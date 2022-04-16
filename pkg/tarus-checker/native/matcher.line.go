package native_judge

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"io"
)

type LineMatcher struct {
	judgeResult   bytes.Buffer
	status        tarus.JudgeStatus
	ErrBuf        bytes.Buffer
	r             *bufio.Scanner
	lineBuf       *bytes.Buffer
	limit         int64
	linePos       int64
	outLines      int64
	ansLines      int64
	caseSensitive bool
}

func (m *LineMatcher) GetJudgeResult() ([]byte, error) {
	if m.status == tarus.JudgeStatus_Unknown {
		if m.lineBuf.Len() != 0 {
			m.outLines += 1
			m.CompLine(m.lineBuf.Bytes())
		}
		for m.r.Scan() {
			m.ansLines += 1
		}

		if m.status == tarus.JudgeStatus_Unknown {
			if m.ansLines == m.outLines {
				m.status = tarus.JudgeStatus_Accepted
				_, _ = fmt.Fprintf(&m.judgeResult, "Accepted %d lines", m.outLines)
			} else if m.ansLines < m.outLines {
				m.status = tarus.JudgeStatus_OutputLimitExceed
				_, _ = fmt.Fprintf(&m.judgeResult, "Output is shorter than answer - expected %d lines but found %d lines", m.ansLines, m.outLines)
			} else {
				m.status = tarus.JudgeStatus_WrongAnswer
				_, _ = fmt.Fprintf(&m.judgeResult, "Output is shorter than answer - expected %d lines but found %d lines", m.ansLines, m.outLines)
			}
		}
	}

	m.r = nil
	m.lineBuf = nil
	b := m.ErrBuf.String()
	var q string
	if len(b) != 0 {
		q = fmt.Sprintf("errout: %q\nres: %q", b, m.judgeResult.String())
	}

	// todo: read again
	return []byte(q), nil
}

func (m *LineMatcher) CompLine(i []byte) {
	if m.status != tarus.JudgeStatus_Unknown {
		return
	}

	var equal = false
	if !m.r.Scan() {
		m.status = tarus.JudgeStatus_OutputLimitExceed
		for m.r.Scan() {
			m.ansLines += 1
		}
		return
	}
	m.ansLines += 1
	j := bytes.TrimSpace(m.r.Bytes())
	i = bytes.TrimSpace(i)
	if m.caseSensitive {
		equal = bytes.Equal(i, j)
	} else {
		equal = bytes.Equal(bytes.ToLower(i), bytes.ToLower(j))
	}

	if !equal {
		for m.r.Scan() {
			m.ansLines += 1
		}
		m.status = tarus.JudgeStatus_WrongAnswer
		_, _ = fmt.Fprintf(&m.judgeResult, "at line %d differs - expected: %q, found: %q", m.outLines, string(j), string(i))
		return
	}
}

func (m *LineMatcher) GetJudgeStatus(_ []byte) (tarus.JudgeStatus, error) {
	return m.status, nil
}

func (m *LineMatcher) Write(p []byte) (n int, err error) {
	if m.lineBuf == nil {
		return 0, io.EOF
	}

	n, err = m.lineBuf.Write(p)
	if tarus.JudgeStatus_WrongAnswer == m.status {
		m.lineBuf.Reset()
		return
	}

	b := m.lineBuf.Bytes()
	var lineConsume = int64(-1)
	for i, l := m.linePos, int64(len(b)); i < l; i++ {
		if b[i] == '\n' {
			m.outLines += 1
			m.CompLine(b[lineConsume+1 : i])
			lineConsume = i
		}
	}

	if lineConsume != -1 {
		m.lineBuf.Next(int(lineConsume + 1))
		m.linePos = int64(len(b)) - (lineConsume + 1)
	}
	return
}

func LineMatch(r io.Reader) *LineMatcher {
	return &LineMatcher{r: bufio.NewScanner(r), lineBuf: bytes.NewBuffer(nil)}
}
