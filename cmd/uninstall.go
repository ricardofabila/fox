package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/utils"
)

// uninstallCmd yeets packages from your system
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove packages from your system",
	Long: `
 You can remove packages from your system.
 Useful if you want to try installing a different version
 or want to reinstall with a different name (via the --as flag)

 (┛ಠ_ಠ)┛彡┻━┻
`,
	Example: `
	Uninstall a package:
	$ fox uninstall <package_name>

	Uninstall a package that was installed with a different name using the --as flag during 'fox install':
	$ fox uninstall <custom_name>
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		if len(args) != 1 {
			color.Red("I only support one argument. Given: [%s]", strings.Join(args, ", "))
			return
		}

		pkgName := strings.TrimSpace(args[0])
		if pkgName == "fox" {
			fmt.Println()
			color.Yellow("                 Was it something I said?")
			color.Yellow(" .·´¯`(>▂<)´¯`·. \n\n")
			return
		}

		color.Blue("Uninstalling: %s", pkgName)
		install := installations.FindInstallation(pkgName)
		if install == nil {
			utils.CheckErr(fmt.Errorf("Error. No installation found for "+pkgName), cmd)
			return // add return so that linter stops complaining
		}

		err := utils.RemoveFile(constants.FoxBinPath + pkgName)
		if err != nil {
			utils.CheckErr(err, cmd)
		}

		installations.DeleteInstallation(*install)
		color.Green("Uninstalled: %s", pkgName)
	},
}

func init() {
	uninstallCmd.Aliases = []string{"yeet", "remove"}
	rootCmd.AddCommand(uninstallCmd)
}
