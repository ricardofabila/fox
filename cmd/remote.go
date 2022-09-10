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

type RemoteFlags struct {
	url  string
	kind string // type is a keyword
}

var remoteFlags = RemoteFlags{
	url:  "",
	kind: "",
}

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Add a remote to your repositories.yaml file",
	Long:  `Add a remote to your repositories.yaml file`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			utils.CheckErr(fmt.Errorf(fmt.Sprintf("'remote' takes not arguments, given: [%s]", strings.Join(args, ", "))), cmd)
		}

		remoteFlags.url = strings.TrimSpace(remoteFlags.url)
		if remoteFlags.url == "" {
			utils.CheckErr(fmt.Errorf("--url is required, and can't be empty"), cmd)
		}

		remoteFlags.kind = strings.TrimSpace(remoteFlags.kind)
		if remoteFlags.kind == "" {
			utils.CheckErr(fmt.Errorf("--kind is required, and can't be empty"), cmd)
		}

		if !lo.Contains([]string{"github", "open"}, remoteFlags.kind) {
			utils.CheckErr(fmt.Errorf("error, the remote type '"+remoteFlags.kind+"' is not supported. Only 'github' and 'open' are valid values."), cmd)
		}

		// check for duplicates
		for _, remote := range repositoriesConfig.Remotes {
			if strings.EqualFold(remote.URL, remoteFlags.url) && strings.EqualFold(remote.Type, remoteFlags.kind) {
				utils.CheckErr(fmt.Errorf("the remote with the url '"+remoteFlags.url+"' "+
					"and the type '"+remoteFlags.kind+"' already exists"), nil)
			}
		}

		// read the repositories.yaml file and add to the list
		home, err := os.UserHomeDir()
		utils.CheckErr(err, cmd)
		viper.AddConfigPath(home + constants.ConfigDirectoryPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("repositories")
		repositoriesConfig.Remotes = append(repositoriesConfig.Remotes, repositoriesTypes.Remote{URL: remoteFlags.url, Type: remoteFlags.kind})
		viper.Set("remotes", repositoriesConfig.Remotes)
		viper.Set("packages", repositoriesConfig.Packages)
		err = viper.WriteConfig()
		utils.CheckErr(err, cmd)
		viper.Reset() // reset viper since we use it for different config files
		color.Green(" 🦊 remote added, Run 'fox update' to rebuild the cache.")
	},
}

func init() {
	remoteCmd.Flags().StringVarP(&remoteFlags.url, "url", "u", "", "The url of the repository")
	remoteCmd.Flags().StringVarP(&remoteFlags.kind, "type", "t", "", "The type of the remote. It can be one of: github|open")
	addCmd.AddCommand(remoteCmd)
}
