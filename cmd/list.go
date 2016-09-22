package cmd

import (
	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List clusters",
		Long:    "List clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters, err := cxt.Client.ListClusters(cxt.Account)
			if err != nil {
				return err
			}

			console.WriteClusters(clusters)

			return nil
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newListCommand())
}
