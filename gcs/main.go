package main

import (
	"flag"
)

func main() {
	into := flag.String("into", "", "Target URL to upload values")
	put := flag.String("put", "-", "File to upload to GCS")
	cred := flag.String("service-account", "", "Service account to use")

	flag.Parse()

	if isUploadOperation(*into, *put) {
		doUpload(*into, *put, *cred)
	} else {
		flag.Usage()
	}
}

func isUploadOperation(target, source string) bool {
	return len(target) > 0 && len(source) > 0
}
