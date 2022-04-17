package tarus_compiler

import "strings"

type CompilerArgs struct {
	CompileTarget int
	CompiledRole  string
	Workdir       string
	Args          []string
	// Note: map might be initialized as null
	ExtraArgs    map[string]string
	Environments map[string]string
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

	CompileTargetBinary
	CompileTargetPythonBinary

	CompileTargetLanguageC
	CompileTargetLanguageCpp
	CompileTargetLanguageDelphi
	CompileTargetLanguageJava
	CompileTargetLanguageJavascript
	CompileTargetLanguageCSharp
	CompileTargetLanguageFSharp
	CompileTargetLanguageGolang
	CompileTargetLanguageHaskell
	CompileTargetLanguageKotlin
	CompileTargetLanguageNodeJs
	CompileTargetLanguageOCaml
	CompileTargetLanguagePascal
	CompileTargetLanguagePerl
	CompileTargetLanguagePhp
	CompileTargetLanguagePython
	CompileTargetLanguagePython2
	CompileTargetLanguagePython3
	CompileTargetLanguageRuby
	CompileTargetLanguageRust
	CompileTargetLanguageScala
	CompileTargetLanguageSwift
	CompileTargetLanguageTypescript
)

func ExtToCompileTarget(ext string) int {
	switch ext {
	case ".exe":
		return CompileTargetBinary
	case ".pyc":
		return CompileTargetPythonBinary
	case ".c":
		return CompileTargetLanguageC
	case ".dpr":
		return CompileTargetLanguageDelphi
	case ".java":
		return CompileTargetLanguageJava
	case ".py":
		return CompileTargetLanguagePython
	case ".py2":
		return CompileTargetLanguagePython2
	case ".py3":
		return CompileTargetLanguagePython3
	case ".perl", ".pl", ".PL":
		return CompileTargetLanguagePerl
	case ".php":
		return CompileTargetLanguagePhp
	case ".cc", ".c++", ".cpp":
		return CompileTargetLanguageCpp
	case ".cs":
		return CompileTargetLanguageCSharp
	case ".fsi":
		return CompileTargetLanguageFSharp
	case ".go":
		return CompileTargetLanguageGolang
	case ".js":
		return CompileTargetLanguageJavascript
	//case "nodejs":
	//	return CompileTargetLanguageNodeJs
	case ".hs", ".lhs":
		return CompileTargetLanguageHaskell
	case ".kt", ".kts", ".ktm":
		return CompileTargetLanguageKotlin
	case ".ml", ".mli":
		return CompileTargetLanguageOCaml
	case ".pas":
		return CompileTargetLanguagePascal
	case ".rb", ".ruby":
		return CompileTargetLanguageRuby
	case ".rs":
		return CompileTargetLanguageRust
	case ".swift":
		return CompileTargetLanguageSwift
	case ".scala":
		return CompileTargetLanguageScala
	case ".ts", ".tsx":
		return CompileTargetLanguageTypescript
	default:
		return CompileTargetUnknown
	}
}

func MimeToCompileTarget(s string) int {
	if strings.HasPrefix(s, "language/") {
		s = strings.TrimPrefix(s, "language/")
		switch s {
		case "c":
			return CompileTargetLanguageC
		case "delphi":
			return CompileTargetLanguageDelphi
		case "java":
			return CompileTargetLanguageJava
		case "python":
			return CompileTargetLanguagePython
		case "python2":
			return CompileTargetLanguagePython2
		case "python3":
			return CompileTargetLanguagePython3
		case "perl":
			return CompileTargetLanguagePerl
		case "php":
			return CompileTargetLanguagePhp
		case "c++", "cpp":
			return CompileTargetLanguageCpp
		case "c#", "csharp":
			return CompileTargetLanguageCSharp
		case "f#", "fsharp":
			return CompileTargetLanguageFSharp
		case "go", "golang":
			return CompileTargetLanguageGolang
		case "js", "javascript":
			return CompileTargetLanguageJavascript
		case "nodejs":
			return CompileTargetLanguageNodeJs
		case "haskell":
			return CompileTargetLanguageHaskell
		case "kotlin":
			return CompileTargetLanguageKotlin
		case "ocaml":
			return CompileTargetLanguageOCaml
		case "pascal":
			return CompileTargetLanguagePascal
		case "ruby":
			return CompileTargetLanguageRuby
		case "rust":
			return CompileTargetLanguageRust
		case "swift":
			return CompileTargetLanguageSwift
		case "scala":
			return CompileTargetLanguageScala
		case "ts", "typescript":
			return CompileTargetLanguageTypescript
		default:
			return CompileTargetUnknown
		}
	} else if strings.HasPrefix(s, "application/") {
		s = strings.TrimPrefix(s, "application/")
		switch s {
		case "octet-stream":
			return CompileTargetBinary
		default:
			return CompileTargetUnknown
		}
	}
	return CompileTargetUnknown
}

func CompileTargetToMime(c int) string {
	switch c {
	case CompileTargetLanguageC:
		return "language/c"
	case CompileTargetLanguageDelphi:
		return "language/delphi"
	case CompileTargetLanguageJava:
		return "language/java"
	case CompileTargetLanguagePython:
		return "language/python"
	case CompileTargetLanguagePython2:
		return "language/python2"
	case CompileTargetLanguagePython3:
		return "language/python3"
	case CompileTargetLanguagePerl:
		return "language/perl"
	case CompileTargetLanguagePhp:
		return "language/php"
	case CompileTargetLanguageCpp:
		return "language/cpp"
	case CompileTargetLanguageCSharp:
		return "language/csharp"
	case CompileTargetLanguageFSharp:
		return "language/fsharp"
	case CompileTargetLanguageGolang:
		return "language/golang"
	case CompileTargetLanguageJavascript:
		return "language/javascript"
	case CompileTargetLanguageNodeJs:
		return "language/nodejs"
	case CompileTargetLanguageHaskell:
		return "language/haskell"
	case CompileTargetLanguageKotlin:
		return "language/kotlin"
	case CompileTargetLanguageOCaml:
		return "language/ocaml"
	case CompileTargetLanguagePascal:
		return "language/pascal"
	case CompileTargetLanguageRuby:
		return "language/ruby"
	case CompileTargetLanguageRust:
		return "language/rust"
	case CompileTargetLanguageSwift:
		return "language/swift"
	case CompileTargetLanguageScala:
		return "language/scala"
	case CompileTargetLanguageTypescript:
		return "language/typescript"
	default:
		return ""
	}
}
