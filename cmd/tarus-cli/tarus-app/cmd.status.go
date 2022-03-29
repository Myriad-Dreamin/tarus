package tarus_app

import "github.com/urfave/cli"

var commandStatus = Command{
	Name:   "status",
	Usage:  "check directory resource status",
	Action: actStatus,
}

func actStatus(c *Client, args *cli.Context) error {
	return nil
}
