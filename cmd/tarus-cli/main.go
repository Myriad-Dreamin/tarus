package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "tarus-cli"
	app.Usage = "Cli for Online Judge Engine -- tarus."
	app.Description = app.Usage
	app.Version = "v0.0.0"
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "tarus-cli: %s\n", err)
		os.Exit(1)
	}
}
