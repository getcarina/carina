package cmd

import (
	"fmt"
	"os"

	"strings"
	"time"

	"github.com/getcarina/carina/client"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/carina/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cxt *context

var rootCmd = &cobra.Command{
	Use:   "carina",
	Short: "Create and interact with clusters on both Rackspace Public and Private Cloud",
	Long:  "Create and interact with clusters on both Rackspace Public and Private Cloud",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := cxt.initialize()
		if err != nil {
			return err
		}

		return checkIsLatest()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	cxt = &context{}

	authHelp := `Authentication:
The user credentials are used to automatically detect the cloud with which the cli should communicate. First, it looks for the Rackspace Public Cloud environment variables, such as CARINA_USERNAME/CARINA_APIKEY or RS_USERNAME/RS_API_KEY. Then it looks for Rackspace Private Cloud environment variables, such as OS_USERNAME/OS_PASSWORD. Use --cloud flag to explicitly select a cloud.

In the following example, the detected cloud is 'private' because --password is specified:
    carina --username bob --password --project admin --auth-endpoint http://example.com/auth/v3 ilovepuppies ls

In the following example, the detected cloud is 'public' because --apikey is specified:
    carina --username bob --apikey abc123 ls

In the following example, 'private' is used, even though the Rackspace Public Cloud environment variables are present, because the --cloud is specified:
    carina --cloud private ls

See https://github.com/getcarina/carina for additional documentation, FAQ and examples
`

	baseDir, err := client.GetCredentialsDir()
	if err != nil {
		panic(err)
	}
	envHelp := fmt.Sprintf(`Environment Variables:
  CARINA_HOME
    directory that stores your cluster tokens and credentials
    current setting: %s
`, baseDir)
	rootCmd.SetUsageTemplate(fmt.Sprintf("%s\n%s\n\n%s", rootCmd.UsageTemplate(), envHelp, authHelp))

	cobra.OnInitialize(initConfig)

	// Global configuration flags
	//rootCmd.PersistentFlags().StringVar(&cxt.ConfigFile, "config", "", "config file (default is $HOME/.carina/config)")
	rootCmd.PersistentFlags().BoolVar(&cxt.CacheEnabled, "cache", true, "Cache API tokens and update times")
	rootCmd.PersistentFlags().BoolVar(&cxt.Debug, "debug", false, "Print additional debug messages to stdout")
	rootCmd.PersistentFlags().BoolVar(&cxt.Silent, "silent", false, "Do not print to stdout")

	// Account flags
	rootCmd.PersistentFlags().StringVar(&cxt.Username, "username", "", "Username [CARINA_USERNAME/RS_USERNAME/OS_USERNAME]")
	rootCmd.PersistentFlags().StringVar(&cxt.APIKey, "api-key", "", "Public Cloud API Key [CARINA_APIKEY/RS_API_KEY]")
	rootCmd.PersistentFlags().StringVar(&cxt.Password, "password", "", "Private Cloud Password [OS_PASSWORD]")
	rootCmd.PersistentFlags().StringVar(&cxt.Project, "project", "", "Private Cloud Project Name [OS_PROJECT_NAME]")
	rootCmd.PersistentFlags().StringVar(&cxt.Domain, "domain", "", "Private Cloud Domain Name [OS_DOMAIN_NAME]")
	rootCmd.PersistentFlags().StringVar(&cxt.Region, "region", "", "Private Cloud Region Name [OS_REGION_NAME]")
	// --auth-endpoint can also override the authentication endpoint for public Carina as well, but that's only helpful for local development
	rootCmd.PersistentFlags().StringVar(&cxt.AuthEndpoint, "auth-endpoint", "", "Private Cloud Authentication endpoint [OS_AUTH_URL]")
	rootCmd.PersistentFlags().StringVar(&cxt.Endpoint, "endpoint", "", "Custom API endpoint [CARINA_ENDPOINT/OS_ENDPOINT]")
	rootCmd.PersistentFlags().StringVar(&cxt.CloudType, "cloud", "", "The cloud type: public or private")

	// --endpoint can override the API endpoint for both Carina and Magnum, hidden since it's only helpful for local development
	rootCmd.PersistentFlags().MarkHidden("endpoint")

	// Don't show usage on errors
	rootCmd.SilenceUsage = true
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cxt.ConfigFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cxt.ConfigFile)
	}

	viper.SetConfigName("carina")        // name of config file (without extension)
	viper.AddConfigPath("$HOME/.carina") // adding home directory as first search path
	viper.AutomaticEnv()                 // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func checkIsLatest() error {
	if !cxt.CacheEnabled {
		return nil
	}

	ok, err := shouldCheckForUpdate()
	if !ok {
		return err
	}
	common.Log.WriteDebug("Checking for newer releases of the carina cli...")

	rel, err := version.LatestRelease()
	if err != nil {
		common.Log.WriteWarning("# Unable to fetch information about the latest release of %s. %s\n.", os.Args[0], err)
		return nil
	}
	common.Log.WriteDebug("Latest: %s", rel.TagName)

	latest, err := version.ExtractSemver(rel.TagName)
	if err != nil {
		common.Log.WriteWarning("# Trouble parsing latest tag (%v): %s", rel.TagName, err)
		return nil
	}

	current, err := version.ExtractSemver(version.Version)
	if err != nil {
		common.Log.WriteWarning("# Trouble parsing current tag (%v): %s", version.Version, err)
		return nil
	}
	common.Log.WriteDebug("Installed: %s", version.Version)

	if latest.Greater(current) {
		common.Log.WriteWarning("# A new version of the Carina client is out, go get it!")
		common.Log.WriteWarning("# You're on %v and the latest is %v", current, latest)
		common.Log.WriteWarning("# https://github.com/getcarina/carina/releases")
	}

	return nil
}

func shouldCheckForUpdate() (bool, error) {
	lastCheck := cxt.Client.Cache.LastUpdateCheck

	// If we last checked recently, don't check again
	if lastCheck.Add(12 * time.Hour).After(time.Now()) {
		common.Log.Debug("Skipping check for a new release as we have already checked recently")
		return false, nil
	}

	err := cxt.Client.Cache.SaveLastUpdateCheck(time.Now())
	if err != nil {
		return false, err
	}

	if strings.Contains(version.Version, "-dev") || version.Version == "" {
		common.Log.Debug("Skipping check for new release because this is a developer build")
		return false, nil
	}

	return true, nil
}
