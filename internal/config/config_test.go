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
				Template: newString("cat {{ filename (index . 0) }}"),
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
							PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/a"}},
						},
					},
					"bravo": {
						Name: "bravo",
						Path: "/tmp/wrestic_test/testdata/repos/bravo",
						Defaults: config.Defaults{
							PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/a"}},
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
							PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/b"}},
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

	t.Run("parses PasswordConfig", func(t *testing.T) {
		const input = `
[defaults]
[defaults.password-config]
template = 'age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}'
args = ['secrets/id', 'secrets/foo']
`
		actual, err := config.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}

		expected := config.Params{
			Defaults: config.Defaults{
				PasswordConfig: &config.PasswordConfig{
					Template: newString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
					Args:     []string{"secrets/id", "secrets/foo"},
				},
			},
		}

		testDefaults(t, "", actual.Defaults, expected.Defaults)
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

	testStringPointer(t, errPrefix+".PasswordConfig.Template", got.Template, exp.Template)
	testStrings(t, errPrefix+".PasswordConfig.Args", got.Args, exp.Args)
}

func testStringPointer(t *testing.T, errPrefix string, got, exp *string) {
	if got == nil && exp == nil {
		return // test OK
	} else if got != nil && exp == nil {
		t.Errorf("%s got %q, expected %v", errPrefix, *got, exp)
	} else if got == nil && exp != nil {
		t.Errorf("%s got %v, expected %q", errPrefix, got, *exp)
	} else if *got != *exp {
		t.Errorf("%s got %q, expected %q", errPrefix, *got, *exp)
	}
}

func testStrings(t *testing.T, errPrefix string, actual, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("%s wrong number of Flags; got %d, expected %d", errPrefix, len(actual), len(expected))
		return
	}

	for i, got := range actual {
		exp := expected[i]

		if got != exp {
			t.Errorf("%s[%d] got %q, expected %q", errPrefix, i, got, exp)
		}
	}
}
