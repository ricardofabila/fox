package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
	"golang.org/x/sys/execabs"

	"github.com/ricardofabila/build"
	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/types"
	repositoriesTypes "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

var repositoriesConfig repositoriesTypes.Config
var userConfig types.UserConfig

const VERSION = "1.0.3"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fox",
	Short: "Use our APIs the easy way",
	Long: `
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⡀⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣾⠙⠻⢶⣄⡀⠀⠀⠀⢀⣤⠶⠛⠛⡇⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⣇⠀⠀⣙⣿⣦⣤⣴⣿⣁⠀⠀⣸⠇⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⣡⣾⣿⣿⣿⣿⣿⣿⣿⣷⣌⠋⠀⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⣿⣷⣄⡈⢻⣿⡟⢁⣠⣾⣿⣦⠀⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⣿⣿⣿⣿⠘⣿⠃⣿⣿⣿⣿⡏⠀⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⠀⠈⠛⣰⠿⣆⠛⠁⠀⡀⠀⠀⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣼⣿⣦⠀⠘⠛⠋⠀⣴⣿⠁⠀⠀⠀⠀⠀
             ⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣶⣾⣿⣿⣿⣿⡇⠀⠀⠀⢸⣿⣏⠀⠀⠀⠀⠀     ____  ⠀
             ⠀⠀⠀⠀⠀⣠⣶⣿⣿⣿⣿⣿⣿⣿⣿⠿⠿⠀⠀⠀⠾⢿⣿⠀⠀⠀⠀⠀    / __/___  _  __
             ⠀⠀⠀⣠⣿⣿⣿⣿⣿⣿⡿⠟⠋⣁⣠⣤⣤⡶⠶⠶⣤⣄⠈⠀⠀⠀⠀⠀   / /_/ __ \| |/_/
             ⠀⠀⢰⣿⣿⣮⣉⣉⣉⣤⣴⣶⣿⣿⣋⡥⠄⠀⠀⠀⠀⠉⢻⣄⠀⠀⠀⠀  / __/ /_/ />  <  ⠀
             ⠀⠀⠸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣟⣋⣁⣤⣀⣀⣤⣤⣤⣤⣄⣿⡄⠀⠀⠀ /_/  \____/_/|_|  ⠀
             ⠀⠀⠀⠙⠿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠛⠋⠉⠁⠀⠀⠀⠀⠈⠛⠃⠀⠀⠀ ⠀
             ⠀⠀⠀⠀⠀⠀⠉⠉⠉⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀

    fox is a (simple!) package manager to install your own tools with ease`,
	Version: VERSION,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cc.Init(&cc.Config{
		RootCmd:       rootCmd,
		Headings:      cc.HiMagenta + cc.Underline,
		Commands:      cc.Yellow + cc.Bold,
		Example:       cc.Italic,
		ExecName:      cc.HiBlue + cc.Bold,
		Flags:         cc.Green,
		FlagsDataType: cc.Underline + cc.HiWhite,
	})
	build.Boostrap()

	// Check that gh is installed
	_, err := execabs.LookPath("gh")
	if err != nil {
		if len(os.Args) == 1 || (len(os.Args) > 1 && os.Args[1] != "gh") {
			color.Yellow("\n Looks like you don't have gh installed or is not in your $PATH.\n\n")
			color.Yellow("\n I can install it for you, just run `fox gh`.\n\n")
			os.Exit(1)
		}
	}

	// register the current version
	err = checkForNewFoxVersion()
	if err != nil {
		_ = utils.PrintAndReturnError(err.Error())
	}

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
	initRepositories()
	initConfig()
}

func checkForNewFoxVersion() error {
	err := utils.CreateFileIfNotExists(constants.FoxVersionPath)
	if err != nil {
		return err
	}

	if !utils.FileExists(constants.FoxVersionPath) {
		return fmt.Errorf("error, version file doesn't exist")
	}

	stats, err := os.Stat(constants.FoxVersionPath)
	if err != nil {
		return err
	}

	sixHours := time.Hour * 6
	if (time.Now().UnixMilli() - stats.ModTime().UnixMilli()) > sixHours.Milliseconds() {
		// go fetch the latest version from the internet
		latestFoxRelease, er := utils.GetFromAPI("https://api.github.com/repos/ricardofabila/fox/releases/latest")
		if er != nil {
			return er
		}
		if !utils.IsValidJSON(string(latestFoxRelease)) {
			return fmt.Errorf("Error, the response by GitHub was not valid JSON: \n" + string(latestFoxRelease))
		}
		release := repositoriesTypes.Release{}
		er = json.Unmarshal(latestFoxRelease, &release)
		if er != nil {
			return er
		}

		latestVersion := release.Tag
		if !strings.Contains(strings.ToLower(latestVersion), strings.ToLower(VERSION)) {
			color.Yellow(" There is a new version of fox available: " + latestVersion)
			color.Yellow("    Your version is: " + VERSION)
			color.Yellow("    run 'fox upgrade fox' to install it")
		}

		er = os.WriteFile(constants.FoxVersionPath, []byte(VERSION), 0666)
		if er != nil {
			return err
		}
	}

	return nil
}
