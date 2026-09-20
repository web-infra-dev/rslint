package utils

import (
	"testing"
	"testing/fstest"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/iovfs"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func TestJestVersionPackageLookup(t *testing.T) {
	for _, test := range []struct{ metadata, want string }{
		{`{"dependencies":{"jest":"^29"},"devDependencies":{"jest":"^30"}}`, "^29"},
		{`{"peerDependencies":{"jest":"^28"}}`, "^28"},
		{`{"dependencies":{"jest":29}}`, ""},
		{`{"dependencies":{"jest":"^29"},"devDependencies":{"other":42}}`, ""},
		{`{}`, ""}, {`null`, ""}, {`["jest"]`, ""}, {`{`, ""},
	} {
		t.Run(test.metadata, func(t *testing.T) {
			fs := internalUtils.NewOverlayVFS(iovfs.From(fstest.MapFS{}, true), map[string]string{
				"/jest-metadata/nested/file.js":      "jest.resetModuleRegistry();",
				"/jest-metadata/package.json":        `{"dependencies":{"jest":"^20"}}`,
				"/jest-metadata/nested/package.json": test.metadata,
			})
			p, err := program.NewFromRoots(program.RootOptions{
				Host: internalUtils.CreateCompilerHost("/jest-metadata", fs), CompilerOptions: &core.CompilerOptions{NoLib: core.TSTrue, AllowJs: core.TSTrue}, RootFileNames: []string{"/jest-metadata/nested/file.js"},
			})
			if err != nil {
				t.Fatal(err)
			}
			file := p.GetSourceFile("/jest-metadata/nested/file.js")
			ctx := (rule.RuleContext{SourceFile: file}).WithProgram(p)
			if got := readJestVersionFromPackageJson(ctx); got != test.want {
				t.Fatalf("Jest version = %q, want %q", got, test.want)
			}
		})
	}
}
