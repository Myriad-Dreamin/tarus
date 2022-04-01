package tarus_judge

import (
	"context"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
)

type WithContainerRequest = tarus.CreateContainerRequest

func WithContainerEnvironment(
	c tarus.JudgeServiceServer, rawCtx context.Context, req *WithContainerRequest, cb func(rawCtx context.Context) error) (err error) {
	_, err = c.CreateContainer(rawCtx, req)
	if err != nil {
		return err
	}
	defer func() {
		err2 := err
		_, err = c.RemoveContainer(rawCtx, &tarus.RemoveContainerRequest{
			TaskKey: req.TaskKey,
		})
		if err2 != nil {
			err = err2
		}
	}()

	err = cb(rawCtx)
	return err
}
