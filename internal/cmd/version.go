package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rafaelespinoza/alf"
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

func makeVersion(parentName, name string) alf.Directive {
	fullName := parentName + " " + name

	return &alf.Command{
		Description: "output build info",
		Setup: func(p flag.FlagSet) *flag.FlagSet {
			flags := flag.NewFlagSet(fullName, flag.ExitOnError)
			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `Usage: %s

Description:

	%s displays versioning data about the current binary.

`, fullName, fullName)
			}
			return flags
		},
		Run: func(_ context.Context) error {
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
