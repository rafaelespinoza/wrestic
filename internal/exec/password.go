package exec

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

const (
	// pwcmdConfigFileKey is the name of a key from the configuration file
	// that's used in producing a password command.
	pwcmdConfigFileKey = "password-config"
	// pwcmdFlagKey is the name of the flag that the restic CLI expects to find
	// the password command.
	pwcmdFlagKey = "--password-command"
)

func makePasswordFlag(configDir string, pw *config.PasswordConfig) (out string, err error) {
	if pw == nil || pw.Template == nil {
		return
	}

	var tmpl *template.Template
	fns := template.FuncMap{
		"filename": func(argFilename string) (out string) { return formatFilenameFlag(configDir, argFilename) },
	}

	tmpl, err = template.New(pwcmdFlagKey).Funcs(fns).Parse(*pw.Template)
	if err != nil {
		err = fmt.Errorf("%w: %s.template is invalid", err, pwcmdConfigFileKey)
		return
	}

	var bld strings.Builder
	if err = tmpl.Execute(&bld, pw.Args); err != nil {
		if xerr, ok := err.(template.ExecError); ok {
			err = fmt.Errorf("%w: %s.template does not agree with args", xerr, pwcmdConfigFileKey)
		}
		return
	}

	out = formatPasswordFlag(pwcmdFlagKey, bld.String())
	return
}

// formatPasswordFlag outputs a command line flag for a restic password. The
// caller may want to also write the value somewhere, like STDERR, in order to
// make it easier to manually intervene if necessary. To accommodate that use
// case, make it easy for the flag to be deconstructed into the key, value parts
// by using the `=` to delimit the split.
func formatPasswordFlag(key, val string) string {
	return key + "=" + val
}

func formatFilenameFlag(configDir, filename string) string {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(configDir, filename)
	}

	filename = filepath.Clean(filename)

	if strings.Contains(filename, " ") {
		filename = fmt.Sprintf("%q", filename)
	}

	return filename
}

// quotePasswordFlag makes it easier to copy and paste a flag value into a
// terminal by putting quotes around the shell command or filename. Use single
// quotes here, rather than double quotes, because the makePasswordFlag function
// may have already put double quotes around the paths in this command.
func quotePasswordFlag(in string) (out string) {
	key, val, _ := strings.Cut(in, "=")
	out = key + `='` + val + `'`
	return
}
