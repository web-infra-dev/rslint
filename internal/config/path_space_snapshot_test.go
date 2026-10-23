package config

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

func TestExactPathIDWindowsDrive(t *testing.T) {
	for _, test := range []struct {
		left, right string
		equal       bool
	}{
		{"", "", true},
		{".", "./", true},
		{"src/../file.ts", "file.ts", true},
		{`C:\Repo\src\a.ts`, "c:/Repo/src/a.ts", true},
		{"./C:/Repo/src/../a.ts", "c:/Repo/a.ts", true},
		{"C:/Repo/a.ts", "D:/Repo/a.ts", false},
		{"C:/Repo/a.ts", "c:/repo/a.ts", false},
		{"C:/Repo/A.ts", "c:/Repo/a.ts", false},
		{"A:leaf", "a:leaf", false},
		{"C:", "c:", false},
		{"//Server/share/a.ts", "//server/share/a.ts", false},
		{`\\Server\share\src\..\a.ts`, "//Server/share/a.ts", true},
		{"//Server/Share/a.ts", "//Server/share/a.ts", false},
		{"/Repo/src/../a.ts", "/Repo/a.ts", true},
		{"/Repo/link.ts", "/Physical/a.ts", false},
		{"/Repo/a.ts", "/repo/a.ts", false},
	} {
		t.Run(test.left+" vs "+test.right, func(t *testing.T) {
			if equal := ExactPathID(test.left) == ExactPathID(test.right); equal != test.equal {
				t.Fatalf("identity equality = %v, want %v", equal, test.equal)
			}
		})
	}
}

func TestPathSpaceSnapshotExtensionRetainsFrozenAnchors(t *testing.T) {
	fsys := &pathIdentityHookFS{FS: &pathSpaceTestFS{}}
	fsys.realpath = func(path string) string {
		if strings.HasPrefix(path, "/alias") {
			t.Fatalf("re-resolved a frozen lexical alias: %s", path)
		}
		if path == "/outside" {
			return "/independent/outside"
		}
		return path
	}
	original := NewPathSpaceSnapshot(nil, fsys, DirectoryIdentity{LexicalPath: "/alias", CanonicalPath: "/physical"})
	inside, outside := "src", "../outside"
	entries := RslintConfig{{BasePath: &inside}, {BasePath: &outside}}
	extended := original.WithConfigBases(map[string]RslintConfig{"/alias": entries}, fsys)
	for lexical, physical := range map[string]string{
		"/alias": "/physical", "/alias/src": "/physical/src", "/outside": "/independent/outside",
	} {
		if got, exists := extended.PhysicalDirectory(lexical); !exists || got != physical {
			t.Fatalf("base %s = %s, %t; want %s", lexical, got, exists, physical)
		}
	}
	if _, exists := original.PhysicalDirectory("/alias/src"); exists {
		t.Fatal("extending a snapshot changed the original")
	}
	owner := tspath.CombinePaths("/alias", "pkg")
	nested := extended.WithConfigBases(map[string]RslintConfig{owner: nil}, fsys)
	if got, _ := nested.PhysicalDirectory(owner); got != "/physical/pkg" {
		t.Fatalf("nested owner escaped its original anchor: %s", got)
	}
}

func TestPathSpaceSnapshotRequiresEveryAuthoredBase(t *testing.T) {
	configDirectory := "/repo/config"
	config := ConfigWithAuthoredPathBase(
		RslintConfig{{Files: []string{"src/*.ts"}}},
		"/repo/invocation",
	)

	if _, err := NewTargetMatcherWithPathSpaces(config, configDirectory, nil, nil, nil); err == nil ||
		!strings.Contains(err.Error(), "path-space snapshot is required") {
		t.Fatalf("nil snapshot error = %v", err)
	}

	emptySnapshot := NewPathSpaceSnapshot(nil, nil)
	if _, err := NewTargetMatcherWithPathSpaces(nil, configDirectory, nil, emptySnapshot, nil); err == nil ||
		!strings.Contains(err.Error(), configDirectory) {
		t.Fatalf("missing owner base error = %v", err)
	}

	ownerOnlySnapshot := NewPathSpaceSnapshot(
		map[string]RslintConfig{configDirectory: nil},
		nil,
	)
	if _, err := NewTargetMatcherWithPathSpaces(config, configDirectory, nil, ownerOnlySnapshot, nil); err == nil ||
		!strings.Contains(err.Error(), "/repo/invocation") {
		t.Fatalf("missing authored entry base error = %v", err)
	}

	completeSnapshot := NewPathSpaceSnapshot(
		map[string]RslintConfig{configDirectory: config},
		nil,
	)
	if _, err := NewTargetMatcherWithPathSpaces(config, configDirectory, nil, completeSnapshot, nil); err != nil {
		t.Fatalf("complete snapshot rejected: %v", err)
	}
}

func TestPathSpaceSnapshotAllowsEmptyConfigAtFrozenOwner(t *testing.T) {
	const configDirectory = "/repo"
	snapshot := NewPathSpaceSnapshot(
		map[string]RslintConfig{configDirectory: nil},
		nil,
	)
	matcher, err := NewTargetMatcherWithPathSpaces(nil, configDirectory, nil, snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	if decision := matcher.MatchFile(PathIdentity{Path: "/repo/index.ts"}); !decision.Selected() || decision.GloballyIgnored() {
		t.Fatalf("empty config default-baseline match = %+v", decision)
	}
}

func TestPathSpaceSnapshotResolvePathRejectsMissingBase(t *testing.T) {
	snapshot := NewPathSpaceSnapshot(nil, nil)
	if _, _, err := snapshot.ResolvePath(
		"/repo/index.ts",
		"/repo",
		nil,
	); err == nil || !strings.Contains(err.Error(), "/repo") {
		t.Fatalf("ResolvePath missing-base error = %v", err)
	}
}
