package target

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type sourceMappingFS struct {
	vfs.FS
	realPaths map[string]string
	observed  []string
}

func (fs *sourceMappingFS) Realpath(path string) string {
	fs.observed = append(fs.observed, path)
	return fs.realPaths[path]
}

func TestLookupSourceTargetPrefersExactThenCanonicalIdentity(t *testing.T) {
	lexicalTarget := File{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/link.ts"}}
	canonicalTarget := File{PathIdentity: rslintconfig.PathIdentity{Path: "/repo/real.ts"}}
	targets := map[string]File{
		rslintconfig.ExactPathID("/repo/link.ts"):  lexicalTarget,
		rslintconfig.ExactPathID("/physical/a.ts"): canonicalTarget,
	}
	fsys := &sourceMappingFS{
		FS:        osvfs.FS(),
		realPaths: map[string]string{"/repo/alias.ts": "/physical/a.ts"},
	}

	if got, ok := LookupSourceTarget(targets, "/repo/link.ts", fsys); !ok || got.Path != lexicalTarget.Path {
		t.Fatalf("exact lookup = (%+v, %v), want lexical target", got, ok)
	}
	if got, ok := LookupSourceTarget(targets, "/repo/alias.ts", fsys); !ok || got.Path != canonicalTarget.Path {
		t.Fatalf("canonical lookup = (%+v, %v), want physical target", got, ok)
	}
}

func TestLookupSourceTargetNormalizesCrossPlatformPathBeforeRealpath(t *testing.T) {
	aliases := []string{
		`C:\repo\alias.ts`,
		`\\server\share\alias.ts`,
	}
	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			normalizedAlias := tspath.NormalizePath(alias)
			canonicalPath := tspath.NormalizePath(alias + `.physical`)
			fsys := &sourceMappingFS{
				FS:        osvfs.FS(),
				realPaths: map[string]string{normalizedAlias: canonicalPath},
			}
			want := File{PathIdentity: rslintconfig.PathIdentity{Path: normalizedAlias}}
			targets := map[string]File{rslintconfig.ExactPathID(canonicalPath): want}

			got, ok := LookupSourceTarget(targets, alias, fsys)
			if !ok || got.Path != want.Path {
				t.Fatalf("lookup = (%+v, %v), want canonical target %+v", got, ok, want)
			}
			if len(fsys.observed) != 1 || fsys.observed[0] != normalizedAlias {
				t.Fatalf("Realpath inputs = %q, want normalized %q", fsys.observed, normalizedAlias)
			}
		})
	}
}
