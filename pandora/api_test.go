package pandora

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
	"encoding/hex"
	"testing"
)

func TestSHA1KeyWriter(t *testing.T) {
	input := "hello world"
	expectedHexStr := "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"

	buf, err := hex.DecodeString(expectedHexStr)
	if err != nil {
		t.Fatalf("error decoding hash: %v", err)
	}

	w := SHA1KeyWriter{}
	w.Write([]byte(input))
	k := w.Key()

	if !bytes.Equal(buf, k.Bytes()) {
		t.Errorf("error writing key. expected value is %v got %v", buf, k.Bytes())
	}
}

func TestKeyPrinterWrite(t *testing.T) {
	kp := KeyPrinter{}
	key := SHA1Key{}

	expectedHexStr := "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
	buf, err := hex.DecodeString(expectedHexStr)
	if err != nil {
		t.Fatalf("error decoding hash: %v", err)
	}
	copy(key.Bytes(), buf)

	if kp.PrintString(&key) != expectedHexStr {
		t.Errorf("error printing key. expected value is %v got %v", expectedHexStr, kp.PrintString(&key))
	}
}

func TestKeyPrinterRead(t *testing.T) {
	kp := KeyPrinter{}
	key := SHA1Key{}

	inputHexStr := "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
	buf, err := hex.DecodeString(inputHexStr)
	if err != nil {
		t.Fatalf("error decoding hash: %v", err)
	}

	if err := kp.ReadString(&key, inputHexStr); err != nil {
		t.Fatalf("error reading from string to key. %v", err)
	}
	if !bytes.Equal(key.Bytes(), buf) {
		t.Errorf("error reading key. expected value is %v got %v", buf, key.Bytes())
	}
}
