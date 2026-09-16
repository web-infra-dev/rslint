// cspell:ignore npmignore
package nodeutil

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
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
		if !isBinFile("/pkg/bin/test.js", bin, "/pkg", true) {
			t.Errorf("bin resolution failed: %#v", bin)
		}
	}
	for _, bin := range []any{nil, false, 42, "", map[string]any{"cli": false}} {
		if isBinFile("/pkg/bin/test.js", bin, "/pkg", true) {
			t.Errorf("invalid bin matched: %#v", bin)
		}
	}
	for _, sensitive := range []bool{true, false} {
		for _, test := range []struct {
			directory, file, bin string
			want                 bool
		}{
			{"/pkg", "/pkg/bin/CLI.js", "bin/cli.js", !sensitive},
			{"/pkg", "/pkg/bin/CLI.js", "bin/cli", !sensitive},
			{"/pkg", "/pkg/bin/CLI/index.js", "bin/cli", !sensitive},
			{"/pkg", "/pkg/BIN/CLI.JS", "bin/cli", !sensitive},
			{"/pkg", "/pkg/BIN/CLI/INDEX.JS", "bin/cli", !sensitive},
			{"/pkg", "/pkg/BIN/CLI/INDEX", "bin/cli", !sensitive},
			{"/pkg", "/pkg/bin/cli.jsx", "bin/cli", false},
			{"/pkg", "/pkg/bin/cli/myindex.js", "bin/cli", false},
			{"/pkg", "/pkg/bin/cli/index.js.js", "bin/cli", false},
			{"/pkg", "/pkg/中文😀/CLI/INDEX.JS", "中文😀/cli", !sensitive},
			{"/pkg", "/pkg/bin/İ.js", "bin/i.js", false},
			{"/pkg", "/pkg/bin/i.js", "bin/İ.js", false},
			{"/pkg", "/pkg/bin/K.js", "bin/k.js", !sensitive},
			{"C:/pkg", "c:/pkg/bin/cli.js", `bin\cli.js`, true},
			{"C:/pkg", "D:/pkg/bin/cli.js", "bin/cli.js", false},
			{"//server/share/pkg", "//SERVER/share/pkg/bin/cli.js", "bin/cli.js", true},
			{"//server/share/pkg", "//server/other/pkg/bin/cli.js", "bin/cli.js", false},
		} {
			if got := isBinFile(test.file, test.bin, test.directory, sensitive); got != test.want {
				t.Errorf("bin %q matching %q (case sensitive %v) = %v, want %v", test.bin, test.file, sensitive, got, test.want)
			}
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
			tspath.ResolvePath(root.Dir, "pkg/package.json"):            `{"files":["lib"],"main":"main.js"}`,
			tspath.ResolvePath(root.Dir, "pkg/.npmignore"):              ignore,
			tspath.ResolvePath(root.Dir, "pkg/nested/package.json"):     `{"files":["lib"],"main":"README.js"}`,
			tspath.ResolvePath(root.Dir, "pkg/lib/nested/package.json"): `{"files":[]}`,
			tspath.ResolvePath(root.Dir, "pkg/lib/nested/.npmignore"):   "ignored.js\n!restored.js",
			tspath.ResolvePath(root.Dir, "pkg/lib/git/.gitignore"):      "ignored.js",
			tspath.ResolvePath(root.Dir, "pkg/lib/blocked/.npmignore"):  "!restored.js",
			tspath.ResolvePath(root.Dir, "pkg/lib/empty/.npmignore"):    "",
			tspath.ResolvePath(root.Dir, "pkg/lib/empty/.gitignore"):    "ignored.js",
			tspath.ResolvePath(root.Dir, "empty/package.json"):          `{}`,
		})
		raw, _, err := rule_tester.NewProgramHelper(root).CreateTestProgram("hello();", "pkg/lib/a.js", "tsconfig.json")
		if err != nil {
			t.Fatal(err)
		}
		return lintprogram.NewFromCompiler(raw)
	}
	first, second := create("lib/a.js\nlib/nested/restored.js\nlib/blocked/"), create("lib/b.js")
	file := tspath.ResolvePath(base.Dir, "pkg/lib/a.js")
	directory := tspath.GetDirectoryPath(tspath.GetDirectoryPath(file))
	unpublished := func(p *lintprogram.Program, relative string) bool {
		return IsUnpublished(p, FindPackage(p, file), tspath.ResolvePath(directory, relative))
	}
	var group sync.WaitGroup
	for range 16 {
		group.Go(func() {
			if !unpublished(first, "lib/a.js") || unpublished(second, "lib/a.js") {
				t.Error("ignore state leaked across Program generations")
			}
			// One Program can serve different configured files.
			if !MatchIgnorePatterns(first, []string{"lib/**"}, "lib/a.js") || MatchIgnorePatterns(first, []string{"lib/**", "!lib/a.js"}, "lib/a.js") {
				t.Error("pattern order or configuration was lost")
			}
			if MatchIgnorePatterns(first, []string{"*.js\n!a.js"}, "b.js") || !MatchIgnorePatterns(first, []string{"*.js", "!a.js"}, "b.js") {
				t.Error("pattern list boundaries collided")
			}
			if unpublished(first, "README.js") {
				t.Error("root README was affected by the Program working directory")
			}
			if unpublished(first, "main.js") || unpublished(first, "package.json") {
				t.Error("always-published metadata was excluded")
			}
			if !unpublished(first, "nested/lib/a.js") || !unpublished(first, "nested/README.js") || unpublished(first, "lib/nested/private.js") {
				t.Error("nested package metadata changed the source package's publication policy")
			}
			if !unpublished(first, "lib/nested/ignored.js") || unpublished(first, "lib/nested/restored.js") || !unpublished(first, "lib/git/ignored.js") {
				t.Error("nested ignore files lost their matching or negation semantics")
			}
			if !unpublished(first, "lib/blocked/restored.js") || unpublished(first, "lib/empty/ignored.js") {
				t.Error("nested ignore files reopened a directory or ignored npmignore precedence")
			}
			if unpublished(first, "../pkg/lib/nested/private.js") || !unpublished(first, "../pkg-other/lib/a.js") || !unpublished(first, "D:/outside.js") {
				t.Error("converted target escaped the publishing package boundary")
			}
			if IsUnpublished(first, FindPackage(first, tspath.ResolvePath(base.Dir, "empty/..hidden.js")), tspath.ResolvePath(base.Dir, "empty/..hidden.js")) || !unpublished(first, "../lib/a.js") || !unpublished(first, "dir/../../lib/a.js") {
				t.Error("ancestor detection did not respect path components")
			}
		})
	}
	group.Wait()
}

// Expectations from Node's path.posix.isAbsolute and path.win32.isAbsolute.
func TestIsAbsolutePath(t *testing.T) {
	for _, tc := range []struct {
		name           string
		posix, windows bool
	}{
		{"", false, false}, {".", false, false}, {"../module", false, false},
		{"file:///a", false, false}, {"^/untitled", false, false},
		{"/server", true, true}, {`\server`, false, true},
		{"//server/share", true, true}, {`\\server\share`, false, true},
		{"C:", false, false}, {"C:module", false, false},
		{"C:/server", false, true}, {`C:\server`, false, true}, {"1:/server", false, false},
	} {
		want := tc.posix
		if filepath.Separator == '\\' {
			want = tc.windows
		}
		if got := IsAbsolutePath(tc.name); got != want {
			t.Errorf("IsAbsolutePath(%q) = %v, want %v", tc.name, got, want)
		}
	}
}

// Publication only reads metadata; keep filesystem casing independent of the host.
type publicationFS struct {
	vfs.FS
	files         map[string]string
	caseSensitive bool
}

func (fs *publicationFS) UseCaseSensitiveFileNames() bool { return fs.caseSensitive }

func (fs *publicationFS) ReadFile(name string) (string, bool) {
	text, ok := fs.files[tspath.GetCanonicalFileName(name, fs.caseSensitive)]
	return text, ok
}

func TestPublicationCrossPlatformPaths(t *testing.T) {
	for _, directory := range []string{"/Work/Pkg", "C:/Work/Pkg", "//Server/Share/Work/Pkg"} {
		for _, caseSensitive := range []bool{true, false} {
			name := directory + "/insensitive"
			if caseSensitive {
				name = directory + "/sensitive"
			}
			t.Run(name, func(t *testing.T) {
				fs := &publicationFS{caseSensitive: caseSensitive, files: map[string]string{}}
				for relative, text := range map[string]string{
					".npmignore":                 "lib/ignored.js\r\nlib/nested/restored.js\r\nlib/blocked/\r\n",
					"lib/nested/.npmignore":      "ignored.js\r\n!restored.js\r\n",
					"lib/blocked/.npmignore":     "!restored.js\r\n",
					"lib/中文 space[1]/.npmignore": "ignored.js\r\n",
				} {
					fs.files[tspath.GetCanonicalFileName(tspath.ResolvePath(directory, relative), caseSensitive)] = text
				}
				p, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
					Host: utils.CreateCompilerHost(directory, fs), CompilerOptions: &core.CompilerOptions{},
				})
				if err != nil {
					t.Fatal(err)
				}
				pkg := &PackageJSON{directory: directory, data: map[string]any{"files": []any{"lib", "..hidden.js"}, "main": "main.js"}}
				for _, test := range []struct {
					path string
					want bool
				}{
					{"lib/a.js", false}, {`lib\nested\..\a.js`, false},
					{"lib/ignored.js", true}, {"lib/nested/ignored.js", true},
					{"lib/nested/restored.js", false}, {"lib/blocked/restored.js", true},
					{"lib/NESTED/ignored.js", !caseSensitive},
					{"lib/中文 space[1]/ignored.js", true}, {"lib/中文 space[1]/a.js", false},
					{"main.js", false}, {"MAIN.js", caseSensitive}, {"README.js", false}, {"..hidden.js", false},
					{"maİn.js", true},
					{"../Pkg/lib/a.js", false}, {"../Pkg-other/lib/a.js", true},
					{`..\outside.js`, true}, {"lib/../../outside.js", true},
					{"D:/Work/Pkg/lib/a.js", true}, {"//Other/Share/Work/Pkg/lib/a.js", true},
					{"//Server/Other/Work/Pkg/lib/a.js", true},
					{strings.Replace(directory, "C:", "c:", 1) + "/lib/a.js", false},
					{strings.Replace(directory, "Server", "server", 1) + "/lib/a.js", false},
					{strings.Replace(directory, "Share", "share", 1) + "/lib/a.js", caseSensitive && strings.Contains(directory, "/Share/")},
					{strings.Replace(directory, "Pkg", "pkg", 1) + "/lib/a.js", caseSensitive},
					{strings.Replace(directory, "Pkg", "pkg", 1) + "/lib/nested/restored.js", caseSensitive},
					{strings.Replace(directory, "Pkg", "pkg", 1) + "/lib/nested/ignored.js", true},
				} {
					if got := IsUnpublished(p, pkg, tspath.ResolvePath(directory, test.path)); got != test.want {
						t.Errorf("IsUnpublished(%q) = %v, want %v", test.path, got, test.want)
					}
				}
			})
		}
	}
}
