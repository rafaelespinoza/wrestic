package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rafaelespinoza/alf"
	"github.com/rafaelespinoza/wrestic/internal/config"
)

var defaultConfigDir string

func init() {
	dir, err := config.DefaultBaseConfigDir()
	if err != nil {
		// Yeah, this would be an unfortunate. Here it's only for documentation
		// purposes; it will be empty in the message.
		fmt.Fprintf(os.Stderr, "could not determine default user configuration directory: %s\n", err)
	}
	defaultConfigDir = dir
}

func makeConfig(parentName, name string) alf.Directive {
	fullName := parentName + " " + name
	var params struct {
		configDir string
	}

	out := alf.Delegator{
		Description: "manage configuration",
		Flags:       flag.NewFlagSet(name, flag.ExitOnError),
		Subs: map[string]alf.Directive{
			"init": &alf.Command{
				Description: "initialize configuration directory",
				Setup: func(_ flag.FlagSet) *flag.FlagSet {
					flags := flag.NewFlagSet(fullName, flag.ExitOnError)
					flags.StringVar(&params.configDir, "C", defaultConfigDir, "base configuration directory")
					flags.Usage = func() {
						fmt.Fprintf(flags.Output(), `Usage: %s [flags]

Description:

	Prepare configuration directory structure. Some application data, such as
	encrypted passwords, may also live here.

	The current configuration directory is:

		%q

Flags:

`, fullName, params.configDir)
						flags.PrintDefaults()
					}

					return flags
				},
				Run: func(ctx context.Context) error {
					dir, err := config.Init(params.configDir)
					if err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "config directory initialized at %q\n", dir)
					return nil
				},
			},
		},
	}

	out.Flags.Usage = func() {
		fmt.Fprintf(out.Flags.Output(), `Usage: %s subcmd [subflags]

Description:

	Manage configuration.

	The default configuration directory is:

		%q

Subcommands:

	These may have their own set of flags. Put them after the subcommand.

	%v

`, fullName, defaultConfigDir, formatSubcommandsDescriptions(&out))
	}

	return &out
}
