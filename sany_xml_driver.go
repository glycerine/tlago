package tlago

import (
	"fmt"
	"github.com/glycerine/tlago/tlc"
	"io"
)

// XMLExporterExitCode follows the XML exporter enum, independently of SANY's
// exit codes (in particular, semantic spec failures are code 2 here).
type XMLExporterExitCode int

const (
	XMLExporterOK XMLExporterExitCode = iota
	XMLArgsParsingFailure
	XMLSpecParsingFailure
	XMLCannotFindEmbeddedSchemaFile
	XMLConfigurationFailure
	XMLTransformationFailure
	XMLSchemaValidationFailure
	XMLUnrepresentableCharacter
)

func (code XMLExporterExitCode) IsBug() bool {
	return code >= XMLCannotFindEmbeddedSchemaFile && code <= XMLSchemaValidationFailure
}

type XMLExportingException struct {
	Code    XMLExporterExitCode
	Message string
	Nested  error
}

func (failure *XMLExportingException) Error() string { return failure.Message }
func (failure *XMLExportingException) Unwrap() error { return failure.Nested }

// ParseSanyXMLSpec follows XMLExporter.parseSpec's validAstSettings: syntax,
// generation and level checking, with linting disabled and errors only output.
func ParseSanyXMLSpec(path string, includeDirs []string, stderr io.Writer) (spec *Spec, err error) {
	if stderr == nil {
		stderr = io.Discard
	}
	defer func() {
		if failure := recover(); failure != nil {
			if nested, ok := failure.(*tlc.FrontEndException); ok {
				err = &XMLExportingException{XMLSpecParsingFailure, "Failed to parse module.", nested}
				return
			}
			panic(failure)
		}
	}()
	settings := defaultSanyDriverSettings()
	settings.lint = false
	spec, status := parseSanyWithSettings(path, LoadOptions{LibraryPaths: includeDirs}, settings,
		func(level Severity, message string) {
			if level == SeverityError {
				fmt.Fprintln(stderr, message)
			}
		})
	if status != sanyExitOK {
		return nil, &XMLExportingException{Code: XMLSpecParsingFailure, Message: "Failed to parse module."}
	}
	return spec, nil
}

// SanySpecToXMLString retains the original four explicit export options.
func SanySpecToXMLString(spec *Spec, restricted, uncomment, prettyPrint, offline bool) (string, error) {
	x := newSanyXMLExporter(spec, SanyXMLOptions{Restricted: restricted, UncommentPreComments: uncomment, Terse: !prettyPrint, Offline: offline})
	data, diagnostics := x.xml()
	if diagnostics.HasErrors() {
		code := XMLTransformationFailure
		for _, diagnostic := range diagnostics {
			if diagnostic.Code == "E7007" {
				code = XMLUnrepresentableCharacter
				break
			}
		}
		return "", &XMLExportingException{Code: code, Message: diagnostics.Error()}
	}
	if !offline {
		if len(sanyXMLSchema) == 0 {
			return "", &XMLExportingException{Code: XMLCannotFindEmbeddedSchemaFile, Message: "Unable to find embedded sany.xsd schema file."}
		}
		if err := ValidateSanyXML(data); err != nil {
			return "", &XMLExportingException{XMLSchemaValidationFailure, "Failed to validate XML", err}
		}
	}
	return string(data), nil
}

// XMLModuleToXML is XMLExporter.moduleToXML's library entry point. It writes
// output and returns the structured source exception rather than an exit int.
func XMLModuleToXML(args []string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if len(args) == 0 {
		return &XMLExportingException{Code: XMLArgsParsingFailure, Message: "at least one .tla file must be given"}
	}
	opts, xmlOpts, err := parseSanyXMLCLIOptions(args)
	if err != nil {
		return &XMLExportingException{Code: XMLArgsParsingFailure, Message: err.Error()}
	}
	if cliArgsContainHelp(args) {
		printSanyXMLUsage(stdout)
		return nil
	}
	if len(opts.files) != 1 {
		return &XMLExportingException{Code: XMLArgsParsingFailure, Message: "Only one TLA file to check allowed!"}
	}
	spec, err := ParseSanyXMLSpec(opts.files[0], opts.load.LibraryPaths, stderr)
	if err != nil {
		return err
	}
	text, err := SanySpecToXMLString(spec, xmlOpts.Restricted, xmlOpts.UncommentPreComments, !xmlOpts.Terse, xmlOpts.Offline)
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, text)
	if err != nil {
		return &XMLExportingException{XMLTransformationFailure, "Failed to transform XML", err}
	}
	return nil
}
