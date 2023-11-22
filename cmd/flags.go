package cmd

import (
	"flag"
	"github.com/elijahelrod/cql_tui/tui"
)

func Run() {
	clusterPtr := flag.String("address", "localhost:6379,", "Cassandra Node server addresses")
	keyspacePtr := flag.String("keyspace", "", "Cassandra keyspace (optional)")
	usernamePtr := flag.String("username", "", "Cassandra username (optional)")
	passwordPtr := flag.String("password", "", "Cassandra password (optional)")

	prettyPrintJson := flag.Bool("pp-json", true, "Pretty print JSON values")

	flag.Parse()

	tui.RunTUI(*clusterPtr, *keyspacePtr, *usernamePtr, *passwordPtr, *prettyPrintJson)
}
