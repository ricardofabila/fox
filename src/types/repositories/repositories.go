package repositories

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/samber/lo"

	"github.com/ricardofabila/fox/src/constants"
	"github.com/ricardofabila/fox/src/utils"
)

type Config struct {
	Remotes  Remotes        `yaml:"remotes"`
	Packages ConfigPackages `yaml:"packages"`
}

type Remotes = []Remote

type Remote struct {
	URL  string `yaml:"url"`
	Type string `yaml:"type"`
}

type ConfigPackages = []ConfigPackage

type ConfigPackage = struct {
	Path           string   `yaml:"path"`
	ExecutableName string   `yaml:"executableName"`
	Type           string   `yaml:"type"`
	DependsOn      []string `yaml:"dependsOn"`
}

type Package struct {
	Description     string            `json:"description"`
	Name            string            `json:"name"`
	NameWithOwner   string            `json:"nameWithOwner"`
	UpdatedAt       string            `json:"updatedAt"`
	URL             string            `json:"url"`
	PrimaryLanguage map[string]string `json:"primaryLanguage"`
	// {"primaryLanguage": {
	// 	"name": "Go"
	// },
	// I fetch these separately
	LatestVersion     string
	DependsOn         []string
	ExecutableName    string
	Type              string
	Releases          []Release
	InstalledVersions []string `yaml:"installedVersions"`
	Aliases           []string `yaml:"aliases"`
	Conflicts         string
}

// Release represents a GitHub release in a repository.
type Release struct {
	Draft      bool `json:"draft,omitempty"`
	Prerelease bool `json:"prerelease,omitempty"`
	// The following fields are not used in CreateRelease or EditRelease:
	ZipballURL string  `json:"zipball_url,omitempty"`
	Tag        string  `json:"tag_name,omitempty"`
	ID         int64   `json:"id,omitempty"`
	Assets     []Asset `json:"assets,omitempty"`
	Name       string  `json:"name,omitempty"`
	Type       string  `json:"type,omitempty"`
	CreatedAt  string  `json:"createdAt,omitempty"`
}

// Asset represents a GitHub release asset in a repository.
type Asset struct {
	Name               string `json:"name,omitempty"`
	ID                 int    `json:"id,omitempty"`
	Tag                string `json:"tag_name,omitempty"`
	Size               int    `json:"size,omitempty"` // bytes
	BrowserDownloadURL string `json:"browser_download_url,omitempty"`
}

var HardcodedPackages = []ConfigPackage{
	{
		Path:           "ricardofabila/fox",
		ExecutableName: "fox",
		Type:           "binary",
	},
	{
		Path:           "cli/cli",
		ExecutableName: "gh",
		Type:           "binary",
	},
}

func (p *Package) IsVisible() bool {
	return !lo.Contains(constants.DoNotShow, p.ExecutableName)
}

func (p *Package) SetLatestVersion(verbose bool) error {
	// https://docs.github.com/en/rest/releases/releases#get-the-latest-release
	// gh api /repos/bishopfox/bf/releases/latest --jq ".name"
	data, err := utils.ExecuteCommandAndGetOutput("gh", []string{"api", "repos/" + p.NameWithOwner + "/releases/latest", "--jq", ".tag_name"}...)

	if err != nil {
		if strings.Contains(data, "Not Found") {
			warn := fmt.Sprintf("Warning! the repo %s has no releases", p.NameWithOwner)
			if verbose {
				color.Yellow(warn)
			}
			return fmt.Errorf(warn)
		}

		color.Red("%s", data)
		return utils.PrintAndReturnError(err.Error())
	}

	p.LatestVersion = strings.TrimSpace(data)

	return nil
}

func (p *Package) GetReleases() ([]Release, error) {
	var releases []Release
	data, err := utils.ExecuteCommandAndGetOutput("gh", []string{"api", "repos/" + p.NameWithOwner + "/releases"}...)
	if err != nil {
		color.Red("%s", data)
		return releases, utils.PrintAndReturnError(err.Error())
	}

	if !utils.IsValidJSON(data) {
		return releases, fmt.Errorf("Error, the response by GitHub was not valid JSON: \n" + data)
	}

	err = json.Unmarshal([]byte(data), &releases)
	if err != nil {
		color.Red("%s", data)
		return releases, utils.PrintAndReturnError(err.Error())
	}

	// filtering our drafts
	releases = lo.Filter(releases, func(r Release, _ int) bool {
		return !r.Draft && !r.Prerelease
	})

	p.Releases = releases
	return releases, nil
}

func (asset *Asset) DownloadAsset(repo string) error {
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

	// unknown flag: --clobber even-though is on the docs https://cli.github.com/manual/gh_release_download 🤷
	// Delete the file manually
	err := utils.RemoveFile("./" + asset.Name)
	if err != nil {
		return err
	}

	// gh release download --repo bishopfox/bf --pattern fox_darwin_amd64_v1 v1.0.0 --dir . --clobber
	// maybe can also use api call directly since ge saved the id
	// curl -vLJO -H 'Authorization: token my_access_token' 'https://api.github.com/repos/:owner/:repo/releases/assets/:id'
	// we get the built url from the asset itself
	// "apiUrl": "https://api.github.com/repos/BishopFox/bf/releases/assets/<id>",
	data, err := utils.ExecuteCommandAndGetOutput("gh", []string{"release", "download", "--repo", repo, "--pattern", asset.Name, asset.Tag, "--dir", "."}...)
	if err != nil {
		color.Red(data)
		return utils.PrintAndReturnError(err.Error())
	}

	spin.Stop()
	ended := time.Now().UnixMilli()
	timeItTook := float64(ended-started) / 1000
	color.Green(" Downloaded in %.2f seconds!", timeItTook)

	return nil
}
