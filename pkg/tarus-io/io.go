package tarus_io

import (
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/containerd/containerd/cio"
	"io"
)

type JudgeChecker interface {
	GetJudgeResult() ([]byte, error)
	GetJudgeStatus(b []byte) (tarus.JudgeStatus, error)
}

type Factory interface {
	AsCreator() cio.Creator
	JudgeChecker
}

type nopCio struct {
	c       cio.Creator
	r       JudgeChecker
	closers []io.Closer
}

func (n *nopCio) AsCreator() cio.Creator {
	return n.c
}

func (n *nopCio) GetJudgeResult() (b []byte, err error) {
	if n.r != nil {
		b, err = n.r.GetJudgeResult()
	}
	for i := range n.closers {
		_ = n.closers[i].Close()
	}
	n.closers = nil
	return nil, nil
}

func (n *nopCio) GetJudgeStatus(b []byte) (tarus.JudgeStatus, error) {
	if n.r != nil {
		return n.r.GetJudgeStatus(b)
	}
	return tarus.JudgeStatus_Unknown, nil
}

func NopCIO(c cio.Creator) *nopCio {
	return &nopCio{c: c}
}

func NopCIO2(c cio.Creator, r JudgeChecker) *nopCio {
	return &nopCio{c: c, r: r}
}
