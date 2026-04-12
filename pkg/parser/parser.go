package parser

import "io"

// Package represents a resolved dependency.
type Package struct {
	Name      string
	Version   string
	Ecosystem string
}

// ParseOptions controls parser behavior.
type ParseOptions struct {
	IncludeDev bool
}

// Parser parses dependency files into a list of packages.
type Parser interface {
	Parse(r io.Reader, opts ParseOptions) ([]Package, error)
	Ecosystem() string
}
