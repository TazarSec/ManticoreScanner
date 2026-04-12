package npm

import (
	"os"
	"sort"
	"testing"

	"github.com/etsubu/manticore-scanner/pkg/parser"
)

func TestPackageJSONParser_WithDev(t *testing.T) {
	f, err := os.Open("../../../testdata/package.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p := &PackageJSONParser{}
	pkgs, err := p.Parse(f, parser.ParseOptions{IncludeDev: true})
	if err != nil {
		t.Fatal(err)
	}

	names := packageNames(pkgs)
	sort.Strings(names)

	expected := []string{"express", "jest", "lodash"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d packages, got %d: %v", len(expected), len(names), names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected %s at position %d, got %s", name, i, names[i])
		}
	}
}

func TestPackageJSONParser_WithoutDev(t *testing.T) {
	f, err := os.Open("../../../testdata/package.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p := &PackageJSONParser{}
	pkgs, err := p.Parse(f, parser.ParseOptions{IncludeDev: false})
	if err != nil {
		t.Fatal(err)
	}

	names := packageNames(pkgs)
	sort.Strings(names)

	expected := []string{"express", "lodash"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d packages, got %d: %v", len(expected), len(names), names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected %s at position %d, got %s", name, i, names[i])
		}
	}
}

func TestPackageJSONParser_Ecosystem(t *testing.T) {
	p := &PackageJSONParser{}
	if p.Ecosystem() != "npm" {
		t.Errorf("expected npm, got %s", p.Ecosystem())
	}
}

func packageNames(pkgs []parser.Package) []string {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	return names
}
