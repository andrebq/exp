package graphdb

import (
	"errors"
	"github.com/andrebq/gas"
	"os"
	"testing"
)

func createDbStructure(pg *PgConn, t *testing.T) {
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
}

func dropDbStructure(pg *PgConn, t *testing.T) {
	_, err := pg.db.Exec("DROP DATABASE graphdb_1;")
	if err != nil {
		t.Logf("Error dropping db. %v", err)
	}
}

func postgresUser() (string, string, error) {
	user, pwd := os.Getenv("PGUSER"), os.Getenv("PGPWD")
	if user == "" || pwd == "" {
		return "", "", errors.New("invalid username or password")
	}
	return user, pwd, nil
}

func TestOpenPgConn(t *testing.T) {
	user, pwd, err := postgresUser()
	if err != nil {
		t.Fatalf("%v", err)
	}
	pg, err := NewPgConn(user, pwd, "localhost", "postgres", "disable")
	if err != nil {
		t.Fatalf("error creating database %v", err)
	}
	createDbStructure(pg, t)
	err = pg.Reopen()
	if err != nil {
		t.Errorf("Unable to reopen the database. %v", err)
	}
	defer dropDbStructure(pg, t)
	defer func() {
		err := pg.Close()
		if err != nil {
			t.Errorf("error closing database connection. %v", err)
		}
	}()
}
