package commands

import (
	"strings"

	"github.com/base-al/bui/version"
	"github.com/base-go/mamba"
)

var versionCmd = &mamba.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *mamba.Command, args []string) {
		info := version.GetBuildInfo()

		// Print version info
		cmd.PrintInfo(info.String())

		// Check for updates
		release, err := version.CheckLatestVersion()
		if err != nil {
			return
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		if version.HasUpdate(info.Version, latestVersion) {
			// Check if it's a major version upgrade
			if isMajorVersionUpgrade(info.Version, latestVersion) {
				cmd.PrintWarning("\nMAJOR VERSION AVAILABLE: " + info.Version + " -> " + latestVersion)
				if strings.HasPrefix(latestVersion, "2.") && strings.HasPrefix(info.Version, "1.") {
					cmd.PrintInfo("NEW in v2.0.0: Automatic Relationship Detection!")
					cmd.PrintInfo("   Fields ending with '_id' now auto-generate GORM relationships")
				}
				cmd.PrintWarning("This is a major version with potential breaking changes.")
				cmd.PrintInfo("Changelog: " + release.HTMLURL)
				cmd.PrintInfo("\nTo upgrade: bui upgrade")
			} else {
				updateMsg := version.FormatUpdateMessage(
					info.Version,
					latestVersion,
					release.HTMLURL,
					release.Body,
				)
				cmd.PrintWarning(updateMsg)
			}
		} else {
			cmd.PrintSuccess("\nYou're up to date! Using the latest version " + info.Version)
		}
	},
}


func init() {
	rootCmd.AddCommand(versionCmd)
}

// isMajorVersionUpgrade checks if the upgrade is a major version change
func isMajorVersionUpgrade(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	if len(currentParts) == 0 || len(latestParts) == 0 {
		return false
	}

	return currentParts[0] != latestParts[0]
}
