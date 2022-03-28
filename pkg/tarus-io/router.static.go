package tarus_io

import (
	"bytes"
	"encoding/hex"
	native_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-checker/native"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"strings"
)

type StaticRouter struct{}

func (s *StaticRouter) MakeIOChannel(iop string) (ChannelFactory, error) {
	u, err := url.Parse(iop)
	if err != nil {
		return nil, err
	}
	protocol := u.Scheme
	if len(protocol) == 0 {
		protocol = u.Path
	}

	switch protocol {
	case "", "memory":
		return s.RouteMemory(u), nil
	default:
		return nil, errors.Wrapf(errdefs.ErrNotFound, "io provider type")
	}
}

func getHexBuffer(s string) (*bytes.Buffer, error) {
	if strings.HasPrefix(s, "hexbytes://") {
		s = strings.TrimPrefix(s, "hexbytes://")
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid hex format when decoding hex url")
		}

		return bytes.NewBuffer(b), nil
	}

	return nil, errors.Wrapf(errdefs.ErrNotFound, "memory protocol")
}

var Statics = &StaticRouter{}

func (s *StaticRouter) RouteMemory(_ *url.URL) ChannelFactory {
	return func(inp, oup string) (Factory, error) {
		r, err := getHexBuffer(inp)
		if err != nil {
			return nil, err
		}
		r2, err := getHexBuffer(oup)
		if err != nil {
			return nil, err
		}

		return NewStd(r, native_judge.StrictMatch(r2), os.Stderr), nil
	}
}
