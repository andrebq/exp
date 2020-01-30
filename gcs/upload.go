package main

import (
	"context"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

func Logout(oldCred string) {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", oldCred)
}

func Login(newCred string) (oldCred string) {
	oldCred = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if newCred != "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", newCred)
	}
	return
}

func doUpload(into, put, cred string) {
	defer Logout(Login(cred))
	t, err := url.Parse(into)
	if err != nil {
		panic(err)
	}
	key := path.Clean(t.Path)
	if strings.HasPrefix(key, "/") {
		key = key[1:]
	}
	reader := openInput(put)
	defer reader.Close()

	t.Path = ""
	t.RawQuery = ""
	bucket, err := blob.OpenBucket(context.Background(), t.String())
	if err != nil {
		panic(err)
	}
	w, err := bucket.NewWriter(context.Background(), key, &blob.WriterOptions{})
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(w, reader)
	if err != nil {
		panic(err)
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}
}

func openInput(name string) io.ReadCloser {
	switch name {
	case "-", "":
		return os.Stdin
	}
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	return file
}
