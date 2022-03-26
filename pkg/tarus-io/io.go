package tarus_io

import (
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/containerd/containerd/cio"
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
	c cio.Creator
	r JudgeChecker
}

func (n nopCio) AsCreator() cio.Creator {
	return n.c
}

func (n nopCio) GetJudgeResult() ([]byte, error) {
	if n.r != nil {
		return n.r.GetJudgeResult()
	}
	return nil, nil
}

func (n nopCio) GetJudgeStatus(b []byte) (tarus.JudgeStatus, error) {
	if n.r != nil {
		return n.r.GetJudgeStatus(b)
	}
	return tarus.JudgeStatus_Unknown, nil
}

func NopCIO(c cio.Creator) Factory {
	return &nopCio{c: c}
}

func NopCIO2(c cio.Creator, r JudgeChecker) Factory {
	return &nopCio{c: c, r: r}
}
