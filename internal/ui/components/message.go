package components

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	userMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86"))

	assistantMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212"))

	systemMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)

	typingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
)

// MessageComponent handles rendering of chat messages
type MessageComponent struct {
	glamourRenderer *glamour.TermRenderer
}

// NewMessageComponent creates a new message component
func NewMessageComponent(width int) (*MessageComponent, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-10),
	)
	if err != nil {
		return nil, err
	}

	return &MessageComponent{
		glamourRenderer: r,
	}, nil
}

// RenderMessage renders a single message with proper formatting
func (m *MessageComponent) RenderMessage(role, content string) string {
	switch role {
	case "user":
		// User messages: render as plain text with ">" prefix (no markdown)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				lines[i] = userMessageStyle.Render(">")
			} else {
				lines[i] = userMessageStyle.Render("> ") + line
			}
		}
		return strings.Join(lines, "\n")

	case "assistant":
		// Assistant messages: render with markdown, no prefix
		rendered, err := m.glamourRenderer.Render(content)
		if err != nil {
			// Fallback to plain text if markdown rendering fails
			rendered = content
		}
		// Remove trailing newlines
		rendered = strings.TrimRight(rendered, "\n")
		return rendered + "\n"

	case "system":
		// System messages: render with label
		rendered, err := m.glamourRenderer.Render(content)
		if err != nil {
			rendered = content
		}
		rendered = strings.TrimRight(rendered, "\n")
		headerLine := systemMessageStyle.Render("System:")
		return headerLine + "\n" + rendered + "\n"

	default:
		return content + "\n"
	}
}

// RenderTyping renders a typing indicator
func (m *MessageComponent) RenderTyping() string {
	return typingStyle.Render("typing...")
}
