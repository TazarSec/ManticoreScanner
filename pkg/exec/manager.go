package exec

// WrapStrategy defines how to safely wrap a package manager install command.
// It separates the dependency resolution step (lockfile generation) from the
// actual installation, allowing security scanning between the two.
type WrapStrategy struct {
	// LockfileCmd is the command to run to generate/update the lockfile
	// without installing packages or running lifecycle scripts.
	// If nil, the lockfile is expected to already exist (e.g., npm ci).
	LockfileCmd []string

	// LockfilePath is the absolute path to the lockfile to scan.
	LockfilePath string

	// InstallCmd is the full command to run after scanning passes.
	InstallCmd []string
}

// PackageManager abstracts wrapping a package manager's install commands
// with pre-install security scanning.
type PackageManager interface {
	// Name returns the package manager name (e.g., "npm").
	Name() string

	// Plan returns a WrapStrategy for the given command arguments,
	// or nil if the command is not an install-type operation that
	// should be scanned before execution.
	//
	// args contains the arguments after the package manager binary
	// (e.g., for "npm install lodash", args is ["install", "lodash"]).
	// dir is the working directory where the command will be executed.
	Plan(args []string, dir string) *WrapStrategy
}
