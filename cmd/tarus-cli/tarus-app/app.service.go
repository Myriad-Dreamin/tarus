package tarus_app

import (
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

var appFlagServiceAddr = cli.StringFlag{
	Name:   "address, a",
	Usage:  "address for tarus service",
	Value:  "",
	EnvVar: "TARUS_SERVICE_ADDRESS",
}

func init() {
	appFlagServiceAddr.Value = ".config/tarus/service.sock"
	if h, _ := os.UserHomeDir(); len(h) != 0 {
		appFlagServiceAddr.Value = filepath.Join(h, appFlagServiceAddr.Value)
	}
}

func (c *Client) initService(args *cli.Context) (err error) {
	c.grpcConn, err = c.connectToGrpcService(args)
	if err == nil {
		c.grpcClient = tarus.NewJudgeServiceClient(c.grpcConn)
		c.closers = append(c.closers, c.grpcConn)
	}
	return
}
