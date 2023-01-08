package config

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
)

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

		currDatastore.parent = &out.Defaults
		out.Datastores[storeName] = currDatastore

		for destName, dest := range currDatastore.Destinations {
			dest.parent = &currDatastore
			currDatastore.Destinations[destName] = dest
		}
	}
	return
}

type Params struct {
	Defaults   Defaults             `toml:"defaults"`
	Datastores map[string]Datastore `toml:"datastores"`
}

type Defaults struct {
	PasswordConfig *PasswordConfig `toml:"password-config"`
}

func mergeDefaults(dst, src *Defaults) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	} else if dst == nil && src != nil {
		dupe := duplicateDefaults(*src)
		*dst = dupe
		return
	}

	mergePasswordConfig(dst.PasswordConfig, src.PasswordConfig)
}

func duplicateDefaults(in Defaults) (out Defaults) {
	out.PasswordConfig = duplicatePasswordConfig(in.PasswordConfig)
	return
}

type PasswordConfig struct {
	File *string `toml:"file"`
}

func mergePasswordConfig(dst, src *PasswordConfig) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	} else if dst == nil && src != nil {
		dupe := duplicatePasswordConfig(src)
		if dupe != nil {
			*dst = *dupe
		}
		return
	}

	if dst.File == nil && src.File != nil {
		file := *src.File
		dst.File = &file
	}
	// Unlike the merging logic for the Command field, we don't want to let a
	// default File value override when the input file has that field set to an
	// empty string. Interpret this as user intention. If the user intends to
	// let source File value be overridden by a default File value, then the
	// corresponding field in the input file should be omitted entirely.
}

func duplicatePasswordConfig(in *PasswordConfig) (out *PasswordConfig) {
	out = &PasswordConfig{}

	if in == nil {
		return
	}

	if in.File != nil {
		file := *in.File
		out.File = &file
	}

	return
}
