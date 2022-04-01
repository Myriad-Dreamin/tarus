package domjudge_driver

import (
	"context"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	tarus_driver "github.com/Myriad-Dreamin/tarus/pkg/tarus-driver"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
)

type Driver struct {
	Problem string
}

func New() (*Driver, error) {
	return &Driver{}, nil
}

func (d *Driver) CreateJudgeRequest(ctx context.Context) (*tarus.MakeJudgeRequest, error) {
	if len(d.Problem) == 0 {
		return nil, errors.Wrap(errdefs.ErrInvalidArgument, "problem option required")
	}

	return CreateLocalJudgeRequest(d.Problem)
}

func init() {
	tarus_driver.Register(&tarus_driver.Registration{
		Id: "domjudge",
		Init: func(args *tarus_driver.InitContext) (tarus_driver.Driver, error) {
			d, err := New()
			if err != nil {
				return nil, err
			}
			d.Problem = args.Arguments["problem"]

			return d, nil
		},
	})
}
