package main

import (
	"fmt"
	"os"
)

type Stack struct {
	data []interface{}
	register map[string]interface{}
}

func NewStack() *Stack {
	return &Stack{data: make([]interface{}, 0),
		register: make(map[string]interface{})}
}

func (s *Stack) pushData(d interface{}) {
	s.data = append(s.data, d)
}

func (s *Stack) popData() interface{} {
	if len(s.data) == 0 {
		panic("stack empty")
	}
	d := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return d
}

func (s *Stack) writeRegister(name string, data interface{}) {
	s.register[name] = data
}

func (s *Stack) readRegister(name string) interface{} {
	d, has := s.register[name]
	if !has {
		panic("register not loaded")
	}
	return d
}

func pushInt(s *Stack, literal int) {
	s.pushData(literal)
}

func addInt(s *Stack) {
	a := s.popData().(int)
	b := s.popData().(int)
	pushInt(s, a+b)
}

func store(s *Stack, name string) {
	s.writeRegister(name, s.popData())
}

func load(s *Stack, name string) {
	s.pushData(s.readRegister(name))
}

func printTop(s *Stack) {
	d := s.popData()
	fmt.Fprintf(os.Stdout, "%v\n", d)
	s.pushData(d)
}

func main() {
	s := NewStack()
	pushInt(s, 10)
	pushInt(s, 20)
	addInt(s)
	store(s, "i1")
	pushInt(s, 10)
	load(s, "i1")
	addInt(s)
	printTop(s)
}
