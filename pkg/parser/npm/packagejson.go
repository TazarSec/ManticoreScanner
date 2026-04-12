package npm

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/etsubu/manticore-scanner/pkg/parser"
)

// packageJSON represents the relevant fields of a package.json file.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// PackageJSONParser parses npm package.json files.
type PackageJSONParser struct{}

func (p *PackageJSONParser) Ecosystem() string { return "npm" }

func (p *PackageJSONParser) Parse(r io.Reader, opts parser.ParseOptions) ([]parser.Package, error) {
	var pkg packageJSON
	if err := json.NewDecoder(r).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("parsing package.json: %w", err)
	}

	seen := make(map[string]bool)
	var packages []parser.Package

	for name, version := range pkg.Dependencies {
		key := name + "@" + version
		if seen[key] {
			continue
		}
		seen[key] = true
		packages = append(packages, parser.Package{
			Name:      name,
			Version:   version,
			Ecosystem: "npm",
		})
	}

	if opts.IncludeDev {
		for name, version := range pkg.DevDependencies {
			key := name + "@" + version
			if seen[key] {
				continue
			}
			seen[key] = true
			packages = append(packages, parser.Package{
				Name:      name,
				Version:   version,
				Ecosystem: "npm",
			})
		}
	}

	return packages, nil
}
