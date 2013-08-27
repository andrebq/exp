package main

import (
	"fmt"
	"os"
)

type any interface{}

type Object struct {
	fields map[string]any
	array []any
}

func NewObject() *Object {
	return &Object{
		fields: make(map[string]any),
		array: make([]any, 0),
	}
}

func (o *Object) writeField(name string, value any) {
	o.fields[name] = value
}

func (o *Object) readField(name string) any {
	value, has := o.fields[name]
	if !has {
		panic("field not defined")
	}
	return value
}

func (o *Object) readIndex(idx int) any {
	return o.array[idx]
}

func (o *Object) writeIndex(idx int, value any) {
	o.array[idx] = value
}

func (o *Object) appendValue(value any) {
	o.array = append(o.array, value)
}

type Stack struct {
	data []any
	register map[string]any
	memory map[int]*Object
	freeMemory int
}

func NewStack() *Stack {
	return &Stack{data: make([]any, 0),
		register: make(map[string]any),
		memory: make(map[int]*Object)}
}

func (s *Stack) pushData(d any) {
	s.data = append(s.data, d)
}

func (s *Stack) popData() any {
	if len(s.data) == 0 {
		panic("stack empty")
	}
	d := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return d
}

func (s *Stack) writeRegister(name string, data any) {
	s.register[name] = data
}

func (s *Stack) readRegister(name string) any {
	d, has := s.register[name]
	if !has {
		panic("register not loaded")
	}
	return d
}

func pushInt(s *Stack, literal any) {
	s.pushData(literal.(int))
}

func addInt(s *Stack) {
	a := s.popData().(int)
	b := s.popData().(int)
	pushInt(s, a+b)
}

func allocObj(s *Stack) {
	s.freeMemory++
	s.memory[s.freeMemory] = NewObject()
	pushInt(s, s.freeMemory)
}

func store(s *Stack, name any) {
	s.writeRegister(name.(string), s.popData())
}

func load(s *Stack, name any) {
	s.pushData(s.readRegister(name.(string)))
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
