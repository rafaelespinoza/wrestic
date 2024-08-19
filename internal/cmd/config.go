package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	execSubcmd := parentName + " exec"
	showOutputFormats := []string{"toml", "json"}

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
			{
				Name:  "show",
				Usage: "select datastores, destinations for display",
				Description: fmt.Sprintf(`Inspect configured datastores and destinations.

This subcommand performs the same filtering that happens when invoking the
subcommand, %s. So this functionality may be useful for previewing which
destinations (restic repositories) would be affected, and with which restic
flags, before invoking a configured restic subcommand via %s.

By default, the displayed configuration values for a destination are merged in
from the parent datastore. Likewise, the datastore's configuration values are
merged in from any top-level configuration values.`,
					execSubcmd, execSubcmd),

				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:    "config-dir",
						Aliases: []string{"C"},
						Usage:   "base configuration directory",
						Value:   defaultConfigDir,
					},
					&cli.StringSliceFlag{
						Name:    "storenames",
						Aliases: []string{"s"},
						Usage:   "names of stores to operate on",
					},
					&cli.StringSliceFlag{
						Name:    "destnames",
						Aliases: []string{"d"},
						Usage:   "names of destinations to operate on",
					},
					&cli.BoolFlag{
						Name:    "merge",
						Aliases: []string{"m"},
						Usage:   "merge configuration values into destination",
						Value:   true,
					},
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   fmt.Sprintf("output format, one of %q", showOutputFormats),
						Value:   showOutputFormats[0],
					},
				},
				Action: func(c *cli.Context) error {
					configDir := c.Path("config-dir")
					if configDir == "" {
						return errors.New("config dir cannot be empty; possibly could not determine a default either")
					}
					outputFormat := c.String("format")
					{
						var formatOK bool // TODO: use slices.Contains if upgrading min golang version to >= v1.21.0
						for i := range showOutputFormats {
							if outputFormat == showOutputFormats[i] {
								formatOK = true
								break
							}
						}
						if !formatOK {
							return fmt.Errorf("unknown format %q, should be one of %q", outputFormat, showOutputFormats)
						}
					}
					stores, err := fetchDatastores(configDir, c.StringSlice("storenames"), c.StringSlice("destnames"))
					if err != nil {
						return err
					}

					return displayDatastores(os.Stdout, c.Bool("merge"), outputFormat, stores)
				},
			},
		},
	}

	return &out
}

func fetchDatastores(configDir string, storenames, destnames []string) (out []config.Datastore, err error) {
	file, err := os.Open(filepath.Clean(filepath.Join(configDir, "wrestic.toml")))
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	params, err := config.Parse(file)
	if err != nil {
		return
	}

	out = config.SelectDatastores(params.Datastores, storenames, destnames)
	return
}

func displayDatastores(w io.Writer, merge bool, format string, stores []config.Datastore) (err error) {
	for _, store := range stores {
		if merge {
			for name, dest := range store.Destinations {
				defs, err := dest.Merge()
				if err != nil {
					return err
				}
				dest.Defaults = defs
				store.Destinations[name] = dest
			}
		}

		switch format {
		case "toml":
			fmt.Fprintf(w, "#\n# %s\n#\n", store.Name)
			if err = config.EncodeTOML(w, store); err != nil {
				return
			}
			fmt.Fprintln(w)
		case "json":
			var raw []byte
			if raw, err = json.Marshal(store); err != nil {
				return
			}
			fmt.Fprintf(w, "%s\n", raw)
		}
	}

	return nil
}
