package gocommons

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/shlex"
)

func TestExec(t *testing.T) {
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
