package loader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	programLoader "github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type staticPlanProvider struct {
	plan linter.GenerationPlan
}

func (p staticPlanProvider) PlanGeneration(context.Context, linter.SourceSnapshot) (linter.GenerationPlan, error) {
	return p.plan, nil
}

func TestGenerationLoadsPartialStreamingWithEmptyCoupledUnits(t *testing.T) {
	root := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
	files := map[string]string{
		"alias/tsconfig.json":           `{"files":["real.ts"],"compilerOptions":{"types":[]}}`,
		"alias/real.ts":                 `export const ok: number = 1;`,
		"unused-fallback/tsconfig.json": `{"files":["unused.ts"],"compilerOptions":{"types":[]}}`,
		"unused-fallback/unused.ts":     `export const unused = 0;`,
		"ind/tsconfig.json":             `{"files":["ind.ts"],"compilerOptions":{"types":[]}}`,
		"ind/ind.ts":                    `export const bad: number = "wrong";`,
	}
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	linkPath := tspath.ResolvePath(root, "alias/link.ts")
	realPath := tspath.ResolvePath(root, "alias/real.ts")
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	indPath := tspath.ResolvePath(root, "ind/ind.ts")
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	linkTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: linkPath, CanonicalPath: realPath},
		ConfigDirectory: root,
	}
	indTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: indPath, CanonicalPath: indPath},
		ConfigDirectory: root,
	}
	aliasDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"alias/tsconfig.json", "unused-fallback/tsconfig.json"},
		BaseDirectory: root,
	}
	indDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"ind/tsconfig.json"},
		BaseDirectory: root,
	}
	request := programLoader.ProjectBuildRequest{
		Targets: target.Plan{Files: []target.File{linkTarget, indTarget}},
		Policies: map[target.File]rslintconfig.ProjectPolicy{
			linkTarget: {ExplicitProject: aliasDecl},
			indTarget:  {ExplicitProject: indDecl},
		},
		Scope:                        programLoader.ActiveOwners,
		IncludeProgramRootIdentities: true,
	}

	// Emits 4 loads: coupled alias, unused empty fallback, independent ind, empty trailing unbound.
	for _, mode := range []linter.ProgramExecution{linter.ProgramPhases, linter.ProgramComplete} {
		session := programLoader.NewSession(fsys)
		loads, err := session.PlanLoads(request, root, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(loads) != 4 {
			t.Fatalf("expected 4 loads (alias + empty fallback + independent + empty unbound), got %d", len(loads))
		}
		genLoads := GenerationLoads(loads, session.FS(), func(binding programLoader.LoadResult, generationFS vfs.FS) linter.Generation {
			return linter.Generation{
				Native: linter.NativeGeneration{
					Programs:         binding.Programs,
					TargetsByProgram: binding.TargetsByProgram,
					TypeCheck:        true,
					Cwd:              root,
					RulesForFile:     func(*ast.SourceFile) []rule.ConfiguredRule { return nil },
				},
				Target: linter.TargetProjection{
					Path: func(path string) string {
						if file, ok := target.LookupSourceTarget(binding.LintTargetBySourcePath, path, generationFS); ok {
							return file.Path
						}
						return path
					},
				},
				ProgramRootIdentities: binding.ProgramRootIdentities,
			}
		})
		res, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
			staticPlanProvider{plan: linter.GenerationPlan{Loads: genLoads}},
			linter.ObservationPolicy{
				Demand:           linter.ArtifactDemand{LintedFiles: true},
				ProgramExecution: mode,
			},
			nil,
		))
		if err != nil {
			t.Fatalf("mode=%d RunPipeline: %v", mode, err)
		}
		if res.Observation.Native.Lint.LintedFileCount != 2 {
			t.Fatalf("mode=%d LintedFileCount=%d, want 2", mode, res.Observation.Native.Lint.LintedFileCount)
		}
		gotFiles := make([]string, len(res.Observation.Native.Files))
		for i, f := range res.Observation.Native.Files {
			gotFiles[i] = f.Path
		}
		wantFiles := []string{linkPath, indPath}
		if !slices.Equal(gotFiles, wantFiles) {
			t.Fatalf("mode=%d linted files=%v, want %v", mode, gotFiles, wantFiles)
		}
		diags := res.Observation.Native.Diagnostics
		if len(diags) != 1 || diags[0].RuleName != "TypeScript(TS2322)" || diags[0].FilePath != indPath {
			t.Fatalf("mode=%d diagnostics=%+v, want one TS2322 at %q", mode, diags, indPath)
		}
		wantRoots := []string{rslintconfig.ExactPathID(realPath), rslintconfig.ExactPathID(indPath)}
		if !slices.Equal(res.Observation.ProgramRootIdentities, wantRoots) {
			t.Fatalf("mode=%d ProgramRootIdentities=%v, want %v", mode, res.Observation.ProgramRootIdentities, wantRoots)
		}
	}
}

func writeGenerationTestFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func reportVarDeclRule() rule.ConfiguredRule {
	return rule.ConfiguredRule{
		Name:     "test/report-var",
		Severity: rule.SeverityError,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindVariableDeclaration: func(node *ast.Node) {
					ctx.ReportNode(node, rule.RuleMessage{Id: "found", Description: "visited"})
				},
			}
		},
	}
}

func runPlannedGenerationPipeline(
	t *testing.T,
	fsys vfs.FS,
	root string,
	request programLoader.ProjectBuildRequest,
	wantLoadCount int,
	mode linter.ProgramExecution,
	typeCheck bool,
	rules []rule.ConfiguredRule,
) linter.PipelineResult {
	t.Helper()
	session := programLoader.NewSession(fsys)
	loads, err := session.PlanLoads(request, root, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(loads) != wantLoadCount {
		t.Fatalf("mode=%d expected %d loads, got %d", mode, wantLoadCount, len(loads))
	}
	genLoads := GenerationLoads(loads, session.FS(), func(binding programLoader.LoadResult, generationFS vfs.FS) linter.Generation {
		return linter.Generation{
			Native: linter.NativeGeneration{
				Programs:         binding.Programs,
				TargetsByProgram: binding.TargetsByProgram,
				TypeCheck:        typeCheck,
				Cwd:              root,
				RulesForFile:     func(*ast.SourceFile) []rule.ConfiguredRule { return rules },
			},
			Target: linter.TargetProjection{
				Path: func(path string) string {
					if file, ok := target.LookupSourceTarget(binding.LintTargetBySourcePath, path, generationFS); ok {
						return file.Path
					}
					return path
				},
			},
			ProgramRootIdentities: binding.ProgramRootIdentities,
		}
	})
	res, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
		staticPlanProvider{plan: linter.GenerationPlan{Loads: genLoads}},
		linter.ObservationPolicy{
			Demand:           linter.ArtifactDemand{LintedFiles: true},
			ProgramExecution: mode,
		},
		nil,
	))
	if err != nil {
		t.Fatalf("mode=%d RunPipeline: %v", mode, err)
	}
	return res
}

func TestGenerationLoadsCoupledTargetAndUnboundSymlinkAlias(t *testing.T) {
	root := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
	writeGenerationTestFiles(t, root, map[string]string{
		"alias/tsconfig.json": `{"files":["real.ts","rejected.js"],"compilerOptions":{"types":[]}}`,
		"alias/real.ts":       `export const realValue = 1;`,
		"alias/rejected.js":   `export const rejectedValue = 2;`,
		"ind/tsconfig.json":   `{"files":["ind.ts"],"compilerOptions":{"types":[]}}`,
		"ind/ind.ts":          `export const indValue = 3;`,
	})
	realPath := tspath.ResolvePath(root, "alias/real.ts")
	linkPath := tspath.ResolvePath(root, "alias/link.ts")
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	rejectedPath := tspath.ResolvePath(root, "alias/rejected.js")
	indPath := tspath.ResolvePath(root, "ind/ind.ts")
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	realTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: realPath, CanonicalPath: realPath},
		ConfigDirectory: root,
	}
	rejectedTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: rejectedPath, CanonicalPath: rejectedPath},
		ConfigDirectory: root,
	}
	linkTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: linkPath, CanonicalPath: realPath},
		ConfigDirectory: root,
	}
	indTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: indPath, CanonicalPath: indPath},
		ConfigDirectory: root,
	}
	aliasDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"alias/tsconfig.json"},
		BaseDirectory: root,
	}
	indDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"ind/tsconfig.json"},
		BaseDirectory: root,
	}
	request := programLoader.ProjectBuildRequest{
		Targets: target.Plan{Files: []target.File{realTarget, rejectedTarget, linkTarget, indTarget}},
		Policies: map[target.File]rslintconfig.ProjectPolicy{
			realTarget:     {ExplicitProject: aliasDecl},
			rejectedTarget: {ExplicitProject: aliasDecl},
			linkTarget:     {ProjectDisabled: true},
			indTarget:      {ExplicitProject: indDecl},
		},
		Scope: programLoader.ActiveOwners,
	}
	wantPaths := []string{realPath, indPath, linkPath, rejectedPath}
	rules := []rule.ConfiguredRule{reportVarDeclRule()}
	for _, mode := range []linter.ProgramExecution{linter.ProgramPhases, linter.ProgramComplete} {
		res := runPlannedGenerationPipeline(t, fsys, root, request, 3, mode, false, rules)
		gotFiles := make([]string, len(res.Observation.Native.Files))
		for i, f := range res.Observation.Native.Files {
			gotFiles[i] = f.Path
		}
		if !slices.Equal(gotFiles, wantPaths) {
			t.Fatalf("mode=%d linted files=%v, want %v", mode, gotFiles, wantPaths)
		}
		gotDiagPaths := make([]string, len(res.Observation.Native.Diagnostics))
		for i, d := range res.Observation.Native.Diagnostics {
			gotDiagPaths[i] = d.FilePath
		}
		if !slices.Equal(gotDiagPaths, wantPaths) {
			t.Fatalf("mode=%d diagnostic file paths=%v, want %v", mode, gotDiagPaths, wantPaths)
		}
	}
}

func TestGenerationLoadsDistinctSymlinkAliasesAcrossCoupledProjects(t *testing.T) {
	root := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
	writeGenerationTestFiles(t, root, map[string]string{
		"shared/real.ts":      `export const sharedValue = 1;`,
		"proj1/tsconfig.json": `{"files":["../shared/real.ts"],"compilerOptions":{"types":[]}}`,
		"proj2/tsconfig.json": `{"files":["../shared/real.ts"],"compilerOptions":{"types":[]}}`,
		"ind/tsconfig.json":   `{"files":["ind.ts"],"compilerOptions":{"types":[]}}`,
		"ind/ind.ts":          `export const indValue = 2;`,
	})
	realPath := tspath.ResolvePath(root, "shared/real.ts")
	link1Path := tspath.ResolvePath(root, "shared/link1.ts")
	link2Path := tspath.ResolvePath(root, "shared/link2.ts")
	if err := os.Symlink(realPath, link1Path); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.Symlink(realPath, link2Path); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	indPath := tspath.ResolvePath(root, "ind/ind.ts")
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	link1Target := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: link1Path, CanonicalPath: realPath},
		ConfigDirectory: root,
	}
	link2Target := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: link2Path, CanonicalPath: realPath},
		ConfigDirectory: root,
	}
	indTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: indPath, CanonicalPath: indPath},
		ConfigDirectory: root,
	}
	proj1Decl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"proj1/tsconfig.json"},
		BaseDirectory: root,
	}
	proj2Decl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"proj2/tsconfig.json"},
		BaseDirectory: root,
	}
	indDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"ind/tsconfig.json"},
		BaseDirectory: root,
	}
	request := programLoader.ProjectBuildRequest{
		Targets: target.Plan{Files: []target.File{link1Target, link2Target, indTarget}},
		Policies: map[target.File]rslintconfig.ProjectPolicy{
			link1Target: {ExplicitProject: proj1Decl},
			link2Target: {ExplicitProject: proj2Decl},
			indTarget:   {ExplicitProject: indDecl},
		},
		Scope:                        programLoader.ActiveOwners,
		IncludeProgramRootIdentities: true,
	}
	wantPaths := []string{link1Path, link2Path, indPath}
	wantRoots := []string{rslintconfig.ExactPathID(realPath), rslintconfig.ExactPathID(indPath)}
	rules := []rule.ConfiguredRule{reportVarDeclRule()}
	for _, mode := range []linter.ProgramExecution{linter.ProgramPhases, linter.ProgramComplete} {
		res := runPlannedGenerationPipeline(t, fsys, root, request, 4, mode, false, rules)
		gotFiles := make([]string, len(res.Observation.Native.Files))
		for i, f := range res.Observation.Native.Files {
			gotFiles[i] = f.Path
		}
		if !slices.Equal(gotFiles, wantPaths) {
			t.Fatalf("mode=%d linted files=%v, want %v", mode, gotFiles, wantPaths)
		}
		gotDiagPaths := make([]string, len(res.Observation.Native.Diagnostics))
		for i, d := range res.Observation.Native.Diagnostics {
			gotDiagPaths[i] = d.FilePath
		}
		if !slices.Equal(gotDiagPaths, wantPaths) {
			t.Fatalf("mode=%d diagnostic file paths=%v, want %v", mode, gotDiagPaths, wantPaths)
		}
		if !slices.Equal(res.Observation.ProgramRootIdentities, wantRoots) {
			t.Fatalf("mode=%d ProgramRootIdentities=%v, want %v", mode, res.Observation.ProgramRootIdentities, wantRoots)
		}
	}
}

func TestGenerationLoadsDeduplicatesSharedProgramRootIdentities(t *testing.T) {
	root := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
	writeGenerationTestFiles(t, root, map[string]string{
		"shared-a.ts":              `export const sharedA = 1;`,
		"shared-b.ts":              `export const sharedB = 2;`,
		"ind-before/tsconfig.json": `{"files":["ind-before.ts","../shared-a.ts"],"compilerOptions":{"types":[]}}`,
		"ind-before/ind-before.ts": `export const beforeValue = 3;`,
		"coupled/tsconfig.json":    `{"files":["../shared-a.ts","../shared-b.ts","target.js"],"compilerOptions":{"types":[]}}`,
		"coupled/target.js":        `export const jsValue = 4;`,
		"ind-after/tsconfig.json":  `{"files":["../shared-b.ts","../shared-a.ts","ind-after.ts"],"compilerOptions":{"types":[]}}`,
		"ind-after/ind-after.ts":   `export const afterValue = 5;`,
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	beforePath := tspath.ResolvePath(root, "ind-before/ind-before.ts")
	jsPath := tspath.ResolvePath(root, "coupled/target.js")
	afterPath := tspath.ResolvePath(root, "ind-after/ind-after.ts")
	beforeTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: beforePath, CanonicalPath: beforePath},
		ConfigDirectory: root,
	}
	jsTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: jsPath, CanonicalPath: jsPath},
		ConfigDirectory: root,
	}
	afterTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: afterPath, CanonicalPath: afterPath},
		ConfigDirectory: root,
	}
	beforeDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"ind-before/tsconfig.json"},
		BaseDirectory: root,
	}
	coupledDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"coupled/tsconfig.json"},
		BaseDirectory: root,
	}
	afterDecl := &rslintconfig.ProjectDeclaration{
		Patterns:      rslintconfig.ProjectPaths{"ind-after/tsconfig.json"},
		BaseDirectory: root,
	}
	config := rslintconfig.RslintConfig{{
		LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
			Project: rslintconfig.ProjectPaths{
				"ind-before/tsconfig.json",
				"coupled/tsconfig.json",
				"ind-after/tsconfig.json",
			},
		}},
	}}
	for _, tc := range []struct {
		name  string
		scope programLoader.ProjectScope
	}{
		{"ActiveOwners", programLoader.ActiveOwners},
		{"AllDeclared", programLoader.AllDeclared},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := programLoader.ProjectBuildRequest{
				Configs: map[string]rslintconfig.RslintConfig{root: config},
				Targets: target.Plan{Files: []target.File{beforeTarget, jsTarget, afterTarget}},
				Policies: map[target.File]rslintconfig.ProjectPolicy{
					beforeTarget: {ExplicitProject: beforeDecl},
					jsTarget:     {ExplicitProject: coupledDecl},
					afterTarget:  {ExplicitProject: afterDecl},
				},
				Scope:                        tc.scope,
				IncludeProgramRootIdentities: true,
			}
			materializedSet, err := programLoader.NewSession(fsys).BuildProjects(request)
			if err != nil {
				t.Fatalf("materialized BuildProjects: %v", err)
			}
			wantRoots := materializedSet.ProgramRootIdentities()
			sharedAID := rslintconfig.ExactPathID(tspath.ResolvePath(root, "shared-a.ts"))
			sharedBID := rslintconfig.ExactPathID(tspath.ResolvePath(root, "shared-b.ts"))
			if !slices.Contains(wantRoots, sharedAID) || !slices.Contains(wantRoots, sharedBID) {
				t.Fatalf("expected shared roots in materialized ProgramRootIdentities, got %v", wantRoots)
			}
			for _, mode := range []linter.ProgramExecution{linter.ProgramPhases, linter.ProgramComplete} {
				res := runPlannedGenerationPipeline(t, fsys, root, request, 4, mode, true, nil)
				if !slices.Equal(res.Observation.ProgramRootIdentities, wantRoots) {
					t.Fatalf("mode=%d ProgramRootIdentities mismatch:\ngot:  %v\nwant: %v",
						mode, res.Observation.ProgramRootIdentities, wantRoots)
				}
			}
		})
	}
}

func TestGenerationLoadsEmptyUnitsSkipDiagnosticBufferAlloc(t *testing.T) {
	root := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
	const unusedCount = 64
	files := map[string]string{
		"alias/tsconfig.json": `{"files":["real.ts"],"compilerOptions":{"types":[]}}`,
		"alias/real.ts":       `export const ok = 1;`,
		"ind/tsconfig.json":   `{"files":["ind.ts"],"compilerOptions":{"types":[]}}`,
		"ind/ind.ts":          `export const ind = 2;`,
	}
	fallbackPatterns := make(rslintconfig.ProjectPaths, 0, unusedCount+1)
	fallbackPatterns = append(fallbackPatterns, "alias/tsconfig.json")
	for i := range unusedCount {
		dir := fmt.Sprintf("unused-%02d", i)
		files[dir+"/tsconfig.json"] = `{"files":["unused.ts"],"compilerOptions":{"types":[]}}`
		files[dir+"/unused.ts"] = `export const unused = 0;`
		fallbackPatterns = append(fallbackPatterns, dir+"/tsconfig.json")
	}
	writeGenerationTestFiles(t, root, files)

	realPath := tspath.ResolvePath(root, "alias/real.ts")
	linkPath := tspath.ResolvePath(root, "alias/link.ts")
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	indPath := tspath.ResolvePath(root, "ind/ind.ts")
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	linkTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: linkPath, CanonicalPath: realPath},
		ConfigDirectory: root,
	}
	indTarget := target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: indPath, CanonicalPath: indPath},
		ConfigDirectory: root,
	}
	request := programLoader.ProjectBuildRequest{
		Targets: target.Plan{Files: []target.File{linkTarget, indTarget}},
		Policies: map[target.File]rslintconfig.ProjectPolicy{
			linkTarget: {ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: fallbackPatterns, BaseDirectory: root}},
			indTarget:  {ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"ind/tsconfig.json"}, BaseDirectory: root}},
		},
		Scope:          programLoader.ActiveOwners,
		SingleThreaded: true,
	}
	// 1 coupled alias + 64 unused fallback candidates + 1 independent + 1 empty unbound = 67 loads.
	res := runPlannedGenerationPipeline(t, fsys, root, request, unusedCount+3, linter.ProgramPhases, false, []rule.ConfiguredRule{reportVarDeclRule()})
	if res.Observation.Native.Lint.LintedFileCount != 2 {
		t.Fatalf("LintedFileCount = %d, want 2", res.Observation.Native.Lint.LintedFileCount)
	}
	if len(res.Observation.Native.Diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(res.Observation.Native.Diagnostics))
	}
}

func TestGenerationLoadsSingleThreadedStopsOnFirstLoadErrorBeforePlugins(t *testing.T) {
	root := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
	writeGenerationTestFiles(t, root, map[string]string{
		"bad.json":  `{"files":["target.js"],"compilerOptions":{"types":[]}}`,
		"good.json": `{"files":["good.ts"],"compilerOptions":{"types":[]}}`,
		"target.js": `export const bad = 1;`,
		"good.ts":   `export const good = 2;`,
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	badPath := tspath.ResolvePath(root, "target.js")
	goodPath := tspath.ResolvePath(root, "good.ts")
	badTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: badPath, CanonicalPath: badPath}, ConfigDirectory: root}
	goodTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: goodPath, CanonicalPath: goodPath}, ConfigDirectory: root}
	goodOnlyReq := programLoader.ProjectBuildRequest{
		Targets: target.Plan{Files: []target.File{goodTarget}},
		Policies: map[target.File]rslintconfig.ProjectPolicy{
			goodTarget: {ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"good.json"}, BaseDirectory: root}},
		},
		Scope:          programLoader.Targeted,
		SingleThreaded: true,
	}

	// Positive control: prove targetedFallbackProvider actually dispatches plugins on a healthy load.
	var controlDispatches atomic.Int32
	if _, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
		&targetedFallbackProvider{fsys: fsys, root: root, request: goodOnlyReq},
		linter.ObservationPolicy{},
		linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			controlDispatches.Add(1)
			return &linter.EslintPluginLintResult{}, nil
		}),
	)); err != nil {
		t.Fatalf("positive control RunPipeline failed: %v", err)
	}
	if got := controlDispatches.Load(); got != 1 {
		t.Fatalf("positive control plugin dispatches = %d, want 1", got)
	}

	// Single-threaded pipeline stops on first unit error before acquiring or dispatching unit 1.
	wantErr := errors.New("load failed")
	var unit1Acquired atomic.Int32
	var pluginDispatches atomic.Int32
	goodProvider := &targetedFallbackProvider{fsys: fsys, root: root, request: goodOnlyReq}
	goodPlan, err := goodProvider.PlanGeneration(context.Background(), linter.SourceSnapshot{})
	if err != nil || len(goodPlan.Loads) != 1 {
		t.Fatalf("goodProvider PlanGeneration: loads=%d err=%v", len(goodPlan.Loads), err)
	}
	loads := []linter.GenerationLoad{
		{
			TargetCount: 1,
			Acquire: func(context.Context) (linter.Generation, linter.ReleaseFunc, error) {
				return linter.Generation{}, nil, wantErr
			},
		},
		{
			TargetCount: 1,
			Acquire: func(ctx context.Context) (linter.Generation, linter.ReleaseFunc, error) {
				unit1Acquired.Add(1)
				return goodPlan.Loads[0].Acquire(ctx)
			},
		},
	}
	dispatch := linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		pluginDispatches.Add(1)
		return &linter.EslintPluginLintResult{}, nil
	})
	_, err = linter.RunPipeline(context.Background(), linter.NewLintRequest(
		staticPlanProvider{plan: linter.GenerationPlan{Loads: loads, SingleThreaded: true}},
		linter.ObservationPolicy{},
		dispatch,
	))
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunPipeline error = %v, want %v", err, wantErr)
	}
	if got := unit1Acquired.Load(); got != 0 {
		t.Fatalf("single-threaded pipeline acquired %d subsequent units after error, want 0", got)
	}
	if got := pluginDispatches.Load(); got != 0 {
		t.Fatalf("plugin dispatched %d times after load error, want 0", got)
	}

	// End-to-end Targeted scope with a rejected JS root falls back to whole-binding
	// validation before exposing any loads, so healthy sibling projects never dispatch plugins.
	t.Run("targeted absent root falls back to whole-binding with zero plugin dispatches", func(t *testing.T) {
		request := programLoader.ProjectBuildRequest{
			Targets: target.Plan{Files: []target.File{badTarget, goodTarget}},
			Policies: map[target.File]rslintconfig.ProjectPolicy{
				badTarget:  {ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"bad.json"}, BaseDirectory: root}},
				goodTarget: {ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"good.json"}, BaseDirectory: root}},
			},
			Scope: programLoader.Targeted,
		}
		for _, singleThreaded := range []bool{false, true} {
			request.SingleThreaded = singleThreaded
			var dispatches atomic.Int32
			provider := &targetedFallbackProvider{fsys: fsys, root: root, request: request}
			_, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
				provider,
				linter.ObservationPolicy{},
				linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
					dispatches.Add(1)
					return &linter.EslintPluginLintResult{}, nil
				}),
			))
			if err == nil || !strings.Contains(err.Error(), "was absent from its TypeScript Program") {
				t.Fatalf("singleThreaded=%v: expected targeted absent-root error, got %v", singleThreaded, err)
			}
			if got := dispatches.Load(); got != 0 {
				t.Fatalf("singleThreaded=%v: plugin dispatched %d times, want 0", singleThreaded, got)
			}
		}
	})

	// Broad lint (ActiveOwners) and program-wide loading (AllDeclared) with a
	// healthy TS project, a legacy .js/allowJs:false project, and a projectService
	// project with disableSourceOfProjectReferenceRedirect:true whose referenced
	// composite project also lists target.ts (with missing .d.ts output) must
	// abort partial streaming in PlanLoads and fail whole-binding validation
	// before dispatching any plugin.
	t.Run("service disabled-source-redirect absent root falls back to whole-binding with zero plugin dispatches", func(t *testing.T) {
		dir := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
		writeGenerationTestFiles(t, dir, map[string]string{
			"healthy/tsconfig.json": `{"files":["healthy.ts"],"compilerOptions":{"types":[]}}`,
			"healthy/healthy.ts":    `export const healthy = 1;`,
			"legacy/tsconfig.json":  `{"files":["legacy.js"],"compilerOptions":{"types":[],"allowJs":false}}`,
			"legacy/legacy.js":      `export const legacy = 2;`,
			"service/tsconfig.json": `{"compilerOptions":{"noLib":true,"disableSourceOfProjectReferenceRedirect":true},"files":["target.ts"],"references":[{"path":"./leaf.json"}]}`,
			"service/leaf.json":     `{"compilerOptions":{"noLib":true,"composite":true,"outDir":"dist"},"files":["target.ts"]}`,
			"service/target.ts":     `export const value = 3;`,
		})
		testFS := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		healthyPath := tspath.ResolvePath(dir, "healthy/healthy.ts")
		legacyPath := tspath.ResolvePath(dir, "legacy/legacy.js")
		targetPath := tspath.ResolvePath(dir, "service/target.ts")
		serviceConfigPath := tspath.ResolvePath(dir, "service/tsconfig.json")
		healthyTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: healthyPath, CanonicalPath: healthyPath}, ConfigDirectory: dir}
		legacyTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: legacyPath, CanonicalPath: legacyPath}, ConfigDirectory: dir}
		serviceTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: targetPath, CanonicalPath: targetPath}, ConfigDirectory: dir}
		healthyDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"healthy/tsconfig.json"}, BaseDirectory: dir}
		legacyDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"legacy/tsconfig.json"}, BaseDirectory: dir}
		config := rslintconfig.RslintConfig{
			{Files: []string{"healthy/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: healthyDecl.Patterns}}},
			{Files: []string{"legacy/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: legacyDecl.Patterns}}},
			{Files: []string{"service/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}},
		}
		wantErrMsg := fmt.Sprintf("project root %q from %q was absent from its TypeScript Program", targetPath, serviceConfigPath)

		for _, tc := range []struct {
			name  string
			scope programLoader.ProjectScope
		}{
			{"ActiveOwners", programLoader.ActiveOwners},
			{"AllDeclared", programLoader.AllDeclared},
		} {
			t.Run(tc.name, func(t *testing.T) {
				for _, singleThreaded := range []bool{true, false} {
					request := programLoader.ProjectBuildRequest{
						Configs: map[string]rslintconfig.RslintConfig{dir: config},
						Targets: target.Plan{Files: []target.File{healthyTarget, legacyTarget, serviceTarget}},
						Policies: map[target.File]rslintconfig.ProjectPolicy{
							healthyTarget: {ExplicitProject: healthyDecl},
							legacyTarget:  {ExplicitProject: legacyDecl},
							serviceTarget: {ServiceRootDirectory: dir},
						},
						Scope:          tc.scope,
						SingleThreaded: singleThreaded,
					}
					plannedLoads, planErr := programLoader.NewSession(testFS).PlanLoads(request, dir, false)
					if planErr != nil || plannedLoads != nil {
						t.Fatalf("singleThreaded=%v: PlanLoads = (%d loads, %v), want (nil, nil)", singleThreaded, len(plannedLoads), planErr)
					}
					var dispatches atomic.Int32
					provider := &targetedFallbackProvider{fsys: testFS, root: dir, request: request}
					_, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
						provider,
						linter.ObservationPolicy{},
						linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
							dispatches.Add(1)
							return &linter.EslintPluginLintResult{}, nil
						}),
					))
					if err == nil || err.Error() != wantErrMsg {
						t.Fatalf("singleThreaded=%v: RunPipeline error = %v, want %q", singleThreaded, err, wantErrMsg)
					}
					if got := dispatches.Load(); got != 0 {
						t.Fatalf("singleThreaded=%v: plugin dispatched %d times before missing-root error, want 0", singleThreaded, got)
					}
				}
			})
		}
	})

	// Symmetric source-redirect case (default projectService: true): a listed .d.ts
	// root that matches a referenced composite project's output redirects to the
	// referenced .ts source, dropping the .d.ts root from the Program. PlanLoads
	// must fall back to whole-binding validation with zero plugin dispatches.
	t.Run("service default source-redirect .d.ts absent root falls back to whole-binding with zero plugin dispatches", func(t *testing.T) {
		dir := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
		writeGenerationTestFiles(t, dir, map[string]string{
			"healthy/tsconfig.json":  `{"files":["healthy.ts"],"compilerOptions":{"types":[]}}`,
			"healthy/healthy.ts":     `export const healthy = 1;`,
			"legacy/tsconfig.json":   `{"files":["legacy.js"],"compilerOptions":{"types":[],"allowJs":false}}`,
			"legacy/legacy.js":       `export const legacy = 2;`,
			"service/tsconfig.json":  `{"compilerOptions":{"noLib":true},"files":["dist/leaf.d.ts"],"references":[{"path":"./leaf.json"}]}`,
			"service/leaf.json":      `{"compilerOptions":{"noLib":true,"composite":true,"outDir":"dist"},"files":["leaf.ts"]}`,
			"service/leaf.ts":        `export const leaf = 3;`,
			"service/dist/leaf.d.ts": `export declare const leaf: number;`,
		})
		testFS := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		healthyPath := tspath.ResolvePath(dir, "healthy/healthy.ts")
		legacyPath := tspath.ResolvePath(dir, "legacy/legacy.js")
		targetPath := tspath.ResolvePath(dir, "service/dist/leaf.d.ts")
		serviceConfigPath := tspath.ResolvePath(dir, "service/tsconfig.json")
		healthyTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: healthyPath, CanonicalPath: healthyPath}, ConfigDirectory: dir}
		legacyTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: legacyPath, CanonicalPath: legacyPath}, ConfigDirectory: dir}
		serviceTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: targetPath, CanonicalPath: targetPath}, ConfigDirectory: dir}
		healthyDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"healthy/tsconfig.json"}, BaseDirectory: dir}
		legacyDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"legacy/tsconfig.json"}, BaseDirectory: dir}
		config := rslintconfig.RslintConfig{
			{Files: []string{"healthy/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: healthyDecl.Patterns}}},
			{Files: []string{"legacy/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: legacyDecl.Patterns}}},
			{Files: []string{"service/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}},
		}
		wantErrMsg := fmt.Sprintf("project root %q from %q was absent from its TypeScript Program", targetPath, serviceConfigPath)

		for _, tc := range []struct {
			name  string
			scope programLoader.ProjectScope
		}{
			{"ActiveOwners", programLoader.ActiveOwners},
			{"AllDeclared", programLoader.AllDeclared},
		} {
			t.Run(tc.name, func(t *testing.T) {
				for _, singleThreaded := range []bool{true, false} {
					request := programLoader.ProjectBuildRequest{
						Configs: map[string]rslintconfig.RslintConfig{dir: config},
						Targets: target.Plan{Files: []target.File{healthyTarget, legacyTarget, serviceTarget}},
						Policies: map[target.File]rslintconfig.ProjectPolicy{
							healthyTarget: {ExplicitProject: healthyDecl},
							legacyTarget:  {ExplicitProject: legacyDecl},
							serviceTarget: {ServiceRootDirectory: dir},
						},
						Scope:          tc.scope,
						SingleThreaded: singleThreaded,
					}
					plannedLoads, planErr := programLoader.NewSession(testFS).PlanLoads(request, dir, false)
					if planErr != nil || plannedLoads != nil {
						t.Fatalf("singleThreaded=%v: PlanLoads = (%d loads, %v), want (nil, nil)", singleThreaded, len(plannedLoads), planErr)
					}
					var dispatches atomic.Int32
					provider := &targetedFallbackProvider{fsys: testFS, root: dir, request: request}
					_, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
						provider,
						linter.ObservationPolicy{},
						linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
							dispatches.Add(1)
							return &linter.EslintPluginLintResult{}, nil
						}),
					))
					if err == nil || err.Error() != wantErrMsg {
						t.Fatalf("singleThreaded=%v: RunPipeline error = %v, want %q", singleThreaded, err, wantErrMsg)
					}
					if got := dispatches.Load(); got != 0 {
						t.Fatalf("singleThreaded=%v: plugin dispatched %d times before missing-root error, want 0", singleThreaded, got)
					}
				}
			})
		}
	})

	// An unreadable unmatched root (owner == -1) must fail the PlanLoads readability
	// preflight so whole-binding acquisition reports the read error with zero plugin
	// dispatches before healthy sibling projects execute.
	t.Run("unreadable unmatched root falls back to whole-binding with zero plugin dispatches", func(t *testing.T) {
		dir := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
		writeGenerationTestFiles(t, dir, map[string]string{
			"healthy/tsconfig.json": `{"files":["healthy.ts"],"compilerOptions":{"types":[]}}`,
			"healthy/healthy.ts":    `export const healthy = 1;`,
			"legacy/tsconfig.json":  `{"files":["legacy.js"],"compilerOptions":{"types":[],"allowJs":false}}`,
			"legacy/legacy.js":      `export const legacy = 2;`,
			"outside.ts":            `export const outside = 3;`,
		})
		healthyPath := tspath.ResolvePath(dir, "healthy/healthy.ts")
		legacyPath := tspath.ResolvePath(dir, "legacy/legacy.js")
		outsidePath := tspath.ResolvePath(dir, "outside.ts")
		testFS := &unreadablePathFS{
			FS:             bundled.WrapFS(cachedvfs.From(osvfs.FS())),
			unreadablePath: outsidePath,
		}
		healthyTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: healthyPath, CanonicalPath: healthyPath}, ConfigDirectory: dir}
		legacyTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: legacyPath, CanonicalPath: legacyPath}, ConfigDirectory: dir}
		outsideTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: outsidePath, CanonicalPath: outsidePath}, ConfigDirectory: dir}
		healthyDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"healthy/tsconfig.json"}, BaseDirectory: dir}
		legacyDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"legacy/tsconfig.json"}, BaseDirectory: dir}
		config := rslintconfig.RslintConfig{
			{Files: []string{"healthy/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: healthyDecl.Patterns}}},
			{Files: []string{"legacy/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: legacyDecl.Patterns}}},
		}
		wantErrMsg := fmt.Sprintf("program: could not read root %q", outsidePath)

		for _, tc := range []struct {
			name  string
			scope programLoader.ProjectScope
		}{
			{"ActiveOwners", programLoader.ActiveOwners},
			{"AllDeclared", programLoader.AllDeclared},
		} {
			t.Run(tc.name, func(t *testing.T) {
				for _, singleThreaded := range []bool{true, false} {
					request := programLoader.ProjectBuildRequest{
						Configs: map[string]rslintconfig.RslintConfig{dir: config},
						Targets: target.Plan{Files: []target.File{healthyTarget, legacyTarget, outsideTarget}},
						Policies: map[target.File]rslintconfig.ProjectPolicy{
							healthyTarget: {ExplicitProject: healthyDecl},
							legacyTarget:  {ExplicitProject: legacyDecl},
							outsideTarget: {ProjectDisabled: true},
						},
						Scope:          tc.scope,
						SingleThreaded: singleThreaded,
					}
					plannedLoads, planErr := programLoader.NewSession(testFS).PlanLoads(request, dir, false)
					if planErr != nil || plannedLoads != nil {
						t.Fatalf("singleThreaded=%v: PlanLoads = (%d loads, %v), want (nil, nil)", singleThreaded, len(plannedLoads), planErr)
					}
					var dispatches atomic.Int32
					provider := &targetedFallbackProvider{fsys: testFS, root: dir, request: request}
					_, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
						provider,
						linter.ObservationPolicy{},
						linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
							dispatches.Add(1)
							return &linter.EslintPluginLintResult{}, nil
						}),
					))
					if err == nil || err.Error() != wantErrMsg {
						t.Fatalf("singleThreaded=%v: RunPipeline error = %v, want %q", singleThreaded, err, wantErrMsg)
					}
					if got := dispatches.Load(); got != 0 {
						t.Fatalf("singleThreaded=%v: plugin dispatched %d times before unreadable-root error, want 0", singleThreaded, got)
					}
				}
			})
		}
	})

	// Coupled and unmatched roots rejected by SourceOnlyCompilerOptions (e.g.
	// custom extensions like .vue) must fall back to whole-binding validation
	// with zero plugin dispatches before healthy sibling projects execute.
	t.Run("unsupported extension coupled and unmatched roots fall back with zero plugin dispatches", func(t *testing.T) {
		dir := tspath.NormalizePath(osvfs.FS().Realpath(t.TempDir()))
		writeGenerationTestFiles(t, dir, map[string]string{
			"a/tsconfig.json":      `{"files":["a.ts"],"compilerOptions":{"types":[]}}`,
			"a/a.ts":               `export const a = 1;`,
			"b/tsconfig.json":      `{"files":["b.ts"],"compilerOptions":{"types":[]}}`,
			"b/b.ts":               `export const b = 2;`,
			"legacy/tsconfig.json": `{"files":["target.vue"],"compilerOptions":{"types":[]}}`,
			"legacy/target.vue":    `export const vue = 3;`,
			"outside.vue":          `export const outside = 4;`,
		})
		aPath := tspath.ResolvePath(dir, "a/a.ts")
		bPath := tspath.ResolvePath(dir, "b/b.ts")
		legacyPath := tspath.ResolvePath(dir, "legacy/target.vue")
		outsidePath := tspath.ResolvePath(dir, "outside.vue")
		aTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: aPath, CanonicalPath: aPath}, ConfigDirectory: dir}
		bTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: bPath, CanonicalPath: bPath}, ConfigDirectory: dir}
		legacyTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: legacyPath, CanonicalPath: legacyPath}, ConfigDirectory: dir}
		outsideTarget := target.File{PathIdentity: rslintconfig.PathIdentity{Path: outsidePath, CanonicalPath: outsidePath}, ConfigDirectory: dir}
		aDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"a/tsconfig.json"}, BaseDirectory: dir}
		bDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"b/tsconfig.json"}, BaseDirectory: dir}
		legacyDecl := &rslintconfig.ProjectDeclaration{Patterns: rslintconfig.ProjectPaths{"legacy/tsconfig.json"}, BaseDirectory: dir}
		config := rslintconfig.RslintConfig{
			{Files: []string{"a/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: aDecl.Patterns}}},
			{Files: []string{"b/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: bDecl.Patterns}}},
			{Files: []string{"legacy/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: legacyDecl.Patterns}}},
		}
		for _, tc := range []struct {
			name    string
			targets []target.File
			policy  map[target.File]rslintconfig.ProjectPolicy
			wantErr string
		}{
			{
				name:    "coupled",
				targets: []target.File{aTarget, legacyTarget},
				policy: map[target.File]rslintconfig.ProjectPolicy{
					aTarget:      {ExplicitProject: aDecl},
					legacyTarget: {ExplicitProject: legacyDecl},
				},
				wantErr: fmt.Sprintf("program did not contain lint target %q", legacyPath),
			},
			{
				name:    "unmatched",
				targets: []target.File{aTarget, bTarget, outsideTarget},
				policy: map[target.File]rslintconfig.ProjectPolicy{
					aTarget:       {ExplicitProject: aDecl},
					bTarget:       {ExplicitProject: bDecl},
					outsideTarget: {ProjectDisabled: true},
				},
				wantErr: fmt.Sprintf("program did not contain lint target %q", outsidePath),
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				for _, scope := range []programLoader.ProjectScope{programLoader.ActiveOwners, programLoader.AllDeclared} {
					for _, singleThreaded := range []bool{true, false} {
						request := programLoader.ProjectBuildRequest{
							Configs:        map[string]rslintconfig.RslintConfig{dir: config},
							Targets:        target.Plan{Files: tc.targets},
							Policies:       tc.policy,
							Scope:          scope,
							SingleThreaded: singleThreaded,
						}
						plannedLoads, planErr := programLoader.NewSession(fsys).PlanLoads(request, dir, false)
						if planErr != nil || plannedLoads != nil {
							t.Fatalf("scope=%v singleThreaded=%v: PlanLoads = (%d loads, %v), want (nil, nil)", scope, singleThreaded, len(plannedLoads), planErr)
						}
						var dispatches atomic.Int32
						provider := &targetedFallbackProvider{fsys: fsys, root: dir, request: request}
						_, err := linter.RunPipeline(context.Background(), linter.NewLintRequest(
							provider,
							linter.ObservationPolicy{},
							linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
								dispatches.Add(1)
								return &linter.EslintPluginLintResult{}, nil
							}),
						))
						if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
							t.Fatalf("scope=%v singleThreaded=%v: RunPipeline error = %v, want substring %q", scope, singleThreaded, err, tc.wantErr)
						}
						if got := dispatches.Load(); got != 0 {
							t.Fatalf("scope=%v singleThreaded=%v: plugin dispatched %d times, want 0", scope, singleThreaded, got)
						}
					}
				}
			})
		}
	})
}

type unreadablePathFS struct {
	vfs.FS
	unreadablePath string
}

func (f *unreadablePathFS) ReadFile(filePath string) (string, bool) {
	if tspath.NormalizePath(filePath) == f.unreadablePath {
		return "", false
	}
	return f.FS.ReadFile(filePath)
}

type targetedFallbackProvider struct {
	fsys    vfs.FS
	root    string
	request programLoader.ProjectBuildRequest
}

func (p *targetedFallbackProvider) PlanGeneration(context.Context, linter.SourceSnapshot) (linter.GenerationPlan, error) {
	session := programLoader.NewSession(p.fsys)
	loads, err := session.PlanLoads(p.request, p.root, false)
	if err != nil {
		return linter.GenerationPlan{}, err
	}
	pluginRules := []rule.ConfiguredRule{{Name: "plugin/check", Severity: rule.SeverityError, IsEslintPluginRule: true}}
	buildGen := func(binding programLoader.LoadResult, generationFS vfs.FS) linter.Generation {
		return linter.Generation{
			Native: linter.NativeGeneration{
				Programs:         binding.Programs,
				TargetsByProgram: binding.TargetsByProgram,
				Cwd:              p.root,
				RulesForFile:     func(*ast.SourceFile) []rule.ConfiguredRule { return pluginRules },
			},
			Target: linter.TargetProjection{
				Path: func(path string) string {
					if file, ok := target.LookupSourceTarget(binding.LintTargetBySourcePath, path, generationFS); ok {
						return file.Path
					}
					return path
				},
				ReadText: func(_ string, source ast.SourceFileLike) (string, error) {
					return source.Text(), nil
				},
			},
			Plugin: &linter.PluginGeneration{WirePath: func(path string) string { return path }},
		}
	}
	if len(loads) > 0 {
		return linter.GenerationPlan{
			Loads:          GenerationLoads(loads, session.FS(), buildGen),
			SingleThreaded: p.request.SingleThreaded,
		}, nil
	}
	return linter.GenerationPlan{
		SingleThreaded: p.request.SingleThreaded,
		Loads: []linter.GenerationLoad{{
			TargetCount: len(p.request.Targets.Files),
			Acquire: func(context.Context) (linter.Generation, linter.ReleaseFunc, error) {
				set, err := session.BuildProjects(p.request)
				if err != nil {
					return linter.Generation{}, nil, err
				}
				binding, err := session.LoadCLI(set, p.request.Targets, p.root, false)
				if err != nil {
					return linter.Generation{}, nil, err
				}
				return buildGen(binding, session.FS()), nil, nil
			},
		}},
	}, nil
}
