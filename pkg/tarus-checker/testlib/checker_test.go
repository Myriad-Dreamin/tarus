package testlib_checker

import (
	"bytes"
	"os"
	"testing"
)

var yesNoResult = []byte("ok answer is YES")

func testYesNo(t testing.TB) {
	oup, err := os.OpenFile("cmd/tests/containerd-judge/yes_no_testcases/yes.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	ans, err := os.OpenFile("cmd/tests/containerd-judge/yes_no_testcases/yes.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	c, err := TestlibChecker("./bin/testlib/checkers/yesno", nil, oup, ans)
	if err != nil {
		t.Fatal(err)
	}

	if s, err := c.GetJudgeResult(); err != nil {
		t.Fatal(err)
	} else {
		_ = s
		if !bytes.Equal(yesNoResult, bytes.TrimSpace(s)) {
			t.Errorf("invalid judge result %s", string(s))
		}
	}
}

func TestYesNo(t *testing.T) {
	testYesNo(t)
}

func BenchmarkYesNo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testYesNo(b)
	}
}
