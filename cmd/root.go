package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/LETHEVIET/chat-tui/internal/config"
	"github.com/LETHEVIET/chat-tui/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "chat-tui",
	Short: "A terminal chat interface for LLMs",
	Long: `chat-tui is a terminal-based chat interface for interacting with
OpenAI-compatible LLM APIs with streaming support and detailed statistics.`,
	RunE: runChat,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringP("config", "c", "", "config file (default is .chat-tui.yaml)")
	rootCmd.Flags().StringP("model", "m", "", "model to use")
	rootCmd.Flags().Float64P("temperature", "t", 0, "temperature for responses")
	rootCmd.Flags().StringP("base-url", "u", "", "base URL for API")
	rootCmd.Flags().BoolP("no-stats", "n", false, "disable stats panel")
}

func runChat(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with command line flags
	if model, _ := cmd.Flags().GetString("model"); model != "" {
		cfg.Model = model
	}

	if temp, _ := cmd.Flags().GetFloat64("temperature"); temp > 0 {
		cfg.Temperature = temp
	}

	if baseURL, _ := cmd.Flags().GetString("base-url"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	if noStats, _ := cmd.Flags().GetBool("no-stats"); noStats {
		cfg.UI.ShowStats = false
	}

	// Create chat model
	chatModel, err := ui.NewChatModel(cfg)
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}

	// Start Bubble Tea program (inline mode, not full-screen)
	p := tea.NewProgram(chatModel)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
