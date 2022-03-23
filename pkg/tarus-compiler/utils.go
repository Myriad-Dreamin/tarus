package tarus_compiler

import (
	"bytes"
	"os"
	"os/exec"
)

func exists(fp string) bool {
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	return false
}

func output(cmd0 string, args ...string) (string, error) {
	cmd := exec.Command(cmd0, args...)

	var stdOut = bytes.NewBuffer(nil)
	cmd.Stdout = stdOut

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return stdOut.String(), nil
}

func errorOutput(cmd0 string, args ...string) (string, error) {
	cmd := exec.Command(cmd0, args...)

	var errOut = bytes.NewBuffer(nil)
	cmd.Stderr = errOut

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return errOut.String(), nil
}
