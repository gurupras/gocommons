package gocommons

import (
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

	result := InitResult("TestIntSort")

	if chunks, err = ExternalSort("./test_files/sort.gz", memory, IntSortParams); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Failed to sort file:", err))
		success = false
		goto out
	}
	NWayMerge(chunks, "test_files/output.gz", memory, IntSortParams)
	for _, chunk := range chunks {
		_ = chunk
		os.Remove(chunk)
	}
out:
	HandleResult(t, success, result)
}
