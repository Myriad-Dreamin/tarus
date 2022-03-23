package tarus_compiler

import (
	"fmt"
	"github.com/containerd/containerd/errdefs"
	"os/exec"
	"path/filepath"
	"strings"
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

func (g *clangCompiler) Compile(args *CompilerArgs) (cr CompilerResponse, err error) {
	var cmdArgs []string
	for i := 0; i < len(args.Args); i++ {
		if strings.HasPrefix(args.Args[i], "@") {
			if i+1 == len(args.Args) {
				err = fmt.Errorf("incompleted argument: %v", args.Args[i])
				return
			}
			switch args.Args[i][1:] {
			case "input":
				i++
				cmdArgs = append(cmdArgs, args.Args[i])
			case "output":
				i++
				cmdArgs = append(cmdArgs, "-o", args.Args[i])
			case "include_dir":
				i++
				cmdArgs = append(cmdArgs, "-I", args.Args[i])
			default:
				err = fmt.Errorf("unknown common option: %v", args.Args[i])
				return
			}
		} else {
			cmdArgs = append(cmdArgs, args.Args[i])
		}
	}

	err = consumeCmd(exec.Command(g.Clang, cmdArgs...))
	return
}

func (g *clangCompiler) detect(s string) error {
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
