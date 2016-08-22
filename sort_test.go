package gocommons

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Int int

func (a *Int) String() string {
	_ = "breakpoint"
	return strconv.Itoa(int(*a))
}

func (l *Int) Less(s SortInterface) (ret bool, err error) {
	var o *Int
	var ok bool
	if s != nil {
		if o, ok = s.(*Int); !ok {
			err = errors.New(fmt.Sprintf("Failed to convert from SortInterface to *Int:", reflect.TypeOf(s)))
			ret = false
			goto out
		}
	}
	if l != nil && o != nil {
		ret = int(*l) < int(*o)
	} else if l != nil {
		ret = true
	} else {
		ret = false
	}
out:
	return
}

type IntSlice []*Int

func ParseInt(line string) SortInterface {
	if ret, err := strconv.Atoi(line); err != nil {
		return nil
	} else {
		retval := Int(ret)
		return &retval
	}
}

var IntSortParams SortParams = SortParams{LineConvert: ParseInt, Lines: make(SortCollection, 0)}

func TestIntString(t *testing.T) {
	i := 5
	I := Int(i)
	str := I.String()

	assert.Equal(t, "5", str, "Should be equal")
}

func TestIntSort(t *testing.T) {
	var success bool = true
	var err error
	var chunks []string
	var memory int = 10485760

	merge_out_channel := make(chan SortInterface, 10000)

	result := InitResult("TestIntSort")

	callback := func(channel chan SortInterface, quit chan bool) {
		var expected_fstruct *File
		if expected_fstruct, err = Open("./test_files/sort_expected.gz", os.O_RDONLY, GZ_TRUE); err != nil {
			return
		}
		defer expected_fstruct.Close()

		expected_chan := make(chan string, 10000)
		go expected_fstruct.AsyncRead(bufio.ScanLines, expected_chan)

		var expected string
		var object SortInterface
		var ok bool
		var lines int64 = 0
		for {
			if expected, ok = <-expected_chan; !ok {
				if _, ok = <-channel; !ok {
					// Both streams are done
					break
				} else {
					assert.Fail(t, "Expected file ended while n-way merge generator still has data?")
				}
			} else {
				if object, ok = <-channel; !ok || object == nil {
					assert.Fail(t, "Expected file has data while n-way merge generator has ended?")
				}
			}
			lines++
			if lines%1000 == 0 {
				//fmt.Println("Finished comparing:", lines)
			}
			assert.Equal(t, expected, object.String())
		}
		quit <- true
	}

	if chunks, err = ExternalSort("./test_files/sort.gz", memory, IntSortParams); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Failed to sort file:", err))
		success = false
		goto out
	}
	//fmt.Println("Merging...")
	NWayMergeGenerator(chunks, IntSortParams, merge_out_channel, callback)
	for _, chunk := range chunks {
		_ = chunk
		os.Remove(chunk)
	}
out:
	HandleResult(t, success, result)
}
