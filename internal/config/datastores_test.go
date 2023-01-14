package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

func TestSelectDatastores(t *testing.T) {
	// some expected values here.
	expAlfa := config.Destination{
		Name: "alfa",
		Path: "/tmp/wrestic_test/testdata/repos/alfa",
		Defaults: config.Defaults{
			PasswordConfig: &config.PasswordConfig{
				Args: []string{"secrets/a"},
			},
		},
	}
	expBravo := config.Destination{
		Name: "bravo",
		Path: "/tmp/wrestic_test/testdata/repos/bravo",
		Defaults: config.Defaults{
			PasswordConfig: &config.PasswordConfig{
				Args: []string{"secrets/a"},
			},
		},
	}
	expCharlie := config.Destination{
		Name: "charlie",
		Path: "/tmp/wrestic_test/testdata/repos/charlie",
		Defaults: config.Defaults{
			PasswordConfig: &config.PasswordConfig{
				Args: []string{"secrets/b"},
			},
		},
	}

	tests := []struct {
		name        string
		inNames     []string
		inDestnames []string
		expected    []config.Datastore
	}{
		{
			name:    "names: specify one",
			inNames: []string{"stuff"},
			expected: []config.Datastore{
				{
					Name: "stuff",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/foo"},
					},
					Destinations: map[string]config.Destination{
						"alfa":  expAlfa,
						"bravo": expBravo,
					},
				},
			},
		},
		{
			name:    "names: specify another",
			inNames: []string{"things"},
			expected: []config.Datastore{
				{
					Name: "things",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/bar"},
						{Path: "/tmp/wrestic_test/testdata/srcdata/qux"},
					},
					Destinations: map[string]config.Destination{
						"charlie": expCharlie,
					},
				},
			},
		},
		{
			name:        "destnames: specify all",
			inDestnames: []string{"alfa", "bravo", "charlie"},
			expected: []config.Datastore{
				{
					Name: "stuff",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/foo"},
					},
					Destinations: map[string]config.Destination{
						"alfa":  expAlfa,
						"bravo": expBravo,
					},
				},
				{
					Name: "things",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/bar"},
						{Path: "/tmp/wrestic_test/testdata/srcdata/qux"},
					},
					Destinations: map[string]config.Destination{
						"charlie": expCharlie,
					},
				},
			},
		},
		{
			name:        "specify some destinations",
			inDestnames: []string{"alfa", "bravo"},
			expected: []config.Datastore{
				{
					Name: "stuff",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/foo"},
					},
					Destinations: map[string]config.Destination{
						"alfa":  expAlfa,
						"bravo": expBravo,
					},
				},
			},
		},
		{
			name:        "specify destinations involving two datastores",
			inDestnames: []string{"bravo", "charlie"},
			expected: []config.Datastore{
				{
					Name: "stuff",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/foo"},
					},
					Destinations: map[string]config.Destination{
						"bravo": expBravo,
					},
				},
				{
					Name: "things",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/bar"},
						{Path: "/tmp/wrestic_test/testdata/srcdata/qux"},
					},
					Destinations: map[string]config.Destination{
						"charlie": expCharlie,
					},
				},
			},
		},
		{
			name:        "names and destnames",
			inNames:     []string{"things"},
			inDestnames: []string{"charlie"},
			expected: []config.Datastore{
				{
					Name: "things",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/bar"},
						{Path: "/tmp/wrestic_test/testdata/srcdata/qux"},
					},
					Destinations: map[string]config.Destination{
						"charlie": expCharlie,
					},
				},
			},
		},
		{
			name: "all",
			expected: []config.Datastore{
				{
					Name: "stuff",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/foo"},
					},
					Destinations: map[string]config.Destination{
						"alfa":  expAlfa,
						"bravo": expBravo,
					},
				},
				{
					Name: "things",
					Sources: []config.Source{
						{Path: "/tmp/wrestic_test/testdata/srcdata/bar"},
						{Path: "/tmp/wrestic_test/testdata/srcdata/qux"},
					},
					Destinations: map[string]config.Destination{
						"charlie": expCharlie,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := os.Open(filepath.Clean(filepath.Join("testdata", "datastores.toml")))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			params, err := config.Parse(file)
			if err != nil {
				t.Fatal(err)
			}

			actual := config.SelectDatastores(params.Datastores, test.inNames, test.inDestnames)

			if len(actual) != len(test.expected) {
				t.Fatalf("wrong number of Datastores; got %d, expected %d", len(actual), len(test.expected))
			}

			for i, got := range actual {
				errPrefix := fmt.Sprintf("item [%d]", i)

				exp := test.expected[i]
				if got.Name != exp.Name {
					t.Errorf("%s wrong Name; got %q, expected %q", errPrefix, got.Name, exp.Name)
				}

				testSources(t, errPrefix, got.Sources, exp.Sources)
				testDestinations(t, errPrefix, got.Destinations, exp.Destinations)
			}
		})
	}
}

func TestDestinationMerge(t *testing.T) {
	type testCase struct {
		name     string
		inputs   string
		expected config.Defaults
	}

	tests := []testCase{
		{
			name: "PasswordConfig.Args: use Destination values",
			inputs: `
[defaults.password-config]
args = ['secrets/defaults']

[datastores.stuff]
name = 'stuff'

[datastores.stuff.defaults.password-config]
args = ['secrets/stuff']

[datastores.stuff.destinations.foo]
name = 'foo'

[datastores.stuff.destinations.foo.defaults.password-config]
args = ['secrets/foo']
`,
			expected: config.Defaults{
				PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/foo"}},
			},
		},
		{
			name: "PasswordConfig.Args: use Datastore values",
			inputs: `
[defaults.password-config]
args = ['secrets/defaults']

[datastores.stuff]
name = 'stuff'

[datastores.stuff.defaults.password-config]
args = ['secrets/stuff']

[datastores.stuff.destinations.foo]
name = 'foo'
`,
			expected: config.Defaults{
				PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/stuff"}},
			},
		},
		{
			name: "PasswordConfig.Args: use default values",
			inputs: `
[defaults.password-config]
args = ['secrets/defaults']

[datastores.stuff]
name = 'stuff'

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
name = 'foo'

[datastores.stuff.destinations.foo.defaults.password-config]
`,
			expected: config.Defaults{
				PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/defaults"}},
			},
		},
		{
			name: "PasswordConfig.Template: use Destination values",
			inputs: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]
name = 'stuff'

[datastores.stuff.defaults.password-config]
template = 'run_stuff'

[datastores.stuff.destinations.foo]
name = 'foo'

[datastores.stuff.destinations.foo.defaults.password-config]
template = 'run_foo'
`,
			expected: config.Defaults{
				PasswordConfig: &config.PasswordConfig{Template: newString("run_foo")},
			},
		},
		{
			name: "PasswordConfig.Template: use Datastore values",
			inputs: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]
name = 'stuff'

[datastores.stuff.defaults.password-config]
template = 'run_stuff'

[datastores.stuff.destinations.foo]
name = 'foo'

[datastores.stuff.destinations.foo.defaults.password-config]
`,
			expected: config.Defaults{
				PasswordConfig: &config.PasswordConfig{Template: newString("run_stuff")},
			},
		},
		{
			name: "PasswordConfig.Template: use default values",
			inputs: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]
name = 'stuff'

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
name = 'foo'

[datastores.stuff.destinations.foo.defaults.password-config]
`,
			expected: config.Defaults{
				PasswordConfig: &config.PasswordConfig{Template: newString("run_defaults")},
			},
		},
		{
			name: "PasswordConfig.Template: Destination has empty value",
			inputs: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]
name = 'stuff'

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
name = 'foo'

[datastores.stuff.destinations.foo.defaults.password-config]
template = '' # user specifies empty value for some reason
`,
			expected: config.Defaults{
				PasswordConfig: &config.PasswordConfig{Template: newString("")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params, err := config.Parse(strings.NewReader(test.inputs))
			if err != nil {
				t.Fatal(err)
			}

			if err != nil {
				t.Fatal(err)
			}

			dest := params.Datastores["stuff"].Destinations["foo"]
			merged := dest.Merge()
			testDefaults(t, "", merged, test.expected)
		})
	}
}

func testSources(t *testing.T, errPrefix string, got, exp []config.Source) {
	t.Helper()

	if len(got) != len(exp) {
		t.Errorf("%s wrong number of Sources; got %d, expected %d", errPrefix, len(got), len(exp))
		return
	}

	for j, gotSrc := range got {
		expSrc := exp[j]

		if gotSrc.Path != expSrc.Path {
			t.Errorf("%s[%d] wrong Source.Path; got %q, expected %q", errPrefix, j, gotSrc.Path, expSrc.Path)
		}
	}
}

func testDestinations(t *testing.T, errPrefix string, got, exp map[string]config.Destination) {
	t.Helper()

	if len(got) != len(exp) {
		t.Errorf("%s wrong number of Destinations; got %d, expected %d", errPrefix, len(got), len(exp))
		return
	}

	for destName, gotDest := range got {
		nestedErrPrefix := fmt.Sprintf("%s[%q]", errPrefix, destName)

		expDest, ok := exp[destName]
		if !ok {
			t.Errorf("%s unexpected destination %q", nestedErrPrefix, destName)
			continue
		}

		if gotDest.Name != expDest.Name {
			t.Errorf("%s wrong Destination.Path; got %q, expected %q", nestedErrPrefix, gotDest.Name, expDest.Name)
		}

		if gotDest.Path != expDest.Path {
			t.Errorf("%s wrong Destination.Path; got %q, expected %q", nestedErrPrefix, gotDest.Path, expDest.Path)
		}

		testDefaults(t, nestedErrPrefix, gotDest.Defaults, expDest.Defaults)
	}
}

// newString is a convenience func for setting up expected data in a test.
// The go compiler says:
// "invalid operation: cannot take address of ("foo") (untyped string constant "foo")".
func newString(in string) (out *string) { return &in }
