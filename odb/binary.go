package odb

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
)

type BinaryBuffer struct {
	io.Writer
	io.Reader
}

func (bw *BinaryBuffer) SwitchToReader(reader io.Reader) {
	bw.Reader = reader
	bw.Writer = nil
}

func (bw *BinaryBuffer) SwitchToWriter(writer io.Writer) {
	bw.Reader = nil
	bw.Writer = writer
}

func (bw *BinaryBuffer) WriteValues(values ...interface{}) (int, error) {
	count := int(0)
	var err error
	for _, v := range values {
		switch v := v.(type) {
		case int32:
			count += 4
			err = bw.WriteInt32(v)
		case int64:
			count += 8
			err = bw.WriteInt64(v)
		case string:
			var sz int
			sz, err = bw.WriteString(v)
			count += sz
		case *TypedMap:
			var sz int
			sz, err = bw.WriteTypedMap(v)
			count += sz
		default:
			err = errInvalidType
		}
		if err != nil {
			return count, err
		}
	}
	return count, nil
}
func (bw *BinaryBuffer) WriteInt32(val int32) error {
	return binary.Write(bw, binary.BigEndian, val)
}
func (bw *BinaryBuffer) WriteInt64(val int64) error {
	return binary.Write(bw, binary.BigEndian, val)
}
func (bw *BinaryBuffer) WriteString(val string) (int, error) {
	buf := []byte(val)
	err := bw.WriteInt32(int32(len(buf)))
	if err != nil {
		return 0, err
	}
	return bw.Write(buf)
}
func (bw *BinaryBuffer) WriteTypedMap(obj *TypedMap) (int, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return 0, err
	}
	return bw.Write(buf.Bytes())
}

func (bw *BinaryBuffer) ReadInt32() (int32, error) {
	var out int32
	err := binary.Read(bw, binary.BigEndian, &out)
	return out, err
}

func (bw *BinaryBuffer) ReadInt64() (int64, error) {
	var out int64
	err := binary.Read(bw, binary.BigEndian, &out)
	return out, err
}

func (bw *BinaryBuffer) ReadString() (string, error) {
	sz, err := bw.ReadInt32()
	if err != nil {
		return "", err
	}
	buf := make([]byte, int(sz))
	_, err = bw.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (bw *BinaryBuffer) ReadTypedMap(out *TypedMap) error {
	dec := gob.NewDecoder(bw)
	err := dec.Decode(out)
	return err
}
