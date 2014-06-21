package pandora

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

func TestLineMessageContent(t *testing.T) {
	// this is a valid message, made with valid utf-8 headers
	// a empty line
	// and a body
	mc := MessageContent{}
	mc.Set([]byte("Header1: ValueHeader1\r\nHeader2: ValueHeader2\n\nthis is the body\nwith a new line"))

	hdr1 := mc.Header("Header1", nil)
	hdr2 := mc.Header("Header2", nil)
	body := mc.Body()

	if !bytes.Equal([]byte("ValueHeader1"), hdr1) {
		t.Errorf("unable to reader value header 1. got: %v", string(hdr1))
	}

	if !bytes.Equal([]byte("ValueHeader2"), hdr2) {
		t.Errorf("unable to reader value header 2. got: %v", string(hdr2))
	}

	if !bytes.Equal([]byte("this is the body\nwith a new line"), body) {
		t.Errorf("unable to reader value body. got: %v", string(body))
	}
}
