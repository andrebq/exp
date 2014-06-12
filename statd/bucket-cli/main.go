package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	bucketServer = flag.String("host", "http://localhost:4001/buckets", "Host to store the bucket")
	bucketName   = flag.String("b", "", "Bucket name")
	doPost        = flag.Bool("post", false, "Should make a GET instead of POST")
	doDelete     = flag.Bool("delete", false, "Remove the data from the server")
	help         = flag.Bool("h", false, "Help")
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	if len(*bucketName) == 0 && (*doPost || *doDelete) {
		log.Printf("Invalid bucket name")
		flag.Usage()
		return
	}

	if *doPost {
		cmdPost()
		return
	}

	if *doDelete {
		cmdDelete()
		return
	}

	cmdGet()

}

func cmdGet() {
	target, err := url.Parse(*bucketServer)
	if err != nil {
		log.Printf("unable to parse url: %v", err)
		return
	}

	if len(*bucketName) > 0 {
		params := make(url.Values)
		params.Add("bucket", *bucketName)
		target.RawQuery = params.Encode()
	}

	urlStr := target.String()
	log.Printf("url: %v", urlStr)
	if req, err := http.NewRequest("GET", urlStr, nil); err != nil {
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
			} else {
				_, err := io.Copy(os.Stdout, res.Body)
				if err != nil {
					log.Printf("error: %v", err)
				}
			}
		}
	}
}

func cmdDelete() {
	target, err := url.Parse(*bucketServer)
	if err != nil {
		log.Printf("unable to parse url: %v", err)
		return
	}
	params := make(url.Values)
	params.Add("bucket", *bucketName)
	target.RawQuery = params.Encode()
	urlStr := target.String()
	log.Printf("url: %v", urlStr)
	if req, err := http.NewRequest("DELETE", urlStr, nil); err != nil {
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

func cmdPost() {
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
			} else {
				data, _ := ioutil.ReadAll(res.Body)
				log.Printf(string(data))
			}
		}
	}
}
