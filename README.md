# wrestic

This is a thin wrapper around [restic](https://github.com/restic/restic).

It's meant to:
- keeps track of what to backup (sourcepaths) and paths to backup repositories
  (destinations) so they're easier to associate to one another.
- generate command line arguments and flags to pass to restic.
