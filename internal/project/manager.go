package project

import (
	"RNPilot/internal/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CreateProject creates a React Native project based on the provided project type and name.
func CreateProject(projectType, projectName string) error {
	switch projectType {
	case "expo":
		return createExpoApp(projectName)
	case "bare":
		return createBareReactNativeApp(projectName)
	default:
		return fmt.Errorf("unknown project type: %s", projectType)
	}
}

// createExpoApp initializes an Expo app.
func createExpoApp(projectName string) error {
	cmd := exec.Command("npx", "create-expo-app@latest", projectName, "--template")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Creating an Expo-managed app...\n")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create Expo app: %v", err)
	}

	// Save the project type in the config file
	projectPath := filepath.Join(".", projectName)

	err := config.InitConfig("expo", projectPath)

	if err != nil {
		return fmt.Errorf("failed to save project config: %v", err)
	}

	return nil
}

// createBareReactNativeApp initializes a bare React Native app.
func createBareReactNativeApp(projectName string) error {
	// TODO: fix yarn for installing dependencies
	cmd := exec.Command("npx", "@react-native-community/cli@latest", "init", projectName, "--pm", "npm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Creating a bare React Native app...\n")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create React Native app: %v", err)
	}

	// Save the project type in the config file
	projectPath := filepath.Join(".", projectName)

	err := config.InitConfig("bare", projectPath)

	if err != nil {
		return fmt.Errorf("failed to save project config: %v", err)
	}

	return nil
}
