package npm

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/etsubu/manticore-scanner/pkg/parser"
)

// lockfile represents the relevant fields of a package-lock.json file.
type lockfile struct {
	LockfileVersion int                        `json:"lockfileVersion"`
	Packages        map[string]lockfilePackage  `json:"packages"`
	Dependencies    map[string]lockfileLegacyDep `json:"dependencies"`
}

type lockfilePackage struct {
	Version string `json:"version"`
	Dev     bool   `json:"dev"`
}

type lockfileLegacyDep struct {
	Version      string                     `json:"version"`
	Dev          bool                       `json:"dev"`
	Dependencies map[string]lockfileLegacyDep `json:"dependencies"`
}

// PackageLockParser parses npm package-lock.json files.
type PackageLockParser struct{}

func (p *PackageLockParser) Ecosystem() string { return "npm" }

func (p *PackageLockParser) Parse(r io.Reader, opts parser.ParseOptions) ([]parser.Package, error) {
	var lf lockfile
	if err := json.NewDecoder(r).Decode(&lf); err != nil {
		return nil, fmt.Errorf("parsing package-lock.json: %w", err)
	}

	switch {
	case lf.LockfileVersion >= 2 && lf.Packages != nil:
		return parseV2V3(lf.Packages, opts)
	case lf.Dependencies != nil:
		return parseV1(lf.Dependencies, opts)
	default:
		return nil, fmt.Errorf("unsupported lockfile version %d or empty lockfile", lf.LockfileVersion)
	}
}

// parseV2V3 handles lockfileVersion 2 and 3, which use the "packages" map.
func parseV2V3(packages map[string]lockfilePackage, opts parser.ParseOptions) ([]parser.Package, error) {
	seen := make(map[string]bool)
	var result []parser.Package

	for key, pkg := range packages {
		if key == "" {
			// Root project entry, skip.
			continue
		}
		if pkg.Dev && !opts.IncludeDev {
			continue
		}

		name := extractPackageName(key)
		if name == "" {
			continue
		}

		dedup := name + "@" + pkg.Version
		if seen[dedup] {
			continue
		}
		seen[dedup] = true

		result = append(result, parser.Package{
			Name:      name,
			Version:   pkg.Version,
			Ecosystem: "npm",
		})
	}

	return result, nil
}

// extractPackageName extracts the npm package name from a node_modules path.
// Examples:
//
//	"node_modules/lodash" -> "lodash"
//	"node_modules/@scope/name" -> "@scope/name"
//	"node_modules/a/node_modules/@scope/name" -> "@scope/name"
func extractPackageName(key string) string {
	const prefix = "node_modules/"
	idx := strings.LastIndex(key, prefix)
	if idx == -1 {
		return ""
	}
	return key[idx+len(prefix):]
}

// parseV1 handles lockfileVersion 1, which uses nested "dependencies".
func parseV1(deps map[string]lockfileLegacyDep, opts parser.ParseOptions) ([]parser.Package, error) {
	seen := make(map[string]bool)
	var result []parser.Package
	walkV1(deps, opts, seen, &result)
	return result, nil
}

func walkV1(deps map[string]lockfileLegacyDep, opts parser.ParseOptions, seen map[string]bool, result *[]parser.Package) {
	for name, dep := range deps {
		if dep.Dev && !opts.IncludeDev {
			continue
		}

		dedup := name + "@" + dep.Version
		if seen[dedup] {
			continue
		}
		seen[dedup] = true

		*result = append(*result, parser.Package{
			Name:      name,
			Version:   dep.Version,
			Ecosystem: "npm",
		})

		if dep.Dependencies != nil {
			walkV1(dep.Dependencies, opts, seen, result)
		}
	}
}
