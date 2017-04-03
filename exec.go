package gocommons

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/google/shlex"
)

var (
	ShellPath = "/bin/bash"
	LogBuf    *bufio.Writer
)

//export Execv
func Execv(cmd string, args []string, shell bool) (ret int, stdout string, stderr string) {
	var buf_stdout, buf_stderr bytes.Buffer
	var err error
	var command *exec.Cmd

	if shell == true {
		args = append([]string{cmd}, args...)
		argstring := "-c '" + strings.Join(args, " ") + "'"
		args, err = shlex.Split(argstring)
		cmd = ShellPath
	}

	// Create a string to log for the command that we're running
	var cmd_string string = cmd + " "
	for i, arg := range args {
		cmd_string += arg
		if i != len(args)-1 {
			cmd_string += " "
		}
	}
	//fmt.Println("cmd: ", cmd)
	//fmt.Println("args:", args)
	//fmt.Println("cmd_string", cmd_string)

	command = exec.Command(cmd, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	command.Stdout = &buf_stdout
	command.Stderr = &buf_stderr
	if err = command.Run(); err != nil {
		ret = -1
	} else {
		ret = 0
	}
	stdout = buf_stdout.String()
	stderr = buf_stderr.String()

	return
}

func SliceArgs(args string) (ret []string) {
	ret, _ = shlex.Split(args)
	return
}

//export Execv1
func Execv1(cmd string, args string, shell bool) (ret int, stdout string, stderr string) {
	return Execv(cmd, SliceArgs(args), shell)
}

func ExecvNoWait(cmd string, args []string, shell bool) (*exec.Cmd, error) {
	var err error
	var command *exec.Cmd

	if shell == true {
		args = append([]string{cmd}, args...)
		argstring := "-c '" + strings.Join(args, " ") + "'"
		args, err = shlex.Split(argstring)
		cmd = ShellPath
	}

	// Create a string to log for the command that we're running
	var cmd_string string = cmd + " "
	for i, arg := range args {
		cmd_string += arg
		if i != len(args)-1 {
			cmd_string += " "
		}
	}

	command = exec.Command(cmd, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = command.Start()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to start process: %v", err))

	}

	return command, nil
}
