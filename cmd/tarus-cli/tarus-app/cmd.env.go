package tarus_app

import (
	container_build_golang "github.com/Myriad-Dreamin/tarus/containers/golang"
	"github.com/urfave/cli"
)

var commandEnvBuild = Command{
	Name:  "env",
	Usage: "environment management",
	Subcommands: []cli.Command{
		{
			Name:   "build",
			Usage:  "build container environment",
			Action: actEnvBuild,
		},
	},
}

func actEnvBuild(c *Client, args *cli.Context) error {
	container_build_golang.Build()
	return nil
}
