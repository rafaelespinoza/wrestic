package config

import "sort"

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

type Datastore struct {
	Name         string                 `toml:"name"`
	Sources      []Source               `toml:"sources"`
	Destinations map[string]Destination `toml:"destinations"`
	Defaults     Defaults               `toml:"defaults"`

	parent *Defaults
}

func (d *Datastore) merge() Datastore {
	out, _ := makeDatastore(*d)
	mergeDefaults(&out.Defaults, d.parent)
	return out
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

type Source struct {
	Path string `toml:"path"`
}

type Destination struct {
	Name     string   `toml:"name"`
	Path     string   `toml:"path"`
	Defaults Defaults `toml:"defaults"`

	parent *Datastore
}

func (d *Destination) Merge() Defaults {
	var srcDefaults Defaults
	if d.parent != nil {
		mergedParent := d.parent.merge()
		d.parent = &mergedParent
		srcDefaults = d.parent.Defaults
	}

	out := duplicateDestination(*d)
	mergeDefaults(&out.Defaults, &srcDefaults)
	return out.Defaults
}

func duplicateDestination(in Destination) (out Destination) {
	out.Name = in.Name
	out.Path = in.Path
	out.Defaults = duplicateDefaults(in.Defaults)
	out.parent = in.parent
	return
}
