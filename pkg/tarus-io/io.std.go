package tarus_io

import (
	"github.com/containerd/containerd/cio"
	"io"
)

func NewStd(inp io.Reader, oup io.Writer, erp io.Writer) Factory {
	if erp == nil {
		erp = io.Discard
	}

	if j, ok := oup.(JudgeChecker); ok {
		return NopCIO2(cio.NewCreator(cio.WithStreams(inp, oup, erp)), j)
	}
	return NopCIO(cio.NewCreator(cio.WithStreams(inp, oup, erp)))
}
