package config

import (
	"strings"
	"testing"
)

func TestPathSpaceSnapshotRequiresEveryAuthoredBase(t *testing.T) {
	configDirectory := "/repo/config"
	config := ConfigWithAuthoredPathBase(
		RslintConfig{{Files: []string{"src/*.ts"}}},
		"/repo/invocation",
	)

	if _, err := NewTargetMatcherWithPathSpaces(config, configDirectory, nil, nil); err == nil ||
		!strings.Contains(err.Error(), "path-space snapshot is required") {
		t.Fatalf("nil snapshot error = %v", err)
	}

	emptySnapshot := NewPathSpaceSnapshot(nil, nil)
	if _, err := NewTargetMatcherWithPathSpaces(nil, configDirectory, nil, emptySnapshot); err == nil ||
		!strings.Contains(err.Error(), configDirectory) {
		t.Fatalf("missing owner base error = %v", err)
	}

	ownerOnlySnapshot := NewPathSpaceSnapshot(
		map[string]RslintConfig{configDirectory: nil},
		nil,
	)
	if _, err := NewTargetMatcherWithPathSpaces(config, configDirectory, nil, ownerOnlySnapshot); err == nil ||
		!strings.Contains(err.Error(), "/repo/invocation") {
		t.Fatalf("missing authored entry base error = %v", err)
	}

	completeSnapshot := NewPathSpaceSnapshot(
		map[string]RslintConfig{configDirectory: config},
		nil,
	)
	if _, err := NewTargetMatcherWithPathSpaces(config, configDirectory, nil, completeSnapshot); err != nil {
		t.Fatalf("complete snapshot rejected: %v", err)
	}
}

func TestPathSpaceSnapshotAllowsEmptyConfigAtFrozenOwner(t *testing.T) {
	const configDirectory = "/repo"
	snapshot := NewPathSpaceSnapshot(
		map[string]RslintConfig{configDirectory: nil},
		nil,
	)
	matcher, err := NewTargetMatcherWithPathSpaces(nil, configDirectory, nil, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if decision := matcher.MatchFile(PathIdentity{Path: "/repo/index.ts"}); !decision.Selected || decision.GloballyIgnored {
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
