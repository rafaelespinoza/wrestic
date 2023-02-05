package config

// Destination is a restic repository.
type Destination struct {
	// Name is not specified in the config file, but is implied by the
	// Destination's place in the config data. The intention is to ease
	// maintenance of the configuration file.
	Name string `toml:"-"`
	// Defaults are any configuration values specific to the Destination.
	// Unspecified fields will be merged in from the Datastore.
	Defaults Defaults `toml:"defaults"`
	// Path is the restic repository path.
	Path string `toml:"path"`

	parent *Datastore
}

// Merge combines the Defaults from the config file's top-level Defaults into
// the Datastore's Defaults, and then combines that into the Destination's
// Defaults. Any config values specified for the Destination are not overridden
// by the same config value specified in the Datastore.
func (d *Destination) Merge() (out Defaults, err error) {
	var srcDefaults Defaults
	if d.parent != nil {
		mergedParent := d.parent.merge()
		d.parent = &mergedParent
		srcDefaults = d.parent.Defaults
	}

	dupe := duplicateDestination(*d)
	mergeDefaults(&dupe.Defaults, &srcDefaults)
	out = dupe.Defaults
	return
}

func duplicateDestination(in Destination) (out Destination) {
	out.Name = in.Name
	out.Path = in.Path
	out.Defaults = duplicateDefaults(in.Defaults)
	out.parent = in.parent
	return
}

// BuildFlags merges in default config values and outputs a list of tuples
// representing the merged config.
func (d *Destination) BuildFlags(configDir string, subcmd string) ([]Flag, error) {
	defaults, err := d.Merge()
	if err != nil {
		return nil, err
	}

	out := []Flag{{Key: "repo", Val: d.Path}}

	if pwcmd, err := parsePasswordCommand(configDir, defaults.PasswordConfig); err != nil {
		return nil, err
	} else if pwcmd != "" {
		out = append(out, Flag{Key: "password-command", Val: pwcmd})
	}

	var restic interface {
		makeFlags(*ResticGlobal) ([]Flag, error)
	}

	switch subcmd {
	case "backup":
		restic = defaults.Restic.Backup
	case "check":
		restic = defaults.Restic.Check
	case "ls":
		restic = defaults.Restic.LS
	case "snapshots":
		restic = defaults.Restic.Snapshots
	case "stats":
		restic = defaults.Restic.Stats
	default:
		break
	}

	if restic == nil {
		return out, nil
	}

	resticFlags, err := restic.makeFlags(defaults.Restic.Global)
	if err != nil {
		return nil, err
	}

	return append(out, resticFlags...), nil
}
