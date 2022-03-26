package main

import (
	"context"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

var handledSignals = []os.Signal{
	unix.SIGTERM,
	unix.SIGINT,
	unix.SIGUSR1,
	unix.SIGPIPE,
}

func handleSignals(ctx context.Context, signals chan os.Signal, stop func(), cancel func()) chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case s := <-signals:

				// Do not print message when dealing with SIGPIPE, which may cause
				// nested signals and consume lots of cpu bandwidth.
				if s == unix.SIGPIPE {
					continue
				}

				fmt.Println("received signal")
				switch s {
				case unix.SIGUSR1:
					break
				default:
					cancel()
					stop()
					close(done)
					return
				}
			}
		}
	}()
	return done
}
