package config

// ResticDefaults are any default configuration values for restic subcommands.
// Asides from Global, which is configuration for shared flags, the struct
// fields here correspond to flags for a restic subcommand.
type ResticDefaults struct {
	// Global refers to any restic flags that are made available for any restic
	// subcommand. In restic's usage menus, they may appear as "global flags".
	Global    *ResticGlobal    `toml:"global"`
	Backup    *ResticBackup    `toml:"backup"`
	Check     *ResticCheck     `toml:"check"`
	LS        *ResticLS        `toml:"ls"`
	Snapshots *ResticSnapshots `toml:"snapshots"`
	Stats     *ResticStats     `toml:"stats"`
}

func duplicateResticDefaults(in *ResticDefaults) (out *ResticDefaults) {
	out = &ResticDefaults{}
	if in == nil {
		return
	}

	mergeConfig(out, in)

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

type ResticCheck struct {
	ReadData       *bool   `toml:"read-data"`
	ReadDataSubset *string `toml:"read-data-subset"`
	WithCache      *bool   `toml:"with-cache"`
}

func (r *ResticCheck) makeFlags(g *ResticGlobal) (out []Flag, err error) {
	out, err = makeMergedFlags(r, g)
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
