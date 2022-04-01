package tarus_judge

import (
	"context"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
)

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
