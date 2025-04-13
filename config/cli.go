package config

type CLI struct {
	RepoPath  string `short:"r" name:"repo" help:"Path of restic repository" required:""`
	MountPath string `short:"m" name:"mount" help:"Path of the mount point" required:""`
}
