package testlib_frontend

import (
	tarus_compiler "github.com/Myriad-Dreamin/tarus/pkg/tarus-compiler"
	"github.com/containerd/containerd/errdefs"
	"path/filepath"
	"testing"
)

func TestCompile(t *testing.T) {
	c := tarus_compiler.GetSystemClang()
	if c == nil {
		t.Fatal(errdefs.ErrNotFound)
	}

	var compiler = &Compiler{
		C:           c.SystemToolchain,
		ProjectRoot: "third_party/testlib",
	}
	var BinRoot = "bin/testlib"
	if err := compiler.CompileChecker(
		filepath.Join(compiler.ProjectRoot, "checkers/yesno.cpp"),
		filepath.Join(BinRoot, "checkers/yesno")); err != nil {
		t.Fatal(err)
	}
}
