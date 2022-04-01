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
	},
}.WithInitService().WithInitDriver()

func actSubmit(c *Client, _ *cli.Context) error {
	var (
		err error
		ctx = context.Background()
		svc = tarus_judge.NewClientAdaptor(c.grpcClient)
	)

	desc, err := c.Driver.CreateJudgeRequest(ctx)
	if err != nil {
		return err
	}

	binTarget, _ := filepath.Abs("data/workdir-judge-engine0/bapc2019_a_accepted_test")

	resp, err := tarus_judge.TransientJudge(svc, ctx, &tarus_judge.TransientJudgeRequest{
		MakeJudgeRequest: desc,
		ImageId:          "docker.io/library/ubuntu:20.04",
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
