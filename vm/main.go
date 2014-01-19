package main

import (
	"fmt"
)

type NodeType byte

const (
	SYM  = NodeType(1)
	STR  = NodeType(2)
	NUM  = NodeType(3)
	OEXP = NodeType(4)
	CEXP = NodeType(5)
)

func (nt NodeType) String() string {
	switch nt {
	case SYM:
		return "sym"
	case STR:
		return "str"
	case NUM:
		return "num"
	case OEXP:
		return "oexp"
	case CEXP:
		return "cexp"
	default:
		panic("not reached")
	}
}

type Node struct {
	Kind NodeType
	Text string
}

func NewNode(kind NodeType, text string) *Node {
	return &Node{Kind: kind, Text: text}
}

func (n *Node) String() string {
	return fmt.Sprintf("<%v:%v>", n.Kind, n.Text)
}

func main() {
	input := "( ab        )\r\n((\r\n()\t))\r\n()"
	for len(input) > 0 {
		var node *Node
		var err error
		input, node, err = getNode(input)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			break
		} else {
			fmt.Printf("node: %v\n", node)
			if len(input) == 0 {
				fmt.Printf(">>>EOF")
			}
		}
	}
}
