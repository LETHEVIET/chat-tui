package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	APIKey        string        `mapstructure:"api_key"`
	BaseURL       string        `mapstructure:"base_url"`
	Model         string        `mapstructure:"model"`
	Temperature   float64       `mapstructure:"temperature"`
	MaxTokens     int           `mapstructure:"max_tokens"`
	SystemPrompt  string        `mapstructure:"system_prompt"`
	UI            UIConfig      `mapstructure:"ui"`
	Debug         DebugConfig   `mapstructure:"debug"`
}

// UIConfig holds UI-specific settings
type UIConfig struct {
	Theme            string `mapstructure:"theme"`
	ShowStats        bool   `mapstructure:"show_stats"`
	SyntaxHighlight  bool   `mapstructure:"syntax_highlight"`
}

// DebugConfig holds debug-related settings
type DebugConfig struct {
	Verbose bool   `mapstructure:"verbose"`
	LogFile string `mapstructure:"log_file"`
}

// Default configuration values
var defaultConfig = Config{
	APIKey:       "not_needed",
	BaseURL:      "https://api.openai.com/v1",
	Model:        "gpt-4",
	Temperature:  0.7,
	MaxTokens:    4096,
	SystemPrompt: "You are a helpful assistant",
	UI: UIConfig{
		Theme:           "dark",
		ShowStats:       true,
		SyntaxHighlight: true,
	},
	Debug: DebugConfig{
		Verbose: false,
		LogFile: ".chat-tui.log",
	},
}

// Load reads configuration from .chat-tui.yaml file and environment variables
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Set up viper to look for .chat-tui.yaml
	viper.SetConfigName(".chat-tui")
	viper.SetConfigType("yaml")

	// Add search paths: current directory, then home directory
	viper.AddConfigPath(".")
	viper.AddConfigPath(homeDir)

	// Set defaults
	setDefaults()

	// Read environment variables with CHAT_TUI prefix
	viper.SetEnvPrefix("CHAT_TUI")
	viper.AutomaticEnv()

	// Try to read config file
	configFileNotFound := false
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configFileNotFound = true
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// If config file not found, run interactive setup
	if configFileNotFound {
		config, err := InteractiveSetup()
		if err != nil {
			return nil, fmt.Errorf("failed to complete interactive setup: %w", err)
		}

		// Save the config to file
		if err := saveConfig(config, ".chat-tui.yaml"); err != nil {
			return nil, fmt.Errorf("failed to save config: %w", err)
		}

		// Override API key from environment if set
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			config.APIKey = apiKey
		}

		fmt.Println("\nConfiguration saved to .chat-tui.yaml")
		fmt.Println("Starting chat...\n")

		return config, nil
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override API key from environment if set
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("api_key", defaultConfig.APIKey)
	viper.SetDefault("base_url", defaultConfig.BaseURL)
	viper.SetDefault("model", defaultConfig.Model)
	viper.SetDefault("temperature", defaultConfig.Temperature)
	viper.SetDefault("max_tokens", defaultConfig.MaxTokens)
	viper.SetDefault("system_prompt", defaultConfig.SystemPrompt)
	viper.SetDefault("ui.theme", defaultConfig.UI.Theme)
	viper.SetDefault("ui.show_stats", defaultConfig.UI.ShowStats)
	viper.SetDefault("ui.syntax_highlight", defaultConfig.UI.SyntaxHighlight)
	viper.SetDefault("debug.verbose", defaultConfig.Debug.Verbose)
	viper.SetDefault("debug.log_file", defaultConfig.Debug.LogFile)
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string) error {
	defaultYAML := `# Chat TUI Configuration
# OpenAI-compatible API settings
api_key: "not_needed"  # Optional: Set your API key here or use OPENAI_API_KEY environment variable
base_url: "https://api.openai.com/v1"  # Can be changed to any OpenAI-compatible endpoint
model: "gpt-4"
temperature: 0.7
max_tokens: 4096
system_prompt: "You are a helpful assistant"

ui:
  theme: dark  # or light
  show_stats: true
  syntax_highlight: true

debug:
  verbose: false
  log_file: .chat-tui.log
`

	return os.WriteFile(path, []byte(defaultYAML), 0644)
}

// InteractiveSetup prompts the user for configuration values
func InteractiveSetup() (*Config, error) {
	fmt.Println("Welcome to Chat TUI!")
	fmt.Println("No configuration file found. Let's set one up.\n")

	reader := bufio.NewReader(os.Stdin)
	config := defaultConfig

	// API Key
	apiKey, err := promptWithDefault(reader, "API Key", defaultConfig.APIKey)
	if err != nil {
		return nil, err
	}
	config.APIKey = apiKey

	// Base URL
	baseURL, err := promptWithDefault(reader, "Base URL", defaultConfig.BaseURL)
	if err != nil {
		return nil, err
	}
	config.BaseURL = baseURL

	// Model
	model, err := promptWithDefault(reader, "Model", defaultConfig.Model)
	if err != nil {
		return nil, err
	}
	config.Model = model

	// Temperature
	tempStr, err := promptWithDefault(reader, "Temperature (0.0-1.0)", fmt.Sprintf("%.1f", defaultConfig.Temperature))
	if err != nil {
		return nil, err
	}
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		fmt.Printf("Invalid temperature, using default: %.1f\n", defaultConfig.Temperature)
		temp = defaultConfig.Temperature
	}
	config.Temperature = temp

	// Max Tokens
	maxTokensStr, err := promptWithDefault(reader, "Max Tokens", fmt.Sprintf("%d", defaultConfig.MaxTokens))
	if err != nil {
		return nil, err
	}
	maxTokens, err := strconv.Atoi(maxTokensStr)
	if err != nil {
		fmt.Printf("Invalid max tokens, using default: %d\n", defaultConfig.MaxTokens)
		maxTokens = defaultConfig.MaxTokens
	}
	config.MaxTokens = maxTokens

	// System Prompt
	systemPrompt, err := promptWithDefault(reader, "System Prompt", defaultConfig.SystemPrompt)
	if err != nil {
		return nil, err
	}
	config.SystemPrompt = systemPrompt

	return &config, nil
}

// promptWithDefault prompts the user for input with a default value
func promptWithDefault(reader *bufio.Reader, prompt, defaultValue string) (string, error) {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}

	return input, nil
}

// saveConfig saves a configuration to a file
func saveConfig(cfg *Config, path string) error {
	configYAML := fmt.Sprintf(`# Chat TUI Configuration
# OpenAI-compatible API settings
api_key: "%s"  # Optional: Set your API key here or use OPENAI_API_KEY environment variable
base_url: "%s"  # Can be changed to any OpenAI-compatible endpoint
model: "%s"
temperature: %.1f
max_tokens: %d
system_prompt: "%s"

ui:
  theme: %s  # or light
  show_stats: %t
  syntax_highlight: %t

debug:
  verbose: %t
  log_file: %s
`,
		cfg.APIKey,
		cfg.BaseURL,
		cfg.Model,
		cfg.Temperature,
		cfg.MaxTokens,
		cfg.SystemPrompt,
		cfg.UI.Theme,
		cfg.UI.ShowStats,
		cfg.UI.SyntaxHighlight,
		cfg.Debug.Verbose,
		cfg.Debug.LogFile,
	)

	return os.WriteFile(path, []byte(configYAML), 0644)
}

// Save writes the current configuration to disk
func (c *Config) Save() error {
	// Save to current directory
	configPath := ".chat-tui.yaml"

	viper.Set("base_url", c.BaseURL)
	viper.Set("model", c.Model)
	viper.Set("temperature", c.Temperature)
	viper.Set("max_tokens", c.MaxTokens)
	viper.Set("system_prompt", c.SystemPrompt)
	viper.Set("ui.theme", c.UI.Theme)
	viper.Set("ui.show_stats", c.UI.ShowStats)
	viper.Set("ui.syntax_highlight", c.UI.SyntaxHighlight)
	viper.Set("debug.verbose", c.Debug.Verbose)
	viper.Set("debug.log_file", c.Debug.LogFile)

	return viper.WriteConfigAs(configPath)
}
