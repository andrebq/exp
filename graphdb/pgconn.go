package graphdb

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PgConn struct {
	db                                   *sql.DB
	username, pwd, dbname, sslmode, host string
}

func NewPgConn(username, pwd, host, db, sslmode string) (*PgConn, error) {
	pg := &PgConn{
		username: username,
		pwd:      pwd,
		dbname:   db,
		sslmode:  sslmode,
		host:     host,
	}
	err := pg.open()
	return pg, err
}

func (pg *PgConn) open() error {
	var err error
	pg.db, err = sql.Open("postgres",
		fmt.Sprintf("user=%v dbname=%v password=%v host=%v sslmode=%v",
			pg.username,
			pg.dbname,
			pg.pwd,
			pg.host,
			pg.sslmode))
	return err
}

func (pg *PgConn) Close() error {
	return pg.db.Close()
}

func (pg *PgConn) Reopen() error {
	err := pg.Close()
	if err != nil {
		return err
	}
	return pg.open()
}

// GetKeyword can be used to return the valid keyword from the given
// PgConn. If the keyword already exists in the database nothing is
// done and the key variable is configured to represent the old keyword.
//
// If the keyword is new, the it is created and the new value is
// configured in the key variable.
//
// This function is atomic and don't need to be inside a transaction
func (pg *PgConn) GetKeyword(key *Keyword) error {
	row := pg.db.QueryRow("select fn_keyword($1)", key.name)
	return row.Scan(&(*key).code)
}
