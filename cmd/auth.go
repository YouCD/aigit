package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/YouCD/aigit/llm"
)

var authCmd = &cobra.Command{
	Use:                   "auth",
	Short:                 "Manage the LLM provider",
	Long:                  "Configure the openai-compatible provider.",
	DisableFlagsInUseLine: true,
}

var authAddCmd = &cobra.Command{
	Use:                   "add <api_key> <model>",
	Short:                 "Configure the openai-compatible provider",
	Long:                  "Configure the openai-compatible provider. Both model and --base-url are required.",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			color.Red("Not enough arguments")
			color.Red(cmd.Long)
			color.Red("\nUsage: aigit auth add <api_key> <model> --base-url <url>")
			exit(ExitInvalidArgs)
		}

		apiKey := strings.TrimSpace(args[0])
		provider := llm.ProviderOpenAICompatible

		config := llm.NewConfig()
		if err := config.Load(); err != nil {
			fmt.Printf("Error reading config: %v\n", err)
			exit(ExitConfigLoad)
		}

		if err := config.AddProvider(provider, apiKey, args[1:]...); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			exit(ExitConfigSave)
		}

		color.Green("Successfully added API key for %s", provider)

		baseURL, _ := cmd.Flags().GetString("base-url")
		if baseURL == "" {
			color.Red("--base-url is required for openai-compatible provider")
			exit(ExitUnsupportedProvider)
		}
		p := config.Providers[provider]
		p.BaseURL = baseURL
		config.Providers[provider] = p
		if err := config.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			exit(ExitConfigSave)
		}
	},
}

func init() {
	authAddCmd.Flags().String("base-url", "", "API base URL (required for openai-compatible provider)")
	authCmd.AddCommand(authAddCmd)
}
