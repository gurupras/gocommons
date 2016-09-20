package gocommons

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	t.Parallel()

	var success bool = true
	var args []string
	var err error
	var result string

	result = InitResult("TestExec-1")
	if args, err = (shlex.Split("ls -l -i -s -a")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		success = false
		goto out
	} else {
		if ret, _, _ := Execv(args[0], args[1:], false); ret < 0 {
			success = false
		}
	}
	HandleResult(t, success, result)

	result = InitResult("TestExec-2")
	if args, err = (shlex.Split("programmustnotexist -l -i -s -a")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		success = false
		goto out
	} else {
		if ret, _, _ := Execv(args[0], args[1:], false); ret >= 0 {
			success = false
		}
	}
out:
	HandleResult(t, success, result)
}

func TestExecShell(t *testing.T) {
	t.Parallel()

	var success bool = true
	var args []string
	var err error
	var result string

	result = InitResult("TestExecShell-1")
	if args, err = (shlex.Split("ls -l -i -s -a")); err != nil {
		fmt.Println(os.Stderr, err)
		success = false
		goto out
	} else {
		if ret, _, _ := Execv(args[0], args[1:], true); ret < 0 {
			success = false
		}
	}
	HandleResult(t, success, result)

	result = InitResult("TestExecShell-2")
	if args, err = (shlex.Split("programmustnotexist -l -i -s -a")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		success = false
		goto out
	} else {
		if ret, _, _ := Execv(args[0], args[1:], true); ret >= 0 {
			success = false
		}
	}
out:
	HandleResult(t, success, result)
}

func TestSliceArgs(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	cmd := "ls -l -i -sa /tmp"

	stringSplit := strings.Split(cmd, " ")
	sliceArgs := SliceArgs(cmd)

	assert.Equal(stringSplit, sliceArgs, "Did not get expected slice")
}
