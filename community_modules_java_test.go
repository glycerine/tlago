// Copyright (c) 2019 TLA+. Licensed under the MIT license in
// test_vectors/CommunityModules/LICENSE.
// Port of the complete CommunityModules build.xml "test" target: the original
// AllTests/AllTestsUnix assumptions and the separate GH037 ShiViz invocation.
package tlago

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaCommunityModulesAnt(t *testing.T) {
	const phaseVariable = "TLAGO_COMMUNITY_ANT_TEST_PHASE"
	if phase := os.Getenv(phaseVariable); phase != "" {
		runJavaCommunityModulesAntPhase(t, phase)
		return
	}
	if runtime.GOOS != "windows" {
		if _, err := exec.LookPath("java"); err != nil {
			t.Skip("original IOUtilsUnixTests launches Java TLC through IOExec; java is unavailable")
		}
	}
	root, err := filepath.Abs("test_vectors")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, name := range []string{"tests", "modules"} {
		if err := os.CopyFS(filepath.Join(dir, name), os.DirFS(filepath.Join(root, "CommunityModules", name))); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "build", "tests"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tlc"), 0755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "java-sany", "tla2tools.jar"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tlc", "tla2tools.jar"), data, 0644); err != nil {
		t.Fatal(err)
	}
	// Preserve the original nested IOExec classpath, including class directories
	// populated by Ant's build and unzip tasks. These nested processes exercise
	// IOUtils; the outer assumptions and ShiViz model run through Go TLC.
	extractCommunityAntClasses(t, filepath.Join(root, "java-sany", "CommunityModules.jar"), filepath.Join(dir, "build", "modules"))
	for _, name := range []string{
		"gson-2.8.6.jar", "jgrapht-core-1.5.1.jar", "jungrapht-layout-1.4-SNAPSHOT.jar",
		"slf4j-api-1.7.30.jar", "slf4j-nop-1.7.30.jar", "commons-lang3-3.12.0.jar", "commons-math3-3.6.1.jar",
	} {
		extractCommunityAntClasses(t, filepath.Join(root, "CommunityModules", "lib", name), filepath.Join(dir, "build", "deps"))
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"all", "shiviz"} {
		t.Run(phase, func(t *testing.T) {
			// Ant forks each Java invocation. A fresh Go test process likewise
			// gives IOEnv its startup environment and isolates all TLC statics.
			cmd := exec.Command(executable, "-test.run=^TestJavaCommunityModulesAnt$", "-test.v")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), phaseVariable+"="+phase)
			if phase == "all" {
				cmd.Env = append(cmd.Env, "SOME_TEST_ENV_VAR=TLCFTW", "SOME-TEST-ENV-VAR=TLCFTW", "SOME_TEST_ENV_VAR_N23=23")
			}
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("original Ant %s invocation: %v\n%s", phase, err, output)
			}
		})
	}
}

func runJavaCommunityModulesAntPhase(t *testing.T, phase string) {
	t.Helper()
	args := []string{"-fp", "1", "-metadir", filepath.Join("build", "states"), "-cleanup", "-config", filepath.Join("tests", "AllTests.cfg")}
	wantExit := tlc.ExitStatusSuccess
	switch phase {
	case "all":
		root := filepath.Join("tests", "AllTestsUnix")
		if runtime.GOOS == "windows" {
			root = filepath.Join("tests", "AllTests")
		}
		args = append(args, root)
	case "shiviz":
		args = []string{"-metadir", filepath.Join("build", "states"), "-cleanup", "-deadlock", "-noGenerateSpecTE", filepath.Join("tests", "GH037", "ShiVizTests")}
		wantExit = tlc.ExitStatusViolationLiveness
	default:
		t.Fatalf("unknown Ant phase %q", phase)
	}
	args = append([]string{"-Dtlc2.TLC.ide=azure-pipeline", "-Dutil.ExecutionStatisticsCollector.id=01ed03e40ba44f278a934849dd2b1038"}, args...)
	opts, err := tlc.ParseTLCOptions(args)
	if err != nil {
		t.Fatal(err)
	}
	classpath, err := tlcApplicationClasspath(nil)
	if err != nil {
		t.Fatal(err)
	}
	resolver := tlc.NewSimpleFilenameToStream([]string{filepath.Dir(opts.SpecFile), "modules"}, tlc.FilenameResolverOptions{Classpath: classpath})
	opts.LoadTool = func() (*tlc.Tool, error) {
		tool, diags, err := loadTLCAppTool(opts.SpecFile, opts.ConfigFile, resolver, opts.RuntimeParams)
		requireNoErrors(t, diags)
		return tool, err
	}
	result, err := tlc.NewTLC(opts).Process(context.Background())
	if result == nil {
		t.Fatalf("original Ant %s: nil result, error %v", phase, err)
	}
	// Ant asserts successful completion of the first fork and exactly exit 13
	// for ShiViz. Retain those complete source assertions without adding gates.
	if result.ExitStatus != wantExit {
		t.Fatalf("original Ant %s exit=%d, want %d; error code=%d, error=%v", phase, result.ExitStatus, wantExit, result.ErrorCode, err)
	}
}

func extractCommunityAntClasses(t *testing.T, path, dest string) {
	t.Helper()
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if !strings.HasSuffix(file.Name, ".class") || strings.HasSuffix(file.Name, "rensenIndexLinkPrediction.class") {
			continue
		}
		name := filepath.Join(dest, filepath.FromSlash(file.Name))
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		input, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		output, err := os.Create(name)
		if err != nil {
			input.Close()
			t.Fatal(err)
		}
		_, copyErr := io.Copy(output, input)
		input.Close()
		closeErr := output.Close()
		if copyErr != nil {
			t.Fatal(copyErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
}
