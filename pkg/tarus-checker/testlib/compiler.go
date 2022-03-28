package testlib_checker

import (
	tarus_compiler "github.com/Myriad-Dreamin/tarus/pkg/tarus-compiler"
)

type Compiler struct {
	C           tarus_compiler.Compiler
	ProjectRoot string
}

func (c *Compiler) CompileChecker(checkerSource string, checkerBin string) error {
	_, err := c.C.Compile(&tarus_compiler.CompilerArgs{
		CompileTarget: tarus_compiler.CompileTargetDefault,
		CompiledRole:  "judge",
		Args: []string{
			"@input", checkerSource,
			"@output", checkerBin,
			"@include_dir", c.ProjectRoot,
			"-lstdc++",
		},
		Environments: nil,
	})
	return err
}
