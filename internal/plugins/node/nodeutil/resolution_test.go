package nodeutil

import (
	"maps"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/program"
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
		if got := ResolveModule(p, "linked", fileName, ResolutionOptions{Modules: []string{"node_modules/custom"}}); got != want {
			t.Errorf("symlinked directory export = %q, want %q", got, want)
		}
	})
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
