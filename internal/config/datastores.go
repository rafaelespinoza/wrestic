package config

import "sort"

// SelectDatastores filters for Datastores with an exactly-matching Name in
// names, or have a Destination exactly matching one in destNames. If names is
// length 0, then the name of the Datastore is not considered. If destNames is
// length 0, then the name of the Destination is not considered either.
func SelectDatastores(datastores map[string]Datastore, names, destNames []string) (out []Datastore) {
	for name, datastore := range datastores {
		if len(names) < 1 && len(destNames) < 1 {
			// capture everything.
			if dupe, ok := makeDatastore(datastore); ok {
				out = append(out, dupe)
			}
		} else if len(destNames) < 1 {
			// only capture Datastore if match on .Name, capture all Destinations.
			for _, targetName := range names {
				if name == targetName {
					if dupe, ok := makeDatastore(datastore); ok {
						out = append(out, dupe)
					}
				}
			}
		} else if len(names) < 1 {
			// capture the Datastore regardless of .Name, but filter matching Destinations.
			if dupe, ok := makeDatastore(datastore, destNames...); ok {
				out = append(out, dupe)
			}
		} else {
			// only capture Datastore if match on .Name, but filter matching Destinations.
			for _, targetName := range names {
				if name == targetName {
					if dupe, ok := makeDatastore(datastore, destNames...); ok {
						out = append(out, dupe)
					}
				}
			}
		}
	}

	// Iteration order for maps is never guaranteed to be the same between one
	// loop and another loop. Ensure predictable order here.
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return
}

// Datastore is an abstraction for source paths to backup and the destination
// restic repositories.
type Datastore struct {
	// Defaults are any configuration values specific to the Datastore.
	// Unspecified fields will be merged in from top-level Defaults.
	Defaults     Defaults               `toml:"defaults"`
	Sources      []Source               `toml:"sources"`
	Destinations map[string]Destination `toml:"destinations"`
	// Name is not specified in the config file, but is implied by the
	// Datastore's place in the config data. The intention is to ease
	// maintenance of the configuration file.
	Name string `toml:"-"`

	parent *Defaults
}

func (d *Datastore) merge() (out Datastore) {
	out, _ = makeDatastore(*d)
	mergeDefaults(&out.Defaults, d.parent)
	return
}

// makeDatastore also returns a boolean to signal to the caller that at least 1
// destination matched the input destNames. This will also be true if the
// destNames input is length 0.
func makeDatastore(in Datastore, destNames ...string) (out Datastore, destMatch bool) {
	srcs := make([]Source, len(in.Sources))
	copy(srcs, in.Sources)

	dests := make(map[string]Destination)
	for name, dest := range in.Destinations {
		if len(destNames) < 1 {
			dests[name] = dest
			destMatch = true
			continue
		}

		for _, targetName := range destNames {
			if name == targetName {
				dests[name] = dest
				destMatch = true
			}
		}
	}

	out = Datastore{
		Name:         in.Name,
		Sources:      srcs,
		Destinations: dests,
		Defaults:     duplicateDefaults(in.Defaults),
		parent:       in.parent,
	}

	return
}

// Source is something to backup to a restic repository.
type Source struct {
	// Path should be an absolute path to either a file or directory.
	Path string `toml:"path"`
}
