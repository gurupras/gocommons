package gocommons

import (
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var args []string
	var err error

	args, err = shlex.Split("ls -l -i -s -a")
	assert.Nil(err, "Failed to split args", err)
	ret, _, stderr := Execv(args[0], args[1:], false)
	assert.Zero(ret, "Failed to run valid command", stderr)

	args, err = shlex.Split("programmustnotexist -l -i -s -a")
	assert.Nil(err, "Failed to split args", err)

	ret, _, stderr = Execv(args[0], args[1:], false)
	assert.NotZero(ret, "Succeeded on illegal command")
}

func TestExecShell(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var args []string
	var err error

	args, err = shlex.Split("ls -l -i -s -a")
	assert.Nil(err, "Failed to split args")
	ret, _, stderr := Execv(args[0], args[1:], true)
	assert.Zero(ret, "Failed to run valid command", stderr)

	args, err = shlex.Split("programmustnotexist -l -i -s -a")
	ret, _, stderr = Execv(args[0], args[1:], true)
	assert.NotZero(ret, "Succeeded on illegal command")
}

func TestSliceArgs(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	cmd := "ls -l -i -sa /tmp"

	stringSplit := strings.Split(cmd, " ")
	sliceArgs := SliceArgs(cmd)

	assert.Equal(stringSplit, sliceArgs, "Did not get expected slice")
}

func TestExecv1(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	cmd := "ls"
	args := "-l -i -s -a -d /tmp"
	ret, _, stderr := Execv1(cmd, args, true)
	assert.Zero(ret, "Got non-zero error code")
	assert.Equal("", stderr, "Stderr is not empty")

}
