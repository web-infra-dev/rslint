package tsconfig

import (
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func configProgram(t testing.TB, files map[string]string, fileName string) *program.Program {
	t.Helper()
	fs := utils.NewOverlayVFS(iovfs.From(fstest.MapFS{}, true), files)
	p, err := program.NewFromRoots(program.RootOptions{Host: utils.CreateCompilerHost(tspath.GetDirectoryPath(fileName), fs), CompilerOptions: &core.CompilerOptions{NoLib: core.TSTrue}})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNearestCompilerOptionsGeneration(t *testing.T) {
	files := map[string]string{
		"/nearest-config/input.ts":      "import 'pkg';",
		"/nearest-config/base.json":     `{"compilerOptions":{"allowImportingTsExtensions":true,"paths":{"alias/*":["src/*"]}}}`,
		"/nearest-config/tsconfig.json": `{"extends":"./base.json","files":["input.ts"]}`,
	}
	p := configProgram(t, files, "/nearest-config/input.ts")
	options := FindNearest(p, "/nearest-config/deep/input.ts")
	if options == nil || options.AllowImportingTsExtensions != core.TSTrue || options.Paths == nil || options.Paths.Size() != 1 {
		t.Fatalf("inherited options were not parsed: %#v", options)
	}
	if FindNearest(p, "/nearest-config/deep/input.ts") != options {
		t.Error("nearest config was not cached")
	}
	files["/nearest-config/tsconfig.json"] = `{"compilerOptions":{"allowImportingTsExtensions":false},"files":["input.ts"]}`
	updated := configProgram(t, files, "/nearest-config/input.ts")
	if got := FindNearest(updated, "/nearest-config/deep/input.ts"); got == nil || got.AllowImportingTsExtensions != core.TSFalse {
		t.Fatalf("config leaked between generations: %#v", got)
	}
	if FindNearest(nil, "/input.ts") != nil {
		t.Fatal("invalid Program returned options")
	}
}

func TestExplicitAndNearestConfig(t *testing.T) {
	p := configProgram(t, map[string]string{
		"/config-query/tsconfig.json":        `{"compilerOptions":{"jsx":"react"}}`,
		"/config-query/base.json":            `{"compilerOptions":{"jsx":"preserve"}}`,
		"/config-query/nested/tsconfig.json": `{"extends":"../base.json"}`,
	}, "/config-query/input.ts")
	nearest := FindNearest(p, "/config-query/nested/input.ts")
	if nearest == nil || nearest.Jsx != core.JsxEmitPreserve {
		t.Fatal("nearest config did not honor extends")
	}
	if FindNearest(p, "nested/input.ts") != nearest {
		t.Fatal("relative source filename did not resolve from the Program directory")
	}
	if Read(p, "nested/tsconfig.json") != nearest {
		t.Fatal("explicit and nearest queries did not share parsing")
	}
	if root := Read(p, "tsconfig.json"); root == nil || root.Jsx != core.JsxEmitReact {
		t.Fatal("explicit query used the wrong config")
	}
	if Read(nil, "/tsconfig.json") != nil {
		t.Fatal("invalid Program returned a config")
	}
}
