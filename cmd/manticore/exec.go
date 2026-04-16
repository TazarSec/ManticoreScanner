package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/etsubu/manticore-scanner/internal/config"
	execpkg "github.com/etsubu/manticore-scanner/pkg/exec"
)

const defaultExecFailOn = 70.0

var execCmd = &cobra.Command{
	Use:   "exec [flags] -- <command> [args...]",
	Short: "Wrap a package manager command with pre-install security scanning",
	Long: `Wrap a package manager install command with pre-install security scanning.

The exec command intercepts package manager install operations, generates a
lockfile without installing packages or running lifecycle scripts, scans all
dependencies for security risks, and only proceeds with the actual installation
if the scan passes.

This prevents malicious packages from executing postinstall hooks before they
have been scanned.

Supported package managers: npm

Examples:
  manticore exec -- npm install
  manticore exec -- npm ci
  manticore exec -- npm install lodash
  manticore exec --fail-on 50 -- npm install`,
	Args: cobra.MinimumNArgs(1),
	RunE: runExec,
}

func init() {
	f := execCmd.Flags()
	f.String("api-key", "", "API key (or set MANTICORE_API_KEY)")
	f.String("api-url", "", "Backend API base URL (or set MANTICORE_API_URL)")
	f.String("format", "", "Output format for scan results: table, json, sarif (default: table)")
	f.Float64("fail-on", defaultExecFailOn, "Block install if any package suspicion score >= threshold")
	f.Int("timeout", 0, "Scan polling timeout in seconds (default: 300)")
	f.Bool("production", false, "Skip devDependencies")
	f.Bool("quiet", false, "Suppress manticore progress output")
	f.Bool("verbose", false, "Verbose logging")
}

func runExec(cmd *cobra.Command, args []string) error {
	f := cmd.Flags()

	apiKey, _ := f.GetString("api-key")
	apiURL, _ := f.GetString("api-url")
	format, _ := f.GetString("format")
	failOn, _ := f.GetFloat64("fail-on")
	timeout, _ := f.GetInt("timeout")
	production, _ := f.GetBool("production")
	quiet, _ := f.GetBool("quiet")
	verbose, _ := f.GetBool("verbose")

	// Resolve scanner config through existing infrastructure.
	cliFlags := config.CLIFlags{
		APIKey:     apiKey,
		APIURL:     apiURL,
		Format:     format,
		Timeout:    timeout,
		Production: production,
		Quiet:      quiet,
		Verbose:    verbose,
	}
	scanCfg := config.Resolve(cliFlags)

	if scanCfg.APIKey == "" {
		return fmt.Errorf("API key required: set --api-key or MANTICORE_API_KEY environment variable")
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	exitCode, err := execpkg.Run(ctx, execpkg.Config{
		Command:    args,
		Dir:        dir,
		ScanConfig: scanCfg,
		FailOn:     failOn,
		Quiet:      quiet,
	})

	if err != nil {
		return err
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}
