// Package version reports the gitmate build version, shared by the CLI and GUI.
package version

import "runtime/debug"

// Stamped is set at release time via -ldflags "-X .../internal/version.Stamped=vX.Y.Z".
// When empty, Resolve falls back to the module version recorded in the build
// info (set for `go install`/release builds), then to "dev".
var Stamped = ""

// Resolve returns the best available version string.
func Resolve() string {
	if Stamped != "" {
		return Stamped
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}
