package tarus_compiler

import "testing"

func TestGetSystemGnu(t *testing.T) {
	if c := GetSystemGnu(); c == nil {
		t.Error("can not found gnu compiler")
	}
}
