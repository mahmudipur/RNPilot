package dotenv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mahmudipur/RNPilot/internal/config"
)

// DotenvFile represents the .env file operations
type DotenvFile struct {
	Path      string
	Variables map[string]string
}

// EnsureExpoPrefix ensures that Expo keys start with EXPO_PUBLIC_
func EnsureExpoPrefix(key string) string {
	if !strings.HasPrefix(key, "EXPO_PUBLIC_") {
		return "EXPO_PUBLIC_" + key
	}
	return key
}

// LoadEnvFile reads the .env file into memory
func LoadEnvFile(path string) (*DotenvFile, error) {
	envFile := &DotenvFile{
		Path:      path,
		Variables: make(map[string]string),
	}

	// Check if .env file exists, if not create it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create .env file: %v", err)
		}
		defer file.Close()
	}

	// Read the .env file line by line
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open .env file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Split key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envFile.Variables[key] = value
	}

	return envFile, nil
}

// SaveEnvFile writes the environment variables back to the .env file
func (e *DotenvFile) SaveEnvFile() error {
	file, err := os.Create(e.Path)
	if err != nil {
		return fmt.Errorf("failed to open .env file for writing: %v", err)
	}
	defer file.Close()

	// Write each key-value pair to the file
	for key, value := range e.Variables {
		_, err := file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if err != nil {
			return fmt.Errorf("failed to write key-value pair to .env: %v", err)
		}
	}

	return nil
}

// AddOrUpdateKey adds or updates a key in the .env file (ensures UPPER CASE for keys)
func (e *DotenvFile) AddOrUpdateKey(key, value string) {
	key = strings.ToUpper(key) // Ensure the key is UPPER CASE
	e.Variables[key] = value
}

// RemoveKey removes a key from the .env file
func (e *DotenvFile) RemoveKey(key string) {
	delete(e.Variables, key)
}

// ListKeys returns a list of keys currently present in the .env file
func (e *DotenvFile) ListKeys() []string {
	keys := make([]string, 0, len(e.Variables))
	for key := range e.Variables {
		keys = append(keys, key)
	}
	return keys
}

// CheckEnvInitialized checks whether the environment has been initialized by reading config
func CheckEnvInitialized(projectPath string) (bool, error) {
	cfg, err := config.LoadConfig(projectPath)
	if err != nil {
		return false, err
	}
	return cfg.EnvInitialized, nil
}

// MarkEnvInitialized updates the config to mark the environment as initialized
func MarkEnvInitialized(projectPath string) error {
	err := config.UpdateConfig(projectPath, func(cfg *config.ProjectConfig) {
		cfg.EnvInitialized = true
	})
	if err != nil {
		return fmt.Errorf("failed to mark env as initialized: %v", err)
	}
	return nil
}

// CreateEnvFiles handles initial creation of .env, env.d.ts, and Babel modifications
func CreateEnvFiles(projectPath, projectType string) error {
	// Create .env file if it doesn't exist
	envFilePath := filepath.Join(projectPath, ".env")
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		_, err := os.Create(envFilePath)
		if err != nil {
			return fmt.Errorf("failed to create .env file: %v", err)
		}
	}

	// If Bare, set up react-native-dotenv and env.d.ts
	if projectType == "bare" {
		err := installReactNativeDotenv(projectPath)
		if err != nil {
			return err
		}

		err = modifyBabelConfig(projectPath)
		if err != nil {
			return err
		}

		// Create env.d.ts
		err = UpdateEnvDTS(projectPath, "", "", false)
		if err != nil {
			return fmt.Errorf("failed to create env.d.ts: %v", err)
		}
	}

	// Mark environment as initialized
	err := MarkEnvInitialized(projectPath)
	if err != nil {
		return err
	}

	fmt.Println("Environment successfully initialized.")
	return nil
}

// EnsureCorrectEnvDTSStructure ensures the env.d.ts file is properly initialized
func EnsureCorrectEnvDTSStructure(envPath string) error {
	envDTSPath := filepath.Join(envPath, "env.d.ts")

	// If the file does not exist, create it with the correct structure
	if _, err := os.Stat(envDTSPath); os.IsNotExist(err) {
		file, err := os.Create(envDTSPath)
		if err != nil {
			return fmt.Errorf("failed to create env.d.ts file: %v", err)
		}
		defer file.Close()

		// Write the initial structure
		_, err = file.WriteString(`declare module "@env" {
}
`)
		if err != nil {
			return fmt.Errorf("failed to initialize env.d.ts: %v", err)
		}
	}

	// Check if the file is malformed or missing the correct structure
	content, err := os.ReadFile(envDTSPath)
	if err != nil {
		return fmt.Errorf("failed to read env.d.ts file: %v", err)
	}

	if !strings.Contains(string(content), `declare module "@env" {`) {
		// The file is corrupted or improperly formatted, rewrite it
		err := os.WriteFile(envDTSPath, []byte(`declare module "@env" {
}
`), 0644)
		if err != nil {
			return fmt.Errorf("failed to reset env.d.ts structure: %v", err)
		}
	}

	return nil
}

// UpdateEnvDTS updates the env.d.ts file with new or updated environment variables
func UpdateEnvDTS(envPath string, key, value string, isRemove bool) error {
	envDTSPath := filepath.Join(envPath, "env.d.ts")

	// Ensure the file is correctly initialized
	err := EnsureCorrectEnvDTSStructure(envPath)
	if err != nil {
		return err
	}

	// Read the content of the file into memory
	file, err := os.OpenFile(envDTSPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open env.d.ts: %v", err)
	}
	defer file.Close()

	var content []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	// Process adding or removing keys
	if isRemove {
		// Remove the key
		for i, line := range content {
			if strings.Contains(line, fmt.Sprintf("export const %s", key)) {
				content = append(content[:i], content[i+1:]...)
				break
			}
		}
	} else {
		// Add or update the key
		found := false
		for i, line := range content {
			if strings.Contains(line, fmt.Sprintf("export const %s", key)) {
				// Update the existing key
				content[i] = fmt.Sprintf("  export const %s: string;", strings.ToUpper(key))
				found = true
				break
			}
		}
		if !found {
			// Insert the new key before the closing brace
			if key != "" {
				content = append(content[:len(content)-1], fmt.Sprintf("  export const %s: string;", strings.ToUpper(key)), "}")
			}
		}
	}

	// Write the updated content back to the file
	err = os.WriteFile(envDTSPath, []byte(strings.Join(content, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated env.d.ts file: %v", err)
	}

	return nil
}

// Install react-native-dotenv for Bare React Native projects
func installReactNativeDotenv(projectPath string) error {
	cmd := exec.Command("npm", "install", "--save-dev", "react-native-dotenv")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Installing react-native-dotenv...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install react-native-dotenv: %v", err)
	}
	return nil
}

// modifyBabelConfig modifies babel.config.js or .babelrc to include the react-native-dotenv plugin
func modifyBabelConfig(projectPath string) error {
	configFilePaths := []string{"babel.config.js", ".babelrc"}

	for _, fileName := range configFilePaths {
		filePath := filepath.Join(projectPath, fileName)

		if _, err := os.Stat(filePath); err == nil {
			if strings.HasSuffix(fileName, ".js") {
				return modifyBabelJSConfig(filePath)
			} else if strings.HasSuffix(fileName, ".babelrc") {
				return modifyBabelJSONConfig(filePath)
			}
		}
	}
	return fmt.Errorf("babel config file not found in project root")
}

// modifyBabelJSConfig modifies babel.config.js to include the react-native-dotenv plugin
func modifyBabelJSConfig(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", filePath, err)
	}

	pluginConfig := `
  ['module:react-native-dotenv', {
    moduleName: '@env',
    blocklist: null,
    allowlist: null,
    safe: false,
    allowUndefined: false,
    verbose: false,
  }],
`

	// Check if the file already contains react-native-dotenv plugin
	if strings.Contains(string(content), "react-native-dotenv") {
		fmt.Printf("react-native-dotenv plugin already present in %s\n", filePath)
		return nil
	}

	// Handle the case where babel.config.js is a function returning an object
	if strings.Contains(string(content), "module.exports = function") {
		return modifyBabelFunctionJSConfig(filePath, string(content), pluginConfig)
	}

	// Handle the case where babel.config.js directly exports an object
	return modifyBabelObjectJSConfig(filePath, string(content), pluginConfig)
}

// modifyBabelFunctionJSConfig modifies a babel.config.js function to add the react-native-dotenv plugin
func modifyBabelFunctionJSConfig(filePath, content, pluginConfig string) error {
	// We need to locate the return statement and modify the object returned by the function
	regex := regexp.MustCompile(`(return\s*\{)`)
	updatedContent := regex.ReplaceAllString(content, "$1\n  plugins: ["+pluginConfig+"],")
	err := os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to update function-based babel.config.js: %v", err)
	}

	fmt.Printf("Successfully added react-native-dotenv plugin to %s\n", filePath)
	return nil
}

// modifyBabelObjectJSConfig modifies a babel.config.js that exports an object directly
func modifyBabelObjectJSConfig(filePath, content, pluginConfig string) error {
	// If plugins array does not exist, add it after the presets key if it exists
	if !strings.Contains(content, "plugins") {
		if strings.Contains(content, "presets") {
			// Insert plugins after presets
			regex := regexp.MustCompile(`(presets:\s*\[.*?\],)`)
			updatedContent := regex.ReplaceAllString(content, "$1\n  plugins: ["+pluginConfig+"],")
			return os.WriteFile(filePath, []byte(updatedContent), 0644)
		} else {
			// If no presets, just add plugins at the top level
			regex := regexp.MustCompile(`(module\.exports\s*=\s*\{)`)
			updatedContent := regex.ReplaceAllString(content, "$1\n  plugins: ["+pluginConfig+"],")
			return os.WriteFile(filePath, []byte(updatedContent), 0644)
		}
	}

	// If plugins array exists, append the plugin
	regex := regexp.MustCompile(`(plugins:\s*\[)`)
	updatedContent := regex.ReplaceAllString(content, `$1`+pluginConfig)
	err := os.WriteFile(filePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to update babel.config.js: %v", err)
	}

	fmt.Printf("Successfully added react-native-dotenv plugin to %s\n", filePath)
	return nil
}

// modifyBabelJSONConfig modifies a .babelrc file (JSON) to include the react-native-dotenv plugin
func modifyBabelJSONConfig(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", filePath, err)
	}

	// Parse the JSON content
	var babelConfig map[string]interface{}
	if err := json.Unmarshal(content, &babelConfig); err != nil {
		return fmt.Errorf("failed to parse JSON in %s: %v", filePath, err)
	}

	// Check if the plugins array exists
	plugins, ok := babelConfig["plugins"].([]interface{})
	if !ok {
		// If not, create the plugins array
		plugins = []interface{}{}
	}

	// Check if react-native-dotenv is already in the plugins
	for _, plugin := range plugins {
		if pluginStr, isString := plugin.(string); isString && strings.Contains(pluginStr, "react-native-dotenv") {
			fmt.Printf("react-native-dotenv plugin already present in %s\n", filePath)
			return nil
		}
	}

	// Add react-native-dotenv to the plugins
	dotenvPlugin := map[string]interface{}{
		"module":         "react-native-dotenv",
		"moduleName":     "@env",
		"blocklist":      nil,
		"allowlist":      nil,
		"safe":           false,
		"allowUndefined": false,
		"verbose":        false,
	}

	plugins = append(plugins, dotenvPlugin)
	babelConfig["plugins"] = plugins

	// Write the updated JSON back to the file
	updatedContent, err := json.MarshalIndent(babelConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize updated JSON for %s: %v", filePath, err)
	}

	err = os.WriteFile(filePath, updatedContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated JSON to %s: %v", filePath, err)
	}

	fmt.Printf("Successfully added react-native-dotenv plugin to %s\n", filePath)
	return nil
}

// DeleteFile removes the specified file from the project
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist, no need to delete
		return nil
	}

	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %v", filePath, err)
	}
	fmt.Printf("Deleted file: %s\n", filePath)
	return nil
}

// DestroyEnvFiles deletes the .env and env.d.ts (if applicable) files
func DestroyEnvFiles(projectPath, projectType string) error {
	// Delete the .env file
	envFilePath := filepath.Join(projectPath, ".env")
	err := DeleteFile(envFilePath)
	if err != nil {
		return err
	}

	// If Bare React Native, also delete env.d.ts
	if projectType == "bare" {
		envDTSPath := filepath.Join(projectPath, "env.d.ts")
		err = DeleteFile(envDTSPath)
		if err != nil {
			return err
		}
	}

	// Update the config to mark EnvInitialized as false
	err = config.UpdateConfig(projectPath, func(cfg *config.ProjectConfig) {
		cfg.EnvInitialized = false
	})
	if err != nil {
		return fmt.Errorf("failed to update project config: %v", err)
	}

	fmt.Println("Environment configuration destroyed successfully.")
	return nil
}
