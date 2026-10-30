package contentmappers_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program/contentmappers"
	"github.com/web-infra-dev/rslint/internal/testutil/contentmappertest"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const mapperPackage = "fake-mapper"

// writeMappedProject writes a project whose index.ts imports a file only the
// test mapper can provide, and returns its tsconfig path.
func writeMappedProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	contentmappertest.WritePackage(t, dir, mapperPackage)
	files := map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "strict": true, "noEmit": true },
  "files": ["index.ts"],
  "contentMappers": [{ "package": "` + mapperPackage + `", "extensions": ["` + contentmappertest.Extension + `"] }]
}`,
		"index.ts":                           "import { answer } from './data" + contentmappertest.Extension + "';\nexport const doubled: number = answer * 2;\n",
		"data" + contentmappertest.Extension: "answer = 42\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return tspath.NormalizePath(filepath.Join(dir, "tsconfig.json"))
}

func parse(t *testing.T, runtime *contentmappers.Runtime, configPath string) *tsoptions.ParsedCommandLine {
	t.Helper()
	host := utils.CreateCompilerHost(tspath.GetDirectoryPath(configPath), bundled.WrapFS(osvfs.FS()))
	config, _ := tsoptions.GetParsedCommandLineOfConfigFile(configPath, runtime.ParseOptions(), nil, host, nil)
	if config == nil {
		t.Fatalf("no parsed config for %s", configPath)
	}
	return config
}

func build(t *testing.T, runtime *contentmappers.Runtime, config *tsoptions.ParsedCommandLine) *compiler.Program {
	t.Helper()
	project, _ := runtime.Open(config)
	host := utils.CreateCompilerHostWithContentMappers(tspath.GetDirectoryPath(config.ConfigName()), bundled.WrapFS(osvfs.FS()), project)
	program, err := utils.CreateProgramFromParsedConfigLenient(true, config, host)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func semanticCodes(program *compiler.Program, fileName string) []int32 {
	file := program.GetSourceFile(fileName)
	if file == nil {
		return nil
	}
	return mapCodes(program.GetSemanticDiagnostics(context.Background(), file))
}

func mapCodes(diagnostics []*ast.Diagnostic) []int32 {
	codes := make([]int32, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		codes = append(codes, diagnostic.Code())
	}
	return codes
}

func TestNilRuntimeRunsNothing(t *testing.T) {
	t.Parallel()
	var runtime *contentmappers.Runtime
	if runtime.Enabled() {
		t.Fatal("nil Runtime must not be enabled")
	}
	if runtime.ParseOptions().RunExternalCode != core.TSUnknown {
		t.Fatal("nil Runtime must not request runExternalCode")
	}
	config := parse(t, runtime, writeMappedProject(t))
	if runtime.Observe(config) {
		t.Fatal("nil Runtime must not record configs")
	}
	if project, release := runtime.Open(config); project != nil || release() != nil {
		t.Fatal("nil Runtime must not open projects")
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDisabledRuntimeReportsIgnoredMappers(t *testing.T) {
	t.Parallel()
	spawner := &contentmappertest.Spawner{Output: "export const answer = 42;\n"}
	runtime := contentmappers.New(contentmappers.Options{Spawner: spawner})
	defer runtime.Close()
	configPath := writeMappedProject(t)

	config := parse(t, runtime, configPath)
	if len(config.ContentMappers()) != 0 {
		t.Fatal("ts-go must drop mappers without runExternalCode")
	}
	if !runtime.Observe(config) {
		t.Fatal("first Observe must record the config")
	}
	if runtime.Observe(parse(t, runtime, configPath)) {
		t.Fatal("a config must be recorded once")
	}
	if got := runtime.IgnoredConfigs(); !slices.Equal(got, []string{configPath}) {
		t.Fatalf("IgnoredConfigs() = %v, want [%s]", got, configPath)
	}
	var warnings strings.Builder
	runtime.WriteIgnoredWarnings(&warnings, "--run-external-code")
	if want := "Warning: " + configPath + " declares contentMappers, which run only with --run-external-code; the files they map provide no types.\n"; warnings.String() != want {
		t.Fatalf("warnings = %q, want %q", warnings.String(), want)
	}

	program := build(t, runtime, config)
	if program.GetSourceFile(tspath.GetDirectoryPath(configPath)+"/data"+contentmappertest.Extension) != nil {
		t.Fatal("a dropped mapper must not map files")
	}
	if codes := semanticCodes(program, tspath.GetDirectoryPath(configPath)+"/index.ts"); !slices.Contains(codes, 2307) {
		t.Fatalf("index.ts diagnostics = %v, want TS2307", codes)
	}
	if spawner.Spawns() != 0 {
		t.Fatalf("disabled Runtime spawned %d mappers", spawner.Spawns())
	}
}

func TestEnabledRuntimeMapsFilesAndStopsMappers(t *testing.T) {
	t.Parallel()
	spawner := &contentmappertest.Spawner{Output: "export const answer: number = 42;\n"}
	runtime := contentmappers.New(contentmappers.Options{Enabled: true, Spawner: spawner})
	configPath := writeMappedProject(t)
	dir := tspath.GetDirectoryPath(configPath)

	config := parse(t, runtime, configPath)
	if runtime.Observe(config) || len(runtime.IgnoredConfigs()) != 0 {
		t.Fatal("an enabled Runtime must not report ignored mappers")
	}
	program := build(t, runtime, config)
	mapped := program.GetSourceFile(dir + "/data" + contentmappertest.Extension)
	if mapped == nil || mapped.ContentMapper() == "" {
		t.Fatal("the mapper must provide data" + contentmappertest.Extension)
	}
	if codes := semanticCodes(program, dir+"/index.ts"); len(codes) != 0 {
		t.Fatalf("index.ts diagnostics = %v, want none", codes)
	}
	if spawner.Spawns() != 1 {
		t.Fatalf("spawned %d mappers, want 1", spawner.Spawns())
	}

	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
	if spawner.Closes() != 1 {
		t.Fatalf("stopped %d mappers, want 1", spawner.Closes())
	}
	if project, _ := runtime.Open(config); project != nil {
		t.Fatal("a closed Runtime must not open projects")
	}
}

func TestReleaseStopsMapperBeforeRuntimeCloses(t *testing.T) {
	t.Parallel()
	spawner := &contentmappertest.Spawner{Output: "export const answer: number = 42;\n"}
	runtime := contentmappers.New(contentmappers.Options{Enabled: true, Spawner: spawner})
	defer runtime.Close()
	config := parse(t, runtime, writeMappedProject(t))

	project, release := runtime.Open(config)
	host := utils.CreateCompilerHostWithContentMappers(tspath.GetDirectoryPath(config.ConfigName()), bundled.WrapFS(osvfs.FS()), project)
	if _, err := utils.CreateProgramFromParsedConfigLenient(true, config, host); err != nil {
		t.Fatal(err)
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
	if err := release(); err != nil {
		t.Fatal("release must be idempotent:", err)
	}
	if spawner.Closes() != 1 {
		t.Fatalf("stopped %d mappers after the last release, want 1", spawner.Closes())
	}
}

func TestRetainKeepsMapperRunningBetweenBuilds(t *testing.T) {
	t.Parallel()
	spawner := &contentmappertest.Spawner{Output: "export const answer: number = 42;\n"}
	runtime := contentmappers.New(contentmappers.Options{Enabled: true, Spawner: spawner})
	configPath := writeMappedProject(t)

	// Each build parses the tsconfig again and releases its project once the
	// Program is built, as a standalone language server request does.
	for range 2 {
		config := parse(t, runtime, configPath)
		project, release := runtime.Open(config)
		runtime.Retain(config)
		host := utils.CreateCompilerHostWithContentMappers(tspath.GetDirectoryPath(configPath), bundled.WrapFS(osvfs.FS()), project)
		if _, err := utils.CreateProgramFromParsedConfigLenient(true, config, host); err != nil {
			t.Fatal(err)
		}
		if err := release(); err != nil {
			t.Fatal(err)
		}
	}
	if spawner.Spawns() != 1 || spawner.Closes() != 0 {
		t.Fatalf("%d mappers started, %d stopped; want 1, 0", spawner.Spawns(), spawner.Closes())
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
	if spawner.Closes() != 1 {
		t.Fatalf("stopped %d mappers after Close, want 1", spawner.Closes())
	}

	var disabled *contentmappers.Runtime
	disabled.Retain(parse(t, runtime, configPath))
}
