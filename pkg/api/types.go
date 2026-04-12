package api

import "time"

// Ecosystem represents a package ecosystem.
type Ecosystem string

const (
	EcosystemNPM      Ecosystem = "npm"
	EcosystemPyPI     Ecosystem = "pypi"
	EcosystemRubyGems Ecosystem = "rubygems"
	EcosystemCrates   Ecosystem = "crates"
	EcosystemMaven    Ecosystem = "maven"
	EcosystemGo       Ecosystem = "go"
)

// ScanStatus represents the status of a scan.
type ScanStatus string

const (
	StatusCompleted  ScanStatus = "completed"
	StatusInProgress ScanStatus = "in_progress"
	StatusQueued     ScanStatus = "queued"
)

// Severity represents the severity level of a suspicion reason.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Phase represents the phase during which suspicious behavior was detected.
type Phase string

const (
	PhaseInstall Phase = "install"
	PhaseRequire Phase = "require"
)

// ScanRequestItem represents a single package in a batch request.
type ScanRequestItem struct {
	Package   string    `json:"package,omitempty"`
	Version   string    `json:"version,omitempty"`
	Ecosystem Ecosystem `json:"ecosystem,omitempty"`
	Hash      string    `json:"hash,omitempty"`
}

// ScanRequest is the request body for the scan endpoint.
type ScanRequest struct {
	Packages []ScanRequestItem `json:"packages"`
}

// SuspicionReason describes why a package is considered suspicious.
type SuspicionReason struct {
	Type     string   `json:"type"`
	Detail   string   `json:"detail"`
	Severity Severity `json:"severity"`
	Phase    Phase    `json:"phase"`
}

// Profile contains the full scan result for a package.
type Profile struct {
	ID                    string            `json:"id"`
	Ecosystem             string            `json:"ecosystem"`
	PackageName           string            `json:"package_name"`
	Version               string            `json:"version"`
	SuspicionScore        float64           `json:"suspicion_score"`
	SuspicionReasons      []SuspicionReason `json:"suspicion_reasons"`
	HasUnknownNetwork     bool              `json:"has_unknown_network"`
	HasSensitiveFileAcces bool              `json:"has_sensitive_file_access"`
	HasUnexpectedProcess  bool              `json:"has_unexpected_processes"`
	ScannedAt             time.Time         `json:"scanned_at"`
}

// BatchResultItem is a single result in a batch response.
type BatchResultItem struct {
	Package   string     `json:"package,omitempty"`
	Version   string     `json:"version,omitempty"`
	Ecosystem string     `json:"ecosystem,omitempty"`
	Hash      string     `json:"hash,omitempty"`
	Status    ScanStatus `json:"status"`
	Profile   *Profile   `json:"profile,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// BatchResponse is the response body for a batch scan request.
type BatchResponse struct {
	Results []BatchResultItem `json:"results"`
}

// APIError is the error response body from the API.
type APIError struct {
	Error string `json:"error"`
}
