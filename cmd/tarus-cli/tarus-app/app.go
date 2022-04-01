package tarus_app

import (
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"io"
	"os"
	"path/filepath"
)

type ActionFunc = func(c *Client, args *cli.Context) error
type Command cli.Command

type Client struct {
	closers  []io.Closer
	grpcConn *grpc.ClientConn

	grpcClient tarus.JudgeServiceClient
}

func New() *cli.App {
	app := cli.NewApp()
	app.Name = "tarus-cli"
	app.Usage = "Cli for Online Judge Engine -- tarus."
	app.Description = app.Usage
	app.Version = "v0.0.0"
	app.EnableBashCompletion = true

	var tarusCommands = []Command{
		commandStatus,
		commandService,
		commandSubmit,
		commandEnvBuild,
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
	app.Commands = append(app.Commands, c.inject(toCliCommands(tarusCommands))...)
	app.Before = func(args *cli.Context) error {
		if args.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		args.App.Metadata["$client"] = c
		return nil
	}
	app.After = func(args *cli.Context) (err error) {
		c := args.App.Metadata["$client"].(*Client)
		for i := range c.closers {
			err2 := c.closers[i].Close()
			if err2 != nil {
				err = err2
			}
		}
		return
	}

	return app
}

func toCliCommands(commands []Command) (cc []cli.Command) {
	for i := range commands {
		cc = append(cc, cli.Command(commands[i]))
	}
	return cc
}

func (c *Client) inject(commands []cli.Command) (cc []cli.Command) {
	if len(commands) == 0 {
		return commands
	}
	for i := range commands {
		commands[i].Subcommands = c.inject(commands[i].Subcommands)

		if commands[i].Action == nil {
			continue
		}
		a := commands[i].Action.(ActionFunc)
		commands[i].Action = func(args *cli.Context) error {
			return a(c, args)
		}
	}

	return commands
}

func hookBefore(before, afterBefore cli.BeforeFunc) cli.BeforeFunc {
	if before == nil {
		return afterBefore
	}
	return func(args *cli.Context) error {
		if err := before(args); err != nil {
			return err
		}
		return afterBefore(args)
	}
}

func hookAfter(after, beforeAfter cli.AfterFunc) cli.AfterFunc {
	if after == nil {
		return beforeAfter
	}
	return func(args *cli.Context) error {
		if err := beforeAfter(args); err != nil {
			_ = after(args)
			return err
		}
		return after(args)
	}
}

func (c Command) WithInitService() Command {
	c.Before = hookBefore(c.Before, func(args *cli.Context) error {
		c := args.App.Metadata["$client"].(*Client)
		return c.initService(args)
	})
	return c
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
