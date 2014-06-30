package pgstore

// Copyright (c) 2014 André Luiz Alves Moraes
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"database/sql"
	"fmt"
	"github.com/andrebq/exp/pandora"
	_ "github.com/lib/pq"
	"io"
	"io/ioutil"
	"sync"
	"time"
)

type querier interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

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

	messageStoreDef = []string{
		`do
		$$
		begin
			create sequence pgstore_seq_messages increment 1 minvalue 1 maxvalue 9223372036854775807 start 1 cache 1;
		exception when duplicate_table then
		end
		$$ language plpgsql;`,
		`create table if not exists pgstore_messages(
			id integer not null default nextval('pgstore_seq_messages'),
			mid bytea not null,
			lid bytea,
			leaseuntil timestamp,
			status int not null,
			receivedat timestamp not null,
			sendwhen timestamp not null,
			deliverycount int not null,
			senderid int not null,
			receiverid int not null
		)`,
		`do
		$$
		begin
			alter table pgstore_messages
				add constraint pgstore_unq_messages_mid unique(mid);
		exception when duplicate_table then
		end
		$$ language plpgsql;`,
		`create table if not exists pgstore_messageboxes (id integer not null default nextval('pgstore_seq_messages'), name text not null)`,
	}

	ErrKeyNotFound = pandora.ApiError("key not found")
)

type memoryBuffer interface {
	Bytes() []byte
}

// MessageStore implements pandora.MessageStore using postgresql as backend
type MessageStore struct {
	sync.RWMutex
	conn *sql.DB
	bs   *BlobStore
}

// OpenMessageStore opens the message store with the given parameters
func OpenMessageStore(user, pwd, host, dbname string) (*MessageStore, error) {
	sqldb, err := sql.Open("postgres", fmt.Sprintf("user=%v dbname=%v password=%v host=%v sslmode=disable", user, dbname, pwd, host))
	if err != nil {
		return nil, err
	}
	return &MessageStore{
		conn: sqldb,
	}, err
}

func (ms *MessageStore) DeleteMessages() error {
	_, err := ms.conn.Exec("delete from pgstore_messages")
	return err
}

func (ms *MessageStore) InitTables() (err error) {
	for _, cmd := range messageStoreDef {
		_, thisErr := ms.conn.Exec(cmd)
		if err == nil && thisErr != nil {
			err = thisErr
		}
	}
	return
}

// take all messages that have a lease time expired and
// remove the lock information.
//
// Confirmed messages aren't touched
func reEnqueueMessages(db querier, now time.Time) error {
	_, err := db.Exec(`update pgstore_messages
		set lid = null, leaseuntil = null
		where lid is not null and leaseuntil < $1 and status <> $2`,
		now, pandora.StatusConfirmed)
	return err
}

func fetchLatestMessage(msg *pandora.Message, db querier, inbox string, now time.Time, dur time.Duration) error {
	err := reEnqueueMessages(db, now)
	if err != nil {
		return err
	}
	inboxId, err := findInbox(db, inbox, false)
	if err != nil {
		return err
	}
	var buf []byte
	var id int64
	err = db.QueryRow(`select id, mid, status, receivedat, sendwhen, deliverycount
		from pgstore_messages
		where receiverid = $1
			and lid is null
			and sendwhen <= $2
			and status <> $3
		order by sendwhen asc
		limit 1`, inboxId, now, pandora.StatusConfirmed).Scan(&id, &buf, &msg.Status, &msg.ReceivedAt, &msg.SendWhen, &msg.DeliveryCount)
	if err != nil {
		return err
	}
	msg.Mid = &pandora.SHA1Key{}
	copy(msg.Mid.Bytes(), buf)
	msg.CalcualteLeaseFor(now, dur)

	_, err = db.Exec("update pgstore_messages set lid = $1, deliverycount = deliverycount + 1, leaseuntil = $2 where id = $3", msg.Lid.Bytes(), msg.LeasedUntil, id)
	return err
}

func findInbox(db querier, inbox string, create bool) (id int64, err error) {
	if len(inbox) == 0 {
		err = pandora.ErrInvalidMailBox
		return
	}
	err = db.QueryRow("select id from pgstore_messageboxes where name = $1", inbox).Scan(&id)
	if err == sql.ErrNoRows {
		if create {
			// create the inbox
			err = db.QueryRow("insert into pgstore_messageboxes(name) values ($1) returning id", inbox).Scan(&id)
		} else {
			err = pandora.ErrSenderNotFound
		}
	}
	return
}

func findSenderReceiver(db querier, msg *pandora.Message, create bool) (sid int64, rid int64, err error) {
	senderName, receiverName := msg.Sender(), msg.Receiver()
	sid, err = findInbox(db, senderName, create)
	if err != nil {
		return
	}

	rid, err = findInbox(db, receiverName, create)
	if err != nil {
		return
	}
	return
}

// Enqueue will place the message inside the receiver inbox
func (ms *MessageStore) Enqueue(msg *pandora.Message) error {
	msg.Status = pandora.StatusNotDelivered
	msg.CalculateMid()
	return doInsideTransaction(ms.conn, func(tx querier) error {
		senderId, receiverId, err := findSenderReceiver(tx, msg, true)
		if err != nil {
			return err
		}
		var id int64
		err = tx.QueryRow("insert into pgstore_messages(mid, status, receivedat, sendwhen, deliverycount, senderid, receiverid) values ($1, $2, $3, $4, $5, $6, $7) returning id",
			msg.Mid.Bytes(), msg.Status, msg.ReceivedAt, msg.SendWhen, msg.DeliveryCount, senderId, receiverId).Scan(&id)
		return err
	})
}

// FetchAndLockLatest will read and lock the latest message for the given receiver
func (ms *MessageStore) FetchAndLockLatest(recv string, dur time.Duration) (*pandora.Message, error) {
	var msg pandora.Message
	err := doInsideTransaction(ms.conn, func(tx querier) error {
		return fetchLatestMessage(&msg, tx, recv, time.Now(), dur)
	})
	return &msg, err
}

func (ms *MessageStore) Ack(mid, lid pandora.Key, status pandora.AckStatus) error {
	return doInsideTransaction(ms.conn, func(tx querier) error {
		var id int64
		switch status {
		case pandora.StatusConfirmed, pandora.StatusRejected:
		default:
			return pandora.ErrUnableToChangeStatus
		}

		err := tx.QueryRow("update pgstore_messages set status = $1 where mid = $2 and lid = $3 and leaseuntil >= $4 returning id", status, mid.Bytes(), lid.Bytes(), time.Now()).Scan(&id)
		if err != nil {
			return err
		}
		if id == 0 {
			return pandora.ErrUnableToChangeStatus
		}
		return nil
	})
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

	if k != nil && len(k.Bytes()) >= len(actual.Bytes()) {
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

func doInsideTransaction(db *sql.DB, fn func(tx querier) error) (err error) {
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		// if we had a panic
		// ensure that we rollback
		// and resend the panic
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
		// save the error from the function
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	err = fn(tx)
	return
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
	return doInsideTransaction(bs.conn, func(tx querier) error {
		err := tx.QueryRow(`insert into pgstore_blobs(blobid, data) values ($1, $2) returning id`, out.Bytes(), data).Scan(&id)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`insert into pgstore_blobs_ref(id, blobid, delta) values ($1, $2, $3)`, id, out.Bytes(), 0)
		if err != nil {
			return err
		}
		return err
	})
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
