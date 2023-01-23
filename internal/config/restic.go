package config

import (
	"fmt"
	"os"
	"reflect"

	"github.com/imdario/mergo"
)

type ResticDefaults struct {
	Global    *ResticGlobal    `toml:"global"`
	Backup    *ResticBackup    `toml:"backup"`
	Check     *ResticCheck     `toml:"check"`
	LS        *ResticLS        `toml:"ls"`
	Snapshots *ResticSnapshots `toml:"snapshots"`
	Stats     *ResticStats     `toml:"stats"`
}

func mergeResticDefaults(dst, src *ResticDefaults) (err error) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	if dst == nil && src != nil {
		err = mergo.Merge(dst, src)
		return
	}

	if err = mergeResticGlobal(dst, src.Global); err != nil {
		return
	}

	if err = mergeResticBackup(dst, src.Backup); err != nil {
		return
	}

	if err = mergeResticCheck(dst, src.Check); err != nil {
		return
	}

	if err = mergeResticLS(dst, src.LS); err != nil {
		return
	}

	if err = mergeResticSnapshots(dst, src.Snapshots); err != nil {
		return
	}

	if err = mergeResticStats(dst, src.Stats); err != nil {
		return
	}

	return
}

func duplicateResticDefaults(in *ResticDefaults) (out *ResticDefaults) {
	out = &ResticDefaults{}
	if in == nil {
		return
	}

	// The library, github.com/imdario/mergo, may return an error if:
	// - the type of the 1st input is not a pointer to a struct.
	// - the types of both inputs are not same type structs.
	// Neither should be concerns for this tool.
	//
	// To keep the signatures of each duplication function in this package
	// consistent, don't return the error. But make some noise.
	if err := mergo.Merge(out, in); err != nil {
		fmt.Fprintf(os.Stderr, "wrestic: %#v\n", err)
	}

	return
}

type ResticGlobal struct {
	CACert          *string              `toml:"cacert"`
	CacheDir        *string              `toml:"cache-dir"`
	CleanupCache    *bool                `toml:"cleanup-cache"`
	Compression     *string              `toml:"compression"`
	InsecureTLS     *bool                `toml:"insecure-tls"`
	JSON            *bool                `toml:"json"`
	KeyHint         *string              `toml:"key-hint"`
	LimitDownload   *int                 `toml:"limit-download"`
	LimitUpload     *int                 `toml:"limit-upload"`
	NoCache         *bool                `toml:"no-cache"`
	NoLock          *bool                `toml:"no-lock"`
	Option          *[]map[string]string `toml:"option"`
	PackSize        *uint                `toml:"pack-size"`
	PasswordCommand *string              `toml:"password-command"`
	PasswordFile    *string              `toml:"password-file"`
	Quiet           *bool                `toml:"quiet"`
	Repo            *string              `toml:"repo"`
	RepositoryFile  *string              `toml:"repository-file"`
	TLSClientCert   *string              `toml:"tls-client-cert"`
	Verbose         *int                 `toml:"verbose"`
}

func mergeResticGlobal(dst *ResticDefaults, src *ResticGlobal) (err error) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	var target ResticGlobal
	if dst.Global != nil {
		target = *dst.Global
	}

	if err = mergeResticConfig(&target, src); err != nil {
		return
	}
	dst.Global = &target
	return
}

type ResticBackup struct {
	DryRun            *bool     `toml:"dry-run"`
	Exclude           *[]string `toml:"exclude"`
	ExcludeCaches     *bool     `toml:"exclude-caches"`
	ExcludeFile       *[]string `toml:"exclude-file"`
	ExcludeIfPresent  *[]string `toml:"exclude-if-present"`
	ExcludeLargerThan *string   `toml:"exclude-larger-than"`
	FilesFrom         *[]string `toml:"files-from"`
	FilesFromRaw      *[]string `toml:"files-from-raw"`
	FilesFromVerbatim *[]string `toml:"files-from-verbatim"`
	Force             *bool     `toml:"force"`
	Host              *string   `toml:"host"`
	Iexclude          *[]string `toml:"iexclude"`
	IexcludeFile      *[]string `toml:"iexclude-file"`
	IgnoreCtime       *bool     `toml:"ignore-ctime"`
	IgnoreInode       *bool     `toml:"ignore-inode"`
	OneFileSystem     *bool     `toml:"one-file-system"`
	Parent            *string   `toml:"parent"`
	Stdin             *bool     `toml:"stdin"`
	StdinFilename     *string   `toml:"stdin-filename"`
	Tag               *[]string `toml:"tag"`
	Time              *string   `toml:"time"` // is type string because "now" is accepted by restic.
	WithAtime         *bool     `toml:"with-atime"`
}

func (r *ResticBackup) makeFlags(g *ResticGlobal) (out []Flag, err error) {
	out, err = makeMergedFlags(r, g)
	return
}

func mergeResticBackup(dst *ResticDefaults, src *ResticBackup) (err error) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	var target ResticBackup
	if dst.Backup != nil {
		target = *dst.Backup
	}

	if err = mergeResticConfig(&target, src); err != nil {
		return
	}
	dst.Backup = &target

	return
}

type ResticCheck struct {
	ReadData       *bool   `toml:"read-data"`
	ReadDataSubset *string `toml:"read-data-subset"`
	WithCache      *bool   `toml:"with-cache"`
}

func (r *ResticCheck) makeFlags(g *ResticGlobal) (out []Flag, err error) {
	out, err = makeMergedFlags(r, g)
	return
}

func mergeResticCheck(dst *ResticDefaults, src *ResticCheck) (err error) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	var target ResticCheck
	if dst.Check != nil {
		target = *dst.Check
	}

	if err = mergeResticConfig(&target, src); err != nil {
		return
	}
	dst.Check = &target

	return
}

type ResticLS struct {
	Host      *[]string `toml:"host"`
	Long      *bool     `toml:"long"`
	Path      *[]string `toml:"path"`
	Recursive *bool     `toml:"recursive"`
	Tag       *[]string `toml:"tag"`
}

func (r *ResticLS) makeFlags(g *ResticGlobal) (out []Flag, err error) {
	out, err = makeMergedFlags(r, g)
	return
}

func mergeResticLS(dst *ResticDefaults, src *ResticLS) (err error) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	var target ResticLS
	if dst.LS != nil {
		target = *dst.LS
	}

	if err = mergeResticConfig(&target, src); err != nil {
		return
	}
	dst.LS = &target

	return
}

type ResticSnapshots struct {
	Compact *bool     `toml:"compact"`
	GroupBy *[]string `toml:"group-by"`
	Host    *[]string `toml:"host"`
	Latest  *int      `toml:"latest"`
	Path    *[]string `toml:"path"`
	Tag     *[]string `toml:"tag"`
}

func (r *ResticSnapshots) makeFlags(g *ResticGlobal) (out []Flag, err error) {
	out, err = makeMergedFlags(r, g)
	return
}

func mergeResticSnapshots(dst *ResticDefaults, src *ResticSnapshots) (err error) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	var target ResticSnapshots
	if dst.Snapshots != nil {
		target = *dst.Snapshots
	}

	if err = mergeResticConfig(&target, src); err != nil {
		return
	}
	dst.Snapshots = &target

	return
}

type ResticStats struct {
	Host *[]string `toml:"host"`
	Mode *string   `toml:"mode"`
	Path *[]string `toml:"path"`
	Tag  *[]string `toml:"tag"`
}

func (r *ResticStats) makeFlags(g *ResticGlobal) (out []Flag, err error) {
	out, err = makeMergedFlags(r, g)
	return
}

func mergeResticStats(dst *ResticDefaults, src *ResticStats) (err error) {
	if (dst == nil && src == nil) || (dst != nil && src == nil) {
		return
	}

	var target ResticStats
	if dst.Stats != nil {
		target = *dst.Stats
	}

	if err = mergeResticConfig(&target, src); err != nil {
		return
	}
	dst.Stats = &target

	return
}

type mergeTransformer struct{ okType reflect.Type }

func (t mergeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ != t.okType {
		return nil
	}

	return func(dst, src reflect.Value) error {
		// Only merge the values if the configuration file does not specify the
		// field. If dst is an initialized but zero-length slice, do nothing.
		if dst.CanSet() && dst.IsNil() {
			dst.Set(src)
			return nil
		}

		return nil
	}
}

// resticConfig represents a set of command flag values for restic.
type resticConfig interface {
	ResticGlobal | ResticBackup | ResticCheck | ResticLS | ResticSnapshots | ResticStats
}

func mergeResticConfig[C resticConfig](dst, src *C) (err error) {
	transformer := mergeTransformer{okType: reflect.TypeOf(new([]string))}
	err = mergo.Merge(dst, src, mergo.WithTransformers(transformer))
	return
}
