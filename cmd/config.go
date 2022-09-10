package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/utils"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display your fox configuration",
	Long: `Display your fox configuration.

The following are the available values and what they mean:

  • autoUpdate (bool) [default: true]:
       You can control if fox updates the available
       packages cache automatically before 'install' and 'info'
  • notifyOutdatedVersions (bool) [default: true]:
       You can control if fox notifies you about if a new version is available
       for your installed packages before 'install' and 'info''
`,
	Run: func(cmd *cobra.Command, args []string) {
		out, err := yaml.Marshal(&userConfig)
		utils.CheckErr(err, cmd)
		fmt.Println()
		color.Blue("        This is your configuration, sir.")
		color.Blue("        Located at '~" + constants.ConfigFilePath + "'.")
		color.Blue(" o(_ _)o\n\n")
		err = utils.PrintYAML(string(out))
		utils.CheckErr(err, cmd)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

// initConfig reads in the user's config file.
func initConfig() {
	// create config file if it doesn't exist
	if !utils.FileExistsInHome(constants.ConfigFilePath) {
		err := writeDefaultConfig()
		utils.CheckErr(err, nil)
	}

	// Find home directory.
	home, err := os.UserHomeDir()
	utils.CheckErr(err, nil)
	viper.AddConfigPath(home + constants.ConfigDirectoryPath)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	err = viper.ReadInConfig()
	utils.CheckErr(err, nil)

	err = viper.Unmarshal(&userConfig)
	utils.CheckErr(err, nil)
	viper.Reset() // reset viper since we use it for different config files
	// fmt.Printf("%v", userConfig)
}

// writeDefaultConfig Creates a default configuration file
func writeDefaultConfig() error {
	utils.CreateDirectoryIfNotExistsInHome(constants.ConfigDirectoryPath)
	err := utils.CreateFileIfNotExistsInHome(constants.ConfigFilePath)
	utils.CheckErr(err, nil)

	// Find home directory.
	home, err := os.UserHomeDir()
	utils.CheckErr(err, nil)
	viper.AddConfigPath(home + constants.ConfigDirectoryPath)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	// Global config
	viper.Set("autoUpdate", true)
	viper.Set("notifyOutdatedVersions", true)

	err = viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}
