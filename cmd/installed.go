package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/repositories"
	"github.com/ricardofabila/fox/src/types"
	repositoriesTypes "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

type InstalledFlags struct {
	print bool
}

var installedFlags = InstalledFlags{
	print: false,
}

// installedCmd represents the installed command
var installedCmd = &cobra.Command{
	Use:   "installed",
	Short: "List the packages you have installed",
	Long:  `Display a list of the packages you have installed`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			fmt.Println()
			color.Red("This command doesn't receive any arguments. Got: [" + strings.Join(args, ", ") + "]")
			fmt.Println()
			_ = cmd.Help()
			return
		}

		installs := installations.LoadInstallations()

		if len(installs.Installations) == 0 {
			fmt.Println()
			color.Blue("You have no installed packages, sir ヾ(_ _。）")
			fmt.Println()
			return
		}

		if installedFlags.print {
			if userConfig.NotifyOutdatedVersions {
				color.Blue(" ⏳ Loading packages list...")
				packages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, false)
				utils.CheckErr(err, listCmd)
				installations.NotifyNewVersions(packages, installs)
			}

			color.Blue("\n ~(^._.) Listing installed packages:\n\n")
			fmt.Println("        ______________________________________________________")
			fmt.Println()

			installs.Installations = lo.Filter(installs.Installations, func(i types.Installation, _ int) bool {
				return i.IsVisible()
			})
			sort.Slice(installs.Installations, func(i, j int) bool {
				return installs.Installations[i].RealName < installs.Installations[j].RealName
			})

			for _, i := range installs.Installations {
				color.Green("        • Package name: " + i.ExecutableName)
				color.Magenta("	• Version: " + i.Version)
				if i.Alias != "" {
					color.Yellow("          Alias: " + i.Alias)
				}
				fmt.Println("          Real executable path: " + constants.FoxBinPath + i.RealName)
				fmt.Println("          Installed at: " + time.UnixMilli(i.Timestamp).String())
				fmt.Println("        ______________________________________________________")
				fmt.Println()
			}

			fmt.Println()
			return
		}

		packages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, false)
		utils.CheckErr(err, listCmd)

		if userConfig.NotifyOutdatedVersions {
			installations.NotifyNewVersions(packages, installs)
		}

		packages = lo.Filter(packages, func(p repositoriesTypes.Package, _ int) bool {
			return len(p.InstalledVersions) > 0
		})
		err = interactiveRender(packages, "Installed packages")
		utils.CheckErr(err, cmd)
	},
}

func init() {
	installedCmd.Flags().BoolVarP(&installedFlags.print, "print", "p", false, "Outputs the list of installed packages to stdout")
	rootCmd.AddCommand(installedCmd)
}
