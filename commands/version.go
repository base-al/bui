package commands

import (
	"fmt"
	"strings"

	"github.com/base-al/bui/version"
	"github.com/base-go/mamba"
)

var versionCmd = &mamba.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *mamba.Command, args []string) {
		info := version.GetBuildInfo()
		fmt.Println(info.String())

		// Check for updates
		release, err := version.CheckLatestVersion()
		if err != nil {
			return
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		if version.HasUpdate(info.Version, latestVersion) {
			// Check if it's a major version upgrade
			if isMajorVersionUpgrade(info.Version, latestVersion) {
				fmt.Printf("\n🚨 MAJOR VERSION AVAILABLE: %s → %s\n", info.Version, latestVersion)
				if strings.HasPrefix(latestVersion, "2.") && strings.HasPrefix(info.Version, "1.") {
					fmt.Println("🎉 NEW in v2.0.0: Automatic Relationship Detection!")
					fmt.Println("   Fields ending with '_id' now auto-generate GORM relationships")
				}
				fmt.Println("⚠️  This is a major version with potential breaking changes.")
				fmt.Printf("📚 Changelog: %s\n", release.HTMLURL)
				fmt.Println("\nTo upgrade: bui upgrade")
			} else {
				fmt.Print(version.FormatUpdateMessage(
					info.Version,
					latestVersion,
					release.HTMLURL,
					release.Body,
				))
			}
		} else {
			fmt.Printf("\n✨ You're up to date! Using the latest version %s\n", info.Version)
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
