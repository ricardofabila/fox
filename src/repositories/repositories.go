package repositories

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/installations"
	"github.com/ricardofabila/fox/src/types"
	"github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

func LoadPackagesFromCache(repositoriesConfig repositories.Config, userConfig types.UserConfig, forceUpdate bool) ([]repositories.Package, error) {
	if userConfig.AutoUpdate || forceUpdate {
		err := UpdatePackagesCache(repositoriesConfig, forceUpdate)
		if err != nil {
			return nil, err
		}
	}

	var repositoriesStruct struct {
		Packages []repositories.Package `yaml:"packages"`
	}
	repositoriesStruct.Packages = []repositories.Package{}
	data, err := os.ReadFile(constants.CacheFilePath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &repositoriesStruct)
	if err != nil {
		return nil, err
	}

	return repositoriesStruct.Packages, nil
}

// writeDefaultConfig Creates a default configuration file
func writeDefaultCacheFile() error {
	err := utils.CreateFileIfNotExists(constants.CacheFilePath)
	if err != nil {
		return err
	}

	var repositoriesStruct struct {
		Packages []repositories.Package `yaml:"packages"`
	}
	repositoriesStruct.Packages = []repositories.Package{}
	data, err := yaml.Marshal(&repositoriesStruct)
	if err != nil {
		return err
	}

	return os.WriteFile(constants.CacheFilePath, data, 0666)
}

func UpdatePackagesCache(repositoriesConfig repositories.Config, force bool) error {
	// create cache file if it doesn't exist
	if !utils.FileExists(constants.CacheFilePath) {
		err := writeDefaultCacheFile()
		return err
	}

	// cache is still fresh no need to rebuild
	if utils.FileExists(constants.CacheFilePath) && !force {
		stats, err := os.Stat(constants.CacheFilePath)
		if err != nil {
			return err
		}

		thirtyMinutes := time.Minute * 30
		if (time.Now().UnixMilli() - stats.ModTime().UnixMilli()) < thirtyMinutes.Milliseconds() {
			return nil
		}
	}

	packages, err := LoadPackages(repositoriesConfig, true)
	if err != nil {
		return err
	}

	// Load the users custom packages
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(home + constants.RepositoriesFilePath)
	if err != nil {
		return err
	}

	var customPackagesStruct repositories.Config
	err = yaml.Unmarshal(data, &customPackagesStruct)
	if err != nil {
		return err
	}

	// Hard coding some hidden packages to be able to upgrade them
	customPackagesStruct.Packages = append(customPackagesStruct.Packages, repositories.HardcodedPackages...)
	customPackages := LoadPackagesFromRepository(customPackagesStruct.Packages, true)

	// save the packages file
	var repositoriesStruct struct {
		Packages []repositories.Package `yaml:"packages"`
	}

	packages = append(packages, customPackages...)

	for i, p := range packages {
		installs := installations.FindInstallations(p.ExecutableName)
		if len(installs) > 0 {
			installed := lo.Map[types.Installation, string](installs, func(i types.Installation, _ int) string {
				return i.Version
			})
			aliases := lo.Map[types.Installation, string](installs, func(i types.Installation, _ int) string {
				return i.Alias
			})
			aliases = lo.Filter(aliases, func(a string, _ int) bool {
				return a != ""
			})
			// can't naively modify with range
			(&packages[i]).InstalledVersions = installed
			(&packages[i]).Aliases = aliases
		}

		// check for conflicts with packages already installed by other sources
		conflict := utils.IsOnPath(p.ExecutableName)
		if len(installs) == 0 && conflict != "" {
			(&packages[i]).Conflicts = conflict
		}
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	repositoriesStruct.Packages = packages
	data, err = yaml.Marshal(&repositoriesStruct)
	if err != nil {
		return err
	}

	err = os.WriteFile(constants.CacheFilePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func LoadPackages(repositoriesConfig repositories.Config, verbose bool) ([]repositories.Package, error) {
	// add the global remote. A curated list of packages people can submit packages into to share with the world.
	globalRemote := repositories.Remote{
		URL:  constants.GlobalRemote,
		Type: "open",
	}

	repositoriesConfig.Remotes = append(repositoriesConfig.Remotes, globalRemote)

	var fetchedPackages []repositories.Package
	for _, remote := range repositoriesConfig.Remotes {
		fetched, err := LoadPackagesFromRemote(remote, verbose)
		if err != nil {
			if verbose {
				color.Red("Error fetching repositories from remote: " + remote.URL)
			}

			continue
		}

		fetchedPackages = append(fetchedPackages, fetched...)
	}

	return fetchedPackages, nil
}

func LoadPackagesFromRemote(remote repositories.Remote, verbose bool) ([]repositories.Package, error) {
	var fetchedPackages []repositories.Package

	var b []byte
	switch remote.Type {
	case "github":
		// gh api "repos/ricardofabila/fox/contents/configPackages.yaml" -- jq ".download_url"
		downloadURL, err := utils.ExecuteCommandAndGetOutput("gh", []string{"api", remote.URL, "--jq", ".download_url"}...)
		downloadURL = strings.TrimSpace(downloadURL)
		if err != nil {
			return fetchedPackages, utils.PrintAndReturnError(err.Error())
		}

		b, err = utils.GetFromAPI(downloadURL)
		if err != nil {
			return fetchedPackages, utils.PrintAndReturnError(err.Error())
		}
	case "open":
		temp, err := utils.GetFromAPI(remote.URL)
		b = temp
		if err != nil {
			return fetchedPackages, utils.PrintAndReturnError(err.Error())
		}
	default:
		return fetchedPackages, fmt.Errorf("error, the remote type '" + remote.Type + "' is not supported. Only 'github' and 'open' are valid values.")
	}

	var repositoriesStruct struct {
		Packages repositories.ConfigPackages `yaml:"packages"`
	}
	err := yaml.Unmarshal(b, &repositoriesStruct)
	if err != nil {
		return fetchedPackages, err
	}
	configPackages := repositoriesStruct.Packages
	if len(configPackages) == 0 {
		warn := fmt.Sprintf("Error. The package remote '" + remote.URL + "' has no packages defined")
		if verbose {
			color.Yellow(warn)
		}
	}

	var executableNames []string
	for _, configPackage := range configPackages {
		if !strings.EqualFold(configPackage.Type, constants.Binary) && !strings.EqualFold(configPackage.Type, constants.Script) {
			warn := fmt.Sprintf("Error. The package '" + configPackage.Path + "' has an unsupported type: '" + configPackage.Type + "'. Only 'script' and 'binary' are valid values.")
			if verbose {
				color.Yellow(warn)
			}
			return fetchedPackages, fmt.Errorf(warn)
		}
		executableNames = append(executableNames, configPackage.ExecutableName)
	}

	// TODO: allow for duplicates, prompt the user which package to install
	duplicates := utils.DuplicateStrings(executableNames)
	if len(duplicates) > 0 {
		return fetchedPackages, fmt.Errorf("the repos list contains duplicate repos with the same executableName [%s]", strings.Join(duplicates, ", "))
	}

	fetchedPackages = append(fetchedPackages, LoadPackagesFromRepository(configPackages, verbose)...)

	return fetchedPackages, nil
}

func LoadPackagesFromRepository(configPackages repositories.ConfigPackages, verbose bool) []repositories.Package {
	var packages []repositories.Package
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(configPackages))
	// Just add some delay to not get rate limited

	delay := int(math.Round(float64(len(configPackages) / constants.GitHubRateLimit)))
	// Do these in go routines
	for _, configPackage := range configPackages {
		go func(configPackage repositories.ConfigPackage) {
			if delay > 0 {
				time.Sleep(time.Millisecond * time.Duration(delay))
			}

			out, err := utils.ExecuteCommandAndGetOutput("gh", []string{"repo", "view", configPackage.Path, "--json", "name,description,updatedAt,primaryLanguage,url,nameWithOwner"}...)
			if err != nil {
				color.Red("%s", out)
				color.Red("%s", err)
				waitGroup.Done()
				return
			}

			var fetchedPackage repositories.Package
			err = json.Unmarshal([]byte(out), &fetchedPackage)
			if err != nil {
				color.Red("%s", out)
				color.Red("%s", err)
				waitGroup.Done()
				return
			}

			fetchedPackage.ExecutableName = configPackage.ExecutableName
			fetchedPackage.Type = configPackage.Type
			fetchedPackage.DependsOn = configPackage.DependsOn
			err = fetchedPackage.SetLatestVersion(verbose)
			if err != nil {
				waitGroup.Done()
				return
			}

			packages = append(packages, fetchedPackage)
			waitGroup.Done()
		}(configPackage)
	}

	waitGroup.Wait()

	return packages
}
