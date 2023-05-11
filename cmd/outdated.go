package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

// outdatedCmd represents the outdated command
var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "List the packages you can upgrade",
	Long: `

	Print the packages that can be upgraded to a new version
	$ fox outdated`,
	Run: func(cmd *cobra.Command, args []string) {
		color.Blue(" ⏳ Loading packages list...")
		packages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, false)
		utils.CheckErr(err, cmd)

		err = checkForNewFoxVersion()
		utils.CheckErr(err, nil)
		installs := installations.LoadInstallations()
		installations.NotifyNewVersions(packages, installs)
	},
}

func init() {
	rootCmd.AddCommand(outdatedCmd)
}
