package version

import "fmt"

// Version, Commit, and Date are set via -ldflags at release build time.
var (
	Version = "0.0.0-dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// String returns a single-line build description.
func String() string {
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
}
