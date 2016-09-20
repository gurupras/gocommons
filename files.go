package gocommons

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar"
)

const (
	GZ_TRUE    = 1
	GZ_FALSE   = 0
	GZ_UNKNOWN = -1
)

const (
	DEFAULT_BUFSIZE = 128 * 1024 * 1024
)

type File struct {
	Path string
	File *os.File
	mode int
	gz   int
}

type IWriter interface {
	Write(bytes []byte) (int, error)
	Reset(w io.Writer)
	Flush() error
	//Close() error
}

type Writer struct {
	writer IWriter
	gz     int
}

func (w *Writer) fix_mode() {
	if w.gz == GZ_UNKNOWN {
		if _, ok := w.writer.(*gzip.Writer); !ok {
			w.gz = GZ_FALSE
		} else {
			w.gz = GZ_TRUE
		}
	}
}
func (w *Writer) Write(bytes []byte) (int, error) {
	return w.writer.Write(bytes)
}

func (w *Writer) Flush() (err error) {
	w.fix_mode()
	if w.gz == GZ_TRUE {
		if v, ok := w.writer.(*gzip.Writer); ok {
			err = v.Flush()
		} else {
			err = errors.New("Could not find underlying Writer type to flush")
		}
	} else {
		if v, ok := w.writer.(*bufio.Writer); ok {
			err = v.Flush()
		} else {
			err = errors.New("Could not find underlying Writer type to flush")
		}
	}
	return
}

func (w *Writer) Close() (err error) {
	w.fix_mode()
	if w.gz == GZ_TRUE {
		if v, ok := w.writer.(*gzip.Writer); ok {
			err = v.Close()
		} else {
			err = errors.New("Could not find underlying Writer type to close")
		}
	} else {
		if _, ok := w.writer.(*bufio.Writer); ok {
		} else {
			err = errors.New("Could not find underlying Writer type to close")
		}
	}
	return
}

func (f *File) Error() string {
	return fmt.Sprintf("Error: (%v, %v, %v)", f.Path, f.mode, f.gz)
}

func (f *File) Reader(bufsize int) (*bufio.Scanner, error) {
	gz_open := false
	var reader io.Reader
	var err error

	if f.gz == GZ_TRUE {
		gz_open = true
	} else if f.gz == GZ_FALSE {
		// Nothing to do
	} else {
		if strings.HasSuffix(f.Path, ".gz") {
			gz_open = true
		} else {
		}
	}

	if gz_open == true {
		reader, err = gzip.NewReader(f.File)
	} else {
		reader = bufio.NewReader(f.File)
	}

	scanner := bufio.NewScanner(reader)

	var buf []byte
	if bufsize != 0 {
		buf = make([]byte, 0, bufsize)
	} else {
		buf = make([]byte, 0, 1048576)
	}

	scanner.Buffer(buf, bufsize)
	return scanner, err
}

func (f *File) AsyncReadWithBufsize(splitFunction bufio.SplitFunc, bufsize int, channel chan string) {
	defer close(channel)
	reader, err := f.Reader(bufsize)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get reader")
		return
	}

	reader.Split(splitFunction)
	for reader.Scan() {
		line := reader.Text()
		channel <- line
	}
}

func (f *File) AsyncRead(splitFunction bufio.SplitFunc, channel chan string) {
	f.AsyncReadWithBufsize(splitFunction, 1048576, channel)
}

func (f *File) Writer(bufsize int) (Writer, error) {
	gz_open := false
	var writer IWriter
	var err error

	switch f.gz {
	case GZ_TRUE:
		gz_open = true
	case GZ_FALSE:
		// Nothing to do
	default:
		if strings.HasSuffix(f.Path, ".gz") {
			gz_open = true
			f.gz = GZ_TRUE
		} else {
			f.gz = GZ_FALSE
		}
	}

	if gz_open == true {
		writer = gzip.NewWriter(f.File)
	} else {
		writer = bufio.NewWriter(f.File)
	}

	if bufsize != 0 {
		writer = bufio.NewWriterSize(writer, bufsize)
	}
	return Writer{writer, f.gz}, err
}

func (f *File) Close() {
	f.File.Close()
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.File.Seek(offset, whence)
}

func Open(filepath string, mode int, gz int) (*File, error) {
	var retfile *File
	var err error

	file, err := os.OpenFile(filepath, mode, 0664)
	if err == nil {
		retfile = &File{filepath, file, mode, gz}
	}
	return retfile, err
}

func ListFiles(fpath string, patterns []string) (matches []string, err error) {
	_, err = os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	visit := func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		var matched bool
		for _, pattern := range patterns {
			var m bool
			m, err = filepath.Match(pattern, fi.Name())
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
			matched = m || matched
		}
		if matched {
			matches = append(matches, fp)
		}
		return nil
	}
	filepath.Walk(fpath, visit)
	sort.Strings(matches)
	return
}

func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	return fileInfo.IsDir(), err
}

func ListDirs(fpath string, patterns []string) (matches []string, err error) {
	var dirs []string
	abs, _ := filepath.Abs(fpath)
	for _, pattern := range patterns {
		globPattern := abs + "/" + pattern
		if dirs, err = doublestar.Glob(globPattern); err != nil {
			err = errors.New(fmt.Sprintf("Failed to glob: %v", err))
			return
		}
		for _, d := range dirs {
			if ok, _ := IsDir(d); ok {
				matches = append(matches, d)
			}
		}
	}
	sort.Strings(matches)
	return
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func Makedirs(path string) error {
	exist, err := Exists(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	if !exist {
		return os.MkdirAll(path, 0775)
	}
	return nil
}
