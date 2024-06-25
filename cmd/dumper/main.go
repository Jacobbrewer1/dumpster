package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

// Set at linking time
var (
	Commit string
	Date   string
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")

	subcommands.Register(new(versionCmd), "")
	subcommands.Register(new(ddlCmd), "")
	subcommands.Register(new(dumpCmd), "")
	subcommands.Register(new(purgeCmd), "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
