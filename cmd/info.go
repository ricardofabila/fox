package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/repositories"
	repositoriesTypes "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info about a specific package",
	Long:  `Get info about a specific package`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		if len(args) != 1 {
			color.Red("I only support one argument. Given: [%s]", strings.Join(args, ", "))
			_ = cmd.Help()
			os.Exit(1)
		}

		desiredRepo := args[0]

		packages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, false)
		utils.CheckErr(err, cmd)

		if userConfig.NotifyOutdatedVersions {
			installs := installations.LoadInstallations()
			installations.NotifyNewVersions(packages, installs)
		}

		var found *repositoriesTypes.Package
		for _, repo := range packages {
			if strings.EqualFold(desiredRepo, repo.ExecutableName) {
				found = &repo
				break
			}
		}

		if found == nil {
			color.Yellow("       Sorry, I couldn't find a package with the name: " + desiredRepo)
			color.Yellow("       Try running 'fox update' first.")
			color.Yellow(" ٩(๏̯๏)۶")
			fmt.Println()
			os.Exit(1)
		}

		detailsText := tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
		detailsText.SetText(renderDetails(*found, func(s string) {}))
		fmt.Println(detailsText.GetText(true))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
