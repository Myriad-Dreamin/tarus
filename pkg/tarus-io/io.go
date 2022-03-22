package tarus_io

import "github.com/containerd/containerd/cio"

type R interface {
	GetJudgeResult() (string, error)
}

type Factory interface {
	AsCreator() cio.Creator
	R
}

type nopCio struct {
	c cio.Creator
	r R
}

func (n nopCio) AsCreator() cio.Creator {
	return n.c
}

func (n nopCio) GetJudgeResult() (string, error) {
	if n.r != nil {
		return n.r.GetJudgeResult()
	}
	return "", nil
}

func NopCIO(c cio.Creator) Factory {
	return &nopCio{c: c}
}

func NopCIO2(c cio.Creator, r R) Factory {
	return &nopCio{c: c, r: r}
}
