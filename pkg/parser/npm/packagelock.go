package npm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/etsubu/manticore-scanner/pkg/parser"
)

// lockfile represents the relevant fields of a package-lock.json file.
type lockfile struct {
	LockfileVersion int                          `json:"lockfileVersion"`
	Packages        map[string]lockfilePackage   `json:"packages"`
	Dependencies    map[string]lockfileLegacyDep `json:"dependencies"`
}

type lockfilePackage struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dev             bool              `json:"dev"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type lockfileLegacyDep struct {
	Version      string                       `json:"version"`
	Dev          bool                         `json:"dev"`
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

func parseRootPackages(packages map[string]lockfilePackage) (map[string]interface{}, error) {
	var rootPackages = make(map[string]interface{}, 64)
	if pkg, ok := packages[""]; ok {
		for k := range pkg.Dependencies {
			rootPackages[k] = true
		}
		for k := range pkg.DevDependencies {
			rootPackages[k] = true
		}
	} else {
		return rootPackages, errors.New("package-json does not contain root dependency")
	}
	return rootPackages, nil
}

// parseV2V3 handles lockfileVersion 2 and 3, which use the "packages" map.
func parseV2V3(packages map[string]lockfilePackage, opts parser.ParseOptions) ([]parser.Package, error) {
	seen := make(map[string]bool)
	var result []parser.Package
	var rootPackages, err = parseRootPackages(packages)
	if err != nil {
		return result, err
	}

	for key, pkg := range packages {
		if pkg.Dev && !opts.IncludeDev {
			continue
		}

		name := extractPackageName(key)
		if name == "" {
			continue
		}
		if _, ok := rootPackages[name]; !ok {
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

func parseV1(deps map[string]lockfileLegacyDep, opts parser.ParseOptions) ([]parser.Package, error) {
	seen := make(map[string]bool)
	var result []parser.Package
	for name, dep := range deps {
		if dep.Dev && !opts.IncludeDev {
			continue
		}

		dedup := name + "@" + dep.Version
		if seen[dedup] {
			continue
		}
		seen[dedup] = true

		result = append(result, parser.Package{
			Name:      name,
			Version:   dep.Version,
			Ecosystem: "npm",
		})
	}
	return result, nil
}
