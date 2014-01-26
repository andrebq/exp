package graphdb

import (
	"math"
	"strings"
)

// Keyword define a Graph keyword, keywords are similar
// to strings but they are stored using much less space
// in the graph.
//
// All Keywords start with a :
type Keyword struct {
	name string
	code int
}

// NewKeyword return a formated Keyword representing the given
// string
func NewKeyword(keyword string) Keyword {
	if !strings.HasPrefix(keyword, ":") {
		keyword = ":" + keyword
	}
	return Keyword{keyword, math.MinInt32}
}

func newKeyword(code int) Keyword {
	return Keyword{code: code}
}

func (k Keyword) Equals(other *Keyword) bool {
	if k.code <= 0 || other.code <= 0 {
		// check the name
		return k.name == other.name
	}
	return k.code == other.code
}

// Node is the representation of a graph node
type Node struct {
	Id       uint64
	Kind     Keyword
	contents []*NodeContent
}

type NodeContent struct {
	kind  Keyword
	value string
}

func newNodeContent(kind Keyword) NodeContent {
	return NodeContent{
		kind: kind,
	}
}

func (nc *NodeContent) Set(value string) *NodeContent {
	nc.value = value
	return nc
}

func (nc *NodeContent) Get() string {
	return nc.value
}

// NewNode return a empty node of the given kind
func NewNode(kind Keyword) *Node {
	return &Node{Id: 0,
		Kind:     kind,
		contents: make([]*NodeContent, 0),
	}
}

func (n *Node) Set(kind Keyword, propValue string) *Node {
	for _, nc := range n.contents {
		if nc.kind.Equals(&kind) {
			nc.Set(propValue)
			return n
		}
	}
	nc := newNodeContent(kind)
	n.contents = append(n.contents, nc.Set(propValue))
	return n
}

func (n *Node) Get(kind Keyword) (string, bool) {
	val := n.Value(kind)
	if val == nil {
		return "", false
	} else {
		return val.Get(), true
	}
}

func (n *Node) ValidId() bool {
	return n.Id > 0
}

func (n *Node) Value(kind Keyword) *NodeContent {
	for _, nc := range n.contents {
		if nc.kind.Equals(&kind) {
			return nc
		}
	}
	return nil
}

func (n *Node) ContentSize() int {
	return len(n.contents)
}
