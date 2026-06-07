// internal/version/version.go

package version

//
// Injected during release builds via ldflags.
// Defaults to "dev" for local builds.
//

var Version = "dev"
