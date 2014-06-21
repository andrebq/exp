package pgstore

import (
	"bytes"
	"github.com/andrebq/exp/pandora"
	"testing"
)

func TestOpenBlobStore(t *testing.T) {
	store, err := OpenBlobStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	err = store.InitTables()
	if err != nil {
		t.Fatalf("unable to initialize tables: %v", err)
	}
	defer store.Close()
}

func mustCreateBlobStore(store *BlobStore, err error) *BlobStore {
	if err != nil {
		panic(err)
	}
	err = store.InitTables()
	if err != nil {
		panic(err)
	}
	return store
}

func TestPandoraAPI(t *testing.T) {
	var bs pandora.BlobStore
	var err error
	bs = mustCreateBlobStore(OpenBlobStore("pandora", "pandora", "localhost", "pandora"))
	data := []byte("this is just a dummy text")

	var kw pandora.SHA1KeyWriter
	kw.Write(data)

	var outKey pandora.Key
	outKey = &pandora.SHA1Key{}
	outKey, err = bs.PutData(outKey, data)

	if err != nil {
		t.Fatalf("error saving key: %v", err)
	}

	if !bytes.Equal(outKey.Bytes(), kw.Key().Bytes()) {
		t.Fatalf("expected key %v got %v", kw.Key().Bytes(), outKey.Bytes())
	}

	outData, err := bs.GetData(nil, outKey)
	if err != nil {
		t.Errorf("error reading data from blobstore: %v", err)
	}

	if !bytes.Equal(outData, data) {
		t.Errorf("expected %v got %v for data", data, outData)
	}

	err = bs.UpdateRefCount(outKey, 1)
	if err != nil {
		t.Errorf("unexpected error when incrementing the ref count: %v", err)
	}
	err = bs.UpdateRefCount(outKey, -1)
	if err != nil {
		t.Errorf("unexpected error when decrementing the ref count: %v", err)
	}
}
