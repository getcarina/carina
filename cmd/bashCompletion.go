package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newBashCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "bash-completion",
		Short:  "Generate a bash completion file for the carina cli",
		Long:   "Generate a bash completion file for the carina cli",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Parent().GenBashCompletion(os.Stdout)
		},
	}

	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
