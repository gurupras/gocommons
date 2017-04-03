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

type FileType int

const (
	GZ_TRUE    FileType = 1
	GZ_FALSE   FileType = 0
	GZ_UNKNOWN FileType = -1
)

const (
	DEFAULT_BUFSIZE = 128 * 1024 * 1024
)

type File struct {
	Path string
	File *os.File
	mode int
	gz   FileType
}

type IWriter interface {
	Write(bytes []byte) (int, error)
	Reset(w io.Writer)
	Flush() error
	//Close() error
}

type Writer struct {
	IWriter
	gz FileType
}

func (f *File) fixMode() {
	// First, the simple case
	if strings.HasSuffix(f.Path, ".gz") {
		f.gz = GZ_TRUE
	} else {
		// Remember, all of this only occurs when gz is set to GZ_UNKNOWN
		// So if a file is in write mode, has a non .gz suffix and is
		// set to GZ_UNKNOWN, we're obviously going to give back a regular
		// non-gz file
		f.gz = GZ_FALSE

		// Try to get a reader to figure it out
		if f.mode|os.O_RDONLY|os.O_RDWR != 0 {
			// We have read privilege..try to get a gzip reader
			reader, err := gzip.NewReader(f.File)
			if err == nil {
				f.gz = GZ_TRUE
				defer reader.Close()
			} else {
				f.gz = GZ_FALSE
			}
		}
		// We can freely seek at this point
		// This occurs on Open at which point the user is just
		// opening the file and cannot do any operation on it.
		// So, we can seek back and return as Open always does
		// - at the start of the file
		f.File.Seek(0, os.SEEK_SET)
	}
}

func (w *Writer) Flush() (err error) {
	if w.gz == GZ_TRUE {
		v, _ := w.IWriter.(*gzip.Writer)
		err = v.Flush()
	} else {
		v, _ := w.IWriter.(*bufio.Writer)
		err = v.Flush()
	}
	return
}

func (w *Writer) Close() (err error) {
	if w.gz == GZ_TRUE {
		v, _ := w.IWriter.(*gzip.Writer)
		err = v.Close()
	}
	return
}

func (f *File) Error() string {
	return fmt.Sprintf("Error: (%v, %v, %v)", f.Path, f.mode, f.gz)
}

func (f *File) RawReader() (io.Reader, error) {
	gz_open := false
	var reader io.Reader
	var err error

	switch f.gz {
	case GZ_TRUE:
		gz_open = true
	case GZ_FALSE:
		// Nothing to do
	case GZ_UNKNOWN:
		panic("Should not have occured..mode should have been fixed on open")
	}

	if gz_open == true {
		reader, err = gzip.NewReader(f.File)
	} else {
		reader = bufio.NewReader(f.File)
	}
	return reader, err
}

func (f *File) Reader(bufsize int) (*bufio.Scanner, error) {
	reader, err := f.RawReader()
	if err != nil {
		return nil, err
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
		panic("Should not have occured..mode should have been fixed on open")
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

func Open(filepath string, mode int, gz FileType) (*File, error) {
	var retfile *File
	var err error

	file, err := os.OpenFile(filepath, mode, 0664)
	if err == nil {
		retfile = &File{filepath, file, mode, gz}
		if gz == GZ_UNKNOWN {
			retfile.fixMode()
		}
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
