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
