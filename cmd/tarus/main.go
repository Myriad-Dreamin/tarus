package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
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
	app.Version = "v0.0.0"
	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "tarus: %s\n", err)
		os.Exit(1)
	}
}
