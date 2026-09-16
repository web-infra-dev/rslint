package nodeutil

import (
	"encoding/json"
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

func TestResolverEntryFields(t *testing.T) {
	cases := []struct {
		name, request, file, config, want string
		missing                           bool
	}{
		{"default", "./main", "input.js", `{}`, "main/main.js", false},
		{"mainFiles-api", "./plain", "input.js", `{"mainFiles":["api"]}`, "plain/api.js", false},
		{"mainFiles-fallback", "./plain", "input.js", `{"mainFiles":["absent","api","index"]}`, "plain/api.js", false},
		{"mainFiles-empty", "./plain", "input.js", `{"mainFiles":[]}`, "", true},
		{"mainFiles-nested", "./plain", "input.js", `{"mainFiles":["nested/entry"]}`, "plain/nested/entry.js", false},
		{"mainFiles-explicit-extension", "./plain", "input.js", `{"mainFiles":["api.js"]}`, "plain/api.js", false},
		{"mainFiles-./plain/", "./plain/", "input.js", `{"mainFiles":["api"]}`, "plain/api.js", false},
		{"mainFiles-./plain/index", "./plain/index", "input.js", `{"mainFiles":["api"]}`, "plain/index.js", false},
		{"mainFiles-./plain/index.js", "./plain/index.js", "input.js", `{"mainFiles":["api"]}`, "plain/index.js", false},
		{"mainFiles-index", "index", "input.js", `{"mainFiles":["api"]}`, "node_modules/index/api.js", false},
		{"mainFiles-index/index", "index/index", "input.js", `{"mainFiles":["api"]}`, "node_modules/index/index.js", false},
		{"mainFiles-index/index.js", "index/index.js", "input.js", `{"mainFiles":["api"]}`, "node_modules/index/index.js", false},
		{"mainFields-[\"module\",\"main\"]", "./main", "input.js", `{"mainFields":["module","main"]}`, "main/module.js", false},
		{"mainFields-[\"missing\",\"main\"]", "./main", "input.js", `{"mainFields":["missing","main"]}`, "main/main.js", false},
		{"mainFields-[\"invalid\",\"main\"]", "./main", "input.js", `{"mainFields":["invalid","main"]}`, "main/main.js", false},
		{"mainFields-[[\"nested\",\"entry\"]]", "./main", "input.js", `{"mainFields":[["nested","entry"]]}`, "main/nested.js", false},
		{"mainFields-[\"self\",\"selfSlash\",\"main\"]", "./main", "input.js", `{"mainFields":["self","selfSlash","main"]}`, "main/main.js", false},
		{"mainFields-[\"directory\",\"main\"]", "./main", "input.js", `{"mainFields":["directory","main"]}`, "main/lib/deep.js", false},
		{"mainFields-[\"parent\"]", "./main", "input.js", `{"mainFields":["parent"]}`, "src/browser.js", false},
		{"mainFields-[\"absolute\"]", "./main", "input.js", `{"mainFields":["absolute"]}`, "main/index.js", false},
		{"mainFields-[]", "./main", "input.js", `{"mainFields":[]}`, "main/index.js", false},
		{"mainFields-[{\"name\":\"absolute\",\"forceRelative\":false}]", "./main", "input.js", `{"mainFields":[{"name":"absolute","forceRelative":false}]}`, "src/browser.js", false},
		{"mainFields-[{\"name\":[\"nested\",\"entry\"],\"forceRelative\":true}]", "./main", "input.js", `{"mainFields":[{"name":["nested","entry"],"forceRelative":true}]}`, "main/nested.js", false},
		{"mainFields-[\"browser\"]", "./main", "input.js", `{"mainFields":["browser"]}`, "main/browser.js", false},
		{"both-empty", "./main", "input.js", `{"mainFields":[],"mainFiles":[]}`, "", true},
		{"main-before-empty-files", "./main", "input.js", `{"mainFields":["module"],"mainFiles":[]}`, "main/module.js", false},
		{"directory-in-main", "./main", "input.js", `{"mainFields":["directory","module","main"]}`, "main/lib/deep-module.js", false},
		{"main-cycle", "./cycle-a", "input.js", `{"mainFields":["main"]}`, "", true},
		{"exports-before-main", "pkg-exports", "input.js", `{"mainFields":["module"]}`, "node_modules/pkg-exports/node.js", false},
		{"exports-directory", "pkg-exports/dir", "input.js", `{"mainFiles":["api"]}`, "node_modules/pkg-exports/lib/api.js", false},
		{"exports-index-name", "pkg-exports/index", "input.js", `{"mainFields":[],"mainFiles":["module"]}`, "", true},
		{"imports-directory", "#dir", "input.js", `{"mainFiles":["api"]}`, "plain/api.js", false},
		{"aliasFields-bare", "bare", "input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-fs", "fs", "input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-./src/server.js", "./src/server.js", "input.js", `{"aliasFields":["browser"]}`, "src/alternate.js", false},
		{"aliasFields-./src/server", "./src/server", "input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-./src/without-extension", "./src/without-extension", "input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-./src/disabled.js", "./src/disabled.js", "input.js", `{"aliasFields":["browser"]}`, "", false},
		{"aliasFields-./src/absent.js", "./src/absent.js", "input.js", `{"aliasFields":["browser"]}`, "", false},
		{"aliasFields-./src/identical.js", "./src/identical.js", "input.js", `{"aliasFields":["browser"]}`, "src/identical.js", false},
		{"aliasFields-./src/missing-mapping.js", "./src/missing-mapping.js", "input.js", `{"aliasFields":["browser"]}`, "", true},
		{"aliasFields-loop-a", "loop-a", "input.js", `{"aliasFields":["browser"]}`, "", true},
		{"aliasFields-query?raw#source", "query?raw#source", "input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-pkg-entry", "pkg-entry", "input.js", `{"aliasFields":["browser"]}`, "node_modules/pkg-entry/browser.js", false},
		{"aliasFields-pkg-exports", "pkg-exports", "input.js", `{"aliasFields":["browser"]}`, "node_modules/pkg-exports/browser.js", false},
		{"aliasFields-#entry", "#entry", "input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-#dep", "#dep", "input.js", `{"aliasFields":["browser"]}`, "node_modules/pkg-entry/browser.js", false},
		{"aliasFields-pkg-exports/dir", "pkg-exports/dir", "input.js", `{"aliasFields":["browser"]}`, "node_modules/pkg-exports/lib/index.js", false},
		{"aliasFields-disabled", "./src/server.js", "input.js", `{"aliasFields":[]}`, "src/server.js", false},
		{"aliasFields-nested", "bare", "input.js", `{"aliasFields":[["nested","aliases"]]}`, "src/alternate.js", false},
		{"aliasFields-priority", "./src/server.js", "input.js", `{"aliasFields":["alternate","browser"]}`, "src/alternate.js", false},
		{"aliasFields-failed-first", "./src/missing-mapping.js", "input.js", `{"aliasFields":["browser","alternate"]}`, "", true},
		{"aliasFields-subdirectory", "../src/server.js", "sub/input.js", `{"aliasFields":["browser"]}`, "src/alternate.js", false},
		{"aliasFields-package-context", "dep", "node_modules/pkg-entry/sub/input.js", `{"aliasFields":["browser"]}`, "node_modules/pkg-entry/dep-browser.js", false},
		{"aliasFields-and-mainFields", "pkg-entry", "input.js", `{"aliasFields":["browser"],"mainFields":["main"]}`, "node_modules/pkg-entry/browser.js", false},
		{"aliasFields-and-mainFiles", "pkg-exports/dir", "input.js", `{"aliasFields":["browser"],"mainFiles":["api"]}`, "node_modules/pkg-exports/browser.js", false},
		{"aliasFields-string-ignored", "./main", "input.js", `{"aliasFields":["browser"]}`, "main/main.js", false},
		{"mainFields-keeps-extension", "./main", "input.js", `{"mainFields":["main"],"extensionAlias":{".js":[".ts"]}}`, "main/main.js", false},
		{"mainFields-browser-remap-extension", "pkg-entry", "input.js", `{"mainFields":["main"],"aliasFields":["browser"],"extensionAlias":{".js":[".ts"]}}`, "node_modules/pkg-entry/browser.ts", false},
		{"mainFiles-extension-directory", "./extension-directory.js", "input.js", `{"mainFiles":["api"]}`, "extension-directory.js/api.js", false},
		{"mainFields-no-entry-files", "./main", "input.js", `{"mainFields":["missing"],"mainFiles":[]}`, "", true},
		{"mainFiles-absolute-entry", "./plain", "input.js", `{"mainFiles":["@ROOT@/src/browser.js"]}`, "src/browser.js", false},
		{"mainFields-field-name-array", "./main", "input.js", `{"mainFields":[["nested","entry"],"main"]}`, "main/nested.js", false},
		{"aliasFields-shadowed-package", "dep", "nested-owner/input.js", `{"aliasFields":["browser"]}`, "nested-owner/mapped.js", false},
		{"aliasFields-into-nested-package", "./nested-owner/local.js", "input.js", `{"aliasFields":["browser"]}`, "nested-owner/mapped.js", false},
		{"aliasFields-chain", "a", "maps/input.js", `{"aliasFields":["browser"]}`, "maps/mapped.js", false},
		{"aliasFields-file-cycle", "./one.js", "maps/input.js", `{"aliasFields":["browser"]}`, "", true},
		{"aliasFields-directory", "./dir", "maps/input.js", `{"aliasFields":["browser"]}`, "maps/mapped.js", false},
		{"aliasFields-invalid", "invalid", "maps/input.js", `{"aliasFields":["browser","alternate"]}`, "", true},
		{"aliasFields-self", "self", "maps/input.js", `{"aliasFields":["browser"]}`, "", true},
		{"aliasFields-empty", "empty", "maps/input.js", `{"aliasFields":["browser"]}`, "maps/index.js", false},
		{"aliasFields-ignored-missing", "./missing", "maps/input.js", `{"aliasFields":["browser"]}`, "", false},
		{"aliasFields-before-global-file", "bare", "input.js", `{"aliasFields":["browser"],"alias":{"@ROOT@/src/browser.js":"@ROOT@/src/alternate.js"}}`, "src/alternate.js", false},
		{"global-alias-before-field", "bare", "input.js", `{"aliasFields":["browser"],"alias":{"bare":"@ROOT@/src/alternate.js"}}`, "src/alternate.js", false},
		{"mainFiles-backslash", "./plain", "input.js", `{"mainFiles":["nested\\entry"]}`, "", true},
		{"mainFiles-parent", "./plain", "input.js", `{"mainFiles":["../src/browser"]}`, "src/browser.js", false},
		{"mainFiles-empty-name", "./plain", "input.js", `{"mainFiles":["","api"]}`, "plain/api.js", false},
		{"mainFiles-query", "./plain", "input.js", `{"mainFiles":["api?raw"]}`, "", true},
		{"mainFields-query", "./main", "input.js", `{"mainFields":["query"]}`, "main/index.js", false},
		{"mainFields-backslash", "./main", "input.js", `{"mainFields":["backslash"]}`, "main/index.js", false},
		{"mainFields-package-self-dot", "./main", "input.js", `{"mainFields":["self"],"mainFiles":["main"]}`, "main/main.js", false},
		{"mainFields-global-file-alias", "./main", "input.js", `{"mainFields":["module"],"alias":{"@ROOT@/main/module.js":"@ROOT@/src/browser.js"}}`, "src/browser.js", false},
		{"mainFiles-global-file-alias", "./plain", "input.js", `{"mainFiles":["api"],"alias":{"@ROOT@/plain/api.js":"@ROOT@/src/browser.js"}}`, "src/browser.js", false},
		{"aliasFields-backslash-request", ".\\src\\server.js", "input.js", `{"aliasFields":["browser"]}`, "", true},
		{"aliasFields-backslash-subdir", "..\\src\\server.js", "sub/input.js", `{"aliasFields":["browser"]}`, "src/alternate.js", false},
		{"aliasFields-absolute-request", "@ROOT@/src/server.js", "input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-relative-parent", "../src/server.js", "maps/input.js", `{"aliasFields":["browser"]}`, "src/browser.js", false},
		{"aliasFields-backslash-target", "backslash", "maps/input.js", `{"aliasFields":["browser"]}`, "", true},
		{"aliasFields-invalid-owner", "a", "invalid-owner/input.js", `{"aliasFields":["browser"]}`, "", true},
		{"aliasFields-foreign-directory", "./plain", "sub/input.js", `{"aliasFields":["browser"],"mainFiles":["api"]}`, "", true},
	}
	archive := txtarfs.MustParseFile(t, "testdata/resolution_entries.txtar")
	names, err := archive.FileNames("")
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{"/entry-project", "C:/entry-project", "//server/share/entry-project"} {
		t.Run(root, func(t *testing.T) {
			files := map[string]string{}
			for _, name := range names {
				data, err := archive.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				files[tspath.ResolvePath(root, name)] = strings.ReplaceAll(string(data), "@ROOT@", root)
			}
			p := programForResolution(t, files, tspath.ResolvePath(root, "input.js"))
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					var config map[string]any
					if err := json.Unmarshal([]byte(strings.ReplaceAll(tc.config, "@ROOT@", root)), &config); err != nil {
						t.Fatal(err)
					}
					var options ResolutionOptions
					applyResolverConfig(&options, config)
					// Windows entry filenames accept either separator. Requests using
					// backslashes follow Node's path semantics; enhanced-resolve treats
					// .\\ as a package and may resolve ..\\ through node_modules.
					missing := tc.missing
					want := tc.want
					if tspath.GetRootLength(root) > 1 {
						switch tc.name {
						case "mainFiles-backslash":
							want, missing = "plain/nested/entry.js", false
						case "mainFields-backslash":
							want = "main/lib/deep.js"
						case "aliasFields-backslash-request":
							want, missing = "src/alternate.js", false
						case "aliasFields-backslash-target":
							want, missing = "maps/mapped.js", false
						}
					} else if tc.name == "aliasFields-backslash-subdir" {
						want, missing = "", true
					}
					got, resolveError := ResolveModuleWithError(p, strings.ReplaceAll(tc.request, "@ROOT@", root), tspath.ResolvePath(root, tc.file), options)
					if want != "" {
						want = tspath.ResolvePath(root, want)
					}
					if got != want || (resolveError != "") != missing {
						t.Errorf("ResolveModule(%q, %s) = (%q, %q), want (%q, missing=%v)", tc.request, tc.config, got, resolveError, want, missing)
					}
				})
			}
			t.Run("import defaults and overrides", func(t *testing.T) {
				for _, tc := range []struct {
					request string
					config  map[string]any
					want    string
				}{
					{"./plain", nil, ""},
					{"./main", nil, ""},
					{"pkg-entry", nil, "node_modules/pkg-entry/index.js"},
					{"./plain", map[string]any{"mainFiles": []string{"api"}}, "plain/api.js"},
					{"./main", map[string]any{"mainFiles": []string{"index"}}, "main/index.js"},
					{"./main", map[string]any{"mainFields": []string{"module", "main"}}, "main/module.js"},
					{"bare", map[string]any{"aliasFields": []string{"browser"}}, "src/browser.js"},
					{"./plain", map[string]any{"aliasFields": []string{"browser"}}, ""},
				} {
					var options ResolutionOptions
					applyResolverConfig(&options, tc.config)
					result := resolveImport(p, tc.request, tspath.ResolvePath(root, "input.js"), false, options)
					want := tc.want
					if want != "" {
						want = tspath.ResolvePath(root, want)
					}
					if result.path != want || (result.resolveError != "") != (want == "") {
						t.Errorf("import %q (%v) = %+v, want %q", tc.request, tc.config, result, want)
					}
				}
			})
			t.Run("package context and resource suffix", func(t *testing.T) {
				options := ResolutionOptions{Paths: []string{tspath.ResolvePath(root, "search")}, AliasFields: [][]string{{"browser"}}}
				fileName := tspath.ResolvePath(root, "input.js")
				if got := ResolveModule(p, "lookup", fileName, options); got != tspath.ResolvePath(root, "search/lookup.js") {
					t.Errorf("resolvePaths package alias = %q", got)
				}
				result := resolveModuleCached(p, "query?raw#source", fileName, options)
				if result.path != tspath.ResolvePath(root, "src/browser.js") || result.resourceSuffix != "?mapped#source" {
					t.Errorf("package alias resource suffix = %+v", result)
				}
			})
		})
	}
}

func TestResolverEntryFieldSymlinks(t *testing.T) {
	archive := txtarfs.MustParseFile(t, "testdata/resolution_entries.txtar")
	root := tspath.NormalizePath(archive.Materialize(t, ""))
	for _, link := range []struct{ name, target string }{
		{"node_modules/linked", "node_modules/pkg-entry"},
		{"linked-plain", "plain"},
	} {
		if err := os.Symlink(tspath.ResolvePath(root, link.target), tspath.ResolvePath(root, link.name)); err != nil {
			if runtime.GOOS == "windows" {
				t.Skipf("symlink creation is unavailable: %v", err)
			}
			t.Fatal(err)
		}
	}
	fileName := tspath.ResolvePath(root, "input.js")
	p := programForResolution(t, nil, fileName)
	for _, tc := range []struct {
		request string
		config  map[string]any
		want    string
	}{
		{"linked", map[string]any{"mainFields": []string{"main"}, "aliasFields": []string{"browser"}}, "node_modules/pkg-entry/browser.js"},
		{"linked", map[string]any{"mainFields": []string{"main"}, "alias": map[string]any{tspath.ResolvePath(root, "node_modules/linked/index.js"): tspath.ResolvePath(root, "src/alternate.js")}}, "src/alternate.js"},
		{"./linked-plain", map[string]any{"mainFiles": []string{"api"}}, "plain/api.js"},
	} {
		var options ResolutionOptions
		applyResolverConfig(&options, tc.config)
		want := osvfs.FS().Realpath(tspath.ResolvePath(root, tc.want))
		if got, err := ResolveModuleWithError(p, tc.request, fileName, options); got != want || err != "" {
			t.Errorf("symlink %q (%v) = %q, %s; want %q", tc.request, tc.config, got, err, want)
		}
	}
	if !osvfs.FS().UseCaseSensitiveFileNames() {
		options := ResolutionOptions{MainFields: []nodeMainField{{Name: []string{"module"}, ForceRelative: true}}}
		want := osvfs.FS().Realpath(tspath.ResolvePath(root, "main/module.js"))
		if got := ResolveModule(p, "./MAIN", fileName, options); tspath.GetCanonicalFileName(got, false) != tspath.GetCanonicalFileName(want, false) {
			t.Errorf("case-insensitive package entry = %q, want %q", got, want)
		}
	}
}
