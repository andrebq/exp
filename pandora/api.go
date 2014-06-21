package pandora

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"regexp"
	"time"
	"unicode/utf8"
)

var (
	lineRegexp = regexp.MustCompile("[^\r\n]*\r?\n")
)

// ApiError is used to define the possible error types
type ApiError string

// Error implement the error interface
func (ae ApiError) Error() string {
	return string(ae)
}

const (
	// Key is too short to be used by the given function
	ErrKeyTooShort = ApiError("key is too short")

	// When the given input is too short to be processed by the given function
	ErrInputTooShort = ApiError("input is to short to process")

	// Header isn't a valid utf8 string
	ErrInvalidHeaderEncoding = ApiError("header should be a valid utf-8")
)

// KeyPrinter is used to print any key to a human readable format,
// and parsing it back.
//
// The format used is hex.Encode/hex.Decode
type KeyPrinter struct{}

// PrintKeyString uses the default KeyPrinter to print a key
func PrintKeyString(k Key) string {
	return KeyPrinter{}.PrintString(k)
}

// String return the string representation of the value
func (kp KeyPrinter) PrintString(k Key) string {
	out := make([]byte, hex.EncodedLen(len(k.Bytes())))
	out = kp.Print(out, k)
	return string(out)
}

// Print will return the encoded value of k, if out is large enough
// no allocation is made.
//
// It is valid to pass nil as out
func (kp KeyPrinter) Print(out []byte, k Key) []byte {
	out = kp.Grow(out, k)
	hex.Encode(out, k.Bytes())
	return out
}

// Read will decode in to out, out should be large enough to hold
// the decoded length of in. If this isn't true ErrKeyTooShort is returned
//
// If len(in) != kp.EncodedSize(out) then an error is returned, it is
// invalid to pass nil as out.
func (kp KeyPrinter) Read(out Key, in []byte) (err error) {
	size := kp.EncodedSize(out)
	if len(in) > size {
		err = ErrKeyTooShort
		return
	} else if len(in) < size {
		err = ErrInputTooShort
		return
	}
	_, err = hex.Decode(out.Bytes(), in)
	return

}

// ReadString works just like Read but expects a string
func (kp KeyPrinter) ReadString(out Key, in string) (err error) {
	inByte := []byte(in)
	return kp.Read(out, inByte)
}

// Grow will return an slice that is large enough to hold the encoded
// key.
//
// If old have enough space (cap(old) > kp.EncodedSize(k)) then no allocation
// is done
func (kp KeyPrinter) Grow(old []byte, k Key) []byte {
	es := kp.EncodedSize(k)
	if cap(old) >= es {
		if len(old) < es {
			old = old[0:es]
		}
		return old
	}
	return make([]byte, es)
}

// EncodedSize return the size required to output k using this printer
func (kp KeyPrinter) EncodedSize(k Key) int {
	return hex.EncodedLen(len(k.Bytes()))
}

// Key is used to identify a value inside the BlobStore
type Key interface {
	// Bytes should return the contents of the key
	Bytes() []byte
}

// KeyWriter is used to calculate keys from a given value
type KeyWriter interface {
	io.Writer
	// Key should return the Key calculated from the data
	// passed via Write.
	Key() Key
}

// BlobStore is a simple CAS store used to save messages body
// and other information that is usually searched by a key.
//
// Reference counting is used to identify unused data. When a refcount
// reaches 0 the data don't need to be collected. Calls to Get MIGHT return
// the old value until the data is collected and this isn't considered
// a error.
type BlobStore interface {
	// PutData take data and returns the key under which that data
	// can be retreived later.
	//
	// If key is large enough to hold the Key, then no allocation is done.
	//
	// Is valid to pass nil as "key"
	PutData(key Key, data []byte) (Key, error)

	// GetData will return the contents under key to the user,
	// if the key isn't found, then nil, nil is returned.
	//
	// If out is large enough to hold the data, no allocation is done,
	// otherwise a new buffer is used.
	//
	// Is valid to pass nil as out
	GetData(out []byte, key Key) ([]byte, error)

	// UpdateRef will change the ref-count of the given key by delta.
	//
	// delta can be positive or negative. If the ref-count becomes 0
	// or less, then the key SHOULD BE collected.
	UpdateRefCount(key Key, delta int) error
}

// MessageHeader is the header used to index the message
type MessageHeader struct {
	// Id of the message, calculated based on the
	// MessageContent
	Id Key
	// SendAt holds the client time when the message was sent
	ReceivedAt time.Time
	// ServerTime holds the time of the server when the message was sent
	SendTime time.Time
	// DeliveryCount count how many times the message were delivered to a client.
	//
	// Only one client can access the message at any given time, but when the client crashes
	// or don't complete the message, then another client might access the message.
	DeliveryCount int
}

func bufOfSize(in []byte, sz int) []byte {
	if cap(in) >= sz {
		in = in[:sz]
		return in
	}
	return make([]byte, sz)
}

func copyBuf(out []byte, in []byte) []byte {
	out = bufOfSize(out, len(in))
	copy(out, in)
	return out
}

type lineScan struct {
	data []byte
	sz   int
}

// scanUntilEmptyLine will read data until a empty line is found
//
// an empty line is just a line that have the '\n|\r\n' on the first position
func (ls *lineScan) scanUntilEmptyLine() bool {
	for ls.scanLine(false) {
		// consume all lines until we have a negative
	}
	// consume the next line that should be an empty line
	ret := ls.scanLine(true)
	return ret
}

func (ls *lineScan) scanLine(allowEmpty bool) bool {
	if ls.eof() {
		return false
	}
	buf := ls.data[ls.sz:]
	found := lineRegexp.Find(buf)
	trimmed := bytes.Trim(found, "\r\n")
	if len(trimmed) == 0 {
		if allowEmpty {
			ls.sz += len(found)
			// consume the newline
			return true
		}
		return false
	}
	ls.sz += len(found)
	return true
}

func (ls *lineScan) peek() []byte {
	return ls.data[:ls.sz]
}

func (ls *lineScan) discard() {
	ls.data = ls.data[ls.sz:]
	ls.sz = 0
}

func (ls *lineScan) read(out []byte) []byte {
	out = bufOfSize(out, ls.sz)
	copy(out, ls.data[:ls.sz])
	ls.data = ls.data[ls.sz:]
	ls.sz = 0
	return out
}

func (ls *lineScan) eol() bool {
	buf := ls.data[ls.sz:]
	found := lineRegexp.Find(buf)
	trimmed := bytes.Trim(found, "\r\n")
	return len(trimmed) == 0
}

func (ls *lineScan) eof() bool {
	return ls.sz >= len(ls.data)
}

// MessageContent holds the body of the message
type MessageContent struct {
	full []byte
	hdrs []byte
	body []byte
}

// Set can be used to update the contents of the message,
// the only validation here is to ensure that the headers inside
// contents are utf8 encoded.
//
// The body of the message is just a raw array of bytes and no validation
// is done.
func (mc *MessageContent) Set(contents []byte) error {
	ls := lineScan{contents, 0}
	if err := mc.validContent(&ls); err != nil {
		return err
	}
	mc.full = contents
	mc.hdrs = ls.peek()
	ls.discard()
	mc.body = ls.data
	return nil
}

// Return ErrInvalidHeaderEncoding if the header isn't encoded with
// utf8.
//
// scanner will be placed after the headers, ie, calling scanner.peek()
// will return the headers with the empty line and the body
func (mc *MessageContent) validContent(scanner *lineScan) error {
	scanner.scanUntilEmptyLine()
	if !utf8.Valid(scanner.peek()) {
		return ErrInvalidHeaderEncoding
	}
	return nil
}

// Header search for the given header inside the MessageBody.
//
// If the header is found then the header value is copied to out
func (mc *MessageContent) Header(name string, out []byte) []byte {
	nameB := []byte(name)
	ls := lineScan{mc.hdrs, 0}
	for ls.scanLine(false) {
		parts := bytes.Split(ls.peek(), []byte(":"))
		if bytes.Equal(parts[0], nameB) {
			return copyBuf(out, bytes.Trim(parts[1], " \r\n"))
		}
		ls.discard()
	}
	return nil
}

// Body return the
func (mc *MessageContent) Body() []byte {
	return mc.body
}

// WriteTo writes the message to the given writer
func (mc *MessageContent) WriteTo(w io.Writer) (int, error) {
	return w.Write(mc.full)
}

// Write will set the contents of the message, it will write everything or
// nothing.
//
// The data is copied from msg to the message, if you don't want allocations,
// use Set instead.
//
// Consecutive calls to Write will overwrite the content.
func (mc *MessageContent) Write(msg []byte) (int, error) {
	ls := lineScan{msg, 0}
	if err := mc.validContent(&ls); err != nil {
		return 0, err
	}
	mc.Set(copyBuf(nil, msg))
	return len(msg), nil
}

// Read copy the contents of the message to the given buffer,
// if there isn't enough space in out, a short read error is returned.
//
// Calling Read two times will write the entire message two times or return an error
// two times.
func (mc *MessageContent) Read(out []byte) (int, error) {
	if len(out) < len(mc.full) {
		return 0, io.ErrShortBuffer
	}
	return copy(out, mc.full), nil
}

// Message holds the body of the message and extra headers
type Message struct {
	MessageHeader
	MessageContent
}

// SHA1Key store a sha1 hash as the key
type SHA1Key [20]byte

// Bytes return the slice holding all 20 bytes of the hash
func (s *SHA1Key) Bytes() []byte {
	return (*s)[:]
}

// Type used to calculate the value of a SHA1Key
type SHA1KeyWriter struct {
	h       hash.Hash
	k       SHA1Key
	invalid bool
}

// Write update the sha1 hash with b bytes
func (s *SHA1KeyWriter) Write(b []byte) (int, error) {
	s.invalid = true
	if s.h == nil {
		s.h = sha1.New()
	}
	return s.h.Write(b)
}

// Key will return the updated hash, if no write were made
// the key will be a 0 byte array
func (s *SHA1KeyWriter) Key() Key {
	if s.invalid {
		if s.h == nil {
			return &(s.k)
		}
		copy(s.k.Bytes(), s.h.Sum(nil))
	}
	return &(s.k)
}
