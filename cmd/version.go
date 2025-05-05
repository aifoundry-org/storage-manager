package cmd

import (
	"fmt"
)

// Version information set by build flags
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return fmt.Sprintf("Storage Manager %s (commit: %s, built: %s)", Version, Commit, BuildDate)
}
