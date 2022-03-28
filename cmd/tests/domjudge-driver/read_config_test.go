package main

import (
	"fmt"
	domjudge_driver "github.com/Myriad-Dreamin/tarus/pkg/tarus-driver/domjudge"
	"testing"
)

func TestDomJudgeRead(t *testing.T) {
	desc, err := domjudge_driver.CreateLocalJudgeRequest("fuzzers/corpora/domjudge/bapc2019-A")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(desc.IoProvider)
	for _, c := range desc.Testcases {
		fmt.Println(c)
	}
}
