// internal/version/version.go

package version

import "runtime/debug"

//
// Injected during release builds via ldflags.
// Defaults to Go build version for local builds.
// Falls back to "dev" when build metadata is unavailable.
//

var Version = "dev"

func Get() string {

	if Version != "dev" {
		return Version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return Version
	}

	if info.Main.Version != "" &&
		info.Main.Version != "(devel)" {

		return info.Main.Version
	}

	return Version
}
