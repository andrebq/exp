package main

import (
	"bytes"
	"flag"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var (
	bucketServer = flag.String("host", "http://localhost:4001/buckets", "Host to store the bucket")
	bucketName = flag.String("b", "", "Bucket name")
	doGet = flag.Bool("get", false, "Should make a GET instead of POST")
	help = flag.Bool("h", false, "Help")
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	if len(*bucketName) == 0 {
		log.Printf("Invalid bucket name")
		flag.Usage()
		return
	}

	if *doGet {
		log.Printf("GET isn't implemented")
		flag.Usage()
		return
	}

	obj := make(map[string]interface{})
	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&obj); err != nil {
		log.Printf("error decoding input: %v", err)
		return
	}

	reqData := make(map[string]interface{})
	reqData["Bucket"] = bucketName
	reqData["Info"] = obj

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.Encode(reqData)
	url := *bucketServer + "/new"
	log.Printf("Sending to: %v", url)

	if req, err := http.NewRequest("POST", url, buf); err != nil {
		log.Printf("Error preparing the request: %v", err)
		return
	} else {
		if res, err := http.DefaultClient.Do(req); err != nil {
			log.Printf("Error sending bucket to server: %v", err)
			return
		} else {
			defer res.Body.Close()
			if res.StatusCode != 200 {
				log.Printf("invalid status code. expecting 200 got %v", res.StatusCode)
				data, _ := ioutil.ReadAll(res.Body)
				log.Printf(string(data))
			}
		}
	}
}
