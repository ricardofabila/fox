package cmd

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add remote and package entries to your repositories.yaml file",
	Long: `
	Add remote and package entries to your repositories.yaml file

	Add a remote:
	$ fox add remote --url "your.url.com/path-to-a-packages-yaml-file" --type "open"

	Add a package:
	$ fox add package --path="OWNER/REPO" --executableName="a-name" --type="script" --dependsOn="bash,curl"
`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
