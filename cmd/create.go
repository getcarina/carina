package cmd

import (
	"errors"

	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var options struct {
		name     string
		template string
		nodes    int
		wait     bool
	}

	var cmd = &cobra.Command{
		Use:   "create <cluster-name>",
		Short: "Create a cluster",
		Long:  "Create a cluster",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if options.nodes < 1 {
				return errors.New("--nodes must be >= 1")
			}

			return bindClusterNameArg(args, &options.name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := cxt.Client.CreateCluster(cxt.Account, options.name, options.template, options.nodes, options.wait)
			if err != nil {
				return err
			}

			console.WriteCluster(cluster)

			return nil
		},
	}

	cmd.ValidArgs = []string{"cluster-name"}
	cmd.Flags().StringVar(&options.template, "template", "", "Name of the template, defining the cluster topology and configuration")
	cmd.Flags().IntVar(&options.nodes, "nodes", 1, "Number of nodes for the initial cluster")
	cmd.Flags().BoolVar(&options.wait, "wait", false, "Wait for the cluster to become active")

	return cmd
}
