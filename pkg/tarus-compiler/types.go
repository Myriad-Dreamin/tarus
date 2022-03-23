package tarus_compiler

type CompilerArgs struct {
	CompileTarget int
	CompiledRole  string
	Workdir       string
	Args          []string
	Environments  map[string]string
}

type CompilerResponse struct {
	DiagPage    string
	ExitSummary int
}

type Compiler interface {
	Compile(args *CompilerArgs) (CompilerResponse, error)
}

type CompilerSerial struct {
	Compiler
	Major    int
	Minor    int
	Version  string
	Features [][2]string
}

const (
	CompileTargetUnknown = iota
	CompileTargetDefault
)
