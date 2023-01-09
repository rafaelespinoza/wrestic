package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/rafaelespinoza/alf"
)

// Root abstracts a top-level command from package main.
type Root interface {
	// Run is the entry point. It should be called with os.Args[1:].
	Run(ctx context.Context, args []string) error
}

// New establishes the root command and subcommands.
func New() Root {
	const name = "wrestic"

	deleg := alf.Delegator{
		Description: "main command for " + name,
		Flags:       flag.NewFlagSet(name, flag.ExitOnError),
		Subs: map[string]alf.Directive{
			"config":  makeConfig(name, "config"),
			"restic":  makeRestic(name, "restic"),
			"version": makeVersion(name, "version"),
		},
	}

	deleg.Flags.Usage = func() {
		fmt.Fprintf(deleg.Flags.Output(), `Usage: %s subcommand [subflags]

Description:

	%s is a tool to help you manage backups of your data.
	It does not do any backing up directly, though the restic subcommand here
	can invoke the real restic to do that.
	Configuratoin

	- A "source" is a path on a host, where the data at that path is backed
	  up to a restic repository.
	- A "destination" is the actual restic backup repository.
	- A "datastore", or "store" is an abstraction for backed up data which
	  encompasses sources and destinations.

Subcommands:

	These may have their own set of flags. Put them after the subcommand.

	%v

`, name, name, formatSubcommandsDescriptions(&deleg))
	}

	return &alf.Root{
		Delegator: &deleg,
	}
}

func formatSubcommandsDescriptions(cmd *alf.Delegator) string {
	return strings.Join(cmd.DescribeSubcommands(), "\n\t")
}

type stringList struct {
	inputs  string
	outputs []string
}

func (s *stringList) Set(in string) error {
	s.inputs = in
	s.outputs = strings.Split(in, ",")
	return nil
}

func (s *stringList) String() string { return s.inputs }
