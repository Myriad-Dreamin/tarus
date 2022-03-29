package tarus_judge

import (
	"context"
	"encoding/hex"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"math/rand"
)

type TransientJudgeRequest struct {
	*tarus.MakeJudgeRequest
	ImageId   string
	BinTarget string

	// Pause      JudgeInfra
}

func WithContainerEnvironment(
	c tarus.JudgeServiceServer, rawCtx context.Context, req *TransientJudgeRequest, cb func(rawCtx context.Context, req *TransientJudgeRequest) error) (err error) {
	_, err = c.CreateContainer(rawCtx, &tarus.CreateContainerRequest{
		ImageId: req.ImageId,
		TaskKey: req.TaskKey,

		BinTarget: req.BinTarget,
	})
	if err != nil {
		return err
	}
	defer func() {
		err2 := err
		// todo: task key
		_, err = c.RemoveContainer(rawCtx, &tarus.RemoveContainerRequest{
			TaskKey: req.TaskKey,
		})
		if err2 != nil {
			err = err2
		}
	}()

	err = cb(rawCtx, req)
	return err
}

var tk = []byte("transient:")

func TransientJudge(c tarus.JudgeServiceServer, rawCtx context.Context, req *TransientJudgeRequest) (resp *tarus.MakeJudgeResponse, err error) {
	if req.TaskKey == nil {
		token := make([]byte, 12)
		_, err = rand.Read(token)
		if err != nil {
			return nil, err
		}
		key := make([]byte, 24+len(tk))
		copy(key[:len(tk)], tk)
		copy(key[len(tk):], hex.EncodeToString(token))
		req.TaskKey = key
	}

	err = WithContainerEnvironment(c, rawCtx, req, func(rawCtx context.Context, req *TransientJudgeRequest) error {
		req.IsAsync = false
		r, err := c.MakeJudge(rawCtx, req.MakeJudgeRequest)
		if err != nil {
			return err
		}
		resp = r
		return nil
	})

	return
}
