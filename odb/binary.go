package odb

import (
	"encoding/binary"
	"encoding/gob"
	"io"
)

type BinaryWriter struct {
	io.Writer
	io.Reader
}

func (bw *BinaryWriter) SwitchToReader(reader io.Reader) {
	bw.Reader = reader
	bw.Writer = nil
}

func (bw *BinaryWriter) SwitchToWriter(writer io.Writer) {
	bw.Reader = nil
	bw.Writer = writer
}

func (bw *BinaryWriter) WriteValue(val interface{}) (int, error) {
	return -1, nil
}
func (bw *BinaryWriter) WriteInt32(val int32) error {
	return binary.Write(bw, binary.BigEndian, val)
}
func (bw *BinaryWriter) WriteInt64(val int64) error {
	return binary.Write(bw, binary.BigEndian, val)
}
func (bw *BinaryWriter) WriteString(val string) (int, error) {
	buf := []byte(val)
	err := bw.WriteInt32(int32(len(buf)))
	if err != nil {
		return 0, err
	}
	return bw.Write(buf)
}
func (bw *BinaryWriter) WriteTypedMap(obj *TypedMap) error {
	enc := gob.NewEncoder(bw)
	return enc.Encode(obj)
}

func (bw *BinaryWriter) ReadInt32() (int32, error) {
	var out int32
	err := binary.Read(bw, binary.BigEndian, &out)
	return out, err
}

func (bw *BinaryWriter) ReadInt64() (int64, error) {
	var out int64
	err := binary.Read(bw, binary.BigEndian, &out)
	return out, err
}

func (bw *BinaryWriter) ReadString() (string, error) {
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

func (bw *BinaryWriter) ReadTypedMap(out *TypedMap) error {
	dec := gob.NewDecoder(bw)
	err := dec.Decode(out)
	return err
}
