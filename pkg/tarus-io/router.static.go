package tarus_io

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

type StaticRouter struct{}

func (s *StaticRouter) MakeIOChannel(iop string) (ChannelFactory, error) {
	switch iop {
	case "":
	case "std":
		return s.RouteStatic, nil
	case "memory":
		return s.RouteMemory, nil
	default:
		return nil, errors.Wrapf(errdefs.ErrNotFound, "io provider type")
	}
	if iop == "" {
	}

	return nil, errdefs.ErrNotFound
}

func (s *StaticRouter) RouteStatic(inp, oup string) (Factory, error) {
	return NewStd(bytes.NewReader(nil), os.Stdout, os.Stderr), nil
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

type matcher struct {
	r   io.Reader
	pos int64
}

func (m *matcher) GetJudgeResult() (string, error) {
	q := fmt.Sprintf("matched: %v", m.pos)
	// todo: read again
	return q, nil
}

func (m *matcher) Write(p []byte) (n int, err error) {
	var x = make([]byte, len(p))
	for len(p) > 0 {
		// todo: read cache
		if n2, err2 := m.r.Read(x); err2 != nil {
			n += m.comp(p, x[:n2])
			err = err2
			break
		} else if n3 := m.comp(p, x[:n2]); n3 < n2 {
			n += n3
			err = fmt.Errorf("not matched at position %v", m.pos)
			break
		} else {
			p = p[n2:]
			x = x[n2:]
			n += n2
		}
	}

	m.pos += int64(n)
	return
}

func (m matcher) comp(l []byte, r []byte) int {
	// fmt.Printf("matching: %s %s\n", hex.EncodeToString(l), hex.EncodeToString(r))
	n := len(l)
	if n > len(r) {
		n = len(r)
	}
	for i := 0; i < n; i++ {
		if l[i] != r[i] {
			return i
		}
	}

	return n
}

func match(r io.Reader) io.Writer {
	return &matcher{r: r, pos: 0}
}

func (s *StaticRouter) RouteMemory(inp, oup string) (Factory, error) {
	r, err := getHexBuffer(inp)
	if err != nil {
		return nil, err
	}
	r2, err := getHexBuffer(oup)
	if err != nil {
		return nil, err
	}

	return NewStd(r, match(r2), os.Stderr), nil
}

var Statics = &StaticRouter{}
