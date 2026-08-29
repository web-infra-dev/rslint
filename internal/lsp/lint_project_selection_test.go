package lsp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func lintProjectMetadataForTest(
	configPath string,
	rootFiles []string,
	options *core.CompilerOptions,
	fs vfs.FS,
) *lintProjectMetadata {
	if options == nil {
		options = &core.CompilerOptions{}
	}
	return newLintProjectMetadata(
		configPath,
		tsoptions.NewParsedCommandLine(options, rootFiles, tspath.ComparePathsOptions{
			CurrentDirectory:          tspath.GetDirectoryPath(configPath),
			UseCaseSensitiveFileNames: true,
		}),
		fs,
	)
}

func TestSelectConfiguredLintProjectDirectRootOutranksEarlierImport(t *testing.T) {
	const (
		firstConfig  = "/repo/tsconfig.import.json"
		secondConfig = "/repo/tsconfig.direct.json"
		targetPath   = "/repo/src/target.ts"
	)
	metadata := map[string]*lintProjectMetadata{
		firstConfig:  lintProjectMetadataForTest(firstConfig, []string{"/repo/importer.ts"}, nil, nil),
		secondConfig: lintProjectMetadataForTest(secondConfig, []string{targetPath}, nil, nil),
	}
	sourceFile := &ast.SourceFile{}
	var programCalls []string
	selected, found, err := selectConfiguredLintProject(
		[]string{firstConfig, secondConfig},
		target.File{PathIdentity: config.PathIdentity{Path: targetPath, CanonicalPath: targetPath}},
		lintProjectLoaders{
			metadata: func(configPath string) (*lintProjectMetadata, bool, error) {
				return metadata[configPath], true, nil
			},
			program: func(configPath string) (*compiler.Program, *ast.SourceFile, error) {
				programCalls = append(programCalls, configPath)
				return new(compiler.Program), sourceFile, nil
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !found || selected.configPath != secondConfig || !selected.directRoot {
		t.Fatalf("selected project = %+v, want direct %q", selected, secondConfig)
	}
	if len(programCalls) != 1 || programCalls[0] != secondConfig {
		t.Fatalf("Program calls = %v, want only direct winner", programCalls)
	}
}

func TestSelectConfiguredLintProjectFallbackOrderAndExtensionFilter(t *testing.T) {
	const (
		firstConfig  = "/repo/tsconfig.ts.json"
		secondConfig = "/repo/tsconfig.js.json"
		targetPath   = "/repo/src/target.js"
	)
	metadata := map[string]*lintProjectMetadata{
		firstConfig: lintProjectMetadataForTest(
			firstConfig,
			[]string{"/repo/first.ts"},
			&core.CompilerOptions{AllowJs: core.TSFalse},
			nil,
		),
		secondConfig: lintProjectMetadataForTest(
			secondConfig,
			[]string{"/repo/second.ts"},
			&core.CompilerOptions{AllowJs: core.TSTrue},
			nil,
		),
	}
	sourceFile := &ast.SourceFile{}
	var programCalls []string
	selected, found, err := selectConfiguredLintProject(
		[]string{firstConfig, secondConfig},
		target.File{PathIdentity: config.PathIdentity{Path: targetPath, CanonicalPath: targetPath}},
		lintProjectLoaders{
			metadata: func(configPath string) (*lintProjectMetadata, bool, error) {
				return metadata[configPath], true, nil
			},
			program: func(configPath string) (*compiler.Program, *ast.SourceFile, error) {
				programCalls = append(programCalls, configPath)
				return new(compiler.Program), sourceFile, nil
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !found || selected.configPath != secondConfig || selected.directRoot {
		t.Fatalf("selected project = %+v, want import fallback %q", selected, secondConfig)
	}
	if len(programCalls) != 1 || programCalls[0] != secondConfig {
		t.Fatalf("Program calls = %v, want unsupported project skipped", programCalls)
	}
}

type configReadCountingFS struct {
	vfs.FS
	target string
	reads  int
}

func (fs *configReadCountingFS) ReadFile(path string) (string, bool) {
	if tspath.NormalizePath(path) == fs.target {
		fs.reads++
	}
	return fs.FS.ReadFile(path)
}

func TestStandaloneLintProjectRequestReusesParsedConfigSnapshot(t *testing.T) {
	dir := t.TempDir()
	firstSource := filepath.Join(dir, "first.ts")
	secondSource := filepath.Join(dir, "second.ts")
	for _, source := range []string{firstSource, secondSource} {
		if err := os.WriteFile(source, []byte("export const value = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	configPath := filepath.Join(dir, "tsconfig.json")
	if err := os.WriteFile(configPath, []byte(`{"files":["first.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &configReadCountingFS{
		FS:     bundled.WrapFS(osvfs.FS()),
		target: tspath.NormalizePath(configPath),
	}
	request := newStandaloneLintProjectRequestWithFS(
		target.File{
			PathIdentity: config.PathIdentity{
				Path:          firstSource,
				CanonicalPath: firstSource,
			},
		},
		fs,
	)
	metadata, err := request.metadata(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !metadata.Contains(firstSource, "") {
		t.Fatal("initial parsed metadata does not contain its configured root")
	}
	if err := os.WriteFile(configPath, []byte(`{"files":["second.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	program, sourceFile, err := request.program(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if sourceFile == nil || sourceFile.FileName() != tspath.NormalizePath(firstSource) {
		t.Fatalf("Program did not use the selected parsed snapshot: %v", sourceFile)
	}
	if program.GetSourceFile(tspath.NormalizePath(secondSource)) != nil {
		t.Fatal("Program reparsed the changed config during the same selection request")
	}
	if fs.reads != 1 {
		t.Fatalf("tsconfig read count = %d, want 1", fs.reads)
	}
}

func TestRunConfiguredLintForContentDirectRootSkipsEarlierImportProgram(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "src", "target.ts")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatal(err)
	}
	const targetContent = "export const target = 1;\n"
	if err := os.WriteFile(targetPath, []byte(targetContent), 0o644); err != nil {
		t.Fatal(err)
	}
	importerPath := filepath.Join(dir, "importer.ts")
	if err := os.WriteFile(
		importerPath,
		[]byte(`import "./src/target";`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	importConfig := filepath.Join(dir, "tsconfig.import.json")
	directConfig := filepath.Join(dir, "tsconfig.direct.json")
	if err := os.WriteFile(importConfig, []byte(`{"files":["importer.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directConfig, []byte(`{"files":["src/target.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	fs := &configReadCountingFS{
		FS:     bundled.WrapFS(osvfs.FS()),
		target: tspath.NormalizePath(importerPath),
	}
	server := newTestServer()
	server.cwd = dir
	server.fs = fs
	uri := documentURIFromPath(targetPath)
	server.documents[uri] = targetContent

	if _, err := configuredSpeculativePipelineResultForTest(server,
		uri,
		context.Background(),
		targetContent,
		config.RslintConfig{{}},
		dir,
		false,
		[]string{importConfig, directConfig},
	); err != nil {
		t.Fatal(err)
	}
	if fs.reads != 0 {
		t.Fatalf("fix-all built the earlier import-only Program %d times", fs.reads)
	}
}

func TestLintSessionProjectRootCacheUsesCommandLineGeneration(t *testing.T) {
	dir := t.TempDir()
	firstSource := filepath.Join(dir, "first.ts")
	secondSource := filepath.Join(dir, "second.ts")
	for _, source := range []string{firstSource, secondSource} {
		if err := os.WriteFile(source, []byte("export const value = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fs := bundled.WrapFS(osvfs.FS())
	host := utils.CreateCompilerHost(dir, fs)
	firstProgram, err := utils.CreateProgramFromOptionsLenient(
		true,
		&core.CompilerOptions{},
		[]string{firstSource},
		host,
	)
	if err != nil {
		t.Fatal(err)
	}
	secondProgram, err := utils.CreateProgramFromOptionsLenient(
		true,
		&core.CompilerOptions{},
		[]string{secondSource},
		host,
	)
	if err != nil {
		t.Fatal(err)
	}

	cache := newLintSessionProjectRootCache()
	configPath := filepath.Join(dir, "tsconfig.json")
	first := cache.metadata(configPath, firstProgram.CommandLine(), fs)
	if reused := cache.metadata(configPath, firstProgram.CommandLine(), fs); reused != first {
		t.Fatal("unchanged Session command line rebuilt its root index")
	}
	second := cache.metadata(configPath, secondProgram.CommandLine(), fs)
	if second == first || !second.Contains(secondSource, "") || second.Contains(firstSource, "") {
		t.Fatal("new Session command line did not replace cached root metadata")
	}
}

func TestResolveTsConfigPathsPreservesSymlinkDeclarationPath(t *testing.T) {
	root := t.TempDir()
	realDir := filepath.Join(root, "real")
	aliasDir := filepath.Join(root, "alias")
	for _, dir := range []string{realDir, aliasDir} {
		if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	realSource := filepath.Join(realDir, "src", "real.ts")
	aliasSource := filepath.Join(aliasDir, "src", "alias.ts")
	for _, source := range []string{realSource, aliasSource} {
		if err := os.WriteFile(source, []byte("export const value = 1;\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	realConfig := filepath.Join(realDir, "tsconfig.json")
	if err := os.WriteFile(realConfig, []byte(`{"include":["src/**/*.ts"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	aliasConfig := filepath.Join(aliasDir, "tsconfig.json")
	if err := os.Symlink(realConfig, aliasConfig); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	fs := bundled.WrapFS(osvfs.FS())
	paths, err := resolveTsConfigPathsWithFS(config.RslintConfig{{
		LanguageOptions: &config.LanguageOptions{
			ParserOptions: &config.ParserOptions{Project: []string{"./tsconfig.json"}},
		},
	}}, aliasDir, fs)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0] != tspath.NormalizePath(aliasConfig) {
		t.Fatalf("resolved project paths = %v, want lexical %q", paths, aliasConfig)
	}
	metadata, err := parseStandaloneLintProject(paths[0], fs, fs)
	if err != nil {
		t.Fatal(err)
	}
	if !metadata.Contains(aliasSource, "") || metadata.Contains(realSource, "") {
		t.Fatal("symlinked tsconfig did not resolve includes from its declared directory")
	}
}
