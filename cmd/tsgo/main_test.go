package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestExternalFileNamespace(t *testing.T) {
	fixture := txtarfs.MustParseFile(t, "testdata/external_namespace.txtar")
	for _, test := range []struct {
		name        string
		packageJSON string
		want        string
	}{
		{name: "without package.json"},
		{name: "named package", packageJSON: `{"name":"named-dependency"}`, want: "named-dependency"},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := fixture.Materialize(t, "")
			if test.packageJSON != "" {
				if err := os.WriteFile(filepath.Join(directory, "node_modules/pkg/package.json"), []byte(test.packageJSON), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			fs := bundled.WrapFS(iovfs.From(os.DirFS(directory), tspath.CaseSensitive))
			program, err := utils.CreateProgram(true, fs, "/", "tsconfig.json", utils.CreateCompilerHost("/", fs))
			if err != nil {
				t.Fatal(err)
			}
			file := program.GetSourceFile(tspath.RootedFilePathFromAbsolute("/node_modules/pkg/index.d.ts"))
			if file == nil || !program.IsSourceFileFromExternalLibrary(file) {
				t.Fatal("dependency was not loaded as an external source file")
			}
			want := test.want
			if want == "" {
				want = file.FileName().AsString()
			}
			if got := externalFileNamespace(program, file); got != want {
				t.Errorf("externalFileNamespace = %q, want %q", got, want)
			}
		})
	}
}
