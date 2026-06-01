//go:build !js

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/ipc"
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
