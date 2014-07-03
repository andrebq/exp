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
