package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ExeceptionTrap struct {
	null *int32
	oob  []int32
	a, b int
	fpe  int
}

func emitSignal(s os.Signal) {
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(s)
}

func Main(injectedActions string) {
	var trap = new(ExeceptionTrap)
	var memoryUsed [][]byte
	trap.oob = make([]int32, 0)
	for _, action := range strings.Split(injectedActions, ",") {
		kv := strings.Split(action, "=")
		switch len(kv) {
		case 0:
			break
		case 1:
			switch action {
			case "null":
				*trap.null = 1
			case "out_of_bound_access":
				trap.oob[1] = 1
			case "fpe":
				trap.a = 1 / trap.fpe
			case "exit":
				os.Exit(0)
			case "abort":
				emitSignal(syscall.SIGABRT)
			default:
				panic(fmt.Errorf("invalid action %s", action))
			}
		case 2:
			action = kv[0]
			value, err := strconv.Atoi(kv[1])
			if err != nil {
				panic(fmt.Errorf("invalid num"))
			}
			switch action {
			case "exit":
				os.Exit(value)
			case "virt_memory":
				memoryUsed = append(memoryUsed, make([]byte, value))
			case "memory":
				tsMemory := make([]byte, value)
				for i := 0; i < len(tsMemory); i++ {
					tsMemory[i] = 1
				}
				memoryUsed = append(memoryUsed, tsMemory)
			case "time":
				time.Sleep(time.Millisecond * time.Duration(value))
			case "signal":
				emitSignal(syscall.Signal(value))
			default:
				panic(fmt.Errorf("invalid action %s", action))
			}
		}
	}
}

func main() {
	if len(os.Args) > 1 {
		Main(os.Args[1])
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			log.Printf("Failed to read line: %v", scanner.Err())
			return
		}
		Main(string(scanner.Bytes()))
	}
}
