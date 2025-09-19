package lists

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexmk92/aws-login/core/auth_drivers"
)

// DriverItem represents an item in the driver selection list
type DriverItem struct {
	title       string
	description string
	driver      auth_drivers.AuthDriverName
	available   bool
}

func (i DriverItem) Title() string       { return i.title }
func (i DriverItem) Description() string { return i.description }
func (i DriverItem) FilterValue() string { return i.title }

// DriverListModel handles the driver selection UI
type DriverListModel struct {
	list     list.Model
	choice   auth_drivers.AuthDriverName
	selected bool
}

// NewDriverListModel creates a new driver selection model
func NewDriverListModel() DriverListModel {
	items := []list.Item{
		DriverItem{
			title:       "Manual",
			description: "Enter MFA code manually",
			driver:      auth_drivers.AuthDriverManual,
			available:   true,
		},
		DriverItem{
			title:       "1Password",
			description: "Use 1Password CLI (requires 1Password CLI)",
			driver:      auth_drivers.AuthDriver1Password,
			available:   auth_drivers.OnePasswordDriver{}.IsInstalled(),
		},
	}

	// Filter out unavailable items
	availableItems := []list.Item{}
	for _, item := range items {
		if driverItem, ok := item.(DriverItem); ok && driverItem.available {
			availableItems = append(availableItems, item)
		}
	}

	l := list.New(availableItems, list.NewDefaultDelegate(), 80, 20)
	l.Title = "🔐 Authentication Driver Selection"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	l.Styles.PaginationStyle = list.Styles{}.PaginationStyle.MarginLeft(2)
	l.Styles.HelpStyle = list.Styles{}.HelpStyle.MarginLeft(2)

	return DriverListModel{
		list: l,
	}
}

// Init initializes the driver selection model
func (m DriverListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the driver selection model
func (m DriverListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if i, ok := m.list.SelectedItem().(DriverItem); ok {
				m.choice = i.driver
				m.selected = true
				return m, nil
			}
		}
	}

	// Update the list with the message
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the driver selection UI
func (m DriverListModel) View() string {
	return "\n" + m.list.View()
}

// GetChoice returns the selected driver
func (m DriverListModel) GetChoice() interface{} {
	return m.choice
}

// IsSelected returns true if a driver has been selected
func (m DriverListModel) IsSelected() bool {
	return m.selected
}
