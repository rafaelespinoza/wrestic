package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

func TestInit(t *testing.T) {
	configDirname, err := config.Init(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("default base config dir", func(t *testing.T) {
		defaultDirname, err := config.DefaultBaseConfigDir()
		if err != nil {
			t.Fatal(err)
		}

		if configDirname == defaultDirname {
			t.Errorf("user should be able to override default base config dir")
		}
	})

	t.Run("structure and permissions", func(t *testing.T) {
		rootDir, err := os.Stat(configDirname)
		if err != nil {
			t.Fatal(err)
		}

		if !rootDir.IsDir() {
			t.Error("expected path to be a directory")
		}
		if got := rootDir.Mode().Perm(); got != 0700 {
			t.Errorf("wrong permissions for directory %s; got %d, expected %d", configDirname, got, 0700)
		}

		for _, basename := range []string{"secrets"} {
			subdirName := filepath.Clean(filepath.Join(configDirname, basename))
			subdir, err := os.Stat(subdirName)
			if err != nil {
				t.Fatalf("%s: %s", subdirName, err)
			}

			if !subdir.IsDir() {
				t.Errorf("expected path %s to be a directory", subdirName)
			}
			if got := subdir.Mode().Perm(); got != 0700 {
				t.Errorf("wrong permissions for directory %s; got %d, expected %d", subdirName, got, 0700)
			}
		}
	})
}
