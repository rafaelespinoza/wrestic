package config

import (
	"errors"
	"fmt"
	"os"
	"path"
)

// Init ensures that the application configuration directory structure is set up
// with an expected layout and access modes are restricted. It returns the path
// to the base configuration directory.
func Init(inConfigDir string) (configDir string, err error) {
	if inConfigDir != "" {
		configDir = inConfigDir
	} else {
		configDir, err = DefaultBaseConfigDir()
		if err != nil {
			return
		}
	}

	for _, dirpath := range []string{configDir, path.Join(configDir, "secrets")} {
		if err = prepDirectory(dirpath); err != nil {
			return
		}
	}

	return
}

// DefaultBaseConfigDir returns the local file system path to find configuration
// and data files specific to the application.
// It's based on the XDG base directory spec for the host system.
func DefaultBaseConfigDir() (out string, err error) {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return
	}

	out = path.Join(baseConfigDir, "wrestic")
	return
}

func prepDirectory(dirpath string) error {
	info, err := os.Stat(dirpath)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		return os.MkdirAll(dirpath, 0700)
	}

	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory, is %s", info.Mode().String())
	}

	// #nosec G302 -- This gosec rule seems to apply to files, but not for
	// directories. The directory must be executable so it can be looked at.
	return os.Chmod(dirpath, 0700)
}
