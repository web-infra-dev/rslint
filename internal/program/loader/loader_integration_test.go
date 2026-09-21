package loader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (s *Session) buildProjectsWithOptionsForTest(
	t *testing.T,
	configs map[string]rslintconfig.RslintConfig,
	plan target.Plan,
	scope ProjectScope,
	singleThreaded bool,
) (ProjectSet, error) {
	t.Helper()
	return s.buildProjectsForTest(t, ProjectBuildRequest{
		Configs: configs, Targets: plan, Scope: scope, SingleThreaded: singleThreaded,
	})
}

func (s *Session) buildProjectsForTest(t *testing.T, request ProjectBuildRequest) (ProjectSet, error) {
	t.Helper()
	if len(request.Targets.Files) > 0 && request.Policies == nil {
		pathSpaces := request.Targets.PathSpaces()
		if pathSpaces == nil {
			pathSpaces = rslintconfig.NewPathSpaceSnapshot(request.Configs, s.FS())
		}
		resolver := configLint.NewResolver(configLint.ResolverOptions{
			ConfigsByOwner: request.Configs, FS: s.FS(), PathSpaces: pathSpaces, Catalog: rule.NewCatalog(),
		})
		var err error
		request.Policies, err = resolver.ProjectPolicies(request.Targets.Files, request.SingleThreaded)
		if err != nil {
			return ProjectSet{}, err
		}
	}
	return s.BuildProjects(request)
}

func TestBuildProjectsSeparatesProjectPolicies(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}},
		{Files: []string{"b/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false), Project: rslintconfig.ProjectPaths{"b/custom.json"}}}},
		{Files: []string{"**/disabled.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false), ProjectDisabled: true}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	session := NewSession(fsys)
	plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{dir + "/a/src/file.ts", dir + "/a/src/disabled.ts", dir + "/b/src/file.ts"}})
	if err != nil {
		t.Fatal(err)
	}
	projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, LintTargets, true)
	if err != nil {
		t.Fatal(err)
	}
	if projects.Len() != 2 {
		t.Fatalf("expected two owning projects, got %d", projects.Len())
	}
	binding, err := session.LoadAPI(projects, plan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(binding.Programs) != 3 {
		t.Fatalf("disabled target must get its own source-only Program: %d", len(binding.Programs))
	}
	for index, sources := range binding.TargetsByProgram {
		if len(sources) != 1 {
			t.Fatalf("exact target scope changed: %v", binding.TargetsByProgram)
		}
		if strings.HasSuffix(sources[0], "/disabled.ts") {
			if index < projects.Len() {
				t.Fatal("disabled target borrowed a configured Program")
			}
			continue
		}
		program := projects.compilerPrograms[index]
		wantConfig := dir + "/a/tsconfig.json"
		if strings.Contains(sources[0], "/b/") {
			wantConfig = dir + "/b/custom.json"
		} else if len(program.CommandLine().FileNames()) != 3 {
			t.Fatal("single-target lint truncated the owning project's roots")
		}
		if program.Options().ConfigFilePath != wantConfig || !program.Options().Strict.IsTrue() {
			t.Fatalf("wrong project/options: %+v", program.Options())
		}
	}
}

func TestBuildProjectsKeepsServiceGapsSourceOnly(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	config := rslintconfig.RslintConfig{
		{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}},
		{Files: []string{"b/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false), Project: rslintconfig.ProjectPaths{"gap-overlap.json"}}}},
	}
	files := []string{tspath.ResolvePath(dir, "b/src/file.ts"), tspath.ResolvePath(dir, "a/src/file.ts")}
	for _, extension := range []string{"ts", "tsx", "mts", "cts", "js", "jsx", "mjs", "cjs"} {
		files = append(files, tspath.ResolvePath(dir, "outside/file."+extension))
	}
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		for _, mode := range []string{"cli", "api"} {
			t.Run(strconv.Itoa(int(scope))+"/"+mode, func(t *testing.T) {
				fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
				session := NewSession(fsys)
				plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: files})
				if err != nil {
					t.Fatal(err)
				}
				projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, scope, true)
				if err != nil || projects.Len() != 2 {
					t.Fatalf("owning projects = %d, error = %v", projects.Len(), err)
				}
				// The explicit target's complete Program also contains the gaps.
				// Those files must retain their own service policy's miss.
				var overlap bool
				for _, program := range projects.compilerPrograms {
					overlap = overlap || program.GetSourceFile(files[2]) != nil
				}
				if !overlap {
					t.Fatal("fixture did not create overlapping project membership")
				}
				load := session.LoadAPI
				if mode == "cli" {
					load = session.LoadCLI
				}
				binding, err := load(projects, plan, dir, true)
				if err != nil {
					t.Fatal(err)
				}
				seen := make(map[string]int)
				for index, sources := range binding.TargetsByProgram {
					program := binding.Programs[index]
					for _, source := range sources {
						seen[source]++
						file := program.GetSourceFile(source)
						wantTypes := !strings.Contains(source, "/outside/")
						if file == nil || program.CanProvideTypeChecker(file) != wantTypes {
							t.Fatalf("%s: wrong checker capability, want types=%v", source, wantTypes)
						}
					}
				}
				if len(seen) != len(files) {
					t.Fatalf("target scope changed: %v", seen)
				}
				for _, file := range files {
					if seen[file] != 1 {
						t.Fatalf("%s linted %d times", file, seen[file])
					}
				}
			})
		}
	}
}

func TestBuildProjectsKeepsProgramModesSeparate(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service_modes.txtar")
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		for _, output := range []string{"stale", "missing"} {
			for _, mode := range []string{"explicit", "mixed", "service"} {
				t.Run(strconv.Itoa(int(scope))+"/"+output+"/"+mode, func(t *testing.T) {
					dir := tspath.NormalizePath(archive.Materialize(t, ""))
					if output == "missing" {
						if err := os.Remove(tspath.ResolvePath(dir, "dist/value.d.ts")); err != nil {
							t.Fatal(err)
						}
					}
					var config rslintconfig.RslintConfig
					files := []string{tspath.ResolvePath(dir, "a.ts"), tspath.ResolvePath(dir, "b.ts")}
					serviceForFile := []bool{mode == "service", mode != "explicit"}
					for index, name := range []string{"a.ts", "b.ts"} {
						options := &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(serviceForFile[index])}
						if !serviceForFile[index] {
							options.Project = rslintconfig.ProjectPaths{"tsconfig.json"}
						}
						config = append(config, rslintconfig.ConfigEntry{
							Files: []string{name}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: options},
						})
					}
					fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
					plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: files})
					if err != nil {
						t.Fatal(err)
					}
					session := NewSession(fsys)
					projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, scope, true)
					if err != nil {
						t.Fatal(err)
					}
					wantPrograms := 1
					if mode == "mixed" {
						wantPrograms = 2
					}
					if projects.Len() != wantPrograms {
						t.Fatalf("got %d Programs for one tsconfig, want %d", projects.Len(), wantPrograms)
					}
					binding, err := session.LoadAPI(projects, plan, dir, true)
					if err != nil {
						t.Fatal(err)
					}
					programForFile := make(map[string]*compiler.Program)
					for index, sources := range binding.TargetsByProgram {
						program := projects.compilerPrograms[index]
						if program.Options().ConfigFilePath != tspath.ResolvePath(dir, "tsconfig.json") || len(program.CommandLine().FileNames()) != 3 {
							t.Fatal("selected project lost its config identity or complete roots")
						}
						for _, file := range sources {
							if programForFile[file] != nil {
								t.Fatalf("target %s was bound twice", file)
							}
							programForFile[file] = program
						}
						if program.GetSourceFile(tspath.ResolvePath(dir, "c.ts")) == nil {
							t.Fatal("sibling outside lint targets must remain in the type context")
						}
					}
					if len(programForFile) != 2 {
						t.Fatalf("expected only the two requested lint targets, got %v", programForFile)
					}
					for index, file := range files {
						program := programForFile[file]
						if program == nil {
							t.Fatalf("target %s has no Program", file)
						}
						usesSource := program.GetSourceFile(tspath.ResolvePath(dir, "lib/value.ts")) != nil
						usesDeclaration := program.GetSourceFile(tspath.ResolvePath(dir, "dist/value.d.ts")) != nil
						if usesSource != serviceForFile[index] || usesDeclaration != (!serviceForFile[index] && output == "stale") {
							t.Fatalf("target %s borrowed the wrong reference context: source=%v, declaration=%v", file, usesSource, usesDeclaration)
						}
						var codes []int
						for _, diagnostic := range program.GetSemanticDiagnostics(context.Background(), program.GetSourceFile(file)) {
							codes = append(codes, int(diagnostic.Code()))
						}
						var wantCodes []int
						if serviceForFile[index] {
							wantCodes = []int{2322}
						} else if output == "missing" {
							wantCodes = []int{6305}
						}
						if !slices.Equal(codes, wantCodes) {
							t.Fatalf("target %s: semantic codes = %v, want %v", file, codes, wantCodes)
						}
					}
				})
			}
		}
	}
}

func TestBuildProjectsValidatesServiceSourcesBeforePublishing(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service_modes.txtar")
	for _, test := range []struct {
		name, fixture, file string
		wantProjects        int
		wantError           bool
	}{
		{name: "configured root", file: "a.ts", wantProjects: 1},
		{name: "selected source absent", fixture: "disabled-source-redirect", file: "target.ts", wantError: true},
		{name: "no configured owner", fixture: "no-config", file: "target.ts"},
	} {
		for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
			t.Run(test.name+"/"+strconv.Itoa(int(scope)), func(t *testing.T) {
				dir := tspath.NormalizePath(archive.Materialize(t, test.fixture))
				file := tspath.ResolvePath(dir, test.file)
				config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{
					ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)},
				}}}
				fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
				plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{file}})
				if err != nil || len(plan.Files) != 1 {
					t.Fatalf("target selection: files=%v, error=%v", plan.Files, err)
				}
				session := NewSession(fsys)
				projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, scope, true)
				if test.wantError {
					// AllDeclared may be consumed without LoadCLI/LoadAPI by
					// --type-check-only, so construction must report this failure.
					if err == nil || !strings.Contains(err.Error(), "was absent") || !strings.Contains(err.Error(), file) {
						t.Fatalf("selected source failure became a published project/gap: projects=%d, error=%v", projects.Len(), err)
					}
					return
				}
				if err != nil || projects.Len() != test.wantProjects {
					t.Fatalf("project construction: projects=%d, error=%v; want %d", projects.Len(), err, test.wantProjects)
				}
				binding, err := session.LoadAPI(projects, plan, dir, true)
				if err != nil {
					t.Fatal(err)
				}
				if len(binding.Programs) != 1 || len(binding.TargetsByProgram) != 1 || !slices.Equal(binding.TargetsByProgram[0], []string{file}) {
					t.Fatalf("target binding changed: %v", binding.TargetsByProgram)
				}
				program := binding.Programs[0]
				source := program.GetSourceFile(file)
				if source == nil || program.CanProvideTypeChecker(source) != (test.wantProjects != 0) {
					t.Fatalf("wrong source/type capability for %s", file)
				}
				if test.wantProjects == 0 && (!program.Options().NoResolve.IsTrue() || !program.Options().NoLib.IsTrue()) {
					t.Fatal("unowned target did not use the existing source-only gap Program")
				}
			})
		}
	}
}

func TestBuildProjectsScopesInactiveOwners(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	configs := map[string]rslintconfig.RslintConfig{
		dir + "/a": {{Files: []string{"src/**/*.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}}},
		dir + "/b": {{Ignores: []string{"**"}}, {LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"custom.json"}}}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	plan, err := target.Resolve(target.Request{ConfigMap: configs, ConfigDirectory: dir, FS: fsys, Files: []string{dir + "/a/src/file.ts"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		t.Run(strconv.Itoa(int(scope)), func(t *testing.T) {
			projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t, configs, plan, scope, true)
			if err != nil {
				t.Fatal(err)
			}
			want := 1
			if scope == AllDeclared {
				want = 2
			}
			if projects.Len() != want {
				t.Fatalf("scope %d built %d Programs, want %d", scope, projects.Len(), want)
			}
			if !containsTSDiagnostic(collectProgramTypeDiagnostics(t, projects.Programs()), "TS2322") {
				t.Fatal("complete selected projects must retain sibling type errors")
			}
		})
	}
}

func TestBuildProjectsSkipsOwnersWithoutTargets(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	for _, reset := range []string{"false", "null"} {
		for _, withActiveOwner := range []bool{false, true} {
			t.Run(reset+"/active-owner="+strconv.FormatBool(withActiveOwner), func(t *testing.T) {
				projectOwner := tspath.ResolvePath(dir, "b")
				otherOwner := tspath.ResolvePath(dir, "a")
				var override rslintconfig.RslintConfig
				if err := json.Unmarshal([]byte(`[{"languageOptions":{"parserOptions":{"projectService":`+reset+`}}}]`), &override); err != nil {
					t.Fatal(err)
				}
				config := rslintconfig.RslintConfig{{Ignores: []string{"**"}}}
				config = append(config, rslintconfig.ConfigWithAuthoredPathBase(rslintconfig.RslintConfig{{
					LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
						ProjectService: rslintconfig.BoolPtr(true), Project: rslintconfig.ProjectPaths{"missing.json"},
					}},
				}}, projectOwner)...)
				config = append(config, rslintconfig.ConfigWithAuthoredPathBase(override, otherOwner)...)
				configs := map[string]rslintconfig.RslintConfig{projectOwner: config}
				fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
				request := target.Request{ConfigMap: configs, ConfigDirectory: dir, FS: fsys, Directories: []string{projectOwner}}
				wantPrograms := 0
				if withActiveOwner {
					configs[otherOwner] = rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
						ProjectService: rslintconfig.BoolPtr(true),
					}}}}
					request.Files = []string{tspath.ResolvePath(otherOwner, "src/file.ts")}
					wantPrograms++
				}
				plan, err := target.Resolve(request)
				if err != nil {
					t.Fatal(err)
				}
				if len(plan.Files) != wantPrograms {
					t.Fatalf("inactive owner selected lint targets: %+v", plan.Files)
				}
				projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t, configs, plan, LintTargets, true)
				if err != nil {
					t.Fatal(err)
				}
				if projects.Len() != wantPrograms {
					t.Fatalf("unselected declarations created Programs: %d", projects.Len())
				}
				foundProject := false
				for _, program := range projects.compilerPrograms {
					foundProject = foundProject || program.Options().ConfigFilePath == tspath.ResolvePath(projectOwner, "missing.json")
				}
				if foundProject {
					t.Fatal("an owner without targets loaded its project declaration")
				}
			})
		}
	}
}

func TestBuildProjectsKeepsEffectiveCandidatesSeparateFromTypeCheckScope(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false)}}},
		{Files: []string{"b/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"a/tsconfig.json", "b/custom.json"}}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	files := []string{dir + "/a/src/file.ts", dir + "/b/src/file.ts"}
	plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: files})
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		t.Run(strconv.Itoa(int(scope)), func(t *testing.T) {
			session := NewSession(fsys)
			projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, scope, true)
			wantProjects := 1
			if scope == AllDeclared {
				wantProjects = 2
			}
			if err != nil || projects.Len() != wantProjects {
				t.Fatalf("projects=%d, want %d; error=%v", projects.Len(), wantProjects, err)
			}
			for _, load := range []func(ProjectSet, target.Plan, string, bool) (LoadResult, error){session.LoadCLI, session.LoadAPI} {
				binding, err := load(projects, plan, dir, true)
				if err != nil {
					t.Fatal(err)
				}
				seen := 0
				for index, boundFiles := range binding.TargetsByProgram {
					for _, file := range boundFiles {
						seen++
						typed := index < projects.Len()
						if typed != (file == files[1]) {
							t.Fatalf("target %s borrowed an inapplicable project: %v", file, binding.TargetsByProgram)
						}
					}
				}
				if seen != len(files) {
					t.Fatalf("project selection changed lint targets: %v", binding.TargetsByProgram)
				}
			}
		})
	}
}

func TestBuildProjectsDeduplicatesExplicitProjectsAcrossOwners(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	legacy := rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"../shared.json"}}}}}
	configs := map[string]rslintconfig.RslintConfig{
		tspath.ResolvePath(dir, "a"):       legacy,
		tspath.ResolvePath(dir, "b"):       legacy,
		tspath.ResolvePath(dir, "service"): {{Files: []string{"**/*.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	plan, err := target.Resolve(target.Request{
		ConfigMap: configs, ConfigDirectory: dir, FS: fsys,
		Files: []string{tspath.ResolvePath(dir, "a/src/file.ts"), tspath.ResolvePath(dir, "b/src/file.ts"), tspath.ResolvePath(dir, "service/file.ts")},
	})
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession(fsys)
	projects, err := session.buildProjectsWithOptionsForTest(t, configs, plan, LintTargets, false)
	if err != nil {
		t.Fatal(err)
	}
	if projects.Len() != 2 {
		t.Fatalf("duplicated the shared explicit project: %d", projects.Len())
	}
	binding, err := session.LoadAPI(projects, plan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(binding.Programs) != 2 || len(binding.TargetsByProgram[0])+len(binding.TargetsByProgram[1]) != 3 {
		t.Fatalf("shared projects lost config ownership: %v", binding.TargetsByProgram)
	}
}

func TestBuildProjectPlanReusesCandidatesAcrossRuleConfigs(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	for _, separateRules := range []bool{false, true} {
		t.Run(strconv.FormatBool(separateRules), func(t *testing.T) {
			config := projectConfig("a/tsconfig*.json")
			if separateRules {
				config = append(config, rslintconfig.ConfigEntry{
					Files: []string{"**/disabled.ts"}, Rules: rslintconfig.Rules{"no-debugger": "error"},
				})
			}
			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			plan, err := target.Resolve(target.Request{
				Config: config, ConfigDirectory: dir, FS: fsys,
				Files: []string{tspath.ResolvePath(dir, "a/src/file.ts"), tspath.ResolvePath(dir, "a/src/disabled.ts")},
			})
			if err != nil {
				t.Fatal(err)
			}
			resolver := configLint.NewResolver(configLint.ResolverOptions{
				Config: config, ConfigDirectory: dir, FS: fsys, PathSpaces: plan.PathSpaces(), Catalog: rule.NewCatalog(),
			})
			policies, err := resolver.ProjectPolicies(plan.Files, false)
			if err != nil {
				t.Fatal(err)
			}
			counting := &projectGlobCountingFS{FS: fsys}
			projects := buildProjectPlan(ProjectBuildRequest{Targets: plan, Policies: policies, Scope: LintTargets}, counting)
			if projects.terminalErr != nil || len(projects.specs) != 1 {
				t.Fatalf("project plan: count=%d, error=%v", len(projects.specs), projects.terminalErr)
			}
			if counting.calls != 1 {
				t.Errorf("expanded one declaration %d times across rule configs, want 1", counting.calls)
			}
			groups := groupTargetsByProjects(plan.Files, projects.targetProjects, func(string) []int {
				t.Fatal("effective candidates must not fall back to owner declarations")
				return nil
			})
			if len(groups) != 1 || len(groups[0].targetIndexes) != 2 {
				t.Fatalf("one project request split into target groups: %+v", groups)
			}
		})
	}
}

type projectGlobCountingFS struct {
	vfs.FS
	calls int
}

func (fsys *projectGlobCountingFS) DirectoryExists(path string) bool {
	fsys.calls++
	return fsys.FS.DirectoryExists(path)
}

func TestBuildProjectPlanKeepsProjectPathContextsSeparate(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	patterns := rslintconfig.ProjectPaths{"tsconfig*.json"}
	base := tspath.ResolvePath(dir, "a")
	otherBase := tspath.ResolvePath(dir, "b")
	first := rslintconfig.ProjectPolicy{ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: patterns, BaseDirectory: base}}
	longerPatterns := rslintconfig.ProjectPaths{"a/tsconfig.json", "b/tsconfig.json"}
	for _, test := range []struct {
		name          string
		first, second rslintconfig.ProjectPolicy
		wantSecond    []string
	}{
		{
			name: "authored base", first: first,
			second:     rslintconfig.ProjectPolicy{ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: patterns, BaseDirectory: otherBase}},
			wantSecond: []string{tspath.ResolvePath(otherBase, "tsconfig.json")},
		},
		{
			name: "root override", first: first,
			second:     rslintconfig.ProjectPolicy{ExplicitProject: first.ExplicitProject, TSConfigRootDirOverride: otherBase},
			wantSecond: []string{tspath.ResolvePath(otherBase, "tsconfig.json")},
		},
		{
			name: "disabled", first: first,
			second: rslintconfig.ProjectPolicy{ExplicitProject: first.ExplicitProject, ProjectDisabled: true},
		},
		{
			name:       "shared patterns with different lengths",
			first:      rslintconfig.ProjectPolicy{ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: longerPatterns[:1], BaseDirectory: dir}},
			second:     rslintconfig.ProjectPolicy{ExplicitProject: &rslintconfig.ProjectDeclaration{Patterns: longerPatterns, BaseDirectory: dir}},
			wantSecond: []string{tspath.ResolvePath(base, "tsconfig.json"), tspath.ResolvePath(otherBase, "tsconfig.json")},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			files := []target.File{
				testLintTarget(fsys, dir, tspath.ResolvePath(base, "src/file.ts")),
				testLintTarget(fsys, dir, tspath.ResolvePath(otherBase, "src/file.ts")),
			}
			plan := buildProjectPlan(ProjectBuildRequest{
				Targets: target.Plan{Files: files}, Scope: LintTargets,
				Policies: map[target.File]rslintconfig.ProjectPolicy{files[0]: test.first, files[1]: test.second},
			}, fsys)
			if plan.terminalErr != nil {
				t.Fatal(plan.terminalErr)
			}
			for i, want := range [][]string{{tspath.ResolvePath(base, "tsconfig.json")}, test.wantSecond} {
				var paths []string
				for _, index := range plan.targetProjects[files[i]] {
					paths = append(paths, plan.specs[index].tsconfigPath)
				}
				if !slices.Equal(paths, want) {
					t.Errorf("target %d projects=%v, want %v", i, paths, want)
				}
			}
		})
	}
}

func TestBuildProjectsPreservesRawDeclarationsAfterEmptyArray(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	config := rslintconfig.RslintConfig{
		{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false), Project: rslintconfig.ProjectPaths{"a/tsconfig.json"}}}},
		{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{}}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{tspath.ResolvePath(dir, "a/src/file.ts")}})
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		t.Run(strconv.Itoa(int(scope)), func(t *testing.T) {
			session := NewSession(fsys)
			projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, scope, true)
			wantProjects := 0
			if scope == AllDeclared {
				wantProjects = 1
			}
			if err != nil || projects.Len() != wantProjects {
				t.Fatalf("cleared lint candidate changed type-check scope: count=%d error=%v", projects.Len(), err)
			}
			binding, err := session.LoadAPI(projects, plan, dir, true)
			if err != nil || len(binding.TargetsByProgram[wantProjects]) != 1 {
				t.Fatalf("cleared target did not use gap lint: %v, %v", binding.TargetsByProgram, err)
			}
		})
	}
}

func TestBuildProjectsPreservesProjectDeclarationOrder(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, test := range []struct {
		name       string
		options    string
		matched    bool
		restore    bool
		wantConfig string
	}{
		{"baseline", `{}`, false, false, "tsconfig.safe.json"},
		{"unmatched service true", `{"projectService":true}`, false, false, "tsconfig.safe.json"},
		{"unmatched service false", `{"projectService":false}`, false, false, "tsconfig.safe.json"},
		{"unmatched service null", `{"projectService":null}`, false, false, "tsconfig.safe.json"},
		{"unmatched root", `{"tsconfigRootDir":"."}`, false, false, "tsconfig.safe.json"},
		{"unmatched root null", `{"tsconfigRootDir":null}`, false, false, "tsconfig.safe.json"},
		{"unmatched project false", `{"project":false}`, false, false, "tsconfig.safe.json"},
		{"unmatched project null", `{"project":null}`, false, false, "tsconfig.safe.json"},
		{"unmatched missing project", `{"project":"missing.json"}`, false, false, "tsconfig.safe.json"},
		{"unmatched project and service", `{"projectService":false,"project":"missing.json"}`, false, false, "tsconfig.safe.json"},
		{"matched project array order", `{"project":["./tsconfig.unsafe.json","./tsconfig.safe.json"]}`, true, false, "tsconfig.unsafe.json"},
		{"matched reversed project array order", `{"project":["./tsconfig.safe.json","./tsconfig.unsafe.json"]}`, true, false, "tsconfig.safe.json"},
		{"matched false then restore", `{"project":false}`, true, true, "tsconfig.safe.json"},
		{"matched null then restore", `{"project":null}`, true, true, "tsconfig.safe.json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := tspath.ResolvePath(archive.Materialize(t, "legacy-policy"), "pkg[1]")
			file := tspath.ResolvePath(dir, "target.ts")
			var options rslintconfig.ParserOptions
			if err := json.Unmarshal([]byte(test.options), &options); err != nil {
				t.Fatal(err)
			}
			selector := "unused.ts"
			if test.matched {
				selector = "target.ts"
			}
			config := rslintconfig.RslintConfig{
				{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"./tsconfig.unsafe.json"}}}},
				{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"./tsconfig.safe.json"}}}},
				{Files: []string{selector}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &options}},
			}
			if strings.Contains(test.name, "array order") {
				config = config[2:]
			}
			wantConfig := test.wantConfig
			if test.restore {
				config = append(config, rslintconfig.ConfigEntry{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"./tsconfig.safe.json"}}}})
				wantConfig = "tsconfig.safe.json"
			}
			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{file}})
			if err != nil {
				t.Fatal(err)
			}
			for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
				t.Run(strconv.Itoa(int(scope)), func(t *testing.T) {
					session := NewSession(fsys)
					projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, scope, true)
					if scope == AllDeclared && strings.Contains(test.options, "missing.json") {
						if err == nil || !strings.Contains(err.Error(), "missing.json") {
							t.Fatalf("type checking lost raw declaration validation: %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					for _, load := range []func(ProjectSet, target.Plan, string, bool) (LoadResult, error){session.LoadAPI, session.LoadCLI} {
						binding, err := load(projects, plan, dir, true)
						if err != nil {
							t.Fatal(err)
						}
						selected := -1
						for index, files := range binding.TargetsByProgram {
							if slices.Contains(files, file) {
								selected = index
							}
						}
						if selected < 0 || selected >= projects.Len() {
							t.Fatal("target lost its configured Program")
						}
						program := projects.compilerPrograms[selected]
						if program.Options().ConfigFilePath != tspath.ResolvePath(dir, wantConfig) {
							t.Fatalf("target selected %q, want %s", program.Options().ConfigFilePath, wantConfig)
						}
						codes := program.GetSemanticDiagnostics(context.Background(), program.GetSourceFile(file))
						if wantConfig == "tsconfig.safe.json" && (len(codes) != 1 || codes[0].Code() != 2322) || wantConfig == "tsconfig.unsafe.json" && len(codes) != 0 {
							t.Fatalf("wrong typed context after policy merge: %v", codes)
						}
					}
				})
			}
		})
	}
}

func TestBuildProjectsPreservesRawProjectBases(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, pattern := range []string{"./tsconfig.unsafe.json", "./tsconfig.u*.json"} {
		t.Run(pattern, func(t *testing.T) {
			dir := tspath.ResolvePath(archive.Materialize(t, "legacy-policy"), "pkg[1]")
			config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false), Project: rslintconfig.ProjectPaths{pattern}}}}}
			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{tspath.ResolvePath(dir, "target.ts")}})
			if err != nil {
				t.Fatal(err)
			}
			projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, LintTargets, true)
			if err != nil || projects.Len() != 1 || projects.compilerPrograms[0].Options().ConfigFilePath != tspath.ResolvePath(dir, "tsconfig.unsafe.json") {
				t.Fatalf("literal directory was treated as a project glob: count=%d error=%v", projects.Len(), err)
			}
		})
	}
	t.Run("same raw path with different bases", func(t *testing.T) {
		dir := tspath.NormalizePath(archive.Materialize(t, ""))
		var config rslintconfig.RslintConfig
		for _, base := range []string{"a", "b"} {
			config = append(config, rslintconfig.ConfigEntry{BasePath: &base, Files: []string{"src/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false), Project: rslintconfig.ProjectPaths{"./tsconfig.json"}}}})
		}
		config = rslintconfig.ConfigWithResolvedBasePaths(config, dir)
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{tspath.ResolvePath(dir, "a/src/file.ts"), tspath.ResolvePath(dir, "b/src/file.ts")}})
		if err != nil {
			t.Fatal(err)
		}
		session := NewSession(fsys)
		projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, LintTargets, true)
		if err != nil || projects.Len() != 2 {
			t.Fatalf("raw paths with different bases merged into one group: count=%d error=%v", projects.Len(), err)
		}
		binding, err := session.LoadAPI(projects, plan, dir, true)
		if err != nil {
			t.Fatal(err)
		}
		for index, files := range binding.TargetsByProgram {
			for _, file := range files {
				want := tspath.ResolvePath(tspath.GetDirectoryPath(tspath.GetDirectoryPath(file)), "tsconfig.json")
				if index >= projects.Len() || projects.compilerPrograms[index].Options().ConfigFilePath != want {
					t.Fatalf("%s borrowed another entry's raw project base", file)
				}
			}
		}
	})
}

func TestBuildProjectsRootContextsShareExecution(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		for _, pattern := range []string{"config.json", "config*.json"} {
			t.Run(strconv.Itoa(int(scope))+"/"+pattern, func(t *testing.T) {
				dir := tspath.NormalizePath(archive.Materialize(t, "root-contexts"))
				owner := tspath.ResolvePath(dir, "owner")
				roots := []string{tspath.ResolvePath(dir, "a[1]"), tspath.ResolvePath(dir, "b")}
				config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{pattern}}}}}
				for index, file := range []string{"x.ts", "y.ts"} {
					config = append(config, rslintconfig.ConfigEntry{Files: []string{file}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{TsconfigRootDir: &roots[index]}}})
				}
				files := []string{tspath.ResolvePath(owner, "x.ts"), tspath.ResolvePath(owner, "y.ts")}
				fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
				plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: owner, FS: fsys, Files: files})
				if err != nil {
					t.Fatal(err)
				}
				session := NewSession(fsys)
				projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{owner: config}, plan, scope, false)
				if err != nil {
					t.Fatal(err)
				}
				if projects.Len() != 2 {
					t.Fatalf("root contexts built %d Programs, want 2", projects.Len())
				}
				for _, load := range []func(ProjectSet, target.Plan, string, bool) (LoadResult, error){session.LoadCLI, session.LoadAPI} {
					binding, err := load(projects, plan, owner, true)
					if err != nil {
						t.Fatal(err)
					}
					seen := 0
					for index, sources := range binding.TargetsByProgram {
						for _, file := range sources {
							seen++
							root := roots[0]
							if file == files[1] {
								root = roots[1]
							}
							if index >= projects.Len() || projects.compilerPrograms[index].Options().ConfigFilePath != tspath.ResolvePath(root, "config.json") {
								t.Fatalf("%s borrowed the other root's Program: %v", file, binding.TargetsByProgram)
							}
							if file == files[1] {
								program := projects.compilerPrograms[index]
								diagnostics := program.GetSemanticDiagnostics(context.Background(), program.GetSourceFile(file))
								if len(diagnostics) != 1 || diagnostics[0].Code() != 2322 {
									t.Fatalf("later root lost its distinct typed context: %v", diagnostics)
								}
							}
						}
					}
					if seen != 2 {
						t.Fatalf("root contexts lost targets: %v", binding.TargetsByProgram)
					}
				}
			})
		}
	}
}

func TestBuildProjectsWithoutEffectiveProjectsKeepTargetsUnbound(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		for _, options := range []string{`{}`, `{"project":[]}`, `{"projectService":false}`, `{"projectService":null}`, `{"project":false}`, `{"project":null}`} {
			t.Run(strconv.Itoa(int(scope))+"/"+options, func(t *testing.T) {
				dir := tspath.NormalizePath(archive.Materialize(t, "a"))
				var disabled rslintconfig.ParserOptions
				if err := json.Unmarshal([]byte(options), &disabled); err != nil {
					t.Fatal(err)
				}
				config := rslintconfig.RslintConfig{{Files: []string{"**/disabled.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &disabled}}}
				files := []string{tspath.ResolvePath(dir, "src/file.ts"), tspath.ResolvePath(dir, "src/disabled.ts")}
				fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
				plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: files})
				if err != nil {
					t.Fatal(err)
				}
				session := NewSession(fsys)
				projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, scope, true)
				wantProjects := 0
				if scope == AllDeclared {
					wantProjects = 1
					if options == `{"project":[]}` {
						wantProjects = 0
					}
				}
				if err != nil || projects.Len() != wantProjects {
					t.Fatalf("implicit project scope: count=%d, want %d, error=%v", projects.Len(), wantProjects, err)
				}
				if wantProjects > 0 && projects.compilerPrograms[0].GetSourceFile(files[1]) == nil {
					t.Fatal("fixture must contain the disabled target in the ordinary Program")
				}
				for _, load := range []func(ProjectSet, target.Plan, string, bool) (LoadResult, error){session.LoadCLI, session.LoadAPI} {
					binding, err := load(projects, plan, dir, true)
					if err != nil {
						t.Fatal(err)
					}
					if len(binding.Programs) != wantProjects+1 || len(binding.TargetsByProgram[wantProjects]) != len(files) || !slices.Contains(binding.TargetsByProgram[wantProjects], files[0]) || !slices.Contains(binding.TargetsByProgram[wantProjects], files[1]) {
						t.Fatalf("disabled target borrowed the default Program: %v", binding.TargetsByProgram)
					}
				}
			})
		}
	}
}

func TestBuildProjectsAllDeclaredKeepsRootForServiceAndClear(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, mode := range []string{"ordinary", "clear", "service"} {
		t.Run(mode, func(t *testing.T) {
			dir := tspath.NormalizePath(archive.Materialize(t, "root-contexts"))
			owner := tspath.ResolvePath(dir, "owner")
			root := tspath.ResolvePath(dir, "a[1]")
			config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"config.json"}, TsconfigRootDir: &root}}}}
			if mode != "ordinary" {
				resetOptions := &rslintconfig.ParserOptions{ProjectDisabled: true}
				if mode == "service" {
					resetOptions.ProjectService = rslintconfig.BoolPtr(true)
				}
				config = append(config, rslintconfig.ConfigEntry{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: resetOptions}})
			}
			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: owner, FS: fsys, Files: []string{tspath.ResolvePath(owner, "x.ts")}})
			if err != nil {
				t.Fatal(err)
			}
			session := NewSession(fsys)
			projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{owner: config}, plan, AllDeclared, true)
			if err != nil {
				t.Fatal(err)
			}
			want := 1
			if mode == "service" {
				want = 2
			}
			if projects.Len() != want || projects.compilerPrograms[0].Options().ConfigFilePath != tspath.ResolvePath(root, "config.json") {
				t.Fatalf("%s changed the raw declaration's root: count=%d", mode, projects.Len())
			}
			binding, err := session.LoadAPI(projects, plan, owner, true)
			if err != nil {
				t.Fatal(err)
			}
			if mode != "ordinary" && len(binding.TargetsByProgram[0]) != 0 {
				t.Fatalf("%s borrowed the program-wide explicit project: %v", mode, binding.TargetsByProgram)
			}
		})
	}
}

func TestBuildProjectsAllDeclaredWithoutTargets(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, options := range []string{`{}`, `{"projectService":true}`, `{"tsconfigRootDir":"/unused"}`, `{"projectService":true,"project":"missing.json"}`} {
		t.Run(options, func(t *testing.T) {
			dir := tspath.NormalizePath(archive.Materialize(t, "a"))
			var parsed rslintconfig.ParserOptions
			if err := json.Unmarshal([]byte(options), &parsed); err != nil {
				t.Fatal(err)
			}
			config := rslintconfig.RslintConfig{{Ignores: []string{"**"}}, {LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &parsed}}}
			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Directories: []string{dir}})
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Files) != 0 {
				t.Fatal("fixture selected targets")
			}
			projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, AllDeclared, true)
			if strings.Contains(options, "missing.json") {
				if err == nil || !strings.Contains(err.Error(), "missing.json") {
					t.Fatalf("program-wide declaration error was hidden: %v", err)
				}
				return
			}
			if err != nil || projects.Len() != 1 || projects.compilerPrograms[0].Options().ConfigFilePath != tspath.ResolvePath(dir, "tsconfig.json") {
				t.Fatalf("zero targets changed program-wide default loading: %d, %v", projects.Len(), err)
			}
		})
	}
}

func TestBuildProjectsRootContextsDeduplicateSharedPath(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, "root-contexts"))
	owner := tspath.ResolvePath(dir, "owner")
	sharedConfig := tspath.ResolvePath(owner, "tsconfig.json")
	roots := []string{tspath.ResolvePath(dir, "a[1]"), tspath.ResolvePath(dir, "b")}
	config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{sharedConfig}}}}}
	for index, name := range []string{"x.ts", "y.ts"} {
		config = append(config, rslintconfig.ConfigEntry{Files: []string{name}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{TsconfigRootDir: &roots[index]}}})
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: owner, FS: fsys, Files: []string{tspath.ResolvePath(owner, "x.ts"), tspath.ResolvePath(owner, "y.ts")}})
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		t.Run(strconv.Itoa(int(scope)), func(t *testing.T) {
			session := NewSession(fsys)
			projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{owner: config}, plan, scope, false)
			if err != nil || projects.Len() != 1 {
				t.Fatalf("root contexts duplicated the same explicit project: %d, %v", projects.Len(), err)
			}
			binding, err := session.LoadAPI(projects, plan, owner, true)
			if err != nil {
				t.Fatal(err)
			}
			if len(binding.Programs) != 1 || len(binding.TargetsByProgram[0]) != 2 {
				t.Fatalf("shared project lost targets: %v", binding.TargetsByProgram)
			}
		})
	}
}

func TestLoadProgramsSingleCandidateRequiresRootMembership(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service_modes.txtar").Materialize(t, ""))
	config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}}}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	requestPlan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{tspath.ResolvePath(dir, "a.ts")}})
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession(fsys)
	projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, requestPlan, LintTargets, true)
	if err != nil || projects.Len() != 1 {
		t.Fatalf("service fixture did not select one Program: %d, %v", projects.Len(), err)
	}
	for _, name := range []string{"a.ts", "lib/value.ts", "outside.ts"} {
		t.Run(name, func(t *testing.T) {
			path := tspath.ResolvePath(dir, name)
			file := target.File{PathIdentity: rslintconfig.PathIdentity{Path: path, CanonicalPath: path, CanonicalParentPath: tspath.GetDirectoryPath(path)}, ConfigDirectory: dir}
			set := projects
			set.targetBinding = nil
			set.targetProjects = map[target.File][]int{file: {0}}
			counting := &targetPlanRealpathCountingFS{FS: fsys, calls: make(map[string]int)}
			owners := directRootProgramOwners(set, []target.File{file}, counting, true)
			wantOwner := -1
			if name == "a.ts" {
				wantOwner = 0
			}
			if !slices.Equal(owners, []int{wantOwner}) {
				t.Fatalf("single candidate bypassed root membership: owners=%v, want %d", owners, wantOwner)
			}
			plan := target.Plan{Files: []target.File{file}}
			binding, err := session.LoadAPI(set, plan, dir, true)
			if err != nil {
				t.Fatal(err)
			}
			if name == "outside.ts" {
				if len(binding.Programs) != 2 || len(binding.TargetsByProgram[0]) != 0 || !slices.Equal(binding.TargetsByProgram[1], []string{path}) {
					t.Fatalf("a candidate without the source was treated as an owner: %v", binding.TargetsByProgram)
				}
			} else if len(binding.Programs) != 1 || !slices.Equal(binding.TargetsByProgram[0], []string{path}) {
				t.Fatalf("single root/import candidate did not bind: %v", binding.TargetsByProgram)
			}
		})
	}
}

type targetPlanRealpathCountingFS struct {
	vfs.FS
	mu    sync.Mutex
	calls map[string]int
}

type retargetingFrozenTargetFS struct {
	vfs.FS
	targetPath        string
	canonicalPath     string
	liveCanonicalPath string
	targetCalls       int
	canonicalCalls    int
}

func (fsys *retargetingFrozenTargetFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if filePath == fsys.targetPath {
		fsys.targetCalls++
		return fsys.liveCanonicalPath
	}
	if filePath == fsys.canonicalPath {
		fsys.canonicalCalls++
		return fsys.liveCanonicalPath
	}
	return fsys.FS.Realpath(filePath)
}

type blockingProgramConfigFS struct {
	vfs.FS
	paths      map[string]struct{}
	waitFor    int
	mu         sync.Mutex
	active     int
	peak       int
	allStarted chan struct{}
	release    chan struct{}
	startOnce  sync.Once
}

type unreadableProjectConfigFS struct {
	vfs.FS
	configPath string
}

func (fsys *unreadableProjectConfigFS) ReadFile(filePath string) (string, bool) {
	if tspath.NormalizePath(filePath) == fsys.configPath {
		return "", false
	}
	return fsys.FS.ReadFile(filePath)
}

func (f *blockingProgramConfigFS) ReadFile(filePath string) (string, bool) {
	if _, blocks := f.paths[tspath.NormalizePath(filePath)]; !blocks {
		return f.FS.ReadFile(filePath)
	}

	f.mu.Lock()
	f.active++
	if f.active > f.peak {
		f.peak = f.active
	}
	if f.active == f.waitFor {
		f.startOnce.Do(func() { close(f.allStarted) })
	}
	f.mu.Unlock()

	<-f.release
	content, ok := f.FS.ReadFile(filePath)
	f.mu.Lock()
	f.active--
	f.mu.Unlock()
	return content, ok
}

func (f *blockingProgramConfigFS) peakConcurrency() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.peak
}

func (f *targetPlanRealpathCountingFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	f.mu.Lock()
	f.calls[filePath]++
	f.mu.Unlock()
	return f.FS.Realpath(filePath)
}

func (f *targetPlanRealpathCountingFS) callCount(filePath string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[tspath.NormalizePath(filePath)]
}

func writeProgramTestFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
}

func collectProgramTypeDiagnostics(
	t *testing.T,
	programs []*lintprogram.Program,
) []rule.RuleDiagnostic {
	t.Helper()

	var diags []rule.RuleDiagnostic
	_, err := linter.RunLinter(linter.RunLinterOptions{
		TypeCheckOnlyPrograms: programs,
		SingleThreaded:        true,
		TypeCheck:             true,
		Consumer: rule.DiagnosticConsumer{
			Report: func(d rule.RuleDiagnostic) {
				diags = append(diags, d)
			},
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	return diags
}

func containsTSDiagnostic(diags []rule.RuleDiagnostic, code string) bool {
	needle := "TypeScript(" + code + ")"
	for _, d := range diags {
		if d.RuleName == needle {
			return true
		}
	}
	return false
}

func resolveAndBindTestTargets(
	t *testing.T,
	set ProjectSet,
	cfg rslintconfig.RslintConfig,
	dir string,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	buildContext *buildContext,
) (target.Plan, LoadResult) {
	t.Helper()
	plan, err := resolveTargetPlanForTest(nil, cfg, dir, nil, fsys, allowFiles, allowDirs, true)
	if err != nil {
		t.Fatalf("resolveTargetPlanForTest: %v", err)
	}
	binding, err := loadAPIForTest(set, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	return plan, binding
}

func TestTypeCheck_SkipsNoTsconfigSourceOnlyProgram(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"bad.ts": `const bad: number = "oops";
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fs)
	programSet, err := buildProjectsForConfig(
		dir,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		true,
		buildContext,
	)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if len(programSet.compilerPrograms) != 0 {
		t.Fatalf("expected no tsconfig-backed programs, got %d", len(programSet.compilerPrograms))
	}

	_, binding := resolveAndBindTestTargets(
		t,
		programSet,
		rslintconfig.RslintConfig{{Files: []string{"**/*.ts"}}},
		dir,
		fs,
		nil,
		nil,
		buildContext,
	)
	programs := binding.compilerPrograms
	if len(programs) != 1 {
		t.Fatalf("expected one compatibility compiler Program, got %d", len(programs))
	}
	if len(binding.Programs) != 1 || binding.Programs[0].CanProvideProgramDiagnostics() {
		t.Fatalf("expected one source-only Program, got %+v", binding.Programs)
	}
	if got := programs[0].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected compatibility Program to have no ConfigFilePath, got %q", got)
	}
	if !programs[0].Options().NoLib.IsTrue() || !programs[0].Options().NoResolve.IsTrue() {
		t.Fatalf("expected compatibility Program to stay non-project-backed, got options %+v", programs[0].Options())
	}
	if diags := collectProgramTypeDiagnostics(t, binding.Programs); containsTSDiagnostic(diags, "TS2322") {
		t.Fatalf("did not expect semantic diagnostics from a source-only Program: %+v", diags)
	}
}

func TestTypeCheck_TsconfigBackedProgramReportsCoveredDeclarationErrors(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "skipLibCheck": false },
  "include": ["rslint.config.ts"]
}
`,
		"rslint.config.ts": `import type { Bad } from './bad';
export const value: Bad | null = null;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	programSet, err := buildProjectsForConfig(
		dir,
		rslintconfig.RslintConfig{{
			Files: []string{"**/*.ts"},
			LanguageOptions: &rslintconfig.LanguageOptions{
				ParserOptions: &rslintconfig.ParserOptions{
					Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
				},
			},
		}},
		true,
		newBuildContext(fs),
	)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	programs := programSet.compilerPrograms
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programs))
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected tsconfig-backed program to carry ConfigFilePath")
	}
	lintPrograms := lintprogram.NewFromCompilers(programs)
	if !lintPrograms[0].CanProvideProgramDiagnostics() {
		t.Fatal("expected tsconfig-backed Program to expose type-check capability")
	}

	diags := collectProgramTypeDiagnostics(t, lintPrograms)
	if !containsTSDiagnostic(diags, "TS2304") {
		var rendered []string
		for _, d := range diags {
			rendered = append(rendered, d.RuleName+": "+d.Message.Description)
		}
		t.Fatalf("expected TS2304 from the tsconfig-covered declaration graph, got:\n%s", strings.Join(rendered, "\n"))
	}
}

func TestTypeCheck_SkipsSourceOnlyPrograms(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"tsconfig.json": `{
  "compilerOptions": { "skipLibCheck": false },
  "include": ["src/in-project.ts"]
}
`,
		"src/in-project.ts": `export const bad: number = "oops";
`,
		"source-only.ts": `import type { Bad } from './bad';
export const value: Bad | null = null;
`,
		"bad.d.ts": `export type Bad = MissingGlobalType;
`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fs)
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
	}}
	programSet, err := buildProjectsForConfig(
		dir,
		cfg,
		true,
		buildContext,
	)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if len(programSet.compilerPrograms) != 1 {
		t.Fatalf("expected one tsconfig-backed program before source-only Program, got %d", len(programSet.compilerPrograms))
	}
	if !lintprogram.NewFromCompiler(programSet.compilerPrograms[0]).CanProvideProgramDiagnostics() {
		t.Fatal("expected tsconfig-backed Program to participate in type-check")
	}

	_, binding := resolveAndBindTestTargets(
		t,
		programSet,
		cfg,
		dir,
		fs,
		nil,
		nil,
		buildContext,
	)
	programs := binding.compilerPrograms
	if len(programs) != 2 {
		t.Fatalf("expected the source-only Program to be appended, got %d programs", len(programs))
	}
	sourceOnlyPath := filepath.ToSlash(filepath.Join(dir, "source-only.ts"))
	if !slices.Contains(binding.TargetsByProgram[1], sourceOnlyPath) {
		t.Fatalf("expected source-only.ts to target the source-only Program, got %v", binding.TargetsByProgram)
	}
	if got := programs[0].Options().ConfigFilePath; got == "" {
		t.Fatal("expected original tsconfig-backed program to carry ConfigFilePath")
	}
	if got := programs[1].Options().ConfigFilePath; got != "" {
		t.Fatalf("expected source-only Program to have no ConfigFilePath, got %q", got)
	}
	if !programs[1].Options().NoLib.IsTrue() || !programs[1].Options().NoResolve.IsTrue() {
		t.Fatalf("expected source-only Program to stay non-project-backed, got options %+v", programs[1].Options())
	}
	if len(binding.Programs) != 2 ||
		!binding.Programs[0].CanProvideProgramDiagnostics() ||
		binding.Programs[1].CanProvideProgramDiagnostics() {
		t.Fatalf("unexpected Program diagnostic capabilities: %+v", binding.Programs)
	}

	diags := collectProgramTypeDiagnostics(t, binding.Programs)
	if !containsTSDiagnostic(diags, "TS2322") {
		t.Fatalf("expected the tsconfig-backed Program to retain semantic diagnostics: %+v", diags)
	}
	if containsTSDiagnostic(diags, "TS2304") {
		t.Fatalf("did not expect declaration diagnostics from a source-only Program: %+v", diags)
	}
}

func TestLoadProgramsBindsImportedNonRootFile(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       "import { value } from './lib';\nconsole.log(value);\n",
		"lib.ts":        "export const value = 1;\n",
		"tsconfig.json": `{"files": ["main.ts"], "compilerOptions": {"module": "ESNext"}}`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fs)
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	programSet, err := buildProjectsForConfig(dir, cfg, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if len(programSet.compilerPrograms) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programSet.compilerPrograms))
	}
	if _, ok := programSet.configOrders[0][exactPathID(dir)]; !ok {
		t.Fatalf("config order was not keyed by normalized directory identity: %v", programSet.configOrders[0])
	}

	libPath := tspath.NormalizePath(filepath.Join(dir, "lib.ts"))
	plan, binding := resolveAndBindTestTargets(
		t,
		programSet,
		cfg,
		dir,
		fs,
		[]string{libPath},
		nil,
		buildContext,
	)
	programs := binding.compilerPrograms
	targetFiles := []string{plan.Files[0].Path}
	targetsByProgram := binding.TargetsByProgram
	if len(programs) != 1 {
		t.Fatalf("imported non-root target should reuse existing Program, got %d programs", len(programs))
	}
	if len(binding.Programs) != 1 || !binding.Programs[0].CanProvideTypeChecker(binding.Programs[0].GetSourceFile(targetsByProgram[0][0])) {
		t.Fatal("imported non-root target lost its compiler-capable Program")
	}
	if len(targetFiles) != 1 || targetFiles[0] != libPath {
		t.Fatalf("expected lib.ts as the only target, got %v", targetFiles)
	}
	if len(targetsByProgram) != 1 || len(targetsByProgram[0]) != 1 ||
		canonicalPathID(targetsByProgram[0][0], fs) != canonicalPathID(libPath, fs) {
		t.Fatalf("expected lib.ts bound to the tsconfig Program, got %v", targetsByProgram)
	}
}

func TestOrderedProgramIndexesForConfig_NormalizesDirectorySeparators(t *testing.T) {
	set := ProjectSet{
		compilerPrograms: []*compiler.Program{nil},
		configOrders: []configOrders{{
			exactPathID(`C:\Repo`): 0,
		}},
	}
	indexes := orderedProgramIndexesForConfig(set, "C:/Repo")
	if len(indexes) != 1 || indexes[0] != 0 {
		t.Fatalf("normalized config directory did not find its Program: %v", indexes)
	}
	if indexes := orderedProgramIndexesForConfig(set, "c:/repo"); len(indexes) != 0 {
		t.Fatalf("config directory identity unexpectedly folded case: %v", indexes)
	}
}

func TestLoadProgramsBindsRealpathTargetToProgramSourceName(t *testing.T) {
	realDir := t.TempDir()
	linkDir := filepath.Join(filepath.Dir(realDir), filepath.Base(realDir)+"-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	defer os.Remove(linkDir)

	writeProgramTestFiles(t, realDir, map[string]string{
		"src/a.ts":      "export const a = 1;\n",
		"tsconfig.json": `{"include": ["src/a.ts"]}`,
	})

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fs)
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
			},
		},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}

	linkDir = tspath.NormalizePath(linkDir)
	realTarget := tspath.NormalizePath(filepath.Join(realDir, "src/a.ts"))
	programSet, err := buildProjectsForConfig(linkDir, cfg, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	programs := programSet.compilerPrograms
	if len(programs) != 1 {
		t.Fatalf("expected one tsconfig-backed program, got %d", len(programs))
	}

	var sourceName string
	for _, sf := range programs[0].GetSourceFiles() {
		if strings.HasSuffix(sf.FileName(), "/src/a.ts") {
			sourceName = sf.FileName()
			break
		}
	}
	if sourceName == "" {
		t.Fatal("expected program to include src/a.ts")
	}
	if sourceName == realTarget {
		t.Skip("compiler already canonicalized source file to realpath")
	}

	plan, binding := resolveAndBindTestTargets(
		t,
		programSet,
		cfg,
		linkDir,
		fs,
		[]string{realTarget},
		nil,
		buildContext,
	)
	programs = binding.compilerPrograms
	targetFiles := []string{plan.Files[0].Path}
	targetsByProgram := binding.TargetsByProgram
	lintTargetBySourcePath := binding.LintTargetBySourcePath
	if len(programs) != 1 {
		t.Fatalf("realpath target should reuse existing Program, got %d programs", len(programs))
	}
	if len(targetFiles) != 1 || targetFiles[0] != realTarget {
		t.Fatalf("expected realpath target as the only discovered target, got %v", targetFiles)
	}
	if len(targetsByProgram) != 1 || len(targetsByProgram[0]) != 1 || targetsByProgram[0][0] != sourceName {
		t.Fatalf("expected realpath target to bind back to source name %q, got %v", sourceName, targetsByProgram)
	}
	if target := lintTargetBySourcePath[exactPathID(sourceName)]; target.Path != realTarget {
		t.Fatalf("expected source path %q to retain lint target %q, got %+v", sourceName, realTarget, target)
	}
}

func TestLoadProgramsUsesPhysicalConfigSpaceForSymlinkedConfigRoot(t *testing.T) {
	realDir := t.TempDir()
	linkDir := filepath.Join(filepath.Dir(realDir), filepath.Base(realDir)+"-config-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	defer os.Remove(linkDir)
	writeProgramTestFiles(t, realDir, map[string]string{
		"src/a.ts":      "debugger;\n",
		"tsconfig.json": `{"include":["src/a.ts"]}`,
	})

	linkDir = tspath.NormalizePath(linkDir)
	realTarget := tspath.NormalizePath(filepath.Join(realDir, "src/a.ts"))
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"src/**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
			Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
		}},
		Rules: rslintconfig.Rules{"no-debugger": "error"},
	}}
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(linkDir, cfg, true, buildContext)
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("create Program through symlinked config root: err=%v programs=%d", err, len(set.compilerPrograms))
	}
	plan := target.Plan{Files: []target.File{testLintTarget(fsys, linkDir, realTarget)}}
	binding, err := loadAPIForTest(set, plan, linkDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("expected real target to bind to config Program, got %v", binding.TargetsByProgram)
	}
	sourcePath := binding.TargetsByProgram[0][0]
	lintTarget, ok := binding.LintTargetBySourcePath[exactPathID(sourcePath)]
	if !ok {
		t.Fatalf("missing lint target for Program source %q", sourcePath)
	}
	if canonicalPathID(lintTarget.CanonicalPath, fsys) != canonicalPathID(realTarget, fsys) {
		t.Fatalf("binding lost canonical target identity: source=%q binding=%+v target=%q", sourcePath, lintTarget, realTarget)
	}

	resolver := configLint.NewResolver(configLint.ResolverOptions{
		Config:                              cfg,
		ConfigDirectory:                     linkDir,
		TargetsBySourcePath:                 binding.LintTargetBySourcePath,
		SourceMappingsIncludeCanonicalPaths: true,
		Catalog:                             rules.All(),
		PathSpaces:                          rslintconfig.NewPathSpaceSnapshot(map[string]rslintconfig.RslintConfig{linkDir: cfg}, fsys),
		FS:                                  fsys,
	})
	rules := resolver.EnabledRulesForSourcePath(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-debugger" {
		t.Fatalf("expected files selector to match in physical config space, got %v", configuredRuleNameSet(rules))
	}
}

func TestLoadProgramsConfigMatchingDoesNotDependOnProgramSourcePath(t *testing.T) {
	rootDir := t.TempDir()
	writeProgramTestFiles(t, rootDir, map[string]string{
		"physical/index.ts": "console.log('value');\n",
		"tsconfig.json":     `{"files":["physical/index.ts"]}`,
	})
	linkPath := filepath.Join(rootDir, "link.ts")
	physicalPath := filepath.Join(rootDir, "physical/index.ts")
	if err := os.Symlink(physicalPath, linkPath); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}

	rootDir = tspath.NormalizePath(rootDir)
	linkPath = tspath.NormalizePath(linkPath)
	physicalPath = tspath.NormalizePath(physicalPath)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	cfg := rslintconfig.RslintConfig{{
		Files: []string{"link.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{
			Project: rslintconfig.ProjectPaths{"./tsconfig.json"},
		}},
		Rules: rslintconfig.Rules{"no-console": "error"},
	}}
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(rootDir, cfg, true, buildContext)
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("create Program: err=%v programs=%d", err, len(set.compilerPrograms))
	}
	plan := target.Plan{Files: []target.File{testLintTarget(fsys, rootDir, linkPath)}}
	binding, err := loadAPIForTest(set, plan, rootDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("expected lexical target to bind to the physical Program source, got %v", binding.TargetsByProgram)
	}
	sourcePath := binding.TargetsByProgram[0][0]
	expectedSourcePath := authoritativePath(physicalPath, fsys)
	if canonicalPathID(sourcePath, fsys) != canonicalPathID(expectedSourcePath, fsys) {
		t.Fatalf("fixture must bind through physical Program source %q, got %q", expectedSourcePath, sourcePath)
	}
	expectedTargetPath := linkPath
	if target := binding.LintTargetBySourcePath[exactPathID(sourcePath)]; target.Path != expectedTargetPath {
		t.Fatalf("binding must retain lexical target %q, got %+v", expectedTargetPath, target)
	}

	resolver := configLint.NewResolver(configLint.ResolverOptions{
		Config:                              cfg,
		ConfigDirectory:                     rootDir,
		TargetsBySourcePath:                 binding.LintTargetBySourcePath,
		SourceMappingsIncludeCanonicalPaths: true,
		Catalog:                             rules.All(),
		PathSpaces:                          rslintconfig.NewPathSpaceSnapshot(map[string]rslintconfig.RslintConfig{rootDir: cfg}, fsys),
		FS:                                  fsys,
	})
	rules := resolver.EnabledRulesForSourcePath(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("Program membership changed the lexical files match: %v", configuredRuleNameSet(rules))
	}
}

func TestLoadProgramsBindsFileSymlinkOutsideProgramRoot(t *testing.T) {
	sharedDir := t.TempDir()
	writeProgramTestFiles(t, sharedDir, map[string]string{
		"shared.ts": `export const value = 1;`,
	})
	repoDir := t.TempDir()
	linkedPath := filepath.Join(repoDir, "linked.ts")
	realTarget := filepath.Join(sharedDir, "shared.ts")
	if err := os.Symlink(realTarget, linkedPath); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}
	writeProgramTestFiles(t, repoDir, map[string]string{
		"tsconfig.json": `{"files":["linked.ts"]}`,
	})

	repoDir = tspath.NormalizePath(repoDir)
	realTarget = tspath.NormalizePath(realTarget)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	cfg := projectConfig("./tsconfig.json")
	set, err := buildProjectsForConfig(repoDir, cfg, true, buildContext)
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("expected one Program for file-symlink fixture, err=%v programs=%d", err, len(set.compilerPrograms))
	}
	var sourceName string
	for _, sourceFile := range set.compilerPrograms[0].GetSourceFiles() {
		if strings.HasSuffix(sourceFile.FileName(), "/linked.ts") || sourceFile.FileName() == realTarget {
			sourceName = sourceFile.FileName()
			break
		}
	}
	if sourceName == "" {
		t.Fatal("expected Program to contain the symlinked source")
	}
	if sourceName == realTarget {
		t.Skip("compiler canonicalized the file symlink before Program lookup")
	}

	plan := target.Plan{Files: []target.File{testLintTarget(fsys, repoDir, realTarget)}}
	binding, err := loadAPIForTest(set, plan, repoDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("real target should bind through the Program's file symlink, targets=%v", binding.TargetsByProgram)
	}
	if len(binding.TargetsByProgram[0]) != 1 || binding.TargetsByProgram[0][0] != sourceName {
		t.Fatalf("expected target to bind to Program source %q, got %v", sourceName, binding.TargetsByProgram)
	}
	if target := binding.LintTargetBySourcePath[exactPathID(sourceName)]; target.ConfigDirectory != repoDir {
		t.Fatalf("expected bound source owner %q, got %+v", repoDir, target)
	}
}

func testLintTarget(fsys vfs.FS, ownerDir string, filePath string) target.File {
	filePath = tspath.NormalizePath(filePath)
	canonicalPath := filePath
	if realPath := fsys.Realpath(filePath); realPath != "" {
		canonicalPath = tspath.NormalizePath(realPath)
	}
	return target.File{PathIdentity: rslintconfig.PathIdentity{Path: filePath,
		CanonicalPath: canonicalPath}, ConfigDirectory: tspath.NormalizePath(ownerDir),
	}
}

func projectConfig(projects ...string) rslintconfig.RslintConfig {
	return rslintconfig.RslintConfig{{
		LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{
				Project: rslintconfig.ProjectPaths(projects),
			},
		},
	}}
}

func TestLoadProgramsDoesNotBorrowParentConfigProgram(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json":   `{"include":["child/target.ts"]}`,
		"child/target.ts": `export const value = 1;`,
	})

	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	configMap := map[string]rslintconfig.RslintConfig{
		tspath.NormalizePath(rootDir):  projectConfig("./tsconfig.json"),
		tspath.NormalizePath(childDir): rslintconfig.RslintConfig{{}},
	}
	set, err := buildProjectsForConfigs(configMap, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	if len(set.compilerPrograms) != 1 {
		t.Fatalf("expected only the root tsconfig Program, got %d", len(set.compilerPrograms))
	}

	targetPath := filepath.Join(childDir, "target.ts")
	plan := target.Plan{Files: []target.File{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := loadAPIForTest(set, plan, rootDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}

	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram[0]) != 0 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("expected target only in a source-only Program, got targets=%v", binding.TargetsByProgram)
	}
	sourceOnlySource := binding.TargetsByProgram[1][0]
	if target := binding.LintTargetBySourcePath[exactPathID(sourceOnlySource)]; target.ConfigDirectory != tspath.NormalizePath(childDir) {
		t.Fatalf("expected source-only owner %q, got %+v", tspath.NormalizePath(childDir), target)
	}
	if binding.Programs[1].CanProvideTypeChecker(binding.Programs[1].SourceFiles()[0]) {
		t.Fatal("child-owned target unexpectedly received type services")
	}
}

func TestTypeCheckDeduplicatesSyntaxFromSourceOnlyAndParentProgram(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json":   `{"include":["child/target.ts"]}`,
		"child/target.ts": `let value: ;`,
	})

	rootDir = tspath.NormalizePath(rootDir)
	childDir = tspath.NormalizePath(childDir)
	configMap := map[string]rslintconfig.RslintConfig{
		rootDir:  projectConfig("./tsconfig.json"),
		childDir: rslintconfig.RslintConfig{{}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfigs(configMap, true, buildContext)
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	targetPath := tspath.NormalizePath(filepath.Join(childDir, "target.ts"))
	plan := target.Plan{Files: []target.File{testLintTarget(fsys, childDir, targetPath)}}
	binding, err := loadAPIForTest(set, plan, rootDir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("child-owned target must remain source-only, got targets %v", binding.TargetsByProgram)
	}

	diagnostics := collectTargetSyntacticDiagnostics(binding.Programs, binding.TargetsByProgram, true, false)
	if len(diagnostics) != 1 {
		t.Fatalf("expected one malformed source-only lint target, got %v", diagnostics)
	}
	diagnostics = append(diagnostics, collectProgramTypeDiagnostics(t, binding.Programs)...)
	for index := range diagnostics {
		if lintTarget, ok := target.LookupSourceTarget(binding.LintTargetBySourcePath, diagnostics[index].FilePath, fsys); ok {
			diagnostics[index].FilePath = lintTarget.Path
		}
	}
	if len(diagnostics) < 2 {
		t.Fatalf("fixture must exercise both source-only syntax and parent Program type-check paths, got %+v", diagnostics)
	}

	diagnostics = deduplicateTypeScriptDiagnostics(diagnostics, fsys)
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "TypeScript(TS1110)" {
		t.Fatalf("expected one TS1110 diagnostic after cross-phase dedupe, got %+v", diagnostics)
	}
	if diagnostics[0].Origin != rule.DiagnosticOriginTypeScript {
		t.Fatalf("deduplicated TypeScript diagnostic lost its origin: %+v", diagnostics[0])
	}
}

func TestBuildProjectsDeduplicatesSharedTsconfigAndRetainsOwners(t *testing.T) {
	rootDir := t.TempDir()
	childDir := filepath.Join(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"tsconfig.json": `{"include":["src/**/*.ts"]}`,
		"src/a.ts":      `export const a = 1;`,
	})
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}

	rootKey := tspath.NormalizePath(rootDir)
	childKey := tspath.NormalizePath(childDir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	set, err := buildProjectsForConfigs(map[string]rslintconfig.RslintConfig{
		rootKey:  projectConfig("./tsconfig.json"),
		childKey: projectConfig("../tsconfig.json"),
	}, true, newBuildContext(fsys))
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	if len(set.compilerPrograms) != 1 || len(set.configOrders) != 1 {
		t.Fatalf("shared tsconfig must produce one Program, got programs=%d orders=%d", len(set.compilerPrograms), len(set.configOrders))
	}
	if order, ok := set.configOrders[0][exactPathID(rootKey)]; !ok || order != 0 {
		t.Fatalf("missing root config association: %v", set.configOrders[0])
	}
	if order, ok := set.configOrders[0][exactPathID(childKey)]; !ok || order != 0 {
		t.Fatalf("missing child config association: %v", set.configOrders[0])
	}
}

func TestBuildProjectsBoundsParallelBuildsAndPreservesOrder(t *testing.T) {
	rootDir := t.TempDir()
	const (
		testGOMAXPROCS = 9 // Deliberately exceeds the former fixed worker limit.
		configCount    = testGOMAXPROCS + 3
	)
	previousGOMAXPROCS := runtime.GOMAXPROCS(testGOMAXPROCS)
	t.Cleanup(func() {
		runtime.GOMAXPROCS(previousGOMAXPROCS)
	})

	files := make(map[string]string, configCount)
	projects := make([]string, 0, configCount)
	configPaths := make(map[string]struct{}, configCount)
	for index := range configCount {
		name := "tsconfig-" + strconv.Itoa(index) + ".json"
		files[name] = `{"compilerOptions":{"noLib":true},"files":["./shared.ts"]}`
		projects = append(projects, "./"+name)
		configPaths[tspath.NormalizePath(filepath.Join(rootDir, name))] = struct{}{}
	}
	files["shared.ts"] = "export const shared = true;\n"
	writeProgramTestFiles(t, rootDir, files)

	expectedConcurrency := testGOMAXPROCS
	fsys := &blockingProgramConfigFS{
		FS:         bundled.WrapFS(osvfs.FS()),
		paths:      configPaths,
		waitFor:    expectedConcurrency,
		allStarted: make(chan struct{}),
		release:    make(chan struct{}),
	}
	type result struct {
		set ProjectSet
		err error
	}
	done := make(chan result, 1)
	go func() {
		set, err := buildProjectsForConfig(
			tspath.NormalizePath(rootDir),
			projectConfig(projects...),
			false,
			newBuildContext(fsys),
		)
		done <- result{set: set, err: err}
	}()

	select {
	case <-fsys.allStarted:
		close(fsys.release)
	case <-time.After(5 * time.Second):
		close(fsys.release)
		t.Fatalf("Program builds did not reach expected concurrency %d", expectedConcurrency)
	}
	got := <-done
	if got.err != nil {
		t.Fatalf("buildProjectsForConfig: %v", got.err)
	}
	if got := fsys.peakConcurrency(); got != expectedConcurrency {
		t.Fatalf("peak Program build concurrency = %d, want %d", got, expectedConcurrency)
	}
	if len(got.set.compilerPrograms) != configCount {
		t.Fatalf("Programs = %d, want %d", len(got.set.compilerPrograms), configCount)
	}
	for index, program := range got.set.compilerPrograms {
		want := tspath.ResolvePath(rootDir, projects[index])
		if got := tspath.NormalizePath(program.Options().ConfigFilePath); got != want {
			t.Fatalf("Program %d config path = %q, want %q", index, got, want)
		}
	}
}

func TestBuildProjectsSingleThreadedBuildsSerially(t *testing.T) {
	rootDir := t.TempDir()
	files := map[string]string{
		"tsconfig-a.json": `{"compilerOptions":{"noLib":true},"files":[]}`,
		"tsconfig-b.json": `{"compilerOptions":{"noLib":true},"files":[]}`,
	}
	writeProgramTestFiles(t, rootDir, files)
	configPaths := make(map[string]struct{}, len(files))
	for name := range files {
		configPaths[tspath.NormalizePath(filepath.Join(rootDir, name))] = struct{}{}
	}
	fsys := &blockingProgramConfigFS{
		FS:         bundled.WrapFS(osvfs.FS()),
		paths:      configPaths,
		waitFor:    1,
		allStarted: make(chan struct{}),
		release:    make(chan struct{}),
	}
	done := make(chan error, 1)
	go func() {
		_, err := buildProjectsForConfig(
			tspath.NormalizePath(rootDir),
			projectConfig("./tsconfig-a.json", "./tsconfig-b.json"),
			true,
			newBuildContext(fsys),
		)
		done <- err
	}()
	<-fsys.allStarted
	close(fsys.release)
	if err := <-done; err != nil {
		t.Fatalf("buildProjectsForConfig: %v", err)
	}
	if got := fsys.peakConcurrency(); got != 1 {
		t.Fatalf("--singleThreaded peak Program build concurrency = %d, want 1", got)
	}
}

func TestExecuteProjectPlanScopesConcurrentProgramQueries(t *testing.T) {
	t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "")
	previousGOMAXPROCS := runtime.GOMAXPROCS(2)
	t.Cleanup(func() {
		runtime.GOMAXPROCS(previousGOMAXPROCS)
	})

	rootDir := tspath.NormalizePath(t.TempDir())
	tests := []struct {
		name           string
		configCount    int
		singleThreaded bool
		wantDerivedFS  bool
	}{
		{name: "one Program", configCount: 1},
		{name: "parallel Programs", configCount: 2, wantDerivedFS: true},
		{name: "single-threaded Programs", configCount: 2, singleThreaded: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := projectPlan{specs: make([]projectSpec, test.configCount)}
			for index := range plan.specs {
				plan.specs[index] = projectSpec{
					tsconfigPath: tspath.ResolvePath(rootDir, "missing-"+strconv.Itoa(index)+".json"),
					programCwd:   rootDir,
				}
			}
			buildContext := newBuildContext(bundled.WrapFS(osvfs.FS()))
			_, _ = executeProjectPlanForTest(plan, test.singleThreaded, buildContext)

			gotDerivedFS := buildContext.newCompilerHostWithCache(rootDir).FS() != buildContext.FS()
			if gotDerivedFS != test.wantDerivedFS {
				t.Fatalf("compiler host derived FS = %t, want %t", gotDerivedFS, test.wantDerivedFS)
			}
		})
	}
}

func TestExecuteProjectPlanPreservesFirstErrorPrecedence(t *testing.T) {
	rootDir := tspath.NormalizePath(t.TempDir())
	first := tspath.ResolvePath(rootDir, "missing-first.json")
	second := tspath.ResolvePath(rootDir, "missing-second.json")
	plan := projectPlan{
		specs: []projectSpec{
			{tsconfigPath: first, programCwd: rootDir},
			{tsconfigPath: second, programCwd: rootDir},
		},
		terminalErr: os.ErrInvalid,
	}
	_, err := executeProjectPlanForTest(
		plan,
		false,
		newBuildContext(bundled.WrapFS(osvfs.FS())),
	)
	if err == nil || !strings.Contains(err.Error(), first) {
		t.Fatalf("error = %v, want first Program path %q", err, first)
	}
	if strings.Contains(err.Error(), second) || strings.Contains(err.Error(), os.ErrInvalid.Error()) {
		t.Fatalf("later error won over the first Program failure: %v", err)
	}
}

func TestBuildProjectsPreservesSymlinkedTsconfigBase(t *testing.T) {
	rootDir := t.TempDir()
	realDir := filepath.Join(rootDir, "z-real")
	aliasDir := filepath.Join(rootDir, "a-alias")
	writeProgramTestFiles(t, realDir, map[string]string{
		"tsconfig.json": `{"include":["src/**/*.ts"]}`,
		"src/real.ts":   `export const source = "real";`,
	})
	writeProgramTestFiles(t, aliasDir, map[string]string{
		"src/alias.ts": `export const source = "alias";`,
	})
	aliasConfig := filepath.Join(aliasDir, "tsconfig.json")
	if err := os.Symlink(filepath.Join(realDir, "tsconfig.json"), aliasConfig); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	realDir = tspath.NormalizePath(realDir)
	aliasDir = tspath.NormalizePath(aliasDir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	set, err := buildProjectsForConfigs(map[string]rslintconfig.RslintConfig{
		aliasDir: projectConfig("./tsconfig.json"),
		realDir:  projectConfig("./tsconfig.json"),
	}, true, newBuildContext(fsys))
	if err != nil {
		t.Fatalf("buildProjectsForConfigs: %v", err)
	}
	if len(set.compilerPrograms) != 2 {
		t.Fatalf("distinct declared tsconfig paths must produce two Programs, got %d", len(set.compilerPrograms))
	}
	programByConfigPath := make(map[string]*compiler.Program, len(set.compilerPrograms))
	for _, program := range set.compilerPrograms {
		programByConfigPath[exactPathID(program.Options().ConfigFilePath)] = program
	}
	aliasConfigPath := tspath.ResolvePath(aliasDir, "tsconfig.json")
	realConfigPath := tspath.ResolvePath(realDir, "tsconfig.json")
	aliasProgram := programByConfigPath[exactPathID(aliasConfigPath)]
	realProgram := programByConfigPath[exactPathID(realConfigPath)]
	if aliasProgram == nil || realProgram == nil {
		t.Fatalf("missing lexical tsconfig Programs: %v", programByConfigPath)
		return
	}
	aliasSource := tspath.ResolvePath(aliasDir, "src/alias.ts")
	realSource := tspath.ResolvePath(realDir, "src/real.ts")
	if aliasProgram.GetSourceFile(aliasSource) == nil || aliasProgram.GetSourceFile(realSource) != nil {
		t.Fatalf("symlinked tsconfig must resolve includes from %q", aliasDir)
	}
	if realProgram.GetSourceFile(realSource) == nil || realProgram.GetSourceFile(aliasSource) != nil {
		t.Fatalf("real tsconfig must resolve includes from %q", realDir)
	}
}

func TestLoadProgramsUsesGoverningConfigProjectOrder(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"shared.ts":       `export const value = 1;`,
		"tsconfig-a.json": `{"files":["shared.ts"]}`,
		"tsconfig-b.json": `{"files":["shared.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(
		dir,
		projectConfig("./tsconfig-a.json", "./tsconfig-b.json"),
		true,
		buildContext,
	)
	if err != nil || len(set.compilerPrograms) != 2 {
		t.Fatalf("expected two ordered Programs, err=%v programs=%d", err, len(set.compilerPrograms))
	}

	targetPath := filepath.Join(dir, "shared.ts")
	plan := target.Plan{Files: []target.File{testLintTarget(fsys, dir, targetPath)}}
	binding, err := loadAPIForTest(set, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.TargetsByProgram[0]) != 1 || len(binding.TargetsByProgram[1]) != 0 {
		t.Fatalf("overlapping target must bind to the first declared project, got %v", binding.TargetsByProgram)
	}
	targetedContext := newBuildContext(fsys)
	targeted, err := sessionForTest(targetedContext).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-a.json", "./tsconfig-b.json")}, Targets: plan, Scope: LintTargets, SingleThreaded: true})
	if err != nil {
		t.Fatalf("BuildTargetProject: %v", err)
	}
	if targeted.Len() != 1 || tspath.NormalizePath(targeted.compilerPrograms[0].Options().ConfigFilePath) != tspath.ResolvePath(dir, "tsconfig-a.json") {
		t.Fatalf("targeted overlap did not retain only the first direct project")
	}
}

func TestLoadProgramsPrefersLaterDirectRootOverEarlierImport(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"target.ts":            `export const target = 1;`,
		"implicit-main.ts":     `import "./target";`,
		"unrelated-main.ts":    `export const unrelated = 1;`,
		"tsconfig-import.json": `{"files":["implicit-main.ts"]}`,
		"tsconfig-direct.json": `{"files":["target.ts"]}`,
		"tsconfig-later.json":  `{"files":["unrelated-main.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	readCounter := &programReadCountingFS{
		FS:    bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		reads: make(map[string]int),
	}
	fsys := vfs.FS(readCounter)
	context := newBuildContext(fsys)
	config := projectConfig(
		"./tsconfig-import.json",
		"./tsconfig-direct.json",
		"./tsconfig-later.json",
	)
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "target.ts")),
	}}

	set, err := sessionForTest(context).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: LintTargets, SingleThreaded: true})
	if err != nil {
		t.Fatalf("BuildTargetProject: %v", err)
	}
	if set.Len() != 1 {
		t.Fatalf("targeted build produced %d Programs, want one direct project", set.Len())
	}
	wantConfig := tspath.ResolvePath(dir, "tsconfig-direct.json")
	if got := tspath.NormalizePath(set.compilerPrograms[0].Options().ConfigFilePath); got != wantConfig {
		t.Fatalf("selected project = %q, want direct project %q", got, wantConfig)
	}
	implicitRoot := tspath.ResolvePath(dir, "implicit-main.ts")
	laterConfig := tspath.ResolvePath(dir, "tsconfig-later.json")
	if got := readCounter.readCount(implicitRoot); got != 0 {
		t.Fatalf("earlier import-only project source was read %d time(s)", got)
	}
	if got := readCounter.readCount(laterConfig); got != 0 {
		t.Fatal("project after the direct winner was unnecessarily read")
	}
	if got := readCounter.readCount(tspath.ResolvePath(dir, "unrelated-main.ts")); got != 0 {
		t.Fatalf("project after the direct winner read its source %d time(s)", got)
	}
	binding, err := sessionForTest(context).LoadAPI(set, plan, dir, true)
	if err != nil {
		t.Fatalf("LoadAPI: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("direct target was not bound to its selected project: %v", binding.TargetsByProgram)
	}

	// The same ownership rule must also hold when broad loading has already
	// materialized every configured Program.
	broadContext := newBuildContext(fsys)
	all, err := sessionForTest(broadContext).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Scope: AllDeclared, SingleThreaded: true})
	if err != nil {
		t.Fatalf("BuildProject: %v", err)
	}
	broadBinding, err := sessionForTest(broadContext).LoadAPI(all, plan, dir, true)
	if err != nil {
		t.Fatalf("broad LoadAPI: %v", err)
	}
	if len(broadBinding.TargetsByProgram) != 3 || len(broadBinding.TargetsByProgram[1]) != 1 {
		t.Fatalf("broad binding did not prefer the direct project: %v", broadBinding.TargetsByProgram)
	}
}

func TestBuildTargetProjectPrefersEarlierDirectRoot(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"nested/target.ts":     `export const target = 1;`,
		"tsconfig-first.json":  `{"files":["nested/target.ts"],"compilerOptions":{"noLib":true}}`,
		"nested/tsconfig.json": `{"files":["target.ts"],"compilerOptions":{"noLib":true}}`,
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	context := newBuildContext(fsys)
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "nested/target.ts")),
	}}

	set, err := sessionForTest(context).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-first.json", "./nested/tsconfig.json")}, Targets: plan, Scope: LintTargets, SingleThreaded: false})
	if err != nil {
		t.Fatalf("BuildTargetProject: %v", err)
	}
	if set.Len() != 1 {
		t.Fatalf("selected Programs = %d, want one", set.Len())
	}
	wantConfig := tspath.ResolvePath(dir, "tsconfig-first.json")
	if got := tspath.NormalizePath(set.compilerPrograms[0].Options().ConfigFilePath); got != wantConfig {
		t.Fatalf("nested project overrode declaration order: got %q, want %q", got, wantConfig)
	}
}

func TestBuildTargetProjectStopsProgramLoadingAfterSelectedRoot(t *testing.T) {
	for _, disableCache := range []bool{false, true} {
		for _, singleThreaded := range []bool{false, true} {
			t.Run(fmt.Sprintf("cache-disabled=%t/serial=%t", disableCache, singleThreaded), func(t *testing.T) {
				t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "")
				if disableCache {
					t.Setenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE", "1")
				}
				dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_root_prefix.txtar").Materialize(t, ""))
				fsys := &programReadCountingFS{FS: bundled.WrapFS(osvfs.FS()), reads: make(map[string]int)}
				plan := target.Plan{Files: []target.File{testLintTarget(fsys, dir, tspath.ResolvePath(dir, "target.ts"))}}
				set, err := NewSession(fsys).buildProjectsForTest(t, ProjectBuildRequest{
					Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig(
						"./tsconfig.json", "./later/tsconfig.json", "./empty.json", "./invalid.json",
					)},
					Targets: plan, Scope: LintTargets, SingleThreaded: singleThreaded,
				})
				if err != nil || set.Len() != 1 {
					t.Fatalf("selected project: count=%d err=%v", set.Len(), err)
				}
				for _, config := range []string{"tsconfig.json", "later/tsconfig.json", "empty.json", "invalid.json"} {
					want := 0
					if config == "tsconfig.json" {
						want = 1
					}
					if got := fsys.readCount(tspath.ResolvePath(dir, config)); got != want {
						t.Fatalf("config %q read %d times, want %d", config, got, want)
					}
				}
				if fsys.readCount(tspath.ResolvePath(dir, "later/base.json")) != 0 ||
					fsys.directoryCount(tspath.ResolvePath(dir, "later/src")) != 0 {
					t.Fatal("a later candidate expanded includes or extended configs after an exact root was found")
				}
				if fsys.readCount(tspath.ResolvePath(dir, "later/src/other.ts")) != 0 {
					t.Fatal("an unselected project constructed a source graph")
				}
			})
		}
	}
}

func TestBuildTargetProjectReportsOnlyReachedConfigErrors(t *testing.T) {
	for _, test := range []struct {
		name        string
		projects    []string
		targets     []string
		unreadable  string
		alias       bool
		allDeclared bool
		wantConfig  string
		wantError   bool
		wantReads   int
	}{
		{name: "direct root skips unreadable later config", projects: []string{"./tsconfig.json", "./later/tsconfig.json"}, targets: []string{"target.ts"}, unreadable: "later/tsconfig.json", wantConfig: "tsconfig.json"},
		{name: "physical root skips unreadable later config", projects: []string{"./tsconfig.json", "./later/tsconfig.json"}, targets: []string{"alias.ts"}, unreadable: "later/tsconfig.json", alias: true, wantConfig: "tsconfig.json"},
		{name: "missed directory hint stops at later direct root", projects: []string{"./import.json", "./tsconfig.json", "./later/tsconfig.json"}, targets: []string{"target.ts"}, unreadable: "later/tsconfig.json", wantConfig: "tsconfig.json"},
		{name: "imported target must read later candidate", projects: []string{"./import.json", "./tsconfig.json"}, targets: []string{"target.ts"}, unreadable: "tsconfig.json", wantError: true, wantReads: 1},
		{name: "another target needs later candidate", projects: []string{"./tsconfig.json", "./later/tsconfig.json"}, targets: []string{"target.ts", "later/src/other.ts"}, unreadable: "later/tsconfig.json", wantError: true, wantReads: 1},
		{name: "first candidate is unreadable", projects: []string{"./tsconfig.json", "./later/tsconfig.json"}, targets: []string{"target.ts"}, unreadable: "tsconfig.json", wantError: true, wantReads: 1},
		{name: "program-wide checking still reads all declarations", projects: []string{"./tsconfig.json", "./later/tsconfig.json"}, targets: []string{"target.ts"}, unreadable: "later/tsconfig.json", allDeclared: true, wantError: true, wantReads: 1},
	} {
		for _, singleThreaded := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/serial=%t", test.name, singleThreaded), func(t *testing.T) {
				dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_root_prefix.txtar").Materialize(t, ""))
				if test.alias {
					if err := os.Symlink(tspath.ResolvePath(dir, "target.ts"), tspath.ResolvePath(dir, "alias.ts")); err != nil {
						t.Skipf("symlinks unavailable: %v", err)
					}
				}
				configPath := tspath.ResolvePath(dir, test.unreadable)
				fsys := &programReadCountingFS{
					FS:    &unreadableProjectConfigFS{FS: bundled.WrapFS(osvfs.FS()), configPath: configPath},
					reads: make(map[string]int),
				}
				var files []target.File
				for _, name := range test.targets {
					files = append(files, testLintTarget(fsys, dir, tspath.ResolvePath(dir, name)))
				}
				scope := LintTargets
				if test.allDeclared {
					scope = AllDeclared
				}
				projects, err := NewSession(fsys).buildProjectsForTest(t, ProjectBuildRequest{
					Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig(test.projects...)},
					Targets: target.Plan{Files: files}, Scope: scope, SingleThreaded: singleThreaded,
				})
				if test.wantError {
					if err == nil || !strings.Contains(err.Error(), test.unreadable) {
						t.Fatalf("expected reached config error for %q, got %v", test.unreadable, err)
					}
				} else {
					if err != nil || projects.Len() != 1 {
						t.Fatalf("selected project: count=%d err=%v", projects.Len(), err)
					}
					if got := tspath.NormalizePath(projects.compilerPrograms[0].Options().ConfigFilePath); got != tspath.ResolvePath(dir, test.wantConfig) {
						t.Fatalf("selected %q, want %q", got, test.wantConfig)
					}
				}
				if got := fsys.readCount(configPath); got != test.wantReads {
					t.Fatalf("unreadable config read %d times, want %d", got, test.wantReads)
				}
			})
		}
	}
}

func TestBuildProjectsReportsPathResolutionBeforeTargetSelection(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	// Path resolution is already unsuccessful. Target selection must return
	// that error without reading configs solely to replace it with another one.
	plan := projectPlan{
		specs:       []projectSpec{{tsconfigPath: tspath.ResolvePath(dir, "removed.json"), programCwd: dir}},
		terminalErr: os.ErrNotExist,
	}
	for _, singleThreaded := range []bool{false, true} {
		fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
		_, eagerError := NewSession(fsys).executeProjectPlan(plan, singleThreaded)
		_, err := NewSession(fsys).executeTargetProjectPlan(plan, ProjectBuildRequest{
			Scope: LintTargets, SingleThreaded: singleThreaded,
		})
		if !errors.Is(err, plan.terminalErr) {
			t.Fatalf("target selection did not preserve path-resolution error: %v", err)
		}
		if eagerError == nil || !strings.Contains(eagerError.Error(), "removed.json") {
			t.Fatalf("program-wide validation changed: %v", eagerError)
		}
	}
}

func TestSelectProjectSourcesKeepsAuthoredEligibilityWithAdaptedProgram(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"keep.ts":       "export const keep = 1;",
		"target.js":     "export const target = 1;",
		"tsconfig.json": `{"files":["keep.ts","target.js"],"compilerOptions":{"noLib":true,"allowJs":false}}`,
	})
	fsys := bundled.WrapFS(osvfs.FS())
	context := newBuildContext(fsys)
	configPath := tspath.ResolvePath(dir, "tsconfig.json")
	parsed, err := context.parseConfig(dir, configPath)
	if err != nil || parsed == nil {
		t.Fatalf("parse config: %v", err)
	}
	targets := []target.File{
		testLintTarget(fsys, dir, tspath.ResolvePath(dir, "keep.ts")),
		testLintTarget(fsys, dir, tspath.ResolvePath(dir, "target.js")),
	}
	indexes := []int{0}
	for _, service := range []bool{false, true} {
		var adapted *compiler.Program
		selected, err := SelectProjectSources(ProjectSelectionRequest{
			Targets: targets, CandidateIndexes: [][]int{indexes, indexes},
			Candidates: []ProjectCandidate{{ConfigPath: configPath, SourceReferences: service}},
			FS:         fsys, SingleThreaded: true,
			Metadata: func(int) (*tsoptions.ParsedCommandLine, error) { return parsed, nil },
			Program: func(_ int, metadata *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
				var err error
				adapted, err = context.createProjectProgramFromParsedConfig(true, dir, metadata, true)
				return adapted, err
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if adapted == nil || adapted.GetSourceFile(targets[1].Path) == nil {
			t.Fatal("fixture must provide JS through the adapter's internal options")
		}
		if selected[0].SourceFile == nil || (selected[1].SourceFile != nil) != service {
			t.Fatalf("service=%t: selection borrowed an adapter's broader eligibility: %+v", service, selected)
		}
	}
}

func TestSelectProjectSourcesOverlapsConfirmedBuildsWithRequiredMetadata(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts":   "export const a = 1;",
		"b.ts":   "export const b = 1;",
		"a.json": `{"files":["a.ts"],"compilerOptions":{"noLib":true}}`,
		"b.json": `{"files":["b.ts"],"compilerOptions":{"noLib":true}}`,
	})
	for _, test := range []struct {
		name          string
		serial        bool
		metadataError bool
	}{
		{name: "parallel"},
		{name: "serial", serial: true},
		{name: "metadata error wins over early build error", metadataError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			serial := test.serial
			if !serial && runtime.GOMAXPROCS(0) < 2 {
				t.Skip("requires at least two Program workers")
			}
			fsys := bundled.WrapFS(osvfs.FS())
			context := newBuildContext(fsys)
			var candidates []ProjectCandidate
			var metadata []*tsoptions.ParsedCommandLine
			var targets []target.File
			for _, name := range []string{"a", "b"} {
				configPath := tspath.ResolvePath(dir, name+".json")
				parsed, err := context.parseConfig(dir, configPath)
				if err != nil {
					t.Fatal(err)
				}
				candidates = append(candidates, ProjectCandidate{ConfigPath: configPath})
				metadata = append(metadata, parsed)
				targets = append(targets, testLintTarget(fsys, dir, tspath.ResolvePath(dir, name+".ts")))
			}
			started := make(chan struct{})
			indexes := []int{0, 1}
			metadataError := errors.New("required later metadata failed")
			buildError := errors.New("earlier confirmed Program failed")
			selected, err := SelectProjectSources(ProjectSelectionRequest{
				Targets: targets, CandidateIndexes: [][]int{indexes, indexes},
				Candidates: candidates, FS: fsys, SingleThreaded: serial,
				Metadata: func(index int) (*tsoptions.ParsedCommandLine, error) {
					if index == 1 {
						if serial {
							select {
							case <-started:
								return nil, errors.New("serial build started before root selection completed")
							default:
							}
						} else {
							select {
							case <-started:
							case <-time.After(5 * time.Second):
								return nil, errors.New("confirmed Program waited for later metadata")
							}
						}
						if test.metadataError {
							return nil, metadataError
						}
					}
					return metadata[index], nil
				},
				Program: func(index int, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
					if index == 0 {
						close(started)
						if test.metadataError {
							return nil, buildError
						}
					}
					return context.createProjectProgramFromParsedConfig(serial, dir, parsed, false)
				},
			})
			if test.metadataError {
				if !errors.Is(err, metadataError) {
					t.Fatalf("metadata error lost priority to early build: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for index, result := range selected {
				if result.CandidateIndex != index || result.SourceFile == nil {
					t.Fatalf("target %d lost its selected Program: %+v", index, result)
				}
			}
		})
	}
}

func TestProjectMetadataPrefetchPreservesDirectoryHints(t *testing.T) {
	for _, test := range []struct {
		name            string
		directories     []string
		targets         []string
		indexes         []int
		caseInsensitive bool
		want            []int
	}{
		{
			name:        "first candidate is the closest directory",
			directories: []string{"/repo/deep", "/other", "/repo"}, targets: []string{"/repo/deep/file.ts"},
			want: []int{0},
		},
		{
			name:        "same directory keeps first declaration",
			directories: []string{"/other", "/repo", "/repo", "/unused"}, targets: []string{"/repo/file.ts"},
			want: []int{0, 1},
		},
		{
			name:        "longer directory later in the declaration",
			directories: []string{"/other", "/repo", "/repo/deep", "/unused"}, targets: []string{"/repo/deep/file.ts"}, want: []int{0, 1, 2},
		},
		{
			name:        "directory boundary is not a string prefix",
			directories: []string{"/other", "/repo/src", "/unused", "/repo/src2"}, targets: []string{"/repo/src2/file.ts"}, want: []int{0, 1, 2, 3},
		},
		{
			name:        "hint position differs from global candidate index",
			directories: []string{"/repo/b", "/unused", "/other", "/repo/a"}, targets: []string{"/repo/a/file.ts"}, indexes: []int{2, 0, 3, 1}, want: []int{2, 0, 3},
		},
		{
			name:        "all targets contribute to the prefix",
			directories: []string{"/other", "/repo", "/repo/a", "/repo/b"}, targets: []string{"/repo/a/file.ts", "/repo/b/file.ts", "/elsewhere/file.ts"}, want: []int{0, 1, 2, 3},
		},
		{
			name:        "case sensitive components",
			directories: []string{"/other", "/repo/a", "/repo/A", "/unused"}, targets: []string{"/repo/A/file.ts"}, want: []int{0, 1, 2},
		},
		{
			name:        "case insensitive ties retain declaration order",
			directories: []string{"/other", "/repo/a", "/repo/A", "/unused"}, targets: []string{"/repo/A/file.ts"}, caseInsensitive: true,
			want: []int{0, 1},
		},
		{
			name:        "Unicode folding retains original directory byte length",
			directories: []string{"/other", "/repo/k", "/repo/K", "/unused"}, targets: []string{"/repo/k/file.ts"}, caseInsensitive: true, want: []int{0, 1, 2},
		},
		{
			name:        "drive root ignores case on a case sensitive filesystem",
			directories: []string{"D:/other", "C:/repo", "c:/repo/deep", "C:/unused"}, targets: []string{"C:/repo/deep/file.ts"}, want: []int{0, 1, 2},
		},
		{
			name:        "UNC root and component case differ",
			directories: []string{"//SERVER/Other", "//server/share", "//SERVER/Share", "//server/unused"}, targets: []string{"//server/Share/file.ts"}, want: []int{0, 1, 2},
		},
		{
			name:        "URL root comparison is retained",
			directories: []string{"https://other/repo", "https://host/repo", "https://HOST/repo/deep", "https://host/unused"}, targets: []string{"https://host/repo/deep/file.ts"}, want: []int{0, 1, 2},
		},
		{
			name:        "unreduced directory length is retained",
			directories: []string{"/other", "/repo/b", "/repo/a/../b", "/unused"}, targets: []string{"/repo/b/file.ts"}, want: []int{0, 1, 2},
		},
		{
			name:        "two candidates can read together",
			directories: []string{"/other", "/repo"}, targets: []string{"/repo/file.ts"},
			want: []int{0, 1},
		},
		{
			name:        "no containing directory",
			directories: []string{"/other", "/repo/a", "/repo/b"}, targets: []string{"/elsewhere/file.ts"},
			want: []int{0},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			candidates := make([]ProjectCandidate, len(test.directories))
			indexes := slices.Clone(test.indexes)
			if indexes == nil {
				indexes = make([]int, len(candidates))
				for index := range indexes {
					indexes[index] = index
				}
			}
			originalIndexes := slices.Clone(indexes)
			for index, directory := range test.directories {
				candidates[index].ConfigPath = fmt.Sprintf("%s/tsconfig-%d.json", directory, index)
			}
			targets := make([]target.File, len(test.targets))
			pending := make([]int, len(targets))
			for index, filePath := range test.targets {
				targets[index] = target.File{PathIdentity: rslintconfig.PathIdentity{Path: filePath}}
				pending[index] = index
			}
			selection := projectSelection{request: ProjectSelectionRequest{
				Candidates: candidates, Targets: targets,
				FS: &bindingIndexTestFS{caseSensitive: !test.caseInsensitive},
			}}
			got := selection.metadataPrefetchPrefix(projectTargetGroup{projectIndexes: indexes, targetIndexes: pending})
			if !slices.Equal(got, test.want) {
				t.Fatalf("prefetch prefix %v, want %v", got, test.want)
			}
			if !slices.Equal(indexes, originalIndexes) {
				t.Fatalf("prediction reordered the candidate list: %v", indexes)
			}
		})
	}
}

func TestSelectProjectSourcesBoundsAndJoinsMetadataAcrossGroups(t *testing.T) {
	previous := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previous)
	dir := tspath.NormalizePath(t.TempDir())
	for _, serial := range []bool{false, true} {
		t.Run(fmt.Sprintf("serial=%t", serial), func(t *testing.T) {
			const count = 6
			candidates := make([]ProjectCandidate, count)
			failures := make([]error, count)
			calls := make([]atomic.Int32, count)
			for index := range candidates {
				candidates[index].ConfigPath = tspath.ResolvePath(dir, fmt.Sprintf("%d.json", index))
				failures[index] = fmt.Errorf("metadata %d failed", index)
			}
			candidates[0].ConfigPath = tspath.ResolvePath(dir, "nested/0.json")
			var targets []target.File
			for _, name := range []string{"nested/a", "b", "c"} {
				targets = append(targets, target.File{
					PathIdentity:    rslintconfig.PathIdentity{Path: tspath.ResolvePath(dir, name+".ts")},
					ConfigDirectory: dir,
				})
			}
			started := make(chan int, count)
			release := make(chan struct{})
			var active, peak atomic.Int32
			done := make(chan error, 1)
			go func() {
				_, err := SelectProjectSources(ProjectSelectionRequest{
					Targets: targets, Candidates: candidates,
					CandidateIndexes: [][]int{{3, 0, 2}, {1, 0, 2}, {2, 1}},
					SingleThreaded:   serial,
					Metadata: func(index int) (*tsoptions.ParsedCommandLine, error) {
						calls[index].Add(1)
						n := active.Add(1)
						for old := peak.Load(); n > old && !peak.CompareAndSwap(old, n); old = peak.Load() {
						}
						started <- index
						<-release
						active.Add(-1)
						return nil, failures[index]
					},
				})
				done <- err
			}()
			workers := 2
			if serial {
				workers = 1
			}
			for range workers {
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					close(release)
					<-done
					t.Fatal("metadata workers did not overlap")
				}
			}
			close(release)
			if err := <-done; !errors.Is(err, failures[3]) {
				t.Fatalf("completion order changed the first group's error: %v", err)
			}
			if got := peak.Load(); got != int32(workers) || active.Load() != 0 {
				t.Fatalf("metadata workers: peak=%d active=%d want peak=%d and all joined", got, active.Load(), workers)
			}
			for index := range candidates {
				want := int32(0)
				if index == 3 || (!serial && index < 4) {
					want = 1
				}
				if got := calls[index].Load(); got != want {
					t.Fatalf("candidate %d parsed %d times, want %d", index, got, want)
				}
			}
		})
	}
}

func TestSelectProjectSourcesBoundsOnDemandMetadataAlongsidePrefetch(t *testing.T) {
	previous := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previous)
	dir := tspath.NormalizePath(t.TempDir())
	started := make(chan int, 3)
	release := make(chan struct{})
	var active, peak atomic.Int32
	selection := projectSelection{
		request: ProjectSelectionRequest{
			Targets: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: tspath.ResolvePath(dir, "nested/target.ts")}}},
			Candidates: []ProjectCandidate{
				{ConfigPath: tspath.ResolvePath(dir, "first.json")},
				{ConfigPath: tspath.ResolvePath(dir, "nested/likely.json")},
				{ConfigPath: tspath.ResolvePath(dir, "elsewhere/later.json")},
			},
			Metadata: func(index int) (*tsoptions.ParsedCommandLine, error) {
				n := active.Add(1)
				for old := peak.Load(); n > old && !peak.CompareAndSwap(old, n); old = peak.Load() {
				}
				started <- index
				<-release
				active.Add(-1)
				return nil, errors.New("metadata is unavailable")
			},
		},
		slots: make([]projectSelectionSlot, 3),
	}
	wait := selection.prefetchMetadata([]projectTargetGroup{{projectIndexes: []int{0, 1, 2}, targetIndexes: []int{0}}})
	if wait == nil {
		t.Fatal("two predicted configs did not start prefetch")
	}
	for range 2 {
		select {
		case index := <-started:
			if index == 2 {
				t.Error("prefetch read beyond the predicted prefix")
			}
		case <-time.After(5 * time.Second):
			close(release)
			wait()
			t.Fatal("prefetch workers did not overlap")
		}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = selection.metadata(2)
	}()
	select {
	case <-started:
		t.Error("an on-demand read exceeded the request's metadata concurrency limit")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	wait()
	<-done
	if peak.Load() != 2 || active.Load() != 0 {
		t.Fatalf("metadata workers: peak=%d active=%d, want 2 and 0", peak.Load(), active.Load())
	}
}

func TestSelectProjectSourcesPrefetchesMetadataWithoutSelectingItsErrors(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"first.json":       `{"files":[],"compilerOptions":{"noLib":true}}`,
		"owner.json":       `{"files":["nested/target.ts"],"compilerOptions":{"noLib":true}}`,
		"unused.json":      `{"files":[],"compilerOptions":{"noLib":true}}`,
		"nested/target.ts": "export const value = 1;",
	})
	for _, serial := range []bool{false, true} {
		t.Run(fmt.Sprintf("serial=%t", serial), func(t *testing.T) {
			if !serial && runtime.GOMAXPROCS(0) < 2 {
				t.Skip("requires concurrent metadata workers")
			}
			fsys := bundled.WrapFS(osvfs.FS())
			context := newBuildContext(fsys)
			candidates := []ProjectCandidate{
				{ConfigPath: tspath.ResolvePath(dir, "first.json")},
				{ConfigPath: tspath.ResolvePath(dir, "owner.json")},
				{ConfigPath: tspath.ResolvePath(dir, "unused.json")},
				{ConfigPath: tspath.ResolvePath(dir, "nested/unreadable.json")},
			}
			var metadata []*tsoptions.ParsedCommandLine
			for _, candidate := range candidates[:3] {
				parsed, err := context.parseConfig(dir, candidate.ConfigPath)
				if err != nil {
					t.Fatal(err)
				}
				metadata = append(metadata, parsed)
			}
			prefetched := make(chan struct{})
			programStarted := make(chan struct{})
			var prefetchWaitTimedOut atomic.Bool
			programCalls := make([]atomic.Int32, len(candidates))
			selected, err := SelectProjectSources(ProjectSelectionRequest{
				Targets:          []target.File{testLintTarget(fsys, dir, tspath.ResolvePath(dir, "nested/target.ts"))},
				CandidateIndexes: [][]int{{0, 1, 2, 3}}, Candidates: candidates, FS: fsys, SingleThreaded: serial,
				Metadata: func(index int) (*tsoptions.ParsedCommandLine, error) {
					if index == 3 {
						close(prefetched)
						select {
						case <-programStarted:
						case <-time.After(5 * time.Second):
							prefetchWaitTimedOut.Store(true)
						}
						return nil, errors.New("unused prefetched metadata failed")
					}
					if index == 0 && !serial {
						select {
						case <-prefetched:
						case <-time.After(5 * time.Second):
							return nil, errors.New("first metadata blocked parallel prefetch")
						}
					}
					return metadata[index], nil
				},
				Program: func(index int, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
					programCalls[index].Add(1)
					if index < len(metadata) && parsed != metadata[index] {
						return nil, errors.New("Program did not reuse the prefetched metadata")
					}
					if index != 1 {
						return nil, fmt.Errorf("unselected candidate %d acquired a Program", index)
					}
					close(programStarted)
					return context.createProjectProgramFromParsedConfig(serial, dir, parsed, false)
				},
			})
			if err != nil || len(selected) != 1 || selected[0].CandidateIndex != 1 || selected[0].SourceFile == nil {
				t.Fatalf("prefetch changed ordered selection: selected=%+v err=%v", selected, err)
			}
			if prefetchWaitTimedOut.Load() {
				t.Fatal("confirmed Program waited for an unused prefetched config")
			}
			for index := range candidates {
				want := int32(0)
				if index == 1 {
					want = 1
				}
				if got := programCalls[index].Load(); got != want {
					t.Fatalf("candidate %d acquired %d Programs, want %d", index, got, want)
				}
			}
			if serial {
				select {
				case <-prefetched:
					t.Fatal("serial selection read past its winner")
				default:
				}
			}
		})
	}
}

type projectSelectionRealpathFS struct {
	vfs.FS
	beforeRealpath func(string)
}

func (fsys *projectSelectionRealpathFS) Realpath(filePath string) string {
	fsys.beforeRealpath(tspath.NormalizePath(filePath))
	return fsys.FS.Realpath(filePath)
}

func TestSelectProjectSourcesOverlapsConfirmedBuildWithRemainingRootIdentities(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts":     "export const a = 1;",
		"b.ts":     "export const b = 1;",
		"extra.ts": "export const extra = 1;",
		"a.json":   `{"files":["a.ts","extra.ts"],"compilerOptions":{"noLib":true}}`,
		"b.json":   `{"files":["b.ts"],"compilerOptions":{"noLib":true}}`,
	})
	for _, serial := range []bool{false, true} {
		t.Run(fmt.Sprintf("serial=%t", serial), func(t *testing.T) {
			if !serial && runtime.GOMAXPROCS(0) < 2 {
				t.Skip("requires concurrent Program workers")
			}
			baseFS := bundled.WrapFS(osvfs.FS())
			context := newBuildContext(baseFS)
			var candidates []ProjectCandidate
			var metadata []*tsoptions.ParsedCommandLine
			var targets []target.File
			for _, name := range []string{"a", "b"} {
				configPath := tspath.ResolvePath(dir, name+".json")
				parsed, err := context.parseConfig(dir, configPath)
				if err != nil {
					t.Fatal(err)
				}
				candidates = append(candidates, ProjectCandidate{ConfigPath: configPath})
				metadata = append(metadata, parsed)
				targets = append(targets, testLintTarget(baseFS, dir, tspath.ResolvePath(dir, name+".ts")))
			}
			started := make(chan struct{})
			var checkedIdentity, wrongOrder atomic.Bool
			fsys := &projectSelectionRealpathFS{FS: baseFS, beforeRealpath: func(filePath string) {
				if filePath != tspath.ResolvePath(dir, "extra.ts") {
					return
				}
				checkedIdentity.Store(true)
				if serial {
					select {
					case <-started:
						wrongOrder.Store(true)
					default:
					}
					return
				}
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					wrongOrder.Store(true)
				}
			}}
			indexes := []int{0, 1}
			calls := make([]atomic.Int32, len(candidates))
			selected, err := SelectProjectSources(ProjectSelectionRequest{
				Targets: targets, CandidateIndexes: [][]int{indexes, indexes},
				Candidates: candidates, FS: fsys, SingleThreaded: serial,
				Metadata: func(index int) (*tsoptions.ParsedCommandLine, error) { return metadata[index], nil },
				Program: func(index int, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error) {
					calls[index].Add(1)
					if index == 0 {
						close(started)
					}
					return context.createProjectProgramFromParsedConfig(serial, dir, parsed, false)
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if !checkedIdentity.Load() || wrongOrder.Load() {
				t.Fatalf("confirmed build/root lookup order: checked=%t wrong=%t", checkedIdentity.Load(), wrongOrder.Load())
			}
			if len(selected) != len(targets) {
				t.Fatalf("selected %d targets, want %d", len(selected), len(targets))
			}
			for index, result := range selected {
				if result.CandidateIndex != index || result.SourceFile == nil || calls[index].Load() != 1 {
					t.Fatalf("target %d: selection=%+v Program calls=%d", index, result, calls[index].Load())
				}
			}
		})
	}
}

func TestSelectProjectSourcesStopsSerialGroupsOnError(t *testing.T) {
	firstError := errors.New("unreadable first config")
	reachedSecond := false
	_, err := SelectProjectSources(ProjectSelectionRequest{
		Targets:          []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: "/a/target.ts"}, ConfigDirectory: "/a"}, {PathIdentity: rslintconfig.PathIdentity{Path: "/b/target.ts"}, ConfigDirectory: "/b"}},
		CandidateIndexes: [][]int{{0}, {1}},
		Candidates:       []ProjectCandidate{{ConfigPath: "/a/tsconfig.json"}, {ConfigPath: "/b/tsconfig.json"}},
		SingleThreaded:   true,
		Metadata: func(index int) (*tsoptions.ParsedCommandLine, error) {
			if index == 0 {
				return nil, firstError
			}
			reachedSecond = true
			return nil, errors.New("unexpected later config read")
		},
	})
	if err == nil || !strings.Contains(err.Error(), firstError.Error()) || reachedSecond {
		t.Fatalf("serial selection continued after its first failure: err=%v second=%t", err, reachedSecond)
	}
}

func TestBuildProjectsPreservesRootFallback(t *testing.T) {
	for _, test := range []struct {
		name, target, wantProject string
		files                     map[string]string
		projects                  []string
	}{
		{
			name: "unsupported ordinary root remains a gap", target: "target.js",
			projects: []string{"./first.json"},
			files: map[string]string{
				"target.js":  `export const value = 1;`,
				"first.json": `{"files":["target.js"],"compilerOptions":{"noLib":true}}`,
			},
		},
		{
			name: "missing first root uses ordered source fallback", target: "target.js", wantProject: "import.json",
			projects: []string{"./first.json", "./import.json", "./direct.json"},
			files: map[string]string{
				"target.js":   `export const value = 1;`,
				"main.ts":     `import "./target.js";`,
				"first.json":  `{"files":["target.js"],"compilerOptions":{"noLib":true}}`,
				"import.json": `{"files":["main.ts"],"compilerOptions":{"noLib":true,"allowJs":true}}`,
				"direct.json": `{"files":["target.js"],"compilerOptions":{"noLib":true,"allowJs":true}}`,
			},
		},
		{
			name: "distinct physical root keeps later direct priority", target: "target.ts", wantProject: "direct.json",
			projects: []string{"./first.json", "./import.json", "./direct.json"},
			files: map[string]string{
				"Target.ts":   `export const upper = 1;`,
				"target.ts":   `export const lower = 2;`,
				"main.ts":     `import "./target";`,
				"first.json":  `{"files":["Target.ts"],"compilerOptions":{"noLib":true}}`,
				"import.json": `{"files":["main.ts"],"compilerOptions":{"noLib":true}}`,
				"direct.json": `{"files":["target.ts"],"compilerOptions":{"noLib":true}}`,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := tspath.NormalizePath(t.TempDir())
			files := make(map[string]string, len(test.files))
			for name, content := range test.files {
				files[tspath.ResolvePath(dir, name)] = content
			}
			fsys := &exactCaseProgramFS{FS: bundled.WrapFS(osvfs.FS()), files: files}
			file := testLintTarget(fsys, dir, tspath.ResolvePath(dir, test.target))
			plan := target.Plan{Files: []target.File{file}}
			for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
				for _, singleThreaded := range []bool{false, true} {
					session := NewSession(fsys)
					projects, err := session.buildProjectsForTest(t, ProjectBuildRequest{
						Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig(test.projects...)},
						Targets: plan, Scope: scope, SingleThreaded: singleThreaded,
					})
					if err != nil {
						t.Fatalf("scope=%d serial=%t: %v", scope, singleThreaded, err)
					}
					binding, err := session.LoadAPI(projects, plan, dir, singleThreaded)
					if err != nil {
						t.Fatal(err)
					}
					bound := 0
					for index, sources := range binding.TargetsByProgram {
						for _, sourceName := range sources {
							bound++
							program := binding.Programs[index]
							wantProject := ""
							if test.wantProject != "" {
								wantProject = tspath.ResolvePath(dir, test.wantProject)
							}
							if sourceName != file.Path || program.Options().ConfigFilePath != wantProject ||
								program.CanProvideTypeChecker(program.GetSourceFile(sourceName)) != (test.wantProject != "") {
								t.Fatalf("scope=%d: wrong binding: source=%q project=%q, want %q", scope, sourceName, program.Options().ConfigFilePath, wantProject)
							}
						}
					}
					if bound != 1 {
						t.Fatalf("target was bound %d times", bound)
					}
				}
			}
		})
	}
}

func TestBuildProjectsFiltersUnsupportedImportedTargets(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_extension_admission.txtar")
	for _, aliasTarget := range []bool{false, true} {
		for _, test := range []struct {
			name                          string
			mixed, later, firstJS, listed bool
		}{
			{name: "single gap"},
			{name: "listed gap", listed: true},
			{name: "mixed gap", mixed: true},
			{name: "single later project", later: true},
			{name: "mixed later project", mixed: true, later: true},
			{name: "listed later project", listed: true, later: true},
			{name: "supported physical alias", firstJS: true},
		} {
			for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
				for _, singleThreaded := range []bool{false, true} {
					t.Run(fmt.Sprintf("alias=%t/%s/scope=%d/serial=%t", aliasTarget, test.name, scope, singleThreaded), func(t *testing.T) {
						fixture := "js_source"
						source, alias := "real.js", "alias.ts"
						lintTarget, imported := source, alias
						if aliasTarget {
							fixture = "js_alias"
							source, alias = "real.ts", "alias.js"
							lintTarget, imported = alias, source
						}
						dir := tspath.NormalizePath(archive.Materialize(t, fixture))
						roots := `"main.ts","keep.ts"`
						if test.listed {
							roots += fmt.Sprintf(",%q", lintTarget)
						}
						writeProgramTestFiles(t, dir, map[string]string{
							"first.json": fmt.Sprintf(`{"files":[%s],"compilerOptions":{"noLib":true,"allowJs":%t}}`, roots, test.firstJS),
						})
						if err := os.Symlink(source, tspath.ResolvePath(dir, alias)); err != nil {
							t.Skipf("symlinks unavailable: %v", err)
						}
						fsys := &programReadCountingFS{FS: bundled.WrapFS(cachedvfs.From(osvfs.FS())), reads: make(map[string]int)}
						file := testLintTarget(fsys, dir, tspath.ResolvePath(dir, lintTarget))
						plan := target.Plan{Files: []target.File{file}}
						if test.mixed {
							plan.Files = append(plan.Files, testLintTarget(fsys, dir, tspath.ResolvePath(dir, "keep.ts")))
						}
						projects := []string{"./first.json"}
						if test.later {
							projects = append(projects, "./second.json")
						}
						session := NewSession(fsys)
						set, err := session.buildProjectsForTest(t, ProjectBuildRequest{
							Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig(projects...)},
							Targets: plan, Scope: scope, SingleThreaded: singleThreaded,
						})
						if err != nil {
							t.Fatal(err)
						}
						if scope == AllDeclared && set.Len() != len(projects) {
							t.Fatalf("type-check scope lost Programs: %d, want %d", set.Len(), len(projects))
						}
						if scope == LintTargets && !test.mixed && !test.firstJS {
							for _, name := range []string{"main.ts", "keep.ts", imported} {
								if count := fsys.readCount(tspath.ResolvePath(dir, name)); count != 0 {
									t.Fatalf("unsupported project's source %q read %d times", name, count)
								}
							}
						}
						binding, err := session.LoadAPI(set, plan, dir, singleThreaded)
						if err != nil {
							t.Fatal(err)
						}
						seen := make(map[string]int)
						for index, sources := range binding.TargetsByProgram {
							program := binding.Programs[index]
							for _, name := range sources {
								selected, ok := binding.LintTargetBySourcePath[exactPathID(name)]
								if !ok {
									t.Fatalf("bound source %q lost its lint target", name)
								}
								seen[selected.Path]++
								wantConfig := ""
								if selected.Path != file.Path || test.firstJS {
									wantConfig = tspath.ResolvePath(dir, "first.json")
								} else if test.later {
									wantConfig = tspath.ResolvePath(dir, "second.json")
								}
								if program.Options().ConfigFilePath != wantConfig ||
									program.CanProvideTypeChecker(program.GetSourceFile(name)) != (wantConfig != "") {
									t.Fatalf("target %q borrowed wrong project: %q, want %q", selected.Path, program.Options().ConfigFilePath, wantConfig)
								}
							}
						}
						if len(seen) != len(plan.Files) {
							t.Fatalf("lint target set changed: %v", seen)
						}
						for _, selected := range plan.Files {
							if seen[selected.Path] != 1 {
								t.Fatalf("target %q bound %d times", selected.Path, seen[selected.Path])
							}
						}
					})
				}
			}
		}
	}
}

func TestBuildProjectsMatchesCompilerExtensionAdmission(t *testing.T) {
	for _, test := range []struct {
		name, file, options      string
		caseSensitive, supported bool
	}{
		{name: "TS case insensitive", file: "target.TS", options: `{}`, supported: true},
		{name: "TS case sensitive", file: "target.TS", options: `{}`, caseSensitive: true},
		{name: "JS case insensitive", file: "target.JS", options: `{"allowJs":true}`, supported: true},
		{name: "JS case sensitive", file: "target.JS", options: `{"allowJs":true}`, caseSensitive: true},
		{name: "JS disabled", file: "target.js", options: `{"allowJs":false}`},
		{name: "checkJs implies allowJs", file: "target.js", options: `{"checkJs":true}`, supported: true},
		{name: "explicit false overrides checkJs", file: "target.js", options: `{"allowJs":false,"checkJs":true}`},
		{name: "module JS enabled", file: "target.mjs", options: `{"allowJs":true}`, caseSensitive: true, supported: true},
		{name: "common JS enabled", file: "target.cjs", options: `{"allowJs":true}`, caseSensitive: true, supported: true},
		{name: "JSX enabled", file: "target.jsx", options: `{"allowJs":true}`, caseSensitive: true, supported: true},
	} {
		for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
			t.Run(fmt.Sprintf("%s/scope=%d", test.name, scope), func(t *testing.T) {
				const dir = "/repo"
				file, configPath := tspath.ResolvePath(dir, test.file), tspath.ResolvePath(dir, "tsconfig.json")
				fsys := newBindingIndexTestFS([]string{file, configPath}, nil)
				fsys.caseSensitive = test.caseSensitive
				fsys.files[configPath] = fmt.Sprintf(`{"files":[%q],"compilerOptions":%s}`, test.file, test.options)
				plan := target.Plan{Files: []target.File{testLintTarget(fsys, dir, file)}}
				config := projectConfig("./tsconfig.json")
				config[0].Files = []string{test.file}
				session := NewSession(fsys)
				set, err := session.buildProjectsForTest(t, ProjectBuildRequest{
					Configs: map[string]rslintconfig.RslintConfig{dir: config},
					Targets: plan, Scope: scope, SingleThreaded: true,
				})
				if err != nil {
					t.Fatal(err)
				}
				if scope == LintTargets && !test.supported && set.Len() != 0 {
					t.Fatal("unsupported root constructed a Program")
				}
				binding, unbound := session.bindTargetsToProjects(set, plan, true)
				if (len(unbound) == 0) != test.supported {
					t.Fatalf("compiler extension admission differs: unbound=%v supported=%t", unbound, test.supported)
				}
				if test.supported && (len(binding.TargetsByProgram) != 1 || !slices.Equal(binding.TargetsByProgram[0], []string{file})) {
					t.Fatalf("supported root identity changed: %v", binding.TargetsByProgram)
				}
			})
		}
	}
}

func TestBuildProjectsKeepsServiceExtensionAdmissionLocal(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_extension_admission.txtar").Materialize(t, "service"))
	config := rslintconfig.RslintConfig{
		{Files: []string{"ordinary.js"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"./tsconfig.json"}}}},
		{Files: []string{"service.js"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}},
	}
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		for _, serial := range []bool{false, true} {
			t.Run(fmt.Sprintf("scope=%d/serial=%t", scope, serial), func(t *testing.T) {
				fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
				plan := target.Plan{Files: []target.File{
					testLintTarget(fsys, dir, tspath.ResolvePath(dir, "service.js")),
					testLintTarget(fsys, dir, tspath.ResolvePath(dir, "ordinary.js")),
				}}
				session := NewSession(fsys)
				set, err := session.buildProjectsForTest(t, ProjectBuildRequest{
					Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: scope, SingleThreaded: serial,
				})
				if err != nil {
					t.Fatal(err)
				}
				binding, err := session.LoadAPI(set, plan, dir, serial)
				if err != nil {
					t.Fatal(err)
				}
				seen := make(map[string]int)
				for index, sources := range binding.TargetsByProgram {
					program := binding.Programs[index]
					for _, file := range sources {
						seen[file]++
						wantTypes := file == plan.Files[0].Path
						if program.CanProvideTypeChecker(program.GetSourceFile(file)) != wantTypes {
							t.Fatalf("service extension permission leaked: %s", file)
						}
						if wantTypes && program.GetSourceFile(plan.Files[1].Path) == nil {
							t.Fatal("fixture did not exercise a service Program containing the ordinary JS target")
						}
					}
				}
				if len(seen) != len(plan.Files) || seen[plan.Files[0].Path] != 1 || seen[plan.Files[1].Path] != 1 {
					t.Fatalf("lint target set changed: %v", seen)
				}
			})
		}
	}
}

func TestBuildTargetProjectFallsBackToFirstImportOnlyAfterRootScan(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"target.ts":           `export const target = 1;`,
		"first-main.ts":       `import "./target";`,
		"later-main.ts":       `import "./target";`,
		"tsconfig-first.json": `{"files":["first-main.ts"]}`,
		"tsconfig-later.json": `{"files":["later-main.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	readCounter := &programReadCountingFS{
		FS:    bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		reads: make(map[string]int),
	}
	fsys := vfs.FS(readCounter)
	context := newBuildContext(fsys)
	config := projectConfig("./tsconfig-first.json", "./tsconfig-later.json")
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "target.ts")),
	}}

	set, err := sessionForTest(context).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: LintTargets, SingleThreaded: true})
	if err != nil {
		t.Fatalf("BuildTargetProject: %v", err)
	}
	if set.Len() != 1 {
		t.Fatalf("fallback built %d retained Programs, want first containing project only", set.Len())
	}
	wantConfig := tspath.ResolvePath(dir, "tsconfig-first.json")
	if got := tspath.NormalizePath(set.compilerPrograms[0].Options().ConfigFilePath); got != wantConfig {
		t.Fatalf("fallback project = %q, want first containing project %q", got, wantConfig)
	}
	if got := readCounter.readCount(tspath.ResolvePath(dir, "later-main.ts")); got != 0 {
		t.Fatalf("fallback built the later project %d time(s) after the first match", got)
	}
	binding, err := sessionForTest(context).LoadAPI(set, plan, dir, true)
	if err != nil {
		t.Fatalf("LoadAPI: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
		t.Fatalf("imported target lost its configured Program: %v", binding.TargetsByProgram)
	}
}

func TestBuildTargetProjectRetainsOnlyActualImportOwners(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":                `import "./target.js";`,
		"target.js":              `export const target = 1;`,
		"tsconfig-no-js.json":    `{"files":["main.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig-allow-js.json": `{"files":["main.ts"],"compilerOptions":{"allowJs":true,"noLib":true}}`,
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "target.js")),
	}}

	for _, test := range []struct {
		name         string
		project      string
		wantPrograms int
	}{
		{name: "allowJs disabled", project: "./tsconfig-no-js.json", wantPrograms: 0},
		{name: "allowJs enabled", project: "./tsconfig-allow-js.json", wantPrograms: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			context := newBuildContext(fsys)
			set, err := sessionForTest(context).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig(test.project)}, Targets: plan, Scope: LintTargets, SingleThreaded: false})
			if err != nil {
				t.Fatalf("BuildTargetProject: %v", err)
			}
			if set.Len() != test.wantPrograms {
				t.Fatalf("selected Programs = %d, want %d", set.Len(), test.wantPrograms)
			}
		})
	}
}

func TestBuildTargetProjectKeepsDirectAndImportFallbackTiersPerTarget(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"direct.ts":       `export const direct = 1;`,
		"fallback.ts":     `export const fallback = 1;`,
		"import-main.ts":  `import "./direct"; import "./fallback";`,
		"tsconfig-a.json": `{"files":["import-main.ts"]}`,
		"tsconfig-b.json": `{"files":["direct.ts"]}`,
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	context := newBuildContext(fsys)
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "direct.ts")),
		testLintTarget(fsys, dir, filepath.Join(dir, "fallback.ts")),
	}}
	set, err := sessionForTest(context).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-a.json", "./tsconfig-b.json")}, Targets: plan, Scope: LintTargets, SingleThreaded: false})
	if err != nil {
		t.Fatalf("BuildTargetProject: %v", err)
	}
	if set.Len() != 2 {
		t.Fatalf("selected Programs = %d, want direct and fallback projects", set.Len())
	}
	binding, err := sessionForTest(context).LoadAPI(set, plan, dir, false)
	if err != nil {
		t.Fatalf("LoadAPI: %v", err)
	}
	if len(binding.TargetsByProgram) != 2 ||
		len(binding.TargetsByProgram[0]) != 1 ||
		len(binding.TargetsByProgram[1]) != 1 ||
		!strings.HasSuffix(binding.TargetsByProgram[0][0], "/fallback.ts") ||
		!strings.HasSuffix(binding.TargetsByProgram[1][0], "/direct.ts") {
		t.Fatalf("target tiers were reordered by construction timing: %v", binding.TargetsByProgram)
	}
}

func TestBuildTargetProjectBuildsMultipleDirectWinnersInParallel(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts":            `export const a = 1;`,
		"b.ts":            `export const b = 1;`,
		"tsconfig-a.json": `{"files":["a.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig-b.json": `{"files":["b.ts"],"compilerOptions":{"noLib":true}}`,
	})
	dir = tspath.NormalizePath(dir)
	aPath := tspath.ResolvePath(dir, "a.ts")
	bPath := tspath.ResolvePath(dir, "b.ts")
	fsys := &blockingProgramConfigFS{
		FS:         bundled.WrapFS(osvfs.FS()),
		paths:      map[string]struct{}{aPath: {}, bPath: {}},
		waitFor:    2,
		allStarted: make(chan struct{}),
		release:    make(chan struct{}),
	}
	context := newBuildContext(fsys)
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, aPath),
		testLintTarget(fsys, dir, bPath),
	}}
	type result struct {
		set ProjectSet
		err error
	}
	done := make(chan result, 1)
	go func() {
		set, err := sessionForTest(context).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-a.json", "./tsconfig-b.json")}, Targets: plan, Scope: LintTargets, SingleThreaded: false})
		done <- result{set: set, err: err}
	}()

	select {
	case <-fsys.allStarted:
		close(fsys.release)
	case <-time.After(5 * time.Second):
		close(fsys.release)
		t.Fatal("direct winner Programs were not built concurrently")
	}
	got := <-done
	if got.err != nil {
		t.Fatalf("BuildTargetProject: %v", got.err)
	}
	if got.set.Len() != 2 {
		t.Fatalf("direct winner Programs = %d, want two", got.set.Len())
	}
	if peak := fsys.peakConcurrency(); peak < 2 {
		t.Fatalf("direct winner build concurrency = %d, want at least two", peak)
	}
}

func TestBuildTargetProjectsDeduplicatesSharedDirectWinnerAcrossOwners(t *testing.T) {
	rootDir := tspath.NormalizePath(t.TempDir())
	childDir := tspath.ResolvePath(rootDir, "child")
	writeProgramTestFiles(t, rootDir, map[string]string{
		"root.ts":        `export const root = 1;`,
		"child/child.ts": `export const child = 1;`,
		"tsconfig.json":  `{"files":["root.ts","child/child.ts"]}`,
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	context := newBuildContext(fsys)
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, rootDir, filepath.Join(rootDir, "root.ts")),
		testLintTarget(fsys, childDir, filepath.Join(childDir, "child.ts")),
	}}
	set, err := sessionForTest(context).buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{
		rootDir:  projectConfig("./tsconfig.json"),
		childDir: projectConfig("../tsconfig.json"),
	}, Targets: plan, Scope: LintTargets, SingleThreaded: false})
	if err != nil {
		t.Fatalf("BuildTargetProjects: %v", err)
	}
	if set.Len() != 1 || !slices.Equal(set.targetProjects[plan.Files[0]], []int{0}) || !slices.Equal(set.targetProjects[plan.Files[1]], []int{0}) {
		t.Fatalf("shared direct winner was not deduplicated: Programs=%d candidates=%v", set.Len(), set.targetProjects)
	}
	binding, err := sessionForTest(context).LoadAPI(set, plan, rootDir, false)
	if err != nil {
		t.Fatalf("LoadAPI: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 2 {
		t.Fatalf("shared direct winner lost an owner's target: %v", binding.TargetsByProgram)
	}
}

func TestBuildTargetProjectsKeepsCandidateOrderAcrossGroups(t *testing.T) {
	for _, firstConfig := range []string{"first.json", "alias.json"} {
		for _, singleThreaded := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/serial=%t", firstConfig, singleThreaded), func(t *testing.T) {
				dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_order_groups.txtar").Materialize(t, ""))
				if firstConfig == "alias.json" {
					if err := os.Symlink("a.ts", tspath.ResolvePath(dir, "alias.ts")); err != nil {
						t.Skipf("symlinks unavailable: %v", err)
					}
				}
				config := projectConfig("./"+firstConfig, "./second.json")
				other := projectConfig("./second.json", "./"+firstConfig)[0]
				other.Files = []string{"b.ts"}
				config = append(config, other)
				fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
				session := NewSession(fsys)
				plan := target.Plan{Files: []target.File{
					testLintTarget(fsys, dir, tspath.ResolvePath(dir, "a.ts")),
					testLintTarget(fsys, dir, tspath.ResolvePath(dir, "b.ts")),
				}}
				projects, err := session.buildProjectsForTest(t, ProjectBuildRequest{
					Configs: map[string]rslintconfig.RslintConfig{dir: config},
					Targets: plan, Scope: LintTargets, SingleThreaded: singleThreaded,
				})
				if err != nil {
					t.Fatal(err)
				}
				binding, err := session.LoadAPI(projects, plan, dir, singleThreaded)
				if err != nil {
					t.Fatal(err)
				}
				bound := make(map[string]string)
				for index, sources := range binding.TargetsByProgram {
					for _, source := range sources {
						file := binding.LintTargetBySourcePath[exactPathID(source)]
						bound[file.Path] = binding.Programs[index].Options().ConfigFilePath
					}
				}
				if len(bound) != 2 || bound[plan.Files[0].Path] != tspath.ResolvePath(dir, firstConfig) ||
					bound[plan.Files[1].Path] != tspath.ResolvePath(dir, "second.json") {
					t.Fatalf("shared candidates changed per-target order: %v", bound)
				}
			})
		}
	}
}

func TestBuildTargetProjectUsesFrozenTargetIdentityForMembership(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "direct root validation",
			files: map[string]string{
				"target.ts":     `export const target = 1;`,
				"tsconfig.json": `{"files":["target.ts"],"compilerOptions":{"noLib":true}}`,
			},
		},
		{
			name: "import fallback membership",
			files: map[string]string{
				"main.ts":       `import "./target";`,
				"target.ts":     `export const target = 1;`,
				"tsconfig.json": `{"files":["main.ts"],"compilerOptions":{"noLib":true}}`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projectDir := tspath.NormalizePath(t.TempDir())
			writeProgramTestFiles(t, projectDir, test.files)
			baseFS := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			frozenCanonicalPath := tspath.NormalizePath(
				baseFS.Realpath(tspath.ResolvePath(projectDir, "target.ts")),
			)
			frozenCanonicalParent := tspath.NormalizePath(baseFS.Realpath(projectDir))

			requestDir := tspath.NormalizePath(t.TempDir())
			requestedPath := tspath.ResolvePath(requestDir, "target.ts")
			liveDir := tspath.NormalizePath(t.TempDir())
			writeProgramTestFiles(t, liveDir, map[string]string{
				"target.ts": `export const replacement = 2;`,
			})
			fsys := &retargetingFrozenTargetFS{
				FS:            baseFS,
				targetPath:    requestedPath,
				canonicalPath: frozenCanonicalPath,
				liveCanonicalPath: tspath.NormalizePath(
					baseFS.Realpath(tspath.ResolvePath(liveDir, "target.ts")),
				),
			}
			plan := target.Plan{Files: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: requestedPath,
				CanonicalPath:       frozenCanonicalPath,
				CanonicalParentPath: frozenCanonicalParent}, ConfigDirectory: projectDir,
			}}}

			session := sessionForTest(newBuildContext(fsys))
			set, err := session.buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{projectDir: projectConfig("./tsconfig.json")}, Targets: plan, Scope: LintTargets, SingleThreaded: true})
			if err != nil {
				t.Fatalf("BuildTargetProject: %v", err)
			}
			if set.Len() != 1 {
				t.Fatalf("selected projects = %d, want frozen project", set.Len())
			}
			binding, err := session.LoadAPI(set, plan, projectDir, true)
			if err != nil {
				t.Fatalf("LoadAPI: %v", err)
			}
			if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 1 {
				t.Fatalf("frozen target binding = %v, want one project target", binding.TargetsByProgram)
			}
			if fsys.targetCalls != 0 || fsys.canonicalCalls != 0 {
				t.Fatalf("target Realpath calls = %d lexical, %d canonical; want frozen identities only", fsys.targetCalls, fsys.canonicalCalls)
			}
		})
	}
}

func TestLoadProgramsPreservesVirtualFileIdentityUnderDirectorySymlink(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_order_groups.txtar")
	for _, withDirectProject := range []bool{false, true} {
		for _, scope := range []ProjectScope{LintTargets, AllDeclared} {
			t.Run(fmt.Sprintf("direct=%t/scope=%d", withDirectProject, scope), func(t *testing.T) {
				baseFS := osvfs.FS()
				dir := tspath.NormalizePath(baseFS.Realpath(archive.Materialize(t, "virtual-alias")))
				realDir := tspath.ResolvePath(dir, "real")
				aliasDir := tspath.ResolvePath(dir, "alias")
				if err := os.MkdirAll(realDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(realDir, aliasDir); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
				targetPath := tspath.ResolvePath(realDir, "a.ts")
				aliasPath := tspath.ResolvePath(aliasDir, "a.ts")
				const targetText = "export const targetBuffer = 111;"
				fsys := utils.NewOverlayVFS(baseFS, map[string]string{
					targetPath:                           targetText,
					aliasPath:                            "export const unrelatedBuffer = 222;",
					tspath.ResolvePath(aliasDir, "b.ts"): "export const companion = 333;",
				})
				// These files have not been saved. Their individual VFS identities
				// stay distinct even though their existing parent directories alias.
				if exactPathID(fsys.Realpath(aliasPath)) == exactPathID(fsys.Realpath(targetPath)) {
					t.Fatal("fixture must provide distinct virtual file identities")
				}
				projects := []string{"./first.json"}
				if withDirectProject {
					projects = append(projects, "./second.json")
				}
				plan := target.Plan{Files: []target.File{testLintTarget(fsys, dir, targetPath)}}
				session := NewSession(fsys)
				set, err := session.buildProjectsForTest(t, ProjectBuildRequest{
					Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig(projects...)},
					Targets: plan, Scope: scope, SingleThreaded: true,
				})
				if err != nil {
					t.Fatal(err)
				}
				wantProjects := 0
				if withDirectProject {
					wantProjects = 1
				}
				if scope == AllDeclared {
					wantProjects = len(projects)
				}
				if set.Len() != wantProjects {
					t.Fatalf("retained projects = %d, want %d", set.Len(), wantProjects)
				}
				for _, load := range []func(ProjectSet, target.Plan, string, bool) (LoadResult, error){session.LoadCLI, session.LoadAPI} {
					binding, err := load(set, plan, dir, true)
					if err != nil {
						t.Fatal(err)
					}
					seen := 0
					for index, paths := range binding.TargetsByProgram {
						for _, path := range paths {
							seen++
							program := binding.Programs[index]
							source := program.GetSourceFile(path)
							if source == nil || source.Text() != targetText || exactPathID(path) != exactPathID(targetPath) ||
								program.CanProvideTypeChecker(source) != withDirectProject {
								t.Fatalf("virtual dependency replaced target source or types: paths=%v source=%v", binding.TargetsByProgram, source)
							}
						}
					}
					if seen != 1 {
						t.Fatalf("lint target count = %d, want 1", seen)
					}
				}
			})
		}
	}
}

func TestBuildTargetProjectDoesNotBorrowAnotherTargetsLiveIdentity(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"physical/a.ts": `export const a = 1;`,
		"physical/b.ts": `export const b = 2;`,
		"main.ts":       `import "./physical/a";`,
		"first.json":    `{"files":["sources/b.ts"],"compilerOptions":{"noLib":true}}`,
		"second.json":   `{"files":["main.ts"],"compilerOptions":{"noLib":true}}`,
	})
	sourceB := tspath.ResolvePath(dir, "sources/b.ts")
	if err := os.MkdirAll(filepath.Dir(sourceB), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(tspath.ResolvePath(dir, "physical/b.ts"), sourceB); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	physicalA := tspath.ResolvePath(dir, "physical/a.ts")
	physicalB := tspath.ResolvePath(dir, "physical/b.ts")
	fsys := &retargetingFrozenTargetFS{
		FS:         bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		targetPath: sourceB, liveCanonicalPath: physicalA,
	}
	plan := target.Plan{Files: []target.File{
		{PathIdentity: rslintconfig.PathIdentity{
			Path: tspath.ResolvePath(dir, "caller/a.ts"), CanonicalPath: physicalA,
			CanonicalParentPath: tspath.GetDirectoryPath(physicalA),
		}, ConfigDirectory: dir},
		{PathIdentity: rslintconfig.PathIdentity{
			Path: sourceB, CanonicalPath: physicalB,
			CanonicalParentPath: tspath.GetDirectoryPath(physicalB),
		}, ConfigDirectory: dir},
	}}
	session := NewSession(fsys)
	set, err := session.buildProjectsForTest(t, ProjectBuildRequest{
		Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./first.json", "./second.json")},
		Targets: plan, Scope: LintTargets, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	binding, err := session.LoadAPI(set, plan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if set.Len() != 2 || len(binding.Programs) != 2 ||
		!slices.Equal(binding.TargetsByProgram[0], []string{sourceB}) ||
		!slices.Equal(binding.TargetsByProgram[1], []string{physicalA}) {
		t.Fatalf("frozen identities lost during project selection: targets=%v, projects=%d", binding.TargetsByProgram, set.Len())
	}
	if fsys.targetCalls != 0 {
		t.Fatalf("re-read another target's physical identity %d time(s)", fsys.targetCalls)
	}
}

func TestLoadProgramsRecomputesProgramMembershipAfterImportGraphChange(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       `import "./extra";`,
		"extra.ts":      `export const value = 1;`,
		"tsconfig.json": `{"files":["main.ts"]}`,
	})

	dir = tspath.NormalizePath(dir)
	config := projectConfig("./tsconfig.json")
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	buildContext := newBuildContext(fsys)
	set, err := buildProjectsForConfig(dir, config, true, buildContext)
	if err != nil {
		t.Fatalf("initial buildProjectsForConfig: %v", err)
	}
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "extra.ts")),
	}}
	initial, err := loadAPIForTest(set, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("initial loadAPIForTest: %v", err)
	}
	if len(initial.Programs) != 1 || len(initial.TargetsByProgram[0]) != 1 {
		t.Fatalf("imported target should initially use the configured Program, got targets=%v", initial.TargetsByProgram)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.ts"), []byte(`export const main = 1;`), 0644); err != nil {
		t.Fatalf("rewrite main.ts: %v", err)
	}
	// Production fix passes reuse the run-scoped filesystem and parse caches.
	// Replacing the source-snapshot generation before rebuilding makes the new
	// text/hash visible while retaining content-keyed AST entries for unchanged
	// files.
	buildContext.invalidateSourceSnapshots()
	rebuilt, err := buildProjectsForConfig(dir, config, true, buildContext)
	if err != nil {
		t.Fatalf("rebuilt buildProjectsForConfig: %v", err)
	}
	afterFix, err := loadAPIForTest(rebuilt, plan, dir, buildContext, true)
	if err != nil {
		t.Fatalf("rebuilt loadAPIForTest: %v", err)
	}
	if len(afterFix.Programs) != 2 || len(afterFix.TargetsByProgram[1]) != 1 {
		t.Fatalf("target must move to a source-only Program after its importing edge is removed, got targets=%v", afterFix.TargetsByProgram)
	}
}

func TestBuildTargetProjectRecomputesImportFallbackAfterFix(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"main.ts":       `import "./target";`,
		"target.ts":     `export const target = 1;`,
		"tsconfig.json": `{"files":["main.ts"]}`,
	})
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	context := newBuildContext(fsys)
	session := sessionForTest(context)
	config := projectConfig("./tsconfig.json")
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "target.ts")),
	}}

	initialSet, err := session.buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: LintTargets, SingleThreaded: true})
	if err != nil {
		t.Fatalf("initial BuildTargetProject: %v", err)
	}
	initial, err := session.LoadAPI(initialSet, plan, dir, true)
	if err != nil {
		t.Fatalf("initial LoadAPI: %v", err)
	}
	if initialSet.Len() != 1 || len(initial.TargetsByProgram[0]) != 1 {
		t.Fatalf("import fallback was not selected initially: %v", initial.TargetsByProgram)
	}

	if err := os.WriteFile(filepath.Join(dir, "main.ts"), []byte(`export const main = 1;`), 0o644); err != nil {
		t.Fatalf("rewrite main.ts: %v", err)
	}
	session.InvalidateSourceSnapshots()
	afterFixSet, err := session.buildProjectsForTest(t, ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: LintTargets, SingleThreaded: true})
	if err != nil {
		t.Fatalf("post-fix BuildTargetProject: %v", err)
	}
	afterFix, err := session.LoadAPI(afterFixSet, plan, dir, true)
	if err != nil {
		t.Fatalf("post-fix LoadAPI: %v", err)
	}
	if afterFixSet.Len() != 0 || len(afterFix.Programs) != 1 || afterFix.Programs[0].CanProvideTypeChecker(afterFix.Programs[0].SourceFiles()[0]) {
		t.Fatalf("removed import did not move target to source-only fallback: projects=%d targets=%v", afterFixSet.Len(), afterFix.TargetsByProgram)
	}
}

func TestTargetResolve_DirectoryWalkAvoidsPerTargetRealpath(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"src/a.ts": `export const a = 1;`,
		"src/b.ts": `export const b = 2;`,
	})
	dir = tspath.NormalizePath(dir)
	fileA := tspath.ResolvePath(dir, "src/a.ts")
	fileB := tspath.ResolvePath(dir, "src/b.ts")
	counter := &targetPlanRealpathCountingFS{FS: osvfs.FS(), calls: make(map[string]int)}
	fsys := bundled.WrapFS(cachedvfs.From(counter))

	plan, err := resolveTargetPlanForTest(
		nil,
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		dir,
		nil,
		fsys,
		nil,
		[]string{dir},
		true,
	)
	if err != nil {
		t.Fatalf("resolveTargetPlanForTest: %v", err)
	}
	if len(plan.Files) != 2 {
		t.Fatalf("targets = %v, want two files", plan.Files)
	}
	if counter.callCount(fileA) != 0 || counter.callCount(fileB) != 0 {
		t.Fatalf("regular targets performed realpath IO: a=%d b=%d", counter.callCount(fileA), counter.callCount(fileB))
	}
}

func TestTargetResolve_RejectsCanonicalTargetWithDifferentOwners(t *testing.T) {
	sharedDir := t.TempDir()
	writeProgramTestFiles(t, sharedDir, map[string]string{
		"target.ts": `export const value = 1;`,
	})
	ownersRoot := t.TempDir()
	ownerA := filepath.Join(ownersRoot, "owner-a")
	ownerB := filepath.Join(ownersRoot, "owner-b")
	if err := os.MkdirAll(ownerA, 0755); err != nil {
		t.Fatalf("mkdir owner A: %v", err)
	}
	if err := os.MkdirAll(ownerB, 0755); err != nil {
		t.Fatalf("mkdir owner B: %v", err)
	}
	sharedTarget := filepath.Join(sharedDir, "target.ts")
	targetA := filepath.Join(ownerA, "target.ts")
	targetB := filepath.Join(ownerB, "target.ts")
	if err := os.Symlink(sharedTarget, targetA); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.Symlink(sharedTarget, targetB); err != nil {
		t.Skipf("second symlink unavailable: %v", err)
	}

	ownerA = tspath.NormalizePath(ownerA)
	ownerB = tspath.NormalizePath(ownerB)
	targetA = tspath.NormalizePath(targetA)
	targetB = tspath.NormalizePath(targetB)
	configMap := map[string]rslintconfig.RslintConfig{
		ownerA: {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		ownerB: {{Rules: rslintconfig.Rules{"no-console": "error"}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	_, err := resolveTargetPlanForTest(
		configMap,
		nil,
		tspath.NormalizePath(ownersRoot),
		nil,
		fsys,
		[]string{targetA, targetB},
		nil,
		true,
	)
	if err == nil {
		t.Fatal("expected aliases governed by different configs to be rejected")
		return
	}
	if !strings.Contains(err.Error(), "resolve to the same file") || !strings.Contains(err.Error(), ownerA) || !strings.Contains(err.Error(), ownerB) {
		t.Fatalf("unexpected ownership conflict error: %v", err)
	}
}

func TestPlainProgramSetSkipsInactiveConfigProjects(t *testing.T) {
	root := t.TempDir()
	activeDir := filepath.Join(root, "active")
	inactiveDir := filepath.Join(root, "inactive")
	writeProgramTestFiles(t, root, map[string]string{
		"active/index.ts":      "export const value = 1;\n",
		"active/tsconfig.json": `{"files":["index.ts"]}`,
	})
	if err := os.MkdirAll(inactiveDir, 0o755); err != nil {
		t.Fatalf("mkdir inactive config: %v", err)
	}
	activeDir = tspath.NormalizePath(activeDir)
	inactiveDir = tspath.NormalizePath(inactiveDir)
	configMap := map[string]rslintconfig.RslintConfig{
		activeDir:   projectConfig("./tsconfig.json"),
		inactiveDir: projectConfig("./missing.json"),
	}
	plan := target.Plan{Files: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: tspath.ResolvePath(activeDir, "index.ts"),
		CanonicalPath: tspath.ResolvePath(activeDir, "index.ts")}, ConfigDirectory: activeDir,
	}}}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	emptySet, err := sessionForTest(newBuildContext(fsys)).buildProjectsForTest(t, ProjectBuildRequest{Configs: configMap, Targets: target.Plan{}, Scope: LintTargets, SingleThreaded: true})
	if err != nil || len(emptySet.compilerPrograms) != 0 {
		t.Fatalf("an empty target plan must not build config projects: programs=%d err=%v", len(emptySet.compilerPrograms), err)
	}

	set, err := sessionForTest(newBuildContext(fsys)).buildProjectsForTest(t, ProjectBuildRequest{Configs: configMap, Targets: plan, Scope: LintTargets, SingleThreaded: true})
	if err != nil || len(set.compilerPrograms) != 1 {
		t.Fatalf("plain lint should build only the active config Program: programs=%d err=%v", len(set.compilerPrograms), err)
	}
	if _, err := sessionForTest(newBuildContext(fsys)).buildProjectsForTest(t, ProjectBuildRequest{Configs: configMap, Scope: AllDeclared, SingleThreaded: true}); err == nil || !strings.Contains(err.Error(), "missing.json") {
		t.Fatalf("the all-project type-check scope must still reject the inactive missing project, got %v", err)
	}
}

type canonicalIdentityTestFS struct {
	vfs.FS
	realPaths map[string]string
}

type exactCaseProgramFS struct {
	vfs.FS
	files map[string]string
}

func (fs *exactCaseProgramFS) UseCaseSensitiveFileNames() bool { return false }
func (fs *exactCaseProgramFS) FileExists(filePath string) bool {
	if _, ok := fs.files[tspath.NormalizePath(filePath)]; ok {
		return true
	}
	return fs.FS.FileExists(filePath)
}
func (fs *exactCaseProgramFS) ReadFile(filePath string) (string, bool) {
	if content, ok := fs.files[tspath.NormalizePath(filePath)]; ok {
		return content, true
	}
	return fs.FS.ReadFile(filePath)
}
func (fs *exactCaseProgramFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if _, ok := fs.files[filePath]; ok {
		return filePath
	}
	return fs.FS.Realpath(filePath)
}

func (fs *canonicalIdentityTestFS) UseCaseSensitiveFileNames() bool { return false }
func (fs *canonicalIdentityTestFS) FileExists(string) bool          { return true }
func (fs *canonicalIdentityTestFS) Realpath(filePath string) string {
	if realPath := fs.realPaths[tspath.NormalizePath(filePath)]; realPath != "" {
		return realPath
	}
	return tspath.NormalizePath(filePath)
}

func TestTargetResolve_UsesCanonicalIdentityInsteadOfGlobalCaseFlag(t *testing.T) {
	configDir := "C:/Repo"
	upper := "C:/Repo/Src/A.ts"
	lower := "c:/repo/src/a.ts"
	config := rslintconfig.RslintConfig{{}}

	t.Run("same canonical path is deduplicated", func(t *testing.T) {
		fsys := &canonicalIdentityTestFS{
			FS: osvfs.FS(),
			realPaths: map[string]string{
				upper: "C:/Repo/Src/A.ts",
				lower: "C:/Repo/Src/A.ts",
			},
		}
		plan, err := resolveTargetPlanForTest(nil, config, configDir, nil, fsys, []string{upper, lower}, nil, true)
		if err != nil || len(plan.Files) != 1 {
			t.Fatalf("same canonical target should be deduplicated: targets=%v err=%v", plan.Files, err)
		}
	})

	t.Run("distinct canonical paths remain distinct", func(t *testing.T) {
		fsys := &canonicalIdentityTestFS{
			FS: osvfs.FS(),
			realPaths: map[string]string{
				upper: upper,
				lower: lower,
			},
		}
		plan, err := resolveTargetPlanForTest(nil, config, configDir, nil, fsys, []string{upper, lower}, nil, true)
		if err != nil || len(plan.Files) != 2 {
			t.Fatalf("global case behavior must not merge distinct physical paths: targets=%v err=%v", plan.Files, err)
		}
	})
}

func TestLoadProgramsRejectsCaseFoldedSourceWithDifferentCanonicalIdentity(t *testing.T) {
	configDir := "/repo"
	upper := "/repo/Source.ts"
	lower := "/repo/source.ts"
	fsys := &exactCaseProgramFS{
		FS: osvfs.FS(),
		files: map[string]string{
			upper: "export const upper = 1;\n",
			lower: "export const lower = 2;\n",
		},
	}
	host := utils.CreateCompilerHost(configDir, fsys)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{upper}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}
	if source := program.GetSourceFile(lower); source == nil || source.FileName() != upper {
		t.Fatalf("fixture must exercise case-folded Program lookup, got %v", source)
	}

	set := ProjectSet{
		compilerPrograms: []*compiler.Program{program},
		configOrders:     []configOrders{{configDir: 0}},
	}
	plan := target.Plan{Files: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: lower,
		CanonicalPath: lower}, ConfigDirectory: configDir,
	}}}
	binding, err := loadAPIForTest(set, plan, configDir, newBuildContext(fsys), true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram[0]) != 0 {
		t.Fatalf("lower-case target must not bind to the distinct upper-case source: %+v", binding.TargetsByProgram)
	}
	if got := binding.TargetsByProgram[1]; len(got) != 1 || got[0] != lower {
		t.Fatalf("lower-case target must bind to its exact compatibility source, got %v", got)
	}
}

func TestBuildProjectsRejectsCaseFoldedServiceSourceWithDifferentCanonicalIdentity(t *testing.T) {
	const configDir = "/repo"
	const upper = "/repo/Source.ts"
	const lower = "/repo/source.ts"
	fsys := &exactCaseProgramFS{
		FS: osvfs.FS(),
		files: map[string]string{
			configDir + "/tsconfig.json": `{"compilerOptions":{"noLib":true},"files":["Source.ts"]}`,
			upper:                        "export const upper = 1;\n",
			lower:                        "export const lower = 2;\n",
		},
	}
	config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{
		ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)},
	}}}
	plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: configDir, FS: fsys, Files: []string{lower}})
	if err != nil || len(plan.Files) != 1 || plan.Files[0].CanonicalPath != lower {
		t.Fatalf("fixture must retain the distinct physical target: files=%v error=%v", plan.Files, err)
	}
	for _, scope := range []ProjectScope{AllDeclared, LintTargets} {
		t.Run(strconv.Itoa(int(scope)), func(t *testing.T) {
			projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t,
				map[string]rslintconfig.RslintConfig{configDir: config}, plan, scope, true)
			if err == nil || !strings.Contains(err.Error(), "was absent") || !strings.Contains(err.Error(), lower) {
				t.Fatalf("case-folded source escaped selected-root validation: projects=%d error=%v", projects.Len(), err)
			}
		})
	}
}

func TestLoadProgramsSplitsCompatibilityProgramsForCaseFoldedPathCollisions(t *testing.T) {
	configDir := "/repo"
	upper := "/repo/Source.ts"
	lower := "/repo/source.ts"
	fsys := &exactCaseProgramFS{
		FS: osvfs.FS(),
		files: map[string]string{
			upper: "export const upper = 1;\n",
			lower: "export const lower = 2;\n",
		},
	}
	plan := target.Plan{Files: []target.File{
		{PathIdentity: rslintconfig.PathIdentity{Path: upper, CanonicalPath: upper}, ConfigDirectory: configDir},
		{PathIdentity: rslintconfig.PathIdentity{Path: lower, CanonicalPath: lower}, ConfigDirectory: configDir},
	}}
	binding, err := loadAPIForTest(ProjectSet{}, plan, configDir, newBuildContext(fsys), true)
	if err != nil {
		t.Fatalf("loadAPIForTest: %v", err)
	}
	if len(binding.compilerPrograms) != 2 || len(binding.TargetsByProgram) != 2 {
		t.Fatalf("case-folded root names require separate compatibility Programs, got %d", len(binding.compilerPrograms))
	}
	bound := []string{binding.TargetsByProgram[0][0], binding.TargetsByProgram[1][0]}
	slices.Sort(bound)
	want := []string{upper, lower}
	slices.Sort(want)
	if !slices.Equal(bound, want) {
		t.Fatalf("compatibility Programs must preserve both exact source identities: got %v want %v", bound, want)
	}
}
