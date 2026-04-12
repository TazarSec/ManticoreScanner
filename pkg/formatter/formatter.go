package formatter

import "github.com/etsubu/manticore-scanner/pkg/api"

// Options controls formatter behavior.
type Options struct {
	InputFile string // path to the scanned file (used by SARIF)
}

// Formatter renders scan results into a specific output format.
type Formatter interface {
	Format(results []api.BatchResultItem, opts Options) ([]byte, error)
}

// Get returns a Formatter for the given format name.
func Get(name string) Formatter {
	switch name {
	case "json":
		return &JSONFormatter{}
	case "sarif":
		return &SARIFFormatter{}
	case "table":
		return &TableFormatter{}
	default:
		return &TableFormatter{}
	}
}
