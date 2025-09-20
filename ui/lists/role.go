package lists

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexmk92/aws-login/core"
)

// RoleItem represents an item in the role selection list
type RoleItem struct {
	title       string
	role        string // the actual ARN to assume
	description string // the description to render in the list
}

func (i RoleItem) Title() string       { return i.title }
func (i RoleItem) Description() string { return i.description }
func (i RoleItem) FilterValue() string { return i.title }

// RoleListModel handles the role selection UI
type RoleListModel struct {
	list     list.Model
	choice   string
	selected bool
}

// NewRoleListModel creates a new role selection model
func NewRoleListModel(awsService *core.AWSService, profile string) RoleListModel {
	roles := awsService.GetAssumableRoles(profile)

	// Create items list with "None" option first
	items := []list.Item{
		RoleItem{
			title:       profile,
			role:        "",
			description: fmt.Sprintf("Continue as the current user: [%s]", profile),
		},
	}

	// Add role items
	for _, role := range roles {
		// Get the profile name that has this assumable_role_id
		profileName := awsService.GetAssumedProfileName(role)
		roleParts := strings.Split(role, ":")
		formattedRole := roleParts[len(roleParts)-1]

		items = append(items, RoleItem{
			title:       profileName,
			description: fmt.Sprintf("Assume: [%s]", formattedRole),
			role:        role,
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "🔐 Select Role to Assume"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	l.Styles.PaginationStyle = list.Styles{}.PaginationStyle.MarginLeft(2)
	l.Styles.HelpStyle = list.Styles{}.HelpStyle.MarginLeft(2)

	return RoleListModel{
		list:     l,
		choice:   "",
		selected: false,
	}
}

// Init initializes the role selection model
func (m RoleListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the role selection model
func (m RoleListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(RoleItem); ok {
				m.choice = i.role
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

// View renders the role selection UI
func (m RoleListModel) View() string {
	box := lipgloss.NewStyle().
		Padding(1, 2).
		Width(60)

	return box.Render(m.list.View())
}

// GetChoice returns the selected role (empty string for "None")
func (m RoleListModel) GetChoice() interface{} {
	return m.choice
}

// IsSelected returns true if a role has been selected
func (m RoleListModel) IsSelected() bool {
	return m.selected
}
