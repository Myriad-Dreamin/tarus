package tarus_compiler

import (
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
)

type StaticRouter struct{}

var Statics = &StaticRouter{}

func (s *StaticRouter) Compile(args *CompilerArgs) (CompilerResponse, error) {
	switch args.CompileTarget {
	case CompileTargetLanguageCpp:
		c := GetSystemClang()
		if c == nil {
			return CompilerResponse{ExitSummary: 1}, errors.Wrapf(errdefs.ErrNotFound, "system clang not found")
		}

		args.Args = append(args.Args, "-lstdc++")
		return c.SystemToolchain.Compile(args)
	default:
		return CompilerResponse{ExitSummary: 1}, errors.Wrapf(errdefs.ErrNotImplemented, "compile target: %d", args.CompileTarget)
	}
}
