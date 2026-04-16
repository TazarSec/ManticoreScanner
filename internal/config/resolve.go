package config

import (
	"os"
	"strconv"

	"github.com/etsubu/manticore-scanner/pkg/scanner"
)

const (
	defaultAPIURL           = "https://api.aegisengine.com"
	defaultTimeout          = 300
	defaultFormat           = "table"
	defaultFailureThreshold = 1
)

// CLIFlags holds the raw flags from the CLI.
type CLIFlags struct {
	APIKey     string
	APIURL     string
	File       string
	Format     string
	Output     string
	FailOn     float64
	FailOnSet  bool
	Timeout    int
	Production bool
	VCSComment bool
	Quiet      bool
	Verbose    bool
}

// Resolve builds a scanner.Config by merging CLI flags with environment variables.
// Priority: CLI flags > environment variables > defaults.
func Resolve(flags CLIFlags) scanner.Config {
	cfg := scanner.Config{
		APIKey:     envOrDefault("MANTICORE_API_KEY", ""),
		APIBaseURL: envOrDefault("MANTICORE_API_URL", defaultAPIURL),
		TimeoutSec: envIntOrDefault("MANTICORE_TIMEOUT", defaultTimeout),
		Format:     envOrDefault("MANTICORE_FORMAT", defaultFormat),
		FailOn:     envFloatOrDefault("MANTICORE_FAILURE_TRESHOLD", defaultFailureThreshold),
		InputPath:  ".",
		Production: false,
		PostToVCS:  false,
		Quiet:      false,
		Verbose:    false,
	}

	// CLI flags override env.
	if flags.APIKey != "" {
		cfg.APIKey = flags.APIKey
	}
	if flags.APIURL != "" {
		cfg.APIBaseURL = flags.APIURL
	}
	if flags.File != "" {
		cfg.InputPath = flags.File
	}
	if flags.Format != "" {
		cfg.Format = flags.Format
	}
	if flags.Output != "" {
		cfg.OutputPath = flags.Output
	}
	if flags.FailOnSet {
		cfg.FailOn = flags.FailOn
	}
	if flags.Timeout > 0 {
		cfg.TimeoutSec = flags.Timeout
	}
	if flags.Production {
		cfg.Production = true
	}
	if flags.VCSComment {
		cfg.PostToVCS = true
	}
	if flags.Quiet {
		cfg.Quiet = true
	}
	if flags.Verbose {
		cfg.Verbose = true
	}

	return cfg
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envFloatOrDefault(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envIntOrDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
