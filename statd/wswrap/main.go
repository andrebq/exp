package main

import (
	"bufio"
	"code.google.com/p/go.net/websocket"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
)

var (
	h        = flag.Bool("h", false, "Help")
	wsAddr   = flag.String("wsAddr", "", "Address for the websocket")
	wsOrigin = flag.String("origin", "", "Origin")
)

func main() {
	flag.Parse()
	if *h {
		flag.Usage()
		os.Exit(1)
	}
	args := flag.Args()
	if len(args) == 0 {
		log.Printf("you must provide the program to start")
		os.Exit(1)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("error stdin pipe: %v", err)
		os.Exit(1)
	}
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("error stdout pipe: %v", err)
		os.Exit(1)
	}
	conn, err := websocket.Dial(*wsAddr, "", *wsOrigin)
	if err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	go lineCopy(cmdIn, conn)
	go lineCopy(conn, cmdOut)

	err = cmd.Start()
	if err != nil {
		conn.Close()
		log.Printf("error starting command: %v", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		conn.Close()
		log.Printf("error on wait: %v", err)
		os.Exit(2)
	}

	conn.Close()
}

// Close out when there is a error reading from in
func lineCopy(out io.WriteCloser, in io.Reader) {
	defer out.Close()
	reader := bufio.NewScanner(in)
	for reader.Scan() {
		line := reader.Bytes()
		_, err := out.Write(line)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		io.WriteString(out, "\n")
	}
	out.Close()
}
