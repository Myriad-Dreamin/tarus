package main

import (
	"context"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	oci_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge/oci"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/sys"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
)

var checkerFeatures = []string{
	"native::strict_matcher",
	"testlib",
}

var runtimeFeatures = []string{
	"oci::containerd",
}

var languageFeatures = []string{
	"Clang",
	"GNU",
}

var feature = fmt.Sprintf(`Current online features are listed following:
     Checker: %v
     Runtime: %v
     LanguageTarget: %v`,
	strings.Join(checkerFeatures, " "),
	strings.Join(runtimeFeatures, " "),
	strings.Join(languageFeatures, " "),
)

func main() {
	app := cli.NewApp()
	app.Name = "tarus"
	app.Usage = "Online Judge Engine Powered by runC, gVisor and other execution runtime."
	app.Description = app.Usage + " " + feature
	app.EnableBashCompletion = true
	app.Version = "v0.0.0"
	h, _ := os.UserHomeDir()
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output in logs",
		},
		cli.StringFlag{
			Name:   "address, a",
			Usage:  "address for tarus service's GRPC server",
			Value:  filepath.Join(h, ".config/tarus/service.sock"),
			EnvVar: "TARUS_SERVICE_ADDRESS",
		},
	}

	app.Action = func(cliArgs *cli.Context) error {
		var (
			config      = defaultConfig()
			signals     = make(chan os.Signal, 2048)
			ctx, cancel = context.WithCancel(context.Background())
		)

		for _, v := range []struct {
			name string
			d    *string
		}{
			{
				name: "address",
				d:    &config.GRPCAddr,
			},
		} {
			if s := cliArgs.GlobalString(v.name); s != "" {
				*v.d = s
			}
		}

		var closers []io.Closer
		var closer = func() {
			for i := range closers {
				_ = closers[i].Close()
			}
		}

		done := handleSignals(ctx, signals, closer, cancel)
		// start the signal handler as soon as we can to make sure that
		// we don't miss any signals during boot
		signal.Notify(signals, handledSignals...)

		containerdServer, err := oci_judge.NewContainerdServer()
		if err != nil {
			panic(err)
		}
		closers = append(closers, containerdServer)

		var hostServer = grpc.NewServer()
		hostServer.RegisterService(&tarus.JudgeService_ServiceDesc, containerdServer)

		// setup the main grpc endpoint
		l, err := sys.GetLocalListener(config.GRPCAddr, 0, 0)
		if err != nil {
			return fmt.Errorf("failed to get listener for main endpoint: %w", err)
		}
		serve(ctx, l, hostServer.Serve)
		<-done
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "tarus: %s\n", err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, l net.Listener, serveFunc func(net.Listener) error) {
	path := l.Addr().String()
	log.G(ctx).WithField("address", path).Info("serving...")
	go func() {
		defer func() {
			_ = l.Close()
		}()
		if err := serveFunc(l); err != nil {
			log.G(ctx).WithError(err).WithField("address", path).Fatal("serve failure")
		}
	}()
}
