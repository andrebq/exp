package graphdb

import (
	"testing"
)

func TestPutAndGet(t *testing.T) {
	// all keys MUST START with a ":"
	// the Put method is smart enough to include the
	// Prefix but the get method not
	key := NewKeyword("valid/key/2")

	if key.name[0] != ':' {
		t.Errorf("NewKeyword should place the extra :")
	}

	key2 := NewKeyword("valid/key/2")
	if !key2.Equals(&key) {
		t.Errorf("Keys should be equals. A: %v, B: %v", key, key2)
	}
}

func TestNodeContents(t *testing.T) {
	key := NewKeyword("attrs/name")
	kind := NewKeyword("kinds/user")
	node := NewNode(kind).Set(key, "gopher")

	if val, has := node.Get(key); !has || val != "gopher" {
		// keywords are invalid
		t.Errorf("attribute not found")
	}
}
