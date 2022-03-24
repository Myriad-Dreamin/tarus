package tarus_io

import "github.com/containerd/containerd/cio"

type JudgeChecker interface {
	GetJudgeResult() ([]byte, error)
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

func NopCIO(c cio.Creator) Factory {
	return &nopCio{c: c}
}

func NopCIO2(c cio.Creator, r JudgeChecker) Factory {
	return &nopCio{c: c, r: r}
}
