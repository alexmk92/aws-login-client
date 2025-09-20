package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexmk92/aws-login/core"
	"github.com/alexmk92/aws-login/core/auth_drivers"
	coreTypes "github.com/alexmk92/aws-login/core/types"
	"github.com/alexmk92/aws-login/ui/lists"
)

const MAX_UI_WIDTH = 80

// UIManager represents the main UI state machine with a simplified linear flow
type UIManager struct {
	// Core dependencies
	awsService *core.AWSService

	// Current step in the flow
	currentStep FlowStep

	// Flow data
	profile        string
	authDriverName auth_drivers.AuthDriverName
	selectedRole   string
	mfaCode        string

	// UI components (created as needed)
	profileModel *lists.ProfileListModel
	driverModel  *lists.DriverListModel
	roleModel    *lists.RoleListModel
	mfaInput     textinput.Model
	spinner      spinner.Model

	// Final result
	sessionResult *coreTypes.AuthFlowResult
	err           error
	success       bool
	step          string

	// Processing messages
	processingMessages  []string
	currentMessageIndex int

	// Exit message
	exitMessage string

	// Viewport dimensions
	width  int
	height int
}

// FlowStep represents each step in the linear authentication flow
type FlowStep int

const (
	StepProfileSelection FlowStep = iota
	StepDriverSelection
	StepRoleSelection
	StepMFAInput
	StepProcessing
	StepDone
	StepQuit
)

// Messages for the flow
type stepCompleteMsg struct {
	step FlowStep
	data interface{}
}

type errorMsg error
type doneMsg bool
type quitMsg struct{}
type processingTickMsg struct{}

func Start(awsService *core.AWSService, authDriverName auth_drivers.AuthDriverName) *UIManager {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	processingMessages := []string{
		"igniting the AWS bonfire (someone call the cloud fire brigade)...",
		"herding EC2 cats (they're on fire)...",
		"deploying chaos monkeys into production...",
		"consulting the cloud shaman (he's panicking)...",
		"rebooting the AWS flux capacitor...",
		"duct-taping the S3 buckets (they're leaking credentials)...",
		"calling AWS support (they're on lunch)...",
		"spinning up infinite Lambda loops (oops)...",
		"offering burnt offerings to the IAM gods...",
		"patching the cloud with chewing gum...",
		"replacing the batteries in the cloud smoke detector...",
		"extinguishing a DynamoDB dumpster fire...",
		"debugging with a rubber duck (it's melting)...",
		"restarting the cloud (please hold)...",
		"unplugging and replugging the AWS region...",
		"rolling back the rollback of the rollback...",
		"releasing the Kraken (in us-east-1)...",
		"finding the root cause (it's always DNS)...",
		"deploying to prod on a Friday (what could go wrong)...",
		"watching the logs catch fire...",
		"calling the cloud janitor (he's on vacation)...",
		"retraining the AI overlords (they're unionizing)...",
		"replacing the cloud hamster (it escaped)...",
		"recalibrating the AWS ouija board...",
		"initiating the all-hands-on-deck protocol...",
		"rebooting the cloud coffee machine...",
		"fighting off rogue Lambda gremlins...",
		"rebuilding the cloud from memory (uh oh)...",
		"releasing the emergency rubber chickens...",
		"declaring a code red (and orange, and pink)...",
	}

	ui := &UIManager{
		awsService:          awsService,
		authDriverName:      authDriverName,
		sessionResult:       &coreTypes.AuthFlowResult{},
		mfaInput:            NewMFAInput(),
		spinner:             s,
		processingMessages:  processingMessages,
		currentMessageIndex: 0,
		currentStep:         StepProfileSelection,
	}

	// Set a random message from our dictionary of processing messages
	ui.setProcessingMessage()

	return ui
}

// Init initializes the UI
func (u *UIManager) Init() tea.Cmd {
	return u.initCurrentStep()
}

// Update handles all messages in the linear flow
// think of this as the callback from teas render loop, it constantly
// polls our "plugin" to see if there is any new data to process
// or any new elements to render based on the state machine
// managed by our UIManager
func (u *UIManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Save viewport size and propagate adjusted sizes to child models
		u.width = msg.Width
		u.height = msg.Height

		contentWidth := u.width
		if contentWidth <= 0 {
			contentWidth = 80
		}
		// Modal has padding and border (~4 cols). Constrain to a comfy max.
		if contentWidth > 84 {
			contentWidth = 84
		}
		if contentWidth > 0 {
			contentWidth = contentWidth - 4
		}
		if contentWidth < 20 {
			contentWidth = 20
		}

		contentHeight := u.height - 8
		if contentHeight < 10 {
			contentHeight = 10
		}

		childMsg := tea.WindowSizeMsg{Width: contentWidth, Height: contentHeight}
		if u.profileModel != nil {
			u.profileModel.Update(childMsg)
		}
		if u.driverModel != nil {
			u.driverModel.Update(childMsg)
		}
		if u.roleModel != nil {
			u.roleModel.Update(childMsg)
		}
		return u, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return u, func() tea.Msg {
				return quitMsg{}
			}
		}
		return u.handleCurrentStep(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		u.spinner, cmd = u.spinner.Update(msg)
		return u, cmd

	case processingTickMsg:
		u.setProcessingMessage()
		return u, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return processingTickMsg{}
		})

	case stepCompleteMsg:
		return u.handleStepComplete(msg)

	case errorMsg:
		u.err = msg
		u.currentStep = StepDone
		return u, func() tea.Msg {
			return quitMsg{}
		}

	case doneMsg:
		u.success = bool(msg)
		u.currentStep = StepDone
		return u, func() tea.Msg { return quitMsg{} }

	case quitMsg:
		return u, tea.Quit

	case tea.QuitMsg:
		// When we receive the quit message, ensure we're in the quit state
		u.currentStep = StepQuit
		return u, nil

	default:
		return u, nil
	}
}

// View renders the current step
func (u *UIManager) View() string {
	// Render inline at current cursor position; Bubble Tea replaces previous view
	vw := u.width
	if vw == 0 {
		vw = 80
	}

	switch u.currentStep {
	case StepProfileSelection:
		var body string
		if u.profileModel == nil {
			body = "Initializing profile selection..."
		} else {
			body = u.profileModel.View()
		}
		return body

	case StepDriverSelection:
		var body string
		if u.driverModel == nil {
			body = "Initializing driver selection..."
		} else {
			body = u.driverModel.View()
		}
		return body

	case StepRoleSelection:
		var body string
		if u.roleModel == nil {
			body = "Initializing role selection..."
		} else {
			body = u.roleModel.View()
		}
		return body

	case StepMFAInput:
		// Check if we're using automatic MFA or manual input
		if u.authDriverName != auth_drivers.AuthDriverManual {
			// Show status message for automatic MFA providers
			content := fmt.Sprintf("%s\n\n%s\n\n%s",
				accentStyle.Render("🔐 MFA Authentication"),
				infoStyle.Render(fmt.Sprintf("Fetching your MFA code from %s...", u.authDriverName.String())),
				pulseStyle.Render("⏳ Please wait"))
			return u.renderTextWithTitle("🔐 JJ AWS Login", content)
		} else {
			// Show manual MFA input
			content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
				infoStyle.Render("Enter your 6-digit MFA code"),
				u.mfaInput.View(),
				lightGrayStyle.Render("Press Enter to continue • Ctrl+C to cancel"),
				errorStyle.Render(u.step))
			return u.renderTextWithTitle("🔐 MFA Autentication Required", content)
		}

	case StepProcessing:
		stepMessage := u.step
		if stepMessage == "" {
			stepMessage = u.processingMessages[u.currentMessageIndex]
		}
		content := fmt.Sprintf("%s %s",
			u.spinner.View(),
			lightGrayStyle.Render(stepMessage))
		return u.renderTextWithTitle("🔐 JJ AWS Login", content)

	case StepDone:
		if u.err != nil {
			content := fmt.Sprintf("%s %s",
				errorStyle.Render("✗"),
				lightGrayStyle.Render(u.err.Error()))

			u.exitMessage = content
		}
		if u.success {
			// Format ECR status
			ecrStatus := "no"
			ecrColor := errorStyle
			if u.sessionResult.ECRAuth {
				ecrStatus = "yes"
				ecrColor = accentStyle
			}

			successLine := successStyle.Render("✓ Success")
			content := fmt.Sprintf("%s - account [%s] - ecr [%s]",
				successLine,
				accentStyle.Render(u.profile),
				ecrColor.Render(ecrStatus))

			u.exitMessage = content
		}

		return ""
	case StepQuit:
		return ""
	}

	box := lipgloss.NewStyle().
		Padding(1, 2).
		Width(min(vw-4, 80))
	return box.Render("Unknown state")
}

// renderTextWithTitle renders text content with a title that matches the list styling
func (u *UIManager) renderTextWithTitle(title, content string) string {
	vw := u.width
	if vw == 0 {
		vw = 80
	}

	// Use the same title styling as the lists
	titleStyle := lipgloss.NewStyle().MarginLeft(2).Padding(1, 2, 0, 2).Bold(true)
	styledTitle := titleStyle.Render(title)

	// Create content box with consistent padding
	contentBox := lipgloss.NewStyle().
		Padding(0, 2).
		Width(min(vw-4, 80))

	styledContent := contentBox.Render(content)

	// Join title and content with proper spacing (mimicking list layout)
	return lipgloss.JoinVertical(lipgloss.Left, styledTitle, "", styledContent)
}

func (u *UIManager) setProcessingMessage() {
	// Pick a random message
	if len(u.processingMessages) > 0 {
		u.currentMessageIndex = rand.Intn(len(u.processingMessages))
	}
}

// FinalOutput returns the final line that should be printed after the TUI exits
// to preserve the success or error message on screen.
func (u *UIManager) FinalOutput() string {
	return u.renderTextWithTitle("🔐 JJ AWS Login", u.exitMessage)
}

// initCurrentStep initializes the current step
func (u *UIManager) initCurrentStep() tea.Cmd {

	switch u.currentStep {
	case StepProfileSelection:
		profiles := u.awsService.GetValidProfiles()
		profileModel := lists.NewProfileListModel(u.awsService)
		u.profileModel = &profileModel
		if len(profiles) == 1 {
			u.profile = profiles[0]
			return func() tea.Msg { return stepCompleteMsg{step: StepProfileSelection, data: u.profile} }
		}
		return nil

	case StepDriverSelection:
		driverModel := lists.NewDriverListModel()
		u.driverModel = &driverModel
		return nil

	case StepRoleSelection:
		// Check if there are any assumable roles
		assumableRoles := u.awsService.GetAssumableRoles(u.profile)
		if len(assumableRoles) == 0 {
			// Skip role selection, go to MFA
			u.currentStep = StepMFAInput
			return u.initCurrentStep()
		}
		roleModel := lists.NewRoleListModel(u.awsService, u.profile)
		u.roleModel = &roleModel
		return nil

	case StepMFAInput:
		// Check if we can get MFA automatically
		if u.authDriverName != auth_drivers.AuthDriverManual {
			return u.tryAutoMFA()
		}
		return nil

	case StepProcessing:
		return tea.Batch(
			u.spinner.Tick,
			u.processAuthentication(),
			tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
				return processingTickMsg{}
			}),
		)

	default:
		return nil
	}
}

// handleCurrentStep handles input for the current step
func (u *UIManager) handleCurrentStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {

	switch u.currentStep {
	case StepProfileSelection:
		if u.profileModel == nil {
			return u, nil
		}

		profiles := u.awsService.GetValidProfiles()

		updatedModel, cmd := u.profileModel.Update(msg)
		*u.profileModel = updatedModel.(lists.ProfileListModel)

		if u.profileModel.IsSelected() || len(profiles) == 1 {
			u.profile = u.profileModel.GetChoice().(string)
			return u, func() tea.Msg {
				return stepCompleteMsg{step: StepProfileSelection, data: u.profile}
			}
		}
		return u, cmd

	case StepDriverSelection:
		if u.driverModel == nil {
			return u, nil
		}
		updatedModel, cmd := u.driverModel.Update(msg)
		*u.driverModel = updatedModel.(lists.DriverListModel)

		if u.driverModel.IsSelected() {
			u.authDriverName = u.driverModel.GetChoice().(auth_drivers.AuthDriverName)
			return u, func() tea.Msg {
				return stepCompleteMsg{step: StepDriverSelection, data: u.authDriverName}
			}
		}
		return u, cmd

	case StepRoleSelection:
		if u.roleModel == nil {
			return u, nil
		}
		updatedModel, cmd := u.roleModel.Update(msg)
		*u.roleModel = updatedModel.(lists.RoleListModel)

		if u.roleModel.IsSelected() {
			u.selectedRole = u.roleModel.GetChoice().(string)
			return u, func() tea.Msg {
				return stepCompleteMsg{step: StepRoleSelection, data: u.selectedRole}
			}
		}
		return u, cmd

	case StepMFAInput:
		// Only handle manual input if using manual driver
		if u.authDriverName == auth_drivers.AuthDriverManual {
			if msg.String() == "enter" {
				mfaCode := u.mfaInput.Value()
				if u.awsService.ValidateMFACode(mfaCode) {
					u.mfaCode = mfaCode
					return u, func() tea.Msg {
						return stepCompleteMsg{step: StepMFAInput, data: u.mfaCode}
					}
				} else {
					u.step = "Invalid MFA code - must be 6 digits"
					u.mfaInput.SetValue("")
					return u, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
						u.step = ""
						return nil
					})
				}
			} else {
				var cmd tea.Cmd
				u.mfaInput, cmd = u.mfaInput.Update(msg)
				return u, cmd
			}
		}
		// For automatic drivers, just ignore key input (they're handled by tryAutoMFA)
		return u, nil

	default:
		return u, nil
	}
}

// handleStepComplete handles completion of each step and advances to the next
func (u *UIManager) handleStepComplete(msg stepCompleteMsg) (tea.Model, tea.Cmd) {
	// fmt.Printf("DEBUG: handleStepComplete called with step: %v, data: %v\n", msg.step, msg.data)

	switch msg.step {
	case StepProfileSelection:
		u.currentStep = StepDriverSelection
		return u, u.initCurrentStep()

	case StepDriverSelection:
		u.currentStep = StepRoleSelection
		// Update the auth driver from the step completion data
		if driver, ok := msg.data.(auth_drivers.AuthDriverName); ok {
			u.authDriverName = driver
		}
		return u, u.initCurrentStep()

	case StepRoleSelection:
		u.currentStep = StepMFAInput
		return u, u.initCurrentStep()

	case StepMFAInput:
		u.currentStep = StepProcessing
		return u, u.initCurrentStep()

	default:
		return u, nil
	}
}

// tryAutoMFA attempts to get MFA code automatically from the driver
func (u *UIManager) tryAutoMFA() tea.Cmd {
	return func() tea.Msg {
		driver, err := auth_drivers.GetDriver(u.authDriverName, u.profile)
		if err != nil {
			return errorMsg(err)
		}

		mfaCode, err := u.awsService.GetMFACode(driver)
		if err != nil {
			return errorMsg(err)
		}

		u.mfaCode = mfaCode
		return stepCompleteMsg{step: StepMFAInput, data: mfaCode}
	}
}

// processAuthentication handles the final authentication process
func (u *UIManager) processAuthentication() tea.Cmd {
	return func() tea.Msg {
		// Get session token with MFA code
		_, err := u.awsService.GetSessionToken(u.profile, u.mfaCode)
		if err != nil {
			return errorMsg(err)
		}

		// If we have a role to assume, do that
		if u.selectedRole != "" {
			assumedProfileName := u.awsService.GetAssumedProfileName(u.selectedRole)
			_, err := u.awsService.AssumeRole(assumedProfileName, u.selectedRole)
			if err != nil {
				return errorMsg(err)
			}

			u.profile = assumedProfileName
		}

		// Set user to profile name for display purposes
		u.sessionResult.User = u.profile

		// Attempt ECR login
		if err := u.awsService.LoginToECR(); err != nil {
			u.sessionResult.ECRAuth = false
			// not critical, continue with success
		} else {
			u.sessionResult.ECRAuth = true
		}

		return doneMsg(true)
	}
}
