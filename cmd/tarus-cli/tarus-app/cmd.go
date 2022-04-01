package tarus_app

import "github.com/urfave/cli"

type ActionFunc = func(c *Client, args *cli.Context) error
type Command cli.Command

func toCliCommands(commands []Command) (cc []cli.Command) {
	for i := range commands {
		cc = append(cc, cli.Command(commands[i]))
	}
	return cc
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

func (c Command) WithInitDriver() Command {
	c.Before = hookBefore(c.Before, func(args *cli.Context) error {
		c := args.App.Metadata["$client"].(*Client)

		if err := c.initDriver(args.GlobalString("driver")); err != nil {
			return err
		}
		if err := c.initDriver(args.String("driver")); err != nil {
			return err
		}
		return nil
	})
	return c
}
