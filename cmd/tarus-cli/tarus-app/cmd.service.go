package tarus_app

import (
	"context"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/k0kubun/pp/v3"
	"github.com/urfave/cli"
)

var commandService = cli.Command{
	Name:  "service",
	Usage: "service operations",
	Before: func(args *cli.Context) error {
		c := args.App.Metadata["$client"].(*Client)
		return c.initService(args)
	},
	After: func(args *cli.Context) error {
		c := args.App.Metadata["$client"].(*Client)
		if c.grpcConn != nil {
			return c.grpcConn.Close()
		}
		return nil
	},
	Action: actServiceStatus,
	Subcommands: []cli.Command{
		{
			Name:   "status",
			Usage:  "check service status",
			Action: actServiceStatus,
		},
	},
}

func actServiceStatus(c *Client, _ *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.grpcClient.Handshake(ctx, &tarus.HandshakeRequest{
		ApiVersion: []byte("v0.0.0"),
	})
	if err != nil {
		return err
	}

	type ServiceStatus struct {
		ApiVersion      string
		JudgeStatusHash string
		ImplementedApis []string
	}
	fmt.Printf("Response: ")
	_, err = pp.Println(ServiceStatus{
		ApiVersion:      string(resp.ApiVersion),
		JudgeStatusHash: resp.JudgeStatusHash,
		ImplementedApis: resp.ImplementedApis,
	})
	return err
}

func (c *Client) initService(args *cli.Context) (err error) {
	c.grpcConn, err = c.connectToGrpcService(args)
	if err == nil {
		c.grpcClient = tarus.NewJudgeServiceClient(c.grpcConn)
	}
	return
}
