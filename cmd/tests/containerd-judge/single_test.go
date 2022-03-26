package main

import (
	"context"
	"fmt"
	oci_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge/oci"
	"os"
	"sync/atomic"
	"testing"
)

var client *oci_judge.ContainerdJudgeServiceServer

func init() {
	var err error
	_ = os.Chdir("../../../")

	client, err = oci_judge.NewContainerdServer()
	if err != nil {
		panic(err)
	}
}

func BenchmarkEcho(b *testing.B) {
	var n int32
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.WithValue(context.Background(), "No", atomic.AddInt32(&n, 1))
		for pb.Next() {
			echoTest(client, ctx)
		}
	})
	fmt.Println("perf", oci_judge.PerfTime)
}
