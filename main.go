package main

import (
	"fmt"
	"gestic/config"
	"gestic/models/selector"
	"gestic/restic"
	"os"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {

	var cli config.CLI
	ctx := kong.Parse(&cli,
		kong.Name("gestic"),
		kong.Description("A diff tool for restic snapshots."),
		kong.Vars{
			"version": fmt.Sprintf("%s (%s)", version, commit),
		},
	)

	// It returns error if no subcommand was provided
	// We don't use subcommands and are ignoring errors
	err := ctx.Run()
	//if err != nil {
	//	ctx.FatalIfErrorf(err)
	//}

	snapshots, err := restic.GetSnapshots(cli.RepoPath, cli.MountPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: cannot get snapshots: %v\n\n", err)
		_, _ = fmt.Fprintf(os.Stderr, "Did you mount the repository?\n")
		_, _ = fmt.Fprintf(os.Stderr, "Run 'man restic mount' for more information.")
		os.Exit(1)
	}

	p := tea.NewProgram(
		selector.InitialModel(snapshots),
	)

	// Redirects the debug to a local file
	debugFile := "/dev/null"
	if len(os.Getenv("DEBUG")) > 0 {
		debugFile = "debug.log"
	}
	f, err := tea.LogToFile(debugFile, "debug")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: cannot setup debug file:: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: program failed to run:: %v\n", err)
		os.Exit(1)
	}
}
