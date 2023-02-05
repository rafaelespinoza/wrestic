package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type Flag struct{ Key, Val string }

const (
	// pwcmdConfigFileKey is the name of a key from the configuration file
	// that's used in producing a password command.
	pwcmdConfigFileKey = "password-config"
)

func parsePasswordCommand(configDir string, pw *PasswordConfig) (out string, err error) {
	if pw == nil || pw.Template == nil {
		return
	}

	var tmpl *template.Template
	fns := template.FuncMap{
		"filename":    func(argFilename string) (out string) { return formatFilenameFlag(configDir, argFilename) },
		"filenameArg": func(argIndex int) (out string) { return formatFilenameFlag(configDir, pw.Args[argIndex]) },
	}

	tmpl, err = template.New(pwcmdConfigFileKey).Funcs(fns).Parse(*pw.Template)
	if err != nil {
		err = fmt.Errorf("%w: %s.template is invalid", err, pwcmdConfigFileKey)
		return
	}

	var bld strings.Builder
	if err = tmpl.Execute(&bld, pw.Args); err != nil {
		if xerr, ok := err.(template.ExecError); ok {
			err = fmt.Errorf("%w: %s.template does not agree with args", xerr, pwcmdConfigFileKey)
		}
		return
	}

	out = bld.String()
	return
}

func formatFilenameFlag(configDir, filename string) string {
	filename = filepath.Clean(filename)

	if strings.Contains(filename, "$") {
		filename = os.ExpandEnv(filename)
	}

	if !filepath.IsAbs(filename) {
		filename = filepath.Join(configDir, filename)
	}

	if strings.Contains(filename, " ") {
		filename = fmt.Sprintf("%q", filename)
	}

	return filename
}

// resticConfig represents a set of command flag values for restic.
type resticConfig interface {
	ResticGlobal | ResticBackup | ResticCheck | ResticLS | ResticSnapshots | ResticStats
}

// makeMergedFlags should be called after a Destination already merged its own
// configuration defaults via its Merge method.
func makeMergedFlags[C resticConfig](cmdConf *C, globalConf *ResticGlobal) (out []Flag, err error) {
	cmdValues := make(map[string]any)
	globalValues := make(map[string]any)

	if cmdConf != nil {
		cmdValues = mapResticConfig(*cmdConf)
	}
	if globalConf != nil {
		globalValues = mapResticConfig(*globalConf)
	}

	if err = mergeResticConfigMap(cmdValues, globalValues); err != nil {
		return
	}

	out, err = makeResticConfigFlags(cmdValues)
	return
}

func mapResticConfig[C resticConfig](r C) map[string]any {
	out := make(map[string]any)

	inputValue := reflect.ValueOf(r)
	inputType := inputValue.Type()

	for i := 0; i < inputValue.NumField(); i++ {
		fieldValue := inputValue.Field(i)
		if !fieldValue.CanInterface() {
			continue
		}

		structField := inputType.Field(i)
		if !structField.IsExported() {
			continue
		}

		if fieldValue.IsNil() {
			continue
		}

		tomlTag := structField.Tag.Get("toml")

		// The main part of the struct field tag (the non-optional one) is meant
		// to be the same as the long option in the restic CLI. But the struct
		// tag might have an option, such as "omitempty". This may matter when
		// generating CLI flag values from the configured file. So, remove any
		// option things from the struct tag here.
		tomlKey, _, _ := strings.Cut(tomlTag, ",")

		if fieldValue.Kind() == reflect.Pointer {
			out[tomlKey] = fieldValue.Elem().Interface()
		} else {
			out[tomlKey] = fieldValue.Interface()
		}
	}

	return out
}

func mergeResticConfigMap(destination, source map[string]any) error {
	for key, srcValue := range source {
		_, ok := destination[key]
		if ok {
			// The destination already has something. Leave it there.
			continue
		}

		switch srcVal := srcValue.(type) {
		case bool, int, string, uint:
			destination[key] = srcVal
		case []string: // flags that may be specified multiple times will be of this type.
			destination[key] = duplicateStrings(srcVal)
		case []map[string]string: // for restic v0.14.0, it's only ResticGlobal.Option that has this type.
			destination[key] = duplicateOptionMap(srcVal)
		default:
			return fmt.Errorf("unhandled type %T at key %q", srcVal, key)
		}
	}

	return nil
}

func makeResticConfigFlags(in map[string]any) (out []Flag, err error) {
	keys := []string{}
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		anything := in[key]

		switch val := anything.(type) {
		case bool:
			out = append(out, Flag{key, strconv.FormatBool(val)})
		case int:
			out = append(out, Flag{key, strconv.Itoa(val)})
		case uint:
			out = append(out, Flag{key, strconv.Itoa(int(val))})
		case string:
			out = append(out, Flag{key, val})
		case []string: // flags that may be specified multiple times will be of this type.
			for _, v := range val {
				out = append(out, Flag{key, v})
			}
		case []map[string]string: // for restic v0.14.0, it's only ResticGlobal.Option that has this type.
			for _, option := range val {
				for subkey, subval := range option {
					out = append(out, Flag{key, subkey + "=" + subval})
				}
			}
		default:
			err = fmt.Errorf("unhandled type %T at key %q", val, key)
			return
		}
	}

	return
}

func duplicateStrings(in []string) (out []string) {
	out = make([]string, len(in))
	copy(out, in)
	return
}

func duplicateOptionMap(in []map[string]string) (out []map[string]string) {
	out = make([]map[string]string, len(in))

	for i, sources := range in {
		dest := make(map[string]string)

		for subkey, subval := range sources {
			dest[subkey] = subval
		}

		out[i] = dest
	}

	return
}
