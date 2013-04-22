package main

import (
	"bytes"
	"code.google.com/p/goplan9/plan9"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"sync"
)

var (
	listen = flag.String("listen", "127.0.0.1:5640", "Address to listen")
	help   = flag.Bool("h", false, "Help")
)

type Server struct {
	listener net.Listener
	explorer FileExplorer
}

func NewServer(lnet, laddr string, explorer FileExplorer) (*Server, error) {
	s := &Server{}
	l, err := net.Listen(lnet, laddr)
	if err != nil {
		return nil, err
	}
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
	// atomic unit of a message
	iounit uint32
}

// Information about a given file
type Stat struct {
	File
	// Permissions (unix like)
	Mode uint32
	// Time of the last access
	Atime uint32
	// Time of the last modification
	Mtime uint32
	// Name of the user
	Uname string
	// Name of the group
	Gname string
	// Size of the file
	//
	// 0 means undefined
	// for directories, should return the number of child nodes
	// for files, the number of bytes in the file
	Size uint64
}

// Represent the path, version and name of any given location in filesystem
type File struct {
	Path    uint64
	Name    string
	Version uint32
	Type    FileType
}

// Interface used to navigate, open and close files in the system
type FileExplorer interface {
	// Must return the unique identifier for this explorer root.
	//
	// The returned value can't be zero, since 0 is considered a non-existing path
	Root() uint64
	// Open the file for subsequent reading/writing
	//
	// If the error is nil, the file is considered ready for processing
	Open(file uint64, mode FileMode) error
	// Read at most len(buf) bytes from the given file starting at the location pointed by
	// offset
	//
	// Should return the number of bytes copied, returning a non null error, don't send any data to the client
	//
	// The returned byte count will be converted to uint32
	Read(buf []byte, offset uint64, path uint64) (int, error)
	// Return the list of ID's for the given path. path will always point to a directory
	ListDir(path uint64) (FileList, error)
	// Return the information about the given file or directory
	Stat(path uint64) (*Stat, error)
	// Return the size of the given file, returning 0, means a file that don't have a
	// finite size (reading from input devices, like keyboard or serial port)
	//
	// When the returned value is 0, the Read method is responsible for sending the io.EOF to signal
	// the end of file, otherwise, the ClientConn will handle the EOF
	Sizeof(path uint64) (uint64, error)
	// Close the associated file
	Close(path uint64) error
	// Create the given file under the path with name and permission
	//
	// Should return the path of the file
	Create(parent uint64, info File, perm uint32) (uint64, error)
	// Write the buf to the given file starting at the offset position
	//
	// Should return the number of bytes actually written or an error.
	//
	// Returning less bytes than buf will not cause an error to the server, but the client might get a little confused,
	// it's recommended that if the write can't be completed, this call should block or return an error.
	Write(dest uint64, buf []byte, offset uint64) (int, error)
}

// Utility to sort files from a directory
type FileList []File

func (f FileList) Len() int {
	return len(f)
}
func (f FileList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
func (f FileList) Less(i, j int) bool {
	return f[i].Name < f[j].Name
}
func (f FileList) FindExact(name string) (int, bool) {
	idx := sort.Search(len(f), func(i int) bool {
		return f[i].Name >= name
	})
	if idx < len(f) && f[idx].Name == name {
		return idx, true
	}
	return idx, false
}

// This is a complement to FileExplorer, to include some search facilities
//
// The methods listed here aren't required to implement a full fileserver, but they might be useful if a directory have
// lot's of child node.
type FileFinder interface {
	FileExplorer
	// This method should return the File structure for the given name under path.
	//
	// If there is no file with that name, should simply return nil.
	//
	// Returning an error here will result in a Rerror message sent to the client
	FindInDir(path uint64, name string) (*File, error)
}

// Represent a reference to a file
type fileRef struct {
	plan9.Qid
}

func (fr *fileRef) IsDir() bool {
	return fr.Type == uint8(FTDIR)
}

// Type of a file
type FileType uint8

const (
	FTFILE  = FileType(plan9.QTFILE)
	FTDIR   = FileType(plan9.QTDIR)
	FTMOUNT = FileType(plan9.QTMOUNT)
)

type FileMode uint8

const (
	FMREAD  = FileMode(plan9.OREAD)
	FMWRITE = FileMode(plan9.OWRITE)
	FMRDWR  = FileMode(plan9.ORDWR)
)

func NewClientConn(conn net.Conn, server *Server, explorer FileExplorer) *ClientConn {
	cc := &ClientConn{conn: conn, server: server, fileRefs: make(map[uint64]*fileRef),
		explorer: explorer, fidmap: make(map[uint32]uint64), iounit: 0}
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
	case plan9.Tread:
		fc = c.read(fc)
	case plan9.Tclunk:
		fc = c.clunk(fc)
	case plan9.Tcreate:
		fc = c.create(fc)
	case plan9.Twrite:
		fc = c.write(fc)
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
	fc.Iounit = c.iounit
	fref, _ := c.createFileRef(c.explorer.Root(), FTDIR, 0)
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
		if current == 0 {
			return c.fileNotFoundErr(fc)
		}
		if ff, ok := c.explorer.(FileFinder); ok {
			f, err := ff.FindInDir(current, name)
			if err != nil {
				return c.unexpectedErr(fc, err)
			}
			current = f.Path
		} else {
			childs, err := c.explorer.ListDir(current)
			if err != nil {
				return c.unexpectedErr(fc, err)
			}
			idx, have := childs.FindExact(name)
			if !have {
				return c.fileNotFoundErr(fc)
			}
			current = childs[idx].Path
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
		if ft == FTFILE && idx != len(fc.Wname)-1 {
			println("here")
			return c.fileNotFoundErr(fc)
		}
	}
	if len(fc.Wqid) == 0 {
		// newfid and fid will map to the same file
		if fc.Newfid != fc.Fid {
			c.bindFid(fc.Newfid, fref.Path)
		}
	} else {
		// make a bind between the last qid and the new fid
		c.bindFid(fc.Newfid, fc.Wqid[len(fc.Wqid)-1].Path)
	}
	return fc
}

func (c *ClientConn) open(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Ropen
	if fref, has := c.fidRef(fc.Fid); has {
		err := c.explorer.Open(fref.Path, FileMode(fc.Mode))
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		fc.Iounit = c.iounit
		fc.Qid = fref.Qid
		return fc
	}
	return c.invalidFidErr(fc)
}

func (c *ClientConn) read(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rread
	if fref, has := c.fidRef(fc.Fid); has {
		if fref.IsDir() {
			return c.readdir(fc, fref)
		}
		return c.readfile(fc, fref)
	}
	return c.invalidFidErr(fc)
}

func (c *ClientConn) readdir(fc *plan9.Fcall, ref *fileRef) *plan9.Fcall {
	// if the call have an offset, return 0
	// since all readdir call's will return the full directory
	if fc.Offset > 0 {
		fc.Count = 0
		return fc
	}
	childs, err := c.explorer.ListDir(ref.Path)
	if err != nil {
		return c.unexpectedErr(fc, err)
	}
	tmpBuf := allocBuffer(int(c.iounit))
	out := bytes.NewBuffer(tmpBuf[:0])
	defer discardBuffer(tmpBuf)
	for _, id := range childs {
		stat, err := c.explorer.Stat(id.Path)
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		dir := plan9.Dir{
			Qid:    plan9.Qid{Type: uint8(stat.Type), Vers: stat.Version, Path: id.Path},
			Mode:   plan9.Perm(stat.Mode),
			Atime:  stat.Atime,
			Mtime:  stat.Mtime,
			Length: stat.Size,
			Uid:    stat.Uname,
			Gid:    stat.Gname,
			Name:   stat.Name,
		}
		buf, err := dir.Bytes()
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		_, err = out.Write(buf)
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
	}
	fc.Count = uint32(len(fc.Data))
	fc.Data = out.Bytes()
	return fc
}

func min(values ...int) int {
	sort.Ints(values)
	return values[0]
}

func (c *ClientConn) readfile(fc *plan9.Fcall, ref *fileRef) *plan9.Fcall {
	size, err := c.explorer.Sizeof(ref.Path)
	if err != nil {
		return c.unexpectedErr(fc, err)
	}
	if size > 0 && fc.Offset >= size {
		// trying to reading past the end of file.
		// return count == 0 to signal EOF to client
		fc.Count = 0
	}
	fc.Data = allocBuffer(min(int(c.iounit), int(fc.Count), int(size)))
	defer discardBuffer(fc.Data)
	n, err := c.explorer.Read(fc.Data, fc.Offset, ref.Path)
	if err == io.EOF {
		if n == 0 {
			// returned EOF without reading anything, should return fc.Count = 0
			discardBuffer(fc.Data)
			fc.Data = nil
			err = nil
			return fc
		} else {
			// was able to read som data from the file, should return the count
			// but not the error. The next call to read will trigger the EOF
			err = nil
		}
	}
	if err != nil {
		return c.unexpectedErr(fc, err)
	}
	fc.Count = uint32(n)
	return fc
}

func (c *ClientConn) clunk(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rclunk
	oldpath, had := c.unbindFid(fc.Fid)
	if had {
		err := c.explorer.Close(oldpath)
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
	}
	return fc
}

func (c *ClientConn) create(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rcreate
	if fref, have := c.fidRef(fc.Fid); have {
		file := File{Name: fc.Name}
		file.Type = FileType(fc.Mode)
		path, err := c.explorer.Create(fref.Path, file, uint32(fc.Perm))
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		cref, err := c.createFileRef(path, file.Type, 0)
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		fc.Iounit = c.iounit
		fc.Qid = cref.Qid
		println("fidmap: ", fmt.Sprintf("%v", c.fidmap))
		c.unbindFid(fc.Fid)
		c.bindFid(fc.Fid, cref.Path)
		println("fidmap: ", fmt.Sprintf("%v", c.fidmap))
		return fc
	}
	return c.invalidFidErr(fc)
}

func (c *ClientConn) write(fc *plan9.Fcall) *plan9.Fcall {
	fc.Type = plan9.Rwrite
	if fref, have := c.fidRef(fc.Fid); have {
		n, err := c.explorer.Write(fref.Path, fc.Data[:fc.Count], fc.Offset)
		if err != nil {
			return c.unexpectedErr(fc, err)
		}
		fc.Count = uint32(n)
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

func allocBuffer(sz int) []byte {
	if sz == 0 {
		// defautl buffer size
		sz = 8192
	}
	return make([]byte, sz)
}

func discardBuffer(buf []byte) {
	// do nothing,
	// later, implement a way to reuse this buffer
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

// Forget about the fid
func (c *ClientConn) unbindFid(fid uint32) (uint64, bool) {
	c.Lock()
	defer c.Unlock()
	path, has := c.fidmap[fid]
	if has {
		delete(c.fidmap, fid)
	}
	return path, has
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
	println("bindFid: ", fid, path)
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

type dummyExplorer struct {
	tmpfiles map[string]uint64
	tmpdata  map[uint64][]byte
}

func (d *dummyExplorer) Root() uint64 {
	return 1
}

func (d *dummyExplorer) Open(file uint64, mode FileMode) error {
	if file == 1 {
		// want the root directory
		if mode != FMREAD {
			return fmt.Errorf("file is read only")
		}
		return nil
	}
	if _, have := d.tmpdata[file]; have {
		return nil
	}
	return fmt.Errorf("file not found")
}

func (d *dummyExplorer) Read(buf []byte, offset uint64, path uint64) (int, error) {
	data, have := d.tmpdata[path]
	if !have {
		return 0, fmt.Errorf("file not found")
	}
	if int(offset) > len(data) {
		return 0, io.EOF
	}
	return copy(buf, data), nil
}

func (d *dummyExplorer) Sizeof(path uint64) (uint64, error) {
	if path == 1 {
		// only one file in the root directory
		return 1, nil
	}
	data, have := d.tmpdata[path]
	if !have {
		return 0, fmt.Errorf("file not found")
	}
	return uint64(len(data)), nil
}

func (d *dummyExplorer) ListDir(path uint64) (FileList, error) {
	switch path {
	case 1:
		ret := make(FileList, 0)
		for k, v := range d.tmpfiles {
			ret = append(ret, File{Path: v, Name: k, Type: FTFILE})
		}
		return ret, nil
	}
	return nil, fmt.Errorf("file not found")
}

func (d *dummyExplorer) Stat(path uint64) (*Stat, error) {
	st := &Stat{Mode: 0644}
	switch path {
	case 1:
		st.Mode = 0755
		st.Type = FTDIR
		st.Size = uint64(len(d.tmpfiles))
		st.Name = "/"
	default:
		if data, have := d.tmpdata[path]; have {
			st.Type = FTFILE
			st.Size = uint64(len(data))
			st.Name = d.nameForPath(path)
			return st, nil
		}
		return nil, fmt.Errorf("file not found")
	}
	panic("not reached")
}

func (d *dummyExplorer) Close(path uint64) error {
	return nil
}

func (d *dummyExplorer) Create(parent uint64, file File, perm uint32) (uint64, error) {
	if parent != 1 {
		return 0, fmt.Errorf("cannot create directories")
	}
	if _, exist := d.tmpfiles[file.Name]; exist {
		return 0, fmt.Errorf("cannot create the given file")
	}
	path := d.maxQid() + 1
	d.tmpfiles[file.Name] = path
	d.tmpdata[path] = allocBuffer(0)
	return path, nil
}

func (d *dummyExplorer) Write(path uint64, buf []byte, offset uint64) (int, error) {
	if offset != 0 {
		return 0, fmt.Errorf("can't handle offset at this moment")
	}
	d.tmpdata[path] = append(d.tmpdata[path], buf...)
	return len(buf), nil
}

func (d *dummyExplorer) nameForPath(p uint64) string {
	for k, v := range d.tmpfiles {
		if v == p {
			return k
		}
	}
	panic("path not found")
}
func (d *dummyExplorer) maxQid() uint64 {
	next := uint64(1)
	for k, _ := range d.tmpdata {
		if k > next {
			next = k
		}
	}
	return next
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	de := dummyExplorer{tmpfiles: make(map[string]uint64), tmpdata: make(map[uint64][]byte)}
	de.tmpfiles["dummy"] = 2
	de.tmpdata[2] = []byte("hello world")

	server, err := NewServer("tcp", *listen, &de)
	if err != nil {
		log.Fatalf("Unable to create server. Cause: %v", err)
	}
	defer server.Close()
	err = server.Start()
	if err != nil {
		log.Fatalf("Unable to start server. Cause: %v", err)
	}
}
