package tarus_store

import (
	"context"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"go.etcd.io/bbolt"
	"os"
	"testing"
)

func TestJudgeSessionStore(t *testing.T) {
	b, err := bbolt.Open("./test.db", os.FileMode(0644), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = b.Close()
		_ = os.Remove("./test.db")
	}()

	j := NewJudgeSessionStore(NewDB(b))
	err = j.SetJudgeSession(context.Background(), []byte("1"), &tarus.OCIJudgeSession{
		CommitStatus: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	o, err := j.GetJudgeSession(context.Background(), []byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(o)
}
