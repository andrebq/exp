package graphdb

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

type keywordCache []*Keyword

func (kc *keywordCache) byKind(code int) (Keyword, bool) {
	for _, k := range *kc {
		if k.code == code {
			return *k, true
		}
	}
	return Keyword{}, false
}

func (kc *keywordCache) byName(name string) (Keyword, bool) {
	for _, k := range *kc {
		if k.name == name {
			return *k, true
		}
	}
	return Keyword{}, false
}

type PgConn struct {
	db                                   *sql.DB
	username, pwd, dbname, sslmode, host string
	keywordCache                         keywordCache
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
	if tmp, has := pg.keywordCache.byName(key.name); has {
		*key = tmp
		return nil
	} else {
		row := pg.db.QueryRow("select fn_keyword($1)", key.name)
		err := row.Scan(&(*key).code)
		if err == nil {
			toCache := *key
			pg.keywordCache = append(pg.keywordCache, &toCache)
		}
		return err
	}
}

func (pg *PgConn) getKeywordByCode(out *Keyword) error {
	row := pg.db.QueryRow("select fn_keyword_code($1)", out.code)
	err := row.Scan(&(*out).name)
	if err != nil {
		return err
	}
	if out.name == "" {
		return errors.New("keyword not found")
	}
	return nil
}

// SaveNode inserts or updates the node inside the database
// The id in node is updated if the node was created.
func (pg *PgConn) SaveNode(node *Node) (err error) {
	var tx *sql.Tx
	tx, err = pg.db.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}(tx)
	if !node.ValidId() {
		row := tx.QueryRow("select fn_new_node(fn_keyword($1));",
			node.Kind.name)
		err = row.Scan(&(*node).Id)
		if err != nil {
			return err
		}
	}
	for _, nc := range node.contents {
		_, err = tx.Exec("select fn_node_data($1, fn_keyword($2), $3);",
			node.Id, nc.kind.name, nc.value)
		if err != nil {
			return err
		}
	}
	return err
}

func (pg *PgConn) FetchNode(dest *Node) error {
	rows, err := pg.db.Query("select nodeid, kind, attr, contents from vw_node_with_contents where nodeid = $1", dest.Id)
	if err != nil {
		return err
	}
	defer rows.Close()
	var id uint64
	var kind int
	var ncData string
	var ncKind int

	for rows.Next() {
		err = rows.Scan(&id, &kind, &ncKind, &ncData)
		if err != nil {
			return err
		}
		dest.Id = id
		dest.Kind.code = kind
		dest.Set(newKeyword(ncKind), ncData)
	}

	err = pg.getKeywordByCode(&dest.Kind)
	if err != nil {
		return err
	}

	for _, nc := range dest.contents {
		err = pg.getKeywordByCode(&nc.kind)
		if err != nil {
			return err
		}
	}
	return nil
}
