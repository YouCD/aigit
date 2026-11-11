package cmd

import (
	"fmt"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v", "-v", "-version", "--version"},
	Short:   "Print the version of aigit",
	Long:    "Print the current version of the aigit CLI tool.",
	Run: func(cmd *cobra.Command, args []string) {
		if Version != "dev" {
			fmt.Println(Version)
			return
		}

		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			fmt.Println(info.Main.Version)
			return
		}

		version, err := exec.Command("git", "describe", "--tags").Output()
		if err != nil {
			fmt.Println("dev")
			return
		}
		fmt.Printf("%s\n", strings.TrimSpace(string(version)))
	},
}
