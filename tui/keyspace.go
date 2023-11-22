package tui

import "github.com/gocql/gocql"

type Keyspace struct {
	*gocql.KeyspaceMetadata
}
