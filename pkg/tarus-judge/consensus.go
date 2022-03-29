package tarus_judge

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"sort"
)

const (
	JudgeServiceApiAll     = "All"
	JudgeServiceApiMinimum = "Min"
)

func judgeStatusHash() string {
	var values []string
	for k, _ := range tarus.JudgeStatus_value {
		values = append(values, k)
	}
	sort.Sort(sort.StringSlice(values))
	var md5Digest = md5.New()
	var err error
	var hasError bool
	for _, k := range values {
		_, err = md5Digest.Write([]byte(k))
		if err != nil {
			hasError = true
		}
		_, err = md5Digest.Write([]byte("$"))
		if err != nil {
			hasError = true
		}
	}

	if hasError {
		return ""
	}
	return hex.EncodeToString(md5Digest.Sum(nil))
}

var JudgeStatusHash string

func init() {
	JudgeStatusHash = judgeStatusHash()
}
