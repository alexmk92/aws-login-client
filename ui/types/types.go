package types

import tea "github.com/charmbracelet/bubbletea"

// AppState represents the current state of the application
type AppState int

const (
	StateProfileSelection AppState = iota // increments by 1 for each position in this const defintion until we hit a new type hint (i.e string, int etc.)
	StateDriverSelection
	StateRoleSelection
	StateLoading
	StateMFAInput
	StateProcessing
	StateDone
)

// ListModel defines the common interface for all selection models
type ListModel interface {
	tea.Model
	GetChoice() interface{}
	IsSelected() bool
}

// ListItem defines the common interface for all list items
type ListItem interface {
	Title() string
	Description() string
	FilterValue() string
}
