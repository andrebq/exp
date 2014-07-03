package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

var (
	addr   = flag.String("addr", "http://localhost:4001/", "HttpFS root folder")
	editor = flag.String("editor", "-", "Editor to start, by default just print to stdout")
	write = flag.Bool("write", false, "Read stdin and write to the remote file")
	h      = flag.Bool("h", false, "Help")
)

func main() {
	flag.Parse()
	if *h || len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	root, err := url.Parse(*addr)
	if err != nil {
		log.Fatalf("Error parsing server address: %v", err)
	}

	destPath := flag.Args()[0]
	root.Path = path.Join(root.Path, destPath)

	if *write {
		doWrite(root)
	} else {
		res, err := http.Get(root.String())
		if err != nil {
			log.Fatalf("Error fetching from server: %v", err)
		}
		if res.StatusCode == 500 || res.StatusCode == 404 {
			log.Printf("Invalid status code %v", res.StatusCode)
			log.Printf("Status text: %v", res.Status)
			res.Body.Close()
			return
		} else {
			editFile(root, res.Body, *editor, flag.Args()[1:]...)
		}
	}
}

func doWrite(target *url.URL) {
	res, err := http.Post(target.String(), "application/octet-stream", os.Stdin)
	if err != nil {
		log.Printf("Error sending file: %v", err)
	} else {
		if res.StatusCode != 200 {
			log.Printf("Invalid status code: %v/%v", res.StatusCode, res.Status)
		}
	}
}

func editFile(remoteAddr *url.URL, contents io.ReadCloser, editor string, args ...string) {
	// for now, just print on stdout
	defer contents.Close()
	if editor == "-" {
		io.Copy(os.Stdout, contents)
		return
	}

	// copy to a temporary file
	tmpDir, err := ioutil.TempDir("", "httped")
	if err != nil {
		log.Printf("Error creating temporary file: %v", err)
		return
	}

	fullName := filepath.Join(tmpDir, path.Base(remoteAddr.Path))
	args = append(args, fullName)

	log.Printf("Editor to use: %v / arguments: %v", editor, args)
	log.Printf("Saving contents to: %v", fullName)
	file, err := os.Create(fullName)
	if err != nil {
		log.Printf("Error creating temporary file: %v", err)
		return
	}
	_, err = io.Copy(file, contents)
	if err != nil {
		log.Printf("Error copying data to temporary files: %v", err)
		return
	}
	file.Close()

	cmd := exec.Command(editor, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("Error running command: %v", err)
		return
	}

	file, err = os.Open(fullName)
	if err != nil {
		log.Printf("Error on seek: %v", err)
		return
	}

	if res, err := http.Post(remoteAddr.String(), "application/octet-stream", file); err != nil {
		log.Printf("Error sending the file back to the server: %v", err)
	} else {
		if res.StatusCode != 200 {
			log.Printf("Invalid status code %v", res.StatusCode)
			log.Printf("Status text: %v", res.Status)
		}
	}
}
