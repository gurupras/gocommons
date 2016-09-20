package gocommons

import (
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

	var success bool = false
	var err error
	var f *File

	result := InitResult("TestOpenGzFalse-1")

	// Should pass
	f, err = Open("test_files/open-test.txt", os.O_RDONLY, GZ_FALSE)
	if err != nil {
		success = false
		goto out
	}
	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || !success {
		goto out
	}
	HandleResult(t, success, result)

	result = InitResult("TestOpenGzFalse-2")
	// Should success
	f, err = Open("test_files/open-test.gz", os.O_RDONLY, GZ_FALSE)
	if err != nil {
		success = false
		goto out
	}
	success, err = CheckFileContentsMatch(f, "Hello World", false)
	if err == nil && !success {
		success = false
		goto out
	}
out:
	HandleResult(t, success, result)
}

func TestWriteGz(t *testing.T) {
	t.Parallel()

	var success bool = false
	var err error
	var f *File
	var writer Writer

	result := InitResult("TestWriteGz")

	if f, err = Open("test_files/write-gz.gz", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, GZ_TRUE); err != nil {
		success = false
		goto out
	}
	defer f.Close()
	if writer, err = f.Writer(0); err != nil {
		success = false
		goto out
	}
	writer.Write([]byte("Hello World"))
	writer.Flush()
	writer.Close()
	f.Close()

	if f, err = Open("test_files/write-gz.gz", os.O_RDONLY, GZ_TRUE); err != nil {
		success = false
		goto out
	}
	if success, err = CheckFileContentsMatch(f, "Hello World", true); err != nil || !success {
		success = false
		goto out
	}
out:
	HandleResult(t, success, result)
}

func TestOpenGzTrue(t *testing.T) {
	t.Parallel()

	var success bool = false
	var err error
	var f *File

	result := InitResult("TestOpenGzTrue-1")

	// Should pass
	f, err = Open("test_files/open-test.gz", os.O_RDONLY, GZ_TRUE)
	if err != nil {
		success = false
		goto out
	}
	success, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || !success {
		success = false
		goto out
	}
	HandleResult(t, success, result)

	result = InitResult("TestOpenGzTrue-2")
	// Should success
	f, err = Open("test_files/open-test.txt", os.O_RDONLY, GZ_TRUE)
	if err != nil {
		success = false
		goto out
	}
	success, err = CheckFileContentsMatch(f, "Hello World", false)
	if err == nil && !success {
		success = false
		goto out
	}
out:
	HandleResult(t, success, result)
}

func TestOpenGzUnknown(t *testing.T) {
	t.Parallel()

	var fail bool = false
	var err error
	var f *File

	result := InitResult("TestOpenGzUnknown-1")

	// Should pass
	f, err = Open("test_files/open-test.gz", os.O_RDONLY, GZ_UNKNOWN)
	if err != nil {
		fail = true
		goto out
	}
	fail, err = CheckFileContentsMatch(f, "Hello World", true)
	if err != nil || fail {
		goto out
	}
	HandleResult(t, fail, result)

	result = InitResult("TestOpenGzUnknown-2")
	// Should fail
	f, err = Open("test_files/open-test.txt", os.O_RDONLY, GZ_UNKNOWN)
	if err != nil {
		fail = true
		goto out
	}
	fail, err = CheckFileContentsMatch(f, "Hello World", false)
	if err == nil || !fail {
		fail = true
		goto out
	}
out:
	HandleResult(t, fail, result)
}

func TestListFiles(t *testing.T) {
	t.Parallel()

	var success bool = true
	result := InitResult("TestListFiles")

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

		if !reflect.DeepEqual(trimmed, answer) {
			fmt.Println("Expected: %v", answer)
			fmt.Println("Got:      %v", trimmed)
			success = false
		} else {
			//			fmt.Println("Passed %s", p)
		}
	}

	HandleResult(t, success, result)
}

func TestListDirs(t *testing.T) {
	t.Parallel()

	var success bool = true
	result := InitResult("TestListDirs")

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

		assert.Equal(t, answer, trimmed, "Did not match")
	}
	HandleResult(t, success, result)
}

func TestExists(t *testing.T) {
	// Test success
	var exists bool
	var err error
	exists, err = Exists("./test_files")
	assert.Equal(t, nil, err, "Failed to check exists on existing directory")
	assert.Equal(t, true, exists, "Exists failed on existing directory")

	exists, err = Exists("./doesnotexist")
	assert.Equal(t, nil, err, "Failed to check exists on non-existing directory")
	assert.Equal(t, false, exists, "Exists failed on non-existing directory")
}
