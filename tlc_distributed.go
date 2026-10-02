package tlago

import (
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
	configName := tlc.ModelConfigPath(server.GetConfigFileName())
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
	spec, diags := LoadSanySpec(server.GetSpecFileName(), LoadOptions{DistributedResolver: resolver})
	if diags.HasErrors() {
		return nil, diags, nil
	}
	diags = append(diags, CheckSpec(spec)...)
	if diags.HasErrors() {
		return nil, diags, nil
	}
	tool, bridgeDiags := BuildTLCTool(spec, cfg, runtime)
	return tool, append(diags, bridgeDiags...), nil
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
