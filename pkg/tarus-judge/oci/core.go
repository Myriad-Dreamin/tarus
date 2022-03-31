package oci_judge

import (
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"io"
)

type OCIJudgeServiceServer interface {
	tarus.JudgeServiceServer
}

type MemoryJudgeConfig struct {
	CallbackKey []byte
	ImageId     string
	ProcessArgs []string
	Input       io.ReadCloser
	Output      io.WriteCloser
	Timeout     int64 // in microsecond
	Memory      int64 // in bytes
}

type JudgeHint struct {
	Code          int    `json:"code" yaml:"code"`
	Signal        string `json:"signal,omitempty" yaml:"signal,omitempty"`
	CheckerResult string `json:"checker_result" yaml:"checker_result"`
}
