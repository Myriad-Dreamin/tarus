package tarus_compiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func exists(fp string) bool {
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	return false
}

func consumeCmd(c *exec.Cmd) error {
	if err := c.Start(); err != nil {
		return err
	}

	if err := c.Wait(); err != nil {
		return err
	}

	if code := c.ProcessState.ExitCode(); code != 0 {
		return fmt.Errorf("process %v exit with code %d", c.Path, code)
	}
	return nil
}

func output(cmd0 string, args ...string) (string, error) {
	cmd := exec.Command(cmd0, args...)

	var stdOut = bytes.NewBuffer(nil)
	cmd.Stdout = stdOut

	if err := consumeCmd(cmd); err != nil {
		return "", err
	} else {
		return stdOut.String(), nil
	}
}

func errorOutput(cmd0 string, args ...string) (string, error) {
	cmd := exec.Command(cmd0, args...)

	var errOut = bytes.NewBuffer(nil)
	cmd.Stderr = errOut

	if err := consumeCmd(cmd); err != nil {
		return "", err
	} else {
		return errOut.String(), nil
	}
}
