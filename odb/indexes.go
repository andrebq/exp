package odb

import (
	"bytes"
	"fmt"
	"github.com/cznic/kv"
)

var (
	oidKey = []byte("oid")
)

type OidIndex struct {
	*kv.DB
}

func (o *OidIndex) Name() string {
	return "core_oid"
}

func (o *OidIndex) incOid() (int64, error) {
	return o.Inc(oidKey, 1)
}

func (o *OidIndex) Write(dbe *DBEntry) error {
	nid, err := o.incOid()
	if err != nil {
		return err
	}
	dbe.SetId(nid)
	buf := &bytes.Buffer{}
	buf.Grow(8)
	bw := &BinaryWriter{buf, nil}
	bw.WriteInt64(dbe.Id())
	bw.WriteTypedMap(&dbe.TypedMap)
	bin := buf.Bytes()
	return o.Set(bin[:8], bin[8:])
}

func (o *OidIndex) ExplainError(err error, isKey bool) error {
	return fmt.Errorf("error %v (was key? %v)", err, isKey)
}

func (o *OidIndex) Find(values ...interface{}) (*DBEntry, error) {
	buf := &bytes.Buffer{}
	bw := BinaryWriter{buf, nil}
	if len(values) != 1 {
		return nil, errInvalidIndexFind
	}
	if k, ok := values[0].(int64); ok {
		bw.WriteInt64(k)
		val, err := o.Get(nil, buf.Bytes())
		if err != nil {
			return nil, newError(UnableToReadStorage, "unable to read storage. cause: %v", err)
		}
		buf = bytes.NewBuffer(val)
		bw.SwitchToReader(buf)
		out := &DBEntry{
			Object: NewObject(),
		}
		err = bw.ReadTypedMap(&out.TypedMap)
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	return nil, errInvalidIndexFind
}
