package cmd

import (
	"github.com/getcarina/carina/console"
	"github.com/getcarina/carina/version"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Show the application version",
		Long:  "Show the application version",
		Run: func(cmd *cobra.Command, args []string) {
			writeVersion()
		},
	}

	return cmd
}

func writeVersion() {
	console.Write("%s (%s)", version.Version, version.Commit)
}
