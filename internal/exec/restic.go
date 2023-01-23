package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

// ResticBatch is a set of named parameters for operating a restic subcommand
// upon multiple destinations.
type ResticBatch struct {
	ConfigDir  string         // ConfigDir is the parent directory for the age-formatted keypair and encrypted secrets.
	Sink       io.Writer      // Sink may capture the arguments and flags generated for the restic subcommand.
	Subcommand string         // Subcommand is the restic subcommand to run.
	Args       []string       // Args are the flags and positional arguments to pass to Subcommand.
	Run        bool           // Run toggles whether the subcommand is actually invoked or not.
	NewCommand func() Command // NewCommand allows some inversion of control, mostly useful for testing.
}

// Do may invoke a restic subcommand (named by Subcommand), with any positional
// arguments and flags (specified in Args) for each member of destinations. By
// default, the restic subcommand is not actually run, instead the arguments and
// flags that would be passed to restic are written to stderr as a shell
// comment. To actually run restic, set Run to true. If Sink is non-empty then
// generated command line args (prefixed with a # for roll-safe purposes) are
// written to Sink.
func (b ResticBatch) Do(ctx context.Context, datastores []config.Datastore) error {
	for _, store := range datastores {
		for _, dest := range store.Destinations {
			args, err := b.buildArgs(dest, store.Sources...)
			if err != nil {
				return fmt.Errorf("%w: store=%q, destination=%q", err, store.Name, dest.Name)
			}

			if b.Sink != nil {
				printArgs(b.Sink, args...)
			}

			if !b.Run { // is this a preview of commands to run?
				continue
			}

			runner := b.NewCommand()
			if err = runner.Run(ctx, args...); err != nil {
				return fmt.Errorf("%w: store=%q, destination=%q", err, store.Name, dest.Name)
			}
		}
	}

	return nil
}

// A Command is an external command to execute with args.
type Command interface {
	Run(ctx context.Context, args ...string) error
}

// NewRestic constructs a Command capable of running restic. By default, it will
// pick the first restic executable found in PATH. The path to the restic
// binary may be overridden with the environment variable, RESTIC_BIN.
func NewRestic(outSink, errSink io.Writer) Command {
	return restic{outSink, errSink}
}

type restic struct{ outSink, errSink io.Writer }

func (r restic) Run(ctx context.Context, args ...string) (err error) {
	bin := "restic"
	// Optionally, check for alternate restic binaries. The main use case is for
	// running a different version of restic. But tests could also use this env
	// var for sanity checking application behavior in a controlled manner.
	if val := os.Getenv("RESTIC_BIN"); val != "" {
		bin = val
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = r.outSink
	cmd.Stderr = r.errSink

	err = cmd.Run()
	return
}
