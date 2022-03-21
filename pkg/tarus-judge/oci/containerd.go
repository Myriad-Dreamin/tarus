package oci_judge

import (
	"bytes"
	context "context"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"os"
	"path/filepath"
	"time"
)

var fixedContainerId = "tarus-engine0"

type ContainerdJudgeServiceServer struct {
	tarus.UnimplementedJudgeServiceServer
	client *containerd.Client
}

func NewContainerdServer() (*ContainerdJudgeServiceServer, error) {
	client, err := containerd.New("/run/containerd/containerd.sock",
		containerd.WithDefaultNamespace("tarus"))
	if err != nil {
		return nil, err
	}
	return &ContainerdJudgeServiceServer{
		client: client,
	}, nil
}

func (c *ContainerdJudgeServiceServer) Close() error {
	return c.client.Close()
}

func (c *ContainerdJudgeServiceServer) Handshake(ctx context.Context, request *tarus.HandshakeRequest) (*emptypb.Empty, error) {
	return c.UnimplementedJudgeServiceServer.Handshake(ctx, request)
}

func (c *ContainerdJudgeServiceServer) CopyFile(ctx context.Context, request *tarus.CopyRequest) (*emptypb.Empty, error) {
	return c.UnimplementedJudgeServiceServer.CopyFile(ctx, request)
}

func (c *ContainerdJudgeServiceServer) CreateContainer(ctx context.Context, request *tarus.CreateContainerRequest) (_ *emptypb.Empty, err error) {
	ctx = namespaces.WithNamespace(ctx, "tarus")
	snapshotter := containerd.DefaultSnapshotter
	if err = c.prepareImageOnSnapshotter(ctx, request.ImageId, snapshotter); err != nil {
		return nil, err
	}

	fixedContainerSnapshotId := "tarus-engine-snapshot0"

	if err = c.client.SnapshotService(snapshotter).Remove(ctx, fixedContainerSnapshotId); err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	}
	if err = c.client.ContainerService().Delete(ctx, fixedContainerId); err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	}

	image, err := c.client.GetImage(ctx, request.ImageId)
	if err != nil {
		return nil, err
	}
	mounts := make([]specs.Mount, 0)

	m, err := filepath.Abs("./data/workdir-judge-engine0")
	if err != nil {
		return nil, err
	}
	mounts = append(mounts, specs.Mount{
		// notes: linux only flags
		Type:        "bind",
		Source:      m,
		Destination: "/workdir",
		Options:     []string{"rbind", "ro"},
	})

	container, err := c.client.NewContainer(
		ctx,
		fixedContainerId,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(fixedContainerSnapshotId, image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithMounts(mounts),
			oci.WithProcessCwd("/workdir"),
		),
	)

	if err != nil {
		return nil, err
	}
	if container == nil {
		return nil, status.Error(codes.Internal, "container not created when successfully returning")
	}
	return nil, nil
}

func (c *ContainerdJudgeServiceServer) RemoveContainer(ctx context.Context, request *tarus.RemoveContainerRequest) (*emptypb.Empty, error) {
	ctx = namespaces.WithNamespace(ctx, "tarus")
	cc, err := c.client.LoadContainer(ctx, fixedContainerId)
	if err != nil {
		return nil, err
	}
	err = cc.Delete(ctx, containerd.WithSnapshotCleanup)
	return nil, err
}

func (c *ContainerdJudgeServiceServer) MakeJudge(ctx context.Context, request *tarus.MakeJudgeRequest) (*emptypb.Empty, error) {
	return c.UnimplementedJudgeServiceServer.MakeJudge(ctx, request)
}

func (c *ContainerdJudgeServiceServer) QueryJudge(ctx context.Context, request *tarus.QueryJudgeRequest) (*tarus.QueryJudgeResponse, error) {
	return c.UnimplementedJudgeServiceServer.QueryJudge(ctx, request)
}

func (c *ContainerdJudgeServiceServer) ImportOCIArchive(ctx context.Context, fp string) error {
	ctx = namespaces.WithNamespace(ctx, "tarus")

	f, err := os.OpenFile(fp, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	_, err = c.client.Import(ctx, f)
	_ = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func (c *ContainerdJudgeServiceServer) TransientJudge(rawCtx context.Context, req *tarus_judge.TransientJudgeRequest) error {
	return tarus_judge.WithContainerEnvironment(c, rawCtx, req, func(rawCtx context.Context, req *tarus_judge.TransientJudgeRequest) error {
		ctx := namespaces.WithNamespace(rawCtx, "tarus")
		cc, err := c.client.LoadContainer(ctx, fixedContainerId)
		if err != nil {
			return err
		}
		return c.withFreshTask(ctx, cc, func(t containerd.Task) error {
			fmt.Printf("linux container create successfully\n")
			spec, err := cc.Spec(ctx)
			if err != nil {
				return err
			}

			procOpts := spec.Process
			procOpts.Terminal = false
			procOpts.Args = []string{"/workdir/echo_test"}

			var b = bytes.NewBuffer([]byte{})
			process, err := t.Exec(ctx, "judge_exec", procOpts, cio.NewCreator(cio.WithStreams(bytes.NewReader(nil), b, os.Stderr)))
			if err != nil {
				return err
			}
			defer func() {
				_, _ = process.Delete(ctx)
			}()

			statusC, err := process.Wait(ctx)
			if err != nil {
				return err
			}

			if err := process.Start(ctx); err != nil {
				return err
			}
			defer func() {
				// todo: check process status
				_ = c.killProcess(ctx, process)
			}()

			select {
			case st := <-statusC:
				code, _, err := st.Result()
				if err != nil {
					return err
				}
				fmt.Printf("linux container exit: %v\n", code)
				fmt.Printf("judge output: %v", b.String())
			case <-time.After(time.Second * 3):
				fmt.Printf("linux container timeout stop\n")
			}

			return nil
		})
	})

}
