package commands

import (
	"fmt"
	"strconv"
	"strings"
)

// CommandDef represents a command definition
type CommandDef struct {
	Name        string
	Description string
	Usage       string
}

// AvailableCommands lists all available commands
var AvailableCommands = []CommandDef{
	{Name: "help", Description: "Show help message", Usage: "/help"},
	{Name: "new", Description: "Start a new conversation", Usage: "/new"},
	{Name: "clear", Description: "Clear chat history", Usage: "/clear"},
	{Name: "reload", Description: "Reload configuration", Usage: "/reload"},
	{Name: "temp", Description: "Set temperature", Usage: "/temp <0-2>"},
	{Name: "system", Description: "Set system prompt", Usage: "/system <text>"},
	{Name: "delete", Description: "Delete last turn", Usage: "/delete"},
	{Name: "save", Description: "Save conversation", Usage: "/save <file>"},
	{Name: "load", Description: "Load conversation", Usage: "/load <file>"},
	{Name: "tokens", Description: "Show token usage", Usage: "/tokens"},
	{Name: "cost", Description: "Show estimated cost", Usage: "/cost"},
	{Name: "export", Description: "Export as markdown", Usage: "/export"},
	{Name: "stats", Description: "Toggle stats panel", Usage: "/stats"},
	{Name: "debug", Description: "Toggle debug mode", Usage: "/debug"},
	{Name: "retry", Description: "Retry last message", Usage: "/retry"},
	{Name: "copy", Description: "Copy last response", Usage: "/copy"},
	{Name: "edit", Description: "Edit last message", Usage: "/edit"},
	{Name: "multiline", Description: "Toggle multiline mode", Usage: "/multiline"},
	{Name: "exit", Description: "Exit the application", Usage: "/exit"},
}

// Command represents a slash command
type Command struct {
	Name string
	Args []string
}

// ParseCommand parses a slash command from user input
func ParseCommand(input string) (*Command, error) {
	input = strings.TrimSpace(input)

	if !strings.HasPrefix(input, "/") {
		return nil, fmt.Errorf("not a command")
	}

	// Remove leading slash
	input = strings.TrimPrefix(input, "/")

	// Split into parts
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := &Command{
		Name: strings.ToLower(parts[0]),
		Args: parts[1:],
	}

	return cmd, nil
}

// IsCommand checks if the input is a slash command
func IsCommand(input string) bool {
	return strings.HasPrefix(strings.TrimSpace(input), "/")
}

// CommandHelp returns help text for all commands
func CommandHelp() string {
	return `Available Commands:
/help           - Show this help message
/new            - Start a new conversation
/clear          - Clear chat history (alias for /new)
/reload         - Reload configuration from .chat-tui.yaml
/temp <0-1>     - Set temperature (e.g., /temp 0.7)
/system <text>  - Set system prompt
/delete         - Delete last turn (user message + assistant response)
/save <file>    - Save conversation to file
/load <file>    - Load conversation from file
/tokens         - Show token usage
/cost           - Show estimated cost
/export         - Export conversation as markdown
/stats          - Toggle stats panel
/debug          - Toggle debug mode
/retry          - Retry last message
/copy           - Copy last response to clipboard
/edit           - Edit last message
/multiline      - Toggle multiline input mode
/quit           - Exit the application`
}

// ValidateArgs validates command arguments
func (c *Command) ValidateArgs(min, max int) error {
	if len(c.Args) < min {
		return fmt.Errorf("command /%s requires at least %d argument(s)", c.Name, min)
	}
	if max > 0 && len(c.Args) > max {
		return fmt.Errorf("command /%s accepts at most %d argument(s)", c.Name, max)
	}
	return nil
}

// GetFloatArg gets a float argument by index
func (c *Command) GetFloatArg(index int) (float64, error) {
	if index >= len(c.Args) {
		return 0, fmt.Errorf("argument %d not found", index)
	}

	val, err := strconv.ParseFloat(c.Args[index], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float value: %s", c.Args[index])
	}

	return val, nil
}

// GetIntArg gets an integer argument by index
func (c *Command) GetIntArg(index int) (int, error) {
	if index >= len(c.Args) {
		return 0, fmt.Errorf("argument %d not found", index)
	}

	val, err := strconv.Atoi(c.Args[index])
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %s", c.Args[index])
	}

	return val, nil
}

// GetStringArg gets a string argument by index
func (c *Command) GetStringArg(index int) (string, error) {
	if index >= len(c.Args) {
		return "", fmt.Errorf("argument %d not found", index)
	}

	return c.Args[index], nil
}

// GetRestAsString gets all remaining arguments as a single string
func (c *Command) GetRestAsString(startIndex int) string {
	if startIndex >= len(c.Args) {
		return ""
	}

	return strings.Join(c.Args[startIndex:], " ")
}

// GetSuggestions returns command suggestions based on input
func GetSuggestions(input string) []CommandDef {
	input = strings.TrimSpace(input)

	// If not a command, return empty
	if !strings.HasPrefix(input, "/") {
		return nil
	}

	// Remove leading slash
	query := strings.TrimPrefix(input, "/")
	query = strings.ToLower(query)

	// If empty, show all commands
	if query == "" {
		return AvailableCommands
	}

	// Filter commands that start with the query
	var suggestions []CommandDef
	for _, cmd := range AvailableCommands {
		if strings.HasPrefix(cmd.Name, query) {
			suggestions = append(suggestions, cmd)
		}
	}

	return suggestions
}
