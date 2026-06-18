// Package version holds the application's version information.
package version

// Version is the current version of GitSum. It defaults to "dev" for
// local builds and is overridden at release time via:
//
//	go build -ldflags "-X github.com/Kschoffner13/GitSum/internal/version.Version=1.2.3"
var Version = "dev"
