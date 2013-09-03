// This is a simple experiment to provide a bridge between
// a websocket capable client and a old http backend

package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
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
}

var (
	broker = NewBroker()
	port   = flag.String("p", ":8081", "Port to listen for websocket connections")
	help   = flag.Bool("h", false, "Show this menu")
)

func handleWebSocket(conn *websocket.Conn) {
	session, err := broker.CreateSession(conn)
	_ = session
	if err != nil {
		log.Printf("Error creating session: %v", err)
		defer conn.Close()
	}
	//loop:
	//	for {
	//		select {
	//		case data := <-session.DataFromBackend:
	//			session.WriteClient(data)
	//		case data := <-session.DataFromClient:
	//			session.WriteBackend(data)
	//		case <-session.Done:
	//			defer session.CloseClient()
	//			defer session.CloseBackend()
	//			break loop
	//		}
	//	}
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(1)
	}
	http.Handle("/ws", websocket.Handler(handleWebSocket))

	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
