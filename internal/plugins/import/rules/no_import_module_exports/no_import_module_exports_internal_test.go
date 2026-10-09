package no_import_module_exports

import (
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

func TestCompileExceptionMatchersWindowsSeparators(t *testing.T) {
	const (
		pattern  = `C:\repo\src\bridge.js`
		fileName = "C:/repo/src/bridge.js"
	)

	if !compileExceptionMatchers([]string{pattern}, true)[0].Match(fileName) {
		t.Fatal("Windows-native exception did not match the normalized filename")
	}
	if compileExceptionMatchers([]string{pattern}, false)[0].Match(fileName) {
		t.Fatal("POSIX matching unexpectedly treated backslashes as separators")
	}
}

func TestResolveNodePackageEntry(t *testing.T) {
	tests := []struct {
		name        string
		packageJSON string
		files       []string
		want        string
		windows     bool
	}{
		{
			name: "package index skips extensionless file", packageJSON: `{}`,
			files: []string{"index", "index.js"}, want: "index.js",
		},
		{
			name: "explicit main accepts exact extensionless file", packageJSON: `{"main":"entry"}`,
			files: []string{"entry", "entry.js", "index.js"}, want: "entry",
		},
		{
			name: "explicit main probes CommonJS extensions", packageJSON: `{"main":"entry"}`,
			files: []string{"entry.js", "index.js"}, want: "entry.js",
		},
		{
			name: "directory main index skips extensionless file", packageJSON: `{"main":"dist"}`,
			files: []string{"dist/index", "dist/index.js", "index.js"}, want: "dist/index.js",
		},
		{
			name: "directory main ignores nested package metadata", packageJSON: `{"main":"dist"}`,
			files: []string{"dist/package.json", "dist/entry.js", "index.js"}, want: "index.js",
		},
		{
			name: "missing main target falls back to package index", packageJSON: `{"main":"missing"}`,
			files: []string{"index.js"}, want: "index.js",
		},
		{
			name: "package index checks JavaScript before JSON", packageJSON: `{}`,
			files: []string{"index.json", "index.js"}, want: "index.js",
		},
		{
			name: "explicit nonstandard extension is accepted exactly", packageJSON: `{"main":"entry.cjs"}`,
			files: []string{"entry.cjs", "index.js"}, want: "entry.cjs",
		},
		{
			name: "nonstandard extension is not inferred", packageJSON: `{"main":"entry"}`,
			files: []string{"entry.cjs", "index.js"}, want: "index.js",
		},
		{
			name: "dot main resolves the package index", packageJSON: `{"main":"."}`,
			files: []string{"index.js"}, want: "index.js",
		},
		{
			name: "POSIX absolute main remains absolute", packageJSON: `{"main":"/absolute/entry.js"}`,
			files: []string{"/absolute/entry.js", "index.js"}, want: "/absolute/entry.js",
		},
		{
			name: "POSIX backslash does not alias a separator", packageJSON: `{"main":"dist\\entry.js"}`,
			files: []string{"dist/entry.js", "index.js"}, want: "index.js",
		},
		{
			name: "Windows backslash is a separator", packageJSON: `{"main":"dist\\entry.js"}`,
			files: []string{"dist/entry.js", "index.js"}, want: "dist/entry.js", windows: true,
		},
		{
			name: "forward slash resolves on every platform", packageJSON: `{"main":"dist/entry.js"}`,
			files: []string{"dist/entry.js", "index.js"}, want: "dist/entry.js",
		},
		{
			name: "Windows absolute main remains absolute", packageJSON: `{"main":"C:/absolute/entry.js"}`,
			files: []string{"C:/absolute/entry.js", "index.js"}, want: "C:/absolute/entry.js", windows: true,
		},
		{
			name: "extensionless package index alone is not an entry", packageJSON: `{}`,
			files: []string{"index"}, want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := "/node-entry"
			if test.windows {
				root = "C:/node-entry"
			}
			files := map[string]string{tspath.ResolvePath(root, "package.json"): test.packageJSON}
			for _, name := range test.files {
				fileName := name
				if !tspath.IsRootedDiskPath(name) {
					fileName = tspath.ResolvePath(root, name)
				}
				files[fileName] = ""
			}
			fs := utils.NewOverlayVFS(iovfs.From(fstest.MapFS{}, true), files)
			p, err := program.NewFromRoots(program.RootOptions{
				Host: utils.CreateCompilerHost(root, fs), CompilerOptions: program.SourceOnlyCompilerOptions(),
				SingleThreaded: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			pkg := packagejson.FindNearest(p, tspath.ResolvePath(root, "source.js"))
			if pkg == nil {
				t.Fatal("package metadata was not found")
			}
			want := test.want
			if want != "" && !tspath.IsRootedDiskPath(want) {
				want = tspath.ResolvePath(root, want)
			}
			if got := resolveNodePackageEntryForPlatform(p, pkg, test.windows); got != want {
				t.Fatalf("resolveNodePackageEntryForPlatform() = %q, want %q", got, want)
			}
		})
	}
}
