package formatter

import (
	"encoding/json"

	"github.com/etsubu/manticore-scanner/pkg/api"
)

// JSONFormatter outputs scan results as JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(results []api.BatchResultItem, opts Options) ([]byte, error) {
	return json.MarshalIndent(results, "", "  ")
}
