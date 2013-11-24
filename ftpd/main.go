package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"strings"
	"time"
)

const (
	CMD_INVALID = iota
	CMD_USER
	CMD_PASS
	CMD_PWD
	CMD_TYPE
)

func splitMsg(cmd string) (byte, []string) {
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return 0, parts
	}
	cmd = strings.ToLower(parts[0])
	if len(parts) == 1 {
		parts = nil
	} else {
		parts = parts[1:]
	}
	switch cmd {
	case "user":
		return CMD_USER, parts
	case "pass":
		return CMD_PASS, parts
	case "pwd":
		return CMD_PWD, parts
	case "type":
		return CMD_TYPE, parts
	default:
		return 0, parts
	}
}

var (
	port = flag.String("addr", "0.0.0.0:21", "Address to listen for FTP clients")
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type readFn func() (string, error)
type writeFn func(msg string, args ...interface{}) error

type ftpCli struct {
	c    net.Conn
	out  writeFn
	in   readFn
	pwd  string
	root string
}

func handleClient(cli net.Conn) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("CLN>> %v caused error %v", cli.RemoteAddr(), err)
		}
	}()
	if err := cli.SetDeadline(time.Time{}); err != nil {
		panic(err)
	}
	in := textproto.NewReader(bufio.NewReaderSize(cli, 4096))
	out := textproto.NewWriter(bufio.NewWriter(cli))
	defer cli.Close()

	send := func(msg string, args ...interface{}) error {
		log.Printf("CLN>> sending... %v", fmt.Sprintf(msg, args...))
		err := out.PrintfLine(msg, args...)
		if err != nil {
			log.Printf("CLN>> SEND>> ERR: %v", err)
		}
		return err
	}
	read := func() (string, error) {
		return in.ReadLine()
	}
	ftpCli := &ftpCli{c: cli, out: send, in: read, pwd: "/", root: "."}
	ftpCli.out("220 Ready!")

	for {
		msg, err := ftpCli.in()
		if err == io.EOF {
			log.Printf("eof")
			return
		} else if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		log.Printf("CLN>> DBG>> %v", msg)
		cmd, args := splitMsg(msg)
		log.Printf("args: %v", args)
		switch cmd {
		case CMD_USER:
			handleUser(ftpCli, args)
		case CMD_PASS:
			handlePass(ftpCli, args)
		case CMD_PWD:
			handlePwd(ftpCli, args)
		case CMD_TYPE:
			handleType(ftpCli, args)
		default:
			handleNotDefined(ftpCli, args)
		}
	}
}

func handleNotDefined(cli *ftpCli, args []string) {
	cli.out("502 Command not defined")
}

func handleUser(cli *ftpCli, args []string) {
	cli.out("331 Username ok, need password")
}

func handlePass(cli *ftpCli, args []string) {
	cli.out("230 User logged in")
}

func handlePwd(cli *ftpCli, args []string) {
	cli.out(`257 "%v" is the current directory`, cli.pwd)
}

func handleType(cli *ftpCli, args []string) {
	if len(args) == 0 {
		cli.out(`501 need arguments`)
	} else if strings.EqualFold("I", args[0]) {
		cli.out(`200 OK`)
	} else {
		cli.out(`504 Command not implemented for that parameter`)
	}
}

func serve(l net.Listener) {
	defer l.Close()
	for {
		log.Printf("SRV>> Accept on %v", l.Addr())
		c, err := l.Accept()
		if err != nil {
			log.Printf("SRV>> Error: %v", err)
			continue
		}
		go handleClient(c)
	}
}

func main() {
	log.Printf("Starting ftp server @ %v", *port)
	l, err := net.Listen("tcp", *port)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
	serve(l)
}
