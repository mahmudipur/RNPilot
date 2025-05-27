package cmd

import (
	"fmt"

	"github.com/mahmudipur/RNPilot/internal/project"
	"github.com/mahmudipur/RNPilot/internal/ui"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Initialize a new React Native project",
	Long: `This command initializes a new React Native project.
You can choose between an Expo-managed app or a bare React Native app.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get project type and project name from the user
		projectType, projectName, err := ui.PromptProjectDetails()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Create the project based on the selected type
		err = project.CreateProject(projectType, projectName)
		if err != nil {
			fmt.Printf("Error creating project: %v\n", err)
			return
		}

		fmt.Printf("Successfully created the %s project: %s\n", projectType, projectName)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
