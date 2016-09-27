package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getcarina/carina/common"
	"github.com/spf13/cobra"
)

func newEnvCommand() *cobra.Command {
	var options struct {
		name  string
		shell string
		path  string
	}

	var cmd = &cobra.Command{
		Use:   "env <cluster-name>",
		Short: "Show the command to load a cluster's credentials",
		Long:  "Show the command to load a cluster's credentials into the current shell session",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if options.shell == "" {
				shell := os.Getenv("SHELL")
				if shell != "" {
					options.shell = filepath.Base(shell)
					common.Log.WriteDebug("Shell: SHELL (%s)", options.shell)
				} else {
					return errors.New("Shell was not specified. Either use --shell or set SHELL")
				}
			} else {
				common.Log.WriteDebug("Shell: --shell (%s)", options.shell)
			}

			return bindName(args, &options.name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceText, err := cxt.Client.GetSourceCommand(cxt.Account, options.shell, options.name, options.path)
			if err != nil {
				return err
			}

			fmt.Println(sourceText)
			return nil
		},
	}

	cmd.ValidArgs = []string{"cluster-name"}
	cmd.Flags().StringVar(&options.shell, "shell", "", "The parent shell type. Allowed values: bash, fish, powershell, cmd [SHELL]")
	cmd.Flags().StringVar(&options.path, "path", "", "Full path to the directory from which the credentials should be loaded")

	return cmd
}

func init() {
	rootCmd.AddCommand(newEnvCommand())
}
