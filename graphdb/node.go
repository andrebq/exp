package main

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
