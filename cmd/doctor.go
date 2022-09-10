package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/utils"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use: "doctor",
	Short: `🏥 Check for common issues and recommendations with your fox
		  configuration and overall environment.`,
	// • Your configuration file.
	Long: `This commands checks:
• You have dependencies installed.
• That you have the right permissions for the folder fox uses.`,
	Run: giveItToMeStraightDoctorICanTakeIt,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func giveItToMeStraightDoctorICanTakeIt(*cobra.Command, []string) {
	color.Magenta("\U0001FA7A Aight, let me take a look:\n\n")

	color.Blue("  What the emojis mean:")
	color.White("  ✅ Everything is awesome!")
	color.White("  💉 This is not a problem, but you may have not have the most optimal experience. But you do you. I won't tell 🤫.")
	color.White("  ❌ Something is most definitely wrong.")

	fmt.Println()

	warnings := 0
	danger := 0

	// -------------------------------------------- CONFIG --------------------------------------------
	color.Green("    🔍 Looking at your config file located at ~" + constants.ConfigFilePath + ":\n")

	if userConfig.NotifyOutdatedVersions {
		color.White("                ✅ You have 'notifyOutdatedVersions' set to true")
		color.White("                This means fox will tell you when there is a new version")
		color.White("                for packages that you have installed.")
	} else {
		color.Yellow("                💉 You have 'notifyOutdatedVersions' set to false")
		color.Yellow("                This means fox will not tell you when there is a new version")
		color.Yellow("                for packages that you have installed.")
		warnings++
	}

	if userConfig.AutoUpdate {
		color.White("                ✅ You have 'AutoUpdate' set to true")
		color.White("                This means fox will update its cache automatically")
		color.White("                when running various command, so you always install")
		color.White("                the latest and greatest version available for a package.")
	} else {
		color.Yellow("                💉 You have 'AutoUpdate' set to false")
		color.Yellow("                This means fox will NOT update its cache automatically")
		color.Yellow("                when running various command, so may install an outdated")
		color.Yellow("                version of a package.")
		warnings++
	}

	fmt.Println()

	// -------------------------------------------- DEPENDENCIES --------------------------------------------
	color.Green("    🔍 Looking at dependencies:\n")

	_, err := exec.LookPath("gh")
	if err != nil {
		color.Red("            ❌ You don't have `gh` installed or is not in your $PATH.")
		color.Red("               You won't be able to make any API calls and won't be able to download anything.")
		color.Red("               %s", err)
		danger++
	} else {
		color.White("                ✅ gh is installed.")
	}

	if err == nil {
		color.White("                Running `gh auth status`.")
		color.White("                If you are having problems installing packages,")
		color.White("                make sure you have the right credentials set:")

		data, err := utils.ExecuteCommandAndGetOutput("gh", []string{"auth", "status"}...)
		if err != nil {
			_ = utils.PrintAndReturnError(err.Error())
		}

		lines := strings.Split(data, "\n")
		for _, line := range lines {
			color.Cyan("                	" + line)
		}
	}

	fmt.Println()

	// ---------------------------------------- OS and ARCHITECTURE ----------------------------------------

	color.Green("    🔍 Looking at your OS and ARCHITECTURE:\n")
	color.White("                ✅ Your OS is: " + runtime.GOOS)
	color.White("                ✅ Your ARCHITECTURE is: " + runtime.GOARCH)
	fmt.Println()

	// -------------------------------------------- PERMISSIONS --------------------------------------------

	color.Green("    🔍 Looking at file permissions:\n")

	// r - read
	// w - write
	// x - execute

	// 7 = all rights
	// 6 = read and write
	// 5 = read and execute
	// 4 = read only
	// 3 = execute and write
	// 2 = write only
	// 1 = execute only
	// 0 = no rights

	// desired 777
	// drwxrwxr-x
	// Owner: rwx
	// Group: rwx
	// Other: rwx

	// sudo chmod -R 777 /usr/local/Fox/
	// https://stackoverflow.com/questions/45429210/how-do-i-check-a-files-permissions-in-linux-using-go

	permissions, err := utils.GetPermissions(constants.FoxRootPath)
	if err != nil {
		_ = utils.PrintAndReturnError(err.Error())
	}

	if permissions != "Owner: rwx Group: rwx Other: rwx" {
		color.Red("            ❌ You don't have the right permissions for: " + constants.FoxRootPath)
		color.Red("               You probably won't be able download nor run anything. Or will need to use `sudo` constantly")
		color.Cyan("              Run: sudo chmod -R 777 " + constants.FoxRootPath)
		danger++
	} else {
		color.White("                ✅ Correct file permissions for: " + constants.FoxRootPath)
	}

	if warnings+danger == 0 {
		color.Green("\n       If you're having env problems I feel bad for you son.\n       I got 99 problems, but your env ain't one. ✅")
		spin := spinner.New([]string{
			"( •_•)",
			"( •_•)>⌐■-■",
			"( •>⌐■-■",
			"(⌐■_■)",
		}, 750*time.Millisecond, spinner.WithWriter(os.Stderr))
		_ = spin.Color("bold", "fgGreen")
		spin.Start()
		time.Sleep(3000 * time.Millisecond)
		spin.Stop()
		color.Green("(⌐■_■)\n")
	}

	if warnings > 0 {
		plural := ""
		if warnings > 1 {
			plural = "s"
		}
		color.Yellow("💉 Found %d warning%s", warnings, plural)
	}

	if danger > 0 {
		plural := ""
		if danger > 1 {
			plural = "s"
		}
		color.Red("❌ Found %d potential problem%s", danger, plural)
	}

	fmt.Println()
}
