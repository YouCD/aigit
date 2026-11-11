package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/YouCD/aigit/llm"
)

// Version is set at build time with -ldflags "-X github.com/YouCD/aigit/cmd.Version=...".
var Version = "dev"

// updateNotice receives the newer-release tag (or "") from the background
// check started in PersistentPreRun.
var updateNotice <-chan string

var rootCmd = &cobra.Command{
	Use:   "aigit",
	Short: "Generate git commit message including title and body",
	Long:  `AI Git Commi streamlines the git commit process by automatically generating meaningful and standardized commit messages.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		if debug {
			llm.Debug = true
		}
		updateNotice = startUpdateCheck(Version)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		select {
		case latest := <-updateNotice:
			if latest == "" {
				return
			}
			fmt.Printf("\n%s %s → %s\n",
				color.YellowString("A new release of aigit is available:"),
				strings.TrimPrefix(Version, "v"), strings.TrimPrefix(latest, "v"))
			prompt := promptui.Select{
				Label: "Upgrade now",
				Items: []string{"Yes", "No"},
				Size:  2,
			}
			// Non-interactive runs fail the prompt; skip silently.
			choice, _, err := prompt.Run()
			if err != nil || choice != 0 {
				return
			}
			fmt.Println("⬆️  Upgrading aigit to", latest, "...")
			if err := selfUpgrade(latest); err != nil {
				color.Red("Upgrade failed: %v", err)
				return
			}
			color.Green("✅ aigit upgraded to %s", latest)
		// Slightly longer than the update check's HTTP timeout so the
		// once-per-day refresh can finish and persist its state file;
		// cached runs deliver instantly.
		case <-time.After(3 * time.Second):
		}
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		exit(ExitCommand)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(versionCmd)
}
