package testlib_checker

import (
	"io"
	"os/exec"
)

type Checker struct {
	io.Writer
	Cmd *exec.Cmd
	pos int
}

func TestlibChecker(checkerCmd string, r io.Reader) io.Writer {
	var cmd = exec.Command(checkerCmd)
	cmd.Stdin = r

	return &Checker{Cmd: cmd, Writer: cmd.Stdout}
}
