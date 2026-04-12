package config

import "github.com/alecthomas/kong"

type CLI struct {
	RepoPath  string           `short:"r" name:"repo" help:"Path of the restic repository" env:"RESTIC_REPOSITORY" required:""`
	MountPath string           `short:"m" name:"mount" help:"Path of the restic mount point" env:"RESTIC_MOUNTPOINT" required:""`
	Version   kong.VersionFlag `short:"v" name:"version" help:"Show app version"`
}
