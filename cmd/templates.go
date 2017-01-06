package cmd

import (
	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newTemplatesCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "templates",
		Short:             "List cluster templates",
		Long:              "List cluster templates",
		PersistentPreRunE: authenticatedPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := cxt.Client.ListClusterTemplates(cxt.Account)
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

	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
