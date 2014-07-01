package http

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
	"github.com/andrebq/exp/pandora"
	"github.com/andrebq/exp/pandora/pgstore"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func mustCreateServer() *pandora.Server {
	bs, err := pgstore.OpenBlobStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		panic(err)
	}

	ms, err := pgstore.OpenMessageStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		panic(err)
	}
	if err := ms.DeleteMessages(); err != nil {
		panic(err)
	}
	return &pandora.Server{
		BlobStore:    bs,
		MessageStore: ms,
	}
}

func TestPandoraAPI(t *testing.T) {
	server := mustCreateServer()
	handler := &Handler{
		server,
	}

	ts := httptest.NewServer(handler)
	defer ts.Close()

	msgToSend := make(url.Values)
	msgToSend.Set(pandora.KeySender, "a@local")
	msgToSend.Set(pandora.KeyReceiver, "b@local")
	msgToSend.Set(pandora.KeyClientTime, time.Now().Format(time.RFC3339Nano))
	msgToSend.Set("info", "testing....")

	res, err := http.PostForm(ts.URL+"/send", msgToSend)
	if err != nil {
		t.Fatalf("error sending post: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("invalid status code. should be 200 got %v", res.StatusCode)
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading response: %v", err)
	}
	t.Logf("buf: %v", string(buf))
	respValues, err := url.ParseQuery(string(buf))
	if err != nil {
		t.Fatalf("error parsing response url: %v", err)
	}
	if len(respValues.Get("mid")) == 0 {
		t.Fatalf("mid shouldn't be empty")
	}

	createdMid := respValues.Get("mid")

	// now, try to consume the message
	msgToFetch := make(url.Values)
	msgToFetch.Set(pandora.KeyReceiver, "b@local")
	msgToFetch.Set(pandora.KeyLeaseTime, "5m")

	t.Logf("starting fetch")

	res, err = http.PostForm(ts.URL+"/fetch", msgToFetch)
	if err != nil {
		t.Fatalf("error fetching the message: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("invalid status code. should be 200 got %v", res.StatusCode)
	}

	buf, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading response: %v", err)
	}
	t.Logf("buf: %v", string(buf))
	respValues, err = url.ParseQuery(string(buf))
	if err != nil {
		t.Fatalf("error parsing response url: %v", err)
	}
	if len(respValues.Get("mid")) == 0 {
		t.Fatalf("mid shouldn't be empty")
	}

	t.Logf("respValues: %v", respValues.Encode())
	if respValues.Get("mid") != createdMid {
		t.Errorf("expected mid %v got %v", createdMid, respValues.Get("mid"))
	}

	if len(respValues.Get("lid")) == 0 {
		t.Errorf("lid is empty")
	}

	msgToAck := make(url.Values)
	msgToAck.Set("mid", respValues.Get("mid"))
	msgToAck.Set("lid", respValues.Get("lid"))
	msgToAck.Set("statusCode", strconv.FormatInt(int64(pandora.StatusConfirmed), 10))

	res, err = http.PostForm(ts.URL+"/ack", msgToAck)
	if err != nil {
		t.Fatalf("error sending message ack")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: expecting %v got %v", 200, res.StatusCode)
	}

	buf, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("error reading body: %v", err)
	}
	t.Logf("buf: %v", string(buf))
}
