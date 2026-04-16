package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/etsubu/manticore-scanner/pkg/api"
	"github.com/etsubu/manticore-scanner/pkg/formatter"
	"github.com/etsubu/manticore-scanner/pkg/scanner"
)

// Config holds configuration for the exec command.
type Config struct {
	// Command is the full command to wrap (e.g., ["npm", "install"]).
	Command []string

	// Dir is the working directory for the command.
	Dir string

	// ScanConfig holds the scanning configuration (API key, URL, timeout, etc.).
	ScanConfig scanner.Config

	// FailOn is the suspicion score threshold. Packages scoring at or above
	// this value will cause the install to be blocked.
	FailOn float64

	// Quiet suppresses manticore's own progress output.
	Quiet bool
}

// Run wraps a package manager install command with pre-install security scanning.
//
// The flow is:
//  1. Detect the package manager from the command
//  2. Generate/update the lockfile without installing packages or running scripts
//  3. Scan all dependencies via the AegisEngine backend
//  4. If any package exceeds the suspicion threshold, abort
//  5. Otherwise, run the actual install command
//
// Returns the process exit code and any error. A non-zero exit code with a nil
// error means the wrapped command (or scan gate) exited cleanly but unsuccessfully.
func Run(ctx context.Context, cfg Config) (int, error) {
	log := func(format string, args ...any) {
		if !cfg.Quiet {
			fmt.Fprintf(os.Stderr, "[manticore] "+format+"\n", args...)
		}
	}

	if len(cfg.Command) == 0 {
		return 1, fmt.Errorf("no command specified")
	}

	// Step 1: Detect package manager.
	pm := Detect(cfg.Command[0])
	if pm == nil {
		return 1, fmt.Errorf("unsupported package manager: %s (supported: npm)", cfg.Command[0])
	}

	pmArgs := cfg.Command[1:]
	strategy := pm.Plan(pmArgs, cfg.Dir)
	if strategy == nil {
		log("Not an install command, running directly without scanning")
		return runPassthrough(ctx, cfg.Command, cfg.Dir)
	}

	log("Detected package manager: %s", pm.Name())

	// Step 2: Generate lockfile without installing packages.
	if strategy.LockfileCmd != nil {
		log("Resolving dependencies (lockfile-only)...")
		exitCode, err := runPassthrough(ctx, strategy.LockfileCmd, cfg.Dir)
		if err != nil {
			return exitCode, fmt.Errorf("failed to resolve dependencies: %w", err)
		}
		if exitCode != 0 {
			return exitCode, fmt.Errorf("dependency resolution exited with code %d", exitCode)
		}
	}

	// Step 3: Verify lockfile exists.
	if _, err := os.Stat(strategy.LockfilePath); os.IsNotExist(err) {
		return 1, fmt.Errorf("lockfile not found at %s — cannot scan dependencies", strategy.LockfilePath)
	}

	// Step 4: Scan dependencies.
	log("Scanning dependencies...")
	scanCfg := cfg.ScanConfig
	scanCfg.InputPath = strategy.LockfilePath

	var progressFn func(completed, total int)
	if !cfg.Quiet {
		progressFn = func(completed, total int) {
			fmt.Fprintf(os.Stderr, "\r[manticore] Scanning: %d/%d packages completed", completed, total)
		}
	}

	result, err := scanner.Run(ctx, scanCfg, progressFn)
	if err != nil {
		return 1, fmt.Errorf("scan failed: %w", err)
	}

	if !cfg.Quiet && progressFn != nil {
		fmt.Fprintln(os.Stderr)
	}

	if len(result.Items) == 0 {
		log("No packages found to scan")
		log("Running: %s", strings.Join(strategy.InstallCmd, " "))
		return runPassthrough(ctx, strategy.InstallCmd, cfg.Dir)
	}

	// Step 5: Evaluate results against threshold.
	var blocked []api.BatchResultItem
	for _, item := range result.Items {
		if item.Profile != nil && item.Profile.SuspicionScore >= cfg.FailOn {
			blocked = append(blocked, item)
		}
	}

	if len(blocked) > 0 {
		// Show scan results for blocked packages.
		fmtr := formatter.Get(scanCfg.Format)
		output, fmtErr := fmtr.Format(result.Items, formatter.Options{InputFile: result.InputFile})
		if fmtErr == nil {
			fmt.Fprintln(os.Stderr)
			os.Stdout.Write(output)
		}
		fmt.Fprintln(os.Stderr)
		log("Blocked %d package(s) exceeding suspicion threshold (%.0f)", len(blocked), cfg.FailOn)
		log("Aborting install. Review the packages above before proceeding.")
		return 1, nil
	}

	log("All %d packages passed security scan (threshold: %.0f)", len(result.Items), cfg.FailOn)

	// Step 6: Run the actual install command.
	fmt.Fprintln(os.Stderr)
	log("Running: %s", strings.Join(strategy.InstallCmd, " "))
	return runPassthrough(ctx, strategy.InstallCmd, cfg.Dir)
}

// runPassthrough executes a command with stdin/stdout/stderr connected to the
// parent process. Returns the exit code and any execution error.
func runPassthrough(ctx context.Context, args []string, dir string) (int, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}
