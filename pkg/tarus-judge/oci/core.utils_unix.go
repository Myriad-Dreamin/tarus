//go:build !plan9 && !windows

package oci_judge

import (
	"os"
	"regexp"
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
