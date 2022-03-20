package oci_judge

import (
	"context"
	"errors"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"syscall"
)

func (c *ContainerdJudgeServiceServer) prepareImageOnSnapshotter(
	ctx context.Context, imageId string, snapshotter string) (err error) {
	var (
		client = c.client
		image  containerd.Image
	)
	{
		image, err = client.GetImage(ctx, imageId)
		if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
			return err
		}
	}
	if image == nil {
		image, err = client.Pull(ctx, imageId, containerd.WithPullUnpack)
		if err != nil && !errors.Is(err, errdefs.ErrNotFound) {
			return err
		}
	}

	unpacked, err := image.IsUnpacked(ctx, snapshotter)
	if err != nil {
		return err
	}
	if !unpacked {
		if err = image.Unpack(ctx, snapshotter); err != nil {
			return err
		}
	}

	return nil
}

func (c *ContainerdJudgeServiceServer) killTask(ctx context.Context, t containerd.Task) (err error) {
	defer func() {
		if err == nil {
			_, err = t.Delete(ctx)
		} else {
			_, _ = t.Delete(ctx)
		}
	}()
	// make sure we wait before calling start
	exitStatusC, err := t.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}

	// kill the process and get the exit status
	if err = t.Kill(ctx, syscall.SIGKILL); err != nil {
		return err
	}

	// wait for the process to fully exit and print out the exit status
	st := <-exitStatusC
	code, _, err := st.Result()
	if err != nil {
		return err
	}
	fmt.Printf("linux container exited with status: %d, %v\n", code, st)
	return nil
}

func (c *ContainerdJudgeServiceServer) withFreshTask(
	ctx context.Context, cc containerd.Container, cb func(t containerd.Task) error) error {
	t, err := cc.Task(ctx, func(_ *cio.FIFOSet) (cio.IO, error) {
		return cio.NullIO("")
	})
	if err == nil {
		// todo: check task status
		_ = c.killTask(ctx, t)
	} else if !errors.Is(err, errdefs.ErrNotFound) {
		return err
	}

	t, err = cc.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return err
	}
	defer func() {
		// todo: check task status
		_ = c.killTask(ctx, t)
	}()

	// make sure we wait before calling start
	_, err = t.Wait(ctx)
	if err != nil {
		fmt.Println(err)
	}

	// call start on the task to execute the server
	if err := t.Start(ctx); err != nil {
		return err
	}

	err = cb(t)
	return err
}
