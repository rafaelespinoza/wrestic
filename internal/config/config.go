package config

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
)

// Parse not only constructs Params from configuration file data, it also
// prepares some internal state necessary for merging data later on.
func Parse(r io.Reader) (out Params, err error) {
	meta, err := toml.NewDecoder(r).Decode(&out)
	if err != nil {
		return
	}

	unexpectedKeys := meta.Undecoded()
	if len(unexpectedKeys) > 0 {
		err = fmt.Errorf("unexpected keys %q", unexpectedKeys)
		return
	}

	for storeName, datastore := range out.Datastores {
		// Ensure consistent Datastore, otherwise goofy things happen such as
		// the wrong parent value getting assigned to a Destination. Direct use
		// of the variable defined in the range statement above seems to lead to
		// nondeterministic results.
		currDatastore := datastore

		// Why not specify the Name field in the configuration file? It's easier
		// to maintain the configuration file if there are less things to
		// specify. Here, we want the Name to be unique among a set. Naturally,
		// a map data structure provides this. Use the key as Name here.
		currDatastore.Name = storeName

		currDatastore.parent = &out.Defaults
		out.Datastores[storeName] = currDatastore

		for destName, dest := range currDatastore.Destinations {
			dest.parent = &currDatastore
			if dest.Name != destName {
				// Same reasoning as the datastore.Name field described above.
				// Ease maintainence of the configuration file.
				dest.Name = destName
			}

			currDatastore.Destinations[destName] = dest
		}
	}
	return
}

// EncodeTOML formats in to TOML and writes to w.
func EncodeTOML(w io.Writer, in any) error {
	enc := toml.NewEncoder(w)
	return enc.Encode(in)
}

// Params represents the entire config file after it's parsed.
type Params struct {
	Defaults   Defaults             `toml:"defaults"`
	Datastores map[string]Datastore `toml:"datastores"`
}

// Defaults defines configuration values.
type Defaults struct {
	PasswordConfig *PasswordConfig `toml:"password-config"`
	Restic         *ResticDefaults `toml:"restic"`
}

func mergeDefaults(dst, src *Defaults) error {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return nil
	} else if dst == nil && src != nil {
		dupe := duplicateDefaults(*src)
		dst = &dupe
		return nil
	}

	mergePasswordConfig(dst.PasswordConfig, src.PasswordConfig)
	return mergeResticDefaults(dst.Restic, src.Restic)
}

func duplicateDefaults(in Defaults) (out Defaults) {
	out.PasswordConfig = duplicatePasswordConfig(in.PasswordConfig)
	out.Restic = duplicateResticDefaults(in.Restic)

	return
}

// PasswordConfig is a specialized configuration type to manage the
// password-command flag for restic subcommands.
type PasswordConfig struct {
	// Template is the password-command (a restic flag) to run. It is parsed by
	// package text/template from the golang standard library. Arguments may be
	// interjected into placeholders delimited by "{{" and "}}".
	Template *string `toml:"template"`
	// Args are positional arguments that may be referenced by placeholders in a
	// template string.
	Args []string `toml:"args"`
}

func mergePasswordConfig(dst, src *PasswordConfig) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	if dst == nil && src != nil {
		dupe := duplicatePasswordConfig(src)
		if dupe != nil {
			dst = dupe
		}
		return
	}

	if dst.Template == nil {
		dst.Template = src.Template
	}

	if len(dst.Args) < 1 {
		dst.Args = duplicateStrings(src.Args)
	}
}

func duplicatePasswordConfig(in *PasswordConfig) (out *PasswordConfig) {
	out = &PasswordConfig{}

	if in == nil {
		return
	}

	var tmpl *string
	if in.Template != nil {
		tmpl = in.Template
	}
	out = &PasswordConfig{
		Template: tmpl,
		Args:     duplicateStrings(in.Args),
	}

	return
}

func duplicateStrings(in []string) (out []string) {
	out = make([]string, len(in))
	copy(out, in)
	return
}

func duplicateOptionMap(in []map[string]string) (out []map[string]string) {
	out = make([]map[string]string, len(in))

	for i, sources := range in {
		dest := make(map[string]string)

		for subkey, subval := range sources {
			dest[subkey] = subval
		}

		out[i] = dest
	}

	return
}
