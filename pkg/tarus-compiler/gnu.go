package tarus_compiler

import (
	"fmt"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

func getSystemGnuToolchain() (CompilerSerial, error) {
	var compiler = gnuCompiler{}
	if err := compiler.detect("/usr"); err != nil {
		return CompilerSerial{}, err
	}

	compiler.c.Compiler = compiler
	return compiler.c, nil
}

func GetSystemGnu() *MultiVerCompiler {
	toolchain, err := getSystemGnuToolchain()
	if err != nil {
		return nil
	}

	return &MultiVerCompiler{
		SystemToolchain: toolchain,
	}
}

type gnuCompiler struct {
	c     CompilerSerial
	Patch int
	GCC   string
}

func (g gnuCompiler) Compile(args *CompilerArgs) (CompilerResponse, error) {
	return CompilerResponse{}, errdefs.ErrNotImplemented
}

func (g gnuCompiler) detect(s string) error {
	gcc := filepath.Join(s, "bin/gcc")
	if exists(gcc) {
		v, err := output(gcc, "--version")
		if err != nil {
			return err
		}

		versionCheck0 := strings.SplitN(v, "\n", 2)[0]
		versionCheck1 := strings.Split(versionCheck0, " ")
		versionCheck2 := strings.TrimSpace(versionCheck1[len(versionCheck1)-1])
		var major, minor, patch int
		n, err := fmt.Sscanf(versionCheck2, "%d.%d.%d", &major, &minor, &patch)
		if err != nil {
			return errors.Wrapf(err, "version detection failed")
		}
		if n != 3 {
			return errors.New("version pattern matching failed")
		}

		g.c.Version = versionCheck2
		g.c.Major = major
		g.c.Minor = minor
		g.Patch = patch
		g.GCC = gcc
		return nil
	}

	return errdefs.ErrNotFound
}
