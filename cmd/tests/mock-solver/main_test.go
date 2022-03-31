package main

import "testing"

func TestMainProc(t *testing.T) {
	for _, f := range []string{
		"memory=1",
		"time=1",
		"time=1,memory=1",
	} {
		t.Run("Proc", func(t *testing.T) {
			Main(f)
		})
	}
}
