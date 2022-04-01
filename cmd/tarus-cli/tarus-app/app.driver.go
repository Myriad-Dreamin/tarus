package tarus_app

import (
	tarus_driver "github.com/Myriad-Dreamin/tarus/pkg/tarus-driver"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"strings"
)

var appFlagDriver = cli.StringFlag{
	Name:  "driver",
	Usage: "judge driver",
}

func (c *Client) initDriver(s string) error {
	if len(s) == 0 {
		return nil
	}

	if c.Driver != nil {
		return errors.New("multiple driver arguments provided")
	}

	var initCtx tarus_driver.InitContext

	id_rest := strings.Split(s, ",")
	id := id_rest[0]

	initCtx.Arguments = make(map[string]string)
	for _, v := range id_rest[1:] {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) > 1 {
			v = kv[1]
		} else {
			v = ""
		}
		initCtx.Arguments[kv[0]] = v
	}

	regs, err := tarus_driver.FindById(id)
	if err != nil {
		return err
	}
	if len(regs) == 0 {
		return errdefs.ErrNotFound
	}
	if len(regs) > 1 {
		return errors.Wrapf(errdefs.ErrInvalidArgument, "multiple driver with id found, maybe bug")
	}

	reg := regs[0]

	d, err := reg.Init(&initCtx)
	if err != nil {
		return err
	}

	c.Driver = d
	return nil
}
