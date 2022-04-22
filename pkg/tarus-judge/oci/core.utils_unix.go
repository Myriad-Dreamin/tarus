//go:build !plan9 && !windows

package oci_judge

import (
	"context"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/contrib/seccomp"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"os"
	"regexp"
	"strings"
	"syscall"
)

var signalMapping = map[string]os.Signal{
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

var consoleOupRegex = regexp.MustCompile("SIG(?:ABRT|ALRM|BUS|CHLD|CLD|CONT|FPE|HUP|ILL|INT|IO|IOT|KILL|PIPE|POLL|PROF|PWR|QUIT|SEGV|STKFLT|STOP|SYS|TERM|TRAP|TSTP|TTIN|TTOU|UNUSED|URG|USR1|USR2|VTALRM|WINCH|XCPU|XFSZ)")

func captureConsoleSignal(oup []byte) os.Signal {
	if trapped := consoleOupRegex.Find(oup); len(trapped) != 0 {
		return signalMapping[string(trapped)]
	}
	return nil
}

var judgeSeccomp *specs.LinuxSeccomp

func init() {
	var dummySpec specs.Spec

	// no privilege is granted
	dummySpec.Process = new(specs.Process)
	dummySpec.Process.Capabilities = new(specs.LinuxCapabilities)
	dummySpec.Process.Capabilities.Bounding = []string{}
	judgeSeccomp = seccomp.DefaultProfile(&dummySpec)

	var defaultSyscalls = judgeSeccomp.Syscalls
	var allowedSyscalls []specs.LinuxSyscall

	for _, s := range defaultSyscalls {

		// we only modify the main
		var isMainAllow = false
		if s.Action >= specs.ActAllow {
			for i := range s.Names {
				// some linux distribution deprecate open and use openat
				if s.Names[i] == "open" || s.Names[i] == "openat" {
					isMainAllow = true
					break
				}
			}
		}
		if isMainAllow {
			var names = s.Names
			var newNames []string
			for i := range names {
				// network operations should be mitigated by apparmor or runsc
				if /* filesystem */
				strings.Contains(names[i], "inotify") ||
					strings.Contains(names[i], "chroot") ||
					strings.Contains(names[i], "xattr") ||
					strings.HasPrefix(names[i], "fanotify_") ||
					(strings.Contains(names[i], "set") && (strings.Contains(names[i], "gid") || strings.Contains(names[i], "uid"))) ||
					/* dangerous timer operation */
					strings.Contains(names[i], "timer") ||
					/* dangerous memory operation */
					strings.HasPrefix(names[i], "memfd_") ||
					strings.HasPrefix(names[i], "shm") ||
					/* linux classic async operation */
					strings.HasPrefix(names[i], "epoll_") ||
					/* linux unnecessary components */
					strings.HasPrefix(names[i], "eventfd") ||
					strings.HasPrefix(names[i], "mq_") ||
					strings.HasPrefix(names[i], "pidfd_") ||
					/* linux recent components */
					strings.HasPrefix(names[i], "io_") ||
					strings.HasPrefix(names[i], "landlock_") ||
					/* seccomp operation */
					strings.Contains(names[i], "prctl") ||
					strings.Contains(names[i], "seccomp") {
					continue
				}
				newNames = append(newNames, names[i])
			}
			s.Names = newNames
		}

		allowedSyscalls = append(allowedSyscalls, s)
	}

	judgeSeccomp.Syscalls = allowedSyscalls
	judgeSeccomp.DefaultAction = specs.ActKill
	// _, _ = pp.Println(judgeSeccomp)
}

func withSeccomp(j *specs.LinuxSeccomp) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {
		s.Linux.Seccomp = j
		return nil
	}
}
