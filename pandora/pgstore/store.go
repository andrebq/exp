package pgstore

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/andrebq/exp/pandora"
	_ "github.com/lib/pq"
	"io"
	"io/ioutil"
	"sync"
)

var (
	blobStoreDef = []string{
		`do
		$$
		begin
			create sequence pgstore_seq_blobs increment 1 minvalue 1 maxvalue 9223372036854775807 start 1 cache 1;
		exception when duplicate_table then
		end
		$$ language plpgsql;`,
		`create table if not exists pgstore_blobs(id integer not null default nextval('pgstore_seq_blobs'), blobid bytea, data bytea)`,
		`create table if not exists pgstore_blobs_ref(id integer not null, blobid bytea, delta integer not null)`,
	}

	ErrKeyNotFound = errors.New("key not found")
)

type memoryBuffer interface {
	Bytes() []byte
}

// BlobStore implements pandora.BlobStore using postgresql as backend
type BlobStore struct {
	sync.RWMutex
	conn *sql.DB
}

// InitTables create all the required tables to use the blob store
func (bs *BlobStore) InitTables() (err error) {
	for _, cmd := range blobStoreDef {
		_, thisErr := bs.conn.Exec(cmd)
		if err == nil && thisErr != nil {
			err = thisErr
		}
	}
	return
}

// Close the store
func (bs *BlobStore) Close() error {
	return bs.conn.Close()
}

// GetData read the contents stored under the k key
func (bs *BlobStore) GetData(out []byte, k pandora.Key) ([]byte, error) {
	var id int64
	var size int
	var err error
	exists, err := bs.keyExistsInDb(&id, k)
	if !exists {
		return nil, ErrKeyNotFound
	}
	err = bs.conn.QueryRow(`select bs.id, octet_length(bs.data) from pgstore_blobs bs where bs.blobid = $1`, k.Bytes()).Scan(&id, &size)
	if err != nil {
		return nil, err
	}
	out = sliceOfSize(out, size)
	err = bs.conn.QueryRow(`select bs.data from pgstore_blobs bs where bs.id = $1`, id).Scan(&out)
	return out, err
}

func (bs *BlobStore) UpdateRefCount(k pandora.Key, delta int) error {
	var err error
	var id int64
	exists, err := bs.keyExistsInDb(&id, k)
	if !exists {
		return ErrKeyNotFound
	}
	_, err = bs.conn.Exec(`insert into pgstore_blobs_ref(id, blobid, delta) values ($1, $2, $3)`, id, k.Bytes(), delta)
	return err
}

// PutData write the contents of data and return the key used to store the data
func (bs *BlobStore) PutData(k pandora.Key, data []byte) (pandora.Key, error) {
	kw := pandora.SHA1KeyWriter{}
	kw.Write(data)
	actual := kw.Key()

	if len(k.Bytes()) >= len(actual.Bytes()) {
		// avoid allocating outside the stack
		copy(k.Bytes(), actual.Bytes())
	} else {
		// use the heap
		k = actual
	}
	err := bs.insert(k, data)
	return k, err
}

func (bs *BlobStore) keyExistsInDb(id *int64, key pandora.Key) (bool, error) {
	err := bs.conn.QueryRow("select id from pgstore_blobs_ref where blobid = $1", key.Bytes()).Scan(id)
	if err == sql.ErrNoRows {
		return false, nil
	} else {
		if err != nil {
			return false, err
		}
	}
	return *id > 0, nil
}

// if data have a Bytes() []byte method, then it is used instead of a copy
func (bs *BlobStore) insert(out pandora.Key, data []byte) error {
	var id int64
	var err error
	exists, err := bs.keyExistsInDb(&id, out)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	tx, err := bs.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	err = tx.QueryRow(`insert into pgstore_blobs(blobid, data) values ($1, $2) returning id`, out.Bytes(), data).Scan(&id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`insert into pgstore_blobs_ref(id, blobid, delta) values ($1, $2, $3)`, id, out.Bytes(), 0)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func readFull(data io.Reader) ([]byte, error) {
	switch data := data.(type) {
	case memoryBuffer:
		return data.Bytes(), nil
	default:
		return ioutil.ReadAll(data)
	}
	panic("not reached")
	return nil, nil
}

func sliceOfSize(old []byte, sz int) []byte {
	if cap(old) >= sz {
		return old[0:sz]
	}
	return make([]byte, sz)
}

// OpenBlobStore returns the blobstore using the provided user, password, host and dbname
func OpenBlobStore(user, pwd, host, dbname string) (*BlobStore, error) {
	sqldb, err := sql.Open("postgres", fmt.Sprintf("user=%v dbname=%v password=%v host=%v sslmode=disable", user, dbname, pwd, host))
	if err != nil {
		return nil, err
	}
	return &BlobStore{
		conn: sqldb,
	}, err
}
