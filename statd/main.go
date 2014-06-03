// Roberto souza da silva junior
package main

import (
	_ "net/http/pprof"
	"bytes"
	"code.google.com/p/go.net/websocket"
	"log"
	_ "github.com/lib/pq"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"bufio"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	dbuser = flag.String("dbuser", "statsd", "Database user")
	dbpasswd = flag.String("dbpasswd", "statsd", "Database password")
	dbname = flag.String("dbname", "statsd", "Database name")
	dbhost = flag.String("dbhost", "localhost", "Database host")
	initdb = flag.Bool("initdb", false, "Initialize the tables on the database")
	httpaddr = flag.String("httpaddr", "0.0.0.0:4001", "Address to listen for incoming http requests")
	help = flag.Bool("h", false, "Help")

	exitStatus int
)

const (
	DefaultSize = 500
	MaxSize = 1000

	DateTimeFormatFromServer = "2006-01-02 15:04:05.000"
	StatsSelect = `select s.id, s.system, s.subsystem, s.message, s.context, to_char(s.servertime, 'yyyy-mm-dd HH24:MI:SS:MS'), s.clienttime, s.error, si.info from stats s inner join stats_info si on s.id = si.stats_id`
)

type Bucket struct {
	Id int
	Bucket string
	ServerTimeNano uint64
	Info map[string]interface{}
}

type Stats struct {
	Id int
	System string
	SubSystem string
	Message string
	ServerTimeNano uint64
	ServerTime string
	ClientTime string
	Context string
	Error bool
	Info map[string]string
}

type StatsDB struct {
	conn *sql.DB
	newStat chan Stats
	newBucketData chan Bucket
	done chan struct{}
}

func (db *StatsDB) PushBucket(bucket *Bucket) error {
	var lastid int

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(bucket.Info)
	if err != nil {
		return err
	}

	err = db.conn.QueryRow("insert into buckets(bucket, servertime, info) values ($1, $2, $3)) returning id",
		bucket.Bucket, time.Now(), string(buf.Bytes())).Scan(&lastid)
	return err
}

func (db *StatsDB) Push(st *Stats) error {
	var lastid int

	err := db.conn.QueryRow("insert into stats(system, subsystem, message, context, servertime, clienttime, error) values ($1, $2, $3, $4, $5, $6, $7) returning id",
		st.System, st.SubSystem, st.Message, st.Context, time.Now(), st.ClientTime, st.Error).Scan(&lastid)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err = enc.Encode(st.Info)
	if err != nil {
		return err
	}

	_, err = db.conn.Exec("insert into stats_info(stats_id, info) values ($1, $2)",
		lastid, string(buf.Bytes()))
	return err
}

func (db *StatsDB) streamRows(result *sql.Rows, out chan Stats) {
	var err error
	defer result.Close()
LOOP:
	for result.Next() {
		var stat Stats
		var info string
		stat.Info = make(map[string]string)
		err = result.Scan(&stat.Id, &stat.System, &stat.SubSystem, &stat.Message, &stat.Context, &stat.ServerTime, &stat.ClientTime, &stat.Error, &info)
		if err != nil {
			printf("error scanning database: %v", err)
			break
		}
		dec := json.NewDecoder(bytes.NewBufferString(info))
		err = dec.Decode(&stat.Info)
		if err != nil {
			printf("error decoding info from record: %v", err)
			break
		}
		select {
		case out <- stat:
		case <-time.After(time.Second * 10):
			break LOOP
		}
	}
	close(out)
}

func (db *StatsDB) FetchAfterId(lastId, size int) (chan Stats, error) {
	result, err := db.conn.Query(StatsSelect + " where s.id > $1 order by s.context, s.servertime desc limit $2", lastId, size)
	if err != nil {
		printf("error running query: %v", err)
		return nil, err
	}
	out := make(chan Stats, 0)
	go db.streamRows(result, out)
	return out, err
}

func (db *StatsDB) Fetch(size int) (chan Stats, error) {
	result, err := db.conn.Query(StatsSelect + " order by s.context, s.servertime desc limit $1", size)
	if err != nil {
		printf("error running query: %v", err)
		return nil, err
	}
	out := make(chan Stats, 0)
	go db.streamRows(result, out)
	return out, err
}

func NewStatsDB(user, pwd, host, dbname string) (*StatsDB, error) {
	printf("opening database connection to: %v with user %v database %v", host, user, dbname)
	sqldb, err := sql.Open("postgres", fmt.Sprintf("user=%v dbname=%v password=%v host=%v sslmode=disable", user, dbname, pwd, host))
	if err != nil {
		return nil, err
	}
	db := &StatsDB{
		conn: sqldb,
		newStat: make(chan Stats, 1),
		done: make(chan struct{}, 0),
	}
	go db.serve()
	return db, nil
}

func (db *StatsDB) CreateTables() error {
	cmds := []string {
`create sequence stats_seq increment 1 minvalue 1 maxvalue 9223372036854775807 start 1 cache 1`,
`create sequence stats_info_seq increment 1 minvalue 1 maxvalue 9223372036854775807 start 1 cache 1`,
`create sequence buckets_seq increment 1 minvalue 1 maxvalue 9223372036854775807 start 1 cache 1`,
`create table stats(id integer not null default nextval('stats_seq'), system char varying(255) not null, subsystem char varying(255), message char varying(255), context char varying(255), servertime timestamp not null, clienttime char varying(100), error boolean)`,
`create table stats_info(id integer not null default nextval('stats_info_seq'), stats_id integer, info text)`,
`create table buckets(id integer not null default nextval('buckets_seq'), bucket varchar(255), servertime timestamp not null, info char varying(1024))`,
	}
	var firsterr error
	for _, cmd := range cmds {
		printf("running: %v", cmd)
		_, err := db.conn.Exec(cmd)
		if err != nil {
			printf("error: %v", err)
			if firsterr != nil {
				firsterr = err
			}
		}
	}
	printf("done creating tables")
	return firsterr
}

func (db *StatsDB) Done() {
	db.done <- struct{}{}
}

func (db *StatsDB) serve() {
	defer db.conn.Close()
LOOP:
	for {
		select {
		case stat := <-db.newStat:
			err := db.Push(&stat)
			if err != nil {
				log.Printf("error pushing to database %v", err)
			}
		case <-db.done:
			break LOOP
		}
	}
}

func MakeStatsStream(db *StatsDB) websocket.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()
		reader := bufio.NewReader(conn)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			printf("error reading line: %v", err)
			return
		}
		lastId, err := strconv.ParseInt(string(bytes.Trim(line, " \r\n\t")), 10, 32)
		if err != nil {
			printf("error decoding lastid using default. %v", err)
			return
		} else {
			printf("starting stream to: %v at id: %v", conn.Request().RemoteAddr, lastId)
		}
		enc := json.NewEncoder(conn)
		var backtime int
		newData := false
		for {
			data, err := db.FetchAfterId(int(lastId), 100)
			if err != nil {
				printf("error reading data from database: %v", err)
				return
			}
			for v := range data {
				lastId = int64(v.Id)
				newData = true
				err = enc.Encode(&v)
				if err != nil {
					printf("error encoding data to client: %v", err)
					return
				}
				fmt.Fprintf(conn, "\r\n")
			}
			if !newData {
				backtime = backtime + 10
				if backtime > 300 {
					backtime = 300
				}
				printf("no more data to stream, wait a few seconds")
				<-time.After(time.Minute + (time.Second * time.Duration(backtime)))
			} else {
				backtime = 0
			}
			newData = false
		}
	})
}

type StatsHandler struct {
	db *StatsDB
}

func NewStatsHandler(db *StatsDB) *StatsHandler {
	return &StatsHandler{
		db: db,
	}
}

func (sh *StatsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	printf("[%v] %v", req.Method, req.URL)
	if req.Method == "POST" {
		sh.handlePost(w, req)
	} else if req.Method == "GET" {
		sh.handleGet(w, req)
	}
}

func (sh *StatsHandler) handlePost(w http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, "/new") {
		dec := json.NewDecoder(req.Body)
		var stats Stats
		if err := dec.Decode(&stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			sh.db.newStat <- stats
		}
	} else {
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (sh *StatsHandler) handleGet(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	size, err := strconv.ParseInt(req.Form.Get("size"), 10, 32)
	if err != nil || size <= 0 {
		size = DefaultSize
	}
	data, err := sh.db.Fetch(int(size))
	if err != nil {
		http.Error(w, "error reading data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "[")
	enc := json.NewEncoder(w)
	first := true
	for s := range data {
		if !first { fmt.Fprintf(w, ",") }
		err = enc.Encode(&s)
		if err != nil {
			printf("error encoding json to client %v", err)
		}
		first = false
	}
	fmt.Fprintf(w, "]")
}

func printf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func fatalf(msg string, code int, args ...interface{}) {
	if exitStatus == 0 {
		exitStatus = code
	}
	if len(msg) > 0 {
		log.Printf(msg, args...)
	}
}

func abort() bool {
	return exitStatus != 0
}

func setupDatabase() (*StatsDB, error) {
	if statsdb, err := NewStatsDB(*dbuser, *dbpasswd, *dbhost, *dbname); err != nil {
		fatalf("error connecting to database. %v", 1, err)
		return nil, err
	} else {
		if err := statsdb.CreateTables(); err != nil {
			fatalf("error creating tables. %v", 1, err)
			return statsdb, err
		}
		return statsdb, nil
	}
}

func setupHttp() error {
	if statsdb, err := NewStatsDB(*dbuser, *dbpasswd, *dbhost, *dbname); err != nil {
		return err
	} else {
		handler := NewStatsHandler(statsdb)
		http.Handle("/stats/", handler)
		http.Handle("/stats/stream", MakeStatsStream(statsdb))
		return nil
	}
	panic("not reached")
	return nil
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		fatalf("", 1)
	}

	if abort() {
		return
	}

	if *initdb {
		db, err := setupDatabase()
		if err != nil {
			fatalf("error setting up the database. %v", 1, err)
		}
		defer db.Done()
	} else {
		printf("starting server at: %v", *httpaddr)
		if err := setupHttp(); err != nil {
			fatalf("error setting up http server. %v", 1, err)
		}
		if !abort() {
			if err := http.ListenAndServe(*httpaddr, nil); err != nil {
				fatalf("error starting http server. %v", 1, err)
			}
		}
	}

	defer os.Exit(exitStatus)
}
