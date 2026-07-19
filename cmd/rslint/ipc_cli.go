//go:build !js

// ipc_cli.go — the IPC CLI entry. Every user-facing rslint CLI invocation
// reaches the Go binary here: the Node parent (packages/rslint cli.ts →
// engine.ts) spawns this binary, owns its stdin/stdout as a bidirectional
// IPC frame stream (internal/ipc.Channel), drives an `init` handshake,
// forwards lint output as `output` frames, and acks `shutdown`.
//
// Topology:
//
//	┌─ Node parent (cli.ts → engine.ts) ───────────────────────┐
//	│  spawn(binary, ...userArgs)                               │
//	│  stdin/stdout = bidirectional IPC frames (ipc.Channel)    │
//	│  stderr       = inherited                                 │
//	│  sends `init`; awaits `response{ok:true}`                 │
//	│  forwards `output` notifications to its real stdout       │
//	└───────────────────────────────────────────────────────────┘
//
// Pipeline: parseLintFlags → start ipc.Channel on the real stdin/stdout →
// wait `init` (or signal) → redirect stdout through `output` frames →
// dispatch on intent (--help / --init / lint) → run the shared
// executeLintPipeline (using either a typed Go-discovered catalog or the
// native JSON/JSONC loader) → drain output, send `shutdown`, exit.
//
// Exit codes: 0 clean · 1 lint/config errors · 2 IPC failure (peer
// disconnect, init/transport error) · 130 interrupted.
//
// Excluded from the js/wasm build: signal handling needs syscall.SIGHUP,
// undefined there, and there is no Node IPC parent. ipc_cli_js.go provides a
// native fallback runCLI for that target.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/output"
)

// Application-level IPC message kinds for the CLI ⇆ Node engine protocol.
// The transport (ipc.Channel) owns only response/error/handshake/exit; the
// kinds below are declared here and travel through the same opaque envelope.
const (
	kindInit            ipc.MessageKind = "init"            // Node → Go: handshake payload
	kindShutdown        ipc.MessageKind = "shutdown"        // Go → Node: lint done
	kindOutput          ipc.MessageKind = "output"          // Go → Node: forwarded stdout text
	kindPluginLint      ipc.MessageKind = "pluginLint"      // Go → Node: run ESLint-plugin rules in a worker
	kindLoadConfigs     ipc.MessageKind = "loadConfigs"     // Go → Node: evaluate one config frontier
	kindActivateConfigs ipc.MessageKind = "activateConfigs" // Go → Node: prepare the effective config/plugin set
)

// initPayload mirrors the IPC `init` message sent by engine.ts. Field shape
// is the wire contract — keep in sync with packages/rslint/src/engine.ts.
type initPayload struct {
	// Runtime: out-of-band switches without a 1:1 user flag.
	Runtime runtimePayload `json:"runtime,omitempty"`

	// User-flag mirrors the Node parent authoritatively drives: the working
	// directory and the discovered positional file/dir set, plus a few
	// switches it prefers to control. User flags also arrive as Go args via
	// parseLintFlags; payload values supplement (never disable) them.
	Files            []string `json:"files,omitempty"`
	WorkingDirectory string   `json:"workingDirectory,omitempty"`
	Format           string   `json:"format,omitempty"`
	FixMode          bool     `json:"fixMode,omitempty"`

	// ConfigDiscovery asks Go to own JS/TS config discovery. It is ignored for
	// help/init and disabled for the native JSON/JSONC configuration path.
	ConfigDiscovery *configDiscoveryPayload `json:"configDiscovery,omitempty"`
}

type configDiscoveryPayload struct {
	Mode               string `json:"mode"`
	ExplicitConfigPath string `json:"explicitConfigPath,omitempty"`
}

type ipcConfigModuleLoader struct {
	channel *ipc.Channel
}

func (loader *ipcConfigModuleLoader) LoadConfigs(ctx context.Context, request discovery.ConfigLoadBatchRequest) (discovery.ConfigLoadBatchResponse, error) {
	msg, err := loader.channel.SendRequest(ctx, kindLoadConfigs, request)
	if err != nil {
		return discovery.ConfigLoadBatchResponse{}, err
	}
	var response discovery.ConfigLoadBatchResponse
	if err := msg.Decode(&response); err != nil {
		return discovery.ConfigLoadBatchResponse{}, fmt.Errorf("decode loadConfigs response: %w", err)
	}
	return response, nil
}

func (loader *ipcConfigModuleLoader) ActivateConfigs(ctx context.Context, request discovery.ConfigActivationRequest) (discovery.ConfigActivationResponse, error) {
	msg, err := loader.channel.SendRequest(ctx, kindActivateConfigs, request)
	if err != nil {
		return discovery.ConfigActivationResponse{}, err
	}
	var response discovery.ConfigActivationResponse
	if err := msg.Decode(&response); err != nil {
		return discovery.ConfigActivationResponse{}, fmt.Errorf("decode activateConfigs response: %w", err)
	}
	return response, nil
}

// runtimePayload is the IPC-bound subset of runtime knobs that don't have (or
// shouldn't duplicate) a 1:1 user flag. Adding a field is a wire change.
type runtimePayload struct {
	// StdoutIsTTY reports whether the peer's real stdout — the terminal the
	// forwarded lint output lands on — is a TTY. The Go process cannot
	// observe this itself (its own stdout is the IPC pipe). Absent (false)
	// when unavailable (for example in the wasm fallback), which degrades to
	// colorless output.
	StdoutIsTTY    bool `json:"stdoutIsTTY,omitempty"`
	SingleThreaded bool `json:"singleThreaded,omitempty"`
}

// runCLIState carries the init-handshake outcome from the inbound handler
// (which runs on the channel's read-loop goroutine) to the runCLI main
// goroutine, and tracks whether a termination signal fired.
type runCLIState struct {
	payloadCh chan *initPayload // 1-buffered; receives at most one payload
	once      sync.Once
	signalled atomic.Bool // set on SIGINT/SIGTERM/SIGHUP
}

// runCLI is the unified IPC mode entry point — the default for every
// user-facing CLI run. Returns the process exit code.
func runCLI(args []string) int {
	baseArgs, help, parseFatal := parseLintFlags(args)
	if parseFatal != 0 {
		return parseFatal // parseLintFlags already printed to stderr
	}

	// Two signal registrations: sigChInit aborts the init-handshake select
	// before we have a lint context; lintCtx (NotifyContext) cancels the lint
	// file-loop at its per-file boundary. SIGHUP guards against a closed
	// controlling terminal truncating profiling output. On Windows SIGHUP is
	// defined but never delivered, so this stays portable.
	sigChInit := make(chan os.Signal, 1)
	signal.Notify(sigChInit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sigChInit)
	lintCtx, lintCtxStop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer lintCtxStop()

	// Capture stdout before redirecting lint output through `output` frames.
	// The channel keeps the original file handles for its entire lifetime.
	origStdout := os.Stdout
	ch := ipc.NewChannel(os.Stdin, origStdout)

	state := &runCLIState{payloadCh: make(chan *initPayload, 1)}

	// Mirror lintCtx cancellation onto state.signalled so the final exit-code
	// mapping (130 when set) works even when the signal arrives after the
	// init-select released sigChInit.
	sigDrainStop := make(chan struct{})
	go func() {
		select {
		case <-lintCtx.Done():
			state.signalled.Store(true)
		case <-sigDrainStop:
		}
	}()
	defer close(sigDrainStop)

	// Inbound `init` request → hand the payload to the main goroutine.
	ch.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		switch msg.Kind {
		case kindInit:
			var p initPayload
			if err := msg.Decode(&p); err != nil {
				return nil, fmt.Errorf("decode init: %w", err)
			}
			// Exactly-once: engine.ts sends one init; defensive vs a buggy peer.
			state.once.Do(func() { state.payloadCh <- &p })
			return map[string]any{"ok": true}, nil
		default:
			return nil, fmt.Errorf("rslint: unsupported inbound kind %q", msg.Kind)
		}
	})

	ch.Start()

	// Disconnect watcher: if the channel closes (peer closed its end) before
	// init arrives, push nil so the init-select sees a disconnect instead of
	// hanging until initTimeout. once.Do guards the normal-init path.
	go func() {
		<-ch.Done()
		state.once.Do(func() {
			select {
			case state.payloadCh <- nil:
			default:
			}
		})
	}()

	// Wait for the init handshake or a signal. The 60s ceiling guards against
	// a peer that crashes before sending init (signal.Notify alone won't fire
	// on parent EOF; the disconnect watcher covers a clean EOF).
	const initTimeout = 60 * time.Second
	initTimer := time.NewTimer(initTimeout)
	defer initTimer.Stop()

	var payload *initPayload
	select {
	case payload = <-state.payloadCh:
	case <-sigChInit:
		state.signalled.Store(true)
		_ = ch.Close()
		return 130
	case <-initTimer.C:
		fmt.Fprintf(os.Stderr, "rslint: timed out waiting for init after %s\n", initTimeout)
		_ = ch.Close()
		return 2
	}
	if payload == nil {
		_ = ch.Close() // peer disconnected before init
		return 2
	}

	// Apply the payload. Working directory and the positional file set are
	// payload-authoritative; the rest supplement flag values.
	if payload.WorkingDirectory != "" {
		// Hard-fail on chdir: every downstream path (config discovery, scope,
		// gap-file matching) anchors at process cwd; the wrong dir would
		// silently lint the wrong files.
		if err := os.Chdir(payload.WorkingDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "rslint: chdir to %q failed: %v\n", payload.WorkingDirectory, err)
			_ = ch.Close()
			return 2
		}
	}
	if len(payload.Files) > 0 {
		baseArgs.AllowFiles, baseArgs.AllowDirs = classifyPaths(payload.Files)
	}
	if payload.Format != "" {
		baseArgs.Format = payload.Format
	}
	if payload.FixMode {
		baseArgs.Fix = true
	}
	baseArgs.StdoutIsTTY = payload.Runtime.StdoutIsTTY
	if payload.Runtime.SingleThreaded {
		baseArgs.SingleThreaded = true
	}
	baseArgs.FS = bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	// Preserve help/init priority, then reject an invalid payload-authoritative
	// format before config discovery can execute user JavaScript.
	if !help && !baseArgs.Init {
		if _, err := output.ParseFormat(baseArgs.Format); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			_ = ch.Close()
			return 2
		}
	}

	// The host sends only discovery intent. Go scans the reachable staged
	// frontier, asks Node to evaluate exact candidates, and converts the final
	// catalog into the shared lint pipeline input. Help/init and an
	// explicit JSON --config do not need the JavaScript module host.
	if !help && !baseArgs.Init && payload.ConfigDiscovery != nil && payload.ConfigDiscovery.Mode != "disabled" {
		if err := discoverCLIConfigCatalog(lintCtx, &baseArgs, payload, ch); err != nil {
			// Discovery now includes the potentially long Go walk and reverse
			// Node loads. A signal can cancel either before the mirror goroutine
			// publishes state.signalled, so consult the source context as well.
			// Interrupts keep the CLI's documented 130 exit and are not reported
			// as ordinary configuration failures.
			interrupted := markCLIInterrupted(lintCtx, state)
			if !interrupted {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			shutdownPeer(ch, state)
			_ = ch.Close()
			if interrupted {
				return 130
			}
			return 1
		}
	}

	// ── stdout redirect (mandatory for every intent) ──────────────────────
	// The peer holds the real stdout for IPC frames, so any plain-text write
	// (usage banner, "Created rslint.config.*", lint output) would corrupt the
	// frame stream if it shared the fd. Redirect through `output`
	// notifications, which the peer concatenates into its real stdout. stderr
	// is untouched (inherited end-to-end).
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		_ = ch.Close()
		fmt.Fprintf(os.Stderr, "rslint: stdout pipe: %v\n", err)
		return 2
	}
	stdoutDrainDone := make(chan struct{})
	go drainStdoutToIPC(stdoutR, ch, stdoutDrainDone)
	os.Stdout = stdoutW
	// finalizeStdout flushes + restores stdout before the shutdown handshake.
	finalizeStdout := func() {
		os.Stdout = origStdout
		_ = stdoutW.Close()
		<-stdoutDrainDone
		_ = stdoutR.Close()
	}

	// ── Intent dispatch ────────────────────────────────────────────────────
	if help {
		// Usage prints to stderr (inherited); the redirect stays active for
		// uniformity.
		fmt.Fprint(os.Stderr, usage)
		finalizeStdout()
		shutdownPeer(ch, state)
		_ = ch.Close()
		return 0
	}

	// Reverse dispatcher: send each plugin-lint batch back to the Node host
	// over the IPC channel and decode its result. Runs concurrently with the
	// native lint pass (executeLintPipeline awaits it before output / --fix).
	dispatch := func(reqCtx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		msg, sendErr := ch.SendRequest(reqCtx, kindPluginLint, req)
		if sendErr != nil {
			return nil, sendErr
		}
		var res linter.EslintPluginLintResult
		if err := msg.Decode(&res); err != nil {
			return nil, fmt.Errorf("decode pluginLint result: %w", err)
		}
		return &res, nil
	}

	exitCode := executeLintPipeline(baseArgs, lintCtx, dispatch)

	finalizeStdout()
	shutdownPeer(ch, state)
	_ = ch.Close()

	if state.signalled.Load() {
		return 130
	}
	return exitCode
}

func markCLIInterrupted(ctx context.Context, state *runCLIState) bool {
	interrupted := state != nil && state.signalled.Load()
	if !interrupted && ctx != nil {
		interrupted = ctx.Err() != nil
	}
	if interrupted && state != nil {
		state.signalled.Store(true)
	}
	return interrupted
}

// shutdownPeer best-effort tells the peer we're done so it drains its worker
// pool and exits cleanly. Skipped if a signal already fired: the peer sees its
// own stdin disconnect anyway, and pushing another frame races the closing
// pipe.
func shutdownPeer(ch *ipc.Channel, state *runCLIState) {
	if state != nil && state.signalled.Load() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = ch.SendRequest(ctx, kindShutdown, struct{}{})
}

// classifyPaths splits a path slice into (files, dirs) by stat'ing each entry,
// mirroring parseLintFlags's positional handling (filepath.Abs +
// tspath.NormalizePath) so the IPC and flag entry paths produce identical
// FileScope downstream. An Abs failure is skipped with a stderr warning rather
// than silently dropping the path.
func classifyPaths(paths []string) (files []string, dirs []string) {
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rslint: filepath.Abs(%q) failed: %v\n", p, err)
			continue
		}
		normalized := tspath.NormalizePath(absPath)
		info, statErr := os.Stat(absPath)
		if statErr == nil && info.IsDir() {
			dirs = append(dirs, normalized)
		} else {
			files = append(files, normalized)
		}
	}
	return files, dirs
}

func discoverCLIConfigCatalog(
	ctx context.Context,
	args *lintArgs,
	payload *initPayload,
	channel *ipc.Channel,
) error {
	if args == nil || payload == nil || payload.ConfigDiscovery == nil {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory for config discovery: %w", err)
	}
	cwd = tspath.NormalizePath(cwd)
	request := discovery.ConfigDiscoveryRequest{
		CWD:            cwd,
		Directories:    append([]string(nil), args.AllowDirs...),
		ImplicitCWD:    len(args.AllowFiles) == 0 && len(args.AllowDirs) == 0,
		SingleThreaded: args.SingleThreaded,
	}
	for _, filePath := range args.AllowFiles {
		request.Files = append(request.Files, discovery.DiscoveryFile{
			Path:     filePath,
			Explicit: true,
		})
	}
	switch payload.ConfigDiscovery.Mode {
	case "auto", "":
		request.Mode = discovery.ConfigDiscoveryAuto
	case "explicit":
		request.Mode = discovery.ConfigDiscoveryExplicit
		request.ExplicitConfigPath = payload.ConfigDiscovery.ExplicitConfigPath
	default:
		return fmt.Errorf("unsupported config discovery mode %q", payload.ConfigDiscovery.Mode)
	}

	catalog, err := discovery.Build(
		ctx,
		args.FS,
		&ipcConfigModuleLoader{channel: channel},
		request,
	)
	if err != nil {
		var allFailed *discovery.AllConfigsFailedError
		if errors.As(err, &allFailed) {
			printConfigDiscoveryFailures(allFailed.Failures)
		}
		return err
	}
	printConfigDiscoveryFailures(catalog.Failures)
	args.ConfigCatalog = catalog
	return nil
}

// stdoutDrainMinFlushBytes is the soft floor for batching IPC `output` frames.
// Reads below this are buffered and combined with later reads; pending bytes
// at or above it flush immediately. Sized to amortize a JSON frame's fixed
// overhead over a meaningful chunk without stalling interactive runs.
const stdoutDrainMinFlushBytes = 8 * 1024

// drainStdoutToIPC consumes the stdout-redirect pipe and forwards bytes to the
// IPC peer as `output` notifications. Two correctness concerns drive it:
//
//  1. UTF-8 boundary safety. os.Pipe.Read may return chunks ending mid-
//     character; string(buf[:n]) would then emit replacement bytes and the
//     peer would see corrupted non-ASCII. We hold back any incomplete trailing
//     sequence (1-3 bytes) and prepend it to the next read.
//  2. Frame-count overhead. Each SendNotification is one JSON frame; for 100K+
//     diagnostic lines that's 100K frames of CPU + syscall overhead. We
//     coalesce up to stdoutDrainMinFlushBytes per frame; the close path always
//     flushes the remainder.
//
// On the first SendNotification failure (peer closed its end) we flip into
// discard mode: keep draining r so the lint pipeline's writes don't block on a
// full pipe, but drop the bytes silently.
func drainStdoutToIPC(r io.Reader, ch *ipc.Channel, done chan<- struct{}) {
	defer close(done)
	buf := make([]byte, 4096)
	var leftover []byte // UTF-8 incomplete tail held back from last read
	var pending []byte  // bytes ready to send, waiting for batch threshold
	discard := false

	flush := func() {
		if len(pending) == 0 || discard {
			// In discard mode dropping the reference lets the GC reclaim the
			// backing array (pending[:0] keeps capacity); subsequent throwaway
			// appends start from zero. Discard fires only when the peer is gone.
			pending = nil
			return
		}
		if err := ch.SendNotification(kindOutput, map[string]any{
			"stream": "stdout",
			"text":   string(pending),
		}); err != nil {
			// The transport fails SendNotification only once it has closed
			// (peer gone) — there is no transient state to retry past, so stop
			// forwarding and drop the rest. Surface it ONCE (we set discard
			// next, so flush never re-enters this branch) so a truncated run is
			// diagnosable on stderr instead of silently losing lint output.
			fmt.Fprintf(os.Stderr,
				"rslint: stopped forwarding stdout to peer (channel closed): %v\n", err)
			discard = true
		}
		pending = pending[:0]
	}

	for {
		n, readErr := r.Read(buf)
		if n > 0 || readErr != nil {
			// Combine the last iteration's incomplete UTF-8 tail with this
			// read's fresh bytes. The :len:len cap on leftover keeps append
			// from aliasing into a shared backing array.
			var combined []byte
			if len(leftover) > 0 {
				combined = append(leftover[:len(leftover):len(leftover)], buf[:n]...)
				leftover = nil
			} else if n > 0 {
				combined = buf[:n]
			}

			if readErr != nil {
				// Final read: flush everything, including any genuinely invalid
				// tail bytes — better to send them and let the peer render
				// U+FFFD than to silently drop them.
				pending = append(pending, combined...)
			} else if len(combined) > 0 {
				good, tail := splitAtUTF8Boundary(combined)
				pending = append(pending, good...)
				if len(tail) > 0 {
					// Copy: the next Read overwrites buf.
					leftover = make([]byte, len(tail))
					copy(leftover, tail)
				}
			}

			if readErr != nil || len(pending) >= stdoutDrainMinFlushBytes {
				flush()
			}
		}
		if readErr != nil {
			return
		}
	}
}

// splitAtUTF8Boundary returns the longest prefix of buf ending at a complete
// UTF-8 character boundary, plus the trailing 1-3 bytes (if any) of an
// unfinished multi-byte character the caller prepends to the next read.
//
// On genuinely invalid UTF-8 at the tail (no lead byte in the last 4 bytes, or
// an invalid 5+-byte lead), returns (buf, nil): we don't recover from real
// corruption; the peer renders U+FFFD at the boundary, same as without the fix.
func splitAtUTF8Boundary(buf []byte) (complete, incomplete []byte) {
	n := len(buf)
	maxLookback := 4
	if n < maxLookback {
		maxLookback = n
	}
	for i := n - 1; i >= n-maxLookback; i-- {
		b := buf[i]
		// Continuation byte (10xxxxxx) — keep scanning back for a lead.
		if b&0xC0 == 0x80 {
			continue
		}
		// Lead byte (or ASCII). Determine the character's expected width.
		var width int
		switch {
		case b&0x80 == 0:
			width = 1 // 0xxxxxxx — ASCII
		case b&0xE0 == 0xC0:
			width = 2 // 110xxxxx
		case b&0xF0 == 0xE0:
			width = 3 // 1110xxxx
		case b&0xF8 == 0xF0:
			width = 4 // 11110xxx
		default:
			return buf, nil // 11111xxx — invalid lead byte; don't recover
		}
		if avail := n - i; avail >= width {
			return buf, nil // last character is complete
		}
		return buf[:i], buf[i:n]
	}
	// Entire 4-byte tail is continuation bytes (corrupt) or buf is empty.
	return buf, nil
}
