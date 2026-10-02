package tlago

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// LoadDistributedWorkerTool performs the file-loading portion of TLCWorker's
// TLCApp construction using the production parser and TLC bridge. It installs
// a fresh worker interning context before configuration or semantic values are
// created. FP64 is initialized first like Java. Registry discovery and endpoint
// export remain in the transport port. Construct one Tool per worker JVM,
// shared by that JVM's threads.
func LoadDistributedWorkerTool(server *tlc.TLCServer, resolver *tlc.RMIFilenameToStreamResolver, runtime tlc.RuntimeParameters) (*tlc.Tool, Diagnostics, error) {
	if server == nil {
		panic(tlc.NewNullPointerException())
	}
	tlc.FP64InitPoly(server.GetIrredPolyForFP())
	tlc.UniqueStringInitializeWithSource(tlc.NewLocalWorkerInternSource(server))
	if resolver == nil {
		resolver = tlc.NewRMIFilenameToStreamResolver()
	}
	resolver.SetTLCServer(server)
	app, diags, err := LoadTLCApp(server.GetSpecFileName(), server.GetConfigFileName(), server.GetCheckDeadlock(), resolver, runtime)
	if app == nil {
		return nil, diags, err
	}
	return app.Tool, diags, err
}

// StartDistributedWorkerGroup is the local post-registry-lookup bootstrap.
// Configuration and Tool construction complete before registration threads
// start; those threads then register asynchronously with the supplied server.
func StartDistributedWorkerGroup(server *tlc.TLCServer, count int, resolver *tlc.RMIFilenameToStreamResolver, runtime tlc.RuntimeParameters, address ...tlc.DistributedWorkerAddress) (*tlc.DistributedWorkerGroup, Diagnostics, error) {
	count = int(int32(count))
	if count < 0 {
		panic(tlc.NewIllegalArgumentException("count < 0"))
	}
	tool, diags, err := LoadDistributedWorkerTool(server, resolver, runtime)
	if err != nil || diags.HasErrors() {
		return nil, diags, err
	}
	group := tlc.NewDistributedWorkerGroup(count, server, tool, address...)
	group.Start()
	return group, diags, nil
}

// LoadTLCApp implements the resolver-taking TLCApp constructor. Config parsing
// precedes SANY and tool construction; metadata directories and recovery are
// handled by the full constructor and create respectively.
func LoadTLCApp(specFile, configFile string, deadlock bool, resolver tlc.FilenameToStream, runtime tlc.RuntimeParameters) (*tlc.TLCApp, Diagnostics, error) {
	tool, diags, err := loadTLCAppTool(specFile, configFile, resolver, runtime)
	if tool == nil {
		return nil, diags, err
	}
	return tlc.NewTLCApp(tool, deadlock), diags, err
}

func loadTLCAppTool(specFile, configFile string, resolver tlc.FilenameToStream, runtime tlc.RuntimeParameters) (*tlc.Tool, Diagnostics, error) {
	lastSeparator := strings.LastIndexByte(specFile, os.PathSeparator)
	specDir, rootFile := specFile[:lastSeparator+1], specFile[lastSeparator+1:]
	if resolver == nil {
		standard, err := fs.Sub(embeddedJavaStandardModules, "test_vectors/java-sany/StandardModules")
		if err != nil {
			return nil, nil, err
		}
		// Include the process classpath and standard resources shipped with
		// the tool. Filesystem/user/library search stays in the
		// source resolver rather than being recreated in the parser.
		classpath := tlc.DefaultFilenameClasspath()
		classpath = append(classpath, tlc.FilenameClasspathEntry{Files: standard, Prefix: tlc.StandardModulesClasspath})
		resolver = tlc.NewSimpleFilenameToStream(nil, tlc.FilenameResolverOptions{Classpath: classpath})
	}
	configName := tlc.ModelConfigPath(configFile)
	configPath := resolver.Resolve(configName, false).GetPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, &tlc.ConfigFileError{Code: tlc.ECCFGErrorReadingFile, Params: []string{configName, err.Error()}}
	}
	source := string(data)
	if strings.HasSuffix(configName, ".tla") {
		source = tlc.ExtractMonolithConfigSource(source, strings.ReplaceAll(configName, ".tla", ""))
	}
	cfg, err := tlc.ParseModelConfigSource(configName, source)
	if err != nil {
		return nil, nil, err
	}
	spec, diags := LoadSanySpec(rootFile, LoadOptions{FilenameResolver: resolver})
	if diags.HasErrors() {
		return nil, diags, nil
	}
	diags = append(diags, CheckSpec(spec)...)
	if diags.HasErrors() {
		return nil, diags, nil
	}
	// SpecProcessor looks up the constructor's raw root name after SANY.
	// In particular, direct constructors retaining a .tla suffix fail here;
	// create strips that suffix before invoking the constructor.
	if spec.Modules[rootFile] == nil {
		panic(tlc.NewTLCRuntimeException(tlc.ECTLCParsingFailed2, fmt.Sprintf(" Module-Table lookup failure for module name %s derived from %s file name.", rootFile, rootFile)))
	}
	tool, bridgeDiags := BuildTLCTool(spec, cfg, runtime)
	diags = append(diags, bridgeDiags...)
	if tool == nil || diags.HasErrors() {
		return nil, diags, nil
	}
	tool.RootFile, tool.SpecDir, tool.ConfigFile = rootFile, specDir, configFile
	return tool, diags, nil
}

// LoadTLCAppWithMetadata is the full application constructor. A checkpoint
// path is returned verbatim by metadata setup; this constructor does not read
// the intern table, matching Java's separate create recovery step.
func LoadTLCAppWithMetadata(specFile, configFile string, deadlock bool, fromCheckpoint *string, fpSetConfig *tlc.FPSetConfiguration, resolver tlc.FilenameToStream, runtime tlc.RuntimeParameters) (*tlc.TLCApp, Diagnostics, error) {
	tool, diags, err := loadTLCAppTool(specFile, configFile, resolver, runtime)
	if tool == nil {
		return nil, diags, err
	}
	return tlc.NewTLCAppWithMetadata(tool, deadlock, fromCheckpoint, fpSetConfig), diags, err
}
