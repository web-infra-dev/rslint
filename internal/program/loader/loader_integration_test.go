package loader

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
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
	resolver := configLint.NewResolver(configLint.ResolverOptions{
		ConfigsByOwner: configs, FS: s.FS(), PathSpaces: plan.PathSpaces(), Catalog: rule.NewCatalog(),
	})
	policies, err := resolver.ProjectPolicies(plan.Files)
	if err != nil {
		return ProjectSet{}, err
	}
	return s.BuildProjects(ProjectBuildRequest{
		Configs: configs, Targets: plan, Policies: policies, Scope: scope, SingleThreaded: singleThreaded,
	})
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
	projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, Targeted, true)
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
	for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
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
	for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
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
		for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
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
	for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
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
				projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t, configs, plan, Targeted, true)
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

func TestBuildProjectsPreservesOwnerDeclarationsAcrossTargets(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service.txtar").Materialize(t, ""))
	config := rslintconfig.RslintConfig{
		{Files: []string{"**/*.ts"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(false)}}},
		{Files: []string{"b/**"}, LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{Project: rslintconfig.ProjectPaths{"a/tsconfig.json", "b/custom.json"}}}},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{dir + "/a/src/file.ts", dir + "/b/src/file.ts"}})
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession(fsys)
	projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, Targeted, true)
	if err != nil {
		t.Fatal(err)
	}
	if projects.Len() != 2 {
		t.Fatalf("raw declarations did not retain both target owners: %d", projects.Len())
	}
	binding, err := session.LoadAPI(projects, plan, dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(binding.Programs) != 2 || len(binding.TargetsByProgram[0]) != 1 || len(binding.TargetsByProgram[1]) != 1 {
		t.Fatalf("raw declarations lost a target: %v", binding.TargetsByProgram)
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
	projects, err := session.buildProjectsWithOptionsForTest(t, configs, plan, Targeted, false)
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
	projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, Targeted, true)
	if err != nil || projects.Len() != 1 {
		t.Fatalf("an empty later array changed the raw declaration list: count=%d error=%v", projects.Len(), err)
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
		{"baseline", `{}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched service true", `{"projectService":true}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched service false", `{"projectService":false}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched service null", `{"projectService":null}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched root", `{"tsconfigRootDir":"."}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched root null", `{"tsconfigRootDir":null}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched project false", `{"project":false}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched project null", `{"project":null}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched missing project", `{"project":"missing.json"}`, false, false, "tsconfig.unsafe.json"},
		{"unmatched project and service", `{"projectService":false,"project":"missing.json"}`, false, false, "tsconfig.unsafe.json"},
		{"matched project array order", `{"project":["./tsconfig.unsafe.json","./tsconfig.safe.json"]}`, true, false, "tsconfig.unsafe.json"},
		{"matched reversed project array order", `{"project":["./tsconfig.safe.json","./tsconfig.unsafe.json"]}`, true, false, "tsconfig.safe.json"},
		{"matched false then restore", `{"project":false}`, true, true, "tsconfig.unsafe.json"},
		{"matched null then restore", `{"project":null}`, true, true, "tsconfig.unsafe.json"},
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
				wantConfig = "tsconfig.unsafe.json"
			}
			fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
			plan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{file}})
			if err != nil {
				t.Fatal(err)
			}
			session := NewSession(fsys)
			projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, Targeted, true)
			if strings.Contains(test.name, "unmatched missing project") || strings.Contains(test.name, "unmatched project and service") {
				if err == nil || !strings.Contains(err.Error(), "missing.json") {
					t.Fatalf("unmatched new fields changed the original declaration error: %v", err)
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
			projects, err := NewSession(fsys).buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, Targeted, true)
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
		projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, plan, Targeted, true)
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
	for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
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

func TestBuildProjectsKeepsDefaultDisabledTargetsUnbound(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/project_service.txtar")
	for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
		for _, options := range []string{`{"projectService":false}`, `{"projectService":null}`, `{"project":false}`, `{"project":null}`} {
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
				if err != nil || projects.Len() != 1 {
					t.Fatalf("ordinary target lost default project: %d, %v", projects.Len(), err)
				}
				if projects.compilerPrograms[0].GetSourceFile(files[1]) == nil {
					t.Fatal("fixture must contain the disabled target in the ordinary Program")
				}
				for _, load := range []func(ProjectSet, target.Plan, string, bool) (LoadResult, error){session.LoadCLI, session.LoadAPI} {
					binding, err := load(projects, plan, dir, true)
					if err != nil {
						t.Fatal(err)
					}
					if len(binding.Programs) != 2 || !slices.Equal(binding.TargetsByProgram[0], files[:1]) || !slices.Equal(binding.TargetsByProgram[1], files[1:]) {
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
	for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
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

func TestLoadProgramsSingleCandidateSkipsRootRanking(t *testing.T) {
	dir := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/project_service_modes.txtar").Materialize(t, ""))
	config := rslintconfig.RslintConfig{{LanguageOptions: &rslintconfig.LanguageOptions{ParserOptions: &rslintconfig.ParserOptions{ProjectService: rslintconfig.BoolPtr(true)}}}}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	requestPlan, err := target.Resolve(target.Request{Config: config, ConfigDirectory: dir, FS: fsys, Files: []string{tspath.ResolvePath(dir, "a.ts")}})
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession(fsys)
	projects, err := session.buildProjectsWithOptionsForTest(t, map[string]rslintconfig.RslintConfig{dir: config}, requestPlan, ActiveOwners, true)
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
			if !slices.Equal(owners, []int{0}) || len(counting.calls) != 0 {
				t.Fatalf("single candidate performed root ranking: owners=%v Realpath=%v", owners, counting.calls)
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
	liveCanonicalPath string
	targetCalls       int
}

func (fsys *retargetingFrozenTargetFS) Realpath(filePath string) string {
	filePath = tspath.NormalizePath(filePath)
	if filePath == fsys.targetPath {
		fsys.targetCalls++
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

type streamingTargetProjectFS struct {
	vfs.FS
	firstSource  string
	laterConfig  string
	sourceRead   chan struct{}
	release      chan struct{}
	sourceOnce   sync.Once
	laterWaiting chan struct{}
	waitOnce     sync.Once
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

func (fsys *streamingTargetProjectFS) ReadFile(filePath string) (string, bool) {
	filePath = tspath.NormalizePath(filePath)
	if filePath == fsys.firstSource {
		fsys.sourceOnce.Do(func() { close(fsys.sourceRead) })
	}
	if filePath == fsys.laterConfig {
		fsys.waitOnce.Do(func() { close(fsys.laterWaiting) })
		select {
		case <-fsys.sourceRead:
		case <-fsys.release:
		}
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
	targeted, err := sessionForTest(targetedContext).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-a.json", "./tsconfig-b.json")}, Targets: plan, Scope: Targeted, SingleThreaded: true})
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

	set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: Targeted, SingleThreaded: true})
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
		t.Fatalf("project after the direct winner was parsed %d time(s)", got)
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
	all, err := sessionForTest(broadContext).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Scope: AllDeclared, SingleThreaded: true})
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

func TestBuildTargetProjectPredictionCannotOverrideEarlierDirectRoot(t *testing.T) {
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

	set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-first.json", "./nested/tsconfig.json")}, Targets: plan, Scope: Targeted, SingleThreaded: false})
	if err != nil {
		t.Fatalf("BuildTargetProject: %v", err)
	}
	if set.Len() != 1 {
		t.Fatalf("selected Programs = %d, want one", set.Len())
	}
	wantConfig := tspath.ResolvePath(dir, "tsconfig-first.json")
	if got := tspath.NormalizePath(set.compilerPrograms[0].Options().ConfigFilePath); got != wantConfig {
		t.Fatalf("predicted nested project overrode declaration order: got %q, want %q", got, wantConfig)
	}
}

func TestBuildTargetProjectIgnoresUnreachedPredictedConfigError(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"nested/target.ts":       `export const target = 1;`,
		"nested/unreadable.json": `{}`,
		"tsconfig-first.json":    `{"files":["nested/target.ts"],"compilerOptions":{"noLib":true}}`,
	})
	configPath := tspath.ResolvePath(dir, "nested/unreadable.json")
	fsys := &unreadableProjectConfigFS{
		FS:         bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		configPath: configPath,
	}
	context := newBuildContext(fsys)
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, filepath.Join(dir, "nested/target.ts")),
	}}

	set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-first.json", "./nested/unreadable.json")}, Targets: plan, Scope: Targeted, SingleThreaded: false})
	if err != nil {
		t.Fatalf("unreached speculative parse became observable: %v", err)
	}
	if set.Len() != 1 {
		t.Fatalf("selected Programs = %d, want one", set.Len())
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

	set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: Targeted, SingleThreaded: true})
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

func TestBuildTargetProjectSkipsImportFallbackWithUnsupportedExtension(t *testing.T) {
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
			set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig(test.project)}, Targets: plan, Scope: Targeted, SingleThreaded: false})
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
	set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-a.json", "./tsconfig-b.json")}, Targets: plan, Scope: Targeted, SingleThreaded: false})
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
		set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-a.json", "./tsconfig-b.json")}, Targets: plan, Scope: Targeted, SingleThreaded: false})
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
	set, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{
		rootDir:  projectConfig("./tsconfig.json"),
		childDir: projectConfig("../tsconfig.json"),
	}, Targets: plan, Scope: Targeted, SingleThreaded: false})
	if err != nil {
		t.Fatalf("BuildTargetProjects: %v", err)
	}
	if set.Len() != 1 || len(set.configOrders[0]) != 2 {
		t.Fatalf("shared direct winner was not deduplicated: Programs=%d orders=%v", set.Len(), set.configOrders)
	}
	binding, err := sessionForTest(context).LoadAPI(set, plan, rootDir, false)
	if err != nil {
		t.Fatalf("LoadAPI: %v", err)
	}
	if len(binding.TargetsByProgram) != 1 || len(binding.TargetsByProgram[0]) != 2 {
		t.Fatalf("shared direct winner lost an owner's target: %v", binding.TargetsByProgram)
	}
}

func TestBuildTargetProjectStreamsConfirmedBuildsDuringRootScan(t *testing.T) {
	previousGOMAXPROCS := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previousGOMAXPROCS) })
	dir := tspath.NormalizePath(t.TempDir())
	writeProgramTestFiles(t, dir, map[string]string{
		"a.ts":            `export const a = 1;`,
		"b.ts":            `export const b = 1;`,
		"tsconfig-a.json": `{"files":["a.ts"],"compilerOptions":{"noLib":true}}`,
		"tsconfig-b.json": `{"files":["b.ts"],"compilerOptions":{"noLib":true}}`,
	})
	aPath := tspath.ResolvePath(dir, "a.ts")
	bPath := tspath.ResolvePath(dir, "b.ts")
	fsys := &streamingTargetProjectFS{
		FS:           bundled.WrapFS(osvfs.FS()),
		firstSource:  aPath,
		laterConfig:  tspath.ResolvePath(dir, "tsconfig-b.json"),
		sourceRead:   make(chan struct{}),
		release:      make(chan struct{}),
		laterWaiting: make(chan struct{}),
	}
	context := newBuildContext(fsys)
	plan := target.Plan{Files: []target.File{
		testLintTarget(fsys, dir, aPath),
		testLintTarget(fsys, dir, bPath),
	}}
	done := make(chan error, 1)
	go func() {
		_, err := sessionForTest(context).BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: projectConfig("./tsconfig-a.json", "./tsconfig-b.json")}, Targets: plan, Scope: Targeted, SingleThreaded: false})
		done <- err
	}()

	select {
	case <-fsys.laterWaiting:
	case <-time.After(5 * time.Second):
		close(fsys.release)
		t.Fatal("root scan did not reach the later config")
	}
	select {
	case <-fsys.sourceRead:
		close(fsys.release)
	case <-time.After(5 * time.Second):
		close(fsys.release)
		t.Fatal("confirmed Program build did not overlap the later root scan")
	}
	if err := <-done; err != nil {
		t.Fatalf("BuildTargetProject: %v", err)
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
				FS:         baseFS,
				targetPath: requestedPath,
				liveCanonicalPath: tspath.NormalizePath(
					baseFS.Realpath(tspath.ResolvePath(liveDir, "target.ts")),
				),
			}
			plan := target.Plan{Files: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: requestedPath,
				CanonicalPath:       frozenCanonicalPath,
				CanonicalParentPath: frozenCanonicalParent}, ConfigDirectory: projectDir,
			}}}

			session := sessionForTest(newBuildContext(fsys))
			set, err := session.BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{projectDir: projectConfig("./tsconfig.json")}, Targets: plan, Scope: Targeted, SingleThreaded: true})
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
			if fsys.targetCalls != 0 {
				t.Fatalf("requested target Realpath calls = %d, want frozen identity only", fsys.targetCalls)
			}
		})
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

	initialSet, err := session.BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: Targeted, SingleThreaded: true})
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
	afterFixSet, err := session.BuildProjects(ProjectBuildRequest{Configs: map[string]rslintconfig.RslintConfig{dir: config}, Targets: plan, Scope: Targeted, SingleThreaded: true})
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

func TestTargetPlanActiveOwnersSelectOnlyGoverningConfigs(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/repo/a": {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
		"/repo/b": {{LanguageOptions: &rslintconfig.LanguageOptions{
			ParserOptions: &rslintconfig.ParserOptions{Project: []string{"./missing.json"}},
		}}},
	}
	active := configsForActiveOwners(configMap, target.Plan{Files: []target.File{{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/a/index.ts",
		CanonicalPath: "/repo/a/index.ts"}, ConfigDirectory: "/repo/a",
	}}})
	if len(active) != 1 || active["/repo/a"] == nil {
		t.Fatalf("expected only the governing config, got %v", active)
	}
	if _, present := active["/repo/b"]; present {
		t.Fatalf("inactive config unexpectedly selected: %v", active)
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
	emptySet, err := sessionForTest(newBuildContext(fsys)).BuildProjects(ProjectBuildRequest{Configs: configMap, Targets: target.Plan{}, Scope: ActiveOwners, SingleThreaded: true})
	if err != nil || len(emptySet.compilerPrograms) != 0 {
		t.Fatalf("an empty target plan must not build config projects: programs=%d err=%v", len(emptySet.compilerPrograms), err)
	}

	builders := []struct {
		name  string
		build func(*Session) (ProjectSet, error)
	}{
		{
			name: "all projects from active owners",
			build: func(session *Session) (ProjectSet, error) {
				return session.BuildProjects(ProjectBuildRequest{Configs: configMap, Targets: plan, Scope: ActiveOwners, SingleThreaded: true})
			},
		},
		{
			name: "targeted projects from active owners",
			build: func(session *Session) (ProjectSet, error) {
				return session.BuildProjects(ProjectBuildRequest{Configs: configMap, Targets: plan, Scope: Targeted, SingleThreaded: true})
			},
		},
	}
	for _, builder := range builders {
		t.Run(builder.name, func(t *testing.T) {
			set, err := builder.build(sessionForTest(newBuildContext(fsys)))
			if err != nil || len(set.compilerPrograms) != 1 {
				t.Fatalf("plain lint should build only the active config Program: programs=%d err=%v", len(set.compilerPrograms), err)
			}
		})
	}
	if _, err := sessionForTest(newBuildContext(fsys)).BuildProjects(ProjectBuildRequest{Configs: configMap, Scope: AllDeclared, SingleThreaded: true}); err == nil || !strings.Contains(err.Error(), "missing.json") {
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
	for _, scope := range []ProjectScope{AllDeclared, ActiveOwners, Targeted} {
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
