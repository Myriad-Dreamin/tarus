package tarus_judge

import (
	"context"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func NewClientAdaptor(c tarus.JudgeServiceClient) tarus.JudgeServiceServer {
	return &ClientAdaptor{c: c}
}

type ClientAdaptor struct {
	tarus.UnimplementedJudgeServiceServer
	c tarus.JudgeServiceClient
}

func (c *ClientAdaptor) Handshake(ctx context.Context, request *tarus.HandshakeRequest) (*tarus.HandshakeResponse, error) {
	return c.c.Handshake(ctx, request)
}

func (c *ClientAdaptor) CreateContainer(ctx context.Context, request *tarus.CreateContainerRequest) (*emptypb.Empty, error) {
	return c.c.CreateContainer(ctx, request)
}

func (c *ClientAdaptor) BundleContainer(ctx context.Context, request *tarus.BundleContainerRequest) (*emptypb.Empty, error) {
	return c.c.BundleContainer(ctx, request)
}

func (c *ClientAdaptor) RemoveContainer(ctx context.Context, request *tarus.RemoveContainerRequest) (*emptypb.Empty, error) {
	return c.c.RemoveContainer(ctx, request)
}

func (c *ClientAdaptor) CloneContainer(ctx context.Context, request *tarus.CloneContainerRequest) (*emptypb.Empty, error) {
	return c.c.CloneContainer(ctx, request)
}

func (c *ClientAdaptor) CheckContainer(ctx context.Context, request *tarus.CheckContainerRequest) (*emptypb.Empty, error) {
	return c.c.CheckContainer(ctx, request)
}

func (c *ClientAdaptor) CopyFile(ctx context.Context, request *tarus.CopyFileRequest) (*emptypb.Empty, error) {
	return c.c.CopyFile(ctx, request)
}

func (c *ClientAdaptor) CompileProgram(ctx context.Context, request *tarus.CompileProgramRequest) (*emptypb.Empty, error) {
	return c.c.CompileProgram(ctx, request)
}

func (c *ClientAdaptor) MakeJudge(ctx context.Context, request *tarus.MakeJudgeRequest) (*tarus.MakeJudgeResponse, error) {
	return c.c.MakeJudge(ctx, request)
}

func (c *ClientAdaptor) QueryJudge(ctx context.Context, request *tarus.QueryJudgeRequest) (*tarus.QueryJudgeResponse, error) {
	return c.c.QueryJudge(ctx, request)
}
