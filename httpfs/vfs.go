package httpfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrCannotWrite      = errors.New("cannot write")
	ErrCannotRead       = errors.New("cannot read")
	ErrCannotWriteToDir = errors.New("cannot write to directory")
	ErrCannotTruncate = errors.New("file don't allow truncate")
)

type Info struct {
	Name     string
	Dir      bool
	CanRead  bool
	CanWrite bool
}

// TreeNode is used to represent the hierarchy between files
type TreeNode struct {
	Parent *TreeNode
	Childs []*TreeNode
	File File
}

// Will write the abs path of this node to the given output
func (tn *TreeNode) WriteAbsPath(out io.Writer) (n int, err error) {
	if (tn == nil) {
		return 0, nil
	}
	var total int
	if sz, err := tn.Parent.WriteAbsPath(out); err != nil {
		total += sz
		return total, err
	} else {
		total += sz
	}
	if sz, err := io.WriteString(out, tn.File.Info().Name); err != nil {
		total += sz
		return total, err
	} else {
		total += sz
	}
	if tn.File.Info().Dir {
		if sz, err := io.WriteString(out, "/"); err != nil {
			total += sz
			return total, err
		} else {
			total += sz
		}
	}
	return total, nil
}

type Truncable interface {
	Truncate() error
}

type File interface {
	Info() Info
	Open(name string) (File, error)
	Reader() (io.ReadCloser, error)
	Writer() (io.WriteCloser, error)
	Childs() ([]string, error)
}

// Represent a actual disk file
type DiskFile struct {
	abs   string
	isdir bool
}

func Truncate(in File) error {
	if t, ok := in.(Truncable); ok {
		return t.Truncate()
	}
	return ErrCannotTruncate
}

func WriteToFile(out File, in io.Reader) error {
	if out.Info().Dir {
		return ErrCannotWriteToDir
	} else {
		writer, err := out.Writer()
		if err != nil {
			return err
		}
		defer writer.Close()
		_, err = io.Copy(writer, in)
		return err
	}
	return nil
}

func DeepReadTo(out io.Writer, in File, depth int) error {
	tn, err := BuildTreeNode(in, depth)
	if err != nil {
		return err
	}
	return deepReadNode(out, tn)
}

func deepReadNode(out io.Writer, n *TreeNode) error {
	_, err := n.WriteAbsPath(out)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, "\n")
	if err != nil {
		return err
	}

	for _, c := range n.Childs {
		err := deepReadNode(out, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadFileTo(out io.Writer, in File) error {
	if in.Info().Dir {
		childs, err := in.Childs()
		if err != nil {
			return err
		}
		for _, v := range childs {
			f, err := in.Open(v)
			if err != nil {
				return err
			}
			if f.Info().Dir {
				_, err = fmt.Fprintf(out, "%v/\n", v)
			} else {
				_, err = fmt.Fprintf(out, "%v\n", v)
			}
			if err != nil {
				return err
			}
		}
	} else {
		reader, err := in.Reader()
		if err != nil {
			return err
		}
		defer reader.Close()
		_, err = io.Copy(out, reader)
		return err
	}
	return nil
}

func Walk(root File, path ...string) (File, error) {
	var err error
	for _, v := range path {
		root, err = root.Open(v)
		if err != nil {
			return nil, err
		}
	}
	return root, err
}

func NewDiskFile(filename string) (File, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	return &DiskFile{
		abs:   filename,
		isdir: info.IsDir(),
	}, nil
}

func BuildTreeNode(root File, depth int) (*TreeNode, error) {
	tn := &TreeNode{
		Parent: nil,
		File: root,
		Childs: nil,
	}
	if !root.Info().Dir || depth == 0 {
		return tn, nil
	}

	childs, err := root.Childs()
	if err != nil {
		return tn, err
	}
	for _, childName := range childs {
		childFile, err := root.Open(childName)
		if err != nil {
			return tn, err
		}
		childNode, err := BuildTreeNode(childFile, depth-1)
		if err != nil {
			return tn, err
		}
		childNode.Parent = tn
		tn.Childs = append(tn.Childs, childNode)
	}
	return tn, err
}

func (df *DiskFile) Info() Info {
	return Info{
		Name:     filepath.Base(df.abs),
		Dir:      df.isdir,
		CanRead:  false,
		CanWrite: false,
	}
}

func (df *DiskFile) Open(name string) (File, error) {
	info, err := os.Stat(filepath.Join(df.abs, name))
	if err != nil {
		return nil, err
	}
	return &DiskFile{
		abs:   filepath.Join(df.abs, name),
		isdir: info.IsDir(),
	}, nil
}

func (df *DiskFile) Reader() (io.ReadCloser, error) {
	if df.isdir {
		return nil, ErrCannotRead
	}
	return os.Open(df.abs)
}

func (df *DiskFile) Writer() (io.WriteCloser, error) {
	if df.isdir {
		return nil, ErrCannotWrite
	}
	return os.OpenFile(df.abs, os.O_RDWR, 0644)
}

func (df *DiskFile) Childs() ([]string, error) {
	if df.isdir {
		file, err := os.Open(df.abs)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return file.Readdirnames(-1)
	} else {
		return nil, nil
	}
}

func (df *DiskFile) Truncate() error {
	file, err := os.OpenFile(df.abs, os.O_RDWR | os.O_TRUNC, 0644)
	if file != nil {
		defer file.Close()
	}
	return err
}
