package exec

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

const (
	passwordCommandFlagKey = "--password-command" // TODO: support this in config
	passwordFileFlagKey    = "--password-file"
)

func makePasswordFlag(configDir string, pw *config.PasswordConfig) (out string, err error) {
	if pw == nil {
		return
	}

	if pw.File != nil {
		filename := filepath.Clean(filepath.Join(configDir, *pw.File))
		// Use %q printing verb in case the path has spaces.
		if strings.Contains(filename, " ") {
			filename = fmt.Sprintf("%q", filename)
		}
		out = formatPasswordFlag(passwordFileFlagKey, filename)
		return
	}

	// restic will prompt the user for a password.
	return
}

// formatPasswordFlag outputs a command line flag for a restic password. The
// caller may want to also write the value somewhere, like STDERR, in order to
// make it easier to manually intervene if necessary. To accommodate that use
// case, make it easy for the flag to be deconstructed into the key, value parts
// by using the `=` to delimit the split.
func formatPasswordFlag(key, val string) string { return key + "=" + val }

// quotePasswordFlag makes it easier to copy and paste a flag value into a
// terminal by putting quotes around the shell command or filename. Use single
// quotes here, rather than double quotes, because the makePasswordFlag function
// may have already put double quotes around the paths in this command.
func quotePasswordFlag(in string) (out string) {
	key, val, _ := strings.Cut(in, "=")
	out = key + `='` + val + `'`
	return
}
