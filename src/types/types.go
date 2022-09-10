package types

import (
	"github.com/samber/lo"

	"github.com/ricardofabila/fox/src/constants"
)

// File to avoid circle imports

type UserConfig struct {
	AutoUpdate             bool `yaml:"autoUpdate"`
	NotifyOutdatedVersions bool `yaml:"notifyOutdatedVersions"`
}

type Installations struct {
	Installations []Installation `yaml:"installations"`
}

type Installation struct {
	Timestamp      int64  `yaml:"timestamp"`
	Package        string `yaml:"package"`
	ExecutableName string `yaml:"executableName"`
	Alias          string `yaml:"alias"`
	RealName       string `yaml:"realName"`
	Version        string `yaml:"version"`
}

func (i *Installation) IsVisible() bool {
	return !lo.Contains(constants.DoNotShow, i.ExecutableName)
}
