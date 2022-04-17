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
	var cmdArgs = []string{"-x", ""}
	switch args.CompileTarget {
	case CompileTargetLanguageC:
		cmdArgs[1] = "c"
	case CompileTargetLanguageCpp:
		cmdArgs[1] = "c++"
	case CompileTargetDefault:
		cmdArgs = cmdArgs[:0]
	default:
		return CompilerResponse{}, fmt.Errorf("invalid compiler target: %d(%s)", args.CompileTarget, CompileTargetToMime(args.CompileTarget))
	}

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

	s, err := errorOutput(g.Clang, cmdArgs...)
	if er, ok := err.(*exec.ExitError); ok {
		cr.ExitSummary = er.ExitCode()
		cr.DiagPage = s
		err = nil
	}
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
