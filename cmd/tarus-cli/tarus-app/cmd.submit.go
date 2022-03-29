package tarus_app

import (
	"context"
	"encoding/base64"
	"fmt"
	hr_bytes "github.com/Myriad-Dreamin/tarus/pkg/hr-bytes"
	domjudge_driver "github.com/Myriad-Dreamin/tarus/pkg/tarus-driver/domjudge"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	"github.com/k0kubun/pp/v3"
	"github.com/urfave/cli"
	"strconv"
	"strings"
	"time"
)

var commandSubmit = Command{
	Name:   "submit",
	Usage:  "judge submission",
	Action: actSubmit,
}.WithInitService()

func actSubmit(c *Client, _ *cli.Context) error {
	var err error

	client := c.grpcClient

	desc, err := domjudge_driver.CreateLocalJudgeRequest("fuzzers/corpora/domjudge/bapc2019-A")
	if err != nil {
		return err
	}

	var ctx = context.Background()

	resp, err := tarus_judge.TransientJudge(tarus_judge.NewClientAdaptor(client), ctx, &tarus_judge.TransientJudgeRequest{
		MakeJudgeRequest: desc,
		ImageId:          "docker.io/library/ubuntu:20.04",
		BinTarget:        "/workdir/bapc2019_a_accepted_test",
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
