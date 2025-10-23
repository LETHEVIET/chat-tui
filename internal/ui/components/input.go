package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

// InputComponent handles user input
type InputComponent struct {
	textarea      textarea.Model
	multilineMode bool
	history       []string
	historyIndex  int
	currentInput  string
}

// NewInputComponent creates a new input component
func NewInputComponent() *InputComponent {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false

	return &InputComponent{
		textarea:      ta,
		multilineMode: false,
		history:       []string{},
		historyIndex:  -1,
		currentInput:  "",
	}
}

// Init initializes the component
func (i *InputComponent) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles input updates
func (i *InputComponent) Update(msg tea.Msg) (*InputComponent, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if len(i.history) > 0 {
				if i.historyIndex == -1 {
					// Save current input before navigating history
					i.currentInput = i.textarea.Value()
				}
				i.historyIndex++
				if i.historyIndex >= len(i.history) {
					i.historyIndex = len(i.history) - 1
				}
				i.textarea.SetValue(i.history[len(i.history)-1-i.historyIndex])
			}
			return i, nil
		case "down":
			if i.historyIndex >= 0 {
				i.historyIndex--
				if i.historyIndex == -1 {
					// Restore current input
					i.textarea.SetValue(i.currentInput)
				} else {
					i.textarea.SetValue(i.history[len(i.history)-1-i.historyIndex])
				}
			}
			return i, nil
		}
	}

	i.textarea, cmd = i.textarea.Update(msg)
	return i, cmd
}

// View renders the input component
func (i *InputComponent) View() string {
	return i.textarea.View()
}

// GetModeIndicator returns the input mode indicator text
func (i *InputComponent) GetModeIndicator() string {
	mode := "single-line"
	if i.multilineMode {
		mode = "multi-line"
	}
	return helpStyle.Render("(" + mode + " mode)")
}

// Value returns the current input value
func (i *InputComponent) Value() string {
	return i.textarea.Value()
}

// Reset clears the input
func (i *InputComponent) Reset() {
	i.textarea.Reset()
	i.historyIndex = -1
	i.currentInput = ""
}

// Focus focuses the input
func (i *InputComponent) Focus() tea.Cmd {
	return i.textarea.Focus()
}

// Blur blurs the input
func (i *InputComponent) Blur() {
	i.textarea.Blur()
}

// ToggleMultilineMode toggles between single and multiline mode
func (i *InputComponent) ToggleMultilineMode() {
	i.multilineMode = !i.multilineMode
	if i.multilineMode {
		i.textarea.SetHeight(5)
	} else {
		i.textarea.SetHeight(1)
	}
}

// SetWidth sets the width of the input
func (i *InputComponent) SetWidth(width int) {
	i.textarea.SetWidth(width)
}

// SetValue sets the value of the input
func (i *InputComponent) SetValue(value string) {
	i.textarea.SetValue(value)
}

// AddToHistory adds the current input to history if it's not empty and different from the last entry
func (i *InputComponent) AddToHistory(input string) {
	if input != "" && (len(i.history) == 0 || i.history[len(i.history)-1] != input) {
		i.history = append(i.history, input)
		// Reset history navigation
		i.historyIndex = -1
		i.currentInput = ""
	}
}
