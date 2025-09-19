package lists

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexmk92/aws-login/core"
)

// ProfileSelectedMsg is sent when a profile is selected
type ProfileSelectedMsg string

// ProfileItem represents an item in the profile selection list
type ProfileItem struct {
	title   string
	profile string
}

func (i ProfileItem) Title() string       { return i.title }
func (i ProfileItem) Description() string { return i.profile }
func (i ProfileItem) FilterValue() string { return i.title }

// ProfileListModel handles the profile selection UI
type ProfileListModel struct {
	list     list.Model
	choice   string
	selected bool
}

// NewProfileListModel creates a new profile selection model
func NewProfileListModel(awsService *core.AWSService) ProfileListModel {
	profiles := awsService.GetValidProfiles()
	profileItems := make([]list.Item, len(profiles))
	for i, p := range profiles {
		profileItems[i] = ProfileItem{
			title:   p,
			profile: p,
		}
	}

	l := list.New(profileItems, list.NewDefaultDelegate(), 80, 20)
	l.Title = "🌍 AWS Profile Selection"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	l.Styles.PaginationStyle = list.Styles{}.PaginationStyle.MarginLeft(2)
	l.Styles.HelpStyle = list.Styles{}.HelpStyle.MarginLeft(2)

	return ProfileListModel{
		list: l,
	}
}

// Init initializes the profile selection model
func (m ProfileListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the profile selection model
func (m ProfileListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if i, ok := m.list.SelectedItem().(ProfileItem); ok {
				m.choice = i.profile
				m.selected = true
			}
		}
	}

	// Update the list with the message
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the profile selection UI
func (m ProfileListModel) View() string {
	return "\n" + m.list.View()
}

// GetChoice returns the selected profile
func (m ProfileListModel) GetChoice() interface{} {
	return m.choice
}

// IsSelected returns true if a profile has been selected
func (m ProfileListModel) IsSelected() bool {
	return m.selected
}
