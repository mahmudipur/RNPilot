package ui

import (
	"fmt"

	"github.com/mahmudipur/RNPilot/internal/utils"

	"github.com/AlecAivazis/survey/v2"
)

// PromptProjectDetails prompts the user to select the project type and input a project name.
func PromptProjectDetails() (string, string, error) {
	// Define options for project types
	projectTypes := []string{"Expo-managed app", "Bare React Native app"}

	// Prompt for project type selection
	var projectType string
	prompt := &survey.Select{
		Message: "Select the type of React Native project:",
		Options: projectTypes,
	}
	err := survey.AskOne(prompt, &projectType)
	if err != nil {
		return "", "", fmt.Errorf("failed to get project type: %v", err)
	}

	// Prompt for project name
	var projectName string
	promptName := &survey.Input{
		Message: "Enter the name of your project:",
	}
	err = survey.AskOne(promptName, &projectName, survey.WithValidator(utils.ValidateProjectName))
	if err != nil {
		return "", "", fmt.Errorf("failed to get project name: %v", err)
	}

	// Map project type to internal representation
	if projectType == "Expo-managed app" {
		return "expo", projectName, nil
	} else if projectType == "Bare React Native app" {
		return "bare", projectName, nil
	}
	return "", "", fmt.Errorf("invalid project type selected")
}

// PromptEnvOperation prompts the user to select an operation (add, update, remove)
func PromptEnvOperation() (string, error) {
	var operation string
	prompt := &survey.Select{
		Message: "What do you want to do?",
		Options: []string{"add", "update", "remove"},
	}
	err := survey.AskOne(prompt, &operation)
	return operation, err
}

// PromptEnvKeyValue prompts the user to enter a key-value pair for the .env file
func PromptEnvKeyValue(operation string) (string, string, error) {
	var key, value string
	keyPrompt := &survey.Input{
		Message: "Enter the environment variable key:",
	}
	err := survey.AskOne(keyPrompt, &key)
	if err != nil {
		return "", "", err
	}

	if operation != "remove" {
		valuePrompt := &survey.Input{
			Message: "Enter the environment variable value:",
		}
		err = survey.AskOne(valuePrompt, &value)
		if err != nil {
			return "", "", err
		}
	}

	return key, value, nil
}

// PromptSelectKey prompts the user to select a key to update
func PromptSelectKey(keys []string) (string, error) {
	var selectedKey string
	prompt := &survey.Select{
		Message: "Select the environment variable key to update:",
		Options: keys,
	}
	err := survey.AskOne(prompt, &selectedKey)
	return selectedKey, err
}

// PromptNewValue prompts the user to input the new value for the selected key
func PromptNewValue() (string, error) {
	var value string
	prompt := &survey.Input{
		Message: "Enter the new value:",
	}
	err := survey.AskOne(prompt, &value)
	return value, err
}
