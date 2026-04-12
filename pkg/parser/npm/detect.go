package npm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/etsubu/manticore-scanner/pkg/parser"
)

// DetectAndParse auto-detects the dependency file in the given path and parses it.
// If path is a directory, it prefers package-lock.json over package.json.
// If path is a file, it parses that file directly.
func DetectAndParse(path string, opts parser.ParseOptions) ([]parser.Package, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", fmt.Errorf("stat %s: %w", path, err)
	}

	if !info.IsDir() {
		return parseFile(path, opts)
	}

	// Directory: try package-lock.json first, then package.json.
	lockPath := filepath.Join(path, "package-lock.json")
	if _, err := os.Stat(lockPath); err == nil {
		pkgs, _, err := parseFile(lockPath, opts)
		return pkgs, lockPath, err
	}

	pkgPath := filepath.Join(path, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		pkgs, _, err := parseFile(pkgPath, opts)
		return pkgs, pkgPath, err
	}

	return nil, "", fmt.Errorf("no package-lock.json or package.json found in %s", path)
}

func parseFile(path string, opts parser.ParseOptions) ([]parser.Package, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, path, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	p, err := detectParser(path)
	if err != nil {
		return nil, path, err
	}

	pkgs, err := p.Parse(f, opts)
	if err != nil {
		return nil, path, err
	}
	return pkgs, path, nil
}

// detectParser determines the appropriate parser based on filename or file content.
// For exact names "package-lock.json" and "package.json", it uses the filename.
// For other files, it peeks at the JSON to check for lockfileVersion.
func detectParser(path string) (parser.Parser, error) {
	base := filepath.Base(path)

	switch base {
	case "package-lock.json":
		return &PackageLockParser{}, nil
	case "package.json":
		return &PackageJSONParser{}, nil
	}

	// For non-standard filenames, try to detect from content.
	if strings.Contains(base, "package-lock") || strings.Contains(base, "npm-shrinkwrap") {
		return &PackageLockParser{}, nil
	}

	// Peek at JSON structure to auto-detect.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s for detection: %w", path, err)
	}

	var probe struct {
		LockfileVersion *int `json:"lockfileVersion"`
	}
	if err := json.Unmarshal(data, &probe); err == nil && probe.LockfileVersion != nil {
		return &PackageLockParser{}, nil
	}

	return &PackageJSONParser{}, nil
}
