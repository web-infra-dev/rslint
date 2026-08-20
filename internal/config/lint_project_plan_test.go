package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

type projectPlanBarrierFS struct {
	vfs.FS
	path    string
	enabled atomic.Bool
	entered chan struct{}
	release chan struct{}
}

type projectPlanOrderedErrorFS struct {
	vfs.FS
	firstPath     string
	secondPath    string
	thirdPath     string
	firstEntered  chan struct{}
	secondEntered chan struct{}
	thirdEntered  chan struct{}
	releaseFirst  chan struct{}
	firstOnce     sync.Once
	secondOnce    sync.Once
	thirdOnce     sync.Once
}

func (fsys *projectPlanBarrierFS) Realpath(path string) string {
	if fsys.enabled.Load() && tspath.NormalizePath(path) == fsys.path {
		fsys.entered <- struct{}{}
		<-fsys.release
	}
	return fsys.FS.Realpath(path)
}

func (fsys *projectPlanOrderedErrorFS) Realpath(path string) string {
	switch tspath.NormalizePath(path) {
	case fsys.firstPath:
		fsys.firstOnce.Do(func() { close(fsys.firstEntered) })
		<-fsys.releaseFirst
	case fsys.secondPath:
		fsys.secondOnce.Do(func() { close(fsys.secondEntered) })
	case fsys.thirdPath:
		fsys.thirdOnce.Do(func() { close(fsys.thirdEntered) })
	}
	return fsys.FS.Realpath(path)
}

func awaitProjectPlanEvent(t *testing.T, events <-chan struct{}) {
	t.Helper()
	select {
	case <-events:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for project-plan worker")
	}
}

func writeProjectPlanFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func projectEntry(paths ...string) *LanguageOptions {
	if paths == nil {
		paths = []string{}
	}
	return &LanguageOptions{ParserOptions: &ParserOptions{Project: paths}}
}

func TestProjectPathResolverUsesTargetEffectiveProject(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	for _, name := range []string{"tsconfig.json", "tsconfig-a.json", "tsconfig-b.json"} {
		writeProjectPlanFile(t, filepath.Join(dir, name), `{}`)
	}
	config := RslintConfig{
		{Files: []string{"**/*.ts"}, LanguageOptions: projectEntry("./tsconfig-a.json")},
		{Files: []string{"default/**"}},
		{Files: []string{"tests/**"}, LanguageOptions: projectEntry("./tsconfig-b.json")},
		{Files: []string{"inherit/**"}, LanguageOptions: &LanguageOptions{ParserOptions: &ParserOptions{Project: nil}}},
		{Files: []string{"clear/**"}, LanguageOptions: projectEntry()},
	}
	fsys := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	resolver, err := NewProjectPathResolver(nil, config, dir, fsys, false)
	if err != nil {
		t.Fatal(err)
	}
	targets := []struct {
		name      string
		relative  string
		want      string
		wantState ProjectDeclarationState
		effective bool
	}{
		{name: "base", relative: "src/a.ts", want: "tsconfig-a.json", wantState: ProjectExplicit, effective: true},
		{name: "override", relative: "tests/a.ts", want: "tsconfig-b.json", wantState: ProjectExplicit, effective: true},
		{name: "nil inherits", relative: "inherit/a.ts", want: "tsconfig-a.json", wantState: ProjectExplicit, effective: true},
		{name: "empty clears then defaults", relative: "clear/a.ts", want: "tsconfig.json", wantState: ProjectExplicitEmpty, effective: true},
		{name: "matched unspecified defaults", relative: "default/a.js", want: "tsconfig.json", wantState: ProjectUnspecified, effective: true},
		{name: "no matching entry", relative: "outside.js", wantState: ProjectUnspecified},
	}
	for _, test := range targets {
		t.Run(test.name, func(t *testing.T) {
			path := tspath.ResolvePath(dir, test.relative)
			planned, err := resolver.ResolveLintTarget(DiscoveredLintTarget{
				Path: path, CanonicalPath: path, ConfigDirectory: dir,
			})
			if err != nil {
				t.Fatal(err)
			}
			if (planned.Effective != nil) != test.effective {
				t.Fatalf("effective config present = %t, want %t", planned.Effective != nil, test.effective)
			}
			if got := planned.Effective.ProjectState(); got != test.wantState {
				t.Fatalf("project state = %v, want %v", got, test.wantState)
			}
			if test.want == "" {
				if len(planned.ProjectPaths) != 0 {
					t.Fatalf("project paths = %v, want none", planned.ProjectPaths)
				}
				return
			}
			want := tspath.ResolvePath(dir, test.want)
			if !slices.Equal(planned.ProjectPaths, []string{want}) {
				t.Fatalf("project paths = %v, want [%s]", planned.ProjectPaths, want)
			}
		})
	}
}

func TestProjectPathResolverPreservesProjectShapeFromJSON(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	for _, name := range []string{"tsconfig.json", "tsconfig-a.json"} {
		writeProjectPlanFile(t, filepath.Join(dir, name), `{}`)
	}
	var config RslintConfig
	if err := json.Unmarshal([]byte(`[
		{"files":["**/*.ts"],"languageOptions":{"parserOptions":{"project":["./tsconfig-a.json"]}}},
		{"files":["inherit/**"],"languageOptions":{"parserOptions":{}}},
		{"files":["clear/**"],"languageOptions":{"parserOptions":{"project":[]}}}
	]`), &config); err != nil {
		t.Fatal(err)
	}
	if config[2].LanguageOptions.ParserOptions.Project == nil {
		t.Fatal("JSON project:[] was collapsed into an unspecified project")
	}
	resolver, err := NewProjectPathResolver(
		nil,
		config,
		dir,
		bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path string
		want string
	}{
		{path: "inherit/a.ts", want: "tsconfig-a.json"},
		{path: "clear/a.ts", want: "tsconfig.json"},
	}
	for _, test := range tests {
		target := tspath.ResolvePath(dir, test.path)
		planned, err := resolver.ResolveLintTarget(DiscoveredLintTarget{
			Path: target, CanonicalPath: target, ConfigDirectory: dir,
		})
		if err != nil {
			t.Fatal(err)
		}
		want := tspath.ResolvePath(dir, test.want)
		if !slices.Equal(planned.ProjectPaths, []string{want}) {
			t.Fatalf("%s projects = %v, want [%s]", test.path, planned.ProjectPaths, want)
		}
	}
}

func TestProjectPathResolverValidatesUnmatchedDeclarations(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	_, err := NewProjectPathResolver(nil, RslintConfig{{
		Files:           []string{"never/**"},
		LanguageOptions: projectEntry("./missing.json"),
	}}, dir, bundled.WrapFS(cachedvfs.From(osvfs.FS())), false)
	if err == nil {
		t.Fatal("unmatched invalid project declaration was not rejected")
	}
}

func TestProjectPathResolverBuildsTypeCheckCatalogPerOwner(t *testing.T) {
	root := tspath.NormalizePath(t.TempDir())
	child := tspath.ResolvePath(root, "packages/child")
	empty := tspath.ResolvePath(root, "packages/empty")
	shared := tspath.ResolvePath(root, "tsconfig.shared.json")
	for _, path := range []string{
		shared,
		tspath.ResolvePath(child, "tsconfig.json"),
		tspath.ResolvePath(empty, "tsconfig.json"),
	} {
		writeProjectPlanFile(t, path, `{}`)
	}
	configs := map[string]RslintConfig{
		root:  {{LanguageOptions: projectEntry(shared)}},
		child: {{}},
		empty: {{LanguageOptions: projectEntry()}},
	}
	resolver, err := NewProjectPathResolver(
		configs,
		nil,
		root,
		bundled.WrapFS(cachedvfs.From(osvfs.FS())),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		shared,
		tspath.ResolvePath(child, "tsconfig.json"),
		tspath.ResolvePath(empty, "tsconfig.json"),
	}
	got := resolver.CatalogProjectPaths()
	if !slices.Equal(got, want) {
		t.Fatalf("catalog projects = %v, want %v", got, want)
	}
}

func TestResolveLintProjectPlanHonorsParallelismAndTargetOrder(t *testing.T) {
	dir := tspath.NormalizePath(t.TempDir())
	config := RslintConfig{{Files: []string{"**/*.ts"}}}
	targetPlan := LintTargetPlan{Targets: []DiscoveredLintTarget{
		{Path: tspath.ResolvePath(dir, "first.ts"), CanonicalPath: tspath.ResolvePath(dir, "first.ts"), ConfigDirectory: dir},
		{Path: tspath.ResolvePath(dir, "second.ts"), CanonicalPath: tspath.ResolvePath(dir, "second.ts"), ConfigDirectory: dir},
		{Path: tspath.ResolvePath(dir, "third.ts"), CanonicalPath: tspath.ResolvePath(dir, "third.ts"), ConfigDirectory: dir},
	}}
	oldProcs := runtime.GOMAXPROCS(2)
	t.Cleanup(func() { runtime.GOMAXPROCS(oldProcs) })

	newBarrierResolver := func(t *testing.T) (*ProjectPathResolver, *projectPlanBarrierFS) {
		t.Helper()
		fsys := &projectPlanBarrierFS{
			FS:      bundled.WrapFS(cachedvfs.From(osvfs.FS())),
			path:    dir,
			entered: make(chan struct{}, len(targetPlan.Targets)),
			release: make(chan struct{}),
		}
		resolver, err := NewProjectPathResolver(nil, config, dir, fsys, false)
		if err != nil {
			t.Fatal(err)
		}
		fsys.enabled.Store(true)
		return resolver, fsys
	}

	t.Run("parallel", func(t *testing.T) {
		resolver, fsys := newBarrierResolver(t)
		type result struct {
			plan LintProjectPlan
			err  error
		}
		done := make(chan result, 1)
		go func() {
			plan, err := resolver.ResolveLintProjectPlan(targetPlan, false)
			done <- result{plan: plan, err: err}
		}()
		awaitProjectPlanEvent(t, fsys.entered)
		awaitProjectPlanEvent(t, fsys.entered)
		select {
		case <-fsys.entered:
			t.Fatal("planning exceeded the GOMAXPROCS worker bound")
		default:
		}
		close(fsys.release)
		resolved := <-done
		if resolved.err != nil {
			t.Fatal(resolved.err)
		}
		for index, target := range targetPlan.Targets {
			if got := resolved.plan.Targets[index].Target.Path; got != target.Path {
				t.Fatalf("target %d path = %q, want %q", index, got, target.Path)
			}
		}
		awaitProjectPlanEvent(t, fsys.entered)
	})

	t.Run("single threaded", func(t *testing.T) {
		resolver, fsys := newBarrierResolver(t)
		done := make(chan error, 1)
		go func() {
			_, err := resolver.ResolveLintProjectPlan(targetPlan, true)
			done <- err
		}()
		awaitProjectPlanEvent(t, fsys.entered)
		select {
		case <-fsys.entered:
			t.Fatal("single-threaded planning started a second target before the first completed")
		default:
		}
		close(fsys.release)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("errors retain target order", func(t *testing.T) {
		firstOwner := tspath.ResolvePath(dir, "missing-owner-first")
		secondOwner := tspath.ResolvePath(dir, "missing-owner-second")
		thirdOwner := tspath.ResolvePath(dir, "missing-owner-third")
		fsys := &projectPlanOrderedErrorFS{
			FS:            osvfs.FS(),
			firstPath:     firstOwner,
			secondPath:    secondOwner,
			thirdPath:     thirdOwner,
			firstEntered:  make(chan struct{}),
			secondEntered: make(chan struct{}),
			thirdEntered:  make(chan struct{}),
			releaseFirst:  make(chan struct{}),
		}
		resolver, err := NewProjectPathResolver(nil, config, dir, fsys, false)
		if err != nil {
			t.Fatal(err)
		}
		first := tspath.ResolvePath(dir, "missing-first.ts")
		second := tspath.ResolvePath(dir, "missing-second.ts")
		third := tspath.ResolvePath(dir, "missing-third.ts")
		result := make(chan error, 1)
		go func() {
			_, resolveErr := resolver.ResolveLintProjectPlan(LintTargetPlan{Targets: []DiscoveredLintTarget{
				{Path: first, CanonicalPath: first, ConfigDirectory: firstOwner},
				{Path: second, CanonicalPath: second, ConfigDirectory: secondOwner},
				{Path: third, CanonicalPath: third, ConfigDirectory: thirdOwner},
			}}, false)
			result <- resolveErr
		}()
		awaitProjectPlanEvent(t, fsys.firstEntered)
		awaitProjectPlanEvent(t, fsys.secondEntered)
		// With the first worker blocked, reaching the third target proves the
		// second target completed before the first error was released.
		awaitProjectPlanEvent(t, fsys.thirdEntered)
		close(fsys.releaseFirst)
		err = <-result
		if err == nil || !strings.Contains(err.Error(), first) {
			t.Fatalf("error = %v, want first target %q", err, first)
		}
	})

	t.Run("single threaded stops at first error", func(t *testing.T) {
		firstOwner := tspath.ResolvePath(dir, "serial-missing-owner-first")
		secondOwner := tspath.ResolvePath(dir, "serial-missing-owner-second")
		fsys := &projectPlanOrderedErrorFS{
			FS:            osvfs.FS(),
			firstPath:     firstOwner,
			secondPath:    secondOwner,
			firstEntered:  make(chan struct{}),
			secondEntered: make(chan struct{}),
			thirdEntered:  make(chan struct{}),
			releaseFirst:  make(chan struct{}),
		}
		close(fsys.releaseFirst)
		resolver, err := NewProjectPathResolver(nil, config, dir, fsys, false)
		if err != nil {
			t.Fatal(err)
		}
		first := tspath.ResolvePath(dir, "serial-missing-first.ts")
		second := tspath.ResolvePath(dir, "serial-missing-second.ts")
		_, err = resolver.ResolveLintProjectPlan(LintTargetPlan{Targets: []DiscoveredLintTarget{
			{Path: first, CanonicalPath: first, ConfigDirectory: firstOwner},
			{Path: second, CanonicalPath: second, ConfigDirectory: secondOwner},
		}}, true)
		if err == nil || !strings.Contains(err.Error(), first) {
			t.Fatalf("error = %v, want first target %q", err, first)
		}
		select {
		case <-fsys.secondEntered:
			t.Fatal("single-threaded planning reached a target after the first error")
		default:
		}
	})
}
