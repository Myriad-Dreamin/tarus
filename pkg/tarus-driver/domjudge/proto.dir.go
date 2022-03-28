package domjudge_driver

import (
	"encoding/binary"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type sortDomJudgeCase []*tarus.JudgeTestcase

func (s sortDomJudgeCase) Len() int {
	return len(s)
}

func (s sortDomJudgeCase) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortDomJudgeCase) Less(i, j int) bool {
	return s[i].Input < s[j].Input
}

func GetJudgeDescriptors(problemSet fs.FS) (ret []*tarus.JudgeTestcase, _ error) {
	// check root exists
	if _, err := fs.Stat(problemSet, "."); err != nil {
		return nil, err
	}

	// see
	// https://github.com/DOMjudge/domjudge/blob/main/webapp/src/Service/ImportProblemService.php
	// or verified latest commit:
	// https://github.com/DOMjudge/domjudge/blob/5ca3bab8e59d128112e8aee6db20c40525c7cef2/webapp/src/Service/ImportProblemService.php
	for _, ty := range []string{"sample", "secret"} {
		if entries, err := fs.Glob(problemSet, filepath.Join("data", ty, "*.in")); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			for _, entry := range entries {
				ansEntry := strings.TrimSuffix(entry, ".in") + ".ans"
				if _, err := fs.Stat(problemSet, ansEntry); err != nil {
					if !os.IsNotExist(err) {
						return nil, err
					}
					continue
				}
				ret = append(ret, &tarus.JudgeTestcase{
					Input:  entry,
					Answer: ansEntry,
				})
			}
		}
	}

	sort.Sort(sortDomJudgeCase(ret))
	for i := range ret {
		ret[i].JudgeKey = make([]byte, 2)
		binary.LittleEndian.PutUint16(ret[i].JudgeKey, uint16(i))
	}
	return
}

func CreateLocalJudgeRequest(fsArchive string) (req *tarus.MakeJudgeRequest, _ error) {
	fsArchive, _ = filepath.Abs(fsArchive)
	testcases, err := GetJudgeDescriptors(os.DirFS(fsArchive))
	if err != nil {
		return nil, err
	}
	req = new(tarus.MakeJudgeRequest)
	req.IoProvider = fmt.Sprintf("file://%s", fsArchive)
	req.Testcases = testcases

	return req, nil
}
