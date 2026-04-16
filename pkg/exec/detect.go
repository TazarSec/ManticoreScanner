package exec

import "path/filepath"

// Detect identifies the package manager from the command name and returns
// the appropriate PackageManager implementation.
// Returns nil if the command is not a recognized package manager.
func Detect(command string) PackageManager {
	base := filepath.Base(command)
	switch base {
	case "npm":
		return &NPM{}
	default:
		return nil
	}
}
