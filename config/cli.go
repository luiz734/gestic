package config

type CLI struct {
	RepoPath string `short:"r" name:"repo" help:"Path of restic repository" required:""`
}
