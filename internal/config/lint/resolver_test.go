package lint

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func newBaseResolver(options ResolverOptions) *Resolver {
	options.Catalog = rules.All()
	configs := options.ConfigsByOwner
	if configs == nil {
		configs = map[string]config.RslintConfig{
			options.ConfigDirectory: options.Config,
		}
	}
	options.PathSpaces = config.NewPathSpaceSnapshot(configs, options.FS)
	return NewResolver(options)
}

func TestResolverRequiresCatalog(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil rule catalog to panic")
		}
	}()
	NewResolver(ResolverOptions{})
}

func TestResolverRequiresPathSpaceSnapshot(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil path-space snapshot to panic")
		}
	}()
	NewResolver(ResolverOptions{Catalog: rules.All()})
}

func TestResolverUsesOnlyBoundTargetOwnership(t *testing.T) {
	configs := map[string]config.RslintConfig{
		"/repo": {{
			Files: []string{"**/*.ts"},
			Rules: config.Rules{"no-console": "error"},
		}},
		"/repo/packages/app": {{
			Files:   []string{"**/*.ts"},
			Plugins: []string{"@typescript-eslint"},
			Rules: config.Rules{
				"@typescript-eslint/require-await": "error",
				"no-debugger":                      "error",
			},
		}},
	}
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner: configs,
		TargetsBySourcePath: map[string]target.File{
			"/repo/packages/app/src/gap.ts":   targetForTest("/repo/packages/app/src/gap.ts", "/repo/packages/app"),
			"/repo/packages/app/src/typed.ts": targetForTest("/repo/packages/app/src/typed.ts", "/repo/packages/app"),
			"/repo/root.ts":                   targetForTest("/repo/root.ts", "/repo"),
		},
	})

	gapRules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath("/repo/packages/app/src/gap.ts"))
	if !gapRules["@typescript-eslint/require-await"] || !gapRules["no-debugger"] || gapRules["no-console"] {
		t.Fatalf("app target resolved against the wrong owner: %v", gapRules)
	}
	typedRules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath("/repo/packages/app/src/typed.ts"))
	if !typedRules["@typescript-eslint/require-await"] || !typedRules["no-debugger"] {
		t.Fatalf("typed app target lost its owner rules: %v", typedRules)
	}
	rootRules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath("/repo/root.ts"))
	if !rootRules["no-console"] || rootRules["no-debugger"] {
		t.Fatalf("root target resolved against the wrong owner: %v", rootRules)
	}
	if rules := resolver.EnabledRulesForSourcePath("/outside/a.ts"); len(rules) != 0 {
		t.Fatalf("unbound multi-config source received rules: %v", rules)
	}
}

func TestResolverUsesBoundOwnerForAliasedSource(t *testing.T) {
	configs := map[string]config.RslintConfig{
		"/repo": {{
			Files: []string{"packages/app/*.ts"},
			Rules: config.Rules{"no-console": "error"},
		}},
		"/repo/packages/app": {{Rules: config.Rules{"no-debugger": "error"}}},
	}
	sourcePath := "/repo/packages/app/a.ts"
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner: configs,
		TargetsBySourcePath: map[string]target.File{
			sourcePath: targetForTest(sourcePath, "/repo"),
		},
	})

	rules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath(sourcePath))
	if !rules["no-console"] || rules["no-debugger"] {
		t.Fatalf("source path overrode the binding's owner: %v", rules)
	}
}

func TestResolverUsesBoundTargetForRulesAndGlobals(t *testing.T) {
	cfg := config.RslintConfig{{
		Files: []string{"src/**/*.ts"},
		LanguageOptions: &config.LanguageOptions{Raw: map[string]any{
			"globals": map[string]any{"aliasedGlobal": "readonly"},
		}},
		Rules: config.Rules{"no-console": "error"},
	}}
	resolver := newBaseResolver(ResolverOptions{
		Config:          cfg,
		ConfigDirectory: "/repo",
		TargetsBySourcePath: map[string]target.File{
			"/outside/real-a.ts": {
				PathIdentity: config.PathIdentity{
					Path:          "/repo/src/a.ts",
					CanonicalPath: "/outside/real-a.ts",
				},
				ConfigDirectory: "/repo",
			},
		},
	})

	rules := resolver.EnabledRulesForSourcePath("/outside/real-a.ts")
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("aliased source did not use its target config: %v", configuredRuleNameSet(rules))
	}
	if access := rules[0].Environment.Globals["aliasedGlobal"]; access != utils.GlobalAccessReadonly {
		t.Fatalf("aliased source lost target globals: %v", access)
	}
	_, resolved, ok := resolver.ResolveSourcePath("/outside/real-a.ts")
	if !ok || resolved.MergedConfig == nil {
		t.Fatal("aliased source did not resolve its merged config")
	}
}

type caseInsensitiveResolverFS struct {
	vfs.FS
}

func (fs *caseInsensitiveResolverFS) UseCaseSensitiveFileNames() bool { return false }
func (fs *caseInsensitiveResolverFS) Realpath(filePath string) string {
	return strings.ToLower(tspath.NormalizePath(filePath))
}

func TestResolverSourceMappingsUseCanonicalFilesystemIdentity(t *testing.T) {
	fsys := &caseInsensitiveResolverFS{FS: osvfs.FS()}
	sourcePath := "c:/repo/src/a.ts"
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner: map[string]config.RslintConfig{
			"C:/Repo": {{
				Files: []string{"src/**/*.ts"},
				Rules: config.Rules{"no-console": "error"},
			}},
		},
		TargetsBySourcePath: map[string]target.File{
			"C:/REPO/SRC/A.ts": targetForTest("c:/repo/src/a.ts", "c:/repo"),
		},
		FS: fsys,
	})

	rules := resolver.EnabledRulesForSourcePath(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("case-equivalent source mapping lost config rules: %v", configuredRuleNameSet(rules))
	}
}

func TestResolverSingleConfigAcceptsUnboundSource(t *testing.T) {
	resolver := newBaseResolver(ResolverOptions{
		ConfigDirectory: "/repo",
		Config: config.RslintConfig{{
			Rules: config.Rules{"no-debugger": "error"},
		}},
	})
	owner, resolved, ok := resolver.ResolveSourcePath("/repo/a.ts")
	if !ok || owner != "/repo" || len(resolved.EnabledRules) != 1 {
		t.Fatalf("single config did not accept an unbound source: owner=%q resolved=%+v ok=%v", owner, resolved, ok)
	}
}

func TestResolverTargetConfigSurvivesSourceBinding(t *testing.T) {
	for _, single := range []bool{true, false} {
		t.Run(map[bool]string{true: "single", false: "multiple"}[single], func(t *testing.T) {
			entries := config.RslintConfig{{Files: []string{"src/**/*.ts"}, Rules: config.Rules{"no-debugger": "error"}}}
			options := ResolverOptions{Config: entries, ConfigDirectory: "/repo"}
			if !single {
				options.ConfigsByOwner = map[string]config.RslintConfig{"/repo": entries}
			}
			resolver := newBaseResolver(options)
			file := target.File{PathIdentity: config.PathIdentity{
				Path: "/repo/src/file.ts", CanonicalPath: "/physical/file.ts", CanonicalParentPath: "/physical",
			}, ConfigDirectory: "/repo"}
			before, ok := resolver.ResolveTarget(file)
			if !ok || len(before.EnabledRules) != 1 {
				t.Fatalf("unbound target lost its config: %+v, %v", before, ok)
			}
			mapping := map[string]target.File{"/physical/file.ts": file}
			bound := resolver.WithSourceMappings(mapping, &caseInsensitiveResolverFS{FS: osvfs.FS()}, true)
			mapping["/physical/file.ts"] = target.File{}
			if bound.singleResolver != resolver.singleResolver ||
				bound.resolversByOwnerPath["/repo"] != resolver.resolversByOwnerPath["/repo"] {
				t.Fatal("source binding replaced the already-used file config resolver")
			}
			if _, ok := resolver.TargetForSourcePath("/physical/file.ts"); ok {
				t.Fatal("source binding mutated the unbound resolver")
			}
			gotTarget, ok := bound.TargetForSourcePath("/physical/file.ts")
			if !ok || gotTarget != file {
				t.Fatalf("binding lost the frozen target identity: %+v, %v", gotTarget, ok)
			}
			owner, after, ok := bound.ResolveSourcePath("/physical/file.ts")
			if !ok || owner != "/repo" || after.MergedConfig != before.MergedConfig {
				t.Fatalf("binding did not reuse the effective config: owner=%q before=%p after=%p", owner, before.MergedConfig, after.MergedConfig)
			}
		})
	}
}

type ownerAliasResolverFS struct {
	vfs.FS
	root string
}

func (fs *ownerAliasResolverFS) Realpath(path string) string {
	alias := fs.root + "/symlink"
	if path == alias || strings.HasPrefix(path, alias+"/") {
		return fs.root + "/real" + strings.TrimPrefix(path, alias)
	}
	return path
}

func TestResolverLiteralOwnerWinsCanonicalAlias(t *testing.T) {
	for _, root := range []string{"", "C:"} {
		t.Run(root, func(t *testing.T) {
			service := true
			resolver := newBaseResolver(ResolverOptions{
				ConfigsByOwner: map[string]config.RslintConfig{
					root + "/real": {{Rules: config.Rules{"no-debugger": "error"}}},
					root + "/symlink": {{Rules: config.Rules{"no-console": "error"}, LanguageOptions: &config.LanguageOptions{
						ParserOptions: &config.ParserOptions{ProjectService: &service},
					}}},
				},
				FS: &ownerAliasResolverFS{FS: osvfs.FS(), root: root},
			})
			realTarget := targetForTest(root+"/real/file.ts", strings.ToLower(root)+"/real")
			aliasTarget := targetForTest(root+"/symlink/file.ts", root+"/symlink")
			realResolved, _ := resolver.ResolveTarget(realTarget)
			alias, _ := resolver.ResolveTarget(aliasTarget)
			if len(realResolved.EnabledRules) != 1 || realResolved.EnabledRules[0].Name != "no-debugger" ||
				len(alias.EnabledRules) != 1 || alias.EnabledRules[0].Name != "no-console" {
				t.Fatalf("canonical alias changed literal owner: real=%v alias=%v", configuredRuleNameSet(realResolved.EnabledRules), configuredRuleNameSet(alias.EnabledRules))
			}
			policies, err := resolver.ProjectPolicies([]target.File{realTarget, aliasTarget}, false)
			if err != nil || len(policies) != 2 || policies[realTarget] != (config.ProjectPolicy{}) || policies[aliasTarget].ServiceRootDirectory != root+"/symlink" {
				t.Fatalf("policy gate used an alias instead of its literal owner: %v, %v", policies, err)
			}
		})
	}
}

func TestResolverProjectPoliciesUsesEffectiveConfig(t *testing.T) {
	for _, test := range []struct {
		name, input string
		want        config.ProjectPolicy
		error       string
	}{
		{name: "ordinary project uses final declaration", input: `[{"languageOptions":{"parserOptions":{"project":["first.json"]}}},{"languageOptions":{"parserOptions":{"project":["second.json"]}}}]`, want: config.ProjectPolicy{ExplicitProject: &config.ProjectDeclaration{Patterns: config.ProjectPaths{"second.json"}, BaseDirectory: "/repo"}}},
		{name: "unmatched options are neutral", input: `[{"files":["unused.ts"],"languageOptions":{"parserOptions":{"projectService":true,"tsconfigRootDir":"relative","project":true}}}]`},
		{name: "empty match remains zero", input: `[{"files":["unused.ts"],"languageOptions":{"parserOptions":{"project":false}}}]`},
		{name: "matched service", input: `[{"languageOptions":{"parserOptions":{"projectService":true}}}]`, want: config.ProjectPolicy{ServiceRootDirectory: "/repo"}},
		{name: "service false does not resurrect unmatched declaration", input: `[{"files":["unused.ts"],"languageOptions":{"parserOptions":{"project":["first.json"]}}},{"languageOptions":{"parserOptions":{"projectService":false}}}]`, want: config.ProjectPolicy{DefaultProjectDisabled: true}},
		{name: "matched reset", input: `[{"languageOptions":{"parserOptions":{"project":null}}}]`, want: config.ProjectPolicy{ProjectDisabled: true}},
		{name: "root error includes target", input: `[{"languageOptions":{"parserOptions":{"tsconfigRootDir":"relative"}}}]`, error: "absolute path"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var entries config.RslintConfig
			if err := json.Unmarshal([]byte(test.input), &entries); err != nil {
				t.Fatal(err)
			}
			file := targetForTest("/repo/target.ts", "/repo")
			resolver := newBaseResolver(ResolverOptions{ConfigsByOwner: map[string]config.RslintConfig{"/repo": entries}})
			policies, err := resolver.ProjectPolicies([]target.File{file}, false)
			if test.error != "" {
				if err == nil || !strings.Contains(err.Error(), file.Path) || !strings.Contains(err.Error(), test.error) {
					t.Fatalf("target policy error=%v", err)
				}
				return
			}
			_, present := policies[file]
			if err != nil || !present || len(policies) != 1 || !reflect.DeepEqual(policies[file], test.want) {
				t.Fatalf("policies=%v error=%v, want %+v", policies, err, test.want)
			}
		})
	}
}

func TestResolverProjectPoliciesKeepsExplicitGapsPerTarget(t *testing.T) {
	var entries config.RslintConfig
	if err := json.Unmarshal([]byte(`[
		{"rules":{"no-debugger":"error"}},
		{"files":["typed.ts"],"languageOptions":{"parserOptions":{"project":"project.json"}}},
		{"files":["cleared.ts"],"languageOptions":{"parserOptions":{"project":null}}}
	]`), &entries); err != nil {
		t.Fatal(err)
	}
	resolver := newBaseResolver(ResolverOptions{Config: entries, ConfigDirectory: "/repo"})
	files := []target.File{
		targetForTest("/repo/typed.ts", "/repo"),
		targetForTest("/repo/cleared.ts", "/repo"),
		targetForTest("/repo/gap.ts", "/repo"),
	}
	policies, err := resolver.ProjectPolicies(files, false)
	want := map[target.File]config.ProjectPolicy{
		files[0]: {ExplicitProject: &config.ProjectDeclaration{Patterns: config.ProjectPaths{"project.json"}, BaseDirectory: "/repo"}},
		files[1]: {ProjectDisabled: true},
		files[2]: {},
	}
	if err != nil || !reflect.DeepEqual(policies, want) {
		t.Fatalf("policies=%v error=%v, want explicit policies for all targets: %v", policies, err, want)
	}
}

func TestResolverProjectPoliciesKeepsComposedOrigins(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	module, invocation := root+"/module", root+"/invocation"
	for _, replace := range []bool{false, true} {
		t.Run(strconv.FormatBool(replace), func(t *testing.T) {
			entries := config.ConfigWithAuthoredPathBase(config.RslintConfig{{
				LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{"./tsconfig.json"}}},
			}}, module)
			inline := config.ConfigEntry{Rules: config.Rules{"no-debugger": "error"}}
			if replace {
				inline.LanguageOptions = &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{"./tsconfig.json"}}}
			}
			entries = append(entries, config.ConfigWithAuthoredPathBase(config.RslintConfig{inline}, invocation)...)
			resolver := newBaseResolver(ResolverOptions{Config: entries, ConfigDirectory: module, DefaultRootDirectory: invocation})
			file := targetForTest(tspath.ResolvePath(invocation, "target.ts"), module)
			policies, err := resolver.ProjectPolicies([]target.File{file}, false)
			wantBase := module
			if replace {
				wantBase = invocation
			}
			want := config.ProjectPolicy{ExplicitProject: &config.ProjectDeclaration{Patterns: config.ProjectPaths{"./tsconfig.json"}, BaseDirectory: wantBase}}
			if err != nil || !reflect.DeepEqual(policies[file], want) {
				t.Fatalf("policy=%+v error=%v, want %+v", policies[file], err, want)
			}
			bound := resolver.WithSourceMappings(map[string]target.File{tspath.ResolvePath(root, "source-alias.ts"): file}, nil, true)
			boundPolicies, err := bound.ProjectPolicies([]target.File{file}, false)
			if err != nil || !reflect.DeepEqual(boundPolicies, policies) {
				t.Fatalf("source binding changed project origin: %v, %v; want %v", boundPolicies, err, policies)
			}
		})
	}
}

func TestResolverServiceRootUsesConfigSource(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	for _, test := range []struct {
		name, matchingDirectory, defaultRootDirectory string
	}{
		{name: "loaded module", matchingDirectory: root + "/config", defaultRootDirectory: root + "/config"},
		{name: "inline API uses invocation cwd", matchingDirectory: root + "/synthetic-owner", defaultRootDirectory: root + "/invocation"},
	} {
		for _, reset := range []bool{false, true} {
			t.Run(test.name+"/nullReset="+strconv.FormatBool(reset), func(t *testing.T) {
				input := `[{"languageOptions":{"parserOptions":{"projectService":true}}}]`
				if reset {
					input = fmt.Sprintf(`[{"languageOptions":{"parserOptions":{"projectService":true,"tsconfigRootDir":%q}}},{"languageOptions":{"parserOptions":{"tsconfigRootDir":null}}}]`, root+"/previous")
				}
				var entries config.RslintConfig
				if err := json.Unmarshal([]byte(input), &entries); err != nil {
					t.Fatal(err)
				}
				resolver := newBaseResolver(ResolverOptions{
					Config: entries, ConfigDirectory: test.matchingDirectory, DefaultRootDirectory: test.defaultRootDirectory,
				})
				file := targetForTest(test.matchingDirectory+"/file.ts", test.matchingDirectory)
				policies, err := resolver.ProjectPolicies([]target.File{file}, false)
				want := config.ProjectPolicy{ServiceRootDirectory: test.defaultRootDirectory}
				if err != nil || policies[file] != want {
					t.Fatalf("policy=%+v error=%v, want %+v", policies[file], err, want)
				}
				bound := resolver.WithSourceMappings(map[string]target.File{file.Path: file}, nil, true)
				boundPolicies, err := bound.ProjectPolicies([]target.File{file}, false)
				if err != nil || !reflect.DeepEqual(boundPolicies, policies) {
					t.Fatalf("source binding changed config defaults: %v, %v", boundPolicies, err)
				}
			})
		}
	}
}

func TestResolverServiceRootUsesEachModuleOwner(t *testing.T) {
	service := config.RslintConfig{{LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}}}}
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner:       map[string]config.RslintConfig{"/repo": service, "/repo/pkg": service},
		DefaultRootDirectory: "/unrelated-invocation",
	})
	files := []target.File{targetForTest("/repo/file.ts", "/repo"), targetForTest("/repo/pkg/file.ts", "/repo/pkg")}
	policies, err := resolver.ProjectPolicies(files, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if policies[file].ServiceRootDirectory != file.ConfigDirectory {
			t.Fatalf("nested owner used another module's default root: %v", policies)
		}
	}
}

func TestResolverServiceRootKeepsModuleOriginForCanonicalAlias(t *testing.T) {
	service := config.RslintConfig{{LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}}}}
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner: map[string]config.RslintConfig{"/symlink": service},
		FS:             &ownerAliasResolverFS{FS: osvfs.FS()},
	})
	file := targetForTest("/real/file.ts", "/real")
	policies, err := resolver.ProjectPolicies([]target.File{file}, false)
	if err != nil || policies[file].ServiceRootDirectory != "/symlink" {
		t.Fatalf("canonical lookup replaced the loaded module's origin: %v, %v", policies, err)
	}
}

func TestResolverRuleOverridePreservesProjectOptions(t *testing.T) {
	service, root := true, tspath.NormalizePath(t.TempDir())
	entries := config.RslintConfig{{
		Files: []string{"src/*.ts"}, Rules: config.Rules{"no-debugger": "warn"},
		LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: &service, TsconfigRootDir: &root}},
	}, {Files: []string{"unused/*.ts"}, LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{"unused.json"}}}}}
	file := targetForTest("/repo/src/file.ts", "/repo")
	before := newBaseResolver(ResolverOptions{Config: entries, ConfigDirectory: "/repo"})
	want, err := before.ProjectPolicies([]target.File{file}, false)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := config.BuildCLIRuleEntry([]string{"no-debugger: error"})
	if err != nil {
		t.Fatal(err)
	}
	entries = append(entries, *entry)
	entries, optionsErrors := config.ValidateRuleOptions(entries, rules.All())
	if len(optionsErrors) != 0 {
		t.Fatal(optionsErrors)
	}
	after := newBaseResolver(ResolverOptions{Config: entries, ConfigDirectory: "/repo"})
	got, err := after.ProjectPolicies([]target.File{file}, false)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("rule override changed parser policy or entry matching: %v, %v; want %v", got, err, want)
	}
	resolved, _ := after.ResolveTarget(file)
	if len(resolved.EnabledRules) != 1 || resolved.EnabledRules[0].Severity != rule.SeverityError {
		t.Fatalf("rule override was not applied: %v", resolved.EnabledRules)
	}
}

type projectPolicyResolverFS struct {
	vfs.FS
	realpath func(string) string
}

func (fs *projectPolicyResolverFS) Realpath(path string) string {
	if fs.realpath != nil {
		return fs.realpath(path)
	}
	return path
}

func TestResolverProjectPoliciesParallelProjection(t *testing.T) {
	previous := runtime.GOMAXPROCS(4)
	t.Cleanup(func() { runtime.GOMAXPROCS(previous) })
	root := tspath.NormalizePath(t.TempDir())
	owner, serviceOwner := root+"/explicit", root+"/service"
	configs := map[string]config.RslintConfig{
		owner: {
			{Rules: config.Rules{"no-debugger": "error"}},
			{Files: []string{"src/*.ts"}, LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{"project.json"}}}},
			{Files: []string{"cleared.ts"}, LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectDisabled: true}}},
		},
		serviceOwner: {
			{Ignores: []string{"ignored.ts"}},
			{LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{ProjectService: config.BoolPtr(true)}}},
		},
	}
	files := []target.File{
		targetForTest(owner+"/src/a.ts", owner),
		targetForTest(owner+"/src/b.ts", owner),
		targetForTest(owner+"/cleared.ts", owner),
		targetForTest(owner+"/gap.js", owner),
		targetForTest(serviceOwner+"/source.ts", serviceOwner),
		targetForTest(serviceOwner+"/ignored.ts", serviceOwner),
	}
	for index := range files {
		files[index].CanonicalParentPath = tspath.GetDirectoryPath(files[index].Path)
	}
	files = append(files, files[0])
	declaration := &config.ProjectDeclaration{Patterns: config.ProjectPaths{"project.json"}, BaseDirectory: owner}
	want := map[target.File]config.ProjectPolicy{
		files[0]: {ExplicitProject: declaration}, files[1]: {ExplicitProject: declaration},
		files[2]: {ProjectDisabled: true}, files[3]: {},
		files[4]: {ServiceRootDirectory: serviceOwner}, files[5]: {},
	}
	for _, singleThreaded := range []bool{false, true} {
		t.Run("singleThreaded="+strconv.FormatBool(singleThreaded), func(t *testing.T) {
			fsys := &projectPolicyResolverFS{FS: osvfs.FS()}
			resolver := newBaseResolver(ResolverOptions{ConfigsByOwner: configs, FS: fsys})
			fsys.realpath = func(string) string { panic("frozen target required another Realpath") }
			got, err := resolver.ProjectPolicies(files, singleThreaded)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("policies=%v error=%v, want %v", got, err, want)
			}
			first, _ := resolver.ResolveTarget(files[0])
			second, _ := resolver.ResolveTarget(files[1])
			if first.MergedConfig != second.MergedConfig || got[files[0]].ExplicitProject != got[files[1]].ExplicitProject {
				t.Fatal("same config shape was not reused")
			}
			empty, err := resolver.ProjectPolicies(nil, singleThreaded)
			if err != nil || empty == nil || len(empty) != 0 {
				t.Fatalf("empty input = %v, %v, want an empty policy map", empty, err)
			}
		})
	}
}

func TestResolverProjectPoliciesConcurrency(t *testing.T) {
	for _, mode := range []struct {
		name           string
		singleThreaded bool
		processors     int
		concurrent     int
	}{
		{name: "parallel", processors: 2, concurrent: 2},
		{name: "single threaded", singleThreaded: true, processors: 2, concurrent: 1},
		{name: "one processor", processors: 1, concurrent: 1},
	} {
		t.Run(mode.name, func(t *testing.T) {
			previous := runtime.GOMAXPROCS(mode.processors)
			t.Cleanup(func() { runtime.GOMAXPROCS(previous) })
			root := tspath.NormalizePath(t.TempDir())
			fsys := &projectPolicyResolverFS{FS: osvfs.FS()}
			resolver := newBaseResolver(ResolverOptions{ConfigDirectory: root, Config: config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}, FS: fsys})
			files := make([]target.File, 5)
			for index := range files {
				files[index] = targetForTest(fmt.Sprintf("%s/target-%d/file.ts", root, index), root)
				files[index].CanonicalPath = fmt.Sprintf("%s/physical-%d/file.ts", root, index)
			}
			entered, release := make(chan string, len(files)), make(chan struct{})
			var unblock sync.Once
			t.Cleanup(func() { unblock.Do(func() { close(release) }) })
			var active, maximum atomic.Int32
			var mu sync.Mutex
			var calls []string
			fsys.realpath = func(path string) string {
				current := active.Add(1)
				defer active.Add(-1)
				for old := maximum.Load(); current > old && !maximum.CompareAndSwap(old, current); old = maximum.Load() {
				}
				mu.Lock()
				calls = append(calls, path)
				mu.Unlock()
				entered <- path
				<-release
				return path
			}
			done := make(chan error, 1)
			go func() {
				_, err := resolver.ProjectPolicies(files, mode.singleThreaded)
				done <- err
			}()
			for range mode.concurrent {
				select {
				case <-entered:
				case <-time.After(10 * time.Second):
					t.Fatal("first config resolutions did not reach the concurrency barrier")
				}
			}
			select {
			case path := <-entered:
				t.Fatalf("unexpected concurrent resolution of %s", path)
			case <-time.After(20 * time.Millisecond):
			}
			unblock.Do(func() { close(release) })
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if maximum.Load() != int32(mode.concurrent) || len(calls) != len(files) {
				t.Fatalf("maximum/calls = %d/%d, want %d/%d", maximum.Load(), len(calls), mode.concurrent, len(files))
			}
			if mode.concurrent == 1 {
				for index, file := range files {
					if calls[index] != tspath.GetDirectoryPath(file.Path) {
						t.Fatalf("serial resolution order = %v", calls)
					}
				}
			}
		})
	}
}

func TestResolverProjectPoliciesFailureOrder(t *testing.T) {
	previous := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previous) })
	firstPanic, secondPanic := &struct{ name string }{"first"}, &struct{ name string }{"second"}
	for _, test := range []struct {
		name, first, second, errorText string
		trailing                       int
		panicValue                     any
	}{
		{name: "policy before panic", first: "z-bad", second: "panic", errorText: "absolute path"},
		{name: "policy before panic in same worker", first: "z-bad", second: "panic", errorText: "absolute path", trailing: 2},
		{name: "missing owner before panic", first: "missing", second: "panic", errorText: "missing governing configuration"},
		{name: "panic before policy", first: "panic", second: "z-bad", panicValue: firstPanic},
		{name: "panic before panic", first: "panic", second: "later-panic", panicValue: firstPanic},
		{name: "policy before policy", first: "z-bad", second: "a-bad", errorText: "absolute path"},
	} {
		for _, singleThreaded := range []bool{false, true} {
			t.Run(test.name+"/singleThreaded="+strconv.FormatBool(singleThreaded), func(t *testing.T) {
				root, relative := tspath.NormalizePath(t.TempDir()), "relative"
				fsys := &projectPolicyResolverFS{FS: osvfs.FS()}
				resolver := newBaseResolver(ResolverOptions{ConfigsByOwner: map[string]config.RslintConfig{root: {
					{Rules: config.Rules{"no-debugger": "error"}},
					{Files: []string{"*-bad/*.ts"}, LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{TsconfigRootDir: &relative}}},
				}}, FS: fsys})
				files := []target.File{targetForTest(root+"/"+test.first+"/file.ts", root), targetForTest(root+"/"+test.second+"/file.ts", root)}
				for index := range test.trailing {
					files = append(files, targetForTest(fmt.Sprintf("%s/safe-%d/file.ts", root, index), root))
				}
				for index := range files {
					files[index].CanonicalParentPath = tspath.GetDirectoryPath(files[index].Path)
					if strings.Contains(files[index].Path, "panic/") {
						files[index].CanonicalParentPath = ""
						files[index].CanonicalPath += ".physical"
					}
				}
				if test.first == "missing" {
					files[0].ConfigDirectory = root + "/missing"
				}
				var panicCalls atomic.Int32
				fsys.realpath = func(path string) string {
					panicCalls.Add(1)
					if path == root+"/panic" {
						panic(firstPanic)
					}
					panic(secondPanic)
				}
				var gotPanic any
				var err error
				func() {
					defer func() { gotPanic = recover() }()
					_, err = resolver.ProjectPolicies(files, singleThreaded)
				}()
				if gotPanic != test.panicValue {
					t.Fatalf("panic = %#v, want original value %#v (error %v)", gotPanic, test.panicValue, err)
				}
				if test.errorText != "" {
					if err == nil || !strings.HasPrefix(err.Error(), files[0].Path+": ") || !strings.Contains(err.Error(), test.errorText) {
						t.Fatalf("error = %v, want first input's %s", err, test.errorText)
					}
					if singleThreaded && panicCalls.Load() != 0 {
						t.Fatal("serial projection resolved beyond the first error")
					}
				} else if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func targetForTest(path string, owner string) target.File {
	return target.File{
		PathIdentity:    config.PathIdentity{Path: path, CanonicalPath: path},
		ConfigDirectory: owner,
	}
}

func configuredRuleNameSet(configuredRules []rule.ConfiguredRule) map[string]bool {
	names := make(map[string]bool, len(configuredRules))
	for _, configuredRule := range configuredRules {
		names[configuredRule.Name] = true
	}
	return names
}
