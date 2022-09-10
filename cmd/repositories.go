package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/ricardofabila/fox/src/constants"
	repositoriesTypes "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

// repositoriesCmd represents the repositories command
var repositoriesCmd = &cobra.Command{
	Use:   "repositories",
	Short: "Print your repositories file",
	Long:  `Print your repositories file`,
	Run: func(cmd *cobra.Command, args []string) {
		out, err := yaml.Marshal(&repositoriesConfig)
		utils.CheckErr(err, cmd)
		fmt.Println()
		color.Blue("          This is your repositories configuration, sir.")
		color.Blue("          Located at '~" + constants.RepositoriesFilePath + "'.")
		color.Blue(" ( ^-^)_旦\n\n")
		err = utils.PrintYAML(string(out))
		utils.CheckErr(err, cmd)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(repositoriesCmd)
}

// initRepositories reads in the repositories file.
func initRepositories() {
	// create config file if it doesn't exist
	if !utils.FileExistsInHome(constants.RepositoriesFilePath) {
		err := writeDefaultRepositoriesFile()
		utils.CheckErr(err, repositoriesCmd)
	}

	// Find home directory.
	home, err := os.UserHomeDir()
	utils.CheckErr(err, repositoriesCmd)

	viper.AddConfigPath(home + constants.ConfigDirectoryPath)
	viper.SetConfigType("yaml")
	viper.SetConfigName("repositories")

	err = viper.ReadInConfig()
	utils.CheckErr(err, repositoriesCmd)

	err = viper.Unmarshal(&repositoriesConfig)
	utils.CheckErr(err, repositoriesCmd)
	viper.Reset() // reset viper since we use it for different config files
	// fmt.Printf("%v", repositoriesConfig)
}

// writeDefaultRepositoriesFile Creates a default repositories file
func writeDefaultRepositoriesFile() error {
	utils.CreateDirectoryIfNotExistsInHome(constants.ConfigDirectoryPath)
	err := utils.CreateFileIfNotExistsInHome(constants.RepositoriesFilePath)
	utils.CheckErr(err, nil)

	// Find home directory.
	home, err := os.UserHomeDir()
	utils.CheckErr(err, repositoriesCmd)
	viper.AddConfigPath(home + constants.ConfigDirectoryPath)
	viper.SetConfigType("yaml")
	viper.SetConfigName("repositories")

	// Repositories config
	viper.Set("remotes", repositoriesTypes.Remotes{})
	viper.Set("packages", repositoriesTypes.ConfigPackages{})

	err = viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}
