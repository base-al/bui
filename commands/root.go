package commands

import (
	"fmt"
	"strings"

	"github.com/base-al/bui/version"
	"github.com/base-go/mamba"
)

var rootCmd = &mamba.Command{
	Use:   "bui",
	Short: "Bui - Unified CLI for Base Stack",
	Long: `Bui is a unified CLI tool for Base Stack development.
Generate backend modules (Go), frontend modules (Nuxt/TypeScript), and manage your full-stack application.`,
	PersistentPreRun: func(cmd *mamba.Command, args []string) {
		// Skip version check for version and upgrade commands
		if cmd.Name() != "version" && cmd.Name() != "upgrade" {
			if release, err := version.CheckLatestVersion(); err == nil {
				info := version.GetBuildInfo()
				latestVersion := strings.TrimPrefix(release.TagName, "v")
				// Only show update message if there's actually an update
				if version.HasUpdate(info.Version, latestVersion) {
					fmt.Print(version.FormatUpdateMessage(
						info.Version,
						latestVersion,
						release.HTMLURL,
						release.Body,
					))
				}
			}
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add global verbose flag
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
}
