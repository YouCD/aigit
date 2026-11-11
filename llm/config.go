package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Provider struct {
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
	BaseURL  string `json:"base_url"`
}

type Config struct {
	CurrentProvider string              `json:"current_provider"`
	Providers       map[string]Provider `json:"providers"`
	Language        string              `json:"language"`
	Emoji           bool                `json:"emoji"`
	Timeout         int                 `json:"timeout"`
}

func NewConfig() *Config {
	return &Config{
		Providers: make(map[string]Provider),
		Emoji:     true,
		Timeout:   300,
	}
}

// ConfigDir returns the per-user aigit config directory
// (~/.config/aigit on Unix, %AppData%\aigit on Windows).
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config directory: %w", err)
	}
	return filepath.Join(base, "aigit"), nil
}

func (c *Config) Load() error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.json")
	configData, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Return empty config if file doesn't exist
		}
		return fmt.Errorf("reading config file: %w", err)
	}

	if err := json.Unmarshal(configData, c); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	if c.Language == "" {
		c.Language = detectLanguage()
	}

	return nil
}

func (c *Config) Save() error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	jsonData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(configFile, jsonData, 0o600); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

func (c *Config) AddProvider(provider, apiKey string, endpoint ...string) error {
	p := Provider{APIKey: apiKey}
	if len(endpoint) > 0 {
		p.Endpoint = endpoint[0]
	}
	c.Providers[provider] = p
	if c.CurrentProvider == "" {
		c.CurrentProvider = provider
	}
	return c.Save()
}

// GetMessageGenerator returns a MessageGenerator instance based on the current provider.
func (c *Config) GetMessageGenerator() (MessageGenerator, error) {
	provider := c.CurrentProvider
	if provider == "" {
		return nil, fmt.Errorf("no provider configured; run: aigit auth add <api_key> <model> --base-url <url>")
	}

	p, exists := c.Providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider %s is not configured", provider)
	}

	if provider != ProviderOpenAICompatible {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if p.APIKey == "" {
		return nil, fmt.Errorf("API key is required for %s provider", provider)
	}
	if p.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required for %s provider", provider)
	}
	model := p.Endpoint
	if model == "" {
		return nil, fmt.Errorf("model is required for %s provider", provider)
	}
	return NewOpenAICompatibleGenerator(p.APIKey, model, p.BaseURL, c.Language, c.Emoji, c.Timeout), nil
}

func detectLanguage() string {
	lang := os.Getenv("LANG")
	if len(lang) >= 2 && lang[:2] == "zh" {
		return "zh"
	}
	return "en"
}
