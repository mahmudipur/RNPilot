package cmd

import (
	"RNPilot/internal/config"
	"RNPilot/internal/dotenv"
	"RNPilot/internal/ui"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// envCmd represents the environment configuration command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables for the project",
	Long:  `Allows you to initialize, add, update, or remove environment variables for both Expo and Bare React Native projects.`,
}

// envInitCmd handles initializing the environment configuration
var envInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the environment configuration for the project",
	Long:  `Initializes the .env file, env.d.ts (if necessary), and sets up dotenv support for the project.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current project path
		projectPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		// Load the project configuration
		projectConfig, err := config.LoadConfig(projectPath)
		if err != nil {
			fmt.Printf("Error loading project config: %v\n", err)
			return
		}

		// Check if the environment has already been initialized
		envInitialized, err := dotenv.CheckEnvInitialized(projectPath)
		if err != nil {
			fmt.Printf("Error checking env initialization: %v\n", err)
			return
		}
		if envInitialized {
			fmt.Println("Environment has already been initialized.")
			return
		}

		// Initialize the environment configuration
		err = dotenv.CreateEnvFiles(projectPath, projectConfig.ProjectType)
		if err != nil {
			fmt.Printf("Error initializing environment configuration: %v\n", err)
			return
		}
	},
}

// envAddCmd represents adding new environment variables
var envAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new environment variable",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current project path
		projectPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		// Ensure the environment has been initialized
		envInitialized, err := dotenv.CheckEnvInitialized(projectPath)
		if err != nil {
			fmt.Printf("Error checking env initialization: %v\n", err)
			return
		}
		if !envInitialized {
			fmt.Println("Environment is not initialized. Run 'env init' first.")
			return
		}

		// Add new environment variable
		projectConfig, err := config.LoadConfig(projectPath)
		if err != nil {
			fmt.Printf("Error loading project config: %v\n", err)
			return
		}

		// Prompt for new key-value pair
		key, value, err := ui.PromptEnvKeyValue("add")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Handle Expo-specific logic
		if projectConfig.ProjectType == "expo" {
			key = dotenv.EnsureExpoPrefix(key) // Ensure Expo keys have EXPO_PUBLIC_ prefix
		}

		// Handle Expo or Bare
		envFilePath := filepath.Join(projectPath, ".env")
		envFile, err := dotenv.LoadEnvFile(envFilePath)
		if err != nil {
			fmt.Printf("Error loading .env file: %v\n", err)
			return
		}

		envFile.AddOrUpdateKey(key, value)

		if projectConfig.ProjectType == "bare" {
			err = dotenv.UpdateEnvDTS(projectPath, key, value, false)
			if err != nil {
				fmt.Printf("Error updating env.d.ts: %v\n", err)
				return
			}
		}

		// Save changes to .env file
		err = envFile.SaveEnvFile()
		if err != nil {
			fmt.Printf("Error saving .env file: %v\n", err)
			return
		}

		fmt.Println("Environment variable added successfully.")
	},
}

// envUpdateCmd represents updating an existing environment variable
var envUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing environment variable",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current project path
		projectPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		// Ensure the environment has been initialized
		envInitialized, err := dotenv.CheckEnvInitialized(projectPath)
		if err != nil {
			fmt.Printf("Error checking env initialization: %v\n", err)
			return
		}
		if !envInitialized {
			fmt.Println("Environment is not initialized. Run 'env init' first.")
			return
		}

		// Update environment variable
		projectConfig, err := config.LoadConfig(projectPath)
		if err != nil {
			fmt.Printf("Error loading project config: %v\n", err)
			return
		}

		// Load .env file
		envFilePath := filepath.Join(projectPath, ".env")
		envFile, err := dotenv.LoadEnvFile(envFilePath)
		if err != nil {
			fmt.Printf("Error loading .env file: %v\n", err)
			return
		}

		// List existing keys and prompt for selection
		existingKeys := envFile.ListKeys()
		if len(existingKeys) == 0 {
			fmt.Println("No environment variables found to update.")
			return
		}

		key, err := ui.PromptSelectKey(existingKeys)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Prompt for new value
		value, err := ui.PromptNewValue()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if projectConfig.ProjectType == "expo" {
			key = dotenv.EnsureExpoPrefix(key) // Ensure Expo keys have EXPO_PUBLIC_ prefix
		}

		envFile.AddOrUpdateKey(key, value)

		if projectConfig.ProjectType == "bare" {
			err = dotenv.UpdateEnvDTS(projectPath, key, value, false)
			if err != nil {
				fmt.Printf("Error updating env.d.ts: %v\n", err)
				return
			}
		}

		// Save changes to .env file
		err = envFile.SaveEnvFile()
		if err != nil {
			fmt.Printf("Error saving .env file: %v\n", err)
			return
		}

		fmt.Println("Environment variable updated successfully.")
	},
}

// envRemoveCmd represents removing an existing environment variable
var envRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an existing environment variable",
	Run: func(cmd *cobra.Command, args []string) {
		// Get current project path
		projectPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		// Ensure the environment has been initialized
		envInitialized, err := dotenv.CheckEnvInitialized(projectPath)
		if err != nil {
			fmt.Printf("Error checking env initialization: %v\n", err)
			return
		}
		if !envInitialized {
			fmt.Println("Environment is not initialized. Run 'env init' first.")
			return
		}

		// Remove environment variable
		projectConfig, err := config.LoadConfig(projectPath)
		if err != nil {
			fmt.Printf("Error loading project config: %v\n", err)
			return
		}

		// Load .env file
		envFilePath := filepath.Join(projectPath, ".env")
		envFile, err := dotenv.LoadEnvFile(envFilePath)
		if err != nil {
			fmt.Printf("Error loading .env file: %v\n", err)
			return
		}

		// List existing keys and prompt for selection
		existingKeys := envFile.ListKeys()
		if len(existingKeys) == 0 {
			fmt.Println("No environment variables found to remove.")
			return
		}

		key, err := ui.PromptSelectKey(existingKeys)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if projectConfig.ProjectType == "expo" {
			key = dotenv.EnsureExpoPrefix(key) // Ensure Expo keys have EXPO_PUBLIC_ prefix
		}

		envFile.RemoveKey(key)

		if projectConfig.ProjectType == "bare" {
			err = dotenv.UpdateEnvDTS(projectPath, key, "", true)
			if err != nil {
				fmt.Printf("Error updating env.d.ts: %v\n", err)
				return
			}
		}

		// Save changes to .env file
		err = envFile.SaveEnvFile()
		if err != nil {
			fmt.Printf("Error saving .env file: %v\n", err)
			return
		}

		fmt.Println("Environment variable removed successfully.")
	},
}

// envDestroyCmd handles uninitializing the environment configuration
var envDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy the environment configuration",
	Long: `Removes the .env and env.d.ts (for Bare React Native) files, 
and updates the project configuration to mark the environment as uninitialized.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current project path
		projectPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		// Load the project configuration
		projectConfig, err := config.LoadConfig(projectPath)
		if err != nil {
			fmt.Printf("Error loading project config: %v\n", err)
			return
		}

		// Check if the environment has already been initialized
		if !projectConfig.EnvInitialized {
			fmt.Println("Environment has not been initialized.")
			return
		}

		// Destroy the environment configuration
		err = dotenv.DestroyEnvFiles(projectPath, projectConfig.ProjectType)
		if err != nil {
			fmt.Printf("Error destroying environment configuration: %v\n", err)
			return
		}
	},
}

func init() {
	envCmd.AddCommand(envInitCmd)
	envCmd.AddCommand(envAddCmd)
	envCmd.AddCommand(envUpdateCmd)
	envCmd.AddCommand(envRemoveCmd)
	envCmd.AddCommand(envDestroyCmd)
	rootCmd.AddCommand(envCmd)
}
