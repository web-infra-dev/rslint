package lsp

import (
	"context"
	"encoding/json"
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
	"github.com/web-infra-dev/rslint/internal/rule"
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
		"",
		target.File{PathIdentity: config.PathIdentity{Path: targetPath, CanonicalPath: targetPath}},
		nil,
		lintProjectLoaders{
			metadata: func(configPath string) (*lintProjectMetadata, bool, error) {
				return metadata[configPath], true, nil
			},
			program: func(metadata *lintProjectMetadata) (*compiler.Program, *ast.SourceFile, error) {
				configPath := metadata.configPath
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
		"",
		target.File{PathIdentity: config.PathIdentity{Path: targetPath, CanonicalPath: targetPath}},
		nil,
		lintProjectLoaders{
			metadata: func(configPath string) (*lintProjectMetadata, bool, error) {
				return metadata[configPath], true, nil
			},
			program: func(metadata *lintProjectMetadata) (*compiler.Program, *ast.SourceFile, error) {
				configPath := metadata.configPath
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
	if lintProgramLexicalPathID(path, fs.FS) == lintProgramLexicalPathID(fs.target, fs.FS) {
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
	program, sourceFile, err := request.program(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if sourceFile == nil || lintProgramLexicalPathID(sourceFile.FileName(), fs) != lintProgramLexicalPathID(firstSource, fs) {
		t.Fatalf("Program did not use the selected parsed snapshot: %v", sourceFile)
	}
	if program.GetSourceFile(tspath.NormalizePath(secondSource)) != nil {
		t.Fatal("Program reparsed the changed config during the same selection request")
	}
	if fs.reads != 1 {
		t.Fatalf("tsconfig read count = %d, want 1", fs.reads)
	}
}

func TestStandaloneLintProjectRequestKeepsReferenceSnapshot(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "snapshot-references"))
	rootPath := tspath.ResolvePath(directory, "app/tsconfig.json")
	refPath := tspath.ResolvePath(directory, "lib/tsconfig.json")
	initial, err := os.ReadFile(refPath)
	if err != nil {
		t.Fatal(err)
	}
	// Include a reference cycle: the compiler's root identity must remain its
	// service clone even though the request also retains the authored root.
	cyclic := strings.TrimSpace(string(initial))
	cyclic = strings.TrimSuffix(cyclic, "}") + `,"references":[{"path":"../app"}]}`
	if err := os.WriteFile(refPath, []byte(cyclic), 0o644); err != nil {
		t.Fatal(err)
	}
	fsys := bundled.WrapFS(osvfs.FS())
	target := lspConfigTarget(tspath.ResolvePath(directory, "app/target.ts"), directory, fsys)
	request := newStandaloneLintProjectRequestWithFS(target, fsys)
	request.sourceReferences = true
	metadata, err := request.metadata(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := request.metadata(refPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(refPath, []byte(strings.ReplaceAll(cyclic, "unsafe.d.ts", "safe.d.ts")), 0o644); err != nil {
		t.Fatal(err)
	}
	program, source, err := request.program(metadata)
	if err != nil || source == nil {
		t.Fatalf("Program source=%v error=%v", source, err)
	}
	found := false
	program.RangeResolvedProjectReference(func(_ tspath.Path, parsed, _ *tsoptions.ParsedCommandLine, _ int) bool {
		if parsed != nil && lintProgramLexicalPathID(parsed.ConfigName(), fsys) == lintProgramLexicalPathID(refPath, fsys) {
			found = true
			if parsed != reference.commandLine {
				t.Error("construction reparsed a selected reference snapshot")
			}
		}
		return true
	})
	if !found {
		t.Fatal("Program omitted its selected reference")
	}
}

func TestLintProjectSnapshotNormalizesWindowsDrive(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	files := make(map[string]string)
	names, err := archive.FileNames("snapshot-options")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		content, err := archive.ReadFile("snapshot-options/" + name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath("c:/Repo", name)] = string(content)
	}
	fsys := utils.NewOverlayVFS(&caseInsensitiveLSPTestFS{}, files)
	upper, err := parseStandaloneLintProject("C:/Repo/tsconfig.json", fsys, fsys)
	if err != nil {
		t.Fatal(err)
	}
	lower, err := parseStandaloneLintProject("c:/Repo/tsconfig.json", fsys, fsys)
	if err != nil {
		t.Fatal(err)
	}
	if upper.configPath != "C:/Repo/tsconfig.json" {
		t.Fatal("normalization changed the declared project identity")
	}
	if !lintProjectSnapshotsEqual(upper.commandLine, lower.commandLine, fsys) {
		t.Fatal("Session drive normalization made unchanged roots or path options incompatible")
	}
	program, err := createStandaloneLintProgram(lower, fsys)
	if err != nil {
		t.Fatal(err)
	}
	if !lintSessionProgramMatchesConfig(program, upper, nil, fsys) {
		t.Fatal("equivalent Windows configuration could not reuse its Program")
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

func TestLintSessionProjectRootCacheTracksServiceProgramGeneration(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "javascript-reference"))
	configPath := tspath.ResolvePath(directory, "tsconfig.json")
	fsys := bundled.WrapFS(osvfs.FS())
	metadata, err := parseStandaloneLintProject(configPath, fsys, fsys)
	if err != nil {
		t.Fatal(err)
	}
	first, err := createStandaloneLintProgram(metadata, fsys)
	if err != nil {
		t.Fatal(err)
	}
	cache := newLintSessionProjectRootCache()
	if cache.canUseServiceProgram(configPath, first, fsys) {
		t.Fatal("JS reference rejected by configured construction was considered compatible")
	}
	if err := os.WriteFile(tspath.ResolvePath(directory, "target.ts"), []byte("export const result = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := createStandaloneLintProgram(metadata, fsys)
	if err != nil {
		t.Fatal(err)
	}
	if first.CommandLine() != second.CommandLine() || !cache.canUseServiceProgram(configPath, second, fsys) {
		t.Fatal("a source-only Program generation did not refresh construction compatibility")
	}
	cache.metadata(configPath, second.CommandLine(), fsys)
	entry := cache.entries[string(lintProgramLexicalPathID(configPath, fsys))]
	if len(cache.entries) != 1 || entry.program.Value() != second || !entry.serviceCompatible {
		t.Fatal("config metadata refresh discarded or accumulated Program compatibility generations")
	}
	if !cache.Invalidate() || len(cache.entries) != 0 {
		t.Fatal("project invalidation retained a Program compatibility result")
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
		withSession  bool
	}{
		{name: "nearest complete project", fixture: "nested", target: "pkg/src/target.ts", wantConfig: "pkg/tsconfig.json", wantRoots: 3, unreadConfig: "tsconfig.json"},
		{name: "external lint config", fixture: "nested", target: "pkg/src/target.ts", configDir: "tooling", wantConfig: "pkg/tsconfig.json", wantRoots: 3, unreadConfig: "tsconfig.json"},
		{name: "ancestor after nearest exclusion", fixture: "ancestor", target: "pkg/target.ts", wantConfig: "tsconfig.json", wantRoots: 1},
		{name: "root boundary gap", fixture: "ancestor", target: "pkg/target.ts", rootDir: "pkg", wantRoots: 1},
		{name: "disabled solution search gap", fixture: "disabled-search", target: "pkg/target.ts", wantRoots: 1},
		{name: "custom reference", fixture: "solution", target: "pkg/src/target.ts", wantConfig: "pkg/tsconfig.app.json", wantRoots: 1},
		{name: "unowned target gap", fixture: "unowned", target: "target.ts", wantRoots: 1},
		// rslint intentionally requires config roots, even when an upstream
		// service Program would contain these targets through source references.
		{name: "imported target gap", fixture: "imported-gap", target: "target.ts", wantRoots: 1},
		{name: "triple slash target gap", fixture: "triple-slash-gap", target: "target.js", wantRoots: 1},
		{name: "overlapping reference chooses child", fixture: "overlapping-reference", target: "target.ts", wantConfig: "leaf.json", wantRoots: 1},
		{name: "disabled source redirect remains an error", fixture: "disabled-source-redirect", target: "target.ts", wantError: "configured project root"},
		{name: "Session disabled source redirect remains an error", fixture: "disabled-source-redirect", target: "target.ts", wantError: "configured project root", withSession: true},
		{name: "explicit JS root without allowJs", fixture: "explicit-js", target: "target.js", wantConfig: "tsconfig.json", wantRoots: 1, withSession: true},
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
			if test.withSession {
				server.backgroundCtx = context.Background()
				server.defaultLibraryPath = bundled.LibPath()
				server.initializeParams = &lsproto.InitializeParams{}
				if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
					t.Fatal(err)
				}
				defer server.session.Close()
				language := lsproto.LanguageKindTypeScript
				if strings.HasSuffix(fileName, ".js") {
					language = lsproto.LanguageKindJavaScript
				}
				server.session.DidOpenFile(context.Background(), uri, 1, editorText, language)
			}
			options := &config.ParserOptions{ProjectService: config.BoolPtr(true), Project: test.project}
			if test.rootDir != "" {
				rootDir := tspath.ResolvePath(directory, test.rootDir)
				options.TsconfigRootDir = &rootDir
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
				if got := program.Options().ConfigFilePath; lintProgramLexicalPathID(got, fsys) != lintProgramLexicalPathID(wantConfig, fsys) {
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
				if targets := generation.Native.TargetsByProgram; len(targets) != 1 || len(targets[0]) != 1 ||
					config.ExactPathID(targets[0][0]) != config.ExactPathID(fileName) {
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

func TestLSPWindowsAbsoluteRootsKeepGenerationContext(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	globalBytes, err := archive.ReadFile("javascript-reference/globals.js")
	if err != nil {
		t.Fatal(err)
	}
	globalText := string(globalBytes)
	for _, test := range []struct {
		name        string
		service     bool
		resident    bool
		jsReference bool
	}{
		{name: "ordinary Session"},
		{name: "service Session", service: true},
		{name: "service fresh fallback", service: true, jsReference: true},
		{name: "service resident fallback", service: true, resident: true, jsReference: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			const configPath = "C:/Repo/tsconfig.json"
			const sourcePath = "C:/Repo/new.ts"
			const uri = lsproto.DocumentUri("file:///C:/Repo/new.ts")
			roots := []string{sourcePath}
			content := "export function value() { var local = 1; return local; }\n"
			if test.jsReference {
				roots = append(roots, "C:/Repo/globals.js")
				content = "/// <reference path=\"./globals.js\" />\nexport function value() { var local = globalValue; return local; }\n"
			}
			configText, err := json.Marshal(map[string]any{
				"compilerOptions": map[string]any{"noLib": true, "allowJs": false},
				"files":           roots,
			})
			if err != nil {
				t.Fatal(err)
			}
			// Both drive spellings address the same disk files. The target is
			// absent from disk, so Realpath preserves the spelling it receives.
			files := make(map[string]string)
			for _, drive := range []string{"C:", "c:"} {
				files[drive+"/Repo/tsconfig.json"] = string(configText)
				files[drive+"/Repo/globals.js"] = globalText
			}
			server := newTestServer()
			server.cwd = "C:/Repo"
			server.fs = &exactCaseLSPProgramFS{FS: utils.NewOverlayVFS(&mockFS{}, files), files: files}
			server.backgroundCtx = context.Background()
			server.initializeParams = &lsproto.InitializeParams{}
			if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
				t.Fatal(err)
			}
			defer server.session.Close()
			if test.resident {
				server.lintPrograms = newLintProgramStore(server)
				server.lintPrograms.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error {
					return nil
				}
			}
			if server.fs.FileExists(sourcePath) {
				t.Fatal("fixture target must exist only in the editor overlay")
			}
			server.documents[uri] = content
			server.session.DidOpenFile(context.Background(), uri, 1, content, lsproto.LanguageKindTypeScript)
			options := &config.ParserOptions{}
			if test.service {
				options.ProjectService = config.BoolPtr(true)
			} else {
				options.Project = config.ProjectPaths{"./tsconfig.json"}
			}
			entries := config.RslintConfig{{
				LanguageOptions: &config.LanguageOptions{ParserOptions: options},
				Rules:           config.Rules{"no-var": "error"},
			}}
			snapshot := documentLintSnapshotForTest(server, uri, entries, server.cwd, false, nil)
			languageService, err := server.session.GetLanguageService(context.Background(), uri)
			if err != nil {
				t.Fatal(err)
			}
			sessionProgram := languageService.GetProgram()
			sessionSource := sessionProgram.GetSourceFile(snapshot.target.Path)
			if sessionSource == nil || config.ExactPathID(sessionSource.FileName()) != config.ExactPathID(sourcePath) || snapshot.target.Path != "c:/Repo/new.ts" {
				t.Fatalf("fixture lost the source across drive spellings: source=%v target=%q", sessionSource, snapshot.target.Path)
			}
			if compatible := lintSessionProgramSupportsService(sessionProgram); compatible == test.jsReference {
				t.Fatalf("Session service compatibility=%v, want %v", compatible, !test.jsReference)
			}
			for _, speculative := range []bool{false, true} {
				var generation linter.Generation
				var release linter.ReleaseFunc
				if speculative {
					generation, release, err = acquireSpeculativeGeneration(context.Background(), content, snapshot,
						server.freezeSpeculativeLintEnvironment(uri, snapshot.target))
				} else {
					provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
					generation, release, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
				}
				if release != nil {
					defer release()
				}
				if err != nil || len(generation.Native.Programs) != 1 {
					t.Fatalf("speculative=%v: Programs=%d error=%v", speculative, len(generation.Native.Programs), err)
				}
				program := generation.Native.Programs[0]
				if got := program.Options().ConfigFilePath; config.ExactPathID(got) != config.ExactPathID(configPath) {
					t.Fatalf("speculative=%v: selected config=%q, want %q", speculative, got, configPath)
				}
				source := program.GetSourceFile(snapshot.target.Path)
				if source == nil || source.Text() != content {
					t.Fatalf("speculative=%v: selected Program lost editor content", speculative)
				}
				if test.jsReference {
					if reference := program.GetSourceFile("C:/Repo/globals.js"); reference == nil || reference.Text() != globalText {
						t.Fatalf("speculative=%v: fallback lost the selected Program's JavaScript context", speculative)
					}
				}
				if !speculative && !test.jsReference && source != sessionSource {
					t.Fatal("normal diagnostics unnecessarily rebuilt the compatible Session Program")
				}
				result, err := runLSPGenerationForTest(context.Background(), generation, nil,
					linter.ArtifactDemand{Native: rule.EditDemandAutofix})
				if err != nil {
					t.Fatal(err)
				}
				diagnostics := result.Observation.Native.Diagnostics
				if len(diagnostics) != 1 || diagnostics[0].RuleName != "no-var" || len(diagnostics[0].Fixes()) != 1 {
					t.Fatalf("speculative=%v: lost native diagnostic or fix: %+v", speculative, diagnostics)
				}
				if diagnostics[0].Fixes()[0].Text != "let" {
					t.Fatalf("speculative=%v: unexpected native fix: %+v", speculative, diagnostics[0].Fixes())
				}
				if diagnostics[0].SourceFile.Text() != content || server.documents[uri] != content {
					t.Fatalf("speculative=%v: diagnostic or resident document used the wrong content", speculative)
				}
			}
		})
	}
}

func TestSelectConfiguredLintProjectServiceDoesNotProbePrograms(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		fixture, file, wantConfig string
	}{
		{fixture: "imported-gap", file: "target.ts"},
		{fixture: "triple-slash-gap", file: "target.js"},
		{fixture: "solution", file: "pkg/src/target.ts", wantConfig: "pkg/tsconfig.app.json"},
		{fixture: "overlapping-reference", file: "target.ts", wantConfig: "leaf.json"},
	} {
		t.Run(test.fixture, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, test.fixture))
			fsys := bundled.WrapFS(osvfs.FS())
			target := lspConfigTarget(tspath.ResolvePath(directory, test.file), directory, fsys)
			request := newStandaloneLintProjectRequestWithFS(target, fsys)
			request.sourceReferences = true
			loaders := request.loaders()
			var built []string
			loaders.program = func(metadata *lintProjectMetadata) (*compiler.Program, *ast.SourceFile, error) {
				configPath := metadata.configPath
				built = append(built, configPath)
				return request.program(metadata)
			}
			selected, found, err := selectConfiguredLintProject(nil, directory, target, fsys, loaders)
			if err != nil {
				t.Fatal(err)
			}
			if test.wantConfig == "" {
				if found || selected.program != nil || len(built) != 0 {
					t.Fatalf("gap constructed a project: selected=%s Programs=%v", selected.configPath, built)
				}
				return
			}
			wantConfig := tspath.ResolvePath(directory, test.wantConfig)
			if !found || lintProgramLexicalPathID(selected.configPath, fsys) != lintProgramLexicalPathID(wantConfig, fsys) ||
				len(built) != 1 || lintProgramLexicalPathID(built[0], fsys) != lintProgramLexicalPathID(wantConfig, fsys) {
				t.Fatalf("selection=%s Programs=%v, want only %s", selected.configPath, built, wantConfig)
			}
		})
	}
}

func TestProjectServiceLSPDoesNotUseIndirectSessionMembership(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	directory := tspath.NormalizePath(archive.Materialize(t, "imported-gap"))
	fileName := tspath.ResolvePath(directory, "target.ts")
	uri := documentURIFromPath(fileName)
	server := newTestServer()
	server.cwd = directory
	server.fs = bundled.WrapFS(osvfs.FS())
	server.backgroundCtx = context.Background()
	server.defaultLibraryPath = bundled.LibPath()
	server.initializeParams = &lsproto.InitializeParams{}
	if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
		t.Fatal(err)
	}
	defer server.session.Close()
	const content = "export const value = 2;\ndebugger;\n"
	server.documents[uri] = content
	entries := config.RslintConfig{{
		LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}},
		Rules:           config.Rules{"no-debugger": "error"},
	}}
	snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
	request := newStandaloneLintProjectRequestWithFS(snapshot.target, server.fs)
	provider := &documentGenerationProvider{
		server: server, uri: uri, snapshot: snapshot,
		requestPrograms: func(context.Context, lsproto.DocumentUri, target.File) (lintProjectLoaders, linter.ReleaseFunc) {
			return lintProjectLoaders{
				program: func(*lintProjectMetadata) (*compiler.Program, *ast.SourceFile, error) {
					t.Fatal("gap discovery attempted standalone Program construction")
					return nil, nil, nil
				},
				metadata: request.loadMetadata,
			}, nil
		},
	}
	assertGap := func() {
		t.Helper()
		generation, release, err := provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
		if release != nil {
			release()
		}
		if err != nil || len(generation.Native.Programs) != 1 {
			t.Fatalf("gap generation Programs=%d error=%v", len(generation.Native.Programs), err)
		}
		gap := generation.Native.Programs[0]
		if gap.Options().ConfigFilePath != "" || !gap.Options().NoResolve.IsTrue() || gap.GetSourceFile(fileName).Text() != content {
			t.Fatal("indirect Session membership replaced the frozen gap source")
		}
	}
	// The lint selection phase must not load Session projects just to decide
	// ownership. DidOpen below is a separate, existing TypeScript lifecycle.
	if projects := server.session.Snapshot().ProjectCollection.ConfiguredProjects(); len(projects) != 0 {
		t.Fatal("fixture Session was not cold")
	}
	assertGap()
	if projects := server.session.Snapshot().ProjectCollection.ConfiguredProjects(); len(projects) != 0 {
		t.Fatal("gap discovery loaded a Session project")
	}
	server.session.DidOpenFile(context.Background(), uri, 1, content, "typescript")
	ls, err := server.session.GetLanguageService(context.Background(), uri)
	if err != nil {
		t.Fatal(err)
	}
	importingProgram := ls.GetProgram()
	configPath := tspath.ResolvePath(directory, "tsconfig.json")
	caseSensitive := server.fs.UseCaseSensitiveFileNames()
	if got := importingProgram.Options().ConfigFilePath; tspath.ToPath(got, "", caseSensitive) != tspath.ToPath(configPath, "", caseSensitive) {
		t.Fatalf("Session selected config %q, want %q", got, configPath)
	}
	if importingProgram.GetSourceFile(fileName) == nil {
		t.Fatalf("Session Program does not contain imported source %q", fileName)
	}
	assertGap()
}

func TestProjectServiceLSPFrozenRootDirectory(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		name         string
		owner        string
		explicitRoot string
		resetRoot    bool
		wantTyped    bool
	}{
		{name: "parent owner admits ancestor", owner: ".", wantTyped: true},
		{name: "explicit root narrows parent owner", owner: ".", explicitRoot: "pkg"},
		{name: "null restores parent owner", owner: ".", explicitRoot: "pkg", resetRoot: true, wantTyped: true},
		{name: "nested owner excludes ancestor", owner: "pkg"},
		{name: "explicit root expands nested owner", owner: "pkg", explicitRoot: ".", wantTyped: true},
		{name: "null restores nested owner", owner: "pkg", explicitRoot: ".", resetRoot: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, "ancestor"))
			workspace := tspath.ResolvePath(directory, "pkg")
			fileName := tspath.ResolvePath(workspace, "target.ts")
			owner := tspath.ResolvePath(directory, test.owner)
			for _, cwdName := range []string{".", "pkg"} {
				t.Run("cwd="+cwdName, func(t *testing.T) {
					server := newTestServer()
					server.cwd = tspath.ResolvePath(directory, cwdName)
					server.fs = bundled.WrapFS(osvfs.FS())
					server.lintPrograms = newLintProgramStore(server)
					server.lintPrograms.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
					uri := documentURIFromPath(fileName)
					const editorText = "debugger;\ndeclare const value: any;\nexport const result = value.member;\n"
					server.documents[uri] = editorText
					options := &config.ParserOptions{ProjectService: config.BoolPtr(true)}
					if test.explicitRoot != "" {
						rootDir := tspath.ResolvePath(directory, test.explicitRoot)
						options.TsconfigRootDir = &rootDir
					}
					entries := config.RslintConfig{{
						Plugins:         []string{"@typescript-eslint"},
						LanguageOptions: &config.LanguageOptions{ParserOptions: options},
						Rules: config.Rules{
							"no-debugger": "error", "@typescript-eslint/no-unsafe-member-access": "error",
						},
					}}
					if test.resetRoot {
						var reset config.RslintConfig
						if err := json.Unmarshal([]byte(`[{"languageOptions":{"parserOptions":{"tsconfigRootDir":null}}}]`), &reset); err != nil {
							t.Fatal(err)
						}
						entries = append(entries, reset...)
					}
					installJSConfigsForTest(server, map[string]config.RslintConfig{owner: entries})
					snapshot := server.documentLintSnapshot(uri)
					if snapshot.target.ConfigDirectory != owner || snapshot.projectPolicyError != nil {
						t.Fatalf("snapshot owner=%q error=%v, want %q", snapshot.target.ConfigDirectory, snapshot.projectPolicyError, owner)
					}
					environment := server.freezeSpeculativeLintEnvironment(uri, snapshot.target)

					// The same owner is independent of workspace cwd. A later owner
					// refresh must also leave this operation's target/policy frozen.
					changedOwner := directory
					if owner == directory {
						changedOwner = workspace
					}
					server.cwd = changedOwner
					installJSConfigsForTest(server, map[string]config.RslintConfig{changedOwner: entries})
					server.invalidateLintProjectCaches()
					if next := server.documentLintSnapshot(uri); next.target.ConfigDirectory != changedOwner {
						t.Fatalf("new snapshot owner=%q, want changed owner %q", next.target.ConfigDirectory, changedOwner)
					}
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
						if lintProgramLexicalPathID(program.Options().ConfigFilePath, server.fs) != lintProgramLexicalPathID(wantConfig, server.fs) {
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

func TestLSPDisabledProjectKeepsDiagnosticAndFixContext(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		name       string
		project    string
		wantCycles int
		cold       bool
	}{
		{name: "false", project: "false"},
		{name: "null", project: "null"},
		{name: "cold false", project: "false", cold: true},
		{name: "cold null", project: "null", cold: true},
		{name: "empty projects", project: `[]`},
		{name: "unmatched project", project: `["tsconfig.other.json"]`},
		{name: "configured control", project: `["tsconfig.json"]`, wantCycles: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, "disabled-binding"))
			fileName := tspath.ResolvePath(directory, "target.ts")
			server := newTestServer()
			server.cwd = directory
			server.fs = bundled.WrapFS(osvfs.FS())
			server.backgroundCtx = context.Background()
			server.defaultLibraryPath = bundled.LibPath()
			server.initializeParams = &lsproto.InitializeParams{}
			if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
				t.Fatal(err)
			}
			defer server.session.Close()
			content, ok := server.fs.ReadFile(fileName)
			if !ok {
				t.Fatal("missing target fixture")
			}
			uri := documentURIFromPath(fileName)
			server.documents[uri] = content
			if !test.cold {
				server.session.DidOpenFile(context.Background(), uri, 1, content, lsproto.LanguageKindTypeScript)
			}
			var entries config.RslintConfig
			if err := json.Unmarshal([]byte(`[{"plugins":["import"],"languageOptions":{"parserOptions":{"project":`+test.project+`}},"rules":{"import/no-cycle":"error","no-var":"error"}}]`), &entries); err != nil {
				t.Fatal(err)
			}
			snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
			for _, speculative := range []bool{false, true} {
				var result linter.PipelineResult
				var err error
				if speculative {
					result, err = speculativePipelineResultForTest(server, context.Background(), uri, content, snapshot)
				} else {
					result, err = configuredDocumentPipelineResultForTest(server, context.Background(), uri, entries, directory, false, nil)
				}
				if err != nil {
					t.Fatal(err)
				}
				cycles, vars := 0, 0
				for _, diagnostic := range result.Observation.Native.Diagnostics {
					switch diagnostic.RuleName {
					case "import/no-cycle":
						cycles++
					case "no-var":
						vars++
					default:
						t.Errorf("unexpected diagnostic: %+v", diagnostic)
					}
				}
				if cycles != test.wantCycles || vars != 1 {
					t.Errorf("speculative=%v: cycles=%d no-var=%d, want %d/1", speculative, cycles, vars, test.wantCycles)
				}
			}
			wantFixed := strings.Replace(content, "var local", "let local", 1)
			if got := runSpeculativeFixAllForTest(t, server, context.Background(), uri, content, snapshot); got != wantFixed {
				t.Errorf("fix-all=%q, want %q", got, wantFixed)
			}
			if server.documents[uri] != content {
				t.Fatal("fix-all changed editor text")
			}
			if test.cold && len(server.session.Snapshot().ProjectCollection.ConfiguredProjects()) != 0 {
				t.Fatal("disabled binding requested a Session project")
			}
		})
	}
}

func TestProjectServiceLSPSessionConfigSnapshot(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		name, fixture, directory, editFile, replacement string
		extends                                         bool
	}{
		{name: "other roots change", fixture: "flat-policy", directory: "pkg[1]", editFile: "tsconfig.json",
			replacement: `{"compilerOptions":{"noLib":true},"files":["target.ts","safe.d.ts"]}`},
		{name: "extended config changes", fixture: "flat-policy", directory: "pkg[1]", editFile: "base.json", extends: true,
			replacement: `{"compilerOptions":{"noLib":true},"files":["target.ts","safe.d.ts"]}`},
		{name: "paths change", fixture: "snapshot-options", editFile: "tsconfig.json",
			replacement: `{"compilerOptions":{"noLib":true,"module":"esnext","moduleResolution":"bundler","paths":{"payload":["./safe.d.ts"]}},"files":["target.ts"]}`},
		{name: "second referenced config changes", fixture: "snapshot-references", directory: "app", editFile: "../lib/tsconfig.json",
			replacement: `{"compilerOptions":{"noLib":true,"module":"esnext","moduleResolution":"bundler","composite":true,"outDir":"dist","paths":{"payload":["./safe.d.ts"]}},"files":["value.ts"]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := tspath.ResolvePath(tspath.NormalizePath(archive.Materialize(t, test.fixture)), test.directory)
			fileName := tspath.ResolvePath(directory, "target.ts")
			configPath := tspath.ResolvePath(directory, "tsconfig.json")
			if test.extends {
				initial, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tspath.ResolvePath(directory, "base.json"), initial, 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(configPath, []byte(`{"extends":"./base.json"}`), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			server := newTestServer()
			server.cwd = directory
			server.fs = bundled.WrapFS(osvfs.FS())
			server.backgroundCtx = context.Background()
			server.defaultLibraryPath = bundled.LibPath()
			// No dynamic watched-file registration: Session does not hear the
			// disk config edit, but lint discovery observes the new snapshot.
			server.initializeParams = &lsproto.InitializeParams{}
			if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
				t.Fatal(err)
			}
			defer server.session.Close()
			content, ok := server.fs.ReadFile(fileName)
			if !ok {
				t.Fatal("missing target fixture")
			}
			uri := documentURIFromPath(fileName)
			server.documents[uri] = content
			server.session.DidOpenFile(context.Background(), uri, 1, content, lsproto.LanguageKindTypeScript)
			entries := config.RslintConfig{{
				Plugins:         []string{"@typescript-eslint"},
				LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}},
				Rules:           config.Rules{"@typescript-eslint/no-unsafe-member-access": "error"},
			}}
			for _, phase := range []struct {
				name       string
				wantUnsafe int
			}{
				{name: "initial", wantUnsafe: 1},
				{name: "changed"},
			} {
				t.Run(phase.name, func(t *testing.T) {
					if phase.name == "changed" {
						if err := os.WriteFile(tspath.ResolvePath(directory, test.editFile), []byte(test.replacement), 0o644); err != nil {
							t.Fatal(err)
						}
						content += "// edited after disk config changed\n"
						server.documents[uri] = content
						server.session.DidChangeFile(context.Background(), uri, 2, makeDidChangeParams(uri, 2, content).ContentChanges)
					}
					snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
					for _, speculative := range []bool{false, false, true} {
						var result linter.PipelineResult
						var err error
						if speculative {
							result, err = speculativePipelineResultForTest(server, context.Background(), uri, content, snapshot)
						} else {
							result, err = configuredDocumentPipelineResultForTest(server, context.Background(), uri, entries, directory, false, nil)
							if phase.name == "initial" && err == nil {
								languageService, sessionErr := server.session.GetLanguageService(context.Background(), uri)
								if sessionErr != nil {
									t.Fatal(sessionErr)
								}
								if got := result.Observation.Native.Diagnostics; len(got) == 1 && got[0].SourceFile != languageService.GetProgram().GetSourceFile(fileName) {
									t.Error("unchanged configuration rebuilt a Session-owned Program")
								}
							}
						}
						if err != nil {
							t.Fatal(err)
						}
						if got := len(result.Observation.Native.Diagnostics); got != phase.wantUnsafe {
							t.Errorf("speculative=%v: unsafe diagnostics=%d, want %d: %+v", speculative, got, phase.wantUnsafe, result.Observation.Native.Diagnostics)
						}
					}
				})
			}
		})
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
				if lintProgramLexicalPathID(program.Options().ConfigFilePath, server.fs) != lintProgramLexicalPathID(wantConfig, server.fs) {
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
	for _, test := range []struct {
		fixture, target, reference, targetContent, referenceContent string
		wantMismatch                                                bool
	}{
		{
			fixture: "reference-sources", target: "app/src/main.ts", reference: "lib/src/value.ts",
			targetContent:    "import { value } from '../../lib/src/value';\nexport const result: number = value;\n",
			referenceContent: "export const value = 42;\n",
		},
		{
			fixture: "javascript-reference", target: "target.ts", reference: "globals.js",
			targetContent:    "/// <reference path=\"./globals.js\" />\nexport const result: number = globalValue;\n",
			referenceContent: "var globalValue = 'editor';\n", wantMismatch: true,
		},
	} {
		t.Run(test.fixture, func(t *testing.T) {
			directory := tspath.NormalizePath(archive.Materialize(t, test.fixture))
			targetPath := tspath.ResolvePath(directory, test.target)
			referencePath := tspath.ResolvePath(directory, test.reference)
			server := newTestServer()
			server.cwd = directory
			server.fs = bundled.WrapFS(osvfs.FS())
			server.lintSessionRoots = newLintSessionProjectRootCache()
			server.backgroundCtx = context.Background()
			server.defaultLibraryPath = bundled.LibPath()
			server.initializeParams = &lsproto.InitializeParams{}
			if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
				t.Fatal(err)
			}
			defer server.session.Close()
			uri := documentURIFromPath(targetPath)
			referenceURI := documentURIFromPath(referencePath)
			server.documents[uri] = test.targetContent
			server.documents[referenceURI] = test.referenceContent
			language := lsproto.LanguageKindTypeScript
			if strings.HasSuffix(referencePath, ".js") {
				language = lsproto.LanguageKindJavaScript
			}
			server.session.DidOpenFile(context.Background(), referenceURI, 1, test.referenceContent, language)
			server.session.DidOpenFile(context.Background(), uri, 1, test.targetContent, "typescript")
			if test.wantMismatch {
				ls, err := server.session.GetLanguageService(context.Background(), uri)
				if err != nil || ls.GetProgram().GetSourceFile(targetPath) == nil || ls.GetProgram().GetSourceFile(referencePath) != nil {
					t.Fatalf("fixture did not expose a JS reference rejected by a target-containing Session Program: %v", err)
				}
			}
			entries := config.RslintConfig{{LanguageOptions: &config.LanguageOptions{
				ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)},
			}}}
			snapshot := documentLintSnapshotForTest(server, uri, entries, directory, false, nil)
			for _, speculative := range []bool{false, true} {
				var generation linter.Generation
				var release linter.ReleaseFunc
				var err error
				if speculative {
					generation, release, err = acquireSpeculativeGeneration(context.Background(), test.targetContent, snapshot,
						server.freezeSpeculativeLintEnvironment(uri, snapshot.target))
				} else {
					provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
					generation, release, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
				}
				if release != nil {
					defer release()
				}
				if err != nil || len(generation.Native.Programs) != 1 {
					t.Fatalf("speculative=%v: Programs=%d error=%v", speculative, len(generation.Native.Programs), err)
				}
				program := generation.Native.Programs[0]
				if source := program.GetSourceFile(referencePath); source == nil || source.Text() != test.referenceContent {
					t.Fatalf("speculative=%v: did not load unsaved referenced source", speculative)
				}
				if test.wantMismatch {
					foundMismatch := false
					for _, diagnostic := range program.NoEmitDiagnostics(context.Background()) {
						foundMismatch = foundMismatch || diagnostic.Code() == 2322
						if diagnostic.Code() == 2304 {
							t.Fatalf("speculative=%v: JS global was missing from the type context", speculative)
						}
					}
					if !foundMismatch {
						t.Fatalf("speculative=%v: lost the JS global's string-to-number type mismatch", speculative)
					}
				} else if program.GetSourceFile(tspath.ResolvePath(directory, "lib/dist/value.d.ts")) != nil {
					t.Fatalf("speculative=%v: used stale declaration output", speculative)
				}
			}
		})
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
	uri := documentURIFromPath(tspath.ResolvePath(directory, "pkg/src/target.ts"))
	snapshot := resolveDocumentLintSnapshotConfig(documentLintSnapshotForTest(server, uri, entries, directory, false, nil), server.fs)
	if snapshot.projectPolicyError != nil || snapshot.projectPolicy.ServiceRootDirectory == "" || len(snapshot.typeScriptConfigPaths) != 0 {
		t.Fatalf("TypeScript policy=%+v paths=%v error=%v", snapshot.projectPolicy, snapshot.typeScriptConfigPaths, snapshot.projectPolicyError)
	}
	jsURI := documentURIFromPath(tspath.ResolvePath(directory, "target.js"))
	jsSnapshot := resolveDocumentLintSnapshotConfig(documentLintSnapshotForTest(server, jsURI, entries, directory, false, nil), server.fs)
	if jsSnapshot.projectPolicyError == nil || !strings.Contains(jsSnapshot.projectPolicyError.Error(), "missing.json") {
		t.Fatalf("matching explicit project error=%v", jsSnapshot.projectPolicyError)
	}
}

func TestDocumentProjectPolicyUsesFlatConfigAndRawBases(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		name         string
		options      string
		matched      bool
		restore      bool
		firstProject string
		omitSecond   bool
		wantConfig   string
		gap          bool
		wantError    string
		rootOverride bool
	}{
		{name: "ordinary declarations keep original order", options: `{}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched service true", options: `{"projectService":true}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched service false", options: `{"projectService":false}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched service null", options: `{"projectService":null}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched root", options: `{"tsconfigRootDir":"."}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched root null", options: `{"tsconfigRootDir":null}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched project false", options: `{"project":false}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched project null", options: `{"project":null}`, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched project", options: `{"project":"./tsconfig.safe.json"}`, omitSecond: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched project and service", options: `{"projectService":false,"project":"./tsconfig.safe.json"}`, omitSecond: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "unmatched project still validates declaration", options: `{"project":"./missing.json"}`, wantError: "doesn't exist"},
		{name: "matched service false retains declarations", options: `{"projectService":false}`, matched: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "matched root retains declaration order", options: `{}`, matched: true, rootOverride: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "matched false", options: `{"project":false}`, matched: true, gap: true},
		{name: "matched null", options: `{"project":null}`, matched: true, gap: true},
		{name: "matched empty array keeps earlier declarations", options: `{"project":[]}`, matched: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "matched false then restore", options: `{"project":false}`, matched: true, restore: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "matched null then restore", options: `{"project":null}`, matched: true, restore: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "literal project in literal directory", options: `{}`, omitSecond: true, wantConfig: "tsconfig.unsafe.json"},
		{name: "glob project in literal directory", options: `{}`, firstProject: "./tsconfig.u*.json", omitSecond: true, wantConfig: "tsconfig.unsafe.json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := tspath.ResolvePath(archive.Materialize(t, "flat-policy"), "pkg[1]")
			fileName := tspath.ResolvePath(directory, "target.ts")
			server := newTestServer()
			server.cwd = directory
			server.fs = bundled.WrapFS(osvfs.FS())
			server.backgroundCtx = context.Background()
			server.defaultLibraryPath = bundled.LibPath()
			server.initializeParams = &lsproto.InitializeParams{}
			if err := server.handleInitialized(context.Background(), &lsproto.InitializedParams{}); err != nil {
				t.Fatal(err)
			}
			defer server.session.Close()
			server.lintPrograms = newLintProgramStore(server)
			server.lintPrograms.coverage.watchFiles = func(context.Context, project.WatcherID, []*lsproto.FileSystemWatcher) error { return nil }
			uri := documentURIFromPath(fileName)
			const content = "export const result = (() => { var output = value.member; return output; })();\n"
			server.documents[uri] = content
			server.session.DidOpenFile(context.Background(), uri, 1, content, "typescript")
			first := test.firstProject
			if first == "" {
				first = "./tsconfig.unsafe.json"
			}
			entries := config.RslintConfig{{
				Plugins: []string{"@typescript-eslint"}, Rules: config.Rules{"no-var": "error", "@typescript-eslint/no-unsafe-member-access": "error"},
				LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{first}}},
			}}
			if !test.omitSecond {
				entries = append(entries, config.ConfigEntry{LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{"./tsconfig.safe.json"}}}})
			}
			var options config.ParserOptions
			if err := json.Unmarshal([]byte(test.options), &options); err != nil {
				t.Fatal(err)
			}
			if test.rootOverride {
				options.TsconfigRootDir = &directory
			}
			selector := "unused.ts"
			if test.matched {
				selector = "target.ts"
			}
			entries = append(entries, config.ConfigEntry{Files: []string{selector}, LanguageOptions: &config.LanguageOptions{ParserOptions: &options}})
			if test.restore {
				entries = append(entries, config.ConfigEntry{LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{"./tsconfig.safe.json"}}}})
			}
			installJSConfigsForTest(server, map[string]config.RslintConfig{directory: entries})
			snapshot := server.documentLintSnapshot(uri)
			if test.wantError != "" {
				if snapshot.projectPolicyError == nil || !strings.Contains(snapshot.projectPolicyError.Error(), test.wantError) {
					t.Fatalf("expected raw declaration error %q, got %v", test.wantError, snapshot.projectPolicyError)
				}
				return
			}
			if snapshot.projectPolicyError != nil || (len(snapshot.typeScriptConfigPaths) == 0) != test.gap {
				t.Fatalf("target paths=%v error=%v, gap=%v", snapshot.typeScriptConfigPaths, snapshot.projectPolicyError, test.gap)
			}
			for _, speculative := range []bool{false, true} {
				var generation linter.Generation
				var release linter.ReleaseFunc
				var err error
				if speculative {
					generation, release, err = acquireSpeculativeGeneration(context.Background(), content, snapshot,
						server.freezeSpeculativeLintEnvironment(uri, snapshot.target))
				} else {
					provider := &documentGenerationProvider{server: server, uri: uri, snapshot: snapshot}
					generation, release, err = provider.AcquireGeneration(context.Background(), linter.SourceSnapshot{})
				}
				if err != nil || len(generation.Native.Programs) != 1 {
					t.Fatalf("speculative=%v: Programs=%d error=%v", speculative, len(generation.Native.Programs), err)
				}
				program := generation.Native.Programs[0]
				if !test.gap && lintProgramLexicalPathID(program.Options().ConfigFilePath, server.fs) != lintProgramLexicalPathID(tspath.ResolvePath(directory, test.wantConfig), server.fs) {
					t.Fatalf("speculative=%v: selected %q, want %s", speculative, program.Options().ConfigFilePath, test.wantConfig)
				}
				result, err := runLSPGenerationForTest(context.Background(), generation, release, linter.ArtifactDemand{})
				if err != nil {
					t.Fatal(err)
				}
				counts := make(map[string]int)
				for _, diagnostic := range result.Observation.Native.Diagnostics {
					counts[diagnostic.RuleName]++
				}
				wantUnsafe := 0
				if test.wantConfig == "tsconfig.unsafe.json" {
					wantUnsafe = 1
				}
				if counts["@typescript-eslint/no-unsafe-member-access"] != wantUnsafe || counts["no-var"] != 1 {
					t.Fatalf("speculative=%v: diagnostics=%v, want unsafe=%d and no-var=1", speculative, counts, wantUnsafe)
				}
			}
		})
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
		if lintProgramLexicalPathID(program.Options().ConfigFilePath, server.fs) != lintProgramLexicalPathID(wantConfig, server.fs) {
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
