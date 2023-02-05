package config

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
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

func mergeDefaults(dst, src *Defaults) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	mergeConfig(dst.PasswordConfig, src.PasswordConfig)
	mergeConfig(dst.Restic, src.Restic)
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

func duplicatePasswordConfig(in *PasswordConfig) (out *PasswordConfig) {
	out = &PasswordConfig{}
	if in == nil {
		return
	}

	mergeConfig(out, in)
	return
}

type mergeableConfig interface {
	PasswordConfig | ResticDefaults
}

func mergeConfig[C mergeableConfig](dst, src *C) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	transformer := mergeTransformer{
		okTypes: []reflect.Type{
			reflect.TypeOf(new([]string)),
			reflect.TypeOf(new(string)),
		},
	}

	// The library, github.com/imdario/mergo, may return an error if:
	// - the type of the 1st input is not a pointer to a struct.
	// - the types of both inputs are not same type structs.
	// Neither should be concerns for this tool. Though, make some noise.
	err := mergo.Merge(dst, src, mergo.WithTransformers(transformer))
	if err != nil {
		fmt.Fprintf(os.Stderr, "wrestic: %#v\n", err)
	}
	return
}

type mergeTransformer struct{ okTypes []reflect.Type }

func (t mergeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	var typeOK bool
	for _, okType := range t.okTypes {
		if typ == okType {
			typeOK = true
			break
		}
	}
	if !typeOK {
		return nil
	}

	return func(dst, src reflect.Value) error {
		// Only merge the values if the configuration file does not specify the
		// field. If dst is an initialized but zero-length slice, do nothing.
		if dst.CanSet() && dst.IsNil() {
			dst.Set(src)
			return nil
		}

		return nil
	}
}
