package cmd

import (
	"github.com/getcarina/carina/client"
	"github.com/getcarina/carina/console"
	"github.com/spf13/cobra"
)

func newCredentialsCommand() *cobra.Command {
	var options struct {
		name string
		path string
	}

	var cmd = &cobra.Command{
		Use:   "credentials <cluster-name>",
		Short: "Download a cluster's credentials",
		Long:  "Download a cluster's credentials",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return bindClusterNameArg(args, &options.name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			credentialsPath, err := cxt.Client.DownloadClusterCredentials(cxt.Account, options.name, options.path)
			if err != nil {
				return err
			}

			console.Write("#")
			console.Write("# Credentials written to \"%s\"", credentialsPath)
			console.Write(client.CredentialsNextStepsString(options.name))
			console.Write("#")

			return nil
		},
	}

	cmd.ValidArgs = []string{"cluster-name"}
	cmd.Flags().StringVar(&options.path, "path", "", "Full path to the directory where the credentials should be saved")
	cmd.SetUsageTemplate(cmd.UsageTemplate())

	return cmd
}
