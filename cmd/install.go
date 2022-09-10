package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

type InstallFlags struct {
	alias       string
	force       bool
	interactive bool
}

var installFlags = InstallFlags{
	alias:       "",
	force:       false,
	interactive: false,
}

// installCmd installs packages
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a package",
	Long: `Install an available package.
NOTE: I currently only work in unix systems (macOS and linux). Windows (pseudo)support coming soon.
`,
	Example: `
	Install a package (latest version):
	$ fox install <package_name>

	Install a package and do not prompt for confirmation:
	$ fox install <package_name> -y

	Install specific version of a package (run [fox info <package_name>] to list available versions):
	$ fox install <package_name>@<version>
	$ fox install <package_name>@latest
	$ fox install <package_name>@v1.0.3
	$ fox install <package_name>@2.5.1

	Install a package and change its executable name (to avoid overpopulating your shell config more aliases):
	$ fox install <original_package_name> --as "custom_name"

		This way instead of running
			$ original_package_name
		You can run:
			$ custom_name

	Install multiple packages:
	$ fox install <package_name_1> <package_name_2>
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		if len(args) != 1 && strings.TrimSpace(installFlags.alias) != "" {
			utils.CheckErr(fmt.Errorf(fmt.Sprintf("I can only install one package when using the --as. flag Given: [%s]", strings.Join(args, ", "))), cmd)
		}

		interactive := lo.Ternary(installFlags.interactive, false, true)
		availablePackages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, false)
		utils.CheckErr(err, cmd)

		if len(args) == 1 {
			err = installations.InstallPackage(availablePackages, args[0], installFlags.alias, interactive, userConfig, false, installFlags.force)
			utils.CheckErr(err, cmd)
			return
		}

		args = lo.Filter(args, func(t string, _ int) bool {
			return !strings.EqualFold(t, "fox")
		})

		var successfullyInstalled []string
		for _, p := range args {
			err = installations.InstallPackage(availablePackages, p, installFlags.alias, interactive, userConfig, false, installFlags.force)
			if err != nil {
				color.Yellow("\n\n There has been an error while installing: " + p)
				color.Yellow(" The following packages installed successfully:")
				color.Yellow("    [ " + strings.Join(successfullyInstalled, ", ") + " ]")
				color.Yellow(" The left over packages were:")
				color.Yellow("    [ " + strings.Join(utils.DifferenceStrings(args, successfullyInstalled), ", ") + " ]")
				utils.CheckErr(err, cmd)
			}
			successfullyInstalled = append(successfullyInstalled, p)
		}
	},
}

func init() {
	installCmd.Flags().StringVar(&installFlags.alias, "as", "", "Install a package and change its executable name\n(to avoid overpopulating your shell config more aliases)")
	installCmd.Flags().BoolVarP(&installFlags.force, "force", "f", false, "Force the installation of a package even if you are already at the latest version")
	installCmd.Flags().BoolVarP(&installFlags.interactive, "yes", "y", false, "Do not prompt for confirmation when installing a package")
	installCmd.Aliases = []string{"i"}
	rootCmd.AddCommand(installCmd)
}
