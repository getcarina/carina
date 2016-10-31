package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newAutoScaleCommand() *cobra.Command {
	var options struct {
		name      string
		autoscale bool
	}

	cmd := &cobra.Command{
		Use:               "autoscale <cluster-name> off/on",
		Short:             "Change the autoscaling setting on a cluster",
		Long:              "Change the autoscaling setting on a cluster",
		Hidden:            true,
		PersistentPreRunE: authenticatedPreRunE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("A cluster name and the autoscale value (off/on) is required")
			}

			switch strings.ToLower(args[1]) {
			case "off", "false", "0":
				options.autoscale = false
			case "on", "true", "1":
				options.autoscale = true
			default:
				return fmt.Errorf("Invalid autoscale value: %s. Allowed values are off and on", args[1])
			}

			return bindClusterNameArg(args, &options.name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cluster, err := cxt.Client.SetAutoScale(cxt.Account, options.name, options.autoscale)
			if err != nil {
				return err
			}

			console.WriteCluster(cluster)

			return nil
		},
	}

	cmd.ValidArgs = []string{"cluster-name"}
	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
