/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the available packages cache",
	Long: `
To avoid fetching all packages on every command, fox saves a cache
of the available packages from your repositories config.

This command forces an update on such cache.

You can control if fox updates automatically before 'install', 'info', and 'list'
in your config file with the option: autoUpdate`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		color.Blue("             Updating available packages cache")
		color.Blue(" ᕕ(⌐■_■)ᕗ ♪♬\n\n")

		started := time.Now().UnixMilli()
		spin := spinner.New(constants.Clocks, 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		_ = spin.Color("bold", "fgHiYellow")
		spin.Start()

		// So that you get your cursor back
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			spin.Stop()
			color.Yellow(" 😜 Operation aborted")
			os.Exit(1)
		}()

		err := repositories.UpdatePackagesCache(repositoriesConfig, true)
		if err != nil {
			spin.Stop()
			utils.CheckErr(err, cmd)
		}

		spin.Stop()
		ended := time.Now().UnixMilli()
		timeItTook := float64(ended-started) / 1000
		color.Green(" Updated in %.2f seconds!", timeItTook)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
