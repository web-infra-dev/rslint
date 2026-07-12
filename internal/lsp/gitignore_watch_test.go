package lsp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestGitignoreFileWatchers(t *testing.T) {
	watchers := gitignoreFileWatchers("/workspace/packages/app", true)
	if len(watchers) != 4 {
		t.Fatalf("watcher count=%d, want recursive plus three ancestors", len(watchers))
	}
	recursive := watchers[0].GlobPattern.RelativePattern
	if recursive == nil || recursive.BaseUri.URI == nil || string(*recursive.BaseUri.URI) != "file:///workspace/packages/app" || recursive.Pattern != "**/.gitignore" {
		t.Fatalf("recursive watcher=%+v", watchers[0])
	}
	wantBases := []string{"file:///workspace/packages", "file:///workspace", "file:///"}
	for i, want := range wantBases {
		relative := watchers[i+1].GlobPattern.RelativePattern
		if relative == nil || relative.BaseUri.URI == nil || string(*relative.BaseUri.URI) != want || relative.Pattern != ".gitignore" {
			t.Fatalf("ancestor watcher %d=%+v, want base %q", i, relative, want)
		}
	}

	withoutRelativePatterns := gitignoreFileWatchers("/workspace/packages/app", false)
	if len(withoutRelativePatterns) != 4 {
		t.Fatalf("watchers without relative-pattern support=%+v", withoutRelativePatterns)
	}
	wantAbsolute := []string{
		"/workspace/packages/app/**/.gitignore",
		"/workspace/packages/.gitignore",
		"/workspace/.gitignore",
		"/.gitignore",
	}
	for i, want := range wantAbsolute {
		pattern := withoutRelativePatterns[i].GlobPattern.Pattern
		if pattern == nil || *pattern != want {
			t.Fatalf("absolute watcher %d=%+v, want %q", i, withoutRelativePatterns[i], want)
		}
	}
}

func TestGitignoreWatchEventsInvalidateDiagnostics(t *testing.T) {
	for _, eventType := range []lsproto.FileChangeType{
		lsproto.FileChangeTypeCreated,
		lsproto.FileChangeTypeChanged,
		lsproto.FileChangeTypeDeleted,
	} {
		t.Run(eventType.String(), func(t *testing.T) {
			s := newTestServer()
			uri := lsproto.DocumentUri("file:///workspace/src/index.ts")
			s.documents[uri] = "debugger;\n"
			s.diagnostics[uri] = []rule.RuleDiagnostic{{RuleName: "no-debugger"}}
			s.docGeneration[uri] = 7

			err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
				Changes: []*lsproto.FileEvent{{
					Uri:  documentURIFromPath(filepath.Join("/workspace", ".gitignore")),
					Type: eventType,
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := s.diagnostics[uri]; ok {
				t.Fatal("stale diagnostics survived .gitignore change")
			}
			if s.docGeneration[uri] != 8 {
				t.Fatalf("generation=%d, want 8", s.docGeneration[uri])
			}
			select {
			case <-s.refreshCh:
			default:
				t.Fatal(".gitignore change did not request diagnostics refresh")
			}
		})
	}
}

func TestLSPGitignoreReloadReadsFreshState(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "source.ts")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignoreURI := documentURIFromPath(gitignorePath)
	uri := documentURIFromPath(target)

	s := newTestServer()
	s.cwd = dir
	s.fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	s.jsonConfig = config.RslintConfig{{Rules: config.Rules{"no-debugger": "error"}}}
	s.documents[uri] = "debugger;\n"
	isIgnored := func() bool {
		effective, cwd, _ := s.getLintConfigForURI(uri)
		return effective.IsFileIgnored(target, cwd)
	}
	if isIgnored() {
		t.Fatal("file was ignored before .gitignore existed")
	}

	if err := os.WriteFile(gitignorePath, []byte("source.ts\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{{Uri: gitignoreURI, Type: lsproto.FileChangeTypeCreated}},
	}); err != nil {
		t.Fatal(err)
	}
	if !isIgnored() {
		t.Fatal("created .gitignore was not applied")
	}

	if err := os.Remove(gitignorePath); err != nil {
		t.Fatal(err)
	}
	if err := s.handleDidChangeWatchedFiles(context.Background(), &lsproto.DidChangeWatchedFilesParams{
		Changes: []*lsproto.FileEvent{{Uri: gitignoreURI, Type: lsproto.FileChangeTypeDeleted}},
	}); err != nil {
		t.Fatal(err)
	}
	if isIgnored() {
		t.Fatal("deleted .gitignore remained active")
	}
}
