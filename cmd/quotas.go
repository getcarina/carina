package cmd

import (
	"strconv"

	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newQuotasCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "quotas",
		Short:             "Show the user's quotas",
		Long:              "Show the user's quotas",
		PersistentPreRunE: authenticatedPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			quotas, err := cxt.Client.GetQuotas(cxt.Account)
			if err != nil {
				return err
			}

			maxClusters := strconv.Itoa(quotas.GetMaxClusters())
			maxNodesPerCluster := strconv.Itoa(quotas.GetMaxNodesPerCluster())

			data := [][]string{
				[]string{"MaxClusters", "MaxNodesPerCluster"},
				[]string{maxClusters, maxNodesPerCluster},
			}
			console.WriteTable(data)

			return nil
		},
	}

	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
