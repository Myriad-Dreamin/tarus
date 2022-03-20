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
