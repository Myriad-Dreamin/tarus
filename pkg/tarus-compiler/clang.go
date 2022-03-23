package tarus_compiler

import (
	"github.com/containerd/containerd/errdefs"
	"path/filepath"
)

type MultiVerCompiler struct {
	SystemToolchain CompilerSerial
	Toolchains      map[int][]CompilerSerial
}

func getSystemClangToolchain() (CompilerSerial, error) {
	var compiler = clangCompiler{}
	if err := compiler.detect("/usr"); err != nil {
		return CompilerSerial{}, err
	}

	compiler.c.Compiler = &compiler
	return compiler.c, nil
}

func GetSystemClang() *MultiVerCompiler {
	toolchain, err := getSystemClangToolchain()
	if err != nil {
		return nil
	}

	return &MultiVerCompiler{
		SystemToolchain: toolchain,
	}
}

type clangCompiler struct {
	c     CompilerSerial
	Patch int
	Clang string
}

func (g *clangCompiler) Compile(args *CompilerArgs) (CompilerResponse, error) {
	return CompilerResponse{}, errdefs.ErrNotImplemented
}

func (g clangCompiler) detect(s string) error {
	clang := filepath.Join(s, "bin/clang")
	if exists(clang) {
		_, err := output(clang, "--version")
		if err != nil {
			return err
		}

		g.Clang = clang
		return nil
	}

	return errdefs.ErrNotFound
}
