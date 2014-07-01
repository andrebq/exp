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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andrebq/exp/pandora"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type jsonOutput struct {
	val interface{}
}

const (
	ErrNotFound     = pandora.ApiError("not found")
	ErrPOSTRequired = pandora.ApiError("POST is required")
)

var (
	ErrServerPanic = errors.New("bad behavior from the server....")
)

// Handler is the base type used to process
// http request to a pandora server
type Handler struct {
	Server     *pandora.Server
	AllowAdmin bool
}

func (ph *Handler) respondWith(w http.ResponseWriter, req *http.Request, val interface{}) {
	switch val := val.(type) {
	case pandora.ApiError:
		if val == ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else if val == ErrPOSTRequired {
			w.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error=%v", url.QueryEscape(val.Error()))
		}
	case error:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error=%v", url.QueryEscape(val.Error()))
	case int:
		w.WriteHeader(val)
	case url.Values:
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, val.Encode())
	case map[string][]string:
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, url.Values(val).Encode())
	case []byte:
		w.WriteHeader(http.StatusOK)
		w.Write(val)
	case io.Reader:
		w.WriteHeader(http.StatusOK)
		io.Copy(w, val)
	case jsonOutput:
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		enc.Encode(val.val)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "response=%v", url.QueryEscape(fmt.Sprintf("%v", val)))
	}
}

// ServeHTTP processa todas as requisições
func (ph *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var ret interface{}
	defer func() {
		// respond even on panic
		if rec := recover(); rec != nil {
			log.Printf("[PANDORA-HANDLER-panic] %v", rec)
			ret = ErrServerPanic
		}
		ph.respondWith(w, req, ret)
	}()
	ret = ph.parseFormIfNeed(req)
	if ret != nil {
		return
	}
	if strings.HasSuffix(req.URL.Path, "/send") {
		ret = ph.Enqueue(req)
	} else if strings.HasSuffix(req.URL.Path, "/fetch") {
		ret = ph.FetchAndLockLatest(req)
	} else if strings.HasSuffix(req.URL.Path, "/ack") {
		ret = ph.Ack(req)
	} else {
		if ph.AllowAdmin {
			ret = ph.ServeAdmin(req)
		} else {
			ret = ErrNotFound
		}
	}
}

func (ph *Handler) ServeAdmin(req *http.Request) interface{} {
	if strings.HasSuffix(req.URL.Path, "/admin/headers") {
		serverTime, err := time.Parse(time.RFC3339Nano, req.Form.Get("receivedat"))
		if err != nil {
			// maybe a duration
			dur, err := time.ParseDuration(req.Form.Get("receivedat"))
			if err != nil {
				// neither duration or time
				return err
			}
			serverTime = time.Now().Add(dur)
		}
		receiver := req.Form.Get("receiver")
		var out [10]pandora.Message
		sz, err := ph.Server.FetchHeaders(out[:], receiver, serverTime)
		if err != nil {
			return err
		}
		if sz == 0 {
			return jsonOutput{}
		}
		final := make([]url.Values, sz)
		for i, _ := range final {
			msg := &out[i]
			output := make(url.Values)
			msg.WriteTo(output)
			final[i] = output
		}
		return jsonOutput{final}
	} else if strings.HasSuffix(req.URL.Path, "/admin/fetchBlob") {
		var kp pandora.KeyPrinter
		var key pandora.SHA1Key
		err := kp.ReadString(&key, req.Form.Get("mid"))
		if err != nil {
			return err
		}
		data, err := ph.Server.BlobStore.GetData(nil, &key)
		if err != nil {
			return err
		}
		return data
	}
	return ErrNotFound
}

func (ph *Handler) FetchAndLockLatest(req *http.Request) interface{} {
	if req.Method != "POST" {
		return ErrPOSTRequired
	}
	receiver := req.Form.Get(pandora.KeyReceiver)
	duration, _ := time.ParseDuration(req.Form.Get(pandora.KeyLeaseTime))
	msg, err := ph.Server.FetchLatest(receiver, duration)
	if err != nil {
		return err
	}
	msg.WriteTo(msg.Body)
	return msg.Body
}

func (ph *Handler) Enqueue(req *http.Request) interface{} {
	if req.Method != "POST" {
		return ErrPOSTRequired
	}
	delay, err := time.ParseDuration(req.Form.Get("delay"))
	if err != nil {
		delay = 0
	}
	req.Form.Del("delay")

	ctime, err := time.Parse(time.RFC3339Nano, req.Form.Get(pandora.KeyClientTime))
	if err != nil {
		ctime = time.Now()
	}
	req.Form.Del(pandora.KeyClientTime)

	msg, err := ph.Server.Send(req.Form.Get(pandora.KeySender), req.Form.Get(pandora.KeyReceiver), delay, ctime, req.Form)
	if err != nil {
		return err
	}
	resp := make(url.Values)
	resp.Set("mid", pandora.KeyPrinter{}.PrintString(msg.Mid))
	return resp
}

func (ph *Handler) Ack(req *http.Request) interface{} {
	var kp pandora.KeyPrinter
	var midK pandora.SHA1Key
	var lidK pandora.SHA1Key
	err := kp.ReadString(&midK, req.Form.Get("mid"))
	if err != nil {
		return err
	}
	err = kp.ReadString(&lidK, req.Form.Get("lid"))
	if err != nil {
		return err
	}

	status, err := strconv.ParseInt(req.Form.Get("statusCode"), 32, 10)
	if err != nil {
		return err
	}

	err = ph.Server.Ack(&midK, &lidK, pandora.AckStatus(status))
	if err != nil {
		return err
	}
	return http.StatusOK
}

func (ph *Handler) parseFormIfNeed(req *http.Request) error {
	if len(req.Form) <= 0 {
		return req.ParseForm()
	}
	return nil
}
