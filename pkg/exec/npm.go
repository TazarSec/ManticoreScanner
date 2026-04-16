package exec

import "path/filepath"

// NPM implements PackageManager for the npm package manager.
type NPM struct{}

func (n *NPM) Name() string { return "npm" }

func (n *NPM) Plan(args []string, dir string) *WrapStrategy {
	if len(args) == 0 {
		return nil
	}

	lockfilePath := filepath.Join(dir, "package-lock.json")

	switch args[0] {
	case "install", "i", "add":
		// Generate the lockfile without installing packages or running scripts.
		// --package-lock-only: only updates package-lock.json, skips node_modules
		// --ignore-scripts:    defense-in-depth against lifecycle script execution
		lockfileArgs := []string{"npm", args[0], "--package-lock-only", "--ignore-scripts"}
		lockfileArgs = append(lockfileArgs, args[1:]...)

		installArgs := make([]string, 0, len(args)+1)
		installArgs = append(installArgs, "npm")
		installArgs = append(installArgs, args...)

		return &WrapStrategy{
			LockfileCmd:  lockfileArgs,
			LockfilePath: lockfilePath,
			InstallCmd:   installArgs,
		}

	case "ci":
		// npm ci requires an existing lockfile; no generation step needed.
		installArgs := make([]string, 0, len(args)+1)
		installArgs = append(installArgs, "npm")
		installArgs = append(installArgs, args...)

		return &WrapStrategy{
			LockfileCmd:  nil,
			LockfilePath: lockfilePath,
			InstallCmd:   installArgs,
		}

	default:
		return nil
	}
}
