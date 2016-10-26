package cmd

import (
	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newTemplatesCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "templates",
		Short:   "List cluster templates",
		Long:    "List cluster templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := cxt.Client.ListClusterTemplates(cxt.Account)
			if err != nil {
				return err
			}

			data := [][]string{[]string{"Name", "COE", "Host Type"}}
			for _, template := range templates {
				data = append(data, []string{template.GetName(), template.GetCOE(), template.GetHostType()})
			}
			console.WriteTable(data)

			return nil
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newTemplatesCommand())
}
