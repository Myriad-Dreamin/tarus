package tarus_app

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

func (c *Client) connectToGrpcService(args *cli.Context) (*grpc.ClientConn, error) {
	var address = args.GlobalString("address")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var opts = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("unix://%s", address), opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to dial %q", address)
	}
	return conn, nil
}
