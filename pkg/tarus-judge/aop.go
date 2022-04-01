package tarus_judge

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"math/rand"
)

type TransientJudgeRequest struct {
	ImageId string
	// Pause      JudgeInfra

	CompileFile string
	BinTarget   string

	*tarus.MakeJudgeRequest
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

func TransientJudge(c tarus.JudgeServiceServer, rawCtx context.Context, rawReq *TransientJudgeRequest) (resp *tarus.MakeJudgeResponse, err error) {
	req := &*rawReq
	if req.MakeJudgeRequest == nil {
		return nil, errors.Wrap(errdefs.ErrInvalidArgument, "req.MakeJudgeRequest is required")
	}
	if (len(req.BinTarget) == 0) == (len(req.CompileFile) == 0) {
		return nil, errors.Wrap(errdefs.ErrInvalidArgument, "req.BinTarget/CompileFile arguemnt conflicts")
	}
	binTarget := req.BinTarget
	req.BinTarget = "/workdir/judging-program"

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

		if len(binTarget) == 0 {
			r, err := c.CompileProgram(rawCtx, &tarus.CompileProgramRequest{
				TaskKey: req.TaskKey,
				FromUrl: req.CompileFile,
				ToPath:  req.BinTarget,
			})
			if err != nil {
				return err
			}
			fmt.Println(r)
		} else {
			r, err := c.CopyFile(rawCtx, &tarus.CopyFileRequest{
				TaskKey: req.TaskKey,
				FromUrl: binTarget,
				ToPath:  req.BinTarget,
			})
			if err != nil {
				return err
			}
			fmt.Println(r)
		}

		r, err := c.MakeJudge(rawCtx, req.MakeJudgeRequest)
		if err != nil {
			return err
		}
		resp = r
		return nil
	})

	return
}
