package formatter

import (
	"fmt"
	"strings"

	"github.com/etsubu/manticore-scanner/pkg/api"
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBold   = "\033[1m"
)

// TableFormatter outputs scan results as a human-readable table.
type TableFormatter struct{}

func (f *TableFormatter) Format(results []api.BatchResultItem, opts Options) ([]byte, error) {
	if len(results) == 0 {
		return []byte("No packages scanned.\n"), nil
	}

	// Calculate column widths.
	nameW, versionW, statusW, scoreW := 7, 7, 6, 5 // minimum header widths
	for _, r := range results {
		if l := len(r.Package); l > nameW {
			nameW = l
		}
		if l := len(r.Version); l > versionW {
			versionW = l
		}
		if l := len(string(r.Status)); l > statusW {
			statusW = l
		}
	}

	var sb strings.Builder

	// Header.
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %s\n",
		nameW, "PACKAGE",
		versionW, "VERSION",
		statusW, "STATUS",
		scoreW, "SCORE",
		"FLAGS",
	)
	sb.WriteString(colorBold + header + colorReset)
	sb.WriteString(strings.Repeat("-", len(header)+10) + "\n")

	// Rows.
	for _, r := range results {
		score := "-"
		scoreColor := colorReset
		flags := ""

		if r.Profile != nil {
			score = fmt.Sprintf("%.1f", r.Profile.SuspicionScore)
			scoreColor = scoreColorFor(r.Profile.SuspicionScore)

			var flagParts []string
			if r.Profile.HasUnknownNetwork {
				flagParts = append(flagParts, "NET")
			}
			if r.Profile.HasSensitiveFileAcces {
				flagParts = append(flagParts, "FILE")
			}
			if r.Profile.HasUnexpectedProcess {
				flagParts = append(flagParts, "PROC")
			}
			flags = strings.Join(flagParts, ",")
		}

		line := fmt.Sprintf("%-*s  %-*s  %-*s  %s%-*s%s  %s\n",
			nameW, r.Package,
			versionW, r.Version,
			statusW, string(r.Status),
			scoreColor, scoreW, score, colorReset,
			flags,
		)
		sb.WriteString(line)
	}

	// Summary.
	completed := 0
	suspicious := 0
	for _, r := range results {
		if r.Status == api.StatusCompleted {
			completed++
		}
		if r.Profile != nil && r.Profile.SuspicionScore > 0 {
			suspicious++
		}
	}
	sb.WriteString(fmt.Sprintf("\n%d packages scanned, %d completed, %d suspicious\n", len(results), completed, suspicious))

	return []byte(sb.String()), nil
}

func scoreColorFor(score float64) string {
	switch {
	case score >= 70:
		return colorRed
	case score >= 30:
		return colorYellow
	default:
		return colorGreen
	}
}
