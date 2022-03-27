package polygon_driver

type StringVal struct {
	Value      *string `xml:"value,attr"`
	AutoUpdate *bool   `xml:"auto-update,attr"`
}

type BooleanVal struct {
	Value *bool `xml:"value,attr"`
}

type ProblemConfigJudging struct {
	TimeLimit      StringVal  `xml:"time-limit"`
	MemoryLimit    StringVal  `xml:"memory-limit"`
	InputFile      StringVal  `xml:"input-file"`
	OutputFile     StringVal  `xml:"output-file"`
	TestValidator  StringVal  `xml:"test-validator"`
	Checker        StringVal  `xml:"checker"`
	Testlib        StringVal  `xml:"testlib"`
	TestWellFormed BooleanVal `xml:"tests-well-formed"`
}

type ProblemConfigPackaging struct {
	TestsetPattern             StringVal  `xml:"testset-pattern"`
	InputFilePathPattern       StringVal  `xml:"input-file-path-pattern"`
	AnswerFilePathPattern      StringVal  `xml:"answer-file-path-pattern"`
	StatementTemplateFile      StringVal  `xml:"statement-template-file"`
	StatementPathPattern       StringVal  `xml:"statement-path-pattern"`
	CheckerPath                StringVal  `xml:"checker-path"`
	RenderFormulasUsingMathjax BooleanVal `xml:"render-formulas-using-mathjax"`
}

type ProblemConfig struct {
	Judging   ProblemConfigJudging   `xml:"judging"`
	Packaging ProblemConfigPackaging `xml:"packaging"`
	Name      string                 `xml:"name,attr"`
}
