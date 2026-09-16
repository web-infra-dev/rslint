package nodeutil

import (
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func programForResolution(t *testing.T, files map[string]string, fileName string) *program.Program {
	t.Helper()
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), maps.Clone(files))
	host := utils.CreateCompilerHost(tspath.GetDirectoryPath(fileName), fs)
	p, err := program.NewFromRoots(program.RootOptions{
		Host: host, CompilerOptions: &core.CompilerOptions{AllowJs: core.TSTrue},
		RootFileNames: []string{fileName}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestResolveModuleGenerationAndOptions(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/node_resolution.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, name := range names {
		data, err := archive.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		files[tspath.ResolvePath("/node-runtime", name)] = string(data)
	}
	create := func() *program.Program {
		return programForResolution(t, files, "/node-runtime/input.js")
	}
	p := create()
	for _, test := range []struct {
		name    string
		options ResolutionOptions
		want    string
	}{
		{"pkg", ResolutionOptions{}, "/node-runtime/node_modules/pkg/index.js"},
		{"pkg", ResolutionOptions{Modules: []string{}}, ""},
		{"pkg", ResolutionOptions{Modules: []string{"/node-runtime/vendor"}}, "/node-runtime/vendor/pkg/index.js"},
		// Module directories may themselves contain a node_modules component.
		{"pkg", ResolutionOptions{Modules: []string{"node_modules/custom"}}, "/node-runtime/node_modules/custom/pkg/index.js"},
		{"pkg", ResolutionOptions{Modules: []string{"/node-runtime/node_modules/custom"}}, "/node-runtime/node_modules/custom/pkg/index.js"},
		{"directory", ResolutionOptions{Modules: []string{"node_modules/custom"}}, "/node-runtime/node_modules/custom/directory/lib/index.js"},
		{"directory", ResolutionOptions{Modules: []string{"/node-runtime/node_modules/custom"}}, "/node-runtime/node_modules/custom/directory/lib/index.js"},
		{"directory", ResolutionOptions{Modules: []string{"vendor/node_modules"}}, "/node-runtime/vendor/node_modules/directory/lib/index.js"},
		{"directory", ResolutionOptions{Modules: []string{"node_modules/custom/node_modules"}}, "/node-runtime/node_modules/custom/node_modules/directory/lib/index.js"},
		{"only-types", ResolutionOptions{}, ""},
		{"custom", ResolutionOptions{}, ""},
		{"custom", ResolutionOptions{Extensions: []string{".custom"}}, "/node-runtime/node_modules/custom/index.custom"},
		{"conditional", ResolutionOptions{}, "/node-runtime/node_modules/conditional/import.js"},
		{"conditional", ResolutionOptions{Conditions: []string{"types"}}, "/node-runtime/node_modules/conditional/index.d.ts"},
	} {
		if got := ResolveModule(p, test.name, "/node-runtime/input.js", test.options); got != test.want {
			t.Errorf("ResolveModule(%s, %#v) = %q, want %q", test.name, test.options, got, test.want)
		}
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			if got := ResolveModule(p, "pkg", "/node-runtime/input.js", ResolutionOptions{}); got != "/node-runtime/node_modules/pkg/index.js" {
				t.Errorf("concurrent cached resolution = %q", got)
			}
		})
	}
	group.Wait()
	files["/node-runtime/node_modules/new-package/index.js"] = "export default 1;"
	if got := ResolveModule(p, "new-package", "/node-runtime/input.js", ResolutionOptions{}); got != "" {
		t.Errorf("previous generation saw a new file: %q", got)
	}
	if got := ResolveModule(create(), "new-package", "/node-runtime/input.js", ResolutionOptions{}); got != "/node-runtime/node_modules/new-package/index.js" {
		t.Errorf("new generation did not resolve the new file: %q", got)
	}
	if got := ResolveModule(nil, "pkg", "/input.js", ResolutionOptions{}); got != "" {
		t.Errorf("invalid Program resolved %q", got)
	}
	t.Run("symlinked directory export", func(t *testing.T) {
		root := tspath.NormalizePath(archive.Materialize(t, ""))
		target := tspath.ResolvePath(root, "node_modules/custom/directory")
		if err := os.Symlink(target, tspath.ResolvePath(root, "node_modules/custom/linked")); err != nil {
			if runtime.GOOS == "windows" {
				t.Skipf("symlink creation is unavailable: %v", err)
			}
			t.Fatal(err)
		}
		fileName := tspath.ResolvePath(root, "input.ts")
		p := programForResolution(t, map[string]string{fileName: "import 'linked';"}, fileName)
		want := osvfs.FS().Realpath(tspath.ResolvePath(target, "lib/index.js"))

		options := ResolutionOptions{Modules: []string{"node_modules/custom"}, Aliases: []moduleAlias{{Name: "virtual", Targets: []string{"linked"}}}}
		if got := ResolveModule(p, "virtual", fileName, options); got != want {
			t.Errorf("alias symlink = %q, want %q", got, want)
		}
		replacement := tspath.ResolvePath(root, "node_modules/pkg/index.js")
		options.Aliases = []moduleAlias{{Name: tspath.ResolvePath(root, "node_modules/custom/linked/lib/index.js"), Targets: []string{replacement}}}
		if got := ResolveModule(p, "linked", fileName, options); got != osvfs.FS().Realpath(replacement) {
			t.Errorf("alias before symlink resolution = %q", got)
		}

		if got := ResolveModule(p, "linked", fileName, ResolutionOptions{Modules: []string{"node_modules/custom"}}); got != want {
			t.Errorf("symlinked directory export = %q, want %q", got, want)
		}
	})
}

func TestImportResolutionProcessDirectory(t *testing.T) {
	fileName := "/process/project/input.ts"
	p := programForResolution(t, map[string]string{
		fileName:                                         "export {};",
		"/process/project/tsconfig.json":                 `{"compilerOptions":{}}`,
		"/process/project/view.tsx":                      "export {};",
		"/process/extensions.json":                       `{"compilerOptions":{"jsx":"react"}}`,
		"/process/deps/node_modules/pkg/index.js":        "export {};",
		"/process/vendor/deps/node_modules/pkg/index.js": "export {};",
	}, fileName)
	options := map[string]any{"resolvePaths": []any{"deps"}, "tsconfigPath": "./extensions.json"}
	for _, cwd := range []string{"", "vendor"} {
		t.Run(cwd, func(t *testing.T) {
			ctx := (rule.RuleContext{SourceFile: p.SourceFiles()[0], Settings: map[string]any{}}).
				WithProgram(p).WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory("/process"))
			if cwd != "" {
				ctx.Settings["cwd"] = cwd
			}
			for _, resolution := range []ResolutionOptions{ImportResolutionOptions(ctx, false, options), RequireResolutionOptions(ctx, options)} {
				want := tspath.ResolvePath("/process", cwd, "deps/node_modules/pkg/index.js")
				if got := ResolveModule(p, "pkg", fileName, resolution); got != want {
					t.Errorf("process-relative lookup = %q, want %q", got, want)
				}
				// tsconfigPath uses process.cwd even when settings.cwd redirects
				// module lookup. The selected JSX map resolves view.js to TSX.
				if got := ImportResolveError(p, "./view.js", fileName, false, resolution); got != "" {
					t.Errorf("process-relative tsconfigPath: %s", got)
				}
			}
		})
	}
}

func TestNearestCompilerOptionsGeneration(t *testing.T) {
	files := map[string]string{
		"/nearest-config/input.ts":      "import 'pkg';",
		"/nearest-config/base.json":     `{"compilerOptions":{"allowImportingTsExtensions":true,"paths":{"alias/*":["src/*"]}}}`,
		"/nearest-config/tsconfig.json": `{"extends":"./base.json","files":["input.ts"]}`,
	}
	p := programForResolution(t, files, "/nearest-config/input.ts")
	options := nearestCompilerOptions(p, "/nearest-config/deep/input.ts")
	if options == nil || options.AllowImportingTsExtensions != core.TSTrue || options.Paths == nil || options.Paths.Size() != 1 {
		t.Fatalf("inherited options were not parsed: %#v", options)
	}
	if nearestCompilerOptions(p, "/nearest-config/deep/input.ts") != options {
		t.Error("nearest config was not cached")
	}
	files["/nearest-config/tsconfig.json"] = `{"compilerOptions":{"allowImportingTsExtensions":false},"files":["input.ts"]}`
	updated := programForResolution(t, files, "/nearest-config/input.ts")
	if got := nearestCompilerOptions(updated, "/nearest-config/deep/input.ts"); got == nil || got.AllowImportingTsExtensions != core.TSFalse {
		t.Fatalf("config leaked between generations: %#v", got)
	}
	if nearestCompilerOptions(nil, "/input.ts") != nil {
		t.Fatal("invalid Program returned options")
	}
}

// Node's host path.resolve removes trailing separators and preserves the
// importer's drive/UNC share. A nil Program forces the lexical fallback.
func TestImportFilePathLexicalFallback(t *testing.T) {
	fileName := "/project/client/input.js"
	if runtime.GOOS == "windows" {
		fileName = "C:/project/client/input.js"
	}
	for _, tc := range []struct {
		name, posix, windows string
	}{
		{".", "/project/client", `C:\project\client`},
		{"../server/missing.js", "/project/server/missing.js", `C:\project\server\missing.js`},
		{"../server/", "/project/server", `C:\project\server`},
		{"../missing/../", "/project", `C:\project`},
		{"/server//missing/../", "/server", `C:\server`},
		{`\server\missing\..`, `/project/client/\server\missing\..`, `C:\server`},
		{"//server/share/../../missing", "/missing", `\\server\share\missing`},
		{"./missing?raw#part", "/project/client/missing?raw#part", `C:\project\client\missing?raw#part`},
		{"pkg", "", ""}, {"fs", "", ""},
		{"https://example.com/mod.js?raw#part", "https://example.com/mod.js?raw#part", "https://example.com/mod.js?raw#part"},
	} {
		want := tc.posix
		if runtime.GOOS == "windows" {
			want = tc.windows
		}
		if got := ImportFilePath(nil, tc.name, fileName, false, ResolutionOptions{}); got != want {
			t.Errorf("ImportFilePath(%q) = %q, want %q", tc.name, got, want)
		}
	}
	if runtime.GOOS == "windows" {
		got := ImportFilePath(nil, "../../missing", "//server/share/client/input.js", false, ResolutionOptions{})
		if got != `\\server\share\missing` {
			t.Errorf("UNC fallback = %q", got)
		}
	}
}

func TestResolverOverrides(t *testing.T) {
	for _, root := range []string{"/alias-project", "C:/alias-project", "//server/share/alias-project"} {
		t.Run(root, func(t *testing.T) {
			archive := txtarfs.MustParseFile(t, "testdata/node_resolution.txtar")
			names, err := archive.FileNames("")
			if err != nil {
				t.Fatal(err)
			}
			files := map[string]string{}
			for _, name := range names {
				data, err := archive.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				files[tspath.ResolvePath(root, name)] = string(data)
			}
			fileName := tspath.ResolvePath(root, "input.js")
			p := programForResolution(t, files, fileName)
			local := tspath.ResolvePath(root, "vendor/pkg/index.js")
			config := func(alias any) map[string]any { return map[string]any{"alias": alias} }
			if root != "/alias-project" {
				var options ResolutionOptions
				applyResolverConfig(&options, config(map[string]any{"native": strings.ReplaceAll(local, "/", `\`)}))
				if got := ResolveModule(p, "native", fileName, options); got != local {
					t.Errorf("native path = %q, want %q", got, local)
				}
			}
			options := ResolutionOptions{Aliases: []moduleAlias{{Name: tspath.ResolvePath(root, "node_modules/pkg/index.js"), Targets: []string{local + "?from-alias"}}}}
			if got := RequireFilePath(p, "pkg", fileName, options); got != filepath.FromSlash(local)+"?from-alias" {
				t.Errorf("file alias query = %q", got)
			}

			for _, query := range []struct{ request, target, suffix string }{
				{"virtual?raw#part", local, "?raw#part"},
				{"virtual", local + "?custom#part", "?custom#part"},
				{"virtual?raw", local + "?custom", "?custom"},
				{"virtual?raw#source", local + "?custom", "?custom#source"},
				{"virtual?raw#source", local + "#custom", "?raw#custom"},
				{"virtual?raw#source", local + "?", "?#source"},
				{"virtual?raw#source", local + "#", "?raw#"},
			} {
				var options ResolutionOptions
				applyResolverConfig(&options, config(map[string]any{"virtual": query.target}))
				if got := RequireFilePath(p, query.request, fileName, options); got != filepath.FromSlash(local)+query.suffix {
					t.Errorf("alias query = %q", got)
				}
				if got := ResolveModule(p, query.request, fileName, options); got != local {
					t.Errorf("filesystem path includes a query: %q", got)
				}
			}

			for _, test := range []struct {
				name   string
				config map[string]any
				want   string
			}{
				{"virtual", config(map[string]any{"virtual": local}), local},
				{"pkg", config(map[string]any{"pkg": local}), local},
				{"fs", config(map[string]any{"fs": local}), local},
				{"pkg", config(map[string]any{"pkg": false}), ""},
				{"pkg", config(map[string]any{"pkg": []any{}}), tspath.ResolvePath(root, "node_modules/pkg/index.js")},
				{"pkg", config(map[string]any{"pkg": "pkg"}), tspath.ResolvePath(root, "node_modules/pkg/index.js")},
				{"virtual", config(map[string]any{"virtual": []any{"./absent", local}}), local},
				{"virtual/index", config(map[string]any{"virtual": tspath.ResolvePath(root, "vendor/pkg")}), local},
				{"virtual/index", config(map[string]any{"virtual$": tspath.ResolvePath(root, "vendor/pkg")}), ""},
				{"virtual", config(map[string]any{"virtual$": local}), local},
				{"virtual/pkg", config(map[string]any{"virtual/*": tspath.ResolvePath(root, "vendor/*/index.js")}), local},
				{"virtual", config([]any{map[string]any{"name": "virtual", "alias": local, "onlyModule": true}}), local},
				{"a", config(map[string]any{"a": "b", "b": local}), local},
				{"a", config(map[string]any{"a": "b", "b": "a"}), ""},
				{"a", config(map[string]any{"*": "prefix*"}), ""},
				{"./node_modules/pkg/index.js", config(map[string]any{tspath.ResolvePath(root, "node_modules/pkg"): tspath.ResolvePath(root, "vendor/pkg")}), local},
				{"virtual", config(map[string]any{"virtual": "./vendor/pkg/index.js"}), local},
				{"custom", map[string]any{"extensions": []any{".custom"}}, tspath.ResolvePath(root, "node_modules/custom/index.custom")},
				{"pkg", map[string]any{"extensions": []any{}}, ""},
				{"pkg", map[string]any{"modules": []string{"node_modules"}}, tspath.ResolvePath(root, "node_modules/pkg/index.js")},
				{"conditional", map[string]any{"conditionNames": []any{"import"}}, tspath.ResolvePath(root, "node_modules/conditional/import.js")},
				{"conditional", map[string]any{"conditionNames": []any{"types"}}, tspath.ResolvePath(root, "node_modules/conditional/index.d.ts")},
				{"conditional", map[string]any{"conditionNames": []any{}}, ""},
			} {
				t.Run(test.name, func(t *testing.T) {
					ctx := (rule.RuleContext{SourceFile: p.SourceFiles()[0], Settings: map[string]any{"node": map[string]any{"resolverConfig": test.config}}}).WithProgram(p)
					options := RequireResolutionOptions(ctx, nil)
					if got := ResolveModule(p, test.name, fileName, options); got != test.want {
						t.Errorf("ResolveModule(%q, %#v) = %q, want %q", test.name, test.config, got, test.want)
					}
				})
			}
		})
	}
}
