package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	oci_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge/oci"
)

func hexUrl(s string) string {
	return fmt.Sprintf("hexbytes://%s", hex.EncodeToString([]byte(s)))
}

func echoTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if err := client.TransientJudge(ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/echo_test",
		Items: []*tarus.MakeJudgeItem{
			{
				JudgeKey:   []byte("001"),
				IoProvider: "memory",
				InputUrl:   hexUrl(``),
				OutputUrl:  hexUrl(`hello world`),
			},
		},
	}); err != nil {
		panic(err)
	}
}

func sleepTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if err := client.TransientJudge(ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/sleep_test",
		Items: []*tarus.MakeJudgeItem{
			{
				JudgeKey:   []byte("001"),
				IoProvider: "memory",
				InputUrl:   hexUrl(``),
				OutputUrl:  hexUrl(`hello world`),
			},
		},
	}); err != nil {
		panic(err)
	}
}

func sleepHardTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if err := client.TransientJudge(ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/sleep_hard_test",
		Items: []*tarus.MakeJudgeItem{
			{
				JudgeKey:   []byte("001"),
				IoProvider: "memory",
				InputUrl:   hexUrl(``),
				OutputUrl:  hexUrl(`hello world`),
			},
			{
				JudgeKey:   []byte("001"),
				IoProvider: "memory",
				InputUrl:   hexUrl(``),
				OutputUrl:  hexUrl(`hello world`),
			},
		},
	}); err != nil {
		panic(err)
	}
}

func ioTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if err := client.TransientJudge(ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/io_test",
		Items: []*tarus.MakeJudgeItem{
			{
				JudgeKey:   []byte("001"),
				IoProvider: "memory",
				InputUrl:   hexUrl("1 2\n"),
				OutputUrl:  hexUrl(`3`),
			},
		},
	}); err != nil {
		panic(err)
	}
}

func inputTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if err := client.TransientJudge(ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/echo_input_test",
		Items: []*tarus.MakeJudgeItem{
			{
				JudgeKey:   []byte("001"),
				IoProvider: "memory",
				InputUrl:   hexUrl("yes\n"),
				OutputUrl:  hexUrl(`yes`),
			},
		},
	}); err != nil {
		panic(err)
	}
}

func main() {
	client, err := oci_judge.NewContainerdServer()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = client.Close()
	}()

	ctx := context.Background()

	if err = client.ImportOCIArchive(ctx, "ubuntu.tar"); err != nil {
		panic(err)
	}

	// echoTest(client, ctx)
	// sleepTest(client, ctx)
	sleepHardTest(client, ctx)
	//ioTest(client, ctx)
	//inputTest(client, ctx)
}
