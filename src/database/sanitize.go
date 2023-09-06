package database

import (
	"fmt"
	"regexp"
	"strings"
)

func SanitizeSchemaName(name string) (string, error) {
	// Replace spaces with underscores
	newname := strings.ReplaceAll(name, " ", "_")

	// Remove any characters that are not alphanumeric or underscores
	validName := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	newname = validName.ReplaceAllString(newname, "")

	// Truncate the name to a reasonable length if needed
	if len(newname) > 63 {
		newname = newname[:63]
	}

	if newname != name {
		return newname, fmt.Errorf("Project name is not allowed, use only alphanumerical characters")
	}
	return newname, nil
}
