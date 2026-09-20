package packagejson

import (
	"sync"
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func metadataProgram(t testing.TB, files map[string]string) *program.Program {
	t.Helper()
	fs := utils.NewOverlayVFS(iovfs.From(fstest.MapFS{}, true), files)
	p, err := program.NewFromRoots(program.RootOptions{
		Host: utils.CreateCompilerHost("/package-metadata", fs), CompilerOptions: &core.CompilerOptions{NoLib: core.TSTrue},
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPackageGeneration(t *testing.T) {
	create := func(metadata string) *program.Program {
		return metadataProgram(t, map[string]string{
			"/package-metadata/string-bin/package.json":        metadata,
			"/package-metadata/string-bin/nested/package.json": "null",
		})
	}
	first := create(`{"bin":"bin/first.js"}`)
	name := tspath.ResolvePath("/package-metadata", "string-bin/nested/a.js")
	var packages [16]*Package
	var group sync.WaitGroup
	for index := range packages {
		group.Go(func() {
			packages[index] = FindNearestValid(first, name)
		})
	}
	group.Wait()
	pkg := packages[0]
	if pkg == nil || pkg.Field("bin") != "bin/first.js" {
		t.Fatalf("invalid nested metadata did not fall back to parent: %#v", pkg)
	}
	for _, got := range packages {
		if got != pkg {
			t.Error("package decoding was not shared within the Program")
		}
	}
	second := create(`{"bin":"bin/second.js"}`)
	if got := FindNearestValid(second, name); got == nil || got == pkg || got.Field("bin") != "bin/second.js" {
		t.Fatalf("package metadata leaked between Programs: %#v", got)
	}
}

func TestNearestPackagePolicies(t *testing.T) {
	for _, invalid := range []string{"{", "null", "[]", "42", `"text"`} {
		t.Run(invalid, func(t *testing.T) {
			p := metadataProgram(t, map[string]string{
				"/package-metadata/package.json":        `{"name":"parent","engines":{"node":">=20"}}`,
				"/package-metadata/nested/package.json": invalid,
			})
			if FindNearest(p, "/package-metadata/nested/file.js") != nil {
				t.Fatal("invalid nearest metadata changed package ownership")
			}
			pkg := FindNearestValid(p, "/package-metadata/nested/file.js")
			if pkg == nil || pkg.Directory() != "/package-metadata" || pkg.Field("engines", "node") != ">=20" {
				t.Fatalf("parent metadata = %#v", pkg)
			}
			if pkg.Field("name", "child") != nil || pkg.Field("missing") != nil {
				t.Fatal("non-object or absent field resolved")
			}
			if pkg.Text() != `{"name":"parent","engines":{"node":">=20"}}` {
				t.Fatal("original JSON changed")
			}
		})
	}
	p := metadataProgram(t, map[string]string{"/package-metadata/package.json": `{}`})
	if FindNearest(p, "/package-metadata/file.js") != FindNearestValid(p, "/package-metadata/file.js") {
		t.Fatal("lookup policies did not share a decoded package")
	}
	if Read(p, ".") != Read(p, "/package-metadata") {
		t.Fatal("relative and absolute lookup did not share a cache entry")
	}
	if FindNearest(p, "file.js") != Read(p, ".") {
		t.Fatal("relative source filename did not resolve from the Program directory")
	}
	if FindNearest(nil, "/file.js") != nil || FindNearestValid(nil, "/file.js") != nil || Read(nil, "/") != nil {
		t.Fatal("invalid Program returned metadata")
	}
}
