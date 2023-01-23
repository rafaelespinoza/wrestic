package config_test

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	"github.com/rafaelespinoza/wrestic/internal/config"
)

func testResticDefaults(t *testing.T, errPrefix string, actual, expected *config.ResticDefaults) {
	t.Helper()

	if actual == nil && expected == nil {
		return // test OK
	} else if actual != nil && expected == nil {
		t.Fatalf("%s actual %#v, expected %v", errPrefix, *actual, expected)
	} else if actual == nil && expected != nil {
		t.Fatalf("%s actual %v, expected %#v", errPrefix, actual, *expected)
	}

	testResticConfig(t, errPrefix+".Global", actual.Global, expected.Global)
	testResticConfig(t, errPrefix+".Backup", actual.Backup, expected.Backup)
	testResticConfig(t, errPrefix+".Check", actual.Check, expected.Check)
	testResticConfig(t, errPrefix+".LS", actual.LS, expected.LS)
	testResticConfig(t, errPrefix+".Snapshots", actual.Snapshots, expected.Snapshots)
	testResticConfig(t, errPrefix+".Stats", actual.Stats, expected.Stats)
}

type resticConfig interface {
	config.ResticGlobal | config.ResticBackup | config.ResticCheck | config.ResticLS | config.ResticSnapshots | config.ResticStats
}

func testResticConfig[C resticConfig](t *testing.T, errPrefix string, actual, expected *C) {
	t.Helper()

	if actual == nil && expected == nil {
		return // test OK
	} else if actual != nil && expected == nil {
		t.Fatalf("%s actual %#v, expected %v", errPrefix, *actual, expected)
	} else if actual == nil && expected != nil {
		t.Fatalf("%s actual %v, expected %#v", errPrefix, actual, *expected)
	}

	// unexportedField looks for fields defined directly on the input struct
	// type and, based on the field name, says if it's exported or not.
	unexportedField := func(p cmp.Path) bool {
		structField, ok := p.Index(-1).(cmp.StructField)
		if !ok {
			return false
		}

		r, _ := utf8.DecodeRuneInString(structField.Name())
		return !unicode.IsUpper(r)
	}

	ignoreUnexportedFields := cmp.FilterPath(unexportedField, cmp.Ignore())

	if diff := cmp.Diff(actual, expected, ignoreUnexportedFields); diff != "" {
		t.Errorf("%s (- means something in actual) (+ means something in expected)\n%s", errPrefix, diff)
	}
}
