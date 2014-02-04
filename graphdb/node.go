package main

import (
	"strings"
)

// keyword is a shared identifier that can be used
// to identify node types, edge types and name properties
//
// All keywords should start with ":" to be considered valid
type Keyword struct {
	name string
	val  uint32
}

// NewKeyword prepare the string to be used as a valid keyword
// if a invalid char is found then the returned keyword is marked
// as invalid and can't be used for processing
func NewKeyword(val string) *Keyword {
	if !strings.HasPrefix(val, ":") {
		val = ":" + val
	}
	return &Keyword{
		name: val,
		val:  0,
	}
}

// String is a visual representation of the keyword
func (k *Keyword) String() string {
	return k.name
}

// Valid return true only if the keyword can be used to identify a
// property and isn't a reserverd keyword
func (k *Keyword) Valid() bool {
	return k.val >= minKeywordCode
}

// ValidName returns true when the name of the keyword
// can be used by the database.
func (k *Keyword) ValidName() bool {
	return strings.HasPrefix(k.name, ":")
}

// Node represent the data stored in the graphdb
type Node struct {
	// The identification of this given node
	Id uint64

	// The kind of a node
	//
	// Bob is a :user
	Kind uint32

	// The data related to this node
	Props Properties
}

// Properties is a collection of keys and values (any valid Go type)
// that represent the data of a given node
type Properties map[uint32]interface{}

// Edge represent the connection between two nodes.
type Edge struct {
	// Start and End hold the nodes involved in the relation
	//
	// Relations might be uni-direction but the default is to
	// be bi-directional
	Start, End uint64

	// Kind is the identification of the Edge
	//
	// Bob :knows Alice, :knows is maped to a uint32 value
	// and used as the Kind of the Edge
	Kind uint32

	// The data related describing this edge.
	Props Properties
}
