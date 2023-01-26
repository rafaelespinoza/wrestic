package cmd

import (
	"context"

	"github.com/urfave/cli/v2"
)

// Root abstracts a top-level command from package main.
type Root interface {
	RunContext(ctx context.Context, args []string) error
}

// New establishes the root command and subcommands.
func New() Root {
	const name = "wrestic"

	app := cli.NewApp()
	app.Name = name
	app.Usage = "restic and a configuration file"
	app.Commands = []*cli.Command{
		makeConfig(name, "config"),
		makeExec(name, "exec"),
		makeVersion(name, "version"),
	}
	app.Description = `Manage backups of your data.

It does not do any backing up directly, though the exec subcommand here
can invoke the real restic to do that.

- A "source" is a path on a host, where the data at that path is backed
  up to a restic repository.
- A "destination" is the actual restic backup repository.
- A "datastore", or "store" is an abstraction for backed up data which
  encompasses sources and destinations.`

	return app
}
