package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ricardofabila/fox/src/constants"
	repositoriesTypes "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

type PackageFlags struct {
	path           string
	executableName string
	kind           string // type is a keyword
	dependsOn      string
}

var packageFlags = PackageFlags{
	path:           "",
	executableName: "",
	kind:           "",
	dependsOn:      "",
}

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Add a package entry to your repositories.yaml file",
	Long: `
	Add a package entry to your repositories.yaml file
	Add a package:
	$ fox add package --path="OWNER/REPO" --executableName="a-name" --type="script" --dependsOn="bash,curl"
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			utils.CheckErr(fmt.Errorf(fmt.Sprintf("'package' takes not arguments, given: [%s]", strings.Join(args, ", "))), cmd)
		}

		packageFlags.path = strings.TrimSpace(packageFlags.path)
		if packageFlags.path == "" {
			utils.CheckErr(fmt.Errorf("--path is required, and can't be empty"), cmd)
		}

		packageFlags.executableName = strings.TrimSpace(packageFlags.executableName)
		if packageFlags.executableName == "" {
			utils.CheckErr(fmt.Errorf("--executableName is required, and can't be empty"), cmd)
		}

		packageFlags.kind = strings.TrimSpace(packageFlags.kind)
		if packageFlags.kind == "" {
			utils.CheckErr(fmt.Errorf("--kind is required, and can't be empty"), cmd)
		}

		if !lo.Contains([]string{"script", "binary"}, packageFlags.kind) {
			utils.CheckErr(fmt.Errorf("error, the package type '"+packageFlags.kind+"' is not supported. Only 'script' and 'binary' are valid values."), cmd)
		}

		packageFlags.dependsOn = strings.TrimSpace(packageFlags.dependsOn)
		dependsOn := strings.Split(packageFlags.dependsOn, ",")

		// read the repositories.yaml file and add to the list
		home, err := os.UserHomeDir()
		utils.CheckErr(err, cmd)
		viper.AddConfigPath(home + constants.ConfigDirectoryPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("repositories")
		configPackage := repositoriesTypes.ConfigPackage{
			Path:           packageFlags.path,
			Type:           packageFlags.kind,
			ExecutableName: packageFlags.executableName,
		}

		if len(dependsOn) > 0 {
			configPackage.DependsOn = dependsOn
		}

		// check for duplicates
		for _, p := range repositoriesConfig.Packages {
			if strings.EqualFold(p.Path, packageFlags.path) {
				utils.CheckErr(fmt.Errorf("the package with the path '"+packageFlags.path+"' already exists"), nil)
			}
		}

		repositoriesConfig.Packages = append(repositoriesConfig.Packages, configPackage)
		viper.Set("remotes", repositoriesConfig.Remotes)
		viper.Set("packages", repositoriesConfig.Packages)
		err = viper.WriteConfig()
		utils.CheckErr(err, cmd)
		viper.Reset() // reset viper since we use it for different config files
		color.Green(" 🦊 package added, Run 'fox update' to rebuild the cache.")
	},
}

func init() {
	packageCmd.Flags().StringVar(&packageFlags.path, "path", "", "The github path of the repository in official format: OWNER/REPO")
	packageCmd.Flags().StringVar(&packageFlags.executableName, "executableName", "", "The name the package will install as by default")
	packageCmd.Flags().StringVar(&packageFlags.kind, "type", "", "The type of the remote. It can be one of: binary|script")
	packageCmd.Flags().StringVar(&packageFlags.dependsOn, "dependsOn", "", "(optional) - a comma separated list of dependencies")
	addCmd.AddCommand(packageCmd)
}
