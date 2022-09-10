package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rivo/tview"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/repositories"
	repositoriesTypes "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

type ListFlags struct {
	print      bool
	upgradable bool
}

var listFlags = ListFlags{
	print:      false,
	upgradable: false,
}

// listCmd represents the repos command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "See the repositories available",
	Long:  `See the repositories available to install packages from.`,
	Example: `

	List the packages available in interactive mode.
	$ fox list

	Print the packages available to stdout.
	$ fox list --print

	Print the packages that can be upgraded to a new version
	$ fox list --upgradable
`,
	Run: func(cmd *cobra.Command, args []string) {
		color.Blue(" ⏳ Loading packages list...")
		packages, err := repositories.LoadPackagesFromCache(repositoriesConfig, userConfig, false)
		utils.CheckErr(err, cmd)

		if userConfig.NotifyOutdatedVersions || listFlags.upgradable {
			err = checkForNewFoxVersion()
			utils.CheckErr(err, nil)
			installs := installations.LoadInstallations()
			installations.NotifyNewVersions(packages, installs)

			if listFlags.upgradable {
				return
			}
		}

		packages = lo.Filter(packages, func(p repositoriesTypes.Package, _ int) bool {
			return p.IsVisible()
		})

		if listFlags.print {
			color.Blue("Available packages:\n\n")
			for _, p := range packages {
				detailsText := tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
				detailsText.SetText(renderDetails(p, func(s string) {}))
				fmt.Println(detailsText.GetText(true))
				fmt.Println("    ----------------------------------------")
				fmt.Println()
			}
			return
		}

		err = interactiveRender(packages, "Available packages")
		utils.CheckErr(err, cmd)
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listFlags.print, "print", "p", false, "Outputs the list of available packages to stdout")
	listCmd.Flags().BoolVarP(&listFlags.upgradable, "upgradable", "u", false, "Outputs the list of available packages than can be upgraded to a new version")
	rootCmd.AddCommand(listCmd)
}

func renderDetails(recordToDisplay repositoriesTypes.Package, cb func(string)) string {
	fullDetails := "[yellow::b]   " + recordToDisplay.Name
	if len(recordToDisplay.InstalledVersions) == 0 {
		fullDetails += "\n\n\t[blue]🦊 Installation:[green] " + "fox i " + recordToDisplay.ExecutableName
	}
	fullDetails += "\n\n [blue::-]   Latest version:[green] " + recordToDisplay.LatestVersion
	description := strings.Join(utils.ChunkString(recordToDisplay.Description, 50), "\n    ")
	fullDetails += fmt.Sprintf("\n\n\t[blue]📄 Description:[white] \n\n    %s", description)
	fullDetails += fmt.Sprintf("\n\n\t[blue]🔗 URL:[white] \n    " + recordToDisplay.URL)
	fullDetails += "\n\n\t[blue]🗨️  Main language:[magenta]  " + recordToDisplay.PrimaryLanguage["name"]

	if len(recordToDisplay.DependsOn) > 0 {
		fullDetails += "\n\n\t[blue]🔒 Depends on:[orange] " + strings.Join(recordToDisplay.DependsOn, ", ")
	}

	fullDetails += "\n\n\t[blue]🗓  Updated at:[white] " + recordToDisplay.UpdatedAt

	if len(recordToDisplay.InstalledVersions) > 0 {
		fullDetails += "\n\n\t[blue]💾 Installed versions:[white] \n    [" + strings.Join(recordToDisplay.InstalledVersions, ", ") + "]"
	}

	if recordToDisplay.Conflicts != "" {
		fullDetails += "\n\n\t[red]💣 Conflicts with:[white] \n    " + recordToDisplay.Conflicts
	}

	cb(fullDetails)
	return fullDetails
}

//gocyclo:ignore
func interactiveRender(packages []repositoriesTypes.Package, title string) error {
	// Don't include hidden in the list
	packages = lo.Filter(packages, func(p repositoriesTypes.Package, _ int) bool {
		return p.IsVisible()
	})

	app := tview.NewApplication()
	menu := tview.NewTextView().
		SetDynamicColors(true).
		SetTextColor(tcell.ColorBlue).
		SetText("[lightgrey:-:-](Esc) to quit, exit search mode\n(Up, Down) to scroll and show details\n(PgUp, PgDn, End, Home) for faster scrolling\n(Ctrl + f or s) to search\n(Enter) to open package URL").
		SetWordWrap(true)

	recordDetails := tview.NewFlex().SetDirection(tview.FlexRow)
	recordDetails.SetBorder(true).SetBorderColor(tcell.ColorMediumOrchid).SetTitle("Details")

	detailsText := tview.NewTextView().SetDynamicColors(true).SetWordWrap(true)
	recordDetails.AddItem(detailsText, 0, 1, false)
	reposList := tview.NewList().ShowSecondaryText(true)
	reposList.SetMainTextColor(tcell.ColorWhite)
	reposList.SetSecondaryTextColor(tcell.ColorLightSlateGrey)

	fillList := func(listOfRecords []repositoriesTypes.Package) {
		for _, r := range listOfRecords {
			if r.Conflicts != "" {
				reposList.AddItem(" × "+r.Name, "   "+utils.Ellipsis(r.Description, 65), rune(0), nil)
				continue
			}

			if len(r.InstalledVersions) > 0 {
				reposList.AddItem(" ▹ "+r.Name, "   "+utils.Ellipsis(r.Description, 65), rune(0), nil)
				continue
			}

			reposList.AddItem(" • "+r.Name, "   "+utils.Ellipsis(r.Description, 65), rune(0), nil)
		}
	}

	fillList(packages)
	searchField := tview.NewInputField()
	searchField.SetLabel(" 🔍 Searching for: ")
	searchField.SetLabelStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow).Bold(true))
	searchField.SetFieldBackgroundColor(tcell.ColorWhiteSmoke)
	searchField.SetFieldTextColor(tcell.ColorBlack)
	searchField.SetPlaceholderStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Bold(true))
	searchField.SetPlaceholderTextColor(tcell.ColorSnow)
	searchField.SetPlaceholder("press (Ctrl + f or s) to search")
	searchField.SetText("")
	searchField.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(reposList)
	})

	reposList.SetChangedFunc(func(index int, command string, _ string, shortcut rune) {
		recordToDisplay := &packages[index]
		if recordToDisplay != nil {
			renderDetails(*recordToDisplay, func(fullDetails string) {
				detailsText.SetText(fullDetails)
			})
		} else {
			recordDetails.Clear()
		}
	})

	reposList.SetSelectedFunc(func(index int, command string, _ string, shortcut rune) {
		selected := packages[index]
		utils.OpenURL(selected.URL)
	})

	// search on the original records and create a clone of them
	originalRecordsToDisplay := make([]repositoriesTypes.Package, len(packages))
	copy(originalRecordsToDisplay, packages)
	searchField.SetChangedFunc(func(searchTerm string) {
		if searchTerm != "" {
			filtered := lo.Filter(originalRecordsToDisplay, func(p repositoriesTypes.Package, _ int) bool {
				return fuzzy.MatchNormalizedFold(searchTerm, p.Name)
			})

			reposList.Clear()
			packages = filtered
			fillList(filtered)

			detailsText.SetText("")
			detailsText.SetText(fmt.Sprintf("%d results", len(filtered)))

			if len(filtered) == 1 {
				renderDetails(filtered[0], func(fullDetails string) {
					detailsText.SetText(fullDetails).SetWordWrap(true).SetWrap(true)
				})
			}
		} else {
			detailsText.SetText("")
			packages = originalRecordsToDisplay
			reposList.Clear()
			fillList(packages)
		}
	})

	// auto select first item to show details for it on start
	if reposList.GetItemCount() > 0 {
		reposList.SetCurrentItem(0)
		recordToDisplay := &packages[reposList.GetCurrentItem()]
		if recordToDisplay != nil {
			renderDetails(*recordToDisplay, func(fullDetails string) {
				detailsText.SetText(fullDetails).SetWordWrap(true).SetWrap(true)
			})
		}
	}

	reposList.SetBorder(true).SetBorderColor(tcell.ColorBlack).SetTitle(title).SetTitleColor(tcell.ColorPurple)
	menu.SetTitle("Menu").SetBorder(true).SetBorderColor(tcell.ColorCornflowerBlue)

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(searchField, 1, 1, false)
	container := tview.NewFlex().
		AddItem(
			tview.NewFlex().
				AddItem(recordDetails, 0, 15, false).
				AddItem(menu, 0, 5, false).
				SetDirection(tview.FlexRow),
			0, 4, false).
		AddItem(reposList, 0, 5, true)
	flex.AddItem(container, 0, 2, true)

	searching := false
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlS || event.Key() == tcell.KeyCtrlF {
			app.SetFocus(searchField)
			searching = true
			return event
		}

		if searching {
			if event.Key() == tcell.KeyESC {
				searching = false
			}

			return event
		}

		// Only allow scrolling and selecting using Enter
		if event.Key() == tcell.KeyEnter ||
			event.Key() == tcell.KeyUp ||
			event.Key() == tcell.KeyDown ||
			event.Key() == tcell.KeyLeft ||
			event.Key() == tcell.KeyRight ||
			event.Key() == tcell.KeyPgUp ||
			event.Key() == tcell.KeyPgDn ||
			event.Key() == tcell.KeyEnd ||
			event.Key() == tcell.KeyHome {
			return event
		}

		return nil
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			if searching {
				searching = false
				app.SetFocus(flex)
				return nil
			}

			app.Stop()
		}

		return event
	})

	if err := app.SetRoot(flex, true).EnableMouse(false).Run(); err != nil {
		return utils.PrintAndReturnError(fmt.Sprintf("Error starting interactive mode: %s \n", err))
	}

	return nil
}
