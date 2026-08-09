package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type rulelessGapExactCaseFS struct {
	*exactCaseProgramFS
	directories map[string]struct{}
}

func (fsys *rulelessGapExactCaseFS) DirectoryExists(path string) bool {
	if _, ok := fsys.directories[tspath.NormalizePath(path)]; ok {
		return true
	}
	return fsys.exactCaseProgramFS.DirectoryExists(path)
}

func rulelessGapTestPlan(dir string, names ...string) lintTargetPlan {
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

func rulelessGapTestBuildContext() *utils.ProgramBuildContext {
	return utils.NewProgramBuildContext(bundled.WrapFS(cachedvfs.From(osvfs.FS())))
}

func bindRulelessGapTestCLI(
	t *testing.T,
	dir string,
	plan lintTargetPlan,
	cfg rslintconfig.RslintConfig,
	enforcePlugins bool,
) (lintTargetBinding, *lintConfigResolver) {
	t.Helper()
	binding, resolver, err := bindCLILintTargetPlan(
		lintProgramSet{},
		plan,
		dir,
		rulelessGapTestBuildContext(),
		true,
		lintConfigResolverOptions{
			Config:           cfg,
			CurrentDirectory: dir,
			EnforcePlugins:   enforcePlugins,
		},
	)
	if err != nil {
		t.Fatalf("bindCLILintTargetPlan: %v", err)
	}
	return binding, resolver
}

func TestBindCLILintTargetPlan_OnlyRulelessGapsStayStandalone(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts": "export const a = 1;\n",
		"b.ts": "export const b = 2;\n",
	})
	plan := rulelessGapTestPlan(dir, "a.ts", "b.ts")
	binding, resolver := bindRulelessGapTestCLI(t, dir, plan, rslintconfig.RslintConfig{{}}, false)

	if len(binding.Programs) != 0 || len(binding.standaloneGapGroups) != 1 {
		t.Fatalf("ruleless gaps must avoid fallback Programs: programs=%d standalone=%v", len(binding.Programs), binding.standaloneGapGroups)
	}
	if binding.TypeInfoFiles == nil || len(binding.TypeInfoFiles) != 0 {
		t.Fatalf("gap binding must retain an explicit empty type-info set: %v", binding.TypeInfoFiles)
	}
	for _, target := range plan.Targets {
		if rules := resolver.ActiveRulesForFile(target.Path); len(rules) != 0 {
			t.Fatalf("ruleless target %q resolved rules: %+v", target.Path, rules)
		}
		if got := binding.OwnerConfigDirBySourcePath[target.Path]; got != dir {
			t.Fatalf("gap owner mapping = %q, want %q", got, dir)
		}
		wantConfigPath := configPathForLintTarget(target, rulelessGapTestBuildContext().FS())
		if got := binding.ConfigPathBySourcePath[target.Path]; got != wantConfigPath {
			t.Fatalf("gap config path = %q, want %q", got, wantConfigPath)
		}
	}
}

func TestBindCLILintTargetPlan_AnyExecutableRuleKeepsAllGapsInFallback(t *testing.T) {
	rslintconfig.RegisterAllRules()
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts": "debugger;\n",
		"b.ts": "export const b = 2;\n",
	})
	plan := rulelessGapTestPlan(dir, "a.ts", "b.ts")
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"a.ts"},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	binding, _ := bindRulelessGapTestCLI(t, dir, plan, cfg, false)

	if len(binding.Programs) != 1 || len(binding.standaloneGapGroups) != 0 {
		t.Fatalf("one active gap rule must preserve the whole fallback set: programs=%d standalone=%v", len(binding.Programs), binding.standaloneGapGroups)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 2 {
		t.Fatalf("fallback targets = %v, want both gaps", binding.TargetsByProgram)
	}
}

func TestBindCLILintTargetPlan_TypeAwareOnlyGapRuleIsNotExecutable(t *testing.T) {
	rslintconfig.RegisterAllRules()
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{"a.ts": "async function f() {}\n"})
	plan := rulelessGapTestPlan(dir, "a.ts")
	cfg := rslintconfig.RslintConfig{{
		Rules:   rslintconfig.Rules{"@typescript-eslint/require-await": "error"},
		Plugins: []string{"@typescript-eslint"},
	}}
	binding, resolver := bindRulelessGapTestCLI(t, dir, plan, cfg, false)

	if len(rslintconfig.GlobalRuleRegistry.GetActiveRulesForFile(cfg, plan.Targets[0].Path, dir, false, nil)) == 0 {
		t.Fatal("fixture must enable a registered type-aware rule before gap filtering")
	}
	if rules := resolver.ActiveRulesForFile(plan.Targets[0].Path); len(rules) != 0 {
		t.Fatalf("type-aware rule survived the gap gate: %+v", rules)
	}
	if len(binding.Programs) != 0 || len(binding.standaloneGapGroups) != 1 {
		t.Fatalf("type-aware-only gap should stay standalone: programs=%d standalone=%v", len(binding.Programs), binding.standaloneGapGroups)
	}
}

func TestBindCLILintTargetPlan_PluginRuleKeepsFallback(t *testing.T) {
	rslintconfig.RegisterAllRules()
	rslintconfig.RegisterEslintPluginRules([]rslintconfig.EslintPluginEntry{{
		Prefix:    "gap-test-plugin",
		RuleNames: []string{"rule"},
	}})
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{"a.ts": "export {};\n"})
	plan := rulelessGapTestPlan(dir, "a.ts")
	cfg := rslintconfig.RslintConfig{{
		Rules:   rslintconfig.Rules{"gap-test-plugin/rule": "error"},
		Plugins: []string{"gap-test-plugin"},
	}}
	binding, resolver := bindRulelessGapTestCLI(t, dir, plan, cfg, true)

	rules := resolver.ActiveRulesForFile(plan.Targets[0].Path)
	if len(rules) != 1 || !rules[0].IsEslintPluginRule {
		t.Fatalf("fixture did not resolve the plugin rule: %+v", rules)
	}
	if len(binding.Programs) != 1 || len(binding.standaloneGapGroups) != 0 {
		t.Fatalf("plugin gap must retain fallback Program: programs=%d standalone=%v", len(binding.Programs), binding.standaloneGapGroups)
	}
}

func TestBindCLILintTargetPlan_NoGapResolverKeepsNilTypeInfoFilter(t *testing.T) {
	rslintconfig.RegisterAllRules()
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts":          "async function f() {}\n",
		"tsconfig.json": `{"files":["a.ts"]}`,
	})
	cfg := rslintconfig.RslintConfig{{
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
		Plugins: []string{"@typescript-eslint"},
		Rules:   rslintconfig.Rules{"@typescript-eslint/require-await": "error"},
	}}
	buildContext := rulelessGapTestBuildContext()
	set, err := createProgramSetForConfig(dir, cfg, true, buildContext)
	if err != nil {
		t.Fatalf("createProgramSetForConfig: %v", err)
	}
	binding, resolver, err := bindCLILintTargetPlan(
		set,
		rulelessGapTestPlan(dir, "a.ts"),
		dir,
		buildContext,
		true,
		lintConfigResolverOptions{Config: cfg, CurrentDirectory: dir},
	)
	if err != nil {
		t.Fatalf("bindCLILintTargetPlan: %v", err)
	}
	if binding.TypeInfoFiles != nil || resolver.typeInfoFiles != nil {
		t.Fatalf("no-gap binding must preserve the nil type-info invariant: binding=%v resolver=%v", binding.TypeInfoFiles, resolver.typeInfoFiles)
	}
	if rules := resolver.ActiveRulesForFile(binding.TargetsByProgram[0][0]); len(rules) != 1 || !rules[0].RequiresTypeInfo {
		t.Fatalf("real Program target lost its type-aware rule: %+v", rules)
	}
}

func TestBindCLILintTargetPlan_MatchesFallbackRootAdmission(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		caseSensitive bool
		supported     bool
	}{
		{name: "custom extension", fileName: "target.vue", caseSensitive: true},
		{name: "json enabled by fallback defaults", fileName: "target.json", caseSensitive: true, supported: true},
		{name: "extensionless", fileName: "target", caseSensitive: true},
		{name: "upper-case extension on case-sensitive FS", fileName: "target.TS", caseSensitive: true},
		{name: "upper-case extension on case-insensitive FS", fileName: "target.TS", caseSensitive: false, supported: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configDir := "/repo"
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

			type bindOutcome struct {
				binding   lintTargetBinding
				err       error
				panicText string
			}
			runBind := func(bind func() (lintTargetBinding, error)) (outcome bindOutcome) {
				defer func() {
					if value := recover(); value != nil {
						outcome.panicText = fmt.Sprint(value)
					}
				}()
				outcome.binding, outcome.err = bind()
				return outcome
			}

			legacyContext := newContext()
			legacy := runBind(func() (lintTargetBinding, error) {
				return bindLintTargetPlan(lintProgramSet{}, plan, configDir, legacyContext, true)
			})
			directContext := newContext()
			direct := runBind(func() (lintTargetBinding, error) {
				binding, _, err := bindCLILintTargetPlan(
					lintProgramSet{},
					plan,
					configDir,
					directContext,
					true,
					lintConfigResolverOptions{Config: rslintconfig.RslintConfig{{}}, CurrentDirectory: configDir},
				)
				return binding, err
			})

			if !test.supported {
				legacyError, directError := "", ""
				if legacy.err != nil {
					legacyError = legacy.err.Error()
				}
				if direct.err != nil {
					directError = direct.err.Error()
				}
				if legacyError != directError || legacy.panicText != direct.panicText ||
					(legacyError == "" && legacy.panicText == "") {
					t.Fatalf("unsupported-root failure differs: legacy=(%q, %q) direct=(%q, %q)", legacyError, legacy.panicText, directError, direct.panicText)
				}
				return
			}
			if legacy.err != nil || direct.err != nil || legacy.panicText != "" || direct.panicText != "" {
				t.Fatalf("supported root failed: legacy=(%v, %q) direct=(%v, %q)", legacy.err, legacy.panicText, direct.err, direct.panicText)
			}
			if len(direct.binding.Programs) != 0 || len(direct.binding.standaloneGapGroups) != 1 {
				t.Fatalf("supported ruleless root should stay standalone: programs=%d groups=%d", len(direct.binding.Programs), len(direct.binding.standaloneGapGroups))
			}
			legacyDiagnostics, _ := collectTargetSyntacticDiagnostics(
				legacy.binding.Programs,
				legacy.binding.TargetsByProgram,
				buildTypeCheckSkipMask(legacy.binding.Programs),
				false,
				false,
			)
			directDiagnostics, _, err := collectStandaloneGapSyntacticDiagnostics(
				direct.binding.standaloneGapGroups,
				configDir,
				directContext,
				true,
			)
			if err != nil {
				t.Fatalf("collectStandaloneGapSyntacticDiagnostics: %v", err)
			}
			if len(legacyDiagnostics) != len(directDiagnostics) {
				t.Fatalf("supported-root diagnostic count differs: legacy=%+v direct=%+v", legacyDiagnostics, directDiagnostics)
			}
			for i := range legacyDiagnostics {
				if legacyDiagnostics[i].RuleName != directDiagnostics[i].RuleName ||
					legacyDiagnostics[i].Range != directDiagnostics[i].Range ||
					legacyDiagnostics[i].Message.Description != directDiagnostics[i].Message.Description {
					t.Fatalf("supported-root diagnostic %d differs: legacy=%+v direct=%+v", i, legacyDiagnostics[i], directDiagnostics[i])
				}
			}
		})
	}
}

func TestFallbackSourceParser_MatchesProgramSyntax(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	files := map[string]string{
		"a.ts":       "export const a: = 1;\n",
		"a.tsx":      "export const view = <div>;\n",
		"a.js":       "class C { method(@dec value) {} }\n",
		"checked.js": "// @ts-check\nclass C { method(@dec value) {} }\n",
		"a.jsx":      "/** @param { */\nexport const view = <div />;\n",
		"a.mjs":      "export const a = ;\n",
		"a.cjs":      "module.exports = { value: };\n",
		"a.mts":      "export const a: = 1;\n",
		"a.cts":      "export const a: = 1;\n",
		// cspell:disable-next-line
		"bom.ts": "\ufeffexport const a: = 1;\n",
	}
	writeProgramTestFiles(t, dir, files)
	buildContext := rulelessGapTestBuildContext()
	plan := rulelessGapTestPlan(dir, "a.ts", "a.tsx", "a.js", "checked.js", "a.jsx", "a.mjs", "a.cjs", "a.mts", "a.cts", "bom.ts")
	rootFiles := make([]string, 0, len(plan.Targets))
	for _, target := range plan.Targets {
		rootFiles = append(rootFiles, target.Path)
	}
	fallback, err := createFallbackProgram(rootFiles, true, dir, buildContext)
	if err != nil {
		t.Fatalf("createFallbackProgram: %v", err)
	}
	parser := newFallbackSourceParser(dir, buildContext)

	for _, target := range plan.Targets {
		t.Run(filepath.Base(target.Path), func(t *testing.T) {
			programFile := exactProgramSourceFile(fallback, target.Path)
			directFile := parser.parse(target)
			if programFile == nil || directFile == nil {
				t.Fatalf("missing source: Program=%v direct=%v", programFile != nil, directFile != nil)
			}
			if programFile.Text() != directFile.Text() || programFile.ScriptKind != directFile.ScriptKind {
				t.Fatal("direct parser changed source text or ScriptKind")
			}
			if (programFile.ExternalModuleIndicator == nil) != (directFile.ExternalModuleIndicator == nil) {
				t.Fatal("direct parser changed external-module detection")
			}
			if programMetadata := fallback.GetSourceFileMetaData(programFile.Path()); programMetadata != parser.sourceFileMetaData(directFile.FileName()) {
				t.Fatalf("direct parser changed source metadata: Program=%+v direct=%+v", programMetadata, parser.sourceFileMetaData(directFile.FileName()))
			}

			programDiagnostics := fallback.GetSyntacticDiagnostics(context.Background(), programFile)
			directDiagnostics := parser.syntacticDiagnostics(directFile)
			if len(programDiagnostics) != len(directDiagnostics) {
				t.Fatalf("diagnostic count = %d, want %d", len(directDiagnostics), len(programDiagnostics))
			}
			if filepath.Base(target.Path) == "a.js" && len(programDiagnostics) == 0 {
				t.Fatal("unchecked JS fixture did not exercise additional syntactic diagnostics")
			}
			for i := range programDiagnostics {
				left, right := programDiagnostics[i], directDiagnostics[i]
				if left.Code() != right.Code() || left.Pos() != right.Pos() || left.End() != right.End() || left.String() != right.String() {
					t.Fatalf("diagnostic %d differs: Program=%v direct=%v", i, left, right)
				}
			}

			binder.BindSourceFile(directFile)
			programAST, programErr := api.EncodeAST(programFile, target.Path)
			directAST, directErr := api.EncodeAST(directFile, target.Path)
			if programErr != nil || directErr != nil {
				t.Fatalf("EncodeAST: Program=%v direct=%v", programErr, directErr)
			}
			if !bytes.Equal(programAST, directAST) {
				t.Fatal("direct parser produced a different encoded AST")
			}
		})
	}
}

func TestFallbackSourceParser_MatchesPackageTypeInNodeModules(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"node_modules/pkg/package.json": "{\"type\":\"module\"}",
		"node_modules/pkg/index.js":     "await value;\n",
	})
	target := rulelessGapTestPlan(dir, "node_modules/pkg/index.js").Targets[0]
	buildContext := rulelessGapTestBuildContext()
	fallback, err := createFallbackProgram([]string{target.Path}, true, dir, buildContext)
	if err != nil {
		t.Fatalf("createFallbackProgram: %v", err)
	}
	programFile := exactProgramSourceFile(fallback, target.Path)
	parser := newFallbackSourceParser(dir, buildContext)
	directFile := parser.parse(target)
	if programFile == nil || directFile == nil {
		t.Fatal("Program and direct parser must contain the node_modules target")
	}
	if (programFile.ExternalModuleIndicator == nil) != (directFile.ExternalModuleIndicator == nil) {
		t.Fatal("package type changed external-module detection")
	}
	if programMetadata := fallback.GetSourceFileMetaData(programFile.Path()); programMetadata != parser.sourceFileMetaData(directFile.FileName()) {
		t.Fatalf("package metadata differs: Program=%+v direct=%+v", programMetadata, parser.sourceFileMetaData(directFile.FileName()))
	}
}

func TestStandaloneGapParser_IsolatesCaseFoldedPackageScopes(t *testing.T) {
	configDir := "/repo"
	upper := "/repo/node_modules/Pkg/index.js"
	lower := "/repo/node_modules/pkg/index.js"
	files := map[string]string{
		upper:                                 "const upper = ;\n",
		lower:                                 "const lower = ;\n",
		"/repo/node_modules/Pkg/package.json": `{"type":"module"}`,
		"/repo/node_modules/pkg/package.json": `{"type":"commonjs"}`,
	}
	directories := map[string]struct{}{
		"/":                      {},
		"/repo":                  {},
		"/repo/node_modules":     {},
		"/repo/node_modules/Pkg": {},
		"/repo/node_modules/pkg": {},
	}
	newFS := func() *rulelessGapExactCaseFS {
		return &rulelessGapExactCaseFS{
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
	legacyDiagnostics, _ := collectTargetSyntacticDiagnostics(
		legacy.Programs,
		legacy.TargetsByProgram,
		buildTypeCheckSkipMask(legacy.Programs),
		false,
		false,
	)

	directContext := utils.NewProgramBuildContext(newFS())
	direct, _, err := bindCLILintTargetPlan(
		lintProgramSet{},
		plan,
		configDir,
		directContext,
		true,
		lintConfigResolverOptions{Config: rslintconfig.RslintConfig{{}}, CurrentDirectory: configDir},
	)
	if err != nil {
		t.Fatalf("bindCLILintTargetPlan: %v", err)
	}
	if len(direct.standaloneGapGroups) != 2 {
		t.Fatalf("case-folded roots must remain isolated, got %d groups", len(direct.standaloneGapGroups))
	}
	directDiagnostics, _, err := collectStandaloneGapSyntacticDiagnostics(
		direct.standaloneGapGroups,
		configDir,
		directContext,
		true,
	)
	if err != nil {
		t.Fatalf("collectStandaloneGapSyntacticDiagnostics: %v", err)
	}

	externalModuleByPath := func(diagnostics []rule.RuleDiagnostic) map[string]bool {
		result := make(map[string]bool, len(diagnostics))
		for _, diagnostic := range diagnostics {
			sourceFile, ok := diagnostic.SourceFile.(*ast.SourceFile)
			if !ok {
				t.Fatalf("diagnostic source for %q is %T, want *ast.SourceFile", diagnostic.FilePath, diagnostic.SourceFile)
			}
			result[diagnostic.FilePath] = sourceFile.ExternalModuleIndicator != nil
		}
		return result
	}
	legacyExternal := externalModuleByPath(legacyDiagnostics)
	directExternal := externalModuleByPath(directDiagnostics)
	if len(legacyDiagnostics) != 2 || len(directDiagnostics) != 2 {
		t.Fatalf("expected one syntax diagnostic per exact-case target: legacy=%d direct=%d", len(legacyDiagnostics), len(directDiagnostics))
	}
	if !legacyExternal[upper] || legacyExternal[lower] {
		t.Fatalf("fixture did not distinguish package scopes: %v", legacyExternal)
	}
	if directExternal[upper] != legacyExternal[upper] || directExternal[lower] != legacyExternal[lower] {
		t.Fatalf("standalone package-scope semantics differ: legacy=%v direct=%v", legacyExternal, directExternal)
	}
}

func TestStandaloneGapSyntaxDeduplicatesAgainstNonGoverningTypeCheckProgram(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json":   `{"include":["child/target.ts"]}`,
		"child/target.ts": "let value: ;\n",
	})
	rootDir = tspath.NormalizePath(rootDir)
	childDir = tspath.NormalizePath(childDir)
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:  projectConfig("./tsconfig.json"),
		childDir: {{}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := utils.NewProgramBuildContext(fsys)
	set, err := createProgramSetForConfigs(configMap, true, buildContext)
	if err != nil {
		t.Fatalf("createProgramSetForConfigs: %v", err)
	}
	targetPath := tspath.ResolvePath(childDir, "target.ts")
	plan := lintTargetPlan{Targets: []resolvedLintTarget{testLintTarget(fsys, childDir, targetPath)}}
	binding, _, err := bindCLILintTargetPlan(
		set,
		plan,
		rootDir,
		buildContext,
		true,
		lintConfigResolverOptions{ConfigMap: configMap, CurrentDirectory: rootDir},
	)
	if err != nil {
		t.Fatalf("bindCLILintTargetPlan: %v", err)
	}
	if len(binding.Programs) != 1 || len(binding.standaloneGapGroups) != 1 {
		t.Fatalf("fixture must keep the child-owned target standalone beside the parent Program: programs=%d groups=%d", len(binding.Programs), len(binding.standaloneGapGroups))
	}

	diagnostics, _, err := collectStandaloneGapSyntacticDiagnostics(
		binding.standaloneGapGroups,
		rootDir,
		buildContext,
		true,
	)
	if err != nil {
		t.Fatalf("collectStandaloneGapSyntacticDiagnostics: %v", err)
	}
	diagnostics = append(
		diagnostics,
		collectProgramTypeDiagnostics(t, binding.Programs, buildTypeCheckSkipMask(binding.Programs), binding.TypeInfoFiles)...,
	)
	if len(diagnostics) < 2 {
		t.Fatalf("fixture must exercise standalone syntax and parent type-check diagnostics: %+v", diagnostics)
	}
	diagnostics = deduplicateTypeScriptDiagnostics(diagnostics, fsys, preferredCallerTargetPaths(plan))
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "TypeScript(TS1110)" || diagnostics[0].FilePath != targetPath {
		t.Fatalf("expected one caller-path TS1110 after cross-phase dedupe, got %+v", diagnostics)
	}
}

func TestCollectStandaloneGapSyntacticDiagnostics_PreservesCountAndStableMissingError(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	nodeModuleFile := filepath.Join(dir, "node_modules", "pkg", "index.ts")
	if err := os.MkdirAll(filepath.Dir(nodeModuleFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nodeModuleFile, []byte("export const value = ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nodeModuleTarget := resolvedLintTarget{
		Path:           tspath.NormalizePath(nodeModuleFile),
		CanonicalPath:  tspath.NormalizePath(nodeModuleFile),
		OwnerConfigDir: dir,
	}
	diagnostics, count, err := collectStandaloneGapSyntacticDiagnostics(
		[][]resolvedLintTarget{{nodeModuleTarget}},
		dir,
		rulelessGapTestBuildContext(),
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 || len(diagnostics) == 0 {
		t.Fatalf("excluded target count/diagnostics = %d/%d, want 0/non-zero", count, len(diagnostics))
	}

	missingA := resolvedLintTarget{Path: tspath.NormalizePath(filepath.Join(dir, "missing-a.ts"))}
	missingB := resolvedLintTarget{Path: tspath.NormalizePath(filepath.Join(dir, "missing-b.ts"))}
	_, _, err = collectStandaloneGapSyntacticDiagnostics(
		[][]resolvedLintTarget{{missingA, missingB}},
		dir,
		rulelessGapTestBuildContext(),
		false,
	)
	if err == nil || !strings.Contains(err.Error(), missingA.Path) {
		t.Fatalf("parallel parse must return the first stable missing target, got %v", err)
	}
}
