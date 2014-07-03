package httpfs

import (
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
