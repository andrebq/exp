package main

import (
	"code.google.com/p/goplan9/plan9"
	"net"
	"flag"
	"log"
)

var (
	listen = flag.String("listen", "127.0.0.1:5640", "Address to listen")
	help = flag.Bool("h", false, "Help")
)

type Server struct {
	listener net.Listener
}

func NewServer(lnet, laddr string) (*Server, error) {
	s := &Server{}
	l, err := net.Listen(lnet, laddr)
	if err != nil { return nil, err }
	s.listener = l
	return s, nil
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Start() error {
	for {
		client, err := s.listener.Accept()
		println("got client")
		if err != nil {
			println("error", err)
			return err
		}
		go s.handleClient(client)
	}
	return nil
}

func (s *Server) handleError(err error, c net.Conn) {
	log.Printf("Client %v caused error %v", c, err)
}

func (s *Server) handleClient(c net.Conn) {
	defer c.Close()
	for {
		fcall, err := plan9.ReadFcall(c)
		if err != nil {
			s.handleError(err, c)
			break
		}
		println(fcall.String())
	}
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	server,err := NewServer("tcp", *listen)
	if err != nil {
		log.Fatalf("Unable to create server. Cause: %v", err)
	}
	defer server.Close()
	err = server.Start()
	if err != nil {
		log.Fatalf("Unable to start server. Cause: %v", err)
	}
}
