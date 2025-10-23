package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("86")  // Cyan
	secondaryColor = lipgloss.Color("212") // Pink
	accentColor    = lipgloss.Color("220") // Yellow
	mutedColor     = lipgloss.Color("240") // Gray
	errorColor     = lipgloss.Color("196") // Red
	successColor   = lipgloss.Color("46")  // Green

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// User message style
	UserMessageStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Padding(0, 1)

	// Assistant message style
	AssistantMessageStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true).
				Padding(0, 1)

	// System message style
	SystemMessageStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true).
				Padding(0, 1)

	// Input box style
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	// Focused input style
	FocusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				Padding(0, 1)

	// Stats panel style
	StatsPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2).
			MarginTop(1)

	// Stats title style
	StatsTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			Underline(true)

	// Stats label style
	StatsLabelStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Width(20)

	// Stats value style
	StatsValueStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Padding(0, 1)

	// Success message style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Padding(0, 1)

	// Typing indicator style
	TypingStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Padding(0, 1)

	// Code block style
	CodeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("252")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	// Command style (for slash commands)
	CommandStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Divider style
	DividerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Faint(true)

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Underline(true).
			Padding(0, 1)
)

// RenderDivider creates a horizontal divider
func RenderDivider(width int) string {
	return DividerStyle.Render(lipgloss.NewStyle().Width(width).Render("─"))
}

// RenderTitle renders the application title
func RenderTitle() string {
	return TitleStyle.Render("LLM Chat CLI")
}

// RenderHelp renders the help text
func RenderHelp() string {
	return HelpStyle.Render("Type a message or /help for commands • Ctrl+C to exit • Ctrl+S to toggle stats")
}

// RenderBanner renders a banner with program info
func RenderBanner(appName, appDesc, version, model, baseURL string) string {
	bannerStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		MarginBottom(1)

	asciiArtStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(false).
		Align(lipgloss.Center)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Italic(true).
		Align(lipgloss.Center)

	infoStyle := lipgloss.NewStyle().
		Foreground(mutedColor)

	var content strings.Builder

	// ASCII Art
	asciiArt := `░█▀▀░█░█░█▀█░▀█▀░░░▀█▀░█░█░▀█▀
░█░░░█▀█░█▀█░░█░░░░░█░░█░█░░█░
░▀▀▀░▀░▀░▀░▀░░▀░░░░░▀░░▀▀▀░▀▀▀
`

	content.WriteString(asciiArtStyle.Width(48).Render(asciiArt))
	content.WriteString("\n")
	content.WriteString(subtitleStyle.Width(48).Render(appDesc))
	content.WriteString("\n\n")

	// Info
	content.WriteString(infoStyle.Render(fmt.Sprintf("Version:  %s", version)))
	content.WriteString("\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf("Model:    %s", model)))
	content.WriteString("\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf("Endpoint: %s", baseURL)))
	content.WriteString("\n\n")

	// Quick help
	content.WriteString(HelpStyle.Render("Type /help for commands • Ctrl+C to exit"))

	return bannerStyle.Render(content.String())
}
