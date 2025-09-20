package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	Pink      = "#E03189" // JJ Brand Pink
	Orange    = "#ff6b35" // JJ Brand Orange
	Green     = "#BCE921" // JJ Brand Pink
	Red       = "#e74c3c"
	LightGray = "#999999"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Orange)).
			Bold(true).
			MarginLeft(2)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Pink)).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Red)).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green))

	pulseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Orange))

	brandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Pink)).
			Bold(true)

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(Green)).
			Italic(true)

	lightGrayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(LightGray)).
			Italic(true)
)

func NewMFAInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "000000"
	ti.Focus()
	ti.CharLimit = 6
	ti.Width = 20
	ti.Prompt = "MFA Code: "
	ti.PromptStyle = brandStyle
	ti.TextStyle = infoStyle
	ti.PlaceholderStyle = accentStyle.Copy().Faint(true)
	return ti
}
