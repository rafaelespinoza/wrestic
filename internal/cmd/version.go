package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// These are pieces of version metadata that can be set through -ldflags.
var (
	versionBranchName string
	versionBuildTime  string
	versionCommitHash string
	versionGoOSArch   string
	versionGoVersion  string
	versionTag        string
)

func makeVersion(parentName, name string) *cli.Command {
	return &cli.Command{
		Name:        "version",
		Usage:       "output build info",
		Description: `displays versioning data about the current binary.`,
		Action: func(c *cli.Context) error {
			const format = "%-12s %s\n"
			fmt.Fprintf(os.Stdout, format, "BranchName", versionBranchName)
			fmt.Fprintf(os.Stdout, format, "BuildTime", versionBuildTime)
			fmt.Fprintf(os.Stdout, format, "CommitHash", versionCommitHash)
			fmt.Fprintf(os.Stdout, format, "GoOSArch", versionGoOSArch)
			fmt.Fprintf(os.Stdout, format, "GoVersion", versionGoVersion)
			fmt.Fprintf(os.Stdout, format, "Tag", versionTag)
			return nil
		},
	}
}
