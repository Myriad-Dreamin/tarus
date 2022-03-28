package tarus_io

import (
	"fmt"
	"os"
	"testing"
)

func TestStaticRouter_MakeIOChannel(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		iop string
	}
	tests := []struct {
		name    string
		args    args
		want    ChannelFactory
		wantErr bool
	}{
		{
			name: "memory_router",
			args: args{
				iop: "",
			},
			wantErr: false,
		},
		{
			name: "memory_router",
			args: args{
				iop: "memory",
			},
			wantErr: false,
		},
		{
			name: "filesystem_router",
			args: args{
				iop: fmt.Sprintf("file://%s/fuzzers/corpora/domjudge/bapc2019-A", cwd),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StaticRouter{}
			_, err = s.MakeIOChannel(tt.args.iop)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeIOChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestStaticRouterFilesystemProvider(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, ts := range []struct {
		iop      string
		wantErr  bool
		wantErr2 bool
	}{
		{"file:fuzzers/corpora/domjudge/bapc2019-A", true, false},
		{fmt.Sprintf("file://%s/fuzzers/corpora/domjudge/bapc2019-A", cwd), false, false},
	} {
		i, err := Statics.MakeIOChannel(ts.iop)
		if (err != nil) != ts.wantErr {
			t.Fatal(err)
		}
		if err != nil {
			continue
		}
		e, err := i("data/sample/00_sample.ans", "data/sample/00_sample.ans")
		if (err != nil) != ts.wantErr2 {
			t.Fatal(err)
		}
		if err != nil {
			continue
		}
		_, _ = e.GetJudgeResult()
	}
}
