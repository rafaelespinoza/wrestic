package exec

import (
	"fmt"
	"io"
	"strings"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

func (b ResticBatch) buildArgs(dest config.Destination, srcPaths ...config.Source) ([]string, error) {
	tuples, err := dest.BuildFlags(b.ConfigDir, b.Subcommand)
	if err != nil {
		return nil, err
	}

	out := []string{b.Subcommand}
	for _, tuple := range tuples {
		out = append(out, fmt.Sprintf("--%s=%s", tuple.Key, tuple.Val))
	}

	for _, arg := range b.Args {
		out = append(out, arg)
	}

	if b.Subcommand == "backup" {
		// It'll probably be more natural to put paths or directories at the end
		// of the slice.
		for _, src := range srcPaths {
			out = append(out, src.Path)
		}
	}

	return out, nil
}

func printArgs(w io.Writer, args ...string) {
	var bld strings.Builder

	// Reduce chances of mistakenly executing the command by outputting a
	// shell comment.
	bld.WriteRune('#')

	for _, tuple := range args {
		arg := tuple

		if strings.HasPrefix(tuple, pwcmdFlagKey) {
			arg = quotePasswordFlag(tuple)
		}

		bld.WriteString(` ` + arg)

	}

	bld.WriteRune('\n')

	fmt.Fprintf(w, bld.String())
}

// pwcmdFlagKey is the name of the flag that the restic CLI expects to find the
// password command.
const pwcmdFlagKey = "--password-command"

// quotePasswordFlag makes it easier to copy and paste a flag value into a
// terminal by putting quotes around the shell command or filename. Use single
// quotes here, rather than double quotes, because the makePasswordFlag function
// may have already put double quotes around the paths in this command.
func quotePasswordFlag(in string) (out string) {
	key, val, _ := strings.Cut(in, "=")
	out = key + `='` + val + `'`
	return
}
