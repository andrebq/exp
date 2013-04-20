package main

import (
	"code.google.com/p/goplan9/plan9"
	"net"
	"flag"
	"fmt"
	"io"
	"log"
	"sync"
)

var (
	listen = flag.String("listen", "127.0.0.1:5640", "Address to listen")
	help = flag.Bool("h", false, "Help")
)

type Server struct {
	listener net.Listener
	explorer FileExplorer
}

func NewServer(lnet, laddr string, explorer FileExplorer) (*Server, error) {
	s := &Server{}
	l, err := net.Listen(lnet, laddr)
	if err != nil { return nil, err }
	s.listener = l
	s.explorer = explorer
	return s, nil
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) Start() error {
	for {
		client, err := s.listener.Accept()
		println("got client")
		if err != nil {
			println("error", err)
			return err
		}
		cc := NewClientConn(client, s, s.explorer)
		go cc.handle()
	}
	return nil
}

func (s *Server) handleError(err error, c net.Conn) {
	log.Printf("Client %v caused error %v", c.RemoteAddr(), err)
}

func readFCall(out chan *plan9.Fcall, done chan signal, err chan error, input io.Reader) {
loop:
	for {
		select {
		case <-done:
			close(out)
			break loop
		default:
			fc, e := plan9.ReadFcall(input)
			if e != nil {
				err <- e
				close(out)
				break
			}
			out <- fc
		}
	}
}

func writeFCall(out io.Writer, done chan signal, err chan error, input chan *plan9.Fcall) {
loop:
	for {
		select {
		case <-done:
			break loop
		case fc := <-input:
			if fc == nil {
				continue
			}
			e := plan9.WriteFcall(out, fc)
			if e != nil {
				err <- e
				break
			}
		}
	}
}

// Handle all client interaction
type ClientConn struct {
	sync.RWMutex
	// remote connection
	conn net.Conn
	// server
	server *Server
	// map between path's and fileRef's
	fileRefs map[uint64]*fileRef
	// map between a fid and a qid
	fidmap map[uint32]uint64
	// explorer used to navigate the FS
	explorer FileExplorer
}

// Interface used to navigate, open and close files in the system
type FileExplorer interface {
	// Must return the unique identifier for this explorer root.
	//
	// The returned value can't be zero, since 0 is considered a non-existing path 
	Root() uint64
	// Should return the unique identifier and type of the given name under the given directory
	//
	// If the name don't exists under parent, should return 0
	Walk(parent uint64, name string) (uint64, FileType)
	// Open the file for subsequent reading/writing
	//
	// If the error is nil, the file is considered ready for processing
	Open(file uint64, mode FileMode) error
}

// Represent a reference to a file
type fileRef struct {
	plan9.Qid
}

// Type of a file
type FileType uint8

const (
	FTFILE = FileType(plan9.QTFILE)
	FTDIR = FileType(plan9.QTDIR)
	FTMOUNT = FileType(plan9.QTMOUNT)
)

type FileMode uint8
const (
	FMREAD = FileMode(plan9.OREAD)
	FMWRITE = FileMode(plan9.OWRITE)
	FMRDWR = FileMode(plan9.ORDWR)
)

func NewClientConn(conn net.Conn, server *Server, explorer FileExplorer) *ClientConn {
	cc := &ClientConn{conn: conn, server: server, fileRefs: make(map[uint64]*fileRef),
		explorer: explorer, fidmap: make(map[uint32]uint64)};
	return cc
}

// used to send signal's between goroutines
type signal struct{}

func (c *ClientConn) handle() {
	defer c.Close()
	input := make(chan *plan9.Fcall)
	output := make(chan *plan9.Fcall)
	done := make(chan signal)
	err := make(chan error)

	go readFCall(input, done, err, c.conn)
	go writeFCall(c.conn, done, err, output)
loop:
	for {
		select {
		case fc := <-input:
			go c.process(fc, output)
		case e := <-err:
			c.server.handleError(e, c.conn)
			done <- signal{}
			done <- signal{}
			break loop
		}
	}
}

func (c *ClientConn) process(fc *plan9.Fcall, out chan *plan9.Fcall) {
	println(">>>\t", fc.String())
	switch fc.Type {
	case plan9.Tversion:
		fc = c.version(fc)
	case plan9.Tattach:
		fc = c.attach(fc)
	case plan9.Twalk:
		fc = c.walk(fc)
	case plan9.Topen:
		fc = c.open(fc)
	default:
		println("!!!\t", fc.String())
		fc = nil
	}
	if fc != nil {
		println("<<<\t", fc.String())
	}
	out <- fc
}

func (c *ClientConn) version(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rversion
	fc.Version = "9P2000"
	return fc
}

func (c *ClientConn) attach(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rattach
	fref, _ := c.createFileRef(c.explorer.Root(), FTMOUNT, 0)
	fc.Qid = fref.Qid
	c.bindFid(fc.Fid, fc.Qid.Path)
	return fc
}

func (c *ClientConn) walk(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rwalk
	fref, has := c.fidRef(fc.Fid)
	if !has {
		return c.invalidFidErr(fc)
	}
	if _, has := c.fidRef(fc.Newfid); has {
		return c.fidUsedErr(fc)
	}
	current := fref.Path
	for idx, name := range fc.Wname {
		var ft FileType
		current, ft = c.explorer.Walk(current, name)
		if current == 0 {
			return c.fileNotFoundErr(fc)
		}
		ref, err := c.createFileRef(current, ft, 0)
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		fc.Wqid = append(fc.Wqid, ref.Qid)

		// if the last match isn't a directory, there is no need to find
		// another part of the path
		//
		// so, just break here
		if ft == FTFILE && idx != len(fc.Wname) - 1 {
			println("here")
			return c.fileNotFoundErr(fc)
		}
	}
	// make a bind between the last qid and the new fid
	c.bindFid(fc.Newfid, fc.Wqid[len(fc.Wqid)-1].Path)
	return fc
}

func (c *ClientConn) open(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Ropen
	if fref, has := c.fidRef(fc.Fid); has {
		err := c.explorer.Open(fref.Path, FileMode(fc.Mode))
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		fc.Qid = fref.Qid
		return fc
	}
	return c.invalidFidErr(fc)
}

func (c *ClientConn) invalidFidErr(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rerror
	fc.Ename = "fid not found"
	return fc
}

func (c *ClientConn) fileNotFoundErr(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rerror
	fc.Ename = "file not found"
	return fc
}

func (c *ClientConn) fidUsedErr(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rerror
	fc.Ename = "fid in use"
	return fc
}

func (c *ClientConn) unexpectedErr(fc *plan9.Fcall, err error) *plan9.Fcall {
	fc.Type = plan9.Rerror
	fc.Ename = err.Error()
	return fc
}

// Return the file referenced by the given fid
func (c *ClientConn) fidRef(fid uint32) (*fileRef, bool) {
	c.RLock()
	defer c.RUnlock()
	if qid, has := c.fidmap[fid]; has {
		fref, has := c.fileRefs[qid]
		return fref, has
	}
	return nil, false
}

// Bind the given fid to the provided path
//
// if path is 0, remove the fid
func (c *ClientConn) bindFid(fid uint32, path uint64) {
	c.Lock()
	defer c.Unlock()
	_, has := c.fidmap[fid]
	if path == 0 && has {
		delete(c.fidmap, fid)
		return
	}
	c.fidmap[fid] = path
}

// Create a file for the given path, if the path is already present, then just check 
// if the version/type are the same, if they are, just return the existing file, otherwise
// return an error
//
// If the version is 0, then this check is ignored and any version existing on the server is returned
//
// Every new resources created will have a version of 1 instead of 0.
func (c *ClientConn) createFileRef(path uint64, ft FileType, version uint32) (*fileRef, error) {
	c.Lock()
	defer c.Unlock()
	if fref, has := c.fileRefs[path]; has {
		if fref.Path == path && (fref.Vers == version || version == 0) && uint8(ft) == fref.Type {
			return fref, nil
		}
		return fref, fmt.Errorf("path %v already used", path)
	}
	fref := &fileRef{}
	if version == 0 {
		version = 1
	}
	fref.Type = uint8(ft)
	fref.Vers = version
	fref.Path = path
	c.fileRefs[path] = fref
	return fref, nil
}

func (c *ClientConn) Close() error {
	return c.conn.Close()
}

type dummyExplorer struct{}

func (d dummyExplorer) Root() uint64 {
	return 1
}

func (d dummyExplorer) Walk(parent uint64, name string) (uint64, FileType) {
	println("\t...searching ", parent, " for file: ", name)
	if parent == 1 && name == "dummy" {
		return 2, FTFILE
	}
	return 0, FTFILE
}

func (d dummyExplorer) Open(file uint64, mode FileMode) error {
	if file != 2 {
		return fmt.Errorf("file not found")
	}
	if mode != FMREAD {
		return fmt.Errorf("file is read only")
	}
	return nil
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	server,err := NewServer("tcp", *listen, dummyExplorer{})
	if err != nil {
		log.Fatalf("Unable to create server. Cause: %v", err)
	}
	defer server.Close()
	err = server.Start()
	if err != nil {
		log.Fatalf("Unable to start server. Cause: %v", err)
	}
}
