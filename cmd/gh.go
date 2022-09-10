package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/sys/execabs"

	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/types"
	"github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

// ghCmd represents the gh command
var ghCmd = &cobra.Command{
	Use:   "gh",
	Short: "Install the official GitHub CLI",
	Long: `
In order to install packages, I need the GitHub CLI (gh) to be installed. You can get it by using this command.
For private repos, make sure you have auth set https://cli.github.com/manual/gh_auth (login via ssh is recommended).`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := execabs.LookPath("gh")
		if err == nil {
			color.Green("\n ✅ gh is already installed.\n\n")

			color.White(" Running `gh auth status`.")
			color.White(" If you are having problems installing packages,")
			color.White(" make sure you have the right credentials set:")

			data, err := utils.ExecuteCommandAndGetOutput("gh", []string{"auth", "status"}...)
			if err != nil {
				_ = utils.PrintAndReturnError(err.Error())
			}

			lines := strings.Split(data, "\n")
			for _, line := range lines {
				color.Cyan("       " + line)
			}

			return
		}

		color.Yellow("\n Looks like you don't have gh installed or is not in your $PATH.\n\n")
		latestGhRelease, err := utils.GetFromAPI("https://api.github.com/repos/cli/cli/releases/latest")
		utils.CheckErr(err, cmd)
		if !utils.IsValidJSON(string(latestGhRelease)) {
			utils.CheckErr(fmt.Errorf("Error, the response by GitHub was not valid JSON: \n"+string(latestGhRelease)), cmd)
		}

		release := repositories.Release{}
		err = json.Unmarshal(latestGhRelease, &release)
		utils.CheckErr(err, cmd)

		pkg := repositories.Package{
			ExecutableName: "gh",
			Type:           "binary",
		}

		assetToDownload, err := installations.GetAssetToDownloadForBinary(pkg, release.Assets, false)
		utils.CheckErr(err, cmd)
		color.Magenta(" Fetching the asset " + assetToDownload.Name + " of size " + utils.ByteCountIEC(int64(assetToDownload.Size)))

		err = utils.DownloadFile(assetToDownload.BrowserDownloadURL, assetToDownload.Name)
		utils.CheckErr(err, cmd)

		_, err = installations.ExtractAsset(assetToDownload.Name, pkg.ExecutableName)
		utils.CheckErr(err, cmd)

		err = installations.MoveAssetToBin("gh", "gh")
		utils.CheckErr(err, cmd)

		color.Green("\n\n 🦊 Installed gh, now go install more cool packages!")
		color.Green(" You can see a hand-picked list of public packages running:")
		color.Blue("\n    fox list\n\n")

		// save the installation to notify about new releases
		install := types.Installation{
			Timestamp:      time.Now().UnixMilli(),
			Package:        "cli/cli",
			ExecutableName: "gh",
			RealName:       "gh",
			Version:        release.Tag,
		}

		installations.SaveInstallation(install)
	},
}

func init() {
	rootCmd.AddCommand(ghCmd)
}
