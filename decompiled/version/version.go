// Package version defines qodercli-reverse version info
package version

// Version qodercli-reverse version number
// kept in sync with official version
var Version = "0.1.30"

// Commit git commit hash (injected via ldflags)
var Commit = "unknown"

// BuildTime build timestamp (injected via ldflags)
var BuildTime = "unknown"

// BuildInfo build info (optional, for future extension)
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildTime string
	GoVersion string
}

// GetVersion gets version number
func GetVersion() string {
	return Version
}
