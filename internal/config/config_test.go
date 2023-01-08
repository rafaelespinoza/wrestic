package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

func TestParse(t *testing.T) {
	t.Run("it works", func(t *testing.T) {
		file, err := os.Open(filepath.Clean(filepath.Join("testdata", "datastores.toml")))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = file.Close() }()

		actual, err := config.Parse(file)
		if err != nil {
			t.Fatal(err)
		}

		expectedDefaults := config.Defaults{
			PasswordConfig: &config.PasswordConfig{
				File: makePWConfigFile("secrets/default"),
			},
		}

		testDefaults(t, "", actual.Defaults, expectedDefaults)

		expectedDatastores := map[string]config.Datastore{
			"stuff": {
				Name:    "stuff",
				Sources: []config.Source{{Path: "/tmp/wrestic_test/testdata/srcdata/foo"}},
				Destinations: map[string]config.Destination{
					"alfa": {
						Name: "alfa",
						Path: "/tmp/wrestic_test/testdata/repos/alfa",
						Defaults: config.Defaults{
							PasswordConfig: &config.PasswordConfig{File: makePWConfigFile("secrets/a")},
						},
					},
					"bravo": {
						Name: "bravo",
						Path: "/tmp/wrestic_test/testdata/repos/bravo",
						Defaults: config.Defaults{
							PasswordConfig: &config.PasswordConfig{File: makePWConfigFile("secrets/a")},
						},
					},
				},
			},
			"things": {
				Name: "things",
				Sources: []config.Source{
					{Path: "/tmp/wrestic_test/testdata/srcdata/bar"},
					{Path: "/tmp/wrestic_test/testdata/srcdata/qux"},
				},
				Destinations: map[string]config.Destination{
					"charlie": {
						Name: "charlie",
						Path: "/tmp/wrestic_test/testdata/repos/charlie",
						Defaults: config.Defaults{
							PasswordConfig: &config.PasswordConfig{File: makePWConfigFile("secrets/b")},
						},
					},
				},
			},
		}

		if len(actual.Datastores) != len(expectedDatastores) {
			t.Fatalf("wrong number of Datastores; got %d, expected %d", len(actual.Datastores), len(expectedDatastores))
		}

		for key, got := range actual.Datastores {
			exp, ok := expectedDatastores[key]

			if !ok {
				t.Errorf("unexpected Datastore %q", key)
				continue
			}

			errPrefix := fmt.Sprintf(".Datastores[%q]", key)

			if got.Name != exp.Name {
				t.Errorf("%s wrong Name; got %q, expected %q", errPrefix, got.Name, exp.Name)
			}

			testSources(t, errPrefix+".Sources", got.Sources, exp.Sources)
			testDestinations(t, errPrefix+".Destinations", got.Destinations, exp.Destinations)
		}
	})

	t.Run("reject data with unknown keys", func(t *testing.T) {
		const input = `
[defaults]
[defaults.badkey]
file = 'secrets/defaultpassword'
`
		actual, err := config.Parse(strings.NewReader(input))
		if err == nil {
			t.Error("expected an error but got none")
		}

		testDefaults(t, "", actual.Defaults, config.Defaults{})
	})
}

func testDefaults(t *testing.T, errPrefix string, got, exp config.Defaults) {
	testPasswordConfig(t, errPrefix, got.PasswordConfig, exp.PasswordConfig)
}

func testPasswordConfig(t *testing.T, errPrefix string, got, exp *config.PasswordConfig) {
	if got == nil && exp == nil {
		return // test OK
	} else if got != nil && exp == nil {
		t.Fatalf("%s got %#v, expected %v", errPrefix, *got, exp)
	} else if got == nil && exp != nil {
		t.Fatalf("%s got %v, expected %#v", errPrefix, got, *exp)
	}

	if got.File == nil && exp.File == nil {
		// test OK
	} else if got.File != nil && exp.File == nil {
		t.Errorf("%s wrong File; got %q, expected %v", errPrefix, *got.File, exp.File)
	} else if got.File == nil && exp.File != nil {
		t.Errorf("%s wrong File; got %v, expected %q", errPrefix, got.File, *exp.File)
	} else if got.File != nil && exp.File != nil && *got.File != *exp.File {
		t.Errorf("%s wrong File; got %q, expected %q", errPrefix, *got.File, *exp.File)
	}
}
