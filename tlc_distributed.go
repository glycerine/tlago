package tlago

import (
	"fmt"
	"io"
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
	return startDistributedWorkerGroup(server, count, resolver, runtime, "", nil, address...)
}

func startDistributedWorkerGroup(server *tlc.TLCServer, count int, resolver *tlc.RMIFilenameToStreamResolver, runtime tlc.RuntimeParameters, serverURL string, lookup tlc.TLCServerLookup, address ...tlc.DistributedWorkerAddress) (*tlc.DistributedWorkerGroup, Diagnostics, error) {
	count = int(int32(count))
	if count < 0 {
		panic(tlc.NewIllegalArgumentException("count < 0"))
	}
	tool, diags, err := LoadDistributedWorkerTool(server, resolver, runtime)
	if err != nil || diags.HasErrors() {
		return nil, diags, err
	}
	group := tlc.NewDistributedWorkerGroup(count, server, tool, address...)
	if lookup != nil {
		group.Runtime.ConfigureKeepAliveLookup(serverURL, tlc.NewTLCServerStatusLookup(lookup), nil)
	}
	group.Start()
	return group, diags, nil
}

// StartDistributedWorkerGroupWithLookup connects the source discovery loop to
// the existing production config/parser/tool bootstrap. lookup supplies the
// Naming.lookup boundary; registry/RPC transport and command main are separate.
func StartDistributedWorkerGroupWithLookup(serverName string, count int, lookup tlc.TLCServerLookup, sleep tlc.DistributedLookupSleep, output io.Writer, resolver *tlc.RMIFilenameToStreamResolver, runtime tlc.RuntimeParameters, address ...tlc.DistributedWorkerAddress) (*tlc.DistributedWorkerGroup, Diagnostics, error) {
	count = int(int32(count))
	if count < 0 {
		// Java constructs CountDownLatch before entering the lookup try/catch.
		panic(tlc.NewIllegalArgumentException("count < 0"))
	}
	server, url, err := tlc.DiscoverTLCWorkerServer(serverName, lookup, sleep, output)
	if err != nil {
		var diags Diagnostics
		return nil, diags, err
	}
	return startDistributedWorkerGroup(server, count, resolver, runtime, url, lookup, address...)
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
	tlc.InitializeToolProperties()
	if resolver == nil {
		classpath, err := tlcApplicationClasspath(nil)
		if err != nil {
			return nil, nil, err
		}
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
		if failure, ok := err.(*tlc.ConfigFileError); ok {
			// ModelConfig.parse throws ConfigFileException; the public config
			// parser retains its separate native result carrier.
			err = tlc.NewConfigFileException(failure.Code, failure.Params)
		}
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

func tlcApplicationClasspath(entries []tlc.FilenameClasspathEntry) ([]tlc.FilenameClasspathEntry, error) {
	if entries == nil {
		entries = tlc.DefaultFilenameClasspath()
	}
	standard, err := fs.Sub(embeddedJavaStandardModules, "test_vectors/java-sany/StandardModules")
	if err != nil {
		return nil, err
	}
	classpath := append([]tlc.FilenameClasspathEntry{}, entries...)
	return append(classpath, tlc.FilenameClasspathEntry{Files: standard, Prefix: tlc.StandardModulesClasspath}), nil
}

// CreateTLCApp ports Java TLCApp.create. Optional classpath roots supply Go's
// directory/archive/bundled resource adapter for the process class loader.
// Option errors retain Java's printed diagnostics and null application result.
func CreateTLCApp(args []string, runtime tlc.RuntimeParameters, classpath ...[]tlc.FilenameClasspathEntry) (*tlc.TLCApp, Diagnostics, error) {
	opts := tlc.ParseTLCAppOptions(args)
	if opts == nil {
		return nil, nil, nil
	}
	var configured []tlc.FilenameClasspathEntry
	if len(classpath) > 0 {
		configured = classpath[0]
	}
	if opts.SpecFile == nil {
		entries, err := tlcApplicationClasspath(configured)
		if err != nil {
			return nil, nil, err
		}
		model := tlc.NewModelInJar(entries)
		if !model.HasModel() {
			tlc.PrintTLCAppUsageError("Error: Missing input TLA+ module.")
			return nil, nil, nil
		}
		model.LoadProperties()
		tlc.Globals.Tool = true
		tlc.Globals.CheckpointDurationMillis = 0
		tlc.FP64InitIndex(opts.FPIndex)
		// This source branch ignores -config and skips early intern recovery,
		// but retains fromChkpt in the full constructor's metadata setup.
		app, diags, err := LoadTLCAppWithMetadata(tlc.ModelCheckFileBasename, tlc.ModelCheckFileBasename, opts.CheckDeadlock, opts.FromCheckpoint, opts.FPSetConfiguration, model.Resolver(), runtime)
		if app != nil {
			app.Tool.DistributedFiles.Classpath = entries
		}
		return app, diags, err
	}
	if opts.ConfigFile == nil {
		opts.ConfigFile = opts.SpecFile
	}
	if opts.FromCheckpoint != nil {
		if err := tlc.RecoverUniqueStrings(*opts.FromCheckpoint); err != nil {
			return nil, nil, err
		}
	}
	tlc.FP64InitIndex(opts.FPIndex)
	var resolver tlc.FilenameToStream
	var entries []tlc.FilenameClasspathEntry
	if len(classpath) > 0 {
		var err error
		entries, err = tlcApplicationClasspath(configured)
		if err != nil {
			return nil, nil, err
		}
		resolver = tlc.NewSimpleFilenameToStream(nil, tlc.FilenameResolverOptions{Classpath: entries})
	}
	app, diags, err := LoadTLCAppWithMetadata(*opts.SpecFile, *opts.ConfigFile, opts.CheckDeadlock, opts.FromCheckpoint, opts.FPSetConfiguration, resolver, runtime)
	if app != nil && entries != nil {
		app.Tool.DistributedFiles.Classpath = entries
	}
	return app, diags, err
}

// RunDistributedWorker connects TLCWorker.main to the production Go parser and
// TLCApp constructor. The supplied process retains the source static lifecycle
// across calls. Environment adapters supply the registry/process boundaries.
func RunDistributedWorker(process *tlc.DistributedWorkerProcess, args []string, env tlc.DistributedWorkerEnvironment, runtime tlc.RuntimeParameters) (Diagnostics, error) {
	var diags Diagnostics
	if env.LoadApp == nil {
		env.LoadApp = func(server *tlc.TLCServer, resolver *tlc.RMIFilenameToStreamResolver) (*tlc.TLCApp, error) {
			app, loaded, err := LoadTLCApp(server.GetSpecFileName(), server.GetConfigFileName(), server.GetCheckDeadlock(), resolver, runtime)
			diags = append(diags, loaded...)
			if err == nil && loaded.HasErrors() {
				params := make([]string, 0, len(loaded))
				for _, diagnostic := range loaded.Errors() {
					params = append(params, diagnostic.Message)
				}
				err = tlc.NewTLCRuntimeException(tlc.ECTLCParsingFailed, params...)
			}
			return app, err
		}
	}
	err := process.Run(args, env)
	return diags, err
}

// RunDistributedServer connects TLCServer.main to production application
// creation and packaged model properties. Email reporting remains disabled
// by user policy.
func RunDistributedServer(process *tlc.DistributedServerProcess, args []string, env tlc.DistributedServerEnvironment, runtime tlc.RuntimeParameters, classpath ...[]tlc.FilenameClasspathEntry) (Diagnostics, error) {
	var diags Diagnostics
	if env.LoadProperties == nil {
		env.LoadProperties = func() {
			var configured []tlc.FilenameClasspathEntry
			if len(classpath) > 0 {
				configured = classpath[0]
			}
			entries, err := tlcApplicationClasspath(configured)
			if err != nil {
				panic(err)
			}
			tlc.NewModelInJar(entries).LoadProperties()
		}
	}
	if env.CreateApp == nil {
		env.CreateApp = func(args []string) (*tlc.TLCApp, error) {
			app, loaded, err := CreateTLCApp(args, runtime, classpath...)
			diags = append(diags, loaded...)
			if err == nil && loaded.HasErrors() {
				params := make([]string, 0, len(loaded))
				for _, diagnostic := range loaded.Errors() {
					params = append(params, diagnostic.Message)
				}
				err = tlc.NewTLCRuntimeException(tlc.ECTLCParsingFailed, params...)
			}
			return app, err
		}
	}
	err := process.Run(args, env)
	return diags, err
}

// RunDistributedWorkerAndFPServer launches both distributed commands in the
// source order. Application diagnostics become the worker command's reported
// error as in RunDistributedWorker; completion and uncaught failures occur on
// the child threads, independently of this launcher's return.
func RunDistributedWorkerAndFPServer(process *tlc.DistributedWorkerProcess, args []string, env tlc.DistributedWorkerAndFPServerEnvironment, runtime tlc.RuntimeParameters) error {
	if env.WorkerMain == nil {
		env.WorkerMain = func(args []string) error {
			if process == nil {
				process = tlc.NewDistributedWorkerProcess()
			}
			_, err := RunDistributedWorker(process, args, env.Worker, runtime)
			return err
		}
	}
	return tlc.RunDistributedWorkerAndFPServer(process, args, env)
}
