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
				Template: pointTo("cat {{ filenameArg 0 }}"),
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
			t.Fatal("expected an error but got none")
		}
		if !strings.Contains(err.Error(), "unexpected keys") {
			t.Errorf("expected error message %q to contain %q", err, "unexpected keys")
		}

		testDefaults(t, "", actual.Defaults, config.Defaults{})
	})

	t.Run("parses PasswordConfig", func(t *testing.T) {
		const input = `
[defaults]
[defaults.password-config]
template = 'age -d -i {{ filename "secrets/id" }} {{ filenameArg 0 }}'
args = ['secrets/foo']
`
		actual, err := config.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}

		expected := config.Params{
			Defaults: config.Defaults{
				PasswordConfig: &config.PasswordConfig{
					Template: pointTo(`age -d -i {{ filename "secrets/id" }} {{ filenameArg 0 }}`),
					Args:     []string{"secrets/foo"},
				},
			},
		}

		testDefaults(t, "", actual.Defaults, expected.Defaults)
	})
}

func testDefaults(t *testing.T, errPrefix string, got, exp config.Defaults) {
	t.Helper()

	testPasswordConfig(t, errPrefix+".PasswordConfig", got.PasswordConfig, exp.PasswordConfig)
	testResticDefaults(t, errPrefix+".Restic", got.Restic, exp.Restic)
}

func testPasswordConfig(t *testing.T, errPrefix string, got, exp *config.PasswordConfig) {
	t.Helper()

	if got == nil && exp == nil {
		return // test OK
	} else if got != nil && exp == nil {
		t.Fatalf("%s got %#v, expected %v", errPrefix, *got, exp)
	} else if got == nil && exp != nil {
		t.Fatalf("%s got %v, expected %#v", errPrefix, got, *exp)
	}

	testStringPointer(t, errPrefix+".Template", got.Template, exp.Template)
	testStrings(t, errPrefix+".Args", got.Args, exp.Args)
}

// A primitive is any builtin type that is also the field type on a struct type
// from this package.
type primitive interface {
	bool | string | int | uint
}

// pointTo is a convenience func for setting up data in a test.
func pointTo[P primitive](in P) *P { return &in }
func pointToStrings(in ...string) *[]string {
	if len(in) < 1 {
		return &[]string{}
	}

	return &in
}

func testStringPointer(t *testing.T, errPrefix string, got, exp *string) {
	t.Helper()

	if got == nil && exp == nil {
		// test OK
	} else if got != nil && exp == nil {
		t.Errorf("%s got %v, expected %v", errPrefix, *got, exp)
	} else if got == nil && exp != nil {
		t.Errorf("%s got %v, expected %v", errPrefix, got, *exp)
	} else if got != nil && exp != nil && *got != *exp {
		t.Errorf("%s got %v, expected %v", errPrefix, *got, *exp)
	}
}

func testStrings(t *testing.T, errPrefix string, actual, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("%s wrong length; got %d, expected %d", errPrefix, len(actual), len(expected))
		return
	}

	for i, got := range actual {
		exp := expected[i]

		if got != exp {
			t.Errorf("%s[%d] got %q, expected %q", errPrefix, i, got, exp)
		}
	}
}
