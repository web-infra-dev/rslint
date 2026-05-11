package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	api "github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// ─────────────────────────────────────────────────────────────────────
// Inbound handler — init message decoding & unsupported-kind rejection.
// Full runCLI flow is exercised end-to-end by the JS cli test suite at
// packages/rslint-test-tools/tests/cli.
// ─────────────────────────────────────────────────────────────────────

func TestRunCLIInboundHandler_InitDecodesPayload(t *testing.T) {
	state := &runCLIState{payloadCh: make(chan *initPayload, 1)}
	h := &runCLIInboundHandler{state: state}

	payload := initPayload{
		Files:            []string{"a.ts", "b.ts"},
		WorkingDirectory: "/tmp/x",
		Format:           "jsonline",
		FixMode:          true,
		Runtime: runtimePayload{
			ForceColor:     true,
			SingleThreaded: true,
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var asInterface interface{}
	if err := json.Unmarshal(raw, &asInterface); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	msg := &api.Message{Kind: api.KindInit, ID: 1, Data: asInterface}
	resp, err := h.Handle(context.Background(), msg)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	asMap, _ := resp.(map[string]interface{})
	if asMap["ok"] != true {
		t.Errorf("init response: got %v, want {ok: true}", resp)
	}

	select {
	case got := <-state.payloadCh:
		if got == nil {
			t.Fatal("payloadCh delivered nil")
		}
		// Top-level user-flag mirrors round-trip exactly.
		if got.Format != "jsonline" {
			t.Errorf("Format: got %q, want %q", got.Format, "jsonline")
		}
		if !got.FixMode {
			t.Errorf("FixMode: got %v, want true", got.FixMode)
		}
		if got.WorkingDirectory != "/tmp/x" {
			t.Errorf("WorkingDirectory: got %q, want %q", got.WorkingDirectory, "/tmp/x")
		}
		// Files preserved verbatim (order + values).
		if len(got.Files) != 2 || got.Files[0] != "a.ts" || got.Files[1] != "b.ts" {
			t.Errorf("Files: got %v, want [a.ts b.ts]", got.Files)
		}
		// Runtime sub-struct round-trips field-for-field — broken JSON
		// tag drift on any of these would silently lose the user's
		// runtime hint.
		if !got.Runtime.ForceColor {
			t.Errorf("Runtime.ForceColor: got %v, want true", got.Runtime.ForceColor)
		}
		if !got.Runtime.SingleThreaded {
			t.Errorf("Runtime.SingleThreaded: got %v, want true", got.Runtime.SingleThreaded)
		}
	default:
		t.Fatal("payloadCh has no value after init handle")
	}
}

func TestRunCLIInboundHandler_InitTwiceDeliversOnce(t *testing.T) {
	state := &runCLIState{payloadCh: make(chan *initPayload, 1)}
	h := &runCLIInboundHandler{state: state}

	mkMsg := func() *api.Message {
		return &api.Message{
			Kind: api.KindInit,
			ID:   1,
			Data: map[string]interface{}{},
		}
	}

	resp1, err := h.Handle(context.Background(), mkMsg())
	if err != nil {
		t.Fatalf("first init: %v", err)
	}
	if asMap, _ := resp1.(map[string]interface{}); asMap["ok"] != true {
		t.Errorf("first init response: got %v, want {ok: true}", resp1)
	}
	// Drain the first delivery.
	<-state.payloadCh

	// Second init must STILL ack with {ok:true} (the protocol is idempotent
	// from the peer's perspective) but must NOT deliver a second payload
	// onto the channel — the lint pipeline already kicked off from the
	// first one and re-running it would race.
	resp2, err := h.Handle(context.Background(), mkMsg())
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	if asMap, _ := resp2.(map[string]interface{}); asMap["ok"] != true {
		t.Errorf("second init response: got %v, want {ok: true} (must remain idempotent)", resp2)
	}
	select {
	case extra := <-state.payloadCh:
		t.Fatalf("second init unexpectedly delivered a payload: %+v", extra)
	default:
	}
}

func TestRunCLIInboundHandler_UnsupportedKindErrors(t *testing.T) {
	state := &runCLIState{payloadCh: make(chan *initPayload, 1)}
	h := &runCLIInboundHandler{state: state}
	msg := &api.Message{Kind: api.KindLint, ID: 1, Data: nil}
	_, err := h.Handle(context.Background(), msg)
	if err == nil {
		t.Errorf("expected error for unsupported kind, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported inbound kind") {
		t.Errorf("error message: %s", err.Error())
	}
}

// ─────────────────────────────────────────────────────────────────────
// classifyPaths splits absolute paths into files vs dirs the same way
// parseLintFlags does for positional args, and runs each entry through
// tspath.NormalizePath so downstream FileScope comparisons key off the
// same canonical form regardless of which entry path produced them.
// ─────────────────────────────────────────────────────────────────────

func TestClassifyPaths_SplitsFilesAndDirs(t *testing.T) {
	tmp := t.TempDir()
	fileA := filepath.Join(tmp, "a.ts")
	fileB := filepath.Join(tmp, "b.ts")
	if err := os.WriteFile(fileA, []byte(""), 0o644); err != nil {
		t.Fatalf("write fileA: %v", err)
	}
	if err := os.WriteFile(fileB, []byte(""), 0o644); err != nil {
		t.Fatalf("write fileB: %v", err)
	}

	// Mix order on purpose: a dir between two files. classifyPaths must
	// not accidentally couple to insertion position when stat'ing.
	files, dirs := classifyPaths([]string{fileA, tmp, fileB})

	wantDir := tspath.NormalizePath(tmp)
	wantA := tspath.NormalizePath(fileA)
	wantB := tspath.NormalizePath(fileB)

	if len(dirs) != 1 || dirs[0] != wantDir {
		t.Errorf("dirs: got %v, want [%q]", dirs, wantDir)
	}
	// Strict file content + order: `fileA` and `fileB` exactly, in input
	// order (each file appears in `files` only — never in `dirs`).
	if len(files) != 2 {
		t.Fatalf("files len: got %v, want 2", files)
	}
	if files[0] != wantA {
		t.Errorf("files[0]: got %q, want %q", files[0], wantA)
	}
	if files[1] != wantB {
		t.Errorf("files[1]: got %q, want %q", files[1], wantB)
	}
}

func TestClassifyPaths_NonexistentTreatedAsFile(t *testing.T) {
	// Non-existent paths fall through to the file branch (caller emits a
	// `was not found in the project` warning later in the pipeline).
	const ghost = "/no/such/path/that/should/not/exist/12345.ts"
	files, dirs := classifyPaths([]string{ghost})
	wantGhost := tspath.NormalizePath(ghost)
	if len(dirs) != 0 {
		t.Errorf("dirs: got %v, want []", dirs)
	}
	if len(files) != 1 || files[0] != wantGhost {
		t.Errorf("files: got %v, want [%q]", files, wantGhost)
	}
}

// classifyPaths must resolve relative paths to absolute, mirroring
// parseLintFlags. The IPC entry path receives raw user positionals from
// `process.argv` (cli.ts forwards them as-is), which are commonly
// relative. Without filepath.Abs here, downstream FileScope matching
// falls through to the SameFile fallback (still works for in-program
// files) but gap-file detection at cmd.go's discovery loop runs
// FindNearestConfig with a relative path against absolute config keys
// — that lookup silently fails, and a file that genuinely IS only
// matched via config `files` pattern (no tsconfig membership) gets
// skipped from lint entirely.
func TestClassifyPaths_AbsolutizesRelative(t *testing.T) {
	// Chdir into a freshly-created tmpdir so we can pass relative paths
	// and verify they come out absolute. Restore cwd after — other
	// tests rely on the process cwd being stable.
	tmp := t.TempDir()
	t.Chdir(tmp)

	if err := os.WriteFile("a.ts", []byte(""), 0o644); err != nil {
		t.Fatalf("write a.ts: %v", err)
	}
	if err := os.Mkdir("subdir", 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	files, dirs := classifyPaths([]string{"a.ts", "subdir"})

	if len(files) != 1 {
		t.Fatalf("files: got %v, want 1 entry", files)
	}
	if !filepath.IsAbs(files[0]) {
		t.Errorf("files[0] = %q must be absolute", files[0])
	}
	if len(dirs) != 1 {
		t.Fatalf("dirs: got %v, want 1 entry", dirs)
	}
	if !filepath.IsAbs(dirs[0]) {
		t.Errorf("dirs[0] = %q must be absolute", dirs[0])
	}

	// Compare against tspath.NormalizePath(filepath.Abs(input)) — this is
	// the exact contract parseLintFlags's positional-arg block produces.
	wantFile, _ := filepath.Abs("a.ts")
	wantDir, _ := filepath.Abs("subdir")
	if files[0] != tspath.NormalizePath(wantFile) {
		t.Errorf("files[0] = %q, want %q", files[0], tspath.NormalizePath(wantFile))
	}
	if dirs[0] != tspath.NormalizePath(wantDir) {
		t.Errorf("dirs[0] = %q, want %q", dirs[0], tspath.NormalizePath(wantDir))
	}
}

func TestClassifyPaths_NormalizesPaths(t *testing.T) {
	// "/tmp/foo/./bar" should normalize to "/tmp/foo/bar". This is the
	// key behavior that lets IPC-supplied paths and CLI-supplied paths
	// produce equal FileScope strings — the rest of the pipeline keys
	// off these strings for config-dir matching, gitignore checks, etc.
	tmp := t.TempDir()
	target := filepath.Join(tmp, "x.ts")
	if err := os.WriteFile(target, []byte(""), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	noisy := filepath.Join(tmp, ".", "x.ts")

	files, _ := classifyPaths([]string{noisy})
	want := tspath.NormalizePath(target)
	if len(files) != 1 || files[0] != want {
		t.Errorf("files: got %v, want [%q] (normalized form)", files, want)
	}
}

// ─────────────────────────────────────────────────────────────────────
// startSynthStdinWriter — small-payload happy path, large-payload
// cancellation, and goroutine cleanup contract.
// ─────────────────────────────────────────────────────────────────────

func TestStartSynthStdinWriter_SmallPayloadCompletes(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()

	data := []byte(`{"configs":[]}`)
	done := startSynthStdinWriter(context.Background(), w, data)

	// Reader consumes; writer completes; done channel closes.
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("data mismatch: got %q want %q", got, data)
	}
	select {
	case <-done:
		// writer joined
	case <-time.After(2 * time.Second):
		t.Fatal("writer goroutine did not join within 2s after read completed")
	}
}

func TestStartSynthStdinWriter_CtxCancelUnblocksLargePayload(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()

	// > 1 MiB ensures we exceed the kernel pipe buffer (Linux ~64 KiB,
	// macOS ~16 KiB). Without ctx-cancel hookup, the writer would block
	// indefinitely because no reader is consuming.
	data := make([]byte, 1<<20)

	ctx, cancel := context.WithCancel(context.Background())
	done := startSynthStdinWriter(ctx, w, data)

	// Don't read. Cancel the ctx — writer must unblock and join.
	cancel()

	select {
	case <-done:
		// writer joined as expected
	case <-time.After(2 * time.Second):
		t.Fatal("writer goroutine did not join within 2s after ctx cancellation — likely deadlocked on pipe Write")
	}
}

func TestStartSynthStdinWriter_ReturnsImmediately(t *testing.T) {
	// The function must return quickly even for a payload that's larger
	// than the kernel pipe buffer — it kicks off the writer in a goroutine
	// rather than blocking on the first Write.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()

	data := make([]byte, 1<<20)

	start := time.Now()
	done := startSynthStdinWriter(context.Background(), w, data)
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("startSynthStdinWriter blocked for %v; expected near-instant return", elapsed)
	}

	// Drain so the writer can complete and we don't leak.
	go func() {
		_, _ = io.ReadAll(r)
	}()
	<-done
}

// ─────────────────────────────────────────────────────────────────────
// executeLintPipeline — compat dispatcher injection contract.
// Two distinct callers running back-to-back must see exactly the
// dispatcher each one passed, with no leak from a previous call.
// Regression for the package-level `compatRuleDispatcher` global which
// would race or leak when reused.
// ─────────────────────────────────────────────────────────────────────

// The previous test in this slot ("DispatcherDoesNotLeakBetweenCalls")
// only asserted two locally-constructed closures are non-nil; it never
// invoked executeLintPipeline. The actual no-package-globals invariant
// is enforced by the function signature (CompatBatchHandler is taken
// as a parameter, no package-level dispatcher exists) and would
// produce a compile error if regressed — so a runtime test is
// redundant. Removed in favor of the compile-time guard.

// ─────────────────────────────────────────────────────────────────────
// parseLintFlags — fresh-FlagSet contract: callable multiple times per
// process without "flag redefined" panic. The pre-refactor implementation
// used flag.CommandLine (global) and would panic on the second call.
// Tests/future-callers must be able to invoke parseLintFlags repeatedly.
// ─────────────────────────────────────────────────────────────────────

func TestParseLintFlags_CallableMultipleTimes(t *testing.T) {
	// First call with one set of args.
	args1, help1, exit1 := parseLintFlags([]string{"--format=jsonline", "--fix"})
	if exit1 != 0 {
		t.Fatalf("first call: fatalExitCode = %d, want 0", exit1)
	}
	if help1 {
		t.Errorf("first call: help should be false")
	}
	if args1.Format != "jsonline" {
		t.Errorf("first call: Format = %q, want %q", args1.Format, "jsonline")
	}
	if !args1.Fix {
		t.Errorf("first call: Fix should be true")
	}

	// Second call with DIFFERENT args. With the old flag.CommandLine-based
	// implementation, the flag.BoolVar etc. registrations would panic on
	// the duplicate flag name. The fresh FlagSet pattern allows arbitrary
	// repetition.
	args2, help2, exit2 := parseLintFlags([]string{"--help"})
	if exit2 != 0 {
		t.Fatalf("second call: fatalExitCode = %d, want 0", exit2)
	}
	if !help2 {
		t.Errorf("second call: help should be true")
	}
	if args2.Fix {
		t.Errorf("second call: Fix should be false (default), got true — state leaked from first call")
	}

	// Third call: ensure args1 doesn't observe the second call's state.
	// The lintArgs returned is a local value; this is a guard against an
	// accidental switch back to package-level mutables in some future
	// refactor.
	if args1.Format != "jsonline" {
		t.Errorf("args1.Format mutated after second call: got %q, want %q (state leaked)",
			args1.Format, "jsonline")
	}
}

// parseLintFlags must not call os.Exit on a parse error — it must return
// a non-zero fatalExitCode so the IPC shutdown / stdout drain can run.
// Validates the ContinueOnError mode of the new FlagSet.
func TestParseLintFlags_BadFlagReturnsFatal(t *testing.T) {
	_, _, exitCode := parseLintFlags([]string{"--definitely-not-a-flag"})
	if exitCode == 0 {
		t.Errorf("parseLintFlags accepted unknown flag; expected fatalExitCode != 0")
	}
}

// ─────────────────────────────────────────────────────────────────────
// splitAtUTF8Boundary — held-back-tail behavior for the stdout drain.
// ─────────────────────────────────────────────────────────────────────

func TestSplitAtUTF8Boundary(t *testing.T) {
	cases := []struct {
		name           string
		in             []byte
		wantComplete   []byte
		wantIncomplete []byte
	}{
		{"empty", nil, nil, nil},
		{"ascii only", []byte("hello"), []byte("hello"), nil},
		// "é" = U+00E9 = 0xC3 0xA9 (2-byte UTF-8)
		{"complete 2-byte at end", []byte("a\xC3\xA9"), []byte("a\xC3\xA9"), nil},
		{"incomplete 2-byte (lead only)", []byte("ab\xC3"), []byte("ab"), []byte("\xC3")},
		// "中" = U+4E2D = 0xE4 0xB8 0xAD (3-byte UTF-8)
		{"complete 3-byte at end", []byte("ab\xE4\xB8\xAD"), []byte("ab\xE4\xB8\xAD"), nil},
		{"incomplete 3-byte (lead only)", []byte("ab\xE4"), []byte("ab"), []byte("\xE4")},
		{"incomplete 3-byte (lead + 1 cont)", []byte("ab\xE4\xB8"), []byte("ab"), []byte("\xE4\xB8")},
		// "😀" = U+1F600 = 0xF0 0x9F 0x98 0x80 (4-byte UTF-8)
		{"complete 4-byte at end", []byte("ab\xF0\x9F\x98\x80"), []byte("ab\xF0\x9F\x98\x80"), nil},
		{"incomplete 4-byte (lead only)", []byte("ab\xF0"), []byte("ab"), []byte("\xF0")},
		{"incomplete 4-byte (lead + 2 cont)", []byte("ab\xF0\x9F\x98"), []byte("ab"), []byte("\xF0\x9F\x98")},
		// Corrupt stream: tail is all continuation bytes (no lead in last 4).
		// We can't recover; flush as-is so the peer sees U+FFFD.
		{"all-continuation tail", []byte("\x80\x80\x80"), []byte("\x80\x80\x80"), nil},
		// Invalid 5+-byte lead byte: same — don't try to recover.
		{"invalid 5-byte lead", []byte("ab\xFF"), []byte("ab\xFF"), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotComplete, gotIncomplete := splitAtUTF8Boundary(tc.in)
			if !bytes.Equal(gotComplete, tc.wantComplete) {
				t.Errorf("complete: got %x, want %x", gotComplete, tc.wantComplete)
			}
			if !bytes.Equal(gotIncomplete, tc.wantIncomplete) {
				t.Errorf("incomplete: got %x, want %x", gotIncomplete, tc.wantIncomplete)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────
// drainStdoutToIPC — boundary-correctness end-to-end.
// Wires the drain goroutine to a mock IPC peer and verifies that a
// multi-byte UTF-8 string written across a 4096-byte read boundary
// reassembles byte-for-byte on the receiving side.
// ─────────────────────────────────────────────────────────────────────

// captureIPC drives a BidirectionalService whose `output` notifications
// are appended to a strings.Builder. Returns the builder and a cleanup.
//
// The mock peer is a second BidirectionalService wired to A via two
// io.Pipes. A's `output` notification → peer's notification handler
// → accumulator. Returns a thread-safe snapshot accessor instead of a
// bare strings.Builder to keep test readers race-free without exposing
// the internal mutex.
type capturedOutput struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (c *capturedOutput) write(s string) {
	c.mu.Lock()
	c.buf.WriteString(s)
	c.mu.Unlock()
}

func (c *capturedOutput) snapshot() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

func (c *capturedOutput) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.Len()
}

func captureIPC(t *testing.T) (a *api.BidirectionalService, captured *capturedOutput, cleanup func()) {
	t.Helper()
	readBFromA, writeAToB := io.Pipe()
	readAFromB, writeBToA := io.Pipe()
	a = api.NewBidirectionalService(readAFromB, writeAToB)
	b := api.NewBidirectionalService(readBFromA, writeBToA)
	captured = &capturedOutput{}
	b.RegisterNotification(api.KindOutput, func(_ context.Context, msg *api.Message) {
		var rec struct {
			Stream string `json:"stream"`
			Text   string `json:"text"`
		}
		raw, _ := json.Marshal(msg.Data)
		_ = json.Unmarshal(raw, &rec)
		captured.write(rec.Text)
	})
	a.Start()
	b.Start()
	cleanup = func() {
		_ = a.Close()
		_ = b.Close()
		_ = writeAToB.Close()
		_ = writeBToA.Close()
	}
	return a, captured, cleanup
}

func TestDrainStdoutToIPC_PreservesUTF8AcrossBoundary(t *testing.T) {
	a, captured, cleanup := captureIPC(t)
	defer cleanup()

	stdoutR, stdoutW := io.Pipe()
	done := make(chan struct{})
	go drainStdoutToIPC(stdoutR, a, done)

	// Build a payload where a 3-byte CJK character spans the 4096-byte
	// internal Read buffer. We position "中" (0xE4 0xB8 0xAD) so its first
	// byte is at offset 4095 — the next two bytes go into the second
	// Read. A pre-fix implementation would corrupt this character via
	// the U+FFFD replacement at the encode step.
	prefix := strings.Repeat("a", 4095)
	cjk := "中" // 3 bytes
	suffix := strings.Repeat("b", 4096)
	original := prefix + cjk + suffix

	go func() {
		_, _ = stdoutW.Write([]byte(original))
		_ = stdoutW.Close() // signals EOF, drain flushes & exits
	}()

	<-done

	// Allow the IPC notification dispatch to flush (notifications are
	// async on the peer's reader goroutine).
	deadline := time.After(2 * time.Second)
	for captured.len() < len(original) {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for output (got %d/%d bytes)",
				captured.len(), len(original))
		case <-time.After(5 * time.Millisecond):
		}
	}

	got := captured.snapshot()
	if got != original {
		// Diagnose by showing first 5 chars of difference.
		for i := 0; i < len(got) && i < len(original); i++ {
			if got[i] != original[i] {
				lo := i - 3
				if lo < 0 {
					lo = 0
				}
				hi := i + 5
				if hi > len(got) {
					hi = len(got)
				}
				t.Fatalf("byte %d differs: got %x, want %x (context got=%x want=%x)",
					i, got[i], original[i], got[lo:hi], original[lo:hi])
			}
		}
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(original))
	}
}

// drainStdoutToIPC writes pure ASCII end-to-end byte-for-byte too.
// Sanity check that the UTF-8 hold-back logic doesn't false-positive on
// content that has no multi-byte characters at all.
func TestDrainStdoutToIPC_ASCIIPassthrough(t *testing.T) {
	a, captured, cleanup := captureIPC(t)
	defer cleanup()

	stdoutR, stdoutW := io.Pipe()
	done := make(chan struct{})
	go drainStdoutToIPC(stdoutR, a, done)

	// > stdoutDrainMinFlushBytes so we get multiple flushes, exercising
	// the batch path too.
	payload := strings.Repeat("rslint diagnostic line\n", 1024)

	go func() {
		_, _ = stdoutW.Write([]byte(payload))
		_ = stdoutW.Close()
	}()

	<-done

	deadline := time.After(2 * time.Second)
	for captured.len() < len(payload) {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for ASCII output (got %d/%d)",
				captured.len(), len(payload))
		case <-time.After(5 * time.Millisecond):
		}
	}
	if captured.snapshot() != payload {
		t.Errorf("ASCII content not preserved end-to-end")
	}
}

// inboundHandlerFunc adapts a plain function to api.InboundHandler for
// these dispatcher-level tests. The api package doesn't export an
// equivalent helper; declaring it here keeps the test self-contained.
type inboundHandlerFunc func(context.Context, *api.Message) (interface{}, error)

func (f inboundHandlerFunc) Handle(ctx context.Context, m *api.Message) (interface{}, error) {
	return f(ctx, m)
}

// ─────────────────────────────────────────────────────────────────────
// makeIPCDispatcher — cancellation contract.
//
// Background: the previous version of makeIPCDispatcher wrapped each
// SendRequest in `context.WithTimeout(ctx, RSLINT_COMPAT_TIMEOUT_MS)`.
// That layer was removed because the Node-side WorkerPool already
// enforces a per-task watchdog (default 30s, terminate+respawn on
// hang), and the Go-side ceiling was an arbitrary footgun for large
// monorepo batches.
//
// What's left as the cancellation contract:
//   - The caller's ctx (the lint-level SIGINT context) is passed
//     straight through; if it cancels, SendRequest returns ctx.Err.
//   - If the peer closes (bs.Close on the other side or stdin EOF),
//     bs.stopCh fires inside SendRequest and the dispatcher returns
//     a wrapped "peer closed" error.
//
// These tests pin both signals so a future refactor cannot drop them
// without making the test red.
// ─────────────────────────────────────────────────────────────────────

func TestMakeIPCDispatcher_PropagatesCtxCancel(t *testing.T) {
	// A→B pair: A is the rslint Go process (we call its dispatcher),
	// B simulates the Node host. B intentionally never replies — we
	// only want to verify that cancelling A's ctx unblocks the
	// dispatcher.
	readBFromA, writeAToB := io.Pipe()
	readAFromB, writeBToA := io.Pipe()
	defer writeAToB.Close()
	defer writeBToA.Close()
	a := api.NewBidirectionalService(readAFromB, writeAToB)
	b := api.NewBidirectionalService(readBFromA, writeBToA)
	// B accepts the request but never responds — that's the wedge
	// we're simulating. Registering an inbound handler avoids the
	// "unsupported kind" error path muddying the test.
	b.SetInboundHandler(inboundHandlerFunc(func(_ context.Context, _ *api.Message) (interface{}, error) {
		select {} // block forever; ctx cancel on A side is what we test
	}))
	a.Start()
	b.Start()
	defer a.Close()
	defer b.Close()

	dispatch := makeIPCDispatcher(a)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel shortly after dispatch starts — the dispatcher must
	// return ctx.Err quickly, NOT wait for any timeout.
	go func() {
		time.Sleep(80 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := dispatch(ctx, linter.CompatBatch{})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error when ctx is cancelled, got nil")
	}
	// The error wraps ctx.Err. SendRequest returns ctx.Err verbatim,
	// dispatcher wraps it in "lintEslintPlugin IPC: <ctx err>".
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected ctx.Err propagation, got %v", err)
	}
	// Must NOT wait anywhere near the old 60s timeout — cancel
	// propagation has to be near-instant.
	if elapsed > time.Second {
		t.Errorf("dispatcher returned after %v; ctx cancel should be sub-second", elapsed)
	}
}

func TestMakeIPCDispatcher_PropagatesPeerClose(t *testing.T) {
	// Same setup as above, but instead of cancelling A's ctx we close
	// B (the peer). bs.stopCh on A's side fires while A is awaiting
	// the response, and SendRequest returns a peer-closed error.
	readBFromA, writeAToB := io.Pipe()
	readAFromB, writeBToA := io.Pipe()
	defer writeAToB.Close()
	defer writeBToA.Close()
	a := api.NewBidirectionalService(readAFromB, writeAToB)
	b := api.NewBidirectionalService(readBFromA, writeBToA)
	b.SetInboundHandler(inboundHandlerFunc(func(_ context.Context, _ *api.Message) (interface{}, error) {
		// Don't reply — wait for the test to close us.
		select {}
	}))
	a.Start()
	b.Start()
	defer a.Close()

	dispatch := makeIPCDispatcher(a)

	// Kick off the dispatch in a goroutine; close B from the main
	// goroutine to trigger the peer-closed path on A.
	type result struct {
		err error
		dur time.Duration
	}
	resCh := make(chan result, 1)
	go func() {
		start := time.Now()
		_, err := dispatch(context.Background(), linter.CompatBatch{})
		resCh <- result{err: err, dur: time.Since(start)}
	}()

	time.Sleep(80 * time.Millisecond)
	// Close B's pipes so A sees EOF on read; that drives bs.stopCh.
	_ = writeBToA.Close()
	_ = readBFromA.Close()

	select {
	case r := <-resCh:
		if r.err == nil {
			t.Fatal("expected error when peer closes mid-request, got nil")
		}
		// Must NOT block anywhere close to the old 60s ceiling.
		if r.dur > 2*time.Second {
			t.Errorf("dispatcher returned after %v; peer close should be near-instant", r.dur)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("dispatcher never returned after peer closed")
	}
}
