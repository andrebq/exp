package pandora

// Copyright (c) 2014 André Luiz Alves Moraes
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/url"
	"strconv"
	"time"
)

// ApiError is used to define the possible error types
type ApiError string

// Error implement the error interface
func (ae ApiError) Error() string {
	return string(ae)
}

const (
	// MaxSize define the max size that any message can have
	MaxSize = 4096
	// Key is too short to be used by the given function
	ErrKeyTooShort = ApiError("key is too short")

	// When the given input is too short to be processed by the given function
	ErrInputTooShort = ApiError("input is to short to process")

	// Header isn't a valid utf8 string
	ErrInvalidHeaderEncoding = ApiError("header should be a valid utf-8")

	// ErrMessageToBig the message data is larger than MaxSize
	ErrMessageToBig = ApiError("message is too big. max size is " + string(MaxSize))

	// ErrNilBody the body passed is nil
	ErrNilBody = ApiError("body is nil")

	// ErrInvalidHeaderEncoding means that a mailbox (sender or receiver) is invalid
	ErrInvalidMailBox = ApiError("invalid mailbox")

	// The sender wasn't found on the server
	ErrSenderNotFound = ApiError("sender not found")

	// The receiver wasn't found on the server
	ErrReceiverNotFound = ApiError("receiver not found")

	// Unable to change the status
	ErrUnableToChangeStatus = ApiError("unable to change status")

	// No messages that match the criteria at this moment
	ErrNoMessages = ApiError("no messages at this moment")

	// Body field used to store the sender
	KeySender = "sender"

	// Body field used to store the receiver
	KeyReceiver = "receiver"

	// Body field used to store the client time
	KeyClientTime = "clientTime"

	// Key used by the client to inform for how long it
	// want to keep the message locked
	KeyLeaseTime = "leaseTime"

	// DefaultLeaseTime is 5 minutes

	DefaultLeaseTime = time.Minute * 5
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
	if k == nil {
		return ""
	}
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

// Message is the header used to index the message
type Message struct {
	// Mid is the id of the message, calculated based on the
	// body of the message
	Mid Key
	// Lid is the id of the current associated lock
	Lid Key
	// Status holds the ack status of this message
	Status AckStatus
	// LeaseUntil holds the time when the Lid will become invalid
	LeasedUntil time.Time
	// FetchTime holds the time at the server when the message was requested by the client.
	FetchTime time.Time
	// ReceivedAt holds the server time when the message was received
	ReceivedAt time.Time
	// SendWhen is used to store when the message should be delivered. This is the ReceivedAt + Delay
	SendWhen time.Time
	// Delay is the time to wait before sending the message
	Delay time.Duration
	// DeliveryCount count how many times the message were delivered to a client.
	//
	// Only one client can access the message at any given time, but when the client crashes
	// or don't complete the message, then another client might access the message.
	DeliveryCount int
	// Body is a list of urlencoded data
	Body url.Values

	invalidBody bool
}

func (m *Message) ensureBody() {
	if m.Body == nil {
		m.Body = make(url.Values)
	}
}

func (m *Message) ValidBody() bool {
	return !m.invalidBody
}

func (m *Message) Sender() string {
	m.ensureBody()
	return m.Body.Get(KeySender)
}

func (m *Message) Receiver() string {
	m.ensureBody()
	return m.Body.Get(KeyReceiver)
}

func (m *Message) SetSender(sender string) {
	m.ensureBody()
	m.Body.Set(KeySender, sender)
}

func (m *Message) SetReceiver(recv string) {
	m.ensureBody()
	m.Body.Set(KeyReceiver, recv)
}

func (m *Message) SetClientTime(ct time.Time) {
	m.ensureBody()
	m.Body.Set(KeyClientTime, ct.Format(time.RFC3339Nano))
}

func (m *Message) ClientTime() time.Time {
	m.ensureBody()
	t, _ := time.Parse(time.RFC3339Nano, m.Body.Get(KeyClientTime))
	return t
}

func (m *Message) Set(key, value string) {
	m.ensureBody()
	m.Body.Set(key, value)
}

func (m *Message) Get(key string) string {
	m.ensureBody()
	return m.Body.Get(key)
}

func (m *Message) CalcualteLeaseFor(now time.Time, lease time.Duration) {
	var kw SHA1KeyWriter
	m.LeasedUntil = now.Add(lease)
	io.WriteString(&kw, m.LeasedUntil.Format(time.RFC3339Nano))
	m.Lid = kw.Key()
}

func (m *Message) WriteTo(out url.Values) {
	var kp KeyPrinter
	out.Set("mid", kp.PrintString(m.Mid))
	out.Set("lid", kp.PrintString(m.Lid))
	out.Set("statusCode", strconv.FormatInt(int64(m.Status), 10))
	out.Set("status", m.Status.String())
	out.Set("leasedUntil", m.LeasedUntil.Format(time.RFC3339Nano))
	out.Set("receivedAt", m.ReceivedAt.Format(time.RFC3339Nano))
	out.Set("deliveryCount", strconv.FormatInt(int64(m.DeliveryCount), 10))
	out.Set("sendWhen", m.SendWhen.Format(time.RFC3339Nano))
	out.Set("validBody", fmt.Sprintf("%v", m.ValidBody()))
}

// Empty will clean all fields of this message and mark the message as
// if it was sent now
//
// No sender or receiver is configured
func (m *Message) Empty(body url.Values) *Message {
	m.Body = body
	if m.Body == nil {
		m.Body = make(url.Values)
	}
	m.SetClientTime(time.Now())
	m.ReceivedAt = m.ClientTime()
	m.SendWhen = m.ClientTime()
	m.DeliveryCount = 0
	m.Delay = 0
	m.Lid = nil
	m.Status = StatusNotDelivered
	m.LeasedUntil = time.Time{}
	return m
}

func (m *Message) CalculateMid() {
	var kw SHA1KeyWriter
	buf := bytes.Buffer{}
	io.WriteString(&buf, m.Body.Encode())
	kw.Write(buf.Bytes())
	m.Mid = kw.Key()
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

// AckStatus define the list of possible status for a given message'
type AckStatus byte

const (
	// StatusConfirmed means that a message was received and processed
	StatusConfirmed = AckStatus(1)
	// StatusRejected means that a message was received but rejected by the client
	StatusRejected = AckStatus(2)
	// StatusTimeout means that a message was sent but the client didn't sent a valid Ack
	StatusTimeout = AckStatus(4)
	// StatusNotDelivered means that a message is waiting for delivery on the queue
	StatusNotDelivered = AckStatus(8)
)

func (a AckStatus) String() string {
	return ackStatuStr[a]
}

var (
	ackStatuStr = map[AckStatus]string{
		StatusConfirmed:    "confirmed",
		StatusRejected:     "rejected",
		StatusTimeout:      "timeout",
		StatusNotDelivered: "notDelivered",
	}
)

// MessageStore defines the required interface to allow the system to work
type MessageStore interface {
	// Enqueue will put msg in the outputbox of the receiver
	Enqueue(msg *Message) error

	// FetchAndLockLatest will fetch the next pending queue that is available for delivery, ie,
	// messages that aren't locked and the SendWhen is less than the current time.
	//
	// This also returns the Lid to be used with the message
	FetchAndLockLatest(receiver string, leaseTime time.Duration) (*Message, error)

	// Ack will change the status of the given mid message, only if lid is still valid
	Ack(mid, lid Key, status AckStatus) error

	// FetchHeaders fetch at least len(out) messages that have the given receiver
	// and were received after serverTime.
	//
	// Only pending messages are returned
	FetchHeaders(out []Message, receiver string, serverTime time.Time) (int, error)

	// Reenqueue messages considering now
	Reenqueue(now time.Time) error
}

// Server implements the pandora message API
type Server struct {
	BlobStore    BlobStore
	MessageStore MessageStore
}

// WriteBlob save the body of the message to the blobstore and writes
// the value back to msg.Mid
func (s *Server) WriteBlob(msg *Message) error {
	buf := &bytes.Buffer{}
	io.WriteString(buf, msg.Body.Encode())
	_, err := s.BlobStore.PutData(msg.Mid, buf.Bytes())
	return err
}

// Send is used to send the givem message contents from sender to receiver,
// sendAt can be used to inform a duration and delay the actual delivery of the message.
//
// delay will always be calculated by the server time.
//
// The message body might be changed by the server by adding headers to it
func (s *Server) Send(sender, receiver string, delay time.Duration, clientTime time.Time, body url.Values) (Message, error) {
	var msg Message
	if body == nil {
		return msg, ErrNilBody
	}
	msg.Body = body
	msg.SetSender(sender)
	msg.SetReceiver(receiver)
	msg.Body.Add("p-server", "Pandora-Default-Server")

	msg.ReceivedAt = time.Now()
	msg.SendWhen = msg.ReceivedAt.Add(delay)
	msg.SetClientTime(clientTime)
	msg.DeliveryCount = 0

	err := s.doSend(&msg)
	return msg, err
}

// FetchLatest fetch the latest message for the given receiver,
// it is possible to fetch the message and not the body (BlobStore is down),
// when that happens the client can check if the body is valid by calling
// Message.ValidBody.
//
// If no error is returned, then the body is valid and there is no need to check that
func (s *Server) FetchLatest(receiver string, lease time.Duration) (*Message, error) {
	if lease <= 0 {
		lease = DefaultLeaseTime
	}
	msg, err := s.MessageStore.FetchAndLockLatest(receiver, lease)
	if err != nil {
		return nil, err
	}
	return s.doReadMessage(msg)
}

// FetchHeaders output at least len(out) messages headers, no body is returned
func (s *Server) FetchHeaders(out []Message, receiver string, receivedAt time.Time) (int, error) {
	return s.MessageStore.FetchHeaders(out, receiver, receivedAt)
}

func (s *Server) doReadMessage(msg *Message) (*Message, error) {
	data, err := s.BlobStore.GetData(nil, msg.Mid)
	if err != nil {
		msg.invalidBody = false
		return msg, err
	}
	msg.Body, err = url.ParseQuery(string(data))
	if err != nil {
		msg.invalidBody = false
	}
	return msg, err
}

func (s *Server) doSend(msg *Message) error {
	if err := s.WriteBlob(msg); err != nil {
		return err
	}
	return s.MessageStore.Enqueue(msg)
}

// Ack is used to confirm that a message mid was processed o rejected by the client.
func (s *Server) Ack(mid, lockId Key, ack AckStatus) error {
	return s.MessageStore.Ack(mid, lockId, ack)
}
