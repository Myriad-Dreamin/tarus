package tarus_app

import (
	"context"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/k0kubun/pp/v3"
	"github.com/urfave/cli"
)

var commandService = Command{
	Name:   "service",
	Usage:  "service operations",
	Action: actServiceStatus,
	Subcommands: []cli.Command{
		{
			Name:   "status",
			Usage:  "check service status",
			Action: actServiceStatus,
		},
	},
}.WithInitService()

func actServiceStatus(c *Client, args *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.grpcClient.Handshake(ctx, &tarus.HandshakeRequest{
		ApiVersion: []byte(args.App.Version),
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
