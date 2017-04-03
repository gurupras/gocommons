package gocommons

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/require"
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

func TestExecNoWait(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	cmdline, err := shlex.Split("programmustnotexist -l -i -s -a")
	require.Nil(err)

	p, err := ExecvNoWait(cmdline[0], cmdline[1:], true)
	require.Nil(err, "Should not fail for valid process")
	require.NotNil(p, "Should not fail for valid process")
	p.Wait()
	require.False(p.ProcessState.Success())

	cmdline, err = shlex.Split("ls -l -i -s -a")
	require.Nil(err)

	p, err = ExecvNoWait(cmdline[0], cmdline[1:], true)
	require.Nil(err, "Should not fail for valid process")
	require.NotNil(p, "Should not fail for valid process")
	p.Wait()
	require.True(p.ProcessState.Success())

	// TODO: Non-shell version
}
