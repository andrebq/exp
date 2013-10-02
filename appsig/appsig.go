package appsig

import (
	"sync"
)

var (
	// bufferd, to avoid from blocking the calling goroutine
	signals = make(chan *sigData, 10)

	// channel to register new receiver
	register = make(chan *sigReg, 0)
)

// hold all info required to dispatch signals
type sigData struct {
	sig  interface{}
	data []interface{}
}

// hold the information required to register a receiver
type sigReg struct {
	sig  interface{}
	recv Receiver
}

// The handler that must be implemented to receive
// the signals
type Receiver interface {
	// This method is called inside a goroutine
	// so you can block as much as you want and don't need
	// to worry about blocking the process that sent the signal
	Receive(signal interface{}, data ...interface{})
}

// A multi receiver that will dispatch the same signal/data to all childs
type multiReceiver struct {
	sync.RWMutex
	childs []Receiver
}

// add a new receiver
func (mr *multiReceiver) pushChild(c Receiver) {
	mr.Lock()
	defer mr.Unlock()
	mr.childs = append(mr.childs, c)
}

func (mr *multiReceiver) Receive(sig interface{}, data ...interface{}) {
	// prevent panic
	// TODO: think in a better way to do that
	defer func() { _ = recover() }()
	for _, v := range mr.childs {
		go v.Receive(sig, data...)
	}
}

// Receiver function
type ReceiveFunc func(sig interface{}, data ...interface{})

// Implement the receive interface
func (s ReceiveFunc) Receive(sig interface{}, data ...interface{}) {
	s(sig, data...)
}

// Register a new signal receiver to the given signal
// usually the signal would be a constant value with non-native type
//
// If a signal should be handled only by the package that sent the signal,
// then just use a private type.
//
//	package foo
//	type sigFoo byte
//	const (
//		sigA = sigFoo(0)
//	)
//
//	func init() {
//		appsig.RegisterFunc(sigA, recvSigA)
//	}
//
//	func SomePublicFunc() {
//		appsig.Signal(sigA, "this will be handled only by recvSigA")
//	}
//
// In the previous example we declared a sigFoo type that is private to package
// foo, then we created a constant value sigA = sigFoo(0). Inside the init function
// we registered the receiver for sigA.
//
// Since sigFoo is private nobody outside package foo will be able to register
// another receiver to sigA. That way, every time SomePublicFunc is called the recvSigA
// function will be activated by appsig.
func RegisterFunc(sig interface{}, h ReceiveFunc) {
	Register(sig, h)
}

// Same logic used by RegisterFunc
func Register(sig interface{}, h Receiver) {
	register <- &sigReg{sig, h}
}

// just to launch the sighandler routine
func init() {
	go handleSignals()
}

// Send a signal to the registered receivers
func Signal(signal interface{}, data ...interface{}) {
	signals <- &sigData{sig: signal, data: data}
}

func dispatchSignal(v *multiReceiver, sig *sigData) {
	v.Receive(sig.sig, sig.data...)
}

func handleSignals() {
	// receiver registry
	registry := map[interface{}]*multiReceiver{}

	for {
		select {
		case sig := <-signals:
			if v, has := registry[sig.sig]; has {
				go dispatchSignal(v, sig)
			}
		case reg := <-register:
			if mr, has := registry[reg.sig]; has {
				mr.pushChild(reg.recv)
			} else {
				mr = &multiReceiver{childs: make([]Receiver, 0, 1)}
				mr.pushChild(reg.recv)
				registry[reg.sig] = mr
			}
		}
	}
}
