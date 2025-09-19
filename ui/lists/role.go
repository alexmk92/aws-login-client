package lists

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexmk92/aws-login/core"
)

// RoleItem represents an item in the role selection list
type RoleItem struct {
	title string
	role  string
}

func (i RoleItem) Title() string       { return i.title }
func (i RoleItem) Description() string { return i.role }
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
			title: "None (continue as current user)",
			role:  "",
		},
	}

	// Add role items
	for _, role := range roles {
		// Get the profile name that has this assumable_role_id
		profileName := awsService.GetAssumedProfileName(role)

		items = append(items, RoleItem{
			title: profileName, // Display profile name
			role:  role,        // Store the full ARN
		})
	}

	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Select Role to Assume"
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
	// Define styles locally since we can't import from parent package
	const (
		Pink  = "#E03189" // JJ Brand Pink
		Green = "#BCE921" // JJ Brand Green
	)

	brandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(Pink)).
		Bold(true)

	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(Green)).
		Italic(true)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(Green))

	header := brandStyle.Render("JJ") + " " + accentStyle.Render("AWS Login")

	content := fmt.Sprintf("%s\n\n%s\n\n%s",
		accentStyle.Render("🔐 Role Selection"),
		infoStyle.Render("Select a role to assume or continue as the current user"),
		m.list.View(),
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(Green)).
		Padding(1, 2).
		Width(60)

	return fmt.Sprintf("\n%s\n\n%s\n\n", header, box.Render(content))
}

// GetChoice returns the selected role (empty string for "None")
func (m RoleListModel) GetChoice() interface{} {
	return m.choice
}

// IsSelected returns true if a role has been selected
func (m RoleListModel) IsSelected() bool {
	return m.selected
}
