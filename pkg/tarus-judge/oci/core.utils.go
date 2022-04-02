package oci_judge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	hr_bytes "github.com/Myriad-Dreamin/tarus/pkg/hr-bytes"
	tarus_io "github.com/Myriad-Dreamin/tarus/pkg/tarus-io"
	"github.com/containerd/containerd/api/types"
	v1 "github.com/containerd/containerd/metrics/types/v1"
	"github.com/containerd/typeurl"
	"github.com/opencontainers/runtime-spec/specs-go"
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
	ProcessSpec    specs.Process
	MemoryLimit    int64
	StackLimit     int64
	CpuTime        int64
	CpuHard        int64
	ChannelFactory tarus_io.ChannelFactory
	JudgeIO        tarus_io.Factory
}

func (c *ContainerdJudgeServiceServer) createProcSpec(
	_ context.Context, sessionEnv, judgeEnv *JudgeEnvironment, judgePoint *tarus.JudgeTestcase) (err error) {
	judgeEnv.ProcessSpec = sessionEnv.ProcessSpec

	if len(judgePoint.IoProvider) != 0 {
		judgeEnv.ChannelFactory, err = c.ioRouter.MakeIOChannel(judgePoint.IoProvider)
		if err != nil {
			return err
		}
	}
	if judgeEnv.ChannelFactory == nil {
		if sessionEnv.ChannelFactory == nil {
			return errors.New("io provider should be set under either session level or case level")
		}
		judgeEnv.ChannelFactory = sessionEnv.ChannelFactory
	}

	judgeEnv.JudgeIO, err = judgeEnv.ChannelFactory(judgePoint.Input, judgePoint.Answer)
	if err != nil {
		return err
	}

	judgeEnv.CpuHard = getOrDefault(judgePoint.EstimatedCpuhard, sessionEnv.CpuHard)
	judgeEnv.CpuTime = getOrDefault(judgePoint.EstimatedCputime, sessionEnv.CpuTime)
	judgeEnv.MemoryLimit = getOrDefault(judgePoint.EstimatedMemory, sessionEnv.MemoryLimit)
	judgeEnv.StackLimit = getOrDefault(judgePoint.EstimatedStack, sessionEnv.StackLimit)

	// todo: check default rlimit, check rlimit_core
	var rlimits []specs.POSIXRlimit
	if len(judgeEnv.ProcessSpec.Rlimits) != 0 {
		rlimits = judgeEnv.ProcessSpec.Rlimits[:0]
	}
	if judgeEnv.CpuTime > 1 || judgeEnv.CpuHard > 1 {
		if judgeEnv.CpuTime > 1 && judgeEnv.CpuHard > 1 {
			rlimits = append(rlimits, specs.POSIXRlimit{
				Type: "RLIMIT_CPU",
				Soft: uint64(judgeEnv.CpuTime),
				Hard: uint64(judgeEnv.CpuHard),
			})
		} else {
			return errors.New("both cpu hard and cpu time should be set at the same time")
		}
	}
	if judgeEnv.MemoryLimit > 1 {
		rlimits = append(rlimits, specs.POSIXRlimit{
			Type: "RLIMIT_DATA",
			Soft: uint64(judgeEnv.MemoryLimit),
			Hard: uint64(judgeEnv.MemoryLimit + (32 * int64(hr_bytes.MB))),
		})
	}
	if judgeEnv.StackLimit > 1 {
		rlimits = append(rlimits, specs.POSIXRlimit{
			Type: "RLIMIT_STACK",
			Soft: uint64(judgeEnv.StackLimit),
			Hard: uint64(judgeEnv.StackLimit + (32 * int64(hr_bytes.MB))),
		})
	}
	judgeEnv.ProcessSpec.Rlimits = rlimits

	return nil
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
		queryResult.Status, err = judgeEnv.JudgeIO.GetJudgeStatus(rawMetric.JudgeReport)
		if err != nil {
			return err
		}
	}

	return nil
}
