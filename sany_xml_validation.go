package tlago

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed test_vectors/java-sany/xml/sany.xsd
var sanyXMLSchema []byte

// SanyXMLSchema returns the packaged upstream schema, independently of cwd.
func SanyXMLSchema() []byte { return bytes.Clone(sanyXMLSchema) }

// ValidateSanyXML uses libxml2's complete XSD validator via xmllint. Validation
// never falls back to a partial schema check or silently succeeds without it.
// The schema is embedded; validation requires no network access.
func ValidateSanyXML(document []byte) error {
	validator, err := exec.LookPath("xmllint")
	if err != nil {
		return fmt.Errorf("SANY XML schema validation requires xmllint (libxml2): %w", err)
	}
	dir, err := os.MkdirTemp("", "tlago-sany-schema-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "sany.xsd")
	if err := os.WriteFile(path, sanyXMLSchema, 0600); err != nil {
		return err
	}
	command := exec.Command(validator, "--nonet", "--huge", "--noout", "--schema", path, "-")
	command.Stdin = bytes.NewReader(document)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}
