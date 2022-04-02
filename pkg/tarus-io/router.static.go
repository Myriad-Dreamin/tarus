package tarus_io

import (
	"bytes"
	"encoding/hex"
	native_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-checker/native"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"path/filepath"
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
	case "file":
		return s.RouteFilesystem(u)
	default:
		return nil, errors.Wrapf(errdefs.ErrNotFound, "io provider type %s", protocol)
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

		// todo: limit output buffer
		m := native_judge.StrictMatch(r2)
		return NewStd(r, m, &m.ErrBuf), nil
	}
}

func (s *StaticRouter) RouteFilesystem(fs *url.URL) (ChannelFactory, error) {
	p := fs.Path
	if !strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "\\") {
		return nil, errors.Wrapf(errdefs.ErrInvalidArgument, "want absolute path for filesystem io provider: %q", p)
	}

	return func(inp, oup string) (Factory, error) {
		r, err := os.Open(filepath.Join(p, inp))
		if err != nil {
			return nil, err
		}
		r2, err := os.Open(filepath.Join(p, oup))
		if err != nil {
			_ = r.Close()
			return nil, err
		}

		m := native_judge.StrictMatch(r2)
		return NewStd(r, m, &m.ErrBuf), nil
	}, nil
}
