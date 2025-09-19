package lists

import (
	"fmt"

	"github.com/alexmk92/aws-login/core"
	"github.com/alexmk92/aws-login/ui/types"
	tea "github.com/charmbracelet/bubbletea"
)

type ListService struct {
	// Private property accessible to any file in the lists package
	lists      map[types.AppState]tea.Model
	awsService *core.AWSService
}

func NewListService(awsService *core.AWSService) *ListService {
	return &ListService{
		lists:      make(map[types.AppState]tea.Model),
		awsService: awsService,
	}
}

func (f *ListService) Register(state types.AppState, model tea.Model) {
	// Overwrite the model in the map if it already exists if we wanted to not do
	// that we could check if the model exists using the below
	//
	// if _, exists := f.lists[state]; exists {
	// 	return
	// }

	f.lists[state] = model
}

func (f *ListService) Get(state types.AppState) tea.Model {
	if model, exists := f.lists[state]; exists {
		return model
	}
	return nil
}

func (f *ListService) GetLists() map[types.AppState]tea.Model {
	return f.lists
}

// GetActiveModel returns the model for the current state, creating it if needed
func (f *ListService) GetActiveModel(state types.AppState, profile string) (tea.Model, error) {
	// Check if model already exists
	if model := f.Get(state); model != nil {
		return model, nil
	}

	var model tea.Model

	switch state {
	case types.StateProfileSelection:
		awsSvc := f.awsService
		profileModel := NewProfileListModel(awsSvc)
		model = profileModel
	case types.StateDriverSelection:
		driverModel := NewDriverListModel()
		model = driverModel
	case types.StateRoleSelection:
		awsSvc := f.awsService
		roleModel := NewRoleListModel(awsSvc, profile)
		model = roleModel
	default:
		return nil, fmt.Errorf("no active model for state: %v", state)
	}

	// Register the model in the service
	f.Register(state, model)
	return model, nil
}

// UpdateActiveModel updates the active model for the current state
func (f *ListService) UpdateActiveModel(state types.AppState, msg tea.Msg) (tea.Model, tea.Cmd) {
	if model := f.Get(state); model != nil {
		updatedModel, cmd := model.Update(msg)
		f.Register(state, updatedModel)
		return updatedModel, cmd
	}

	return nil, nil
}

// GetActiveModelForState returns the model for a specific state
func (f *ListService) GetActiveModelForState(state types.AppState) tea.Model {
	return f.Get(state)
}
