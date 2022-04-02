package oci_judge

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	tarus_io "github.com/Myriad-Dreamin/tarus/pkg/tarus-io"
	"github.com/containerd/containerd/api/types"
	v1 "github.com/containerd/containerd/metrics/types/v1"
	"github.com/containerd/typeurl"
	"os"
	"regexp"
	"syscall"
	"time"
)

var golangRepMapping = map[string]os.Signal{
	"SIGABRT":   syscall.SIGABRT,
	"SIGALRM":   syscall.SIGALRM,
	"SIGBUS":    syscall.SIGBUS,
	"SIGCHLD":   syscall.SIGCHLD,
	"SIGCLD":    syscall.SIGCLD,
	"SIGCONT":   syscall.SIGCONT,
	"SIGFPE":    syscall.SIGFPE,
	"SIGHUP":    syscall.SIGHUP,
	"SIGILL":    syscall.SIGILL,
	"SIGINT":    syscall.SIGINT,
	"SIGIO":     syscall.SIGIO,
	"SIGIOT":    syscall.SIGIOT,
	"SIGKILL":   syscall.SIGKILL,
	"SIGPIPE":   syscall.SIGPIPE,
	"SIGPOLL":   syscall.SIGPOLL,
	"SIGPROF":   syscall.SIGPROF,
	"SIGPWR":    syscall.SIGPWR,
	"SIGQUIT":   syscall.SIGQUIT,
	"SIGSEGV":   syscall.SIGSEGV,
	"SIGSTKFLT": syscall.SIGSTKFLT,
	"SIGSTOP":   syscall.SIGSTOP,
	"SIGSYS":    syscall.SIGSYS,
	"SIGTERM":   syscall.SIGTERM,
	"SIGTRAP":   syscall.SIGTRAP,
	"SIGTSTP":   syscall.SIGTSTP,
	"SIGTTIN":   syscall.SIGTTIN,
	"SIGTTOU":   syscall.SIGTTOU,
	"SIGUNUSED": syscall.SIGUNUSED,
	"SIGURG":    syscall.SIGURG,
	"SIGUSR1":   syscall.SIGUSR1,
	"SIGUSR2":   syscall.SIGUSR2,
	"SIGVTALRM": syscall.SIGVTALRM,
	"SIGWINCH":  syscall.SIGWINCH,
	"SIGXCPU":   syscall.SIGXCPU,
	"SIGXFSZ":   syscall.SIGXFSZ,
}

func captureProgramRep(oup []byte) os.Signal {
	golangRepCapture := regexp.MustCompile("SIG(?:ABRT|ALRM|BUS|CHLD|CLD|CONT|FPE|HUP|ILL|INT|IO|IOT|KILL|PIPE|POLL|PROF|PWR|QUIT|SEGV|STKFLT|STOP|SYS|TERM|TRAP|TSTP|TTIN|TTOU|UNUSED|URG|USR1|USR2|VTALRM|WINCH|XCPU|XFSZ)")
	if trapped := golangRepCapture.Find(oup); len(trapped) != 0 {
		return golangRepMapping[string(trapped)]
	}
	return nil
}

func analysisContainerSignal(code uint32, oup []byte) os.Signal {
	if 128 < code && code < 128+0x20 { // unix program
		return syscall.Signal(code - 128)
	} else if code == 2 { // golang program
		return captureProgramRep(oup)
	}

	return nil
}

type JudgeEnvironment struct {
	MemoryLimit  int64
	StackLimit   int64
	CpuTime      int64
	CpuHard      int64
	JudgeChecker tarus_io.JudgeChecker
}

type JudgeMetric struct {
	ContainerInfo *types.Metric
	Code          uint32
	IsTimeout     bool
	StartedAt     time.Time
	ExitedAt      time.Time
	JudgeReport   []byte
}

func (c *ContainerdJudgeServiceServer) analysisJudgeResult(
	_ context.Context, judgeEnv *JudgeEnvironment, queryResult *tarus.QueryJudgeItem, rawMetric *JudgeMetric,
) (err error) {

	var jh JudgeHint
	jh.Code = int(rawMetric.Code)
	if rawMetric.ExitedAt.IsZero() {
		queryResult.TimeUseHard = int64(time.Now().Sub(rawMetric.StartedAt) * time.Nanosecond)
	} else {
		queryResult.TimeUseHard = int64(rawMetric.ExitedAt.Sub(rawMetric.StartedAt) * time.Nanosecond)
	}

	jh.CheckerResult = string(rawMetric.JudgeReport)

	m0, err := typeurl.UnmarshalAny(rawMetric.ContainerInfo.Data)
	if err != nil {
		return err
	}
	if m2, ok := m0.(*v1.Metrics); ok {
		//fmt.Printf("metrics cpu result: %v %v %v\n", m2.CPU.Usage.Total, m2.CPU.Usage.Kernel, m2.CPU.Usage.User)
		//fmt.Printf("metrics memory result: %v %v %v\n", m2.Memory.Usage.Max, m2.Memory.Usage.Usage, m2.Memory.RSS)
		queryResult.TimeUse = int64(m2.CPU.Usage.User)
		queryResult.MemoryUse = int64(m2.Memory.Usage.Max)
		if queryResult.MemoryUse < int64(m2.Memory.TotalRSS) {
			queryResult.MemoryUse = int64(m2.Memory.TotalRSS)
		}
	} else {
		fmt.Println("invalid type url for extracting metrics", rawMetric.ContainerInfo.Data.TypeUrl)
	}

	var sig = analysisContainerSignal(rawMetric.Code, rawMetric.JudgeReport)
	if sig != nil {
		jh.Signal = sig.String()
	}

	queryResult.Hint, err = json.Marshal(jh)
	if err != nil {
		return err
	}

	if queryResult.MemoryUse >= judgeEnv.MemoryLimit {
		queryResult.Status = tarus.JudgeStatus_MemoryLimitExceed
	} else if queryResult.TimeUseHard >= judgeEnv.CpuTime+int64(time.Millisecond*50) ||
		queryResult.TimeUse >= judgeEnv.CpuTime+int64(time.Millisecond*50) {
		queryResult.Status = tarus.JudgeStatus_TimeLimitExceed
	}

	if rawMetric.IsTimeout {
		queryResult.Status = tarus.JudgeStatus_TimeLimitExceed
	} else if sig != nil {
		switch sig {
		case syscall.SIGKILL:
			if queryResult.Status == 0 {
				queryResult.Status = tarus.JudgeStatus_RuntimeError
			}
		case syscall.SIGSYS:
			queryResult.Status = tarus.JudgeStatus_SecurityPolicyViolation
		case syscall.SIGABRT:
			queryResult.Status = tarus.JudgeStatus_AssertionFailed
		case syscall.SIGFPE:
			queryResult.Status = tarus.JudgeStatus_FloatingPointException
		// case syscall.SIGSEGV:
		// case syscall.SIGILL:
		default:
			queryResult.Status = tarus.JudgeStatus_RuntimeError
		}
	} else if rawMetric.Code != 0 {
		queryResult.Status = tarus.JudgeStatus_RuntimeError
	}

	if queryResult.Status == 0 {
		queryResult.Status, err = judgeEnv.JudgeChecker.GetJudgeStatus(rawMetric.JudgeReport)
		if err != nil {
			return err
		}
	}

	return nil
}
