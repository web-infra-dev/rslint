//go:build !js

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// newCLIChannelPair wires two ipc.Channels back-to-back over two io.Pipes so
// a test can play the Node peer (b) opposite the CLI side (a). Channels are
// returned unstarted; the caller installs handlers then Start()s.
func newCLIChannelPair(t *testing.T) (a, b *ipc.Channel) {
	t.Helper()
	abR, abW := io.Pipe() // a → b
	baR, baW := io.Pipe() // b → a
	a = ipc.NewChannel(baR, abW)
	b = ipc.NewChannel(abR, baW)
	t.Cleanup(func() {
		_ = a.Close()
		_ = b.Close()
		_ = abW.Close()
		_ = baW.Close()
	})
	return a, b
}

// TestClassifyPaths checks the (files, dirs) split: an existing directory is
// classified as a dir, an existing file and a nonexistent path both as files
// (stat failure → treated as a file, matching parseLintFlags). All paths are
// abs-resolved and normalized.
func TestClassifyPaths(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.ts")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "nope.ts")

	files, dirs := classifyPaths([]string{dir, file, missing})

	wantDir := tspath.NormalizePath(dir)
	if len(dirs) != 1 || dirs[0] != wantDir {
		t.Errorf("dirs = %v, want [%s]", dirs, wantDir)
	}
	wantFiles := []string{tspath.NormalizePath(file), tspath.NormalizePath(missing)}
	if len(files) != 2 || files[0] != wantFiles[0] || files[1] != wantFiles[1] {
		t.Errorf("files = %v, want %v", files, wantFiles)
	}
}

// TestDrainStdoutToIPC_DiscardsOnClosedPeer pins #4: once the channel is
// closed (peer gone) the drain stops forwarding but keeps reading r — so the
// lint pipeline never blocks on a full stdout pipe — and exits cleanly on EOF.
func TestDrainStdoutToIPC_DiscardsOnClosedPeer(t *testing.T) {
	a, _ := newCLIChannelPair(t)
	a.Start()
	_ = a.Close() // peer gone → every SendNotification fails

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	done := make(chan struct{})
	go drainStdoutToIPC(pr, a, done)

	// Past the batch threshold so flush fires and hits the closed channel.
	if _, err := pw.Write(bytes.Repeat([]byte("x"), stdoutDrainMinFlushBytes+1)); err != nil {
		t.Fatal(err)
	}
	_ = pw.Close() // EOF → final flush + return

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("drain hung after the channel closed")
	}
}

// TestSplitAtUTF8Boundary covers the chunk-boundary cases drainStdoutToIPC
// depends on: ASCII and complete multibyte pass through whole; a chunk ending
// mid-character holds back the incomplete lead+continuation bytes; invalid
// lead bytes are passed through untouched (no recovery).
func TestSplitAtUTF8Boundary(t *testing.T) {
	cases := []struct {
		name           string
		in             string
		wantComplete   string
		wantIncomplete []byte
	}{
		{"ascii", "abc", "abc", nil},
		{"complete multibyte", "a世", "a世", nil},
		{"split 3-byte tail", "abc\xe4\xb8", "abc", []byte{0xe4, 0xb8}}, // 世 = e4 b8 96, last byte missing
		{"lead without continuation", "x\xc3", "x", []byte{0xc3}},       // 2-byte lead, continuation missing
		{"empty", "", "", nil},
		{"invalid lead byte", "ab\xff", "ab\xff", nil}, // 11111xxx → (buf, nil)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			complete, incomplete := splitAtUTF8Boundary([]byte(tc.in))
			if string(complete) != tc.wantComplete {
				t.Errorf("complete = %q, want %q", complete, tc.wantComplete)
			}
			if !bytes.Equal(incomplete, tc.wantIncomplete) {
				t.Errorf("incomplete = %v, want %v", incomplete, tc.wantIncomplete)
			}
		})
	}
}

// TestRunCLI_StdoutIsTTYWireToANSI pins the issue-#1080 seam end-to-end in
// process: the runtime.stdoutIsTTY fact of a real init frame must cross the
// wire (json tag), the payload→lintArgs merge, and the pipeline's color
// decision, putting ANSI escapes into the forwarded `output` frames — and
// only then. Every env tier above the TTY fact is neutralized so the result
// is identical on dev machines and CI.
func TestRunCLI_StdoutIsTTYWireToANSI(t *testing.T) {
	cases := []struct {
		name     string
		tty      bool
		wantANSI bool
	}{
		{"stdoutIsTTY=true colors output frames", true, true},
		{"stdoutIsTTY=false stays colorless", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runStdoutTTYCase(t, tc.tty, tc.wantANSI)
		})
	}
}

func runStdoutTTYCase(t *testing.T, tty, wantANSI bool) {
	// Truly unset NO_COLOR/FORCE_COLOR (set-but-empty would trip tiers 3/4);
	// t.Setenv first arranges restoration and marks the test non-parallel,
	// which the os.Stdin/os.Stdout/cwd swaps below require anyway.
	for _, key := range []string{"NO_COLOR", "FORCE_COLOR"} {
		if v, ok := os.LookupEnv(key); ok {
			t.Setenv(key, v)
			os.Unsetenv(key)
		}
	}
	t.Setenv("GITHUB_ACTIONS", "") // non-empty check → falls through to the TTY fact
	t.Setenv("TERM", "xterm-256color")

	// EvalSymlinks: macOS t.TempDir lives under the /var → /private/var
	// symlink; the pipeline matches normalized real paths, so anchor the
	// fixture at the physical path.
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFixture := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeFixture("tsconfig.json", `{"compilerOptions":{"strict":true},"include":["**/*.ts"]}`)
	writeFixture("index.ts", "// @ts-ignore\nconst a = 1;\n")

	// runCLI binds its IPC channel to the os.Stdin/os.Stdout globals and
	// changes directory to the payload workingDirectory; swap in pipe ends
	// and restore everything afterwards. t.Chdir into the current directory
	// is a no-op move that registers automatic cwd restoration.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(wd)
	stdinR, stdinW, err := os.Pipe() // test peer writes → CLI reads
	if err != nil {
		t.Fatal(err)
	}
	stdoutR, stdoutW, err := os.Pipe() // CLI writes → test peer reads
	if err != nil {
		t.Fatal(err)
	}
	origStdin, origStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdinR, stdoutW
	t.Cleanup(func() {
		os.Stdin, os.Stdout = origStdin, origStdout
		_ = stdinR.Close()
		_ = stdinW.Close()
		_ = stdoutR.Close()
		_ = stdoutW.Close()
	})

	// The test plays the Node peer on the other pipe ends.
	peer := ipc.NewChannel(stdoutR, stdinW)
	var outMu sync.Mutex
	var out bytes.Buffer
	peer.RegisterNotification(kindOutput, func(msg *ipc.Message) {
		var d struct {
			Text string `json:"text"`
		}
		if err := msg.Decode(&d); err != nil {
			return
		}
		outMu.Lock()
		out.WriteString(d.Text)
		outMu.Unlock()
	})
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		if msg.Kind == kindShutdown {
			return map[string]any{"ok": true}, nil
		}
		return nil, fmt.Errorf("unexpected inbound kind %q", msg.Kind)
	})
	peer.Start()
	t.Cleanup(func() { _ = peer.Close() })

	codeCh := make(chan int, 1)
	go func() { codeCh <- runCLI(nil) }()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err = peer.SendRequest(ctx, kindInit, map[string]any{
		"workingDirectory": dir,
		"runtime":          map[string]any{"stdoutIsTTY": tty},
		"configs": []map[string]any{{
			"configDirectory": dir,
			"entries": []map[string]any{{
				"files":   []string{"**/*.ts"},
				"rules":   map[string]any{"@typescript-eslint/ban-ts-comment": "error"},
				"plugins": []string{"@typescript-eslint"},
			}},
		}},
	})
	if err != nil {
		t.Fatalf("init request failed: %v", err)
	}

	var code int
	select {
	case code = <-codeCh:
	case <-time.After(60 * time.Second):
		t.Fatal("runCLI did not return within 60s")
	}

	// runCLI returns only after finalizeStdout drained every output frame and
	// the shutdown round-trip completed, so `out` is complete here.
	outMu.Lock()
	text := out.String()
	outMu.Unlock()
	if code != 1 {
		t.Errorf("exit code = %d, want 1 (one lint error)", code)
	}
	if !strings.Contains(text, "ban-ts-comment") {
		t.Fatalf("no diagnostic in forwarded output — lint did not run; output: %q", text)
	}
	if gotANSI := strings.Contains(text, "\x1b["); gotANSI != wantANSI {
		t.Errorf("ANSI in output = %v, want %v (stdoutIsTTY=%v); output: %q", gotANSI, wantANSI, tty, text)
	}
}

// TestHoistTypeSnapshots pins the CLI dispatcher's binary-trailer hoisting:
// snapshots move out of the JSON into the blobs slice, each file gets a 1-based
// index (0 for none), the index round-trips back to its blob, and the caller's
// request is left untouched.
func TestHoistTypeSnapshots(t *testing.T) {
	snapA := []byte{1, 2, 3}
	snapC := []byte{9, 9}
	orig := linter.EslintPluginLintRequest{
		Files: []linter.EslintPluginLintFile{
			{Path: "a.ts", TypeSnapshot: snapA},
			{Path: "b.ts"}, // no snapshot
			{Path: "c.ts", TypeSnapshot: snapC},
		},
	}
	got, blobs := hoistTypeSnapshots(orig)

	// blobs holds exactly the two non-empty snapshots, in file order.
	if len(blobs) != 2 {
		t.Fatalf("got %d blobs, want 2", len(blobs))
	}
	if !bytes.Equal(blobs[0], snapA) || !bytes.Equal(blobs[1], snapC) {
		t.Errorf("blobs = %v, want [%v %v]", blobs, snapA, snapC)
	}

	// Each snapshot-bearing file is hoisted out (TypeSnapshot nil) and gets a
	// 1-based index; a file with no snapshot keeps index 0 and nil.
	if got.Files[0].TypeSnapshot != nil || got.Files[0].TypeSnapshotIndex != 1 {
		t.Errorf("file a: snapshot=%v index=%d, want nil/1", got.Files[0].TypeSnapshot, got.Files[0].TypeSnapshotIndex)
	}
	if got.Files[1].TypeSnapshotIndex != 0 {
		t.Errorf("file b (no snapshot): index=%d, want 0", got.Files[1].TypeSnapshotIndex)
	}
	if got.Files[2].TypeSnapshot != nil || got.Files[2].TypeSnapshotIndex != 2 {
		t.Errorf("file c: snapshot=%v index=%d, want nil/2", got.Files[2].TypeSnapshot, got.Files[2].TypeSnapshotIndex)
	}

	// The 1-based index round-trips: blobs[index-1] is the file's snapshot.
	if !bytes.Equal(blobs[got.Files[0].TypeSnapshotIndex-1], snapA) {
		t.Error("file a's index does not point at its snapshot in blobs")
	}
	if !bytes.Equal(blobs[got.Files[2].TypeSnapshotIndex-1], snapC) {
		t.Error("file c's index does not point at its snapshot in blobs")
	}

	// The caller's original request is untouched (shallow-copy isolation).
	if !bytes.Equal(orig.Files[0].TypeSnapshot, snapA) || orig.Files[0].TypeSnapshotIndex != 0 {
		t.Errorf("original file a mutated: snapshot=%v index=%d", orig.Files[0].TypeSnapshot, orig.Files[0].TypeSnapshotIndex)
	}
	if !bytes.Equal(orig.Files[2].TypeSnapshot, snapC) {
		t.Errorf("original file c snapshot mutated: %v", orig.Files[2].TypeSnapshot)
	}
}

// TestHoistTypeSnapshots_NoSnapshots: with no file carrying a snapshot, blobs is
// empty (so WriteFrame emits a legacy JSON frame) and every index stays 0.
func TestHoistTypeSnapshots_NoSnapshots(t *testing.T) {
	orig := linter.EslintPluginLintRequest{
		Files: []linter.EslintPluginLintFile{{Path: "a.ts"}, {Path: "b.ts"}},
	}
	got, blobs := hoistTypeSnapshots(orig)
	if len(blobs) != 0 {
		t.Fatalf("got %d blobs, want 0", len(blobs))
	}
	for i, f := range got.Files {
		if f.TypeSnapshotIndex != 0 {
			t.Errorf("file %d index=%d, want 0", i, f.TypeSnapshotIndex)
		}
	}
}
