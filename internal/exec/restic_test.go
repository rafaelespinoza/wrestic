package exec_test

import (
	"context"
	"os"
	"testing"

	"github.com/rafaelespinoza/wrestic/internal/config"
	"github.com/rafaelespinoza/wrestic/internal/exec"
)

func TestResticBatch(t *testing.T) {
	tests := []struct {
		name                 string
		datastores           []config.Datastore
		configDir            string
		subcommand           string
		run                  bool
		expectedSinkData     []string
		expectedReceivedArgs [][]string
	}{
		{
			name: "Run=false",
			datastores: []config.Datastore{
				{
					Destinations: map[string]config.Destination{
						"foo": {
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secrets/foo")},
							},
						},
					},
				},
				{
					Destinations: map[string]config.Destination{
						"bar": {
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secrets/bar")},
							},
						},
					},
				},
			},
			configDir:  "/tmp",
			subcommand: "test",
			expectedSinkData: []string{
				`# test --foo=123 deadbeef --bar --repo=foo --password-file='/tmp/secrets/foo'
`,
				`# test --foo=123 deadbeef --bar --repo=bar --password-file='/tmp/secrets/bar'
`,
			},
			expectedReceivedArgs: [][]string{},
		},
		{
			name: "Run=true",
			datastores: []config.Datastore{
				{
					Destinations: map[string]config.Destination{
						"foo": {
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secrets/foo")},
							},
						},
					},
				},
				{
					Destinations: map[string]config.Destination{
						"bar": {
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secrets/bar")},
							},
						},
					},
				},
			},
			configDir:  "/tmp",
			subcommand: "test",
			run:        true,
			expectedSinkData: []string{
				`# test --foo=123 deadbeef --bar --repo=foo --password-file='/tmp/secrets/foo'
`,
				`# test --foo=123 deadbeef --bar --repo=bar --password-file='/tmp/secrets/bar'
`,
			},
			expectedReceivedArgs: [][]string{
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=foo", `--password-file=/tmp/secrets/foo`},
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=bar", `--password-file=/tmp/secrets/bar`},
			},
		},
		{
			name: "PasswordConfig.File has spaces",
			datastores: []config.Datastore{
				{
					Destinations: map[string]config.Destination{
						"foo": {
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secrets/foo")},
							},
						},
					},
				},
				{
					Destinations: map[string]config.Destination{
						"bar": {
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secret bar")},
							},
						},
					},
				},
			},
			configDir:  "/Users/username/Library/Application Support/wrestic",
			subcommand: "test",
			run:        true,
			expectedSinkData: []string{
				`# test --foo=123 deadbeef --bar --repo=foo --password-file='"/Users/username/Library/Application Support/wrestic/secrets/foo"'
`,
				`# test --foo=123 deadbeef --bar --repo=bar --password-file='"/Users/username/Library/Application Support/wrestic/secret bar"'
`,
			},
			expectedReceivedArgs: [][]string{
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=foo", `--password-file="/Users/username/Library/Application Support/wrestic/secrets/foo"`},
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=bar", `--password-file="/Users/username/Library/Application Support/wrestic/secret bar"`},
			},
		},
		{
			name: "Subcommand=backup",
			datastores: []config.Datastore{
				{
					Sources: []config.Source{{Path: "/usr/foo"}},
					Destinations: map[string]config.Destination{
						"foo": {
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secrets/foo")},
							},
						},
					},
				},
				{
					Sources: []config.Source{{Path: "/etc/bar"}},
					Destinations: map[string]config.Destination{
						"bar": {
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{File: pointToFilename("secrets/bar")},
							},
						},
					},
				},
			},
			subcommand: "backup",
			expectedSinkData: []string{
				`# backup --foo=123 deadbeef --bar --repo=foo --password-file='secrets/foo' /usr/foo
`,
				`# backup --foo=123 deadbeef --bar --repo=bar --password-file='secrets/bar' /etc/bar
`,
			},
			expectedReceivedArgs: [][]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sink := Sink{data: make([]string, 0)}
			receivedArgs := make([][]string, 0)
			newCommand := func() exec.Command {
				run := func(ctx context.Context, args ...string) error {
					receivedArgs = append(receivedArgs, args)
					return nil
				}
				return &Command{RunResp: run}
			}

			batch := exec.ResticBatch{
				Sink:       &sink,
				ConfigDir:  test.configDir,
				Subcommand: test.subcommand,
				Args:       []string{"--foo=123", "deadbeef", "--bar"},
				Run:        test.run,
				NewCommand: newCommand,
			}

			err := batch.Do(context.Background(), test.datastores)
			if err != nil {
				t.Fatal(err)
			}

			if len(sink.data) != len(test.expectedSinkData) {
				t.Fatalf("wrong number of sink items; got %d, expected %d", len(sink.data), len(test.expectedSinkData))
			}
			for i, got := range sink.data {
				exp := test.expectedSinkData[i]

				if got != exp {
					t.Errorf("item %d; wrong sink data\ngot %q\nexp %q", i, got, exp)
				}
			}

			if len(receivedArgs) != len(test.expectedReceivedArgs) {
				t.Fatalf("wrong number of received args; got %d, expected %d", len(receivedArgs), len(test.expectedReceivedArgs))
			}
			for i, gotArgs := range receivedArgs {
				expArgs := test.expectedReceivedArgs[i]

				if len(gotArgs) != len(expArgs) {
					t.Errorf("item %d; wrong number of received args; got %d, expected %d", i, len(gotArgs), len(expArgs))
					continue
				}

				for j, got := range gotArgs {
					exp := expArgs[j]
					if got != exp {
						t.Errorf("item [%d][%d]; wrong received arg\ngot %q\nexp %q", i, j, got, exp)
					}
				}
			}
		})
	}
}

func TestResticBatchNewRestic(t *testing.T) {
	// Sanity check on invoking a command upon each destination.
	tests := []struct {
		name      string
		resticBin string // resticBin should be the name of a simple program available on most Unix-like systems.
		expError  bool
	}{
		{name: "ok", resticBin: "true", expError: false},
		{name: "err", resticBin: "false", expError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("RESTIC_BIN", test.resticBin)

			batch := exec.ResticBatch{
				Run:        true,
				NewCommand: func() exec.Command { return exec.NewRestic(os.Stdout, os.Stderr) },
			}

			destinations := map[string]config.Destination{
				"bar": {
					Path: "bar",
					Defaults: config.Defaults{
						PasswordConfig: &config.PasswordConfig{File: pointToFilename("bar")},
					},
				},
			}
			err := batch.Do(context.Background(), []config.Datastore{{Destinations: destinations}})

			if !test.expError && err != nil {
				t.Fatal(err)
			} else if test.expError && err == nil {
				t.Fatal("expected an error, but got empty")
			}
		})
	}
}

type Sink struct{ data []string }

func (s *Sink) Write(p []byte) (n int, err error) {
	s.data = append(s.data, string(p))
	return len(p), nil
}

type Command struct {
	RunResp func(ctx context.Context, args ...string) error
}

func (c *Command) Run(ctx context.Context, args ...string) error {
	if c.RunResp == nil {
		panic("define RunResp")
	}

	return c.RunResp(ctx, args...)
}

func pointToFilename(in string) (out *string) { return &in }
