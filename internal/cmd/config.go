package cmd

import (
	"fmt"
	"os"

	"github.com/rafaelespinoza/wrestic/internal/config"
	"github.com/urfave/cli/v2"
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

func makeConfig(parentName, name string) *cli.Command {
	out := cli.Command{
		Name:  name,
		Usage: "manage application configuration",
		Description: fmt.Sprintf(`Manage configuration.

The default configuration directory is:
	%q`, defaultConfigDir),

		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "prepare configuration directory structure",
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:    "config-dir",
						Aliases: []string{"C"},
						Usage:   "base configuration directory",
						Value:   defaultConfigDir,
					},
				},
				Description: `Prepare configuration directory structure. Some application data, such as
encrypted passwords, may also live here.`,
				Action: func(c *cli.Context) error {
					dir, err := config.Init(c.Path("config-dir"))
					if err != nil {
						return err
					}
					fmt.Fprintf(os.Stderr, "config directory initialized at %q\n", dir)
					return nil
				},
			},
		},
	}

	return &out
}
