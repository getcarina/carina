package cmd

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/getcarina/carina/client"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/carina/version"
	"github.com/spf13/cobra"
)

func bindClusterNameArg(args []string, name *string) error {
	if len(args) < 1 {
		return errors.New("A cluster name is required")
	}
	*name = args[0]
	return nil
}

func authenticatedPreRunE(cmd *cobra.Command, args []string) error {
	err := cxt.initialize()
	if err != nil {
		return err
	}

	return checkIsLatest()
}

func unauthenticatedPreRunE(cmd *cobra.Command, args []string) error {
	cxt.Client = client.NewClient(cxt.CacheEnabled)

	return checkIsLatest()
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

	latest, err := semver.NewVersion(rel.TagName)
	if err != nil {
		common.Log.WriteWarning("# Trouble parsing latest tag (%v): %s", rel.TagName, err)
		return nil
	}

	current, err := semver.NewVersion(version.Version)
	if err != nil {
		common.Log.WriteWarning("# Trouble parsing current tag (%v): %s", version.Version, err)
		return nil
	}
	common.Log.WriteDebug("Installed: %s", version.Version)

	if latest.GreaterThan(current) {
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
