package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/etsubu/manticore-scanner/pkg/api"
	"github.com/etsubu/manticore-scanner/pkg/parser"
	"github.com/etsubu/manticore-scanner/pkg/parser/npm"
)

const (
	maxBatchSize    = 50
	maxConcurrency  = 3
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

	// Chunk and submit batches concurrently.
	chunks := api.ChunkItems(items, maxBatchSize)

	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	pollerCfg := api.DefaultPollerConfig()
	pollerCfg.Timeout = timeout
	pollerCfg.OnProgress = onProgress

	if len(chunks) == 1 {
		// Single batch, poll directly.
		results, err := api.PollUntilComplete(ctx, client, chunks[0], pollerCfg)
		if err != nil {
			return nil, err
		}
		return &Result{Items: results, InputFile: inputFile}, nil
	}

	// Multiple batches: submit concurrently with bounded concurrency.
	type chunkResult struct {
		items []api.BatchResultItem
		err   error
	}

	results := make([]chunkResult, len(chunks))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, chunk := range chunks {
		wg.Add(1)
		go func(idx int, ch []api.ScanRequestItem) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r, err := api.PollUntilComplete(ctx, client, ch, pollerCfg)
			results[idx] = chunkResult{items: r, err: err}
		}(i, chunk)
	}

	wg.Wait()

	// Collect all results.
	var allItems []api.BatchResultItem
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		allItems = append(allItems, r.items...)
	}

	return &Result{Items: allItems, InputFile: inputFile}, nil
}
