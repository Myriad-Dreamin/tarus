package main

import (
	"fmt"
	tarus_app "github.com/Myriad-Dreamin/tarus/cmd/tarus-cli/tarus-app"
	"os"
)

func main() {
	app := tarus_app.New()
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "tarus-cli: %s\n", err)
		os.Exit(1)
	}
}
