package graphdb

import (
	"errors"
	"github.com/andrebq/gas"
	"os"
	"testing"
)

var (
	dbCreated = false
)

func createDbStructure(pg *PgConn, t *testing.T) {
	if dbCreated {
		t.Logf("database already created...")
	}
	createdb, err := gas.ReadFile("github.com/andrebq/exp/graphdb/pg_create_db.sql")
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	structure, err := gas.ReadFile("github.com/andrebq/exp/graphdb/pg_create_database-structure.sql")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	_, err = pg.db.Exec(string(createdb))
	if err != nil {
		t.Logf("Error creating database: %v", err)
	}

	_, err = pg.db.Exec(string(structure))
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	dbCreated = true
}

func postgresUser(t *testing.T) (string, string) {
	user, pwd := os.Getenv("PGUSER"), os.Getenv("PGPWD")
	if user == "" || pwd == "" {
		t.Fatalf(errors.New("invalid username or password").Error())
	}
	return user, pwd
}

func TestOpenPgConn(t *testing.T) {
	user, pwd := postgresUser(t)

	pg, err := NewPgConn(user, pwd, "localhost", "postgres", "disable")
	if err != nil {
		t.Fatalf("error creating database %v", err)
	}
	createDbStructure(pg, t)
	err = pg.Reopen()
	if err != nil {
		t.Errorf("Unable to reopen the database. %v", err)
	}
	err = pg.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func createPgConn(user, pwd, host, db, sslmode string, t *testing.T) *PgConn {
	pg, err := NewPgConn(user, pwd, host, db, sslmode)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return pg
}

func TestCreateKeyword(t *testing.T) {
	user, pwd := postgresUser(t)
	pg := createPgConn(user, pwd, "localhost", "graphdb_1", "disable", t)
	createDbStructure(pg, t)
	defer pg.Close()
	key := NewKeyword(":core/test/kw1")
	err := pg.GetKeyword(&key)
	if err != nil {
		t.Errorf("Error while creating keyword: %v", err)
	}
	if key.code <= 0 {
		t.Errorf("key.code should be positive but got %v", key.code)
	}

	if val, has := pg.keywordCache.byName(key.name); !has {
		t.Errorf("key should be in the cache")
	} else {
		if val.name != key.name || val.code != key.code {
			t.Errorf("Cache is different from actual. cache: %v actual: %v",
				val, key)
		}
	}
}

func TestSaveNode(t *testing.T) {
	node := NewNode(NewKeyword(":users"))
	node.Set(NewKeyword(":core/name"), "gopher")

	user, pwd := postgresUser(t)
	pg := createPgConn(user, pwd, "localhost", "graphdb_1", "disable", t)
	createDbStructure(pg, t)
	defer pg.Close()

	err := pg.SaveNode(node)
	if err != nil {
		t.Fatalf("Error saving node: %v", err)
	}

	if !node.ValidId() {
		t.Errorf("Id should be valid. but got: %v", node.Id)
	}
}

func TestFecthNode(t *testing.T) {
	user, pwd := postgresUser(t)
	pg := createPgConn(user, pwd, "localhost", "graphdb_1", "disable", t)
	createDbStructure(pg, t)
	defer pg.Close()

	node := NewNode(NewKeyword(":core/rootnode"))
	node.Id = 1

	err := pg.FetchNode(node)
	if err != nil {
		t.Fatalf("Error reading node. %v", err)
	}

	if node.ContentSize() != 1 {
		t.Errorf("Content size should be %v but got %v", 1, node.ContentSize())
	}

	if _, has := node.Get(NewKeyword(":core/metadata/created_at")); !has {
		t.Errorf("Unable to get keyword: %v", NewKeyword(":core/metadata/created_at"))
	}
}
