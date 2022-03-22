package tarus_store

import (
	"context"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
)

// task key -> session

type JudgeSessionStore interface {
	SetJudgeSession(ctx context.Context, key []byte, meta *tarus.OCIJudgeSession) error
	GetJudgeSession(ctx context.Context, key []byte) (*tarus.OCIJudgeSession, error)

	// FinishSession requires transactionCb is idempotent
	FinishSession(ctx context.Context, key []byte, transactionCb func() error) error
}

const (
	CommitStatusIdle = iota
	CommitStatusFinished
)
