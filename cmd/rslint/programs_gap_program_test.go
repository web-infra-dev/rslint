package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	default_rule "github.com/web-infra-dev/rslint/internal/plugins/import/rules/default"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_cycle"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func gapProgramTestPlan(dir string, names ...string) lintTargetPlan {
	plan := lintTargetPlan{Targets: make([]resolvedLintTarget, 0, len(names))}
	for _, name := range names {
		fileName := tspath.NormalizePath(filepath.Join(dir, name))
		plan.Targets = append(plan.Targets, resolvedLintTarget{
			Path:           fileName,
			CanonicalPath:  fileName,
			OwnerConfigDir: dir,
		})
	}
	return plan
}

func gapProgramTestBuildContext() *utils.ProgramBuildContext {
	return utils.NewProgramBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
}

func compareGapProgramSyntaxDiagnostics(t *testing.T, legacy, direct []rule.RuleDiagnostic) {
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

func TestGapProgramsMatchFallbackProgram(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"package.json":  "{\"type\":\"module\"}",
		"a.js":          "import './b.js';\nimport './unselected.js';\nclass C { method(@dec value) {} }\n",
		"b.js":          "import './a.js';\nexport const value = 1;\n",
		"unselected.js": "export const unselected = 1;\n",
		"bad.ts":        "export const value: = 1;\n",
		"view.tsx":      "export const view = <div>;\n",
	})
	plan := gapProgramTestPlan(dir, "a.js", "b.js", "bad.ts", "view.tsx")
	rootFiles := make([]string, 0, len(plan.Targets))
	for _, target := range plan.Targets {
		rootFiles = append(rootFiles, target.Path)
	}

	legacyContext := gapProgramTestBuildContext()
	fallback, err := createFallbackProgram(rootFiles, true, dir, legacyContext)
	if err != nil {
		t.Fatalf("createFallbackProgram: %v", err)
	}
	legacyDiagnostics, _ := collectTargetSyntacticDiagnostics(
		[]*compiler.Program{fallback},
		[][]string{rootFiles},
		[]bool{true},
		false,
		false,
	)

	directContext := gapProgramTestBuildContext()
	programs, directDiagnostics, _, err := buildGapPrograms(
		[][]resolvedLintTarget{plan.Targets},
		dir,
		directContext,
		false, // Exercise the production parallel parse/bind path.
	)
	if err != nil {
		t.Fatalf("buildGapPrograms: %v", err)
	}
	compareGapProgramSyntaxDiagnostics(t, legacyDiagnostics, directDiagnostics)
	if len(programs) != 1 || len(programs[0].SourceFiles()) != len(plan.Targets) {
		t.Fatalf("gap Program source count differs: programs=%d files=%d", len(programs), len(programs[0].SourceFiles()))
	}
	sourceProgram := programs[0]

	for i, target := range plan.Targets {
		programFile := exactProgramSourceFile(fallback, target.Path)
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
		if legacyMetadata := fallback.GetSourceFileMetaData(programFile.Path()); legacyMetadata != sourceProgram.SourceFileMetadata(directFile) {
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
			mode := fallback.GetModeForUsageLocation(programFile, specifier)
			legacyResolution := fallback.GetResolvedModule(programFile, specifier.Text(), mode)
			directResolution := sourceProgram.GetResolvedModule(directFile, specifier.Text(), mode)
			if legacyResolution == nil || directResolution == nil ||
				legacyResolution.ResolvedFileName != directResolution.ResolvedFileName ||
				legacyResolution.IsResolved() != directResolution.IsResolved() {
				t.Fatalf("module resolution differs for %q in %q: legacy=%+v direct=%+v", specifier.Text(), target.Path, legacyResolution, directResolution)
			}
			if legacyResolution.IsResolved() {
				legacySource := fallback.GetSourceFileForResolvedModule(legacyResolution.ResolvedFileName)
				directSource := sourceProgram.GetSourceFileForResolvedModule(directResolution.ResolvedFileName)
				if (legacySource == nil) != (directSource == nil) ||
					legacySource != nil && legacySource.FileName() != directSource.FileName() {
					t.Fatalf("resolved source lookup differs for %q", specifier.Text())
				}
			}
		}
	}
	unselected := tspath.ResolvePath(dir, "unselected.js")
	if fallback.GetSourceFile(unselected) != nil || sourceProgram.GetSourceFile(unselected) != nil {
		t.Fatal("NoResolve Programs materialized an unselected dependency")
	}
}

func TestGapProgramsReportUnreadableTarget(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	target := tspath.ResolvePath(dir, "missing.ts")
	_, _, _, err := buildGapPrograms(
		[][]resolvedLintTarget{{{
			Path:           target,
			CanonicalPath:  target,
			OwnerConfigDir: dir,
		}}},
		dir,
		gapProgramTestBuildContext(),
		true,
	)
	want := fmt.Sprintf("program: could not read root %q", target)
	if err == nil || err.Error() != want {
		t.Fatalf("unreadable gap target error = %v, want %q", err, want)
	}
}

func TestGapProgramSupportsCrossFileImportRules(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts": "import value from './b';\nexport const a = value;\n",
		"b.ts": "import './a';\nexport const b = 1;\n",
	})
	plan := gapProgramTestPlan(dir, "a.ts", "b.ts")
	programs, diagnostics, syntaxErrorFiles, err := buildGapPrograms(
		[][]resolvedLintTarget{plan.Targets},
		dir,
		gapProgramTestBuildContext(),
		true,
	)
	if err != nil {
		t.Fatalf("buildGapPrograms: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected gap Program syntax diagnostics: %+v", diagnostics)
	}

	var cycleReports, defaultReports int
	opts := linter.RunLinterOptions{
		Programs:         programs,
		SingleThreaded:   true,
		TypeInfoFiles:    map[string]struct{}{},
		SyntaxErrorFiles: syntaxErrorFiles,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{
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
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			switch diagnostic.RuleName {
			case no_cycle.NoCycleRule.Name:
				cycleReports++
			case default_rule.DefaultRule.Name:
				defaultReports++
			default:
				t.Errorf("unexpected gap Program import diagnostic: %+v", diagnostic)
			}
		}},
	}
	opts.PreparedPlan, err = linter.PrepareLintPlan(opts)
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	result, err := linter.RunLinter(opts)
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if result.LintedFileCount != 2 || cycleReports != 2 || defaultReports != 1 {
		t.Fatalf(
			"gap Program import results differ: files=%d cycles=%d defaults=%d",
			result.LintedFileCount,
			cycleReports,
			defaultReports,
		)
	}
}

type gapProgramBindOutcome struct {
	binding   lintTargetBinding
	err       error
	panicText string
}

func runGapProgramBind(bind func() (lintTargetBinding, error)) (outcome gapProgramBindOutcome) {
	defer func() {
		if value := recover(); value != nil {
			outcome.panicText = fmt.Sprint(value)
		}
	}()
	outcome.binding, outcome.err = bind()
	return outcome
}

func TestBindCLILintTargetPlanMatchesFallbackRootAdmission(t *testing.T) {
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
		{name: "json enabled by fallback defaults", fileName: "target.json", caseSensitive: true, supported: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const configDir = "/repo"
			targetPath := tspath.ResolvePath(configDir, test.fileName)
			newContext := func() *utils.ProgramBuildContext {
				fsys := newBindingIndexTestFS([]string{targetPath}, nil)
				fsys.caseSensitive = test.caseSensitive
				fsys.files[targetPath] = "const value = ;\n"
				return utils.NewProgramBuildContext(fsys)
			}
			plan := lintTargetPlan{Targets: []resolvedLintTarget{{
				Path:           targetPath,
				CanonicalPath:  targetPath,
				OwnerConfigDir: configDir,
			}}}

			legacyContext := newContext()
			legacy := runGapProgramBind(func() (lintTargetBinding, error) {
				return bindLintTargetPlan(lintProgramSet{}, plan, configDir, legacyContext, true)
			})
			directContext := newContext()
			direct := runGapProgramBind(func() (lintTargetBinding, error) {
				return bindCLILintTargetPlan(lintProgramSet{}, plan, configDir, directContext, true)
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
			if len(direct.binding.Programs) != 0 || len(direct.binding.GapGroups) != 1 {
				t.Fatalf("supported root did not use root construction: programs=%d groups=%d", len(direct.binding.Programs), len(direct.binding.GapGroups))
			}
			legacyDiagnostics, _ := collectTargetSyntacticDiagnostics(
				legacy.binding.Programs,
				legacy.binding.TargetsByProgram,
				buildTypeCheckSkipMask(legacy.binding.Programs),
				false,
				false,
			)
			_, directDiagnostics, _, err := buildGapPrograms(
				direct.binding.GapGroups,
				configDir,
				directContext,
				true,
			)
			if err != nil {
				t.Fatalf("buildGapPrograms: %v", err)
			}
			compareGapProgramSyntaxDiagnostics(t, legacyDiagnostics, directDiagnostics)
		})
	}
}

type gapProgramExactCaseFS struct {
	*exactCaseProgramFS
	directories map[string]struct{}
}

func (fsys *gapProgramExactCaseFS) DirectoryExists(path string) bool {
	if _, ok := fsys.directories[tspath.NormalizePath(path)]; ok {
		return true
	}
	return fsys.exactCaseProgramFS.DirectoryExists(path)
}

func TestGapProgramsIsolateCaseFoldedPackageScopes(t *testing.T) {
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
		return &gapProgramExactCaseFS{
			exactCaseProgramFS: &exactCaseProgramFS{FS: osvfs.FS(), files: files},
			directories:        directories,
		}
	}
	plan := lintTargetPlan{Targets: []resolvedLintTarget{
		{Path: upper, CanonicalPath: upper, OwnerConfigDir: configDir},
		{Path: lower, CanonicalPath: lower, OwnerConfigDir: configDir},
	}}

	legacyContext := utils.NewProgramBuildContext(newFS())
	legacy, err := bindLintTargetPlan(lintProgramSet{}, plan, configDir, legacyContext, true)
	if err != nil {
		t.Fatalf("bindLintTargetPlan: %v", err)
	}
	directContext := utils.NewProgramBuildContext(newFS())
	direct, err := bindCLILintTargetPlan(lintProgramSet{}, plan, configDir, directContext, true)
	if err != nil {
		t.Fatalf("bindCLILintTargetPlan: %v", err)
	}
	if len(direct.GapGroups) != 2 {
		t.Fatalf("case-folded targets share a Program: groups=%d", len(direct.GapGroups))
	}
	programs, directDiagnostics, _, err := buildGapPrograms(
		direct.GapGroups,
		configDir,
		directContext,
		true,
	)
	if err != nil {
		t.Fatalf("buildGapPrograms: %v", err)
	}
	legacyDiagnostics, _ := collectTargetSyntacticDiagnostics(
		legacy.Programs,
		legacy.TargetsByProgram,
		buildTypeCheckSkipMask(legacy.Programs),
		false,
		false,
	)
	compareGapProgramSyntaxDiagnostics(t, legacyDiagnostics, directDiagnostics)

	directExternal := make(map[string]bool, 2)
	for _, sourceProgram := range programs {
		for _, file := range sourceProgram.SourceFiles() {
			directExternal[file.FileName()] = file.ExternalModuleIndicator != nil
		}
	}
	legacyExternal := make(map[string]bool, 2)
	for _, program := range legacy.Programs {
		for _, file := range program.GetSourceFiles() {
			legacyExternal[file.FileName()] = file.ExternalModuleIndicator != nil
		}
	}
	if !legacyExternal[upper] || legacyExternal[lower] ||
		directExternal[upper] != legacyExternal[upper] || directExternal[lower] != legacyExternal[lower] {
		t.Fatalf("package scope leaked across exact-case targets: legacy=%v direct=%v", legacyExternal, directExternal)
	}
}

func TestGapProgramSyntaxDeduplicatesAgainstNonGoverningTypeCheckProgram(t *testing.T) {
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
	buildContext := utils.NewProgramBuildContext(fsys)
	set, err := createProgramSetForConfigs(configMap, true, buildContext)
	if err != nil {
		t.Fatalf("createProgramSetForConfigs: %v", err)
	}
	targetPath := tspath.ResolvePath(childDir, "target.ts")
	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := bindCLILintTargetPlan(set, plan, rootDir, buildContext, true)
	if err != nil {
		t.Fatalf("bindCLILintTargetPlan: %v", err)
	}
	if len(binding.Programs) != 1 || len(binding.GapGroups) != 1 {
		t.Fatalf("fixture did not create a non-governing Program plus gap Program: programs=%d groups=%d", len(binding.Programs), len(binding.GapGroups))
	}

	_, diagnostics, _, err := buildGapPrograms(
		binding.GapGroups,
		rootDir,
		buildContext,
		true,
	)
	if err != nil {
		t.Fatalf("buildGapPrograms: %v", err)
	}
	diagnostics = append(
		diagnostics,
		collectProgramTypeDiagnostics(t, binding.Programs, buildTypeCheckSkipMask(binding.Programs), binding.TypeInfoFiles)...,
	)
	if len(diagnostics) < 2 {
		t.Fatalf("fixture did not produce cross-phase duplicate diagnostics: %+v", diagnostics)
	}
	diagnostics = deduplicateTypeScriptDiagnostics(diagnostics, fsys, preferredCallerTargetPaths(plan))
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "TypeScript(TS1110)" || diagnostics[0].FilePath != targetPath {
		t.Fatalf("cross-phase diagnostic dedupe differs: %+v", diagnostics)
	}
}
