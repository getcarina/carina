package cmd

import (
	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newRebuildCommand() *cobra.Command {
	var options struct {
		name string
		wait bool
	}

	var cmd = &cobra.Command{
		Use:   "rebuild <cluster-name>",
		Short: "Rebuild a cluster",
		Long:  "Rebuild a cluster. This rebuilds the cluster infrastructure only and does not affect existing containers or volumes.",
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return bindName(args, &options.name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := cxt.Client.RebuildCluster(cxt.Account, options.name, options.wait)
			if err != nil {
				return err
			}

			console.WriteCluster(cluster)

			return nil
		},
	}

	cmd.ValidArgs = []string{"cluster-name"}
	cmd.Flags().BoolVar(&options.wait, "wait", false, "wait for the previous cluster operation to complete")

	return cmd
}

func init() {
	rootCmd.AddCommand(newRebuildCommand())
}
