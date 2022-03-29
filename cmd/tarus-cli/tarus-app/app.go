package tarus_app

import (
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"os"
	"path/filepath"
)

type ActionFunc = func(c *Client, args *cli.Context) error

type Client struct {
	grpcConn   *grpc.ClientConn
	grpcClient tarus.JudgeServiceClient
}

func New() *cli.App {
	app := cli.NewApp()
	app.Name = "tarus-cli"
	app.Usage = "Cli for Online Judge Engine -- tarus."
	app.Description = app.Usage
	app.Version = "v0.0.0"
	app.EnableBashCompletion = true

	var tarusCommands = []cli.Command{
		commandStatus,
		commandService,
	}

	h, _ := os.UserHomeDir()
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output in logs",
		},
		cli.StringFlag{
			Name:   "address, a",
			Usage:  "address for tarus service's GRPC server",
			Value:  filepath.Join(h, ".config/tarus/service.sock"),
			EnvVar: "TARUS_SERVICE_ADDRESS",
		},
	}
	var c = new(Client)
	app.Commands = append(app.Commands, c.inject(tarusCommands)...)
	app.Before = func(args *cli.Context) error {
		if args.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		args.App.Metadata["$client"] = c
		return nil
	}

	return app
}

func (c *Client) inject(commands []cli.Command) []cli.Command {
	if len(commands) == 0 {
		return commands
	}
	for i := range commands {
		commands[i].Subcommands = c.inject(commands[i].Subcommands)
		a := commands[i].Action.(ActionFunc)
		commands[i].Action = func(args *cli.Context) error {
			return a(c, args)
		}
	}

	return commands
}

func init() {
	pp.SetColorScheme(pp.ColorScheme{
		StructName:      pp.White,
		FieldName:       pp.Blue | pp.Bold,
		Bool:            pp.Yellow,
		Integer:         pp.Yellow,
		Nil:             pp.Yellow,
		Float:           pp.Yellow,
		String:          pp.Green,
		StringQuotation: pp.Green,
	})
}
