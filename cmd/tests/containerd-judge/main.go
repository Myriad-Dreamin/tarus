package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	hr_bytes "github.com/Myriad-Dreamin/tarus/pkg/hr-bytes"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	oci_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge/oci"
	"syscall"
)

func hexUrl(s string) string {
	return fmt.Sprintf("hexbytes://%s", hex.EncodeToString([]byte(s)))
}

func echoTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if _, err := tarus_judge.TransientJudge(client, ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/echo_test",
		MakeJudgeRequest: &tarus.MakeJudgeRequest{
			Testcases: []*tarus.JudgeTestcase{
				{
					JudgeKey:   []byte("001"),
					IoProvider: "memory",
					Input:      hexUrl(``),
					Answer:     hexUrl(`hello world`),
				},
			},
		},
	}); err != nil {
		panic(err)
	}
}

func sleepTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if _, err := tarus_judge.TransientJudge(client, ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/sleep_test",
		MakeJudgeRequest: &tarus.MakeJudgeRequest{
			Testcases: []*tarus.JudgeTestcase{
				{
					JudgeKey:   []byte("001"),
					IoProvider: "memory",
					Input:      hexUrl(``),
					Answer:     hexUrl(`hello world`),
				},
			},
		},
	}); err != nil {
		panic(err)
	}
}

func sleepHardTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if _, err := tarus_judge.TransientJudge(client, ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/sleep_hard_test",
		MakeJudgeRequest: &tarus.MakeJudgeRequest{
			Testcases: []*tarus.JudgeTestcase{
				{
					JudgeKey:   []byte("001"),
					IoProvider: "memory",
					Input:      hexUrl(``),
					Answer:     hexUrl(`hello world`),
				},
				{
					JudgeKey:   []byte("002"),
					IoProvider: "memory",
					Input:      hexUrl(``),
					Answer:     hexUrl(`hello world`),
				},
			},
		},
	}); err != nil {
		panic(err)
	}
}

func ioTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if _, err := tarus_judge.TransientJudge(client, ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/io_test",
		MakeJudgeRequest: &tarus.MakeJudgeRequest{
			Testcases: []*tarus.JudgeTestcase{
				{
					JudgeKey:   []byte("001"),
					IoProvider: "memory",
					Input:      hexUrl("1 2\n"),
					Answer:     hexUrl(`3`),
				},
			},
		},
	}); err != nil {
		panic(err)
	}
}

func inputTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	if _, err := tarus_judge.TransientJudge(client, ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/echo_input_test",
		MakeJudgeRequest: &tarus.MakeJudgeRequest{
			Testcases: []*tarus.JudgeTestcase{
				{
					JudgeKey:   []byte("001"),
					IoProvider: "memory",
					Input:      hexUrl("yes\n"),
					Answer:     hexUrl(`yes`),
				},
			},
		},
	}); err != nil {
		panic(err)
	}
}

func statusTest(client *oci_judge.ContainerdJudgeServiceServer, ctx context.Context) {
	testcases := []*tarus.JudgeTestcase{
		{
			JudgeKey:   []byte("001"),
			IoProvider: "memory",
			Input:      hexUrl("exit\n"),
			Answer:     hexUrl(``),
		},
		{
			JudgeKey:   []byte("002"),
			IoProvider: "memory",
			Input:      hexUrl("abort\n"),
			Answer:     hexUrl(``),
		},
		{
			JudgeKey:   []byte("003"),
			IoProvider: "memory",
			Input:      hexUrl("null\n"),
			Answer:     hexUrl(``),
		},
		{
			JudgeKey:   []byte("004"),
			IoProvider: "memory",
			Input:      hexUrl("fpe\n"),
			Answer:     hexUrl(``),
		},
		// safe usage
		{
			JudgeKey:   []byte("005"),
			IoProvider: "memory",
			Input:      hexUrl(fmt.Sprintf("memory=%d\n", int64(384*hr_bytes.MB))),
			Answer:     hexUrl(``),
		},
		// mle, 384 + 128 >= 512MB
		{
			JudgeKey:   []byte("006"),
			IoProvider: "memory",
			Input:      hexUrl(fmt.Sprintf("memory=%d,memory=%d\n", int64(384*hr_bytes.MB), int64(128*hr_bytes.MB))),
			Answer:     hexUrl(``),
		},
		// ok, unused virtual memory are not traced by page usage
		{
			JudgeKey:   []byte("007"),
			IoProvider: "memory",
			Input:      hexUrl(fmt.Sprintf("virt_memory=%d\n", int64(384*hr_bytes.MB))),
			Answer:     hexUrl(``),
		},
		// early rejected by system, gets runtime error
		{
			JudgeKey:   []byte("008"),
			IoProvider: "memory",
			Input:      hexUrl(fmt.Sprintf("virt_memory=%d\n", int64(512*hr_bytes.MB))),
			Answer:     hexUrl(``),
		},
	}

	for i := 1; i < 0x20; i++ {
		testcases = append(testcases, &tarus.JudgeTestcase{
			JudgeKey:   []byte(fmt.Sprintf("signal:%v", syscall.Signal(i).String())),
			IoProvider: "memory",
			Input:      hexUrl(fmt.Sprintf("signal=%v\n", i)),
			Answer:     hexUrl(``),
		})
		testcases = append(testcases, &tarus.JudgeTestcase{
			JudgeKey:   []byte(fmt.Sprintf("signal:%v (exit code detection)", syscall.Signal(i).String())),
			IoProvider: "memory",
			Input:      hexUrl(fmt.Sprintf("exit=%v\n", i+128)),
			Answer:     hexUrl(``),
		})
	}

	statusCodes, err := tarus_judge.TransientJudge(client, ctx, &tarus_judge.TransientJudgeRequest{
		ImageId:   "docker.io/library/ubuntu:20.04",
		BinTarget: "/workdir/mock_solver",
		MakeJudgeRequest: &tarus.MakeJudgeRequest{
			Testcases: testcases,
		},
	})
	if err != nil {
		panic(err)
	}
	for _, status := range statusCodes.Items {
		fmt.Println(status)
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
	// sleepHardTest(client, ctx)
	// ioTest(client, ctx)
	// inputTest(client, ctx)
	// statusTest(client, ctx)
}
