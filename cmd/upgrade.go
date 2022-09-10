package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/repositories"
	"github.com/ricardofabila/fox/src/types"
	repositories2 "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade installed packages to the latest version",
	Long:  `Upgrade installed packages to the latest version`,
	Example: `
	Upgrade a package:
	$ fox upgrade <package_name>

	Upgrade multiple packages:
	$ fox upgrade <package_name> <package_name> <package_name>

	Upgrade all packages:
	$ fox upgrade

	Upgrade fox:
	$ fox upgrade fox
`,
	Run: func(cmd *cobra.Command, args []string) {
		// To upgrade first find the installations FindInstallations
		// if you find at least one, go fetch the packages. Find the by the original executable name.
		// get the latest version. Check all the installations, get the one for the executable name. No aliases.
		// if they differ, go install latest, with the alias or executable name
		args = lo.Map(args, func(t string, _ int) string {
			return strings.TrimSpace(t)
		})

		if len(args) > 1 && lo.Contains(args, "fox") {
			fmt.Println()
			color.Red("I cannot upgrade myself and other packages at the same time. Given: [%s]", strings.Join(args, ", "))
			color.Red("If you want to upgrade fox. Run 'fox upgrade fox' without other arguments")
			_ = cmd.Help()
			os.Exit(1)
		}

		upgradeAll := len(args) == 0
		// Upgrade fox itself
		if !upgradeAll && strings.TrimSpace(args[0]) == "fox" {
			color.Blue(" Upgrading fox. You may need to execute this command with 'sudo'")
			upgradeName := "fox-upgrade"
			availablePackages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, true)
			utils.CheckErr(err, cmd)
			err = installations.InstallPackage(availablePackages, "fox", upgradeName, false, userConfig, true, true)
			utils.CheckErr(err, cmd)
			// execute a rename of the downloaded file
			err = utils.MoveFile(constants.FoxBinPath+upgradeName, constants.FoxBinPath+"fox")
			utils.CheckErr(err, cmd)
			color.Green(" 🦊 done! You have the latest version of fox.")
			return
		}

		args = lo.Filter(args, func(t string, _ int) bool {
			return !strings.EqualFold(t, "fox")
		})

		installs := installations.LoadInstallations()
		if len(installs.Installations) == 0 {
			fmt.Println()
			color.Blue("You have no installed packages, sir ヾ(_ _。）")
			fmt.Println()
			return
		}

		aliasedOrWithVersion := lo.Filter(installs.Installations, func(i types.Installation, _ int) bool {
			return i.Alias != ""
		})
		if len(aliasedOrWithVersion) > 0 {
			fmt.Println()
			color.Yellow(" Warning! Some of your installations are aliased or have a specific version.")
			color.Yellow(" Upgrading aliased and versioned installations is not supported.")
			color.Yellow(" You must upgrade those manually by uninstalling them and reinstalling them.")
			color.Yellow(" The following have been found:")
			for _, i := range aliasedOrWithVersion {
				color.Yellow("      • " + i.RealName)
			}
			fmt.Println()
		}

		// filter out the packages that are already on the latest versions
		availablePackages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, true)
		utils.CheckErr(err, cmd)
		canBeUpgraded := installations.GetUpgradable(availablePackages, installs)

		if upgradeAll {
			args = lo.Map(canBeUpgraded, func(p repositories2.Package, _ int) string {
				return p.ExecutableName
			})
		}

		args = lo.Filter(args, func(n string, _ int) bool {
			existing := installations.FindInstallation(n)
			if existing == nil {
				color.Yellow("       Sorry, I couldn't find an installation for: " + n)
				color.Yellow(" ٩(๏̯๏)۶")
				fmt.Println()
				return false
			}

			// notify that package is already at the latest version
			pkg := installations.FindPackage(availablePackages, *existing)
			if pkg != nil {
				// if !strings.EqualFold(strings.TrimSpace(installation.Version), strings.TrimSpace(pkg.LatestVersion)) {
				if strings.Contains(strings.TrimSpace(pkg.LatestVersion), strings.TrimSpace(existing.Version)) {
					color.Blue(" %s is already at the latest version", n)
					fmt.Println()
					return false
				}
			}

			return true
		})

		// find the installation for the given package
		var willBeUpgraded []repositories2.Package
		for _, c := range canBeUpgraded {
			for _, n := range args {
				if c.ExecutableName == n {
					willBeUpgraded = append(willBeUpgraded, c)
					continue
				}
			}
		}

		// Avoid printing more things, return early
		if len(willBeUpgraded) == 0 && len(args) == 1 {
			return
		}

		if len(willBeUpgraded) == 0 {
			fmt.Println()
			color.Blue("You have no installed packages I can update, sir ヾ(_ _。）")
			fmt.Println()
			return
		}

		if len(willBeUpgraded) > 0 {
			fmt.Println()
			color.Blue(" The following packages will be upgraded:")
			fmt.Printf(" %s\n", strings.Join(lo.Map(willBeUpgraded, func(p repositories2.Package, _ int) string {
				return p.ExecutableName
			}), ", "))
			fmt.Println()
		}

		// TODO: prompt user for confirmation

		for _, pkg := range willBeUpgraded {
			color.Green(" Upgrading: " + pkg.ExecutableName)
			err = installations.InstallPackage(availablePackages, pkg.ExecutableName, "", false, userConfig, false, false)
			utils.CheckErr(err, cmd)
			fmt.Println()
		}

		fmt.Println(" Thank you for participating in this")
		fmt.Println(" Aperture Science computer-aided enrichment activity.")
		color.Green(" ✅ Upgrade complete!")
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
