package installations

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/types"
	repositoriesTypes "github.com/ricardofabila/fox/src/types/repositories"
	"github.com/ricardofabila/fox/src/utils"
)

var installationsPath = constants.FoxRootPath + "installations.yaml"

func LoadInstallations() types.Installations {
	file, err := os.OpenFile(installationsPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		color.Red("Error writing installations file at "+installationsPath+": %s", err.Error())
		os.Exit(1)
	}

	err = file.Close()
	if err != nil {
		color.Red("Error closing installations file at "+installationsPath+"installations.yaml: %s", err.Error())
		os.Exit(1)
	}

	data, err := os.ReadFile(installationsPath)
	if err != nil {
		color.Red("Error reading installations file at "+installationsPath+"installations.yaml: %s", err.Error())
		os.Exit(1)
	}
	if len(data) == 0 {
		SaveInstallations(types.Installations{})
	}

	data, err = os.ReadFile(installationsPath)
	if err != nil {
		color.Red("Error reading installations file at "+installationsPath+"installations.yaml: %s", err.Error())
		os.Exit(1)
	}

	var installations types.Installations
	err = yaml.Unmarshal(data, &installations)
	if err != nil {
		color.Red("Error reading installations file at "+installationsPath+"installations.yaml: %s", err.Error())
		os.Exit(1)
	}

	return installations
}

func SaveInstallations(installations types.Installations) {
	data, err := yaml.Marshal(&installations)
	if err != nil {
		color.Red("Error marshalling installations.yaml: %s", err.Error())
		os.Exit(1)
	}

	err = os.WriteFile(installationsPath, data, 0666)
	if err != nil {
		color.Red("Error writing installations file at "+installationsPath+": %s", err.Error())
		os.Exit(1)
	}
}

func SaveInstallation(installation types.Installation) {
	installs := LoadInstallations()
	// remove duplicate if exist
	cleanInstalls := lo.Filter[types.Installation](installs.Installations, func(ins types.Installation, _ int) bool {
		return ins.RealName != installation.RealName
	})

	SaveInstallations(types.Installations{Installations: append(cleanInstalls, installation)})
}

func DeleteInstallation(installation types.Installation) {
	installs := LoadInstallations()
	cleanInstalls := lo.Filter[types.Installation](installs.Installations, func(ins types.Installation, _ int) bool {
		return ins.RealName != installation.RealName
	})

	SaveInstallations(types.Installations{Installations: cleanInstalls})
}

func FindPackage(packages []repositoriesTypes.Package, installation types.Installation) *repositoriesTypes.Package {
	install, found := lo.Find(packages, func(pkg repositoriesTypes.Package) bool {
		return pkg.ExecutableName == installation.ExecutableName
	})

	if found {
		return &install
	}

	return nil
}

func FindInstallation(executableNameOrAlias string) *types.Installation {
	installs := LoadInstallations()
	install, found := lo.Find(installs.Installations, func(installation types.Installation) bool {
		return executableNameOrAlias == installation.RealName
	})

	if found {
		return &install
	}

	return nil
}

func FindInstallations(executableName string) []types.Installation {
	installs := LoadInstallations()
	installations := lo.Filter[types.Installation](installs.Installations, func(t types.Installation, _ int) bool {
		return t.ExecutableName == executableName
	})

	return installations
}

func NotifyNewVersions(packages []repositoriesTypes.Package, installations types.Installations) {
	color.Magenta(" Checking for available package updates: \n")
	upgradable := GetUpgradable(packages, installations)
	if len(upgradable) == 0 {
		color.Magenta(" No packages need to be upgraded ~(‾▿‾)~")
		return
	}
	for _, u := range upgradable {
		color.Yellow(" " + u.ExecutableName + " has a newer version: " + u.LatestVersion)
		color.Yellow("    Your version is: [" + strings.Join(u.InstalledVersions, ", ") + "]")
		fmt.Println()
	}
	color.Yellow(" run 'fox upgrade' to upgrade all packages")
	fmt.Println()
}

func GetUpgradable(packages []repositoriesTypes.Package, installations types.Installations) []repositoriesTypes.Package {
	var upgradable []repositoriesTypes.Package

	canBeUpgraded := lo.Filter(installations.Installations, func(i types.Installation, _ int) bool {
		return i.Alias == ""
	})

	for _, installation := range canBeUpgraded {
		pkg, found := lo.Find(packages, func(pkg repositoriesTypes.Package) bool {
			return pkg.NameWithOwner == installation.Package
		})

		if !found {
			continue
		}

		if !strings.Contains(strings.TrimSpace(pkg.LatestVersion), strings.TrimSpace(installation.Version)) {
			upgradable = append(upgradable, pkg)
		}
	}

	return upgradable
}

func InstallPackage(availablePackages []repositoriesTypes.Package, executableName, alias string, interactive bool, userConfig types.UserConfig, installFox, force bool) error {
	pkgParam := strings.Split(executableName, "@")
	pkgName := strings.TrimSpace(pkgParam[0])
	alias = strings.TrimSpace(alias)
	version := ""

	if len(pkgParam) != 1 && len(pkgParam) != 2 {
		return fmt.Errorf("Error. The package name must follow the format: <package_name>@<version>. Given: " + executableName)
	}

	if len(pkgParam) == 2 {
		version = pkgParam[1]
	} else {
		version = "latest"
	}

	if !installFox && (strings.EqualFold(pkgName, "fox") || strings.EqualFold(alias, "fox")) {
		return fmt.Errorf("error I cannot install a package with the name of fox.\nThat would kill me o(╥﹏╥)o\nIf you want to upgrade fox run 'fox upgrade fox'")
	}

	// check for conflicts with packages already installed by other sources
	conflictAlias := utils.IsOnPath(alias)
	if conflictAlias != "" && alias != "" && !installFox {
		if FindInstallation(alias) == nil {
			return fmt.Errorf("The package you want to install with that alias conflicts with: " + conflictAlias)
		}
	}

	conflictPkgName := utils.IsOnPath(pkgName)
	if conflictPkgName != "" && !installFox {
		if FindInstallation(pkgName) == nil {
			return fmt.Errorf("The package you want to install conflicts with: " + conflictPkgName)
		}
	}

	if userConfig.NotifyOutdatedVersions && interactive {
		installs := LoadInstallations()
		NotifyNewVersions(availablePackages, installs)
	}

	var pkg *repositoriesTypes.Package
	for _, repo := range availablePackages {
		if repo.ExecutableName == pkgName {
			pkg = &repo
			break
		}
	}

	if pkg == nil {
		return fmt.Errorf(fmt.Sprintf("Could not find the package '%s'. Try running 'fox update' first.", pkgName))
	}

	// package might already be at the latest version
	if alias == "" {
		existingInstallation := FindInstallation(pkgName)
		if existingInstallation != nil {
			if strings.Contains(strings.TrimSpace(pkg.LatestVersion), strings.TrimSpace(existingInstallation.Version)) {
				if !force {
					color.Green(" The package " + pkgName + " is already at the latest version: " + existingInstallation.Version)
					return nil
				}
			}
		}
	}

	releases, err := pkg.GetReleases()
	if err != nil {
		return err
	}

	if len(releases) == 0 {
		return fmt.Errorf("Error. The package " + pkgName + " has no releases")
	}

	var releaseToInstall *repositoriesTypes.Release
	if version == "latest" {
		releaseToInstall = &releases[0]
	} else {
		for _, release := range releases {
			if strings.EqualFold(release.Name, version) {
				releaseToInstall = &release
				break
			}
		}
	}

	if releaseToInstall == nil {
		return fmt.Errorf("error package version not found")
	}

	color.Blue(" Installing: %s@%s", pkg.ExecutableName, version)
	releaseToInstall.Assets = lo.Filter(releaseToInstall.Assets, func(x repositoriesTypes.Asset, _ int) bool {
		if strings.Contains(x.Name, "windows") {
			return false
		}

		if strings.HasSuffix(x.Name, ".xz") {
			return false
		}

		if strings.HasSuffix(x.Name, ".gz") && !strings.HasSuffix(x.Name, ".tar.gz") {
			return false
		}

		// pre-filtering assets mean for a different operating system
		if strings.Contains(strings.ToLower(runtime.GOOS), "darwin") {
			if strings.Contains(x.Name, "linux") || strings.Contains(x.Name, "windows") {
				return false
			}
		}

		if strings.Contains(strings.ToLower(runtime.GOOS), "linux") {
			if strings.Contains(x.Name, "darwin") || strings.Contains(x.Name, "osx") || strings.Contains(x.Name, "windows") {
				return false
			}
		}

		return true
	})

	assetName, err := DownloadAsset(*pkg, *releaseToInstall, interactive)
	if err != nil {
		return err
	}

	alias = lo.Ternary(alias == "", pkg.ExecutableName, alias)

	if version != "latest" {
		// check if there is no previous installation, we can avoid the @
		if FindInstallation(alias) != nil {
			alias += "@" + releaseToInstall.Tag
		}
	}

	err = MoveAssetToBin(assetName, alias)
	if err != nil {
		return err
	}

	if len(pkg.DependsOn) > 0 {
		color.Yellow(" Warning: '%s' depends on:\n   [%s]\n   make sure you have those installed.", pkg.ExecutableName, strings.Join(pkg.DependsOn, ", "))
	}

	if pkg.NameWithOwner == "ricardofabila/fox" {
		return nil
	}

	color.Green(" 🦊 Installed: %s@%s as %s", pkg.ExecutableName, version, alias)
	// save the installation
	install := types.Installation{
		Timestamp:      time.Now().UnixMilli(),
		Package:        pkg.NameWithOwner,
		ExecutableName: pkg.ExecutableName,
		RealName:       alias,
		Version:        releaseToInstall.Tag,
	}
	if alias != pkg.ExecutableName {
		install.Alias = alias
	}

	SaveInstallation(install)

	return nil
}

func MoveAssetToBin(assetName, alias string) error {
	installationPath := constants.FoxBinPath + alias
	err := utils.RemoveFile(installationPath)
	if err != nil {
		return err
	}

	err = utils.MoveFile("./"+assetName, installationPath)
	if err != nil {
		return err
	}

	err = utils.MakeFileExecutable(installationPath)
	if err != nil {
		return err
	}

	return nil
}

func DownloadAsset(pkg repositoriesTypes.Package, release repositoriesTypes.Release, interactive bool) (string, error) {
	if pkg.Type == constants.Script {
		assetsNames := lo.Map[repositoriesTypes.Asset, string](release.Assets, func(x repositoriesTypes.Asset, _ int) string {
			return x.Name
		})
		// use the pkg.ExecutableName as the file to search for.
		// In case the release has other files like checksums or something.
		assetToSearchFor := pkg.ExecutableName
		ranks := fuzzy.RankFindNormalizedFold(assetToSearchFor, assetsNames)

		// if no match, use the zip source code that every release has and extract the script from there
		if len(ranks) == 0 {
			_, err := utils.ExecuteCommandAndGetOutput("gh", []string{"release", "download", "--repo", pkg.NameWithOwner, release.Tag, "--archive", "zip", "--dir", "."}...)
			if err != nil {
				return "", err
			}

			files, err := ioutil.ReadDir(".")
			if err != nil {
				return "", err
			}

			var filesInCurrentDirectory []string
			for _, file := range files {
				filesInCurrentDirectory = append(filesInCurrentDirectory, file.Name())
			}

			zipRanks := fuzzy.RankFindNormalizedFold(pkg.Name+".zip", filesInCurrentDirectory)
			if len(zipRanks) == 0 {
				return "", fmt.Errorf("error finding zip ball")
			}
			best := lo.MaxBy[fuzzy.Rank](zipRanks, func(rank, max fuzzy.Rank) bool {
				return rank.Distance > max.Distance
			})
			assetName, err := ExtractAsset(best.Target, pkg.ExecutableName)
			if err != nil {
				return "", err
			}

			return assetName, nil
		}

		best := lo.MaxBy[fuzzy.Rank](ranks, func(rank, max fuzzy.Rank) bool {
			return rank.Distance > max.Distance
		})
		assetToDownload := &release.Assets[best.OriginalIndex]

		if assetToDownload == nil {
			return "", fmt.Errorf("Error. Found no installable asset for the given release: " + pkg.ExecutableName)
		}

		color.Magenta(" Fetching the asset " + assetToDownload.Name + " of size " + utils.ByteCountIEC(int64(assetToDownload.Size)))
		err := assetToDownload.DownloadAsset(pkg.NameWithOwner)
		if err != nil {
			return "", err
		}

		if utils.FileHasTarExtension(assetToDownload.Name) || utils.FileHasZIPExtension(assetToDownload.Name) {
			assetName, err := ExtractAsset(assetToDownload.Name, pkg.ExecutableName)
			return assetName, err
		}

		return assetToDownload.Name, nil
	}

	if pkg.Type == constants.Binary {
		assetToDownload, err := GetAssetToDownloadForBinary(pkg, release.Assets, interactive)
		if err != nil {
			return "", err
		}

		color.Magenta(" Fetching the asset " + assetToDownload.Name + " of size " + utils.ByteCountIEC(int64(assetToDownload.Size)))
		err = assetToDownload.DownloadAsset(pkg.NameWithOwner)
		if err != nil {
			return "", err
		}

		if utils.FileHasTarExtension(assetToDownload.Name) || utils.FileHasZIPExtension(assetToDownload.Name) {
			assetName, err := ExtractAsset(assetToDownload.Name, pkg.ExecutableName)
			return assetName, err
		}

		return assetToDownload.Name, nil
	}

	return "", fmt.Errorf("Error. The following package type is not valid: " + pkg.Type)
}

func GetAssetToDownloadForBinary(pkg repositoriesTypes.Package, assets []repositoriesTypes.Asset, interactive bool) (*repositoriesTypes.Asset, error) {
	if len(assets) == 0 {
		return nil, fmt.Errorf("Error. Found no assets for the given release: " + pkg.ExecutableName)
	}

	var assetToDownload *repositoriesTypes.Asset
	assetsNames := lo.Map[repositoriesTypes.Asset, string](assets, func(x repositoriesTypes.Asset, _ int) string {
		return x.Name
	})

	usersRuntime := runtime.GOOS + " " + runtime.GOARCH
	assetToSearchFor := usersRuntime
	ranks := fuzzy.RankFindNormalizedFold(assetToSearchFor, assetsNames)

	// Deal with macOS having different names and architectures
	// (╯°□°）╯︵ ┻━┻
	if len(ranks) == 0 {
		// use macos in the case it is darwin, as a lot of packages use that name
		if strings.Contains(strings.ToLower(usersRuntime), "darwin") {
			for _, r := range constants.MacOS {
				assetToSearchFor = r
				ranks = fuzzy.RankFindNormalizedFold(assetToSearchFor, assetsNames)

				if len(ranks) > 0 {
					break
				}
			}
		}

		// Deal with linux having different names and architectures
		// (╯°□°）╯︵ ┻━┻
		if strings.Contains(strings.ToLower(usersRuntime), "linux") {
			// just check for linux + x86_64
			if strings.Contains(strings.ToLower(runtime.GOARCH), "386") || strings.Contains(strings.ToLower(runtime.GOARCH), "amd64") {
				for _, r := range constants.Linux {
					assetToSearchFor = r
					ranks = fuzzy.RankFindNormalizedFold(assetToSearchFor, assetsNames)

					if len(ranks) > 0 {
						break
					}
				}
			}
		}
	}

	if len(ranks) == 0 {
		assetToSearchFor = usersRuntime
		return nil, fmt.Errorf("Error. Found no assets that match your OS and Architecture: " + assetToSearchFor)
	}

	best := lo.MaxBy[fuzzy.Rank](ranks, func(rank, max fuzzy.Rank) bool {
		return rank.Distance > max.Distance
	})
	assetToDownload = &assets[best.OriginalIndex]

	if assetToDownload == nil {
		return nil, fmt.Errorf("Error. Found no installable asset for the given release: " + pkg.ExecutableName)
	}

	if interactive {
		color.Magenta(" Found the asset: " + assetToDownload.Name)
		prompt := promptui.Select{
			Label: " Proceed with installation?",
			Items: []string{"Yes", "No"},
		}

		_, result, err := prompt.Run()

		// most likely a Control+C
		if err != nil {
			color.Green(" (Ͼ˳Ͽ)..!!! Aborting installation.")
			os.Exit(1)
		}

		if result == "No" {
			color.Green(" (Ͼ˳Ͽ)..!!! Aborting installation.")
			os.Exit(1)
		}

		color.Green(" (＾▽＾) Continuing with your installation!")
	}

	return assetToDownload, nil
}

func ExtractAsset(fileName, executableName string) (string, error) {
	// if the download was a tar or zip file, process it
	if !utils.FileHasTarExtension(fileName) && !utils.FileHasZIPExtension(fileName) {
		return "", fmt.Errorf("the given file was not in a  valid compressed format: " + fileName)
	}

	directoryName := "./fox-temp-" + executableName
	err := utils.RemoveDirectory(directoryName)
	if err != nil {
		return "", err
	}

	utils.CreateDirectoryIfNotExists(directoryName)

	if utils.FileHasTarExtension(fileName) {
		err = utils.ExtractTAR("./"+fileName, directoryName)
		if err != nil {
			return "", err
		}
	}

	if utils.FileHasZIPExtension(fileName) {
		err = utils.ExtractZIP(fileName, directoryName)
		if err != nil {
			return "", err
		}
	}

	// walk the directory until we find the file with the name of the executable
	desiredFile := ""
	err = filepath.Walk(directoryName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.EqualFold(executableName, info.Name()) && !info.IsDir() {
			desiredFile = path
			return nil
		}

		// some bundlers use the name of the folder as the name of the binary, so match without extension
		if strings.EqualFold(utils.ZIPWithoutExtension(fileName), info.Name()) && !info.IsDir() {
			desiredFile = path
			return nil
		}

		if strings.EqualFold(utils.TarWithoutExtension(fileName), info.Name()) && !info.IsDir() {
			desiredFile = path
			return nil
		}

		// in the case for scripts, the file might have an extension or something extra. eg: my-script.js
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(executableName)) && !info.IsDir() {
			// some scripts have man pages that end en a .number
			match, _ := regexp.MatchString("\\.\\d+", info.Name())
			if !match {
				desiredFile = path
			}

			return nil
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if desiredFile == "" {
		return "", fmt.Errorf("Error. Found no installable asset for the given releas of: " + executableName)
	}

	// move file out of the directory
	err = utils.MoveFile(desiredFile, "./"+executableName)
	if err != nil {
		return "", err
	}

	err = utils.RemoveFile("./" + fileName)
	if err != nil {
		return "", err
	}

	// remove dangling directory, we don't need it
	err = utils.RemoveDirectory(directoryName)
	if err != nil {
		return "", err
	}

	return "./" + executableName, nil
}
