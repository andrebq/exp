package httpfs

import (
	"os"
	"strings"
	"testing"
)

func TestDiskFile(t *testing.T) {
	df, err := NewDiskFile(".") // should be a directory
	if err != nil {
		t.Fatalf("should always be capable of reading . %v", err)
	}

	childs, err := df.Childs()
	if err != nil {
		t.Fatalf("error reading childs of . %v", err)
	}

	for _, v := range childs {
		if strings.HasSuffix(v, ".go") {
			// a file and not a dir
			file, err := df.Open(v)

			if err != nil {
				t.Errorf("error opening %v: %v", v, err)
			}

			if reader, err := file.Reader(); err != nil {
				t.Errorf("error opening the file for reading: %v", err)
			} else {
				reader.Close()
			}

			if writer, err := file.Writer(); err != nil {
				t.Errorf("error opening the file for writing: %v", err)
			} else {
				writer.Close()
			}

			// check if walk is working
			_, err = Walk(df, v)
			if err != nil {
				t.Errorf("error on walk: %v", err)
			}
		}
	}
}

func TestOpenOrCreate(t *testing.T) {
	df, err := NewDiskFile(".") // should be a directory
	if err != nil {
		t.Fatalf("should always be capable of reading . %v", err)
	}

	os.Remove("newfile.txt")
	newfile, err := OpenOrCreate(df, true, "newfile.txt")
	if err != nil {
		t.Fatalf("error creating newfile.txt")
	}

	if newfile == nil {
		t.Fatalf("nefile is nil...")
	}
	os.Remove("newfile.txt")

	// now let's create a deep path
	os.RemoveAll("a/lot/of/sub/dirs.txt")
	newfile, err = OpenOrCreate(df, true, "a/lot/of/sub/dirs.txt")
	if err != nil {
		t.Fatalf("error creating sub-dirs: %v", err)
	}
	if newfile == nil {
		t.Fatalf("newfile is nil")
	}
	os.RemoveAll("a/lot/of/sub/dirs.txt")
}
