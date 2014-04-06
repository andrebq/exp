package odb

const (
	idBitCount = 0x0000ffffffffffff
	dbBitCount = 0xffff000000000000
)

type TypedMap map[string]interface{}

func (t *TypedMap) get(name string) interface{} {
	val := (*t)[name]
	if val == nil {
		return struct{}{}
	}
	return val
}

func (t *TypedMap) Put(name string, val interface{}) bool {
	switch val.(type) {
	case string, int64, int32:
		(*t)[name] = val
		return true
	}
	return false
}

func (t *TypedMap) Has(name string) bool {
	_, ok := (*t)[name]
	return ok
}

func (t *TypedMap) Int32(name string) int32 {
	if val, ok := t.get(name).(int32); ok {
		return val
	}
	return 0
}

func (t *TypedMap) Int64(name string) int64 {
	if val, ok := t.get(name).(int64); ok {
		return val
	}
	return 0
}

func (t *TypedMap) Uint64(name string) uint64 {
	if val, ok := t.get(name).(uint64); ok {
		return val
	}
	return 0
}

func (t *TypedMap) UInt32(name string) uint32 {
	if val, ok := t.get(name).(uint32); ok {
		return val
	}
	return 0
}

func (t *TypedMap) String(name string) string {
	if val, ok := t.get(name).(string); ok {
		return val
	}
	return ""
}

type Object struct {
	TypedMap
}

func (o *Object) SetVersion(version int32) {
	o.Put("core_version", version)
}

func (o *Object) SetLocalId(id int64) {
	o.updateOid(o.DB(), id)
}

func (o *Object) SetDB(db int32) {
	o.updateOid(db, o.LocalId())
}

func (o *Object) DB() int32 {
	oid := o.Int64("core_oid")
	db := (uint64(oid) & dbBitCount) >> 48
	return int32(db)
}

func (o *Object) LocalId() int64 {
	oid := o.Oid()
	return oid & idBitCount
}

func (o *Object) Oid() int64 {
	return o.Int64("core_oid")
}

func (o *Object) updateOid(db int32, lid int64) {
	dbu := uint64(db) << 48
	o.Put("core_oid", int64(dbu)|lid&idBitCount)
}

func (o *Object) Version() int32 {
	return o.Int32("core_version")
}

func NewObject() *Object {
	return &Object{make(TypedMap)}
}
