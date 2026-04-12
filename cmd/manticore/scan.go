package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/etsubu/manticore-scanner/internal/config"
	"github.com/etsubu/manticore-scanner/pkg/formatter"
	"github.com/etsubu/manticore-scanner/pkg/scanner"
	"github.com/etsubu/manticore-scanner/pkg/vcs/github"
)

var flags config.CLIFlags

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan packages from package.json / package-lock.json",
	Long:  "Parse npm dependency files, submit packages to AegisEngine for behavioral analysis, and report findings.",
	RunE:  runScan,
}

func init() {
	f := scanCmd.Flags()
	f.StringVar(&flags.APIKey, "api-key", "", "API key (or set MANTICORE_API_KEY)")
	f.StringVar(&flags.APIURL, "api-url", "", "API base URL (or set MANTICORE_API_URL)")
	f.StringVar(&flags.File, "file", "", "Path to package.json or package-lock.json (default: auto-detect in cwd)")
	f.StringVar(&flags.Format, "format", "", "Output format: table, json, sarif (default: table)")
	f.StringVar(&flags.Output, "output", "", "Write output to file (default: stdout)")
	f.Float64Var(&flags.FailOn, "fail-on", 0, "Exit code 1 if any suspicion_score >= this value")
	f.IntVar(&flags.Timeout, "timeout", 0, "Polling timeout in seconds (default: 300)")
	f.BoolVar(&flags.Production, "production", false, "Skip devDependencies")
	f.BoolVar(&flags.VCSComment, "vcs-comment", false, "Post results to VCS PR/MR (requires GITHUB_TOKEN)")
	f.BoolVar(&flags.Quiet, "quiet", false, "Suppress progress output")
	f.BoolVar(&flags.Verbose, "verbose", false, "Verbose logging")
}

func runScan(cmd *cobra.Command, args []string) error {
	// Track whether --fail-on was explicitly set.
	flags.FailOnSet = cmd.Flags().Changed("fail-on")

	cfg := config.Resolve(flags)

	if cfg.APIKey == "" {
		return fmt.Errorf("API key is required. Set --api-key or MANTICORE_API_KEY environment variable")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Progress callback.
	var onProgress func(completed, total int)
	if !cfg.Quiet {
		onProgress = func(completed, total int) {
			fmt.Fprintf(os.Stderr, "\rScanning: %d/%d packages completed", completed, total)
		}
	}

	// Run the scan.
	result, err := scanner.Run(ctx, cfg, onProgress)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if !cfg.Quiet && onProgress != nil {
		fmt.Fprintln(os.Stderr) // newline after progress
	}

	if len(result.Items) == 0 {
		if !cfg.Quiet {
			fmt.Fprintln(os.Stderr, "No packages found to scan.")
		}
		return nil
	}

	// Format output.
	f := formatter.Get(cfg.Format)
	output, err := f.Format(result.Items, formatter.Options{
		InputFile: result.InputFile,
	})
	if err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	// Write output.
	if cfg.OutputPath != "" {
		if err := os.WriteFile(cfg.OutputPath, output, 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		if !cfg.Quiet {
			fmt.Fprintf(os.Stderr, "Results written to %s\n", cfg.OutputPath)
		}
	} else {
		fmt.Print(string(output))
	}

	// Post to VCS if requested.
	if cfg.PostToVCS {
		if err := postToVCS(ctx, result); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to post VCS comment: %v\n", err)
		} else if !cfg.Quiet {
			fmt.Fprintln(os.Stderr, "Results posted to PR.")
		}
	}

	// Check fail threshold.
	if cfg.FailOn != nil {
		for _, item := range result.Items {
			if item.Profile != nil && item.Profile.SuspicionScore >= *cfg.FailOn {
				fmt.Fprintf(os.Stderr, "FAIL: %s@%s has suspicion score %.1f (threshold: %.1f)\n",
					item.Package, item.Version,
					item.Profile.SuspicionScore, *cfg.FailOn,
				)
				os.Exit(1)
			}
		}
	}

	return nil
}

func postToVCS(ctx context.Context, result *scanner.Result) error {
	provider := github.NewProvider(nil)
	vcsCtx, err := provider.Detect()
	if err != nil {
		return fmt.Errorf("detecting VCS environment: %w", err)
	}
	return provider.PostResults(ctx, vcsCtx, result.Items)
}
