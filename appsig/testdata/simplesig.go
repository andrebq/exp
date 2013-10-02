package main

import (
	"fmt"
	"github.com/andrebq/exp/appsig"
	"sync"
)

func handleSig(sig interface{}, data ...interface{}) {
	fmt.Printf("sig: %v / data: %v\n", sig, data)
	counter.Done()
}

type signal byte

const (
	sig1 = signal(1)
)

var (
	counter = &sync.WaitGroup{}
)

func main() {
	println("add counter")
	counter.Add(3)
	appsig.RegisterFunc(sig1, handleSig)
	println("handlesig register")
	appsig.Signal(sig1, "sig 1")
	println("sent sig 1")
	appsig.Signal(sig1, []string{"sig1", "sig2"})
	println("sent sig 1")
	appsig.Signal(sig1, map[string]string{"abc": "123"})
	println("sent sig 1")
	println("waiting...")
	counter.Wait()
}
