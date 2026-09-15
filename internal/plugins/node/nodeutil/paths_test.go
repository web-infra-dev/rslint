// cspell:ignore npmignore
package nodeutil

import (
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func packageRoot(t testing.TB) rule_tester.Root {
	t.Helper()
	base := fixtures.GetRootDir()
	directory := tspath.ResolvePath(base.Dir, "node-package-tests")
	return rule_tester.Root{Dir: directory, FS: utils.NewOverlayVFS(base.FS, map[string]string{
		tspath.ResolvePath(directory, "tsconfig.json"): `{"compilerOptions":{"allowJs":true,"noEmit":true},"include":["**/*"]}`,
	})}
}

func TestPackageGeneration(t *testing.T) {
	base := packageRoot(t)
	create := func(metadata string) *lintprogram.Program {
		t.Helper()
		root := base
		root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
			tspath.ResolvePath(root.Dir, "string-bin/package.json"):        metadata,
			tspath.ResolvePath(root.Dir, "string-bin/nested/package.json"): "null",
		})
		raw, _, err := rule_tester.NewProgramHelper(root).CreateTestProgram("hello();", "string-bin/nested/a.js", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		return lintprogram.NewFromCompiler(raw)
	}
	first := create(`{"bin":"bin/first.js"}`)
	name := tspath.ResolvePath(base.Dir, "string-bin/nested/a.js")
	var packages [16]*PackageJSON
	var group sync.WaitGroup
	for index := range packages {
		group.Go(func() {
			packages[index] = FindPackage(first, name)
		})
	}
	group.Wait()
	pkg := packages[0]
	if pkg == nil || pkg.data["bin"] != "bin/first.js" {
		t.Fatalf("invalid nested metadata did not fall back to parent: %#v", pkg)
	}
	for _, got := range packages {
		if got != pkg {
			t.Error("package decoding was not shared within the Program")
		}
	}
	second := create(`{"bin":"bin/second.js"}`)
	if got := FindPackage(second, name); got == nil || got == pkg || got.data["bin"] != "bin/second.js" {
		t.Fatalf("package metadata leaked between Programs: %#v", got)
	}
}

func TestBinAliases(t *testing.T) {
	for _, bin := range []any{[]any{"bin/test"}, map[string]any{"cli": "bin/test"}, "bin/test"} {
		if !isBinFile("/pkg/bin/test.js", bin, "/pkg") {
			t.Errorf("bin resolution failed: %#v", bin)
		}
	}
	for _, bin := range []any{nil, false, 42, "", map[string]any{"cli": false}} {
		if isBinFile("/pkg/bin/test.js", bin, "/pkg") {
			t.Errorf("invalid bin matched: %#v", bin)
		}
	}
}

func TestWorkspaceDependencyPatterns(t *testing.T) {
	for _, test := range []struct {
		name     string
		patterns any
		want     bool
	}{
		{"packages/app", []any{`./{packages,apps}/*/`}, true},
		{"packages/app", map[string]any{"packages": []any{`packages/@(app|lib)`}}, true},
		{"packages/.hidden", []any{`packages/*`}, true},
		// Upstream treats 1..3 literally.
		{"packages/2", []any{`packages/{1..3}`}, true},
		// Upstream treats ^ as a literal class member; minimatch negates it.
		{"packages/app", []any{`packages/[^a]*`}, false},
		{"packages/lib", []any{`packages/[^a]*`}, true},
		// The ! form negates character classes in both matchers.
		{"packages/app", []any{`packages/[!a]*`}, false},
		{"packages/lib", []any{`packages/[!a]*`}, true},
		{"packages/excluded", []any{`!packages/excluded`, `packages/*`}, false},
		{"packages/app", []any{`packages\*\`}, true},
		{"123", []any{123.0}, true},
		{"packages/app", []any{"", "!", "packages/other"}, false},
		{"packages/app", map[string]any{"packages": false}, false},
	} {
		if got := matchesWorkspace(test.name, test.patterns); got != test.want {
			t.Errorf("matchesWorkspace(%q, %#v) = %v, want %v", test.name, test.patterns, got, test.want)
		}
	}
}

func TestConvertPath(t *testing.T) {
	for _, test := range []struct{ pattern, input, want string }{
		{"src/**", "src/bin/test.js", "bin/test.js"},
		{`src/\*.js`, "src/*.js", "src/*.js"},
		{`src/\*.js`, "src/a.js", "src/a.js"},
		{"src/file?.js", "src/file?.js", "file?.js"},
		{"src/file[ab].js", "src/file[ab].js", "file[ab].js"},
		{"src/a.js ", "src/a.js ", "a.js "},
		{"src/a.js ", "src/a.js", "src/a.js"},
		{"src/file?.js", "src/file1.js", "src/file1.js"},
		{"src/[ab].js", "src/a.js", "src/a.js"},
		{"src/{a,b}.js", "src/a.js", "src/a.js"},
		{"src/**", "src/.hidden.js", ".hidden.js"},
	} {
		got, ok := ConvertPath(test.input, map[string]any{"convertPath": map[string]any{test.pattern: []any{"^src/", ""}}}, nil)
		if !ok || got != test.want {
			t.Errorf("%q with %q = %q, %v; want %q", test.input, test.pattern, got, ok, test.want)
		}
	}
	// Object order is unavailable after config decoding. Use array form when
	// overlapping patterns need priority; this is an explicit documented difference.
	mapping := map[string]any{"src/**": []any{"^src/", "first/"}, "**": []any{"^src/", "second/"}}
	got, ok := ConvertPath("src/a.js", map[string]any{"convertPath": mapping}, nil)
	if !ok || got != "second/a.js" {
		t.Fatalf("object order = %q, %v", got, ok)
	}
	ordered := []any{map[string]any{"include": []any{"src/**"}, "replace": []any{"^src/", "first/"}}, map[string]any{"include": []any{"**"}, "replace": []any{"^src/", "second/"}}}
	got, ok = ConvertPath("src/a.js", map[string]any{"convertPath": ordered}, nil)
	if !ok || got != "first/a.js" {
		t.Fatalf("array order = %q, %v", got, ok)
	}
	// Named and unnamed captures share JavaScript's lexical numbering.
	got, ok = ConvertPath("src/cli.js", map[string]any{"convertPath": map[string]any{"**": []any{`(?<dir>src)/(.*)`, "$1/$2"}}}, nil)
	if !ok || got != "src/cli.js" {
		t.Fatalf("mixed capture numbering = %q, %v", got, ok)
	}
}

func TestPublicationGenerationAndConfiguration(t *testing.T) {
	base := packageRoot(t)
	create := func(ignore string) *lintprogram.Program {
		t.Helper()
		root := base
		root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
			tspath.ResolvePath(root.Dir, "pkg/package.json"):        `{"files":["lib"],"main":"main.js"}`,
			tspath.ResolvePath(root.Dir, "pkg/.npmignore"):          ignore,
			tspath.ResolvePath(root.Dir, "pkg/nested/package.json"): `{"files":["lib"]}`,
			tspath.ResolvePath(root.Dir, "empty/package.json"):      `{}`,
		})
		raw, _, err := rule_tester.NewProgramHelper(root).CreateTestProgram("hello();", "pkg/lib/a.js", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		return lintprogram.NewFromCompiler(raw)
	}
	first, second := create("lib/a.js"), create("lib/b.js")
	file := tspath.ResolvePath(base.Dir, "pkg/lib/a.js")
	directory := tspath.GetDirectoryPath(tspath.GetDirectoryPath(file))
	var group sync.WaitGroup
	for range 16 {
		group.Go(func() {
			if !IsUnpublished(first, file, "lib/a.js") || IsUnpublished(second, file, "lib/a.js") {
				t.Error("ignore state leaked across Program generations")
			}
			// One Program can serve different configured files.
			if !MatchIgnorePatterns(first, []string{"lib/**"}, "lib/a.js") || MatchIgnorePatterns(first, []string{"lib/**", "!lib/a.js"}, "lib/a.js") {
				t.Error("pattern order or configuration was lost")
			}
			if MatchIgnorePatterns(first, []string{"*.js\n!a.js"}, "b.js") || !MatchIgnorePatterns(first, []string{"*.js", "!a.js"}, "b.js") {
				t.Error("pattern list boundaries collided")
			}
			if IsUnpublished(first, tspath.ResolvePath(directory, "README.js"), "README.js") {
				t.Error("root README was affected by the Program working directory")
			}
			if IsUnpublished(first, tspath.ResolvePath(directory, "main.js"), "main.js") || IsUnpublished(first, tspath.ResolvePath(directory, "package.json"), "package.json") {
				t.Error("always-published metadata was excluded")
			}
			if IsUnpublished(first, tspath.ResolvePath(directory, "nested/lib/a.js"), "nested/lib/a.js") {
				t.Error("converted path did not use the nested package's relative path")
			}
			if IsUnpublished(first, tspath.ResolvePath(base.Dir, "empty/..hidden.js"), "..hidden.js") || !IsUnpublished(first, file, "../lib/a.js") || !IsUnpublished(first, file, "dir/../../lib/a.js") {
				t.Error("ancestor detection did not respect path components")
			}
		})
	}
	group.Wait()
}
