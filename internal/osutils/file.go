package osutils

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
)

func DirExist(folderpath string) bool {
	res := false
	fd, err := os.Open(folderpath)
	if err == nil {
		defer fd.Close()
		fi, err := fd.Stat()
		if err == nil {
			if fi.IsDir() {
				res = true
			}
		}
	}
	return res
}

func FileExist(filepath string) bool {
	res := false
	fd, err := os.Open(filepath)
	if err == nil {
		defer fd.Close()
		fi, err := fd.Stat()
		if err == nil {
			if !fi.IsDir() {
				res = true
			}
		}
	}
	return res
}

func LinesFrom(filepath string) []string {
	fi, err := os.OpenFile(filepath, os.O_RDONLY, 0660)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			return nil
		}
		log.Fatalf("Unable to open file '%s': '%+v'", filepath, err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	res := make([]string, 0)
	reader := bufio.NewReader(fi)
	var line []byte
	var isPrefix bool
	for {
		if line, isPrefix, err = reader.ReadLine(); err != nil {
			log.Printf("Error when reading line '%s' (prefix %+v) of '%s': '%+v'\n", line, isPrefix, filepath, err)
			break
		}
		res = append(res, string(line))
	}

	if err == io.EOF {
		err = nil
	}
	if err != nil {
		panic(err)
	}
	return res
}
