package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// createHookCmd represents the create-hook command
var createHookCmd = &cobra.Command{
	Use:   "create-hook [name] [directory]",
	Short: "Generate a custom React Native hook",
	Long: `Generate a new custom React Native hook with the provided name
and place it in the specified directory. If no directory is provided, the default is ./src/lib/hooks.`,
	Args: cobra.MinimumNArgs(1), // At least the hook name is required
	Run: func(cmd *cobra.Command, args []string) {
		hookName := args[0]
		directory := "./src/lib/hooks" // Default directory
		if len(args) > 1 {
			directory = args[1]
		}

		// Ensure the hook name starts with 'use'
		hookName = ensureUsePrefix(hookName)

		// Create the hook file
		err := createHookFile(hookName, directory)
		if err != nil {
			fmt.Printf("Error creating hook: %v\n", err)
			os.Exit(1)
		}

		// Create the corresponding type file
		err = createTypeFile(hookName)
		if err != nil {
			fmt.Printf("Error creating type file: %v\n", err)
			os.Exit(1)
		}

		// Update the hook barrel file (index.ts)
		err = updateHookBarrelFile(directory, hookName)
		if err != nil {
			fmt.Printf("Error updating hook barrel file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created hook %s in %s\n", hookName, directory)
	},
}

func init() {
	rootCmd.AddCommand(createHookCmd)
}

// ensureUsePrefix adds the 'use' prefix if not already present
func ensureUsePrefix(hookName string) string {
	if !strings.HasPrefix(hookName, "use") {
		return "use" + strings.Title(hookName)
	}
	return hookName
}

// toUpperCamelCase converts a string like 'useCustomHook' to 'UseCustomHook'
func toUpperCamelCase(s string) string {
	return strings.Title(s)
}

// createHookFile generates the hook file in the given directory
func createHookFile(hookName, directory string) error {
	// Ensure the directory exists
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return err
	}

	// Define the file path (TSX if JSX is needed, otherwise TS)
	filePath := filepath.Join(directory, fmt.Sprintf("%s.tsx", hookName))

	// Convert the hook name to UpperCamelCase
	hookTypeName := toUpperCamelCase(hookName)

	// Define the content of the hook file
	content := fmt.Sprintf(`import { I%s } from "@src/lib/types/hooks";

export const %s: I%s = () => {
	// Your hook logic here
	return {};
};
`, hookTypeName, hookName, hookTypeName)

	// Write the content to the file
	return os.WriteFile(filePath, []byte(content), 0644)
}

// createTypeFile generates the type file for the hook
func createTypeFile(hookName string) error {
	// Convert hook name to UpperCamelCase for type names
	hookTypeName := toUpperCamelCase(hookName)

	// Ensure the types directory exists
	typesDirectory := "./src/lib/types/hooks"
	if err := os.MkdirAll(typesDirectory, os.ModePerm); err != nil {
		return err
	}

	// Define the type file path
	typeFilePath := filepath.Join(typesDirectory, fmt.Sprintf("%s.type.ts", hookName))

	// Define the content of the type file
	typeFileContent := fmt.Sprintf(`interface I%sProps {}
interface I%sReturnValue {}

export type I%s = ({}: I%sProps) => I%sReturnValue;
`, hookTypeName, hookTypeName, hookTypeName, hookTypeName, hookTypeName)

	// Write the content to the type file
	if err := os.WriteFile(typeFilePath, []byte(typeFileContent), 0644); err != nil {
		return err
	}

	// Update the type barrel file
	return updateTypeBarrelFile(hookName)
}

// updateHookBarrelFile updates (or creates) the index.ts file to export the new hook
func updateHookBarrelFile(directory, hookName string) error {
	barrelFilePath := filepath.Join(directory, "index.ts")

	// Open or create the barrel file
	f, err := os.OpenFile(barrelFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the export statement to the barrel file
	exportStatement := fmt.Sprintf("export * from \"./%s\";\n", hookName)
	_, err = f.WriteString(exportStatement)
	return err
}

// updateTypeBarrelFile updates (or creates) the index.ts file for the types in ./src/lib/types/hooks/
func updateTypeBarrelFile(hookName string) error {
	barrelFilePath := "./src/lib/types/hooks/index.ts"

	// Open or create the barrel file
	f, err := os.OpenFile(barrelFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the export statement to the types barrel file
	exportStatement := fmt.Sprintf("export * from \"./%s.type\";\n", hookName)
	_, err = f.WriteString(exportStatement)
	return err
}
