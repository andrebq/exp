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
