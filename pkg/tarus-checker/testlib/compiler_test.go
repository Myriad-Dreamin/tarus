package testlib_checker

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

	for _, ts := range [][2]string{
		{"checkers/acmp.cpp", "checkers/acmp"},
		{"checkers/casencmp.cpp", "checkers/casencmp"},
		{"checkers/dcmp.cpp", "checkers/dcmp"},
		{"checkers/hcmp.cpp", "checkers/hcmp"},
		{"checkers/lcmp.cpp", "checkers/lcmp"},
		{"checkers/nyesno.cpp", "checkers/nyesno"},
		{"checkers/pointsinfo.cpp", "checkers/pointsinfo"},
		{"checkers/rcmp4.cpp", "checkers/rcmp4"},
		{"checkers/rcmp9.cpp", "checkers/rcmp9"},
		{"checkers/uncmp.cpp", "checkers/uncmp"},
		{"checkers/yesno.cpp", "checkers/yesno"},
		{"checkers/caseicmp.cpp", "checkers/caseicmp"},
		{"checkers/casewcmp.cpp", "checkers/casewcmp"},
		{"checkers/fcmp.cpp", "checkers/fcmp"},
		{"checkers/icmp.cpp", "checkers/icmp"},
		{"checkers/ncmp.cpp", "checkers/ncmp"},
		{"checkers/pointscmp.cpp", "checkers/pointscmp"},
		{"checkers/rcmp.cpp", "checkers/rcmp"},
		{"checkers/rcmp6.cpp", "checkers/rcmp6"},
		{"checkers/rncmp.cpp", "checkers/rncmp"},
		{"checkers/wcmp.cpp", "checkers/wcmp"},
	} {
		if err := compiler.CompileChecker(
			filepath.Join(compiler.ProjectRoot, ts[0]),
			filepath.Join(BinRoot, ts[1])); err != nil {
			t.Fatal(err)
		}
	}

}
