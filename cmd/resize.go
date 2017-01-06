package cmd

import (
	"errors"

	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newResizeCommand() *cobra.Command {
	var options struct {
		name  string
		nodes int
		wait  bool
	}

	var cmd = &cobra.Command{
		Use:               "resize <cluster-name>",
		Short:             "Resize a cluster",
		Long:              "Resize a cluster by setting the number of cluster nodes",
		PersistentPreRunE: authenticatedPreRunE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if options.nodes < 1 {
				return errors.New("--nodes must be >= 1")
			}

			return bindClusterNameArg(args, &options.name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := cxt.Client.ResizeCluster(cxt.Account, options.name, options.nodes, options.wait)
			if err != nil {
				return err
			}

			console.WriteCluster(cluster)

			return nil
		},
	}

	cmd.ValidArgs = []string{"cluster-name"}
	cmd.Flags().IntVar(&options.nodes, "nodes", 1, "The desired number of nodes in the cluster")
	cmd.Flags().BoolVar(&options.wait, "wait", false, "Wait for cluster to finish resizing and return to active")
	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
