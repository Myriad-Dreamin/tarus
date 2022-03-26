package native_judge

import (
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"io"
)

type StrictMatcher struct {
	r   io.Reader
	pos int64
}

func (m StrictMatcher) comp(l []byte, r []byte) int {
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

func StrictMatch(r io.Reader) io.Writer {
	return &StrictMatcher{r: r, pos: 0}
}

func (m *StrictMatcher) GetJudgeResult() ([]byte, error) {
	q := fmt.Sprintf("matched: %v", m.pos)
	// todo: read again
	return []byte(q), nil
}

func (m *StrictMatcher) GetJudgeStatus(_ []byte) (tarus.JudgeStatus, error) {
	return tarus.JudgeStatus_Accepted, nil
}

func (m *StrictMatcher) Write(p []byte) (n int, err error) {
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
