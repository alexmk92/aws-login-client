package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/alexmk92/aws-login/core"
	"github.com/alexmk92/aws-login/core/auth_drivers"
	"github.com/alexmk92/aws-login/ui"
)

// It's nice to keep the main file as lean as possible, use this to set up things like
// logging, singleton invocations, environment variables, etc.
func main() {
	// Set up logger with JJ branding
	log.SetReportTimestamp(false)
	log.SetPrefix("j&j-aws-login")

	// Default to unknown driver, this means the UI layer will prompt the user to select a driver
	// unless the AWS_LOGIN_AUTH_DRIVER environment variable is set, at which point it will be
	var authDriverName auth_drivers.AuthDriverName = auth_drivers.AuthDriverUnknown

	// Just demonstrating how you can assign a variable AND assert on it immediately
	// inside of an if statement
	if driverStr := os.Getenv("AWS_LOGIN_AUTH_DRIVER"); driverStr != "" {
		// Again with our inline assertion so we can assign the driver with the safely parsed type
		if driver, err := auth_drivers.ParseAuthDriver(driverStr); err == nil {
			authDriverName = driver
		}
	}

	// Arg1 is the command name, arg 2 is our profile name
	// when we pass this to the UI layer, it will be
	// initialised to an empty string if no profile is provided
	// this is because Go will default to the most 'sane' value
	// for the specified type, unless it is a pointer, in
	// which case it will be nil (careful as nils can cause
	// nil pointer dereference exceptions if not handled properly)
	var profile string
	if len(os.Args) >= 2 {
		profile = os.Args[1]
		// os.Setenv is only valid for the current process context, it is NOT setting
		// host environment variables
		os.Setenv("AWS_PROFILE", profile)
	}

	// Create the core AWS service to be consumed by the UI manager
	awsService := core.NewAWSService()

	// Create the UI manager for tea to consume: https://github.com/charmbracelet/bubbletea
	uiManager := ui.Start(profile, awsService, authDriverName)
	// Now, delegate tea to utilize our uiManager
	p := tea.NewProgram(uiManager)
	if _, err := p.Run(); err != nil {
		log.Fatal("Error AWS login", "error", err)
	}
}
