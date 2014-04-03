package odb

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

func (o *Object) SetId(id int64) {
	o.Put("core_id", id)
}

func (o *Object) SetDB(db int32) {
	o.Put("core_db", db)
}

func (o *Object) DB() int32 {
	return o.Int32("core_db")
}

func (o *Object) Id() int64 {
	return o.Int64("core_id")
}

func (o *Object) Version() int32 {
	return o.Int32("core_version")
}

func NewObject() *Object {
	return &Object{make(TypedMap)}
}
