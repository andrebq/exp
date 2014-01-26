// A simple client to interact with the graphdb
package main

import (
	"os"
	"flag"
	"fmt"
	"github.com/andrebq/exp/graphdb"
)

var (
	operation = flag.String("op", "none", 
`Operation to execute inside the database can be:
	add-node,
	set-node-data,
	fetch-node,
	create-initial-structure,
	get-keyword`)
	user = flag.String("user", "postgres", "User to open the connection")
	password = flag.String("password", "", "Password to open the connection")
	database = flag.String("database", "graphdb_1", "Database to interact with")
	pgconn *graphdb.PgConn
)

func main() {
	flag.Parse()
	
	var err error
	pgconn, err = graphdb.NewPgConn(*user, *password, "localhost", *database, "disable") 
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening connection with PG. Cause: %v", err)
		os.Exit(2)
	}

	switch *operation {
	case "add-node": cmdAddNode(flag.Args())
	case "set-node-data": cmdSetNodeData()
	case "fetch-node": cmdFetchNode()
	case "create-initial-structure": cmdCreateInitialStructure()
	case "get-keyword": cmdGetKeyword()
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func cmdAddNode(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "You must pass the kind of the node\n")
		flag.Usage()
		os.Exit(3)
	}

	node := graphdb.NewNode(graphdb.NewKeyword(args[0]))
	err := pgconn.SaveNode(node)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving node: %v\n", err)
	}
	fmt.Fprintf(os.Stdout, "%v\n", node.Id)
}

func cmdSetNodeData() {}
func cmdFetchNode() {}
func cmdCreateInitialStructure() {}
func cmdGetKeyword() {}
