package utils

import (
	"errors"
	"regexp"
)

// ValidateProjectName validates the project name to ensure it follows naming conventions.
func ValidateProjectName(val interface{}) error {
	name, ok := val.(string)
	if !ok {
		return errors.New("invalid type of input")
	}

	// Validate project name (should start with a letter and contain only alphanumeric and dashes)
	match, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_-]+$`, name)
	if !match {
		return errors.New("project name must start with a letter and contain only alphanumeric characters, dashes, or underscores")
	}
	return nil
}
