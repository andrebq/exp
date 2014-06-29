package pgstore

// Copyright (c) 2014 André Luiz Alves Moraes
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"bytes"
	"github.com/andrebq/exp/pandora"
	"net/url"
	"testing"
	"time"
)

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

func mustCreateMessageStore(store *MessageStore, err error) *MessageStore {
	if err != nil {
		panic(err)
	}
	err = store.InitTables()
	if err != nil {
		panic(err)
	}
	return store
}

func TestOpenMessageStore(t *testing.T) {
	store, err := OpenMessageStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	err = store.InitTables()
	if err != nil {
		t.Fatalf("error initializing tables: %v", err)
	}
}

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

func TestPandoraServer(t *testing.T) {
	ms := mustCreateMessageStore(OpenMessageStore("pandora", "pandora", "localhost", "pandora"))
	server := pandora.Server{
		BlobStore:    mustCreateBlobStore(OpenBlobStore("pandora", "pandora", "localhost", "pandora")),
		MessageStore: ms,
	}
	defer ms.DeleteMessages()
	body := make(url.Values)
	body.Set("topic", "text")
	msg, err := server.Send("a@local", "b@remote", time.Minute*-5, time.Now(), body)
	if err != nil {
		t.Fatalf("error sending the message: %v", err)
	}

	var emptyKey pandora.SHA1Key
	if bytes.Equal(msg.Mid.Bytes(), emptyKey.Bytes()) {
		t.Errorf("mid cannot be empty or null")
	}

	newMsg, err := server.FetchLatest("b@remote", time.Minute*5)
	if err != nil {
		t.Fatalf("error fetching the message: %v", err)
	}

	if !bytes.Equal(msg.Mid.Bytes(), newMsg.Mid.Bytes()) {
		t.Errorf("mid is different: expecting %v got %v", msg.Mid.Bytes(), newMsg.Mid.Bytes())
	}

	if newMsg.Lid == nil {
		t.Fatalf("lid is nil")
	}

	if bytes.Equal(newMsg.Lid.Bytes(), emptyKey.Bytes()) {
		t.Errorf("lid is empty...")
	}

	if !newMsg.LeasedUntil.After(time.Now()) {
		t.Errorf("invalid lease time, should be after now: %v", newMsg.LeasedUntil)
	}

	if newMsg.Get("topic") != msg.Get("topic") {
		t.Errorf("body is different. expecting %v got %v", msg.Body, newMsg.Body)
	}

	err = server.Ack(newMsg.Mid, newMsg.Lid, pandora.StatusConfirmed)
	if err != nil {
		t.Errorf("error doing ack: %v", err)
	}
}

func TestMessageStorePandoraAPI(t *testing.T) {
	store := mustCreateMessageStore(OpenMessageStore("pandora", "pandora", "localhost", "pandora"))
	defer store.DeleteMessages()

	msg := &pandora.Message{}
	t.Logf("body: %v", msg.Body)
	msg.Empty(nil)
	msg.SetReceiver("test@remote")
	msg.SetSender("test@local")
	msg.Body.Set("hi", "a body")
	t.Logf("body: %v", msg.Body)
	err := store.Enqueue(msg)
	if err != nil {
		t.Fatalf("error saving the message: %v", err)
	}

	newMsg, err := store.FetchAndLockLatest("test@remote", time.Minute*5)
	if err != nil {
		t.Fatalf("error saving fetching the message: %v", err)
	}

	if !bytes.Equal(newMsg.Mid.Bytes(), msg.Mid.Bytes()) {
		t.Fatalf("mid's are different. expecting %v got %v", msg.Mid.Bytes(), newMsg.Mid.Bytes())
	}

	if !newMsg.LeasedUntil.After(time.Now()) {
		t.Fatalf("Invalid lease time. should be %v", newMsg.LeasedUntil)
	}
	var kw pandora.SHA1Key
	if bytes.Equal(kw[:], newMsg.Lid.Bytes()) {
		t.Fatalf("invalid lid. cannot be empty or nil: %v", newMsg.Lid.Bytes())
	}
}

func TestBlobStorePandoraAPI(t *testing.T) {
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
