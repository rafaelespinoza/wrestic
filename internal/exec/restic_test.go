package exec_test

import (
	"context"
	"os"
	"strings"
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
		expectErr            bool
		expectErrMsgContains string // if we're expecting an error, what should the message mention?
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
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString("cat {{ filename (index . 0) }}"),
									Args:     []string{"secrets/foo"},
								},
							},
						},
					},
				},
				{
					Destinations: map[string]config.Destination{
						"bar": {
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString("cat {{ filename (index . 0) }}"),
									Args:     []string{"secrets/bar"},
								},
							},
						},
					},
				},
			},
			configDir:  "/tmp",
			subcommand: "test",
			expectedSinkData: []string{
				`# test --foo=123 deadbeef --bar --repo=foo --password-command='cat /tmp/secrets/foo'
`,
				`# test --foo=123 deadbeef --bar --repo=bar --password-command='cat /tmp/secrets/bar'
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
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString("cat {{ filename (index . 0) }}"),
									Args:     []string{"secrets/foo"},
								},
							},
						},
					},
				},
				{
					Destinations: map[string]config.Destination{
						"bar": {
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString("cat {{ filename (index . 0) }}"),
									Args:     []string{"secrets/bar"},
								},
							},
						},
					},
				},
			},
			configDir:  "/tmp",
			subcommand: "test",
			run:        true,
			expectedSinkData: []string{
				`# test --foo=123 deadbeef --bar --repo=foo --password-command='cat /tmp/secrets/foo'
`,
				`# test --foo=123 deadbeef --bar --repo=bar --password-command='cat /tmp/secrets/bar'
`,
			},
			expectedReceivedArgs: [][]string{
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=foo", `--password-command=cat /tmp/secrets/foo`},
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=bar", `--password-command=cat /tmp/secrets/bar`},
			},
		},
		{
			name: "PasswordConfig.Args absolute paths",
			datastores: []config.Datastore{
				{
					Destinations: map[string]config.Destination{
						"foo": { // Args are filenames without spaces
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"/elsewhere/no_spaces/secrets/id", "/elsewhere/no_spaces/secrets/foo"},
								},
							},
						},
					},
				},
				{
					Destinations: map[string]config.Destination{
						"bar": { // Args are filenames with spaces
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"/elsewhere/has spaces/secrets/id", "/elsewhere/has spaces/secrets/bar"},
								},
							},
						},
					},
				},
			},
			configDir:  "/tmp/config_place",
			subcommand: "test",
			run:        true,
			expectedSinkData: []string{
				`# test --foo=123 deadbeef --bar --repo=foo --password-command='age -d -i /elsewhere/no_spaces/secrets/id /elsewhere/no_spaces/secrets/foo'
`,
				`# test --foo=123 deadbeef --bar --repo=bar --password-command='age -d -i "/elsewhere/has spaces/secrets/id" "/elsewhere/has spaces/secrets/bar"'
`,
			},
			expectedReceivedArgs: [][]string{
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=foo", `--password-command=age -d -i /elsewhere/no_spaces/secrets/id /elsewhere/no_spaces/secrets/foo`},
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=bar", `--password-command=age -d -i "/elsewhere/has spaces/secrets/id" "/elsewhere/has spaces/secrets/bar"`},
			},
		},
		{
			name: "PasswordConfig.Args relative paths",
			datastores: []config.Datastore{
				{
					Destinations: map[string]config.Destination{
						"foo": { // Args are filenames without spaces
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"secrets/id", "secrets/foo"},
								},
							},
						},
					},
				},
				{
					Destinations: map[string]config.Destination{
						"bar": { // Args are filenames with spaces
							Path: "bar",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"secret id", "secret bar"},
								},
							},
						},
					},
				},
			},
			configDir:  "/tmp/config_place",
			subcommand: "test",
			run:        true,
			expectedSinkData: []string{
				`# test --foo=123 deadbeef --bar --repo=foo --password-command='age -d -i /tmp/config_place/secrets/id /tmp/config_place/secrets/foo'
`,
				`# test --foo=123 deadbeef --bar --repo=bar --password-command='age -d -i "/tmp/config_place/secret id" "/tmp/config_place/secret bar"'
`,
			},
			expectedReceivedArgs: [][]string{
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=foo", `--password-command=age -d -i /tmp/config_place/secrets/id /tmp/config_place/secrets/foo`},
				{"test", "--foo=123", "deadbeef", "--bar", "--repo=bar", `--password-command=age -d -i "/tmp/config_place/secret id" "/tmp/config_place/secret bar"`},
			},
		},
		{
			name: "PasswordConfig.Template invalid",
			datastores: []config.Datastore{
				{
					Destinations: map[string]config.Destination{
						"foo": {
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (does_not_work . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"secrets/id", "secrets/foo"},
								},
							},
						},
					},
				},
			},
			configDir:            "/tmp/config_place",
			subcommand:           "test",
			run:                  true,
			expectErr:            true,
			expectErrMsgContains: "template is invalid",
			expectedSinkData:     []string{},
			expectedReceivedArgs: [][]string{},
		},
		{
			name: "PasswordConfig.Args invalid",
			datastores: []config.Datastore{
				{
					Destinations: map[string]config.Destination{
						"foo": {
							Path: "foo",
							Defaults: config.Defaults{
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"secrets/id"},
								},
							},
						},
					},
				},
			},
			configDir:            "/tmp/config_place",
			subcommand:           "test",
			run:                  true,
			expectErr:            true,
			expectErrMsgContains: "template does not agree with args",
			expectedSinkData:     []string{},
			expectedReceivedArgs: [][]string{},
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
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"secrets/id", "secrets/foo"},
								},
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
								PasswordConfig: &config.PasswordConfig{
									Template: pointToString(`age -d -i {{ filename (index . 0) }} {{ filename (index . 1) }}`),
									Args:     []string{"secrets/id", "secrets/bar"},
								},
							},
						},
					},
				},
			},
			configDir:  "/tmp/config_place",
			subcommand: "backup",
			expectedSinkData: []string{
				`# backup --foo=123 deadbeef --bar --repo=foo --password-command='age -d -i /tmp/config_place/secrets/id /tmp/config_place/secrets/foo' /usr/foo
`,
				`# backup --foo=123 deadbeef --bar --repo=bar --password-command='age -d -i /tmp/config_place/secrets/id /tmp/config_place/secrets/bar' /etc/bar
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
			if err != nil && !test.expectErr {
				t.Fatal(err)
			} else if err == nil && test.expectErr {
				t.Error("expected an error")
			} else if err != nil && test.expectErr && !strings.Contains(err.Error(), test.expectErrMsgContains) {
				t.Errorf("expected error message %q to contain %q", err, test.expectErrMsgContains)
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
						PasswordConfig: &config.PasswordConfig{Template: pointToString("foo"), Args: []string{"bar"}},
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

func pointToString(in string) (out *string) { return &in }
