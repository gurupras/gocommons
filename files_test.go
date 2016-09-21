package gocommons

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CheckFileContentsMatch(f *File, contents string, expected bool) (bool, error) {
	var err error
	var match bool
	scanner, err := f.Reader(0)
	if err != nil {
		return !expected, err
	}
	s := ""

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		s += scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return !expected, err
	}

	if s != contents {
		match = false
	} else {
		match = true
	}
	if match != expected {
		return false, err
	} else {
		return true, err
	}
}

func TestOpenGzFalse(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var success bool = false
	var err error
	var f *File

	// Should pass
	f, err = Open("test_files/open-test.txt", os.O_RDONLY, GZ_FALSE)
	assert.Nil(err, "Failed to open valid file", err)

	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}

	// Should succeed
	f, err = Open("test_files/open-test.gz", os.O_RDONLY, GZ_FALSE)
	assert.Nil(err, "Failed to open valid file", err)
	success, err = CheckFileContentsMatch(f, "Hello World", false)
	if err == nil && !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}
}

func TestWriteGz(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var success bool
	var err error
	var f *File
	var writer Writer

	f, err = Open("test_files/write-gz.gz", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, GZ_TRUE)
	assert.Nil(err, "Failed to open valid file", err)
	defer f.Close()

	writer, err = f.Writer(0)
	assert.Nil(err, "Failed to get writer to file", err)

	writer.Write([]byte("Hello World"))
	writer.Flush()
	writer.Close()
	f.Close()

	f, err = Open("test_files/write-gz.gz", os.O_RDONLY, GZ_TRUE)
	assert.Nil(err, "Failed to open valid file", err)

	if success, err = CheckFileContentsMatch(f, "Hello World", true); err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}
}

func TestFlush(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var success bool
	var err error
	var f *File

	// Should pass
	f, err = Open("/tmp/normal.gz", os.O_CREATE|os.O_TRUNC|os.O_RDWR, GZ_TRUE)
	assert.Nil(err, "Failed to open valid file")

	writer, err := f.Writer(0)
	assert.Nil(err, "Failed to open valid file")
	writer.Write([]byte("stuff"))
	writer.Flush()
	writer.Close()

	f.Seek(0, 0)
	if success, err = CheckFileContentsMatch(f, "stuff", true); err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}
	os.Remove(f.Path)

	// Now do it for a normal file
	f, err = Open("/tmp/normal.txt", os.O_CREATE|os.O_TRUNC|os.O_RDWR, GZ_FALSE)
	assert.Nil(err, "Failed to open valid file")

	writer, err = f.Writer(0)
	assert.Nil(err, "Failed to open valid file")
	writer.Write([]byte("stuff"))
	writer.Flush()
	writer.Close()

	f.Seek(0, 0)
	if success, err = CheckFileContentsMatch(f, "stuff", true); err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}
	os.Remove(f.Path)

	// Now unknown .gz
	f, err = Open("/tmp/unknown.gz", os.O_CREATE|os.O_TRUNC|os.O_RDWR, GZ_UNKNOWN)
	assert.Nil(err, "Failed to open valid file")

	writer, err = f.Writer(0)
	assert.Nil(err, "Failed to open valid file")
	writer.Write([]byte("stuff"))
	writer.Flush()
	writer.Close()

	f.Seek(0, 0)
	if success, err = CheckFileContentsMatch(f, "stuff", true); err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}
	os.Remove(f.Path)

	// Now unknown non-gz
	f, err = Open("/tmp/unknown.txt", os.O_CREATE|os.O_TRUNC|os.O_RDWR, GZ_UNKNOWN)
	assert.Nil(err, "Failed to open valid file")

	writer, err = f.Writer(0)
	assert.Nil(err, "Failed to open valid file")
	writer.Write([]byte("stuff"))
	writer.Flush()
	writer.Close()

	f.Seek(0, 0)
	if success, err = CheckFileContentsMatch(f, "stuff", true); err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}
	os.Remove(f.Path)
}

func TestOpenGzTrue(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var success bool
	var err error
	var f *File

	// Should pass
	f, err = Open("test_files/open-test.gz", os.O_RDONLY, GZ_TRUE)
	assert.Nil(err, "Failed to open valid file")

	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}

	// Should succeed
	f, err = Open("test_files/open-test.txt", os.O_RDONLY, GZ_TRUE)
	assert.Nil(err, "Failed to open valid file")

	success, err = CheckFileContentsMatch(f, "Hello World", false)
	if err == nil && !success {
		assert.Fail("Should have failed to verify file contents")
	}
}

func TestOpenGzUnknown(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var success bool
	var err error
	var f *File

	// Should pass
	f, err = Open("test_files/open-test.gz", os.O_RDONLY, GZ_UNKNOWN)
	assert.Nil(err, "Failed to open valid file", err)

	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}

	// Should pass
	f, err = Open("test_files/open-test.txt", os.O_RDONLY, GZ_UNKNOWN)
	assert.Nil(err, "Failed to open valid file", err)

	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}

	// Should pass
	f, err = Open("test_files/open-test.gz", os.O_RDONLY, GZ_UNKNOWN)
	assert.Nil(err, "Failed to open valid file", err)

	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || !success {
		assert.Fail(fmt.Sprintf("Failed to verify file contents: %v", err))
	}

	// Should fail
	f, err = Open("test_files/open-test.fake.gz", os.O_RDONLY, GZ_UNKNOWN)
	assert.Nil(err, "Failed to open valid file", err)

	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err == nil || success {
		assert.Fail(fmt.Sprintf("Should have failed to verify file contents: %v", err))
	}

}

func TestListFiles(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	answer_txt := []string{"a.txt", "b.txt"}
	answer_txt_out := []string{"a.txt.out.1", "c.txt.out.2"}
	answer_gz := []string{"a.gz", "a.sorted.gz", "c.gz"}
	answer_combined := []string{"a.txt", "a.txt.out.1", "b.txt", "c.txt.out.2"}

	patterns := [][]string{[]string{"*.txt"}, []string{"*.txt.out.*"}, []string{"*.gz"}, []string{"*.txt", "*.txt.out.*"}}
	answers := [][]string{answer_txt, answer_txt_out, answer_gz, answer_combined}

	for i := range patterns {
		p := patterns[i]
		answer := answers[i]
		files, _ := ListFiles("test_files/list_files", p)
		trimmed := make([]string, len(files))
		for idx, v := range files {
			trimmed[idx] = path.Base(v)
		}
		//		fmt.Println("files:   %v", files)
		//		fmt.Println("trimmed: %v", trimmed)
		//		for idx, v := range trimmed {
		//			fmt.Println("trimmed[%v] = %v", idx, v)
		//		}

		assert.True(reflect.DeepEqual(trimmed, answer), fmt.Sprintf("Expected: %v\nGot: %v\n", answer, trimmed))
	}
}

func TestListDirs(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	answer_nr := []string{"1", "2", "3"}
	answer_r := []string{"1", "11", "111", "2", "21", "3", "31"}

	patterns := []string{"*/", "**/"}
	answers := [][]string{answer_nr, answer_r}

	for i := range patterns {
		p := patterns[i]
		answer := answers[i]
		files, _ := ListDirs("./test_files/testdir", []string{p})
		trimmed := make([]string, len(files))
		for idx, v := range files {
			trimmed[idx] = path.Base(v)
		}
		//		fmt.Println("files:   %v", files)
		//		fmt.Println("trimmed: %v", trimmed)
		//		for idx, v := range trimmed {
		//			fmt.Println("trimmed[%v] = %v", idx, v)
		//		}

		assert.Equal(answer, trimmed, "Did not match")
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Test success
	var exists bool
	var err error
	exists, err = Exists("./test_files")
	assert.Equal(nil, err, "Failed to check exists on existing directory")
	assert.Equal(true, exists, "Exists failed on existing directory")

	exists, err = Exists("./doesnotexist")
	assert.Equal(nil, err, "Failed to check exists on non-existing directory")
	assert.Equal(false, exists, "Exists failed on non-existing directory")
}
