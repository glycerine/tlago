package tlago

import (
	"os"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// LoadDistributedWorkerTool performs the file-loading portion of TLCWorker's
// TLCApp construction using the production parser and TLC bridge. Registry,
// intern-source setup, and worker-group launch belong to the remaining startup
// port, before this function is called.
func LoadDistributedWorkerTool(server *tlc.TLCServer, resolver *tlc.RMIFilenameToStreamResolver, runtime tlc.RuntimeParameters) (*tlc.Tool, Diagnostics, error) {
	if server == nil {
		panic(tlc.NewNullPointerException())
	}
	if resolver == nil {
		resolver = tlc.NewRMIFilenameToStreamResolver()
	}
	resolver.SetTLCServer(server)
	configName := tlc.ModelConfigPath(server.GetConfigFileName())
	configPath := resolver.Resolve(configName, false)
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
