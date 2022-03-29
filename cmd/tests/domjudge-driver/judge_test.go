package main

import (
	"context"
	domjudge_driver "github.com/Myriad-Dreamin/tarus/pkg/tarus-driver/domjudge"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	oci_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge/oci"
	"testing"
)

func TestDomJudge_BAPC2019A_Accepted(t *testing.T) {
	var err error

	client, err := oci_judge.NewContainerdServer()
	if err != nil {
		t.Fatal(err)
	}

	desc, err := domjudge_driver.CreateLocalJudgeRequest("fuzzers/corpora/domjudge/bapc2019-A")
	if err != nil {
		t.Fatal(err)
	}

	var ctx = context.Background()

	if _, err := tarus_judge.TransientJudge(client, ctx, &tarus_judge.TransientJudgeRequest{
		MakeJudgeRequest: desc,
		ImageId:          "docker.io/library/ubuntu:20.04",
		BinTarget:        "/workdir/bapc2019_a_accepted_test",
	}); err != nil {
		panic(err)
	}
}
