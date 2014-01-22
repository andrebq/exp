package graphdb

import (
	"testing"
)

func TestPutAndGet(t *testing.T) {
	attributes := NewAttributes()
	key := ":valid/key"
	value := "valid_value"
	attributes.Put(key, value)
	
	if found, has := attributes.Get(key); !has {
		t.Fatalf("Should have found the key %v", key)
	} else {
		if value != found {
			t.Errorf("Expecting %v got %v", value, found)
		}
	}

	// all keys MUST START with a ":"
	// the Put method is smart enough to include the
	// Prefix but the get method not
	key = "valid/key/2"

	attributes.Put(key, value)
	if found, has := attributes.Get(":" + key); !has {
		t.Fatalf("The Put method should have placed the :")
	} else {
		if value != found {
			t.Errorf("Expecting %v got %v", value, found)
		}
	}

	if _, has := attributes.Get(key); has {
		t.Fatalf("The Get method shouldn't place the extra :")
	}
}
