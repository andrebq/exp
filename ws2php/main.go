// This is a simple experiment to provide a bridge between
// a websocket capable client and a old http backend

package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type signal struct{}

type method func(*Broker) (interface{}, error)

type brokerMessage struct {
	request method
	done    chan signal
	err     error
	data    interface{}
}

// Hold all websocket sessions that are running at this moment
type Broker struct {
	list       []*Session
	done       chan signal
	message    chan *brokerMessage
	lastSid    uint64
	backendUrl *url.URL
}

func NewBroker() *Broker {
	s := &Broker{
		list: make([]*Session, 0),
	}
	go s.loop()
	return s
}

func (s *Broker) openSessionWithBackend(newSession *Session) error {
	return nil
}

func (s *Broker) CreateSession(cli *websocket.Conn) (*Session, error) {
	m := s.createMessage(func(b *Broker) (interface{}, error) {
		s.lastSid++
		newSession := &Session{
			Cid:     s.lastSid,
			Backend: s.backendUrl,
		}
		err := s.openSessionWithBackend(newSession)
		return newSession, err
	})
	s.message <- m
	<-m.done
	return m.data.(*Session), m.err
}

func (s *Broker) createMessage(m method) *brokerMessage {
	ret := &brokerMessage{
		request: m,
		done:    make(chan signal, 0),
		err:     nil,
		data:    nil,
	}
	return ret
}

func (s *Broker) loop() {
loop:
	for {
		select {
		case <-s.done:
			break loop
		case m := <-s.message:
			s.processMessage(m)
		}
	}
}

func (s *Broker) processMessage(m *brokerMessage) {
	defer func(m *brokerMessage) {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				m.err = err
			default:
				m.err = fmt.Errorf("Error: %v", err)
			}
		}
		go func(m *brokerMessage) { m.done <- signal{} }(m)
	}(m)
	m.data, m.err = m.request(s)
}

// Session between the websocket and the backend
type Session struct {
	// Client id
	Cid uint64

	// Backend base url
	Backend *url.URL

	// Conn
	Socket *websocket.Conn
}

// Send data to the client
func (s *Session) WriteClient(buf []byte) error {
	_, err := s.Socket.Write(buf)
	return err
}

// Send data to the backend
func (s *Session) WriteBackend(buf []byte) error {
	// TODO: put some http status handling here
	_, err := http.Post(s.BackendPOSTURL().String(), "application/json", bytes.NewBuffer(buf))
	return err
}

func (s *Session) Close() error {
	return s.Socket.Close()
}

func (s *Session) readFromClient() chan []byte {
	ch := make(chan []byte, 1)
	go func(ch chan []byte) {
		// TODO should consume buffers from a pool to avoid
		// creating too much garbage.
		for {
			// change this to some tokenizer
			buf := make([]byte, 1024)
			s.Socket.Read(buf)
			if len(buf) == 0 {
				close(ch)
				return
			}
			// send the buffer down the road
			ch <- buf
		}
	}(ch)
	return ch
}

func (s *Session) readFromBackend() chan []byte {
	ch := make(chan []byte, 1)
	url := s.BackendGETURL().String()
	go func(ch chan []byte) {
		// TODO should consume buffers from a pool to avoid
		// creating too much garbage.
		for {
			// http already have a tokenizer, ignore
			resp, err := http.Get(url)
			if err != nil {
				close(ch)
				return
			}
			if resp.StatusCode != 200 {
				// ignore and read next
				continue
			}
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				close(ch)
				return
			}
			// send data down the road
			ch <- buf
		}
	}(ch)
	return ch
}

func (s *Session) BackendGETURL() *url.URL {
	get := &url.URL{Path: fmt.Sprintf("./%v", s.Cid)}
	return s.Backend.ResolveReference(get)
}

func (s *Session) BackendPOSTURL() *url.URL {
	// at this moment they are the same
	// maybe this will change in the future
	//
	// the backend MUST BE able to handle GET/POST properly
	return s.BackendGETURL()
}

var (
	broker  = NewBroker()
	port    = flag.String("p", ":8081", "Port to listen for websocket connections")
	backend = flag.String("backendUrl", "http://localhost:8082", "Backend url")
	runPong = flag.Bool("runPong", false, "Run the sample ping/pong backend")
	help    = flag.Bool("h", false, "Show this menu")
)

func handleWebSocket(conn *websocket.Conn) {
	// need to handle session termination
	// and io errors
	session, err := broker.CreateSession(conn)
	backendInput := session.readFromBackend()
	clientInput := session.readFromClient()
	if err != nil {
		log.Printf("Error creating session: %v", err)
		defer conn.Close()
	}

	t := time.NewTicker(10 * time.Second)
loop:
	for {
		select {
		case data := <-backendInput:
			session.WriteClient(data)
		case data := <-clientInput:
			session.WriteBackend(data)
		case <-t.C:
			session.Close()
			break loop
		}
	}
}

func startProxy() {
	http.Handle("/ws", websocket.Handler(handleWebSocket))

	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(1)
	}

	if *runPong {
		startPong()
	} else {
		startProxy()
	}
}
