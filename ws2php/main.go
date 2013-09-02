// This is a simple experiment to provide a bridge between
// a websocket capable client and a old http backend

package main

import (
	"net/http"
	"code.google.com/p/go.net/websocket"
)

type signal struct{}

type method func() (interface{}, error)

type brokerMessage struct {
	request method
	completed chan signal
	err error
	data interface{}
}

// Hold all websocket sessions that are running at this moment
type SessionBroker struct {
	list []*Session
	done chan signal
}

func NewSessionBroker() *SessionBroker {
	s := &SessionBroker{
		list: make([]*Session, 0),
	}
	go s.loop()
}

func (s *SessionBroker) CreateSession(cli websocket.Conn) (*Session, error) {
}

func (s *SessionBroker) createMessage(m method) *brokerMessage {
	ret := brokerMessage {
		request: m,
		completed: make(chan signal, 0),
		err: nil,
		data: nil,
	}
	return ret
}

func (s *SessionBroker) loop() {
loop:
	for {
		select {
		case <-s.done:
			break loop
		}
	}
}

// Session between the websocket and the backend
type Session struct {
	// Client id
	Cid uint64

	// Backend base url
	Backend *url.URL
}

var broker = NewSessionBroker()

func handleWebSocket(conn websocket.Conn) {
	session := broker.CreateSession(conn)
loop:
	for {
		select {
		case data := <-session.DataFromBackend:
			session.WriteClient(data)
		case data := <-session.DataFromClient:
			session.WriteBackend(data)
		case <-session.Done:
			defer session.CloseClient()
			defer session.CloseBackend()
			break loop
		}
	}
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

