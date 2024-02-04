package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"

	"github.com/google/subcommands"
)

type versionCmd struct{}

func (v versionCmd) Name() string {
	return "version"
}

func (v versionCmd) Synopsis() string {
	return "Print application version information and exit"
}

func (v versionCmd) Usage() string {
	return `version:
  Print application version information and exit.
`
}

func (v versionCmd) SetFlags(f *flag.FlagSet) {}

func (v versionCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Printf(
		"Commit: %s\nRuntime: %s %s/%s\nDate: %s\n",
		Commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		Date,
	)
	return subcommands.ExitSuccess
}
