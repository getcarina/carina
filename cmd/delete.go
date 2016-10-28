package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	var options struct {
		name string
		wait bool
	}

	var cmd = &cobra.Command{
		Use:     "delete <cluster-name>",
		Aliases: []string{"rm"},
		Short:   "Delete a cluster",
		Long:    "Delete a cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return bindClusterNameArg(args, &options.name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cxt.Client.DeleteCluster(cxt.Account, options.name, options.wait)
			if err != nil {
				return err
			}

			fmt.Printf("Deleting cluster (%s)\n", options.name)

			return nil
		},
	}

	cmd.ValidArgs = []string{"cluster-name"}
	cmd.Flags().BoolVar(&options.wait, "wait", false, "Wait for the cluster to be deleted")
	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
