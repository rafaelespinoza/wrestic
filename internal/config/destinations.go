package config

type Destination struct {
	// Name is not specified in the config file, but is implied by the
	// Destination's place in the config data. The intention is to ease
	// maintenance of the configuration file.
	Name     string   `toml:"-"`
	Path     string   `toml:"path"`
	Defaults Defaults `toml:"defaults"`

	parent *Datastore
}

func (d *Destination) Merge() (out Defaults, err error) {
	var srcDefaults Defaults
	if d.parent != nil {
		var mergedParent Datastore
		if mergedParent, err = d.parent.merge(); err != nil {
			return
		}
		d.parent = &mergedParent
		srcDefaults = d.parent.Defaults
	}

	dupe := duplicateDestination(*d)
	if err = mergeDefaults(&dupe.Defaults, &srcDefaults); err != nil {
		return
	}
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
