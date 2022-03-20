package oci_judge

import (
	context "context"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"os"
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
	container, err := c.client.NewContainer(
		ctx,
		fixedContainerId,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(fixedContainerSnapshotId, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
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

			time.Sleep(3 * time.Second)
			fmt.Printf("linux container stop\n")
			return nil
		})
	})

}
