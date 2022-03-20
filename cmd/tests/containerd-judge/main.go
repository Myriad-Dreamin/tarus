package main

import (
	"context"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	oci_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge/oci"
)

func main() {
	client, err := oci_judge.NewContainerdServer()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = client.Close()
	}()

	ctx := context.Background()

	if err = client.ImportOCIArchive(ctx, "busybox.tar"); err != nil {
		panic(err)
	}

	if err = client.TransientJudge(ctx, &tarus_judge.TransientJudgeRequest{
		ImageId: "docker.io/library/busybox:1.35.0",
	}); err != nil {
		panic(err)
	}
}
