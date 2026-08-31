//go:build !js

package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
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

func TestDiscoverCLIExplicitConfigProjectsMixedTargetsOnly(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	selectedDir := filepath.Join(root, "selected")
	unrelatedDir := filepath.Join(root, "unrelated")
	toolsDir := filepath.Join(root, "tools")
	for _, path := range []string{selectedDir, filepath.Join(selectedDir, "deep"), unrelatedDir, toolsDir} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
	}
	for path, content := range map[string]string{
		filepath.Join(root, ".gitignore"):             "root.generated.js\n",
		filepath.Join(selectedDir, ".gitignore"):      "selected.generated.js\n",
		filepath.Join(selectedDir, "deep/.gitignore"): "deep.generated.js\n",
		filepath.Join(unrelatedDir, ".gitignore"):     "unrelated.generated.js\n",
		filepath.Join(toolsDir, ".gitignore"):         "exact.js\n",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	configPath := filepath.Join(root, "custom.config.cjs")
	if err := os.WriteFile(configPath, []byte("export default [];\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	exactFile := filepath.Join(toolsDir, "exact.js")
	if err := os.WriteFile(exactFile, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cli, peer := newCLIChannelPair(t)
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		switch msg.Kind {
		case kindLoadConfigs:
			var request discovery.ConfigLoadBatchRequest
			if err := msg.Decode(&request); err != nil {
				return nil, err
			}
			if len(request.Candidates) != 1 || request.Candidates[0].ConfigPath != tspath.NormalizePath(configPath) {
				return nil, fmt.Errorf("unexpected config candidates: %+v", request.Candidates)
			}
			return discovery.ConfigLoadBatchResponse{
				TransactionID: request.TransactionID,
				Results: []discovery.ConfigLoadResult{{
					ID:     request.Candidates[0].ID,
					Status: "loaded",
				}},
			}, nil
		case kindActivateConfigs:
			var request discovery.ConfigActivationRequest
			if err := msg.Decode(&request); err != nil {
				return nil, err
			}
			return discovery.ConfigActivationResponse{TransactionID: request.TransactionID}, nil
		default:
			return nil, fmt.Errorf("unexpected request kind %q", msg.Kind)
		}
	})
	cli.Start()
	peer.Start()
	spy := &directoryAccessSpyFS{FS: osvfs.FS()}
	args := lintArgs{
		FS:             spy,
		AllowFiles:     []string{tspath.NormalizePath(exactFile)},
		AllowDirs:      []string{tspath.NormalizePath(selectedDir)},
		SingleThreaded: true,
	}
	payload := initPayload{ConfigDiscovery: &configDiscoveryPayload{ExplicitConfigPath: configPath}}

	if err := discoverCLIConfigCatalog(context.Background(), &args, &payload, cli); err != nil {
		t.Fatalf("discoverCLIConfigCatalog: %v", err)
	}
	if args.ConfigCatalog == nil || !args.ConfigCatalog.Explicit {
		t.Fatal("explicit config catalog was not installed")
	}
	unrelatedDir = tspath.NormalizePath(unrelatedDir)
	for _, accessed := range spy.accessedDirs {
		if accessed == unrelatedDir || tspath.StartsWithDirectory(accessed, unrelatedDir, true) {
			t.Fatalf("explicit CLI projection entered unrelated directory %q", accessed)
		}
	}
}

func TestMarkCLIInterruptedUsesCanceledDiscoveryContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	state := &runCLIState{}
	if !markCLIInterrupted(ctx, state) {
		t.Fatal("canceled discovery context was not classified as an interrupt")
	}
	if !state.signalled.Load() {
		t.Fatal("canceled discovery context did not publish the signal state")
	}

	if markCLIInterrupted(context.Background(), &runCLIState{}) {
		t.Fatal("live discovery context was classified as an interrupt")
	}
}

func TestShutdownPeerWaitsForAcknowledgement(t *testing.T) {
	cli, peer := newCLIChannelPair(t)
	requestSeen := make(chan struct{})
	releaseResponse := make(chan struct{})
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		if msg.Kind != kindShutdown {
			return nil, fmt.Errorf("request kind = %q, want %q", msg.Kind, kindShutdown)
		}
		var payload struct{}
		if err := msg.Decode(&payload); err != nil {
			return nil, fmt.Errorf("decode shutdown payload: %w", err)
		}
		close(requestSeen)
		<-releaseResponse
		return map[string]any{"ok": true}, nil
	})
	cli.Start()
	peer.Start()

	shutdownDone := make(chan bool, 1)
	go func() {
		shutdownDone <- shutdownPeer(cli, &runCLIState{})
	}()

	select {
	case <-requestSeen:
	case <-time.After(2 * time.Second):
		t.Fatal("peer did not receive shutdown request")
	}
	select {
	case <-shutdownDone:
		t.Fatal("shutdownPeer returned before the peer acknowledged")
	default:
	}
	close(releaseResponse)
	select {
	case acknowledged := <-shutdownDone:
		if !acknowledged {
			t.Fatal("shutdownPeer did not report the acknowledgement")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdownPeer did not return after the acknowledgement")
	}
}

func TestAcknowledgedOutputWriterWaitsForPeer(t *testing.T) {
	cli, peer := newCLIChannelPair(t)
	requestSeen := make(chan string, 1)
	releaseResponse := make(chan struct{})
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		if msg.Kind != kindOutput {
			return nil, fmt.Errorf("request kind = %q, want %q", msg.Kind, kindOutput)
		}
		var payload struct {
			Stream string `json:"stream"`
			Text   string `json:"text"`
		}
		if err := msg.Decode(&payload); err != nil {
			return nil, err
		}
		if payload.Stream != "stdout" {
			return nil, fmt.Errorf("stream = %q, want stdout", payload.Stream)
		}
		requestSeen <- payload.Text
		<-releaseResponse
		return map[string]any{"ok": true}, nil
	})
	cli.Start()
	peer.Start()

	text := "\x1b[36;1m" + "start" + "\x1b[0;22m   Linting...\n"
	writeDone := make(chan error, 1)
	go func() {
		written, err := (acknowledgedOutputWriter{
			ctx:     context.Background(),
			channel: cli,
		}).Write([]byte(text))
		if err == nil && written != len(text) {
			err = fmt.Errorf("written = %d, want %d", written, len(text))
		}
		writeDone <- err
	}()

	select {
	case got := <-requestSeen:
		if got != text {
			t.Fatalf("output text = %q, want %q", got, text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("peer did not receive acknowledged output request")
	}
	select {
	case err := <-writeDone:
		t.Fatalf("Write returned before peer acknowledgement: %v", err)
	default:
	}
	close(releaseResponse)
	select {
	case err := <-writeDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Write did not return after peer acknowledgement")
	}
}

func TestAcknowledgedOutputWriterReturnsOnCancellation(t *testing.T) {
	cli, peer := newCLIChannelPair(t)
	requestSeen := make(chan struct{})
	releaseResponse := make(chan struct{})
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		if msg.Kind != kindOutput {
			return nil, fmt.Errorf("request kind = %q, want %q", msg.Kind, kindOutput)
		}
		close(requestSeen)
		<-releaseResponse
		return map[string]any{"ok": true}, nil
	})
	cli.Start()
	peer.Start()

	ctx, cancel := context.WithCancel(context.Background())
	writeDone := make(chan error, 1)
	go func() {
		written, err := (acknowledgedOutputWriter{ctx: ctx, channel: cli}).Write([]byte("start\n"))
		if written != 0 {
			err = fmt.Errorf("written = %d, want 0 after cancellation", written)
		}
		writeDone <- err
	}()

	select {
	case <-requestSeen:
	case <-time.After(2 * time.Second):
		t.Fatal("peer did not receive acknowledged output request")
	}
	cancel()
	select {
	case err := <-writeDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Write error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Write did not return after cancellation")
	}
	close(releaseResponse)
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

func TestRunCLIRejectsPayloadFormatBeforeConfigDiscovery(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, "rslint.config.mjs"),
		[]byte("export default [];\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	code, output := runCLIInitForTest(t, map[string]any{
		"workingDirectory": dir,
		"format":           "stylish",
		"configDiscovery":  map[string]any{},
	}, nil, 0, false)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; output=%q", code, output)
	}
	if output != "" {
		t.Fatalf("invalid format unexpectedly produced stdout: %q", output)
	}
}

func TestRunCLIMachineFormatDoesNotRequestAcknowledgedStart(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dir, "index.ts")
	if err := os.WriteFile(filePath, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "rslint.config.mjs"),
		[]byte("export default [];\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	code, output := runCLIInitForTest(t, map[string]any{
		"workingDirectory": dir,
		"files":            []string{filePath},
		"format":           "jsonline",
		"configDiscovery":  map[string]any{},
	}, config.RslintConfig{{
		Files: []string{"**/*.ts"},
		Rules: config.Rules{"no-debugger": "error"},
	}}, 0, true)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1; output=%q", code, output)
	}
	if !strings.Contains(output, `"ruleName":"no-debugger"`) {
		t.Fatalf("machine diagnostic missing: %q", output)
	}
	if strings.Contains(output, "Linting...") || strings.Contains(output, "Lint failed") {
		t.Fatalf("machine output contains default lifecycle text: %q", output)
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
	writeFixture("rslint.config.mjs", "export default [];\n")

	code, text := runCLIInitForTest(t, map[string]any{
		"workingDirectory": dir,
		"runtime":          map[string]any{"stdoutIsTTY": tty},
		"configDiscovery":  map[string]any{},
	}, config.RslintConfig{{
		Files:   []string{"**/*.ts"},
		Rules:   config.Rules{"@typescript-eslint/ban-ts-comment": "error"},
		Plugins: []string{"@typescript-eslint"},
	}}, 1, true)
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

// TestRunCLI_WorkingDirectoryAliases pins the real CLI seam that made released
// versions silently lint zero files when launched from a symlinked cwd. Shells
// preserve the lexical cwd in PWD, while Node reports the physical cwd in the
// init payload. The second case also drives the payload itself through the
// alias, covering Windows directory-junction invocation.
func TestRunCLI_WorkingDirectoryAliases(t *testing.T) {
	baseDir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve physical temp directory: %v", err)
	}
	realDir := filepath.Join(baseDir, "real")
	aliasDir := filepath.Join(baseDir, "alias")
	if err := os.Mkdir(realDir, 0o755); err != nil {
		t.Fatalf("mkdir real working directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "index.ts"), []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write lint target: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(realDir, "rslint.config.mjs"),
		[]byte("export default [];\n"),
		0o644,
	); err != nil {
		t.Fatalf("write lint config: %v", err)
	}
	createWorkingDirectoryAlias(t, realDir, aliasDir)

	for _, test := range []struct {
		name             string
		workingDirectory string
	}{
		{name: "physical payload with lexical PWD", workingDirectory: realDir},
		{name: "lexical payload", workingDirectory: aliasDir},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Reproduce a logical shell cwd. On Unix os.Getwd may return this
			// spelling after runCLI changes to the physical payload directory.
			t.Setenv("PWD", aliasDir)
			t.Setenv("NO_COLOR", "1")
			code, text := runCLIInitForTest(t, map[string]any{
				"workingDirectory": test.workingDirectory,
				"files":            []string{filepath.Join(test.workingDirectory, "index.ts")},
				"configDiscovery":  map[string]any{},
			}, config.RslintConfig{{
				Files: []string{"**/*.ts"},
				Rules: config.Rules{"no-debugger": "error"},
			}}, 1, true)
			if code != 1 {
				t.Fatalf("exit code = %d, want 1; output: %q", code, text)
			}
			if !strings.Contains(text, "no-debugger") || !strings.Contains(text, "(1 file, 1 rule,") {
				t.Fatalf("symlinked working directory did not lint index.ts exactly once; output: %q", text)
			}
		})
	}
}

func createWorkingDirectoryAlias(t *testing.T, target, alias string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		// Directory symlinks can require Developer Mode or elevated privileges
		// on Windows, so use an ordinary-user directory junction instead.
		// cspell:ignore mklink
		if output, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", alias, target).CombinedOutput(); err != nil {
			t.Fatalf("create working-directory junction: %v\n%s", err, output)
		}
	} else if err := os.Symlink(target, alias); err != nil {
		t.Fatalf("create working-directory symlink: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(alias); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove working-directory alias: %v", err)
		}
	})
}

const cliTestInitRequestID = 1

type cliTestPeerResult struct {
	output                     string
	initReplies                int
	acknowledgedOutputRequests int
	shutdownRequests           int
	err                        error
}

// serveCLITestPeer models the ordering that matters in the real Node peer:
// output requests and notifications are handled synchronously as frames are
// decoded, and only then can a later shutdown request be acknowledged. Using
// ipc.Channel here would be incorrect because its Go-side notification and
// request handlers intentionally run in independent goroutines and may
// overtake one another.
func serveCLITestPeer(
	fromCLI io.Reader,
	toCLI io.Writer,
	loadedConfig config.RslintConfig,
) cliTestPeerResult {
	reader := bufio.NewReader(fromCLI)
	var output bytes.Buffer
	result := cliTestPeerResult{}
	shutdownSeen := false
	recordError := func(err error) {
		if result.err == nil {
			result.err = err
		}
	}

	for {
		msg, err := ipc.ReadFrame(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				recordError(fmt.Errorf("read CLI frame: %w", err))
			}
			result.output = output.String()
			return result
		}

		switch {
		case (msg.Kind == ipc.KindResponse || msg.Kind == ipc.KindError) && msg.ID == cliTestInitRequestID:
			result.initReplies++
			if msg.Kind == ipc.KindError {
				var response ipc.ErrorResponseData
				if err := msg.Decode(&response); err != nil {
					recordError(fmt.Errorf("decode init error response: %w", err))
				} else {
					recordError(fmt.Errorf("init rejected: %s", response.Message))
				}
				continue
			}
			var response struct {
				OK bool `json:"ok"`
			}
			if err := msg.Decode(&response); err != nil {
				recordError(fmt.Errorf("decode init response: %w", err))
			} else if !response.OK {
				recordError(errors.New("init response did not acknowledge the request"))
			}

		case msg.ID == 0 && msg.Kind == kindOutput:
			if shutdownSeen {
				recordError(errors.New("received output after shutdown"))
			}
			var notification struct {
				Stream string `json:"stream"`
				Text   string `json:"text"`
			}
			if err := msg.Decode(&notification); err != nil {
				recordError(fmt.Errorf("decode output notification: %w", err))
				continue
			}
			if notification.Stream != "stdout" {
				recordError(fmt.Errorf("output stream = %q, want stdout", notification.Stream))
			}
			output.WriteString(notification.Text)

		case msg.ID > 0 && msg.Kind == kindOutput:
			result.acknowledgedOutputRequests++
			if shutdownSeen {
				recordError(errors.New("received acknowledged output after shutdown"))
			}
			var request struct {
				Stream string `json:"stream"`
				Text   string `json:"text"`
			}
			if err := msg.Decode(&request); err != nil {
				recordError(fmt.Errorf("decode acknowledged output request: %w", err))
				continue
			}
			if request.Stream != "stdout" {
				recordError(fmt.Errorf("output stream = %q, want stdout", request.Stream))
			}
			output.WriteString(request.Text)
			response, err := ipc.NewMessage(ipc.KindResponse, msg.ID, map[string]any{"ok": true})
			if err != nil {
				recordError(fmt.Errorf("build acknowledged output response: %w", err))
				continue
			}
			if err := ipc.WriteFrame(toCLI, response); err != nil {
				recordError(fmt.Errorf("write acknowledged output response: %w", err))
			}

		case msg.ID > 0 && msg.Kind == kindLoadConfigs:
			var request discovery.ConfigLoadBatchRequest
			if err := msg.Decode(&request); err != nil {
				recordError(fmt.Errorf("decode loadConfigs request: %w", err))
				continue
			}
			results := make([]discovery.ConfigLoadResult, len(request.Candidates))
			for index, candidate := range request.Candidates {
				results[index] = discovery.ConfigLoadResult{
					ID:      candidate.ID,
					Status:  "loaded",
					Entries: loadedConfig,
				}
			}
			response, err := ipc.NewMessage(ipc.KindResponse, msg.ID, discovery.ConfigLoadBatchResponse{
				TransactionID: request.TransactionID,
				Results:       results,
			})
			if err != nil {
				recordError(fmt.Errorf("build loadConfigs response: %w", err))
				continue
			}
			if err := ipc.WriteFrame(toCLI, response); err != nil {
				recordError(fmt.Errorf("write loadConfigs response: %w", err))
			}

		case msg.ID > 0 && msg.Kind == kindActivateConfigs:
			var request discovery.ConfigActivationRequest
			if err := msg.Decode(&request); err != nil {
				recordError(fmt.Errorf("decode activateConfigs request: %w", err))
				continue
			}
			response, err := ipc.NewMessage(ipc.KindResponse, msg.ID, discovery.ConfigActivationResponse{
				TransactionID: request.TransactionID,
			})
			if err != nil {
				recordError(fmt.Errorf("build activateConfigs response: %w", err))
				continue
			}
			if err := ipc.WriteFrame(toCLI, response); err != nil {
				recordError(fmt.Errorf("write activateConfigs response: %w", err))
			}

		case msg.ID > 0 && msg.Kind == kindShutdown:
			result.shutdownRequests++
			if shutdownSeen {
				recordError(errors.New("received duplicate shutdown request"))
			}
			shutdownSeen = true
			var payload struct{}
			if err := msg.Decode(&payload); err != nil {
				recordError(fmt.Errorf("decode shutdown payload: %w", err))
			}
			response, err := ipc.NewMessage(ipc.KindResponse, msg.ID, map[string]any{"ok": true})
			if err != nil {
				recordError(fmt.Errorf("build shutdown response: %w", err))
				continue
			}
			if err := ipc.WriteFrame(toCLI, response); err != nil {
				recordError(fmt.Errorf("write shutdown response: %w", err))
			}

		default:
			recordError(fmt.Errorf("unexpected CLI frame kind=%q id=%d", msg.Kind, msg.ID))
			if msg.ID > 0 && msg.Kind != ipc.KindResponse && msg.Kind != ipc.KindError {
				response, responseErr := ipc.NewMessage(ipc.KindError, msg.ID, ipc.ErrorResponseData{
					Message: fmt.Sprintf("unexpected test-peer request kind %q", msg.Kind),
				})
				if responseErr != nil {
					recordError(fmt.Errorf("build unexpected-frame response: %w", responseErr))
					continue
				}
				if responseErr := ipc.WriteFrame(toCLI, response); responseErr != nil {
					recordError(fmt.Errorf("write unexpected-frame response: %w", responseErr))
				}
			}
		}
	}
}

func runCLIInitForTest(
	t *testing.T,
	payload any,
	loadedConfig config.RslintConfig,
	wantAcknowledgedOutputRequests int,
	wantShutdown bool,
) (int, string) {
	t.Helper()

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

	peerDone := make(chan cliTestPeerResult, 1)
	go func() {
		peerDone <- serveCLITestPeer(stdoutR, stdinW, loadedConfig)
	}()

	codeCh := make(chan int, 1)
	go func() { codeCh <- runCLI(nil) }()

	initRequest, err := ipc.NewMessage(kindInit, cliTestInitRequestID, payload)
	if err != nil {
		t.Fatalf("build init request: %v", err)
	}
	if err := ipc.WriteFrame(stdinW, initRequest); err != nil {
		t.Fatalf("write init request: %v", err)
	}

	var code int
	select {
	case code = <-codeCh:
	case <-time.After(60 * time.Second):
		t.Fatal("runCLI did not return within 60s")
	}

	// runCLI has stopped writing frames. Closing its test stdout publishes EOF
	// to the ordered peer so we can join it before inspecting captured output.
	// The join, rather than a sleep or retry, is the happens-before edge that
	// makes the capture complete.
	os.Stdin, os.Stdout = origStdin, origStdout
	if err := stdoutW.Close(); err != nil {
		t.Fatalf("close CLI stdout: %v", err)
	}

	var peerResult cliTestPeerResult
	select {
	case peerResult = <-peerDone:
	case <-time.After(5 * time.Second):
		t.Fatal("test peer did not finish after CLI stdout closed")
	}
	if err := stdinW.Close(); err != nil {
		t.Fatalf("close test peer output: %v", err)
	}
	if peerResult.err != nil {
		t.Fatalf("test peer protocol error: %v", peerResult.err)
	}
	if peerResult.initReplies != 1 {
		t.Fatalf("init replies = %d, want 1", peerResult.initReplies)
	}
	if peerResult.acknowledgedOutputRequests != wantAcknowledgedOutputRequests {
		t.Fatalf(
			"acknowledged output requests = %d, want %d",
			peerResult.acknowledgedOutputRequests,
			wantAcknowledgedOutputRequests,
		)
	}
	wantShutdownRequests := 0
	if wantShutdown {
		wantShutdownRequests = 1
	}
	if peerResult.shutdownRequests != wantShutdownRequests {
		t.Fatalf("shutdown requests = %d, want %d", peerResult.shutdownRequests, wantShutdownRequests)
	}
	return code, peerResult.output
}

func TestWriteConfigDiscoveryFailures(t *testing.T) {
	var output strings.Builder
	writeConfigDiscoveryFailures(&output, []discovery.ConfigFailure{
		{Path: "/repo/broken-a/rslint.config.mjs", Message: "unexpected token"},
		{Path: "/repo/broken-b/rslint.config.ts", Message: "module not found"},
	})

	want := "Warning: skipping config /repo/broken-a/rslint.config.mjs: unexpected token\n" +
		"Warning: skipping config /repo/broken-b/rslint.config.ts: module not found\n"
	if got := output.String(); got != want {
		t.Fatalf("config discovery warnings = %q, want %q", got, want)
	}
}
