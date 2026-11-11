package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/YouCD/aigit/llm"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate git commit message including title and body",
	Long:  `Generate a git commit message using AI. Use --language (-l) flag to quickly switch language without modifying config.`,
	Run: func(cmd *cobra.Command, args []string) {
		yes, _ := cmd.Flags().GetBool("yes")
		language, _ := cmd.Flags().GetString("language")
		timeoutSec, _ := cmd.Flags().GetInt("timeout")

		// Execute git diff --cached command
		diffOutput, err := exec.Command("git", "diff", "--cached").Output()
		if err != nil {
			fmt.Printf("Error getting git diff: %v\n", err)
			exit(ExitGitDiff)
		}

		// If there are no staged changes
		if len(diffOutput) == 0 {
			if yes {
				if err := exec.Command("git", "add", ".").Run(); err != nil {
					fmt.Printf("Error staging changes: %v\n", err)
					exit(ExitGitStage)
				}
				color.Green("All changes staged successfully!")
				diffOutput, err = exec.Command("git", "diff", "--cached").Output()
				if err != nil {
					fmt.Printf("Error getting git diff: %v\n", err)
					exit(ExitGitDiff)
				}
			} else {
				color.Yellow("No staged changes found.")
				stagePrompt := promptui.Select{
					Label: "Would you like to run 'git add .' to stage all changes?",
					Items: []string{"Yes", "No"},
					Size:  2,
				}

				_, stageChoice, err := stagePrompt.Run()
				if err != nil {
					fmt.Printf("Error with prompt: %v\n", err)
					exit(ExitPrompt)
				}

				if stageChoice == "Yes" {
					addCmd := exec.Command("git", "add", ".")
					if err := addCmd.Run(); err != nil {
						fmt.Printf("Error staging changes: %v\n", err)
						exit(ExitGitStage)
					}
					color.Green("All changes staged successfully!")

					diffOutput, err = exec.Command("git", "diff", "--cached").Output()
					if err != nil {
						fmt.Printf("Error getting git diff: %v\n", err)
						exit(ExitGitDiff)
					}
				} else {
					color.Red("No changes staged. Please use 'git add' to stage your changes.")
					exit(ExitUserAbort)
				}
			}
		}

		config := llm.NewConfig()
		if err := config.Load(); err != nil {
			fmt.Printf("Error reading config: %v\n", err)
			exit(ExitConfigLoad)
		}

		// Override language from command line flag
		if language != "" {
			config.Language = language
		}

		// Override timeout from command line flag
		if timeoutSec > 0 {
			config.Timeout = timeoutSec
		}

		if config.CurrentProvider == "" {
			color.Red("No provider configured.")
			color.Red("Run: aigit auth add <api_key> <model> --base-url <url>")
			exit(ExitConfigLoad)
		}
		provider := config.CurrentProvider

		fmt.Println("\n🤖 Generating commit message by", provider)
		var commitMessage string
		for attempt := 1; attempt <= 3; attempt++ {
			commitMessage, err = generateMessage(config, diffOutput)
			if err != nil {
				fmt.Printf("Error generating commit message: %v\n", err)
				exit(ExitLLM)
			}
			if !llm.IsRawCommitJSON(commitMessage) {
				break
			}
			if attempt < 3 {
				color.Yellow("⚠️ Response format incorrect (body must be a string list), retrying (%d/3)...", attempt)
			}
		}

		if yes {
			if llm.IsRawCommitJSON(commitMessage) {
				color.Yellow("⚠️ Template parsing failed, using raw response")
			}
			commitExec := exec.Command("git", "commit", "-m", commitMessage)
			commitExec.Stdout = os.Stdout
			commitExec.Stderr = os.Stderr
			if err := commitExec.Run(); err != nil {
				fmt.Printf("Error committing changes: %v\n", err)
				exit(ExitGitCommit)
			}
			fmt.Println("\n" + commitMessage)
			color.Green("✅ Successfully committed changes!")
			return
		}

		for {
			isRaw := llm.IsRawCommitJSON(commitMessage)
			if isRaw {
				color.Yellow("\n⚠️ Template parsing failed, showing raw response:")
			} else {
				fmt.Println("\n📝 Generated commit message:")
			}
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println(commitMessage)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

			items := []string{"Commit this message", "Regenerate message"}
			if isRaw {
				items = []string{"Use this anyway", "Regenerate message"}
			}
			prompt := promptui.Select{
				Label: "Choose an action",
				Items: items,
				Size:  2,
			}

			commitChoice, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Error with prompt: %v\n", err)
				exit(ExitPrompt)
			}

			switch commitChoice {
			case 0:
				commitExec := exec.Command("git", "commit", "-m", commitMessage)
				if err := commitExec.Run(); err != nil {
					fmt.Printf("Error committing changes: %v\n", err)
					exit(ExitGitCommit)
				}
				color.Green("\n✅ Successfully committed changes!")

				pushPrompt := promptui.Select{
					Label: "Would you like to push these changes to the remote repository?",
					Items: []string{"No", "Yes"},
					Size:  2,
				}

				_, pushChoice, err := pushPrompt.Run()
				if err != nil {
					fmt.Printf("Error with prompt: %v\n", err)
					exit(ExitPrompt)
				}

				if pushChoice == "Yes" {
					pushExec := exec.Command("git", "push", "origin", "HEAD")
					output, err := pushExec.CombinedOutput()
					if err != nil {
						color.Red("Error pushing changes: %v\n%s", err, output)
						exit(ExitGitPush)
					}
					fmt.Printf("%s", output)
					color.Green("✅ Successfully pushed changes to remote repository!")
				} else {
					color.Yellow("Changes committed locally. Remember to push when ready.")
				}
				return
			case 1:
				fmt.Println("\n🤖 Regenerating commit message...")
				for attempt := 1; attempt <= 3; attempt++ {
					commitMessage, err = generateMessage(config, diffOutput)
					if err != nil {
						fmt.Printf("Error generating commit message: %v\n", err)
						exit(ExitLLM)
					}
					if !llm.IsRawCommitJSON(commitMessage) {
						break
					}
					if attempt < 3 {
						color.Yellow("⚠️ Response format incorrect (body must be a string list), retrying (%d/3)...", attempt)
					}
				}
				continue
			default:
				color.Red("Invalid choice")
			}
		}
	},
}

func init() {
	commitCmd.Flags().BoolP("yes", "y", false, "Skip all confirmations and commit directly")
	commitCmd.Flags().StringP("language", "l", "", "Language for commit message: en or zh (overrides config file)")
	commitCmd.Flags().IntP("timeout", "t", 0, "LLM request timeout in seconds (overrides config file)")
}

func generateMessage(config *llm.Config, diffOutput []byte) (string, error) {
	diff := processBinaryDiff(diffOutput)
	truncated := truncateDiff(diff)

	if llm.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] diff: original=%d bytes, after binary removal=%d, after truncation=%d\n",
			len(diffOutput), len(diff), len(truncated))
	}

	generator, err := config.GetMessageGenerator()
	if err != nil {
		return "", fmt.Errorf("error getting message generator: %w", err)
	}
	return generator.GenerateCommitMessage(string(truncated))
}
