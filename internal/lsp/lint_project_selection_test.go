package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
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
		tsoptions.NewParsedCommandLine(options, rootFiles, nil, tspath.ComparePathsOptions{
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
	target     string
	reads      int
	unreadable bool
}

func (fs *configReadCountingFS) ReadFile(path string) (string, bool) {
	if tspath.NormalizePath(path) == fs.target {
		fs.reads++
		if fs.unreadable {
			return "", false
		}
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

func TestProjectServiceLSPGenerationParity(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		name         string
		fixture      string
		target       string
		configDir    string
		rootDir      string
		wantConfig   string
		wantRoots    int
		project      []string
		unreadConfig string
		failConfig   string
		wantError    string
	}{
		{name: "nearest complete project", fixture: "nested", target: "pkg/src/target.ts", wantConfig: "pkg/tsconfig.json", wantRoots: 3, unreadConfig: "tsconfig.json"},
		{name: "external lint config", fixture: "nested", target: "pkg/src/target.ts", configDir: "tooling", wantConfig: "pkg/tsconfig.json", wantRoots: 3, unreadConfig: "tsconfig.json"},
		{name: "ancestor after nearest exclusion", fixture: "ancestor", target: "pkg/target.ts", wantConfig: "tsconfig.json", wantRoots: 1},
		{name: "root boundary gap", fixture: "ancestor", target: "pkg/target.ts", rootDir: "pkg", wantRoots: 1},
		{name: "disabled solution search gap", fixture: "disabled-search", target: "pkg/target.ts", wantRoots: 1},
		{name: "custom reference", fixture: "solution", target: "pkg/src/target.ts", wantConfig: "pkg/tsconfig.app.json", wantRoots: 1},
		{name: "unowned target gap", fixture: "unowned", target: "target.ts", wantRoots: 1},
		{name: "unreadable config remains an error", fixture: "unowned", target: "target.ts", failConfig: "tsconfig.json", wantError: "no parsed config returned"},
		{name: "conflicting explicit project", fixture: "nested", target: "pkg/src/target.ts", project: []string{}, wantError: "enabling parserOptions.project"},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, test.fixture))
			fileName := tspath.ResolvePath(directory, test.target)
			configDirectory := tspath.ResolvePath(directory, test.configDir)
			fsys := &configReadCountingFS{FS: bundled.WrapFS(osvfs.FS())}
			if test.unreadConfig != "" {
				fsys.target = tspath.ResolvePath(directory, test.unreadConfig)
			}
			if test.failConfig != "" {
				fsys.target = tspath.ResolvePath(directory, test.failConfig)
				fsys.unreadable = true
			}
			server := newTestServer()
			server.cwd = directory
			server.fs = fsys
			uri := documentURIFromPath(fileName)
			const editorText = "export const editor = 1;\n"
			server.documents[uri] = editorText
			options := &config.ParserOptions{ProjectService: config.BoolPtr(true), Project: test.project}
			if test.rootDir != "" {
				options.TsconfigRootDir = tspath.ResolvePath(directory, test.rootDir)
			}
			entries := config.RslintConfig{{
				Plugins:         []string{"@typescript-eslint"},
				LanguageOptions: &config.LanguageOptions{ParserOptions: options},
				Rules: config.Rules{
					"no-debugger": "error", "@typescript-eslint/no-unsafe-member-access": "error",
				},
			}}
			snapshot := documentLintSnapshotForTest(server, uri, entries, configDirectory, false, nil)
			for _, speculative := range []bool{false, true} {
				var generation linter.Generation
				var release linter.ReleaseFunc
				var err error
				text := editorText
				if speculative {
					text = "export const preview = 2;\n"
					generation, release, err = acquireSpeculativeGeneration(
						context.Background(), text, snapshot,
						server.freezeSpeculativeLintEnvironment(uri, snapshot.target),
					)
				} else {
					provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
					generation, release, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
				}
				if release != nil {
					defer release()
				}
				if test.wantError != "" {
					if err == nil || !strings.Contains(err.Error(), test.wantError) {
						t.Fatalf("speculative=%v: error=%v, want %q", speculative, err, test.wantError)
					}
					continue
				}
				if err != nil {
					t.Fatalf("speculative=%v: %v", speculative, err)
				}
				if len(generation.Native.Programs) != 1 {
					t.Fatalf("speculative=%v: programs=%d, want 1", speculative, len(generation.Native.Programs))
				}
				program := generation.Native.Programs[0]
				wantConfig := test.wantConfig
				if wantConfig != "" {
					wantConfig = tspath.ResolvePath(directory, wantConfig)
				}
				if got := program.Options().ConfigFilePath; got != wantConfig {
					t.Fatalf("speculative=%v: config=%q, want %q", speculative, got, test.wantConfig)
				}
				if len(program.RootFileNames()) != test.wantRoots {
					t.Fatalf("speculative=%v: roots=%v", speculative, program.RootFileNames())
				}
				source := program.GetSourceFile(fileName)
				if source == nil || source.Text() != text {
					t.Fatalf("speculative=%v: generation did not use its editor text", speculative)
				}
				wantRules := 2
				if wantConfig == "" {
					wantRules = 1
					if !program.Options().NoResolve.IsTrue() || !program.Options().NoLib.IsTrue() {
						t.Fatal("project gap used a dependency-resolving type context")
					}
				}
				if rules := generation.Native.RulesForFile(source); len(rules) != wantRules {
					t.Fatalf("speculative=%v: configured rules=%v, want %d", speculative, rules, wantRules)
				}
				if targets := generation.Native.TargetsByProgram; len(targets) != 1 || len(targets[0]) != 1 || targets[0][0] != fileName {
					t.Fatalf("speculative=%v: extra lint targets=%v", speculative, targets)
				}
				if server.documents[uri] != editorText {
					t.Fatal("speculative generation changed resident editor content")
				}
			}
			if test.unreadConfig != "" && fsys.reads != 0 {
				t.Fatalf("read unrelated root config %d times", fsys.reads)
			}
		})
	}
}

func TestProjectServiceLSPFrozenRootDirectory(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		name         string
		explicitRoot string
		inferredRoot bool
		wantTyped    bool
	}{
		{name: "invocation cwd leaves ancestor unowned"},
		{name: "explicit root admits ancestor", explicitRoot: ".", wantTyped: true},
		{name: "inferred root admits ancestor", inferredRoot: true, wantTyped: true},
		{name: "explicit root overrides inferred root", explicitRoot: "pkg", inferredRoot: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, "ancestor"))
			workspace := tspath.ResolvePath(directory, "pkg")
			fileName := tspath.ResolvePath(workspace, "target.ts")
			server := newTestServer()
			server.cwd = workspace
			server.fs = bundled.WrapFS(osvfs.FS())
			server.lintPrograms = newLintProgramStore(server)
			server.lintPrograms.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
			uri := documentURIFromPath(fileName)
			const editorText = "debugger;\ndeclare const value: any;\nexport const result = value.member;\n"
			server.documents[uri] = editorText
			options := &config.ParserOptions{ProjectService: config.BoolPtr(true)}
			if test.explicitRoot != "" {
				options.TsconfigRootDir = tspath.ResolvePath(directory, test.explicitRoot)
			}
			entries := config.RslintConfig{{
				Plugins:         []string{"@typescript-eslint"},
				LanguageOptions: &config.LanguageOptions{ParserOptions: options},
				Rules: config.Rules{
					"no-debugger": "error", "@typescript-eslint/no-unsafe-member-access": "error",
				},
			}}
			if test.inferredRoot {
				entries[0].InferredTSConfigRootDirs = []string{directory}
			}
			installJSConfigsForTest(server, map[string]config.RslintConfig{directory: entries})
			snapshot := server.documentLintSnapshot(uri)
			if snapshot.cwd != workspace || snapshot.target.ConfigDirectory != directory {
				t.Fatalf("snapshot cwd=%q owner=%q, want nested workspace %q and parent owner %q", snapshot.cwd, snapshot.target.ConfigDirectory, workspace, directory)
			}
			if snapshot.projectPolicyError != nil {
				t.Fatal(snapshot.projectPolicyError)
			}
			environment := server.freezeSpeculativeLintEnvironment(uri, snapshot.target)
			// A later invocation directory would admit the ancestor project. Both
			// adapters must continue using the policy frozen for this operation.
			server.cwd = directory
			for _, speculative := range []bool{false, true} {
				var generation linter.Generation
				var release linter.ReleaseFunc
				var err error
				wantText := editorText
				if speculative {
					wantText = "// speculative text\n" + editorText
					generation, release, err = acquireSpeculativeGeneration(context.Background(), wantText, snapshot, environment)
				} else {
					provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
					generation, release, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
				}
				if release != nil {
					defer release()
				}
				if err != nil || len(generation.Native.Programs) != 1 {
					t.Fatalf("speculative=%v: programs=%d error=%v", speculative, len(generation.Native.Programs), err)
				}
				program := generation.Native.Programs[0]
				wantConfig := ""
				if test.wantTyped {
					wantConfig = tspath.ResolvePath(directory, "tsconfig.json")
				}
				if program.Options().ConfigFilePath != wantConfig {
					t.Fatalf("speculative=%v: config=%q, want %q", speculative, program.Options().ConfigFilePath, wantConfig)
				}
				source := program.GetSourceFile(fileName)
				if source == nil || source.Text() != wantText {
					t.Fatalf("speculative=%v: generation did not use its editor text", speculative)
				}
				foundSyntax, foundTyped := false, false
				for _, configured := range generation.Native.RulesForFile(source) {
					foundSyntax = foundSyntax || configured.Name == "no-debugger"
					foundTyped = foundTyped || configured.RequiresTypeInfo
				}
				if !foundSyntax || foundTyped != test.wantTyped {
					t.Fatalf("speculative=%v: syntax=%v typed=%v, want typed=%v", speculative, foundSyntax, foundTyped, test.wantTyped)
				}
			}
			if server.documents[uri] != editorText {
				t.Fatal("speculative generation changed resident editor content")
			}
		})
	}
}

func TestProjectServiceLSPGapKeepsDiagnosticsAndFixes(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, resident := range []bool{false, true} {
		for _, sourceKind := range []string{"syntax-rule", "syntax-error", "unsaved-file"} {
			malformed := sourceKind == "syntax-error"
			name := "standalone"
			if resident {
				name = "resident"
			}
			name += "/" + sourceKind
			t.Run(name, func(t *testing.T) {
				directory := tspath.NormalizePath(archive.Materialize(t, "unowned"))
				fileName := tspath.ResolvePath(directory, "target.ts")
				if sourceKind == "unsaved-file" {
					if err := os.Remove(fileName); err != nil {
						t.Fatal(err)
					}
				}
				server := newTestServer()
				server.cwd = directory
				server.fs = bundled.WrapFS(osvfs.FS())
				if resident {
					server.lintPrograms = newLintProgramStore(server)
					server.lintPrograms.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
				}
				uri := documentURIFromPath(fileName)
				content := "declare const opaque: any;\nexport const result = (() => { var value = opaque.member; return value; })();\n"
				if malformed {
					content = "var value = ;\n"
				}
				server.documents[uri] = content
				entries := config.RslintConfig{{
					Plugins:         []string{"@typescript-eslint"},
					LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}},
					Rules:           config.Rules{"no-var": "error", "@typescript-eslint/no-unsafe-member-access": "error"},
				}}
				snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
				preview := "// speculative text\n" + content
				for _, speculative := range []bool{false, true} {
					var result linter.PipelineResult
					var err error
					if speculative {
						result, err = speculativePipelineResultForTest(server, context.Background(), uri, preview, snapshot)
					} else {
						result, err = configuredDocumentPipelineResultForTest(server, context.Background(), uri, entries, directory, false, nil)
					}
					if err != nil {
						t.Fatalf("speculative=%v: %v", speculative, err)
					}
					observation := result.Observation.Native
					if observation.HasTargetSyntaxErrors != malformed || len(observation.Diagnostics) != 1 {
						t.Fatalf("speculative=%v: unexpected gap diagnostics %+v", speculative, observation.Diagnostics)
					}
					diagnostic := observation.Diagnostics[0]
					if malformed {
						if !strings.HasPrefix(diagnostic.RuleName, "TypeScript(TS") {
							t.Fatalf("speculative=%v: syntax diagnostic was lost: %+v", speculative, diagnostic)
						}
					} else if diagnostic.RuleName != "no-var" {
						t.Fatalf("speculative=%v: expected only the syntax rule, got %+v", speculative, diagnostic)
					}
				}
				wantFixed := preview
				if !malformed {
					wantFixed = strings.Replace(preview, "var value", "let value", 1)
				}
				if fixed := runSpeculativeFixAllForTest(t, server, context.Background(), uri, preview, snapshot); fixed != wantFixed {
					t.Fatalf("gap fix-all=%q, want %q", fixed, wantFixed)
				}
				if server.documents[uri] != content {
					t.Fatal("gap fix-all mutated the resident editor text")
				}
				if resident && len(server.lintPrograms.programs) != 0 {
					t.Fatal("a source-only fallback became a resident configured Program")
				}
			})
		}
	}
}

func TestProjectServiceLSPTypedGapTyped(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "ancestor"))
	fileName := tspath.ResolvePath(directory, "pkg/target.ts")
	configPath := tspath.ResolvePath(directory, "tsconfig.json")
	server := newTestServer()
	server.cwd = directory
	fsys := &configReadCountingFS{FS: bundled.WrapFS(osvfs.FS()), target: configPath}
	server.fs = fsys
	server.lintPrograms = newLintProgramStore(server)
	server.lintPrograms.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	uri := documentURIFromPath(fileName)
	const content = "declare const opaque: any;\nexport const result = (() => { var value = opaque.member; return value; })();\n"
	server.documents[uri] = content
	entries := config.RslintConfig{{
		Plugins:         []string{"@typescript-eslint"},
		LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}},
		Rules:           config.Rules{"no-var": "error", "@typescript-eslint/no-unsafe-member-access": "error"},
	}}
	snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
	var firstTypedSource *ast.SourceFile
	for _, phase := range []struct {
		name       string
		config     string
		typed      bool
		unreadable bool
	}{
		{name: "typed", config: `{"compilerOptions":{"noLib":true},"files":["pkg/target.ts"]}`, typed: true},
		{name: "gap", config: `{"compilerOptions":{"noLib":true},"files":[]}`},
		{name: "unreadable-config", config: `{"compilerOptions":{"noLib":true},"files":["pkg/target.ts"]}`, unreadable: true},
		{name: "typed-again", config: `{"compilerOptions":{"noLib":true,"strict":true},"files":["pkg/target.ts"]}`, typed: true},
	} {
		t.Run(phase.name, func(t *testing.T) {
			if err := os.WriteFile(configPath, []byte(phase.config), 0o644); err != nil {
				t.Fatal(err)
			}
			fsys.unreadable = phase.unreadable
			server.lintPrograms.DidChangeWatchedFiles([]*lsproto.FileEvent{{Uri: documentURIFromPath(configPath), Type: lsproto.FileChangeTypeChanged}})
			var previousSource *ast.SourceFile
			for _, speculative := range []bool{false, false, true} {
				var generation linter.Generation
				var release linter.ReleaseFunc
				var err error
				wantText := content
				if speculative {
					wantText = "// speculative generation\n" + content
					generation, release, err = acquireSpeculativeGeneration(context.Background(), wantText, snapshot,
						server.freezeSpeculativeLintEnvironment(uri, snapshot.target))
				} else {
					provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
					generation, release, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
				}
				if release != nil {
					defer release()
				}
				if phase.unreadable {
					if err == nil || !strings.Contains(err.Error(), "no parsed config returned") {
						t.Fatalf("speculative=%v: unreadable config was reduced to a gap: %v", speculative, err)
					}
					continue
				}
				if err != nil || len(generation.Native.Programs) != 1 {
					t.Fatalf("speculative=%v: programs=%d error=%v", speculative, len(generation.Native.Programs), err)
				}
				program := generation.Native.Programs[0]
				wantConfig := ""
				if phase.typed {
					wantConfig = configPath
				}
				if program.Options().ConfigFilePath != wantConfig {
					t.Fatalf("speculative=%v: config=%s, want %s", speculative, program.Options().ConfigFilePath, wantConfig)
				}
				source := program.GetSourceFile(fileName)
				if source == nil || source.Text() != wantText {
					t.Fatalf("speculative=%v: generation used stale editor text", speculative)
				}
				foundSyntax, foundTyped := false, false
				for _, configured := range generation.Native.RulesForFile(source) {
					foundSyntax = foundSyntax || configured.Name == "no-var"
					foundTyped = foundTyped || configured.RequiresTypeInfo
				}
				if !foundSyntax || foundTyped != phase.typed {
					t.Fatalf("speculative=%v: syntax=%v typed=%v, want typed=%v", speculative, foundSyntax, foundTyped, phase.typed)
				}
				if !speculative && phase.typed {
					if previousSource != nil && previousSource != source {
						t.Fatal("cache hit rebuilt an unchanged configured Program")
					}
					previousSource = source
					if firstTypedSource == nil {
						firstTypedSource = source
					} else if phase.name == "typed-again" && (source == firstTypedSource || !program.Options().Strict.IsTrue()) {
						t.Fatal("restored membership reused the old configured generation")
					}
				}
			}
			if server.documents[uri] != content {
				t.Fatal("speculative transition changed editor text")
			}
			if !phase.typed && len(server.lintPrograms.programs) != 0 {
				t.Fatal("gap retained a configured Program after membership invalidation")
			}
		})
	}
}

func TestProjectServiceLSPUsesReferencedEditorSources(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "reference-sources"))
	targetPath := tspath.ResolvePath(directory, "app/src/main.ts")
	referencePath := tspath.ResolvePath(directory, "lib/src/value.ts")
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	uri := documentURIFromPath(targetPath)
	const targetContent = "import { value } from '../../lib/src/value';\nexport const result: number = value;\n"
	const referenceContent = "export const value = 42;\n"
	server.documents[uri] = targetContent
	server.documents[documentURIFromPath(referencePath)] = referenceContent
	entries := config.RslintConfig{{LanguageOptions: &config.LanguageOptions{
		ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)},
	}}}
	snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
	for _, speculative := range []bool{false, true} {
		var generation linter.Generation
		var err error
		if speculative {
			generation, _, err = acquireSpeculativeGeneration(context.Background(), targetContent, snapshot,
				server.freezeSpeculativeLintEnvironment(uri, snapshot.target))
		} else {
			provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
			generation, _, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
		}
		if err != nil {
			t.Fatalf("speculative=%v: %v", speculative, err)
		}
		if len(generation.Native.Programs) != 1 {
			t.Fatalf("speculative=%v: programs=%d", speculative, len(generation.Native.Programs))
		}
		program := generation.Native.Programs[0]
		if source := program.GetSourceFile(referencePath); source == nil || source.Text() != referenceContent {
			t.Fatalf("speculative=%v: did not load unsaved referenced source", speculative)
		}
		if program.GetSourceFile(tspath.ResolvePath(directory, "lib/dist/value.d.ts")) != nil {
			t.Fatalf("speculative=%v: used stale declaration output", speculative)
		}
	}
}

func TestDocumentProjectPolicyUsesMatchingEntries(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "nested"))
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	entries := config.RslintConfig{
		{LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}}},
		{Files: []string{"**/*.js"}, LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{
			ProjectService: config.BoolPtr(false), Project: []string{"./missing.json"},
		}}},
	}
	server.jsConfigs = map[string]config.RslintConfig{directory: entries}
	if err := server.rebuildTsConfigPaths(); err != nil {
		t.Fatalf("eagerly expanded an unrelated flat-config project: %v", err)
	}
	uri := documentURIFromPath(tspath.ResolvePath(directory, "pkg/src/target.ts"))
	snapshot := resolveDocumentLintSnapshotConfig(documentLintSnapshotForTest(server, uri, entries, directory, false, nil), server.fs)
	if snapshot.projectPolicyError != nil || !snapshot.projectPolicy.ProjectService || len(snapshot.typeScriptConfigPaths) != 0 {
		t.Fatalf("TypeScript policy=%+v paths=%v error=%v", snapshot.projectPolicy, snapshot.typeScriptConfigPaths, snapshot.projectPolicyError)
	}
	jsURI := documentURIFromPath(tspath.ResolvePath(directory, "target.js"))
	jsSnapshot := resolveDocumentLintSnapshotConfig(documentLintSnapshotForTest(server, jsURI, entries, directory, false, nil), server.fs)
	if jsSnapshot.projectPolicyError == nil || !strings.Contains(jsSnapshot.projectPolicyError.Error(), "missing.json") {
		t.Fatalf("matching explicit project error=%v", jsSnapshot.projectPolicyError)
	}
}

func TestProjectServiceLSPLexicalAliases(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "aliases"))
	realFile := tspath.ResolvePath(directory, "real/file.ts")
	aliasFile := tspath.ResolvePath(directory, "alias/file.ts")
	if err := os.Symlink(realFile, aliasFile); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	server.lintPrograms = newLintProgramStore(server)
	server.lintPrograms.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
	entries := config.RslintConfig{{LanguageOptions: &config.LanguageOptions{
		ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)},
	}}}
	sources := make(map[string]*ast.SourceFile)
	for _, fileName := range []string{aliasFile, realFile, aliasFile, realFile} {
		uri := documentURIFromPath(fileName)
		const content = "export const value = 1;\n"
		server.documents[uri] = content
		snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
		provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
		generation, release, err := provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
		if release != nil {
			defer release()
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(generation.Native.Programs) != 1 {
			t.Fatalf("%s: programs=%d", fileName, len(generation.Native.Programs))
		}
		program := generation.Native.Programs[0]
		wantConfig := tspath.ResolvePath(tspath.GetDirectoryPath(fileName), "tsconfig.json")
		if program.Options().ConfigFilePath != wantConfig {
			t.Fatalf("%s: config=%q, want %q", fileName, program.Options().ConfigFilePath, wantConfig)
		}
		if program.Options().Strict.IsTrue() != (fileName == realFile) {
			t.Fatalf("%s: selected the physical alias's compiler options", fileName)
		}
		source := program.GetSourceFile(fileName)
		if previous := sources[fileName]; previous != nil && previous != source {
			t.Fatalf("%s: unchanged lexical project was rebuilt", fileName)
		}
		sources[fileName] = source
	}
	if sources[aliasFile] == sources[realFile] {
		t.Fatal("lexical alias projects shared one source identity")
	}
}
