package loader

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	default_rule "github.com/web-infra-dev/rslint/internal/plugins/import/rules/default"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
)

func rootProgramTestPlan(dir string, names ...string) target.Plan {
	plan := target.Plan{Files: make([]target.File, 0, len(names))}
	for _, name := range names {
		fileName := tspath.NormalizePath(filepath.Join(dir, name))
		plan.Files = append(plan.Files, target.File{PathIdentity: rslintconfig.PathIdentity{Path: fileName,
			CanonicalPath: fileName}, ConfigDirectory: dir,
		})
	}
	return plan
}

func rootProgramTestBuildContext() *buildContext {
	return newBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
}

func compareProgramSyntaxDiagnostics(t *testing.T, legacy, direct []rule.RuleDiagnostic) {
	t.Helper()
	if len(legacy) != len(direct) {
		t.Fatalf("syntax diagnostic count differs: legacy=%d direct=%d", len(legacy), len(direct))
	}
	for i := range legacy {
		left, right := legacy[i], direct[i]
		if left.RuleName != right.RuleName || left.FilePath != right.FilePath ||
			left.Range != right.Range || left.Message.Description != right.Message.Description {
			t.Fatalf("syntax diagnostic %d differs:\nlegacy=%+v\ndirect=%+v", i, left, right)
		}
	}
}

func TestRootProgramsMatchCompatibilityProgram(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"package.json":  "{\"type\":\"module\"}",
		"a.js":          "import './b.js';\nimport './unselected.js';\nclass C { method(@dec value) {} }\n",
		"b.js":          "import './a.js';\nexport const value = 1;\n",
		"unselected.js": "export const unselected = 1;\n",
		"bad.ts":        "export const value: = 1;\n",
		"view.tsx":      "export const view = <div>;\n",
	})
	plan := rootProgramTestPlan(dir, "a.js", "b.js", "bad.ts", "view.tsx")
	rootFiles := make([]string, 0, len(plan.Files))
	for _, target := range plan.Files {
		rootFiles = append(rootFiles, target.Path)
	}

	legacyContext := rootProgramTestBuildContext()
	compatibility, err := createCompatibilityProgramForTest(rootFiles, true, dir, legacyContext)
	if err != nil {
		t.Fatalf("createCompatibilityProgramForTest: %v", err)
	}
	legacyProgram, err := lintprogram.NewFromBoundSources(compatibility, compatibility.SourceFiles())
	if err != nil {
		t.Fatalf("adapt compatibility Program: %v", err)
	}
	legacyDiagnostics := collectTargetSyntacticDiagnostics(
		[]*lintprogram.Program{legacyProgram},
		[][]string{rootFiles},
		false,
		false,
	)

	directContext := rootProgramTestBuildContext()
	programs, directDiagnostics, err := buildRootProgramsForTest(
		[][]target.File{plan.Files},
		dir,
		directContext,
		false, // Exercise the production parallel parse/bind path.
	)
	if err != nil {
		t.Fatalf("buildRootProgramsForTest: %v", err)
	}
	compareProgramSyntaxDiagnostics(t, legacyDiagnostics, directDiagnostics)
	if len(programs) != 1 || len(programs[0].SourceFiles()) != len(plan.Files) {
		t.Fatalf("source-only Program source count differs: programs=%d files=%d", len(programs), len(programs[0].SourceFiles()))
	}
	sourceProgram := programs[0]

	for i, target := range plan.Files {
		programFile := exactProgramSourceFile(compatibility, target.Path)
		directFile := sourceProgram.SourceFiles()[i]
		if programFile == nil || directFile == nil {
			t.Fatalf("missing source for %q: Program=%v direct=%v", target.Path, programFile != nil, directFile != nil)
		}
		if programFile.Text() != directFile.Text() || programFile.ScriptKind != directFile.ScriptKind {
			t.Fatalf("source text or ScriptKind differs for %q", target.Path)
		}
		if (programFile.ExternalModuleIndicator == nil) != (directFile.ExternalModuleIndicator == nil) {
			t.Fatalf("external-module detection differs for %q", target.Path)
		}
		if legacyMetadata := compatibility.GetSourceFileMetaData(programFile.Path()); legacyMetadata != sourceProgram.SourceFileMetadata(directFile) {
			t.Fatalf("source metadata differs for %q: legacy=%+v direct=%+v", target.Path, legacyMetadata, sourceProgram.SourceFileMetadata(directFile))
		}

		programAST, programErr := api.EncodeAST(programFile, target.Path)
		directAST, directErr := api.EncodeAST(directFile, target.Path)
		if programErr != nil || directErr != nil {
			t.Fatalf("EncodeAST for %q: Program=%v direct=%v", target.Path, programErr, directErr)
		}
		if !bytes.Equal(programAST, directAST) {
			t.Fatalf("bound AST differs for %q", target.Path)
		}

		for _, specifier := range programFile.Imports() {
			mode := compatibility.GetModeForUsageLocation(programFile, specifier)
			legacyResolution := compatibility.GetResolvedModule(programFile, specifier.Text(), mode)
			directResolution := sourceProgram.GetResolvedModule(directFile, specifier.Text(), mode)
			if legacyResolution == nil || directResolution == nil ||
				legacyResolution.ResolvedFileName != directResolution.ResolvedFileName ||
				legacyResolution.IsResolved() != directResolution.IsResolved() {
				t.Fatalf("module resolution differs for %q in %q: legacy=%+v direct=%+v", specifier.Text(), target.Path, legacyResolution, directResolution)
			}
			if legacyResolution.IsResolved() {
				legacySource := compatibility.GetSourceFileForResolvedModule(legacyResolution.ResolvedFileName)
				directSource := sourceProgram.GetSourceFileForResolvedModule(directResolution.ResolvedFileName)
				if (legacySource == nil) != (directSource == nil) ||
					legacySource != nil && legacySource.FileName() != directSource.FileName() {
					t.Fatalf("resolved source lookup differs for %q", specifier.Text())
				}
			}
		}
	}
	unselected := tspath.ResolvePath(dir, "unselected.js")
	if compatibility.GetSourceFile(unselected) != nil || sourceProgram.GetSourceFile(unselected) != nil {
		t.Fatal("NoResolve Programs materialized an unselected dependency")
	}
}

func TestRootProgramsReportUnreadableTarget(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	targetPath := tspath.ResolvePath(dir, "missing.ts")
	_, _, err := buildRootProgramsForTest(
		[][]target.File{{{PathIdentity: rslintconfig.PathIdentity{Path: targetPath,
			CanonicalPath: targetPath}, ConfigDirectory: dir,
		}}},
		dir,
		rootProgramTestBuildContext(),
		true,
	)
	want := fmt.Sprintf("program: could not read root %q", targetPath)
	if err == nil || err.Error() != want {
		t.Fatalf("unreadable root target error = %v, want %q", err, want)
	}
}

func TestRootProgramSupportsCrossFileImportRules(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts": "import value from './b';\nexport const a = value;\n",
		"b.ts": "import './a';\nexport const b = 1;\n",
	})
	plan := rootProgramTestPlan(dir, "a.ts", "b.ts")
	programs, diagnostics, err := buildRootProgramsForTest(
		[][]target.File{plan.Files},
		dir,
		rootProgramTestBuildContext(),
		true,
	)
	if err != nil {
		t.Fatalf("buildRootProgramsForTest: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected source-only Program syntax diagnostics: %+v", diagnostics)
	}
	if len(programs) != 1 {
		t.Fatalf("source-only Program count = %d, want 1", len(programs))
	}

	var cycleReports, defaultReports int
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs: programs,
		TargetsByProgram: [][]string{{
			plan.Files[0].Path,
			plan.Files[1].Path,
		}},
		SingleThreaded: true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{
				{
					Name:     no_cycle.NoCycleRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						if !ctx.Program().IsValid() || ctx.Program().CanProvideTypeChecker(ctx.SourceFile) {
							t.Fatal("source-only import rule received the wrong Program capabilities")
						}
						return no_cycle.NoCycleRule.Run(ctx, nil)
					},
				},
				{
					Name:     default_rule.DefaultRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return default_rule.DefaultRule.Run(ctx, nil)
					},
				},
			}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	result, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			switch diagnostic.RuleName {
			case no_cycle.NoCycleRule.Name:
				cycleReports++
			case default_rule.DefaultRule.Name:
				defaultReports++
			default:
				t.Errorf("unexpected source-only Program import diagnostic: %+v", diagnostic)
			}
		}},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if result.LintedFileCount != 2 || cycleReports != 2 || defaultReports != 1 {
		t.Fatalf(
			"source-only Program import results differ: files=%d cycles=%d defaults=%d",
			result.LintedFileCount,
			cycleReports,
			defaultReports,
		)
	}
}

type rootProgramBindOutcome struct {
	binding   LoadResult
	err       error
	panicText string
}

func runRootProgramBind(bind func() (LoadResult, error)) (outcome rootProgramBindOutcome) {
	defer func() {
		if value := recover(); value != nil {
			outcome.panicText = fmt.Sprint(value)
		}
	}()
	outcome.binding, outcome.err = bind()
	return outcome
}

func TestLoadCLIMatchesCompatibilityRootAdmission(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		caseSensitive bool
		supported     bool
	}{
		{name: "custom extension", fileName: "target.vue", caseSensitive: true},
		{name: "extensionless", fileName: "target", caseSensitive: true},
		{name: "upper-case extension on case-sensitive FS", fileName: "target.TS", caseSensitive: true},
		{name: "upper-case extension on case-insensitive FS", fileName: "target.TS", caseSensitive: false, supported: true},
		{name: "json enabled by compatibility defaults", fileName: "target.json", caseSensitive: true, supported: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const configDir = "/repo"
			targetPath := tspath.ResolvePath(configDir, test.fileName)
			newContext := func() *buildContext {
				fsys := newBindingIndexTestFS([]string{targetPath}, nil)
				fsys.caseSensitive = test.caseSensitive
				fsys.files[targetPath] = "const value = ;\n"
				return newBuildContext(fsys)
			}
			plan := target.Plan{Files: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: targetPath,
				CanonicalPath: targetPath}, ConfigDirectory: configDir,
			}}}

			legacyContext := newContext()
			legacy := runRootProgramBind(func() (LoadResult, error) {
				return loadAPIForTest(ProjectSet{}, plan, configDir, legacyContext, true)
			})
			directContext := newContext()
			direct := runRootProgramBind(func() (LoadResult, error) {
				return sessionForTest(directContext).LoadCLI(ProjectSet{}, plan, configDir, true)
			})

			legacyError, directError := "", ""
			if legacy.err != nil {
				legacyError = legacy.err.Error()
			}
			if direct.err != nil {
				directError = direct.err.Error()
			}
			if !test.supported {
				if legacyError != directError || legacy.panicText != direct.panicText ||
					(legacyError == "" && legacy.panicText == "") {
					t.Fatalf("unsupported-root behavior differs: legacy=(%q, %q) direct=(%q, %q)", legacyError, legacy.panicText, directError, direct.panicText)
				}
				return
			}
			if legacy.err != nil || direct.err != nil || legacy.panicText != "" || direct.panicText != "" {
				t.Fatalf("supported root failed: legacy=(%v, %q) direct=(%v, %q)", legacy.err, legacy.panicText, direct.err, direct.panicText)
			}
			if len(direct.binding.compilerPrograms) != 0 || len(direct.binding.Programs) != 1 {
				t.Fatalf("supported root did not use root construction: compiler programs=%d Programs=%d", len(direct.binding.compilerPrograms), len(direct.binding.Programs))
			}
			legacyDiagnostics := collectTargetSyntacticDiagnostics(
				legacy.binding.Programs,
				legacy.binding.TargetsByProgram,
				false,
				false,
			)
			directDiagnostics := collectTargetSyntacticDiagnostics(
				direct.binding.Programs,
				direct.binding.TargetsByProgram,
				false,
				false,
			)
			compareProgramSyntaxDiagnostics(t, legacyDiagnostics, directDiagnostics)
		})
	}
}

type rootProgramExactCaseFS struct {
	*exactCaseProgramFS
	directories map[string]struct{}
}

func (fsys *rootProgramExactCaseFS) DirectoryExists(path string) bool {
	if _, ok := fsys.directories[tspath.NormalizePath(path)]; ok {
		return true
	}
	return fsys.exactCaseProgramFS.DirectoryExists(path)
}

func TestRootProgramsIsolateCaseFoldedPackageScopes(t *testing.T) {
	const configDir = "/repo"
	const upper = "/repo/node_modules/Pkg/index.js"
	const lower = "/repo/node_modules/pkg/index.js"
	files := map[string]string{
		upper:                                 "const upper = ;\n",
		lower:                                 "const lower = ;\n",
		"/repo/node_modules/Pkg/package.json": "{\"type\":\"module\"}",
		"/repo/node_modules/pkg/package.json": "{\"type\":\"commonjs\"}",
	}
	directories := map[string]struct{}{
		"/": {}, "/repo": {}, "/repo/node_modules": {},
		"/repo/node_modules/Pkg": {}, "/repo/node_modules/pkg": {},
	}
	newFS := func() vfs.FS {
		return &rootProgramExactCaseFS{
			exactCaseProgramFS: &exactCaseProgramFS{FS: osvfs.FS(), files: files},
			directories:        directories,
		}
	}
	plan := target.Plan{Files: []target.File{
		{PathIdentity: rslintconfig.PathIdentity{Path: upper, CanonicalPath: upper}, ConfigDirectory: configDir},
		{PathIdentity: rslintconfig.PathIdentity{Path: lower, CanonicalPath: lower}, ConfigDirectory: configDir},
	}}

	legacyContext := newBuildContext(newFS())
	legacy, err := loadAPIForTest(ProjectSet{}, plan, configDir, legacyContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	directContext := newBuildContext(newFS())
	direct, err := sessionForTest(directContext).LoadCLI(ProjectSet{}, plan, configDir, true)
	if err != nil {
		t.Fatalf("prepareCLIForTest: %v", err)
	}
	if len(direct.Programs) != 2 {
		t.Fatalf("case-folded targets share a Program: Programs=%d", len(direct.Programs))
	}
	directDiagnostics := collectTargetSyntacticDiagnostics(
		direct.Programs,
		direct.TargetsByProgram,
		false,
		false,
	)
	legacyDiagnostics := collectTargetSyntacticDiagnostics(
		legacy.Programs,
		legacy.TargetsByProgram,
		false,
		false,
	)
	compareProgramSyntaxDiagnostics(t, legacyDiagnostics, directDiagnostics)

	directExternal := make(map[string]bool, 2)
	for _, sourceProgram := range direct.Programs {
		for _, file := range sourceProgram.SourceFiles() {
			directExternal[file.FileName()] = file.ExternalModuleIndicator != nil
		}
	}
	legacyExternal := make(map[string]bool, 2)
	for _, program := range legacy.Programs {
		for _, file := range program.SourceFiles() {
			legacyExternal[file.FileName()] = file.ExternalModuleIndicator != nil
		}
	}
	if !legacyExternal[upper] || legacyExternal[lower] ||
		directExternal[upper] != legacyExternal[upper] || directExternal[lower] != legacyExternal[lower] {
		t.Fatalf("package scope leaked across exact-case targets: legacy=%v direct=%v", legacyExternal, directExternal)
	}
}

func TestSourceOnlyProgramSyntaxDeduplicatesAgainstNonGoverningTypeCheckProgram(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json":   "{\"include\":[\"child/target.ts\"]}",
		"child/target.ts": "let value: ;\n",
	})
	rootDir = tspath.NormalizePath(rootDir)
	childDir = tspath.NormalizePath(childDir)
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:  projectConfig("./tsconfig.json"),
		childDir: {},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfigs(configMap, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	targetPath := tspath.ResolvePath(childDir, "target.ts")
	plan := target.Plan{Files: []target.File{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := sessionForTest(buildContext).LoadCLI(set, plan, rootDir, true)
	if err != nil {
		t.Fatalf("prepareCLIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 1 || len(binding.Programs) != 2 {
		t.Fatalf("fixture did not create a non-governing Program plus source-only Program: compiler=%d Programs=%d", len(binding.compilerPrograms), len(binding.Programs))
	}

	diagnostics := collectTargetSyntacticDiagnostics(
		binding.Programs,
		binding.TargetsByProgram,
		true,
		false,
	)
	diagnostics = append(
		diagnostics,
		collectProgramTypeDiagnostics(t, binding.Programs)...,
	)
	if len(diagnostics) < 2 {
		t.Fatalf("fixture did not produce cross-phase duplicate diagnostics: %+v", diagnostics)
	}
	diagnostics = deduplicateTypeScriptDiagnostics(diagnostics, fsys, preferredCallerPathsForTest(plan))
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "TypeScript(TS1110)" || diagnostics[0].FilePath != targetPath {
		t.Fatalf("cross-phase diagnostic dedupe differs: %+v", diagnostics)
	}
}

func sessionForTest(context *buildContext) *Session {
	return &Session{context: context}
}

func buildProjectsForConfigs(
	configs map[string]rslintconfig.RslintConfig,
	singleThreaded bool,
	context *buildContext,
) (ProjectSet, error) {
	return sessionForTest(context).BuildProjects(configs, singleThreaded)
}

func buildProjectsForConfig(
	configDirectory string,
	config rslintconfig.RslintConfig,
	singleThreaded bool,
	context *buildContext,
) (ProjectSet, error) {
	return sessionForTest(context).BuildProject(configDirectory, config, singleThreaded)
}

func executeProjectPlanForTest(
	plan projectPlan,
	singleThreaded bool,
	context *buildContext,
) (ProjectSet, error) {
	return sessionForTest(context).executeProjectPlan(plan, singleThreaded)
}

func resolveTargetPlanForTest(
	configMap map[string]rslintconfig.RslintConfig,
	config rslintconfig.RslintConfig,
	currentDirectory string,
	scopes map[string]target.OwnerScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (target.Plan, error) {
	return target.Resolve(target.Request{
		ConfigMap:       configMap,
		Config:          config,
		ConfigDirectory: currentDirectory,
		ScanRoot:        currentDirectory,
		OwnerScopes:     scopes,
		FS:              fsys,
		Files:           allowFiles,
		Directories:     allowDirs,
		SingleThreaded:  singleThreaded,
	})
}

func preferredCallerPathsForTest(plan target.Plan) map[string]string {
	return plan.PreferredCallerPaths()
}

func loadAPIForTest(
	projects ProjectSet,
	plan target.Plan,
	currentDirectory string,
	context *buildContext,
	singleThreaded bool,
) (LoadResult, error) {
	return sessionForTest(context).LoadAPI(projects, plan, currentDirectory, singleThreaded)
}

func createCompatibilityProgramForTest(
	rootFileNames []string,
	singleThreaded bool,
	currentDirectory string,
	context *buildContext,
) (*compiler.Program, error) {
	program, err := context.createCompatibilityProgram(
		singleThreaded,
		currentDirectory,
		sourceOnlyCompilerOptions(),
		rootFileNames,
	)
	if err != nil {
		return nil, fmt.Errorf("create compatibility Program for %d lint target(s): %w", len(rootFileNames), err)
	}
	return program, nil
}

func buildRootProgramsForTest(
	groups [][]target.File,
	currentDirectory string,
	context *buildContext,
	singleThreaded bool,
) ([]*lintprogram.Program, []rule.RuleDiagnostic, error) {
	result := LoadResult{}
	if err := sessionForTest(context).appendRootPrograms(&result, groups, currentDirectory, singleThreaded); err != nil {
		return nil, nil, err
	}
	diagnostics := collectTargetSyntacticDiagnostics(result.Programs, result.TargetsByProgram, false, false)
	return result.Programs, diagnostics, nil
}

func collectTargetSyntacticDiagnostics(
	programs []*lintprogram.Program,
	targetsByProgram [][]string,
	typeCheck bool,
	typeCheckOnly bool,
) []rule.RuleDiagnostic {
	type diagnosticKey struct {
		path string
		code int32
		pos  int
		end  int
	}
	seen := make(map[diagnosticKey]struct{})
	var diagnostics []rule.RuleDiagnostic
	for index, sourceProgram := range programs {
		if index >= len(targetsByProgram) {
			continue
		}
		coveredByTypeCheck := typeCheck && sourceProgram.CanProvideProgramDiagnostics()
		for _, target := range targetsByProgram[index] {
			file := sourceProgram.GetSourceFile(target)
			if file == nil {
				continue
			}
			for _, diagnostic := range sourceProgram.SyntacticDiagnostics(context.Background(), file) {
				if coveredByTypeCheck || typeCheckOnly {
					continue
				}
				loc := diagnostic.Loc()
				key := diagnosticKey{file.FileName(), diagnostic.Code(), loc.Pos(), loc.End()}
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				diagnostics = append(diagnostics, rule.RuleDiagnostic{
					RuleName:     fmt.Sprintf("TypeScript(TS%d)", diagnostic.Code()),
					SourceFile:   file,
					FilePath:     file.FileName(),
					Range:        loc,
					Message:      rule.RuleMessage{Description: diagnostic.String()},
					Severity:     rule.SeverityError,
					Origin:       rule.DiagnosticOriginTypeScript,
					PreFormatted: true,
				})
			}
		}
	}
	return diagnostics
}

func remapDiagnosticTargetPaths(
	diagnostics []rule.RuleDiagnostic,
	mapping map[string]target.File,
) {
	for index := range diagnostics {
		if target, ok := mapping[diagnostics[index].FilePath]; ok {
			diagnostics[index].FilePath = target.Path
		}
	}
}

func deduplicateTypeScriptDiagnostics(
	diagnostics []rule.RuleDiagnostic,
	fsys vfs.FS,
	preferred ...map[string]string,
) []rule.RuleDiagnostic {
	type diagnosticKey struct {
		path string
		code string
		pos  int
		end  int
		text string
	}
	var preferredPaths map[string]string
	if len(preferred) > 0 {
		preferredPaths = preferred[0]
	}
	seen := make(map[diagnosticKey]int)
	result := make([]rule.RuleDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if diagnostic.Origin != rule.DiagnosticOriginTypeScript {
			result = append(result, diagnostic)
			continue
		}
		canonical := canonicalPathID(diagnostic.FilePath, fsys)
		key := diagnosticKey{
			path: canonical,
			code: diagnostic.RuleName,
			pos:  diagnostic.Range.Pos(),
			end:  diagnostic.Range.End(),
			text: diagnostic.Message.Description,
		}
		if existingIndex, exists := seen[key]; exists {
			preferredPath := preferredPaths[canonical]
			if preferredPath != "" && exactPathID(diagnostic.FilePath) == exactPathID(preferredPath) {
				result[existingIndex] = diagnostic
			}
			continue
		}
		seen[key] = len(result)
		result = append(result, diagnostic)
	}
	return result
}

type lintConfigResolverOptions struct {
	Config                 rslintconfig.RslintConfig
	CurrentDirectory       string
	LintTargetBySourcePath map[string]target.File
	FS                     vfs.FS
}

type testLintConfigResolver struct {
	resolver           *rslintconfig.FileConfigResolver
	lintTargetBySource map[string]target.File
}

func newLintConfigResolver(opts lintConfigResolverOptions) *testLintConfigResolver {
	return &testLintConfigResolver{
		resolver: rslintconfig.NewFileConfigResolverWithFS(
			opts.Config,
			opts.CurrentDirectory,
			opts.FS,
			rules.All(),
		),
		lintTargetBySource: opts.LintTargetBySourcePath,
	}
}

func (resolver *testLintConfigResolver) EnabledRulesForFile(fileName string) []rule.ConfiguredRule {
	target, ok := resolver.lintTargetBySource[fileName]
	if !ok {
		target.Path = fileName
	}
	rules, _ := resolver.resolver.EnabledRulesForTarget(target.Path, target.CanonicalPath)
	return rules
}

func configuredRuleNameSet(rules []rule.ConfiguredRule) map[string]struct{} {
	result := make(map[string]struct{}, len(rules))
	for _, configured := range rules {
		result[configured.Name] = struct{}{}
	}
	return result
}
