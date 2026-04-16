package scanner

// Config holds the interface-agnostic scan configuration.
type Config struct {
	APIKey     string
	APIBaseURL string
	TimeoutSec int
	FailOn     float64 // suspicion_score threshold; -1 = don't fail
	Format     string  // "table", "json", "sarif"
	OutputPath string  // "" = stdout
	InputPath  string  // path to dependency file or directory
	Production bool    // skip devDependencies
	PostToVCS  bool    // publish results to VCS
	Quiet      bool
	Verbose    bool
}
