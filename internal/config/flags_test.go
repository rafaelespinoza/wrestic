package config_test

import (
	"testing"

	"github.com/rafaelespinoza/wrestic/internal/config"
)

func testFlags(t *testing.T, errPrefix string, actual, expected []config.Flag) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("%s wrong number of Flags; got %d, expected %d", errPrefix, len(actual), len(expected))
	}

	for i, got := range actual {
		exp := expected[i]

		if got.Key != exp.Key {
			t.Errorf("%s[%d] wrong Key got %q, expected %q", errPrefix, i, got.Key, exp.Key)
		}

		if got.Val != exp.Val {
			t.Errorf("%s[%d] wrong Val got %q, expected %q", errPrefix, i, got.Val, exp.Val)
		}
	}
}
