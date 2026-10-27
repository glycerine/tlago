package tlago

import (
	"embed"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

//go:embed stdlib/tla2tools/*.tla
var embeddedTLA2ToolsModules embed.FS

//go:embed stdlib/community/*.tla
var embeddedCommunityModules embed.FS

//go:embed stdlib/tlaps/*.tla
var embeddedTLAPSModules embed.FS

//go:embed stdlib/apalache/Apalache.tla
var embeddedApalacheModule embed.FS

type LoadOptions struct {
	// ParsingProgress receives SANY file/provenance progress; nil is silent.
	ParsingProgress      func(string)
	ResolutionError      func(string)
	FilenameResolver     tlc.FilenameToStream
	LibraryPaths         []string
	PreferLibraryModules bool
	ExtraModules         []string
}

// Package the original module definitions. Semantic resolution must use their
// real declarations, dependencies and export visibility, not abbreviated stubs.
var standardModules = loadStandardModules()

func loadStandardModules() map[string]string {
	out := make(map[string]string)
	loadEmbeddedModules(out, embeddedApalacheModule, "stdlib/apalache")
	loadEmbeddedModules(out, embeddedTLAPSModules, "stdlib/tlaps")
	loadEmbeddedModules(out, embeddedCommunityModules, "stdlib/community")
	loadEmbeddedModules(out, embeddedTLA2ToolsModules, "stdlib/tla2tools")
	return out
}

func loadEmbeddedModules(out map[string]string, fs embed.FS, dir string) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tla") {
			continue
		}
		data, err := fs.ReadFile(dir + "/" + entry.Name())
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".tla")
		out[name] = string(data)
	}
}
