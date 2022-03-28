package tarus_io

import (
	"github.com/containerd/containerd/cio"
	"io"
	"os"
)

func NewStd(inp io.Reader, oup io.Writer, erp io.Writer) Factory {
	if erp == nil {
		erp = io.Discard
	}

	var stdCio *nopCio
	if j, ok := oup.(JudgeChecker); ok {
		stdCio = NopCIO2(cio.NewCreator(cio.WithStreams(inp, oup, erp)), j)
	} else {
		stdCio = NopCIO(cio.NewCreator(cio.WithStreams(inp, oup, erp)))
	}

	for _, closable := range []interface{}{
		inp, oup, erp,
	} {
		if c, ok := closable.(io.Closer); ok && c != os.Stdin && c != os.Stdout && c != os.Stderr {
			stdCio.closers = append(stdCio.closers, c)
		}
	}

	return stdCio
}
