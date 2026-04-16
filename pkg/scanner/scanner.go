package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/etsubu/manticore-scanner/pkg/api"
	"github.com/etsubu/manticore-scanner/pkg/parser"
	"github.com/etsubu/manticore-scanner/pkg/parser/npm"
)

// Result holds the output of a scan.
type Result struct {
	Items     []api.BatchResultItem
	InputFile string // the file that was parsed
}

// Run executes the full scan pipeline: parse -> submit -> poll.
func Run(ctx context.Context, cfg Config, onProgress func(completed, total int)) (*Result, error) {
	// Parse dependencies.
	packages, inputFile, err := npm.DetectAndParse(cfg.InputPath, parser.ParseOptions{
		IncludeDev: !cfg.Production,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing dependencies: %w", err)
	}

	if len(packages) == 0 {
		return &Result{InputFile: inputFile}, nil
	}

	// Convert to API request items.
	items := make([]api.ScanRequestItem, len(packages))
	for i, pkg := range packages {
		items[i] = api.ScanRequestItem{
			Package:   pkg.Name,
			Version:   pkg.Version,
			Ecosystem: api.Ecosystem(pkg.Ecosystem),
		}
	}

	// Create API client.
	client := api.NewClient(cfg.APIBaseURL, cfg.APIKey, nil)

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	pollerCfg := api.DefaultPollerConfig()
	pollerCfg.Timeout = timeout
	pollerCfg.OnProgress = onProgress

	results, err := api.PollUntilComplete(ctx, client, items, pollerCfg)
	if err != nil {
		return nil, err
	}
	return &Result{Items: results, InputFile: inputFile}, nil
}
