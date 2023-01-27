package config_test

import (
	"strings"
	"testing"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

func TestDestination(t *testing.T) {
	type mergeTestcase struct {
		expDefaults       config.Defaults
		expError          bool
		expErrMsgContains string
	}

	type flagsTestcase struct {
		inConfigDir       string
		inSubcommand      string
		expFlags          []config.Flag
		expError          bool
		expErrMsgContains string
	}

	type testCase struct {
		name              string
		inputFileContents string
		merge             mergeTestcase
		flags             flagsTestcase
	}

	runTest := func(t *testing.T, test testCase) {
		t.Helper()

		params, err := config.Parse(strings.NewReader(test.inputFileContents))
		if err != nil {
			t.Fatal(err)
		}

		dest := params.Datastores["stuff"].Destinations["foo"]

		merged, err := dest.Merge()
		if err != nil && !test.merge.expError {
			t.Fatal(err)
		} else if err == nil && test.merge.expError {
			t.Error("expected an error from Merge")
		} else if err != nil && test.merge.expError && !strings.Contains(err.Error(), test.merge.expErrMsgContains) {
			t.Errorf("expected error message %q to contain %q", err, test.merge.expErrMsgContains)
		}
		testDefaults(t, "", merged, test.merge.expDefaults)

		flags, err := dest.BuildFlags(test.flags.inConfigDir, test.flags.inSubcommand)
		if err != nil && !test.flags.expError {
			t.Fatal(err)
		} else if err == nil && test.flags.expError {
			t.Error("expected an error from BuildFlags")
		} else if err != nil && test.flags.expError && !strings.Contains(err.Error(), test.flags.expErrMsgContains) {
			t.Errorf("expected error message %q to contain %q", err, test.flags.expErrMsgContains)
		}
		testFlags(t, "Flags", flags, test.flags.expFlags)
	}

	t.Run("PasswordConfig.Args", func(t *testing.T) {
		tests := []testCase{
			{
				name: "use Destination values",
				inputFileContents: `
[defaults.password-config]
args = ['secrets/defaults']

[datastores.stuff]

[datastores.stuff.defaults.password-config]
args = ['secrets/stuff']

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
args = ['secrets/foo']
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/foo"}},
						Restic:         &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					expFlags: []config.Flag{{Key: "repo", Val: "test"}},
				},
			},
			{
				name: "use Datastore values",
				inputFileContents: `
[defaults.password-config]
args = ['secrets/defaults']

[datastores.stuff]

[datastores.stuff.defaults.password-config]
args = ['secrets/stuff']

[datastores.stuff.destinations.foo]
path = 'test'
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/stuff"}},
						Restic:         &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					expFlags: []config.Flag{{Key: "repo", Val: "test"}},
				},
			},
			{
				name: "use default values",
				inputFileContents: `
[defaults.password-config]
args = ['secrets/defaults']

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{Args: []string{"secrets/defaults"}},
						Restic:         &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					expFlags: []config.Flag{{Key: "repo", Val: "test"}},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) { runTest(t, test) })
		}
	})

	t.Run("PasswordConfig.Template", func(t *testing.T) {
		tests := []testCase{
			{
				name: "use Destination values",
				inputFileContents: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]

[datastores.stuff.defaults.password-config]
template = 'run_stuff'

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
template = 'run_foo'
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{Template: pointTo("run_foo")},
						Restic:         &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					expFlags: []config.Flag{
						{Key: "repo", Val: "test"},
						{Key: "password-command", Val: "run_foo"},
					},
				},
			},
			{
				name: "use Datastore values",
				inputFileContents: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]

[datastores.stuff.defaults.password-config]
template = 'run_stuff'

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{Template: pointTo("run_stuff")},
						Restic:         &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					expFlags: []config.Flag{
						{Key: "repo", Val: "test"},
						{Key: "password-command", Val: "run_stuff"},
					},
				},
			},
			{
				name: "use default values",
				inputFileContents: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{Template: pointTo("run_defaults")},
						Restic:         &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					expFlags: []config.Flag{
						{Key: "repo", Val: "test"},
						{Key: "password-command", Val: "run_defaults"},
					},
				},
			},
			{
				name: "Destination has empty value",
				inputFileContents: `
[defaults.password-config]
template = 'run_defaults'

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
template = '' # user specifies empty value for some reason
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{Template: pointTo("")},
						Restic:         &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					expFlags: []config.Flag{
						{Key: "repo", Val: "test"},
					},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) { runTest(t, test) })
		}
	})

	t.Run("PasswordConfig paths", func(t *testing.T) {
		tests := []testCase{
			{
				name: "absolute",
				inputFileContents: `
[defaults]

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
template = 'age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}'
args = ['/elsewhere/secrets/id', '/elsewhere/secrets/foo']
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{
							Template: pointTo(`age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}`),
							Args:     []string{"/elsewhere/secrets/id", "/elsewhere/secrets/foo"},
						},
						Restic: &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					inConfigDir:  "/tmp/config_place",
					inSubcommand: "ls",
					expFlags: []config.Flag{
						{Key: "repo", Val: "test"},
						{Key: "password-command", Val: "age -d -i /elsewhere/secrets/id /elsewhere/secrets/foo"},
					},
				},
			},
			{
				name: "relative",
				inputFileContents: `
[defaults]

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
template = 'age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}'
args = ['secrets/id', 'secrets/foo']
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{
							Template: pointTo(`age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}`),
							Args:     []string{"secrets/id", "secrets/foo"},
						},
						Restic: &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					inConfigDir:  "/tmp/config_place",
					inSubcommand: "ls",
					expFlags: []config.Flag{
						{Key: "repo", Val: "test"},
						{Key: "password-command", Val: "age -d -i /tmp/config_place/secrets/id /tmp/config_place/secrets/foo"},
					},
				},
			},
			{
				name: "has spaces",
				inputFileContents: `
[defaults]

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
template = 'age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}'
args = ['secrets/id', 'secrets/foo']
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{
							Template: pointTo(`age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}`),
							Args:     []string{"secrets/id", "secrets/foo"},
						},
						Restic: &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					inConfigDir:  "/Users/foobar/Library/Application Support/wrestic",
					inSubcommand: "ls",
					expFlags: []config.Flag{
						{Key: "repo", Val: "test"},
						{Key: "password-command", Val: `age -d -i "/Users/foobar/Library/Application Support/wrestic/secrets/id" "/Users/foobar/Library/Application Support/wrestic/secrets/foo"`},
					},
				},
			},
			{
				name: "invalid",
				inputFileContents: `
[defaults]

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
template = 'age -d -i {{ does_not_work 0 }} {{ filenameArg 1 }}'
args = ['/elsewhere/secrets/id', '/elsewhere/secrets/foo']
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{
							Template: pointTo(`age -d -i {{ does_not_work 0 }} {{ filenameArg 1 }}`),
							Args:     []string{"/elsewhere/secrets/id", "/elsewhere/secrets/foo"},
						},
						Restic: &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					inSubcommand:      "ls",
					inConfigDir:       "/tmp/config_place",
					expFlags:          []config.Flag{},
					expError:          true,
					expErrMsgContains: "template is invalid",
				},
			},
			{
				name: "invalid - does not agree",
				inputFileContents: `
[defaults]

[datastores.stuff]

[datastores.stuff.defaults.password-config]

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults.password-config]
template = 'age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}'
args = ['/elsewhere/secrets/id']
`,
				merge: mergeTestcase{
					expDefaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{
							Template: pointTo(`age -d -i {{ filenameArg 0 }} {{ filenameArg 1 }}`),
							Args:     []string{"/elsewhere/secrets/id"},
						},
						Restic: &config.ResticDefaults{},
					},
				},
				flags: flagsTestcase{
					inSubcommand:      "ls",
					inConfigDir:       "/tmp/config_place",
					expFlags:          []config.Flag{},
					expError:          true,
					expErrMsgContains: "template does not agree with args",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) { runTest(t, test) })
		}
	})

	t.Run("Restic.Backup", func(t *testing.T) {
		runTest(t, testCase{
			name: "it works",
			inputFileContents: `
[defaults]
restic.global = { verbose = 2 }
restic.backup = { dry-run = true }

[datastores.stuff.defaults]
restic.backup.iexclude = ['charlie', 'DS_Store']
restic.backup.tag = ['foo', 'bar']

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults]
restic.backup.host = 'custom_host'
`,
			merge: mergeTestcase{
				expDefaults: config.Defaults{
					PasswordConfig: &config.PasswordConfig{},
					Restic: &config.ResticDefaults{
						Global: &config.ResticGlobal{
							Verbose: pointTo(2),
						},
						Backup: &config.ResticBackup{
							DryRun:   pointTo(true),
							Host:     pointTo("custom_host"),
							Iexclude: pointToStrings("charlie", "DS_Store"),
							Tag:      pointToStrings("foo", "bar"),
						},
					},
				},
			},
			flags: flagsTestcase{
				inSubcommand: "backup",
				expFlags: []config.Flag{
					{Key: "repo", Val: "test"},
					{Key: "dry-run", Val: "true"},
					{Key: "host", Val: "custom_host"},
					{Key: "iexclude", Val: "charlie"},
					{Key: "iexclude", Val: "DS_Store"},
					{Key: "tag", Val: "foo"},
					{Key: "tag", Val: "bar"},
					{Key: "verbose", Val: "2"},
				},
			},
		})
	})

	t.Run("Restic.Check", func(t *testing.T) {
		runTest(t, testCase{
			inputFileContents: `
[defaults]
restic.global = { no-lock = true }
restic.check = { read-data = false }

[datastores.stuff.defaults]
restic.check = { with-cache = true }

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults]
restic.check = { read-data = true, read-data-subset = '15%' }
`,
			merge: mergeTestcase{
				expDefaults: config.Defaults{
					PasswordConfig: &config.PasswordConfig{},
					Restic: &config.ResticDefaults{
						Global: &config.ResticGlobal{NoLock: pointTo(true)},
						Check: &config.ResticCheck{
							ReadData:       pointTo(true),
							ReadDataSubset: pointTo("15%"),
							WithCache:      pointTo(true),
						},
					},
				},
			},
			flags: flagsTestcase{
				inSubcommand: "check",
				expFlags: []config.Flag{
					{Key: "repo", Val: "test"},
					{Key: "no-lock", Val: "true"},
					{Key: "read-data", Val: "true"},
					{Key: "read-data-subset", Val: "15%"},
					{Key: "with-cache", Val: "true"},
				},
			},
		})
	})

	t.Run("Restic.LS", func(t *testing.T) {
		runTest(t, testCase{
			inputFileContents: `
[defaults]
restic.global = { json = true }
restic.ls = { long = true, recursive = true }

[datastores.stuff.defaults]
restic.ls = { host = ['other_host'] }

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults]
restic.global = { pack-size = 4, tls-client-cert = 'client.crt' }
restic.global.option = [{ 'local.layout' = 'test' }, { 's3.connections' = '3' }]
restic.ls = { path = ['/echo', '/foxtrot'], tag = ['foo'] }
`,
			merge: mergeTestcase{
				expDefaults: config.Defaults{
					PasswordConfig: &config.PasswordConfig{},
					Restic: &config.ResticDefaults{
						Global: &config.ResticGlobal{
							JSON:     pointTo(true),
							PackSize: pointTo(uint(4)),
							Option: &([]map[string]string{
								{"local.layout": "test"}, {"s3.connections": "3"},
							}),
							TLSClientCert: pointTo("client.crt"),
						},
						LS: &config.ResticLS{
							Host:      pointToStrings("other_host"),
							Long:      pointTo(true),
							Path:      pointToStrings("/echo", "/foxtrot"),
							Recursive: pointTo(true),
							Tag:       pointToStrings("foo"),
						},
					},
				},
			},
			flags: flagsTestcase{
				inSubcommand: "ls",
				expFlags: []config.Flag{
					{Key: "repo", Val: "test"},
					{Key: "host", Val: "other_host"},
					{Key: "json", Val: "true"},
					{Key: "long", Val: "true"},
					{Key: "option", Val: "local.layout=test"},
					{Key: "option", Val: "s3.connections=3"},
					{Key: "pack-size", Val: "4"},
					{Key: "path", Val: "/echo"},
					{Key: "path", Val: "/foxtrot"},
					{Key: "recursive", Val: "true"},
					{Key: "tag", Val: "foo"},
					{Key: "tls-client-cert", Val: "client.crt"},
				},
			},
		})
	})

	t.Run("Restic.Snapshots", func(t *testing.T) {
		runTest(t, testCase{
			name: "use Datastore and Destination values",
			inputFileContents: `
[defaults]
restic.snapshots = { latest = 1, tag = ['alfa', 'bravo'] }

[datastores.stuff.defaults]
restic.snapshots = { compact = true, group-by = ['foo'], latest = 2 }

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults]
# group-by is length 0, which signals "do not merge this field with parent values".
restic.snapshots = { group-by = [], latest = 3, path = ['c'] }
`,
			merge: mergeTestcase{
				expDefaults: config.Defaults{
					PasswordConfig: &config.PasswordConfig{},
					Restic: &config.ResticDefaults{
						Snapshots: &config.ResticSnapshots{
							Compact: pointTo(true),
							GroupBy: pointToStrings(),
							Latest:  pointTo(3),
							Path:    pointToStrings("c"),
							Tag:     pointToStrings("alfa", "bravo"),
						},
					},
				},
			},
			flags: flagsTestcase{
				inSubcommand: "snapshots",
				expFlags: []config.Flag{
					{Key: "repo", Val: "test"},
					{Key: "compact", Val: "true"},
					{Key: "latest", Val: "3"},
					{Key: "path", Val: "c"},
					{Key: "tag", Val: "alfa"},
					{Key: "tag", Val: "bravo"},
				},
			},
		})
	})

	t.Run("Restic.Stats", func(t *testing.T) {
		runTest(t, testCase{
			inputFileContents: `
[defaults]
restic.global = { json = true }
restic.stats = { host = ['some_host'], tag = ['foo'] }

[datastores.stuff.defaults]
restic.stats = { mode = 'files-by-contents', path = ['/charlie', '/delta'] }

[datastores.stuff.destinations.foo]
path = 'test'

[datastores.stuff.destinations.foo.defaults]
restic.stats = { host = [], path = ['/echo', '/foxtrot'] }
`,
			merge: mergeTestcase{
				expDefaults: config.Defaults{
					PasswordConfig: &config.PasswordConfig{},
					Restic: &config.ResticDefaults{
						Global: &config.ResticGlobal{JSON: pointTo(true)},
						Stats: &config.ResticStats{
							Mode: pointTo("files-by-contents"),
							Host: pointToStrings(),
							Path: pointToStrings("/echo", "/foxtrot"),
							Tag:  pointToStrings("foo"),
						},
					},
				},
			},
			flags: flagsTestcase{
				inSubcommand: "stats",
				expFlags: []config.Flag{
					{Key: "repo", Val: "test"},
					{Key: "json", Val: "true"},
					{Key: "mode", Val: "files-by-contents"},
					{Key: "path", Val: "/echo"},
					{Key: "path", Val: "/foxtrot"},
					{Key: "tag", Val: "foo"},
				},
			},
		})
	})
}
