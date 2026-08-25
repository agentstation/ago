package ago

import "runtime/debug"

// Version is the module version. Release builds can set it with a linker
// value. Commands built through go install or go tool read it from Go build
// information. A local source build without a version reports "dev".
var Version = "dev"

func init() {
	info, _ := debug.ReadBuildInfo()
	Version = resolveVersion(Version, info)
}

func resolveVersion(linked string, info *debug.BuildInfo) string {
	if linked != "" && linked != "dev" {
		return linked
	}
	if info != nil && info.Main.Path == "github.com/agentstation/ago" &&
		info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
