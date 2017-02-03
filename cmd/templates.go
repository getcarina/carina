package cmd

import (
	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newTemplatesCommand() *cobra.Command {
	var options struct {
		name  string
	}

	var cmd = &cobra.Command{
		Use:               "templates",
		Short:             "List cluster templates",
		Long:              "List cluster templates",
		PersistentPreRunE: authenticatedPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := cxt.Client.ListClusterTemplates(cxt.Account, options.name)
			if err != nil {
				return err
			}

			data := [][]string{[]string{"Name", "COE", "Host"}}
			for _, template := range templates {
				data = append(data, []string{template.GetName(), template.GetCOE(), template.GetHostType()})
			}
			console.WriteTable(data)

			return nil
		},
	}

	cmd.Flags().StringVar(&options.name, "name", "", "Filter by name, e.g. Kubernetes*")
	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
