package odb

import (
	"testing"
)

func TestInsertAndFindById(t *testing.T) {
	db, err := NewDB("", 1)
	if err != nil {
		t.Fatalf("unable to open db. %v", err)
	}
	obj := NewObject()
	obj.Put("name", "odb")
	obj.SetLocalId(10)
	obj.SetVersion(1)

	obj, err = db.PutObject(obj)
	if err != nil {
		t.Fatalf("unable to put object. %v", err)
	}
	if obj.LocalId() != 10 {
		t.Errorf("object.Id shouldn't be <= 0")
	}
	if obj.DB() != 1 {
		t.Errorf("object.DB should be the same of the database. value: %v", obj.DB())
	}

	nobj, err := db.FindOneByIndex("core_oid", obj.Oid())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nobj.LocalId() != obj.LocalId() {
		t.Errorf("expecting id: %v got %v", obj.LocalId(), nobj.LocalId())
	}
	if nobj.DB() != obj.DB() {
		t.Errorf("expecting db: %v got %v", obj.DB(), nobj.DB())
	}

	if len(nobj.TypedMap) == 0 {
		t.Errorf("nobj len is zero")
	}

	for k, v := range obj.TypedMap {
		if v2, ok := nobj.TypedMap[k]; ok {
			if v != v2 {
				t.Errorf("for key: %v expecting %v got %v", k, v, v2)
			}
		} else {
			t.Errorf("key %v not found on nobj", k)
		}
	}
}
