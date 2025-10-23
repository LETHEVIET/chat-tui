package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/LETHEVIET/chat-tui/internal/llm"
)

var (
	statsPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			MarginTop(1)

	statsTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true).
			Underline(true)

	statsLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(20)

	statsValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	statsHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
)

// StatsComponent displays request statistics
type StatsComponent struct {
	visible bool
	stats   *llm.RequestStats
}

// NewStatsComponent creates a new stats component
func NewStatsComponent() *StatsComponent {
	return &StatsComponent{
		visible: true,
		stats:   nil,
	}
}

// Toggle toggles the visibility of stats
func (s *StatsComponent) Toggle() {
	s.visible = !s.visible
}

// SetStats updates the stats
func (s *StatsComponent) SetStats(stats *llm.RequestStats) {
	s.stats = stats
}

// IsVisible returns whether stats are visible
func (s *StatsComponent) IsVisible() bool {
	return s.visible
}

// View renders the stats panel
func (s *StatsComponent) View() string {
	if !s.visible {
		return ""
	}

	if s.stats == nil {
		return statsPanelStyle.Render(
			statsTitleStyle.Render("Stats") + "\n\n" +
				statsHelpStyle.Render("No request data yet"),
		)
	}

	var content strings.Builder

	// Title
	content.WriteString(statsTitleStyle.Render("Request Statistics"))
	content.WriteString("\n\n")

	// Model info
	content.WriteString(s.renderStat("Model", s.stats.Model))
	content.WriteString(s.renderStat("HTTP Status", fmt.Sprintf("%d", s.stats.HTTPStatus)))
	content.WriteString("\n")

	// Token stats
	content.WriteString(statsTitleStyle.Render("Tokens"))
	content.WriteString("\n")
	content.WriteString(s.renderStat("Input Tokens", fmt.Sprintf("%d", s.stats.InputTokens)))
	content.WriteString(s.renderStat("Output Tokens", fmt.Sprintf("%d", s.stats.OutputTokens)))
	content.WriteString(s.renderStat("Total Tokens", fmt.Sprintf("%d", s.stats.TotalTokens)))
	content.WriteString("\n")

	// Performance stats
	content.WriteString(statsTitleStyle.Render("Performance"))
	content.WriteString("\n")

	// Overall latency
	content.WriteString(s.renderStat("Total Latency", fmt.Sprintf("%.2fs", s.stats.Latency.Seconds())))

	// Time to first token
	if s.stats.TimeToFirstToken > 0 {
		content.WriteString(s.renderStat("Time to 1st Token", fmt.Sprintf("%.2fs", s.stats.TimeToFirstToken.Seconds())))
	}

	// Generation time (after first token)
	if s.stats.GenerationTime > 0 {
		content.WriteString(s.renderStat("Generation Time", fmt.Sprintf("%.2fs", s.stats.GenerationTime.Seconds())))
	}

	// Overall speed
	if s.stats.TokensPerSec > 0 {
		content.WriteString(s.renderStat("Avg Speed", fmt.Sprintf("%.2f tok/s", s.stats.TokensPerSec)))
	}

	// Post-first-token speed
	if s.stats.PostFirstTokenSpeed > 0 {
		content.WriteString(s.renderStat("Gen Speed", fmt.Sprintf("%.2f tok/s", s.stats.PostFirstTokenSpeed)))
	}

	// Cost estimate (if available)
	if s.stats.CostEstimate > 0 {
		content.WriteString("\n")
		content.WriteString(statsTitleStyle.Render("Cost"))
		content.WriteString("\n")
		content.WriteString(s.renderStat("Estimate", fmt.Sprintf("$%.6f", s.stats.CostEstimate)))
	}

	return statsPanelStyle.Render(content.String())
}

// renderStat renders a single stat line
func (s *StatsComponent) renderStat(label, value string) string {
	return statsLabelStyle.Render(label+":") + " " +
		statsValueStyle.Render(value) + "\n"
}

// RenderCompactStats renders a compact version of stats for inline display
func (s *StatsComponent) RenderCompactStats() string {
	if !s.visible || s.stats == nil {
		return ""
	}

	parts := []string{}

	if s.stats.TotalTokens > 0 {
		parts = append(parts, fmt.Sprintf("%d tok", s.stats.TotalTokens))
	}

	if s.stats.TimeToFirstToken > 0 {
		parts = append(parts, fmt.Sprintf("TTFT: %.2fs", s.stats.TimeToFirstToken.Seconds()))
	}

	if s.stats.PostFirstTokenSpeed > 0 {
		parts = append(parts, fmt.Sprintf("%.1f tok/s", s.stats.PostFirstTokenSpeed))
	} else if s.stats.TokensPerSec > 0 {
		parts = append(parts, fmt.Sprintf("%.1f tok/s", s.stats.TokensPerSec))
	}

	if s.stats.Latency > 0 {
		parts = append(parts, fmt.Sprintf("%.2fs", s.stats.Latency.Seconds()))
	}

	if len(parts) == 0 {
		return ""
	}

	return statsHelpStyle.Render("[" + strings.Join(parts, " • ") + "]")
}
