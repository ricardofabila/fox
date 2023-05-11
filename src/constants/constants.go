package constants

import "runtime"

const GlobalRemote = "https://raw.githubusercontent.com/ricardofabila/fox-packages/main/packages.yaml"

const FoxRootPath = "/usr/local/Fox/"
const FoxBinPath = FoxRootPath + "bin/"
const FoxVersionPath = "/usr/local/Fox/version"

const ConfigFilePath = "/.fox/config.yaml"
const ConfigDirectoryPath = "/.fox"

const RepositoriesFilePath = "/.fox/repositories.yaml"

const CacheFilePath = FoxRootPath + "cache.yaml"

const Binary = "binary"
const Script = "script"

// GitHubRateLimit User-to-server requests are limited to 5,000 requests per hour and per authenticated user.
// All requests from OAuth applications authorized by a user or a personal access token owned
// by the user, and requests authenticated with any of the user's authentication credentials,
// share the same quota of 5,000 requests per hour for that user.
// https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting
const GitHubRateLimit = 5000

var DoNotShow = []string{"gh", "fox"}

var TarExtensions = []string{".tar", ".tar.gz", ".tb2", ".tbz", ".tbz2", ".tgz", ".tlz", ".txz", ".tZ"}

var ZIPExtensions = []string{".zip"}

var Clocks = []string{
	" 🕐 '･ˎ--ˎ^^- ",
	" 🕜 ~･ˌ--ˌ^^- ",
	" 🕑 _･,--,^^- ",
	" 🕝 ｡･.--.^^- ",
	" 🕒 ~･ˎ--ˎ^^- ",
	" 🕞 '･ˌ--ˌ^^- ",
	" 🕓 '･,--,^^- ",
	" 🕟 ~･.--.^^- ",
	" 🕔 _･ˎ--ˎ^^- ",
	" 🕠 ｡･ˌ--ˌ^^- ",
	" 🕕 ~･,--,^^- ",
	" 🕡 '･.--.^^- ",
	" 🕖 '･ˎ--ˎ^^- ",
	" 🕢 ~･ˌ--ˌ^^- ",
	" 🕗 _･,--,^^- ",
	" 🕣 ｡･.--.^^- ",
	" 🕘 ~･ˎ--ˎ^^- ",
	" 🕤 '･ˌ--ˌ^^- ",
	" 🕙 '･,--,^^- ",
	" 🕥 ~･.--.^^- ",
	" 🕚 _･ˎ--ˎ^^- ",
	" 🕦 ｡･ˌ--ˌ^^- ",
	" 🕛 ~･,--,^^- ",
	" 🕧 '･.--.^^- "}

var MacOS = []string{
	"macintosh",
	runtime.GOARCH + "macintosh",
	"macintosh",
	"apple" + "darwin" + runtime.GOARCH,
	"apple" + runtime.GOARCH + "darwin",
	"apple" + "darwin",
	"darwin" + "apple",
	"darwin",
	"macos" + runtime.GOARCH,
	runtime.GOARCH + "macos",
	"macos",
	"mac",
	"mac" + runtime.GOARCH,
	runtime.GOARCH + "mac",
	"osx" + runtime.GOARCH,
	runtime.GOARCH + "osx",
	"osx",
}

var Linux = []string{
	"8664" + "linux",
	"linux" + "8664",
	"linux64" + "static",
	"linux64",
	"amd64" + "linux",
	"linux" + "amd64",
	"8664" + "linux",
}
