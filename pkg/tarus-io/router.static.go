package tarus_io

import (
	"bytes"
	"encoding/hex"
	native_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-checker/native"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type StaticRouter struct{}

func (s *StaticRouter) MakeIOChannel(iop string) (ChannelFactory, error) {
	switch strings.SplitN(iop, ",", 2)[0] {
	case "":
	case "memory":
		return s.RouteMemory(iop), nil
	default:
		return nil, errors.Wrapf(errdefs.ErrNotFound, "io provider type")
	}
	if iop == "" {
	}

	return nil, errdefs.ErrNotFound
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

func (s *StaticRouter) RouteMemory(_ string) ChannelFactory {
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
