package testlib_checker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type CmdChecker struct {
	io.Writer
	ctx     context.Context
	Cmd     *exec.Cmd
	pos     int
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
	closers []io.Closer
	cancel  context.CancelFunc
}

// CloseWithTimeout must be invoked in main process to this time (sync process)
func (c *CmdChecker) CloseWithTimeout(t time.Duration) error {
	if c.cancel == nil {
		return nil
	}
	var cancel = c.cancel
	c.cancel = nil
	if cancel == nil {
		return nil
	}

	var timer *time.Timer
	if t != 0 {
		timer = time.AfterFunc(t, cancel)
	}
	if err := c.Cmd.Wait(); err != nil {
		return err
	}
	timer.Stop()

	for i := range c.closers {
		_ = c.closers[i].Close()
	}
	for i := range c.Cmd.ExtraFiles {
		_ = c.Cmd.ExtraFiles[i].Close()
	}
	return nil
}

func (c *CmdChecker) Close() error {
	return c.CloseWithTimeout(time.Second * 5)
}

func (c *CmdChecker) GetJudgeResult() ([]byte, error) {
	if err := c.Close(); err != nil {
		fmt.Println("stdout", c.stdout.String())
		fmt.Println("stderr", c.stderr.String())
		return nil, err
	}
	return c.stderr.Bytes(), nil
}

//

func TestlibChecker(checkerCmd string, inp, oup, ans *os.File) (*CmdChecker, error) {
	var ctx, cancel = context.WithCancel(context.Background())

	var extraFiles []*os.File
	var checkerArgs []string
	var fdAssign = 3
	var appendFd = func(f *os.File, otherwise string) {
		if f != nil {
			extraFiles = append(extraFiles, f)
			checkerArgs = append(checkerArgs, fmt.Sprintf("/proc/self/fd/%d", fdAssign))
			fdAssign += 1
		} else {
			checkerArgs = append(checkerArgs, otherwise)
		}
	}

	// input
	appendFd(inp, os.DevNull)
	// output
	appendFd(oup, "/dev/stdin")
	// answer
	appendFd(ans, os.DevNull)
	var cmd = exec.CommandContext(ctx, checkerCmd, checkerArgs...)

	var solverBuf = bytes.NewBuffer(nil)
	var stdout = bytes.NewBuffer(nil)
	var stderr = bytes.NewBuffer(nil)
	var c = &CmdChecker{Cmd: cmd, Writer: solverBuf, stdout: stdout, stderr: stderr}
	cmd.Stdin = solverBuf
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	c.ctx = ctx
	c.cancel = cancel
	cmd.ExtraFiles = append(cmd.ExtraFiles, extraFiles...)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return c, nil
}
