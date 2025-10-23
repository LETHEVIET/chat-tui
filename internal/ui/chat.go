package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/LETHEVIET/chat-tui/internal/commands"
	"github.com/LETHEVIET/chat-tui/internal/config"
	"github.com/LETHEVIET/chat-tui/internal/llm"
	"github.com/LETHEVIET/chat-tui/internal/ui/components"
	"github.com/LETHEVIET/chat-tui/internal/version"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// ChatModel is the main Bubble Tea model for the chat interface
type ChatModel struct {
	config             *config.Config
	client             llm.Client
	messages           []llm.Message
	input              *components.InputComponent
	messageComp        *components.MessageComponent
	stats              *components.StatsComponent
	streaming          bool
	streamContent      string
	streamChan         <-chan llm.StreamChunk
	streamStats        *llm.RequestStats
	err                error
	width              int
	height             int
	systemPrompt       string
	ready              bool
	suggestions        []commands.CommandDef
	selectedSuggestion int
	showBanner         bool
}

// Messages for async operations
type streamChunkMsg struct {
	chunk llm.StreamChunk
}

type streamCompleteMsg struct {
	stats *llm.RequestStats
}

type errorMsg struct {
	err error
}

type configReloadedMsg struct {
	config *config.Config
}

// NewChatModel creates a new chat model
func NewChatModel(cfg *config.Config) (*ChatModel, error) {
	// Create LLM client
	client := llm.NewOpenAIClient(
		cfg.APIKey,
		cfg.BaseURL,
		cfg.Model,
		cfg.Temperature,
		cfg.MaxTokens,
	)

	// Create UI components
	input := components.NewInputComponent()

	messageComp, err := components.NewMessageComponent(100)
	if err != nil {
		return nil, fmt.Errorf("failed to create message component: %w", err)
	}

	stats := components.NewStatsComponent()
	if !cfg.UI.ShowStats {
		stats.Toggle() // Start with stats hidden if config says so
	}

	// Initialize with system prompt
	messages := []llm.Message{}
	if cfg.SystemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: cfg.SystemPrompt,
		})
	}

	return &ChatModel{
		config:       cfg,
		client:       client,
		messages:     messages,
		input:        input,
		messageComp:  messageComp,
		stats:        stats,
		systemPrompt: cfg.SystemPrompt,
		ready:        true,
		showBanner:   true,
	}, nil
}

// Init initializes the model
func (m *ChatModel) Init() tea.Cmd {
	return m.input.Init()
}

// Update handles messages
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(msg.Width - 4)

	case streamStartMsg:
		m.streamChan = msg.chunks
		m.streamStats = msg.stats
		return m, m.waitForChunk()

	case tea.KeyMsg:
		if m.streaming {
			// Allow Ctrl+C to cancel streaming
			if msg.Type == tea.KeyCtrlC {
				m.streaming = false
				m.err = fmt.Errorf("streaming cancelled")
				return m, nil
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlS:
			m.stats.Toggle()
			return m, nil

		case tea.KeyCtrlD:
			m.input.ToggleMultilineMode()
			return m, nil

		case tea.KeyTab:
			// Autocomplete command
			if len(m.suggestions) > 0 {
				suggestion := m.suggestions[m.selectedSuggestion]
				m.input.SetValue("/" + suggestion.Name + " ")
				m.suggestions = nil
				m.selectedSuggestion = 0
			}
			return m, nil

		case tea.KeyUp:
			// Navigate suggestions up (only if we have suggestions)
			if len(m.suggestions) > 1 {
				m.selectedSuggestion--
				if m.selectedSuggestion < 0 {
					m.selectedSuggestion = len(m.suggestions) - 1
				}
				return m, nil
			}

		case tea.KeyDown:
			// Navigate suggestions down (only if we have suggestions)
			if len(m.suggestions) > 1 {
				m.selectedSuggestion++
				if m.selectedSuggestion >= len(m.suggestions) {
					m.selectedSuggestion = 0
				}
				return m, nil
			}

		case tea.KeyEnter:
			input := strings.TrimSpace(m.input.Value())
			if input == "" {
				return m, nil
			}

			// Auto-complete command if suggestions exist
			if commands.IsCommand(input) && len(m.suggestions) > 0 {
				// Use the selected (or first) suggestion to autocomplete
				suggestion := m.suggestions[m.selectedSuggestion]
				input = "/" + suggestion.Name
			}

			// Clear suggestions
			m.suggestions = nil
			m.selectedSuggestion = 0

			// Check if it's a command
			if commands.IsCommand(input) {
				return m, m.handleCommand(input)
			}

			// Add user message
			m.messages = append(m.messages, llm.Message{
				Role:    "user",
				Content: input,
			})

			// Add input to history before resetting
			m.input.AddToHistory(input)

			m.input.Reset()
			m.streaming = true
			m.streamContent = ""

			return m, m.streamResponse()
		}

	case streamChunkMsg:
		if msg.chunk.Error != nil {
			m.streaming = false
			m.streamChan = nil
			m.err = msg.chunk.Error
			return m, nil
		}

		if msg.chunk.Done {
			m.streaming = false
			m.streamChan = nil
			// Add assistant message
			if m.streamContent != "" {
				m.messages = append(m.messages, llm.Message{
					Role:    "assistant",
					Content: m.streamContent,
				})
			}
			m.stats.SetStats(m.streamStats)
			return m, nil
		}

		m.streamContent += msg.chunk.Content
		return m, m.waitForChunk()

	case streamCompleteMsg:
		m.streaming = false
		m.streamChan = nil
		m.stats.SetStats(msg.stats)
		if m.streamContent != "" {
			m.messages = append(m.messages, llm.Message{
				Role:    "assistant",
				Content: m.streamContent,
			})
		}
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.streaming = false
		return m, nil

	case configReloadedMsg:
		m.config = msg.config
		m.client = llm.NewOpenAIClient(
			msg.config.APIKey,
			msg.config.BaseURL,
			msg.config.Model,
			msg.config.Temperature,
			msg.config.MaxTokens,
		)
		m.err = nil
		return m, nil
	}

	// Update input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// Update suggestions based on input
	currentInput := m.input.Value()
	m.suggestions = commands.GetSuggestions(currentInput)
	if len(m.suggestions) == 0 {
		m.selectedSuggestion = 0
	} else if m.selectedSuggestion >= len(m.suggestions) {
		m.selectedSuggestion = len(m.suggestions) - 1
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *ChatModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var view strings.Builder

	// Banner (always visible)
	view.WriteString(RenderBanner(version.AppName, version.Description, version.Version, m.config.Model, m.config.BaseURL))
	view.WriteString("\n\n")

	// Messages (render all, no height limit in inline mode)
	for _, msg := range m.messages {
		if msg.Role == "system" {
			continue
		}
		view.WriteString(m.messageComp.RenderMessage(msg.Role, msg.Content))
		view.WriteString("\n")
	}

	// Render streaming content
	if m.streaming && m.streamContent != "" {
		view.WriteString(m.messageComp.RenderMessage("assistant", m.streamContent+" "+TypingStyle.Render("▊")))
	} else if m.streaming {
		view.WriteString(m.messageComp.RenderTyping())
		view.WriteString("\n")
	}

	// Error display
	if m.err != nil {
		view.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		view.WriteString("\n\n")
	}

	// Command suggestions
	if len(m.suggestions) > 0 {
		view.WriteString(m.renderSuggestions())
		view.WriteString("\n")
	}

	// Input prompt
	view.WriteString(m.input.View())

	// Bottom status bar: input mode + stats (on same line)
	statusLine := m.input.GetModeIndicator()
	if m.stats.IsVisible() {
		compactStats := m.stats.RenderCompactStats()
		if compactStats != "" {
			statusLine += "  " + compactStats
		}
	}
	if statusLine != "" {
		view.WriteString("\n")
		view.WriteString(statusLine)
	}

	return view.String()
}

// renderMessages renders the message history
func (m *ChatModel) renderMessages(maxHeight int) string {
	var content strings.Builder

	// Render all messages except system prompt
	for _, msg := range m.messages {
		if msg.Role == "system" {
			continue
		}
		content.WriteString(m.messageComp.RenderMessage(msg.Role, msg.Content))
		content.WriteString("\n")
	}

	// Render streaming content
	if m.streaming && m.streamContent != "" {
		// Use the same rendering style as completed messages
		content.WriteString(m.messageComp.RenderMessage("assistant", m.streamContent+" "+TypingStyle.Render("▊")))
		content.WriteString("\n")
	} else if m.streaming {
		content.WriteString(m.messageComp.RenderTyping())
		content.WriteString("\n")
	}

	// Limit height and add scrolling effect
	lines := strings.Split(content.String(), "\n")
	if maxHeight > 0 && len(lines) > maxHeight {
		lines = lines[len(lines)-maxHeight:]
	}

	return strings.Join(lines, "\n")
}

// streamResponse starts streaming a response
func (m *ChatModel) streamResponse() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		chunks, stats, err := m.client.ChatStream(ctx, m.messages)
		if err != nil {
			return errorMsg{err: err}
		}

		// Store the channel and stats for reading chunks
		return streamStartMsg{chunks: chunks, stats: stats}
	}
}

type streamStartMsg struct {
	chunks <-chan llm.StreamChunk
	stats  *llm.RequestStats
}

// waitForChunk waits for the next stream chunk
func (m *ChatModel) waitForChunk() tea.Cmd {
	if m.streamChan == nil {
		return nil
	}

	return func() tea.Msg {
		chunk, ok := <-m.streamChan
		if !ok {
			return streamCompleteMsg{stats: m.streamStats}
		}
		return streamChunkMsg{chunk: chunk}
	}
}

// handleCommand processes slash commands
func (m *ChatModel) handleCommand(input string) tea.Cmd {
	cmd, err := commands.ParseCommand(input)
	if err != nil {
		m.err = err
		return nil
	}

	switch cmd.Name {
	case "help":
		m.err = nil
		// Display help as assistant message so it's visible
		m.messages = append(m.messages, llm.Message{
			Role:    "assistant",
			Content: commands.CommandHelp(),
		})

	case "new", "clear":
		m.messages = []llm.Message{}
		if m.systemPrompt != "" {
			m.messages = append(m.messages, llm.Message{
				Role:    "system",
				Content: m.systemPrompt,
			})
		}
		m.err = nil
		m.streamContent = ""

	case "reload":
		return func() tea.Msg {
			cfg, err := config.Load()
			if err != nil {
				return errorMsg{err: fmt.Errorf("failed to reload config: %w", err)}
			}
			return configReloadedMsg{config: cfg}
		}

	case "delete":
		// Delete last turn (user message + assistant response)
		if len(m.messages) >= 2 {
			// Check if last message is from assistant
			if m.messages[len(m.messages)-1].Role == "assistant" {
				m.messages = m.messages[:len(m.messages)-2]
			} else {
				m.messages = m.messages[:len(m.messages)-1]
			}
			m.err = nil
		} else {
			m.err = fmt.Errorf("no messages to delete")
		}

	case "stats":
		m.stats.Toggle()

	case "temp":
		if err := cmd.ValidateArgs(1, 1); err != nil {
			m.err = err
			return nil
		}
		temp, err := cmd.GetFloatArg(0)
		if err != nil {
			m.err = err
			return nil
		}
		if temp < 0 || temp > 2 {
			m.err = fmt.Errorf("temperature must be between 0 and 2")
			return nil
		}
		m.client.SetTemperature(temp)
		m.err = nil
		m.messages = append(m.messages, llm.Message{
			Role:    "system",
			Content: fmt.Sprintf("Temperature set to %.2f", temp),
		})

	case "system":
		if err := cmd.ValidateArgs(1, 0); err != nil {
			m.err = err
			return nil
		}
		newPrompt := cmd.GetRestAsString(0)
		m.systemPrompt = newPrompt
		// Update system message if it exists
		if len(m.messages) > 0 && m.messages[0].Role == "system" {
			m.messages[0].Content = newPrompt
		} else {
			m.messages = append([]llm.Message{{Role: "system", Content: newPrompt}}, m.messages...)
		}
		m.err = nil

	case "copy":
		// Copy last assistant message
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Role == "assistant" {
				if err := clipboard.WriteAll(m.messages[i].Content); err != nil {
					m.err = fmt.Errorf("failed to copy: %w", err)
				} else {
					m.err = nil
					m.messages = append(m.messages, llm.Message{
						Role:    "system",
						Content: "Last response copied to clipboard",
					})
				}
				break
			}
		}

	case "multiline":
		m.input.ToggleMultilineMode()

	case "exit":
		return tea.Quit

	default:
		m.err = fmt.Errorf("unknown command: /%s (type /help for available commands)", cmd.Name)
	}

	// Add command to history before resetting
	m.input.AddToHistory(input)

	m.input.Reset()
	m.suggestions = nil
	m.selectedSuggestion = 0
	return nil
}

// renderSuggestions renders command suggestions
func (m *ChatModel) renderSuggestions() string {
	if len(m.suggestions) == 0 {
		return ""
	}

	var suggestions strings.Builder

	// Limit suggestions display
	maxSuggestions := 5
	displayCount := len(m.suggestions)
	if displayCount > maxSuggestions {
		displayCount = maxSuggestions
	}

	suggestions.WriteString(HelpStyle.Render("Suggestions (Enter/Tab to complete, ↑↓ to navigate):"))
	suggestions.WriteString("\n")

	for i := 0; i < displayCount; i++ {
		cmd := m.suggestions[i]
		line := fmt.Sprintf("  %-12s - %s", cmd.Usage, cmd.Description)

		if i == m.selectedSuggestion {
			// Highlight selected suggestion
			suggestions.WriteString(SuccessStyle.Render("▸ " + line))
		} else {
			suggestions.WriteString(HelpStyle.Render("  " + line))
		}
		suggestions.WriteString("\n")
	}

	if len(m.suggestions) > maxSuggestions {
		suggestions.WriteString(HelpStyle.Render(fmt.Sprintf("  ... and %d more", len(m.suggestions)-maxSuggestions)))
		suggestions.WriteString("\n")
	}

	return suggestions.String()
}
