package odb

import (
	"testing"
)

func TestInsertAndFindById(t *testing.T) {
	db, err := NewDB("")
	if err != nil {
		t.Fatalf("unable to open db. %v", err)
	}
	obj := NewObject()
	obj.Put("name", "odb")
	obj.SetDB(1)
	obj.SetVersion(1)

	obj, err = db.PutObject(obj)
	if err != nil {
		t.Fatalf("unable to put object. %v", err)
	}
	if obj.Id() <= 0 {
		t.Errorf("object.Id shouldn't be <= 0")
	}

	nobj, err := db.FindOneByIndex("core_oid", obj.Id())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nobj.Id() != obj.Id() {
		t.Errorf("expecting id: %v got %v", obj.Id(), nobj.Id())
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
