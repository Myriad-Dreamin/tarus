//go:build linux

package testlib_checker

import (
	"bytes"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/pkg/memfd"
	"io"
	"os"
	"runtime"
	"time"
)

type ProcessChecker struct {
	Process   *os.Process
	pos       int
	stdout    *os.File
	stderr    *os.File
	stdoutOut []byte
	stderrOut []byte
	closers   []io.Closer
}

func (c *ProcessChecker) closeResources() {
	for i := range c.closers {
		_ = c.closers[i].Close()
	}
}

// CloseWithTimeout must be invoked in main process to this time (sync process)
func (c *ProcessChecker) CloseWithTimeout(t time.Duration) (err error) {
	var timer *time.Timer
	if t != 0 {
		timer = time.AfterFunc(t, func() {
			_ = c.Process.Kill()
		})
	}
	_, err = c.Process.Wait()
	if timer != nil {
		timer.Stop()
	}
	c.stdoutOut = c.stringFromFdFile(c.stdout)
	c.stderrOut = c.stringFromFdFile(c.stderr)
	c.closeResources()
	return
}

func (c *ProcessChecker) Close() error {
	return c.CloseWithTimeout(time.Second * 5)
}

func (c *ProcessChecker) GetJudgeResult() ([]byte, error) {
	if err := c.Close(); err != nil {
		fmt.Println("stdout", c.stdoutOut)
		fmt.Println("stderr", c.stderrOut)
		return nil, err
	}
	//fmt.Println("stdout", c.stdoutOut)
	//fmt.Println("stderr", c.stderrOut)
	return c.stderrOut, nil
}

func (c *ProcessChecker) writerDescriptor() (f *os.File, err error) {
	f, err = memfd.New("golang_mem")
	if err != nil {
		return
	}
	c.closers = append(c.closers, f)
	return
}

func (c *ProcessChecker) stringFromFdFile(s *os.File) []byte {
	var b = make([]byte, 16)
	var parts [][]byte
	_, err := s.Seek(0, 0)
	var n int
	for err == nil {
		n, err = s.Read(b)
		b = b[:n]
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			if len(parts) == 0 {
				return b
			}
			break
		}
		parts = append(parts, b)
		b = make([]byte, 16)
	}

	return bytes.Join(parts, nil)
}

func TestlibCheckerProcessImpl(checkerCmd string, inp, oup, ans *os.File) (c *ProcessChecker, err error) {
	var extraFiles []*os.File
	var checkerArgv = []string{checkerCmd}
	var fdAssign int
	var fdAssignTmpl = []string{"/proc/self/fd/3", "/proc/self/fd/4", "/proc/self/fd/5"}
	var appendFd = func(f *os.File, otherwise string) {
		if f != nil {
			extraFiles = append(extraFiles, f)
			checkerArgv = append(checkerArgv, fdAssignTmpl[fdAssign])
			fdAssign += 1
			c.closers = append(c.closers, f)
		} else {
			checkerArgv = append(checkerArgv, otherwise)
		}
	}

	c = &ProcessChecker{closers: make([]io.Closer, 0, 5)}
	stdout, err := c.writerDescriptor()
	if err != nil {
		c.closeResources()
		return nil, err
	}
	stderr, err := c.writerDescriptor()
	if err != nil {
		c.closeResources()
		return nil, err
	}
	c.stdout = stdout
	c.stderr = stderr

	extraFiles = []*os.File{nil, stdout, stderr}
	// input
	appendFd(inp, os.DevNull)
	// output
	appendFd(oup, "/dev/stdin")
	// answer
	appendFd(ans, os.DevNull)

	var env []string
	if //goland:noinspection GoBoolExpressions
	runtime.GOOS == "windows" {
		env = []string{"SYSTEMROOT=" + os.Getenv("SYSTEMROOT")}
	}

	c.Process, err = os.StartProcess(checkerCmd, checkerArgv, &os.ProcAttr{
		Files: extraFiles,
		Env:   env,
	})
	if err != nil {
		c.closeResources()
		return nil, err
	}

	return c, nil
}
