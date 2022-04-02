package tarus_app

import (
	"context"
	"encoding/base64"
	"fmt"
	hr_bytes "github.com/Myriad-Dreamin/tarus/pkg/hr-bytes"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	"github.com/k0kubun/pp/v3"
	"github.com/urfave/cli"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var commandSubmit = Command{
	Name:   "submit",
	Usage:  "judge submission",
	Action: actSubmit,
	Flags: []cli.Flag{
		appFlagDriver,
		cli.StringFlag{
			Name:     "submission, s",
			Usage:    "source code or binary program",
			Required: true,
		},
		cli.StringFlag{
			Name:  "image",
			Usage: "use container image",
			Value: "docker.io/library/ubuntu:20.04",
		},
	},
}.WithInitService().WithInitDriver()

func actSubmit(c *Client, args *cli.Context) error {
	var (
		err error
		ctx = context.Background()
		svc = tarus_judge.NewClientAdaptor(c.grpcClient)

		codePath = args.String("submission")
		imageId  = args.String("image")
	)

	desc, err := c.Driver.CreateJudgeRequest(ctx)
	if err != nil {
		return err
	}

	binTarget, err := filepath.Abs(codePath)
	if err != nil {
		return err
	}

	resp, err := tarus_judge.TransientJudge(svc, ctx, &tarus_judge.TransientJudgeRequest{
		MakeJudgeRequest: desc,
		ImageId:          imageId,
		BinTarget:        binTarget,
	})
	if err != nil {
		return err
	}

	type JudgeResult struct {
		Index  string
		Status string
	}

	var results []JudgeResult
	for i := range resp.Items {
		results = append(results, JudgeResult{
			Index: strconv.Itoa(i) + ":" + base64.RawURLEncoding.EncodeToString(resp.Items[i].JudgeKey),
			Status: strings.Join([]string{
				resp.Items[i].Status.String(),
				humanTime(resp.Items[i].TimeUse).String(),
				humanTime(resp.Items[i].TimeUseHard).String(),
				humanMemory(resp.Items[i].MemoryUse).String(),
			}, "/"),
		})
	}

	fmt.Printf("Response: ")
	_, err = pp.Println(results)
	return err
}

func humanTime(t int64) time.Duration {
	return (time.Duration(t) * time.Nanosecond) / time.Millisecond * time.Millisecond
}

func humanMemory(t int64) hr_bytes.Byte {
	return hr_bytes.Byte(t)
}
