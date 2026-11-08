// Package version defines qodercli-reverse version info
package version

// Version qodercli-reverse version number
// kept in sync with official version
const Version = "0.1.30"

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
