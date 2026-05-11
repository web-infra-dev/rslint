// IPC CLI entry: every user-facing CLI invocation reaches the Go binary
// here. The binary is an internal npm artifact (users always go through
// `rslint.cjs → cli.ts → engine.ts`), so we can demand a single, uniform
// entry contract: the Node parent owns stdin/stdout, drives the
// handshake, and delivers the user's intent in one `init` message.
//
// Topology:
//
//	┌─ Node parent (cli.ts → engine.ts) ─────────────────┐
//	│  spawn(binary, ...userArgs)                        │
//	│  stdin/stdout = bidirectional IPC frames           │
//	│  stderr       = inherited                          │
//	│  sends `init` first; awaits `response{ok:true}`    │
//	│  answers reverse `lintEslintPlugin` requests       │
//	│    via WorkerPool                                  │
//	│  forwards `output` notifications to its real stdout│
//	└────────────────────────────────────────────────────┘
//
// Pipeline:
//
//  1. parseLintFlags                  — capture --help / --init / --fix /
//                                       --format / --start-time and
//                                       positional files from the user-
//                                       forwarded args.
//  2. start IPC service               — bind to original os.Stdin/Stdout.
//  3. wait `init` or signal           — peer authoritatively declares
//                                       WorkingDirectory, Configs,
//                                       EslintPluginEntries, runtime
//                                       hints. Init is mandatory even
//                                       for `--help` / `--init` so the
//                                       protocol stays uniform.
//  4. redirect stdout                 — every plain-text write goes
//                                       through `output` notifications
//                                       so it doesn't corrupt the IPC
//                                       frame stream sharing the fd.
//  5. dispatch on intent              — `--help` prints usage and returns
//                                       0; `--init` runs the shared
//                                       executeLintPipeline (which
//                                       handles InitDefaultConfig
//                                       internally) without stdin
//                                       reassignment.
//  6. (lint flow only) synthesize stdin
//                                     — os.Stdin = pipe carrying the
//                                       synthesized config payload (the
//                                       same shape parseConfigPayload
//                                       expects in cmd.go).
//  7. build compat dispatcher        — closure over
//                                       bs.SendRequest("lintEslintPlugin");
//                                       passed by parameter to
//                                       executeLintPipeline (no globals).
//  8. executeLintPipeline             — the shared helper covering
//                                       multi-config / gap-file /
//                                       fix-loop / output formatting.
//  9. drain output, send `shutdown`,  — peer acknowledges, closes its WorkerPool;
//     close service, exit               we propagate the lint exit code.
//
// Exit codes:
//
//	0   — clean run, no lint errors / successful --init / --help
//	1   — lint errors / config errors
//	2   — runner failure (peer disconnect, IPC failure, compat dispatch failure)
//	130 — interrupted by SIGINT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// initPayload mirrors the IPC `init` message sent by `engine.ts`. Field
// shape is the wire contract — keep in sync with packages/rslint/src/engine.ts.
type initPayload struct {
	// Configs: array of `{configDirectory, entries}`-shaped JSON objects.
	// Re-marshaled byte-for-byte into the synthesized stdin payload so
	// parseConfigPayload (cmd.go::executeLintPipeline) parses it the
	// same way regardless of whether configs came from the IPC init
	// message or (in legacy direct-invocation paths) stdin. Empty/nil
	// means "no JS config — let the binary load JSON config from disk
	// via LoadConfigurationWithFallback (ConfigStdin=false branch)".
	Configs []json.RawMessage `json:"configs,omitempty"`

	// EslintPluginEntries: registered as placeholder rules so the registry
	// has matching names when the config references them. Empty when no
	// config carries `eslintPlugins`.
	EslintPluginEntries []rslintconfig.EslintPluginEntry `json:"eslintPluginEntries,omitempty"`

	// Runtime: out-of-band switches that don't have a 1:1 user flag.
	Runtime runtimePayload `json:"runtime,omitempty"`

	// User-flag mirrors. cli.ts forwards user flags both as Go args
	// (preserving `flag.Parse` compatibility) AND through the init payload
	// for the few we want to authoritatively control from Node (notably
	// the working directory and the discovered positional file/dir set).
	Files            []string `json:"files,omitempty"`
	WorkingDirectory string   `json:"workingDirectory,omitempty"`
	Format           string   `json:"format,omitempty"`
	FixMode          bool     `json:"fixMode,omitempty"`
}

// runtimePayload is the IPC-bound subset of the runner's `runtime` knobs.
// Out-of-band overrides for switches that don't have (or shouldn't be
// duplicated as) a 1:1 user flag — currently the force-color TTY hint and
// the single-threaded toggle.
//
// Anything that's purely a JS-side tuning knob (worker-pool timeouts) or
// already arrives via flag.Parse (--force-color, --singleThreaded
// user-explicit) lives elsewhere and is intentionally
// not sent here. Adding a new field is a wire-protocol change — keep
// this small.
type runtimePayload struct {
	ForceColor     bool `json:"forceColor,omitempty"`
	SingleThreaded bool `json:"singleThreaded,omitempty"`
}

type logNotification struct {
	Level  string `json:"level"`
	Source string `json:"source"`
	Text   string `json:"text"`
}

// runCLIState carries the init-handshake outcome from the inbound handler
// goroutine to the runCLI main goroutine.
type runCLIState struct {
	payloadCh chan *initPayload // 1-buffered; receives at most one payload
	once      sync.Once
	signalled atomic.Bool // set on SIGINT/SIGTERM
}

// runCLIInboundHandler dispatches the only inbound request the CLI shim
// expects: `init`. Anything else is rejected with a transparent error so
// the peer can surface it.
type runCLIInboundHandler struct{ state *runCLIState }

func (h *runCLIInboundHandler) Handle(_ context.Context, msg *api.Message) (interface{}, error) {
	switch msg.Kind {
	case api.KindInit:
		var p initPayload
		if err := api.DecodeData(msg.Data, &p); err != nil {
			return nil, fmt.Errorf("decode init: %w", err)
		}
		// Exactly-once delivery — engine.ts sends one init; defensive
		// against a buggy peer.
		h.state.once.Do(func() {
			h.state.payloadCh <- &p
		})
		return map[string]interface{}{"ok": true}, nil
	default:
		return nil, fmt.Errorf("ipc-cli: unsupported inbound kind %q", msg.Kind)
	}
}

// runCLI is the unified IPC mode entry point. Returns the exit code that
// the process should propagate.
//
// `args` is the post-mode-prefix argv (i.e. main.go forwards os.Args[1:]
// after stripping nothing — runCLI is the default branch). It is parsed
// through Go's global flag set for the user's --help / --init / --fix
// / --format / --start-time flags.
func runCLI(args []string) int {
	// Parse user-supplied flags eagerly — cli.ts forwards the full user
	// argv through to the binary, so this picks up --help, --init,
	// --fix, --format, --start-time, etc.
	baseArgs, help, parseFatal := parseLintFlags(args)
	if parseFatal != 0 {
		// parseLintFlags already printed the error to stderr.
		return parseFatal
	}

	// Signals we care about:
	//
	//   SIGINT  — user pressed Ctrl-C in the shell. The most common
	//             interruption path.
	//   SIGTERM — supervisor / engine.ts safeKillGo / launchctl etc.
	//             requesting graceful exit.
	//   SIGHUP  — controlling terminal disconnected (user closed the
	//             tab, ssh session ended). Without an explicit handler,
	//             the OS default action terminates the process abruptly
	//             — IPC peer never gets a clean handshake, profiling
	//             output (--trace, --cpuprof) is truncated.
	//
	// On Windows, SIGHUP is defined but never delivered naturally by
	// the OS; registering it via signal.Notify is a no-op there. The
	// same code is therefore portable — no GOOS guard needed.
	//
	// Two registrations:
	//
	//   sigChInit:  consumed by the init-handshake select for a fast
	//               abort BEFORE we have a lint context to cancel.
	//   lintCtx:    populated by signal.NotifyContext so the lint
	//               file-loop's per-file boundary observes cancellation.
	//
	// We keep them separate because the init-select consumes its channel
	// (a ctx-only model would have to poll Done() in the select, which
	// is fine but more verbose).
	sigChInit := make(chan os.Signal, 1)
	signal.Notify(sigChInit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sigChInit)
	lintCtx, lintCtxStop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer lintCtxStop()

	// Capture the original stdin/stdout BEFORE BidirectionalService binds
	// to them. After init we may reassign os.Stdin / os.Stdout to in-process
	// pipes (lint flow only) so executeLintPipeline can read its
	// synthesized config bytes and forward print output as IPC notifications.
	origStdin, origStdout := os.Stdin, os.Stdout

	// IPC service over the original (real) stdin/stdout pair. The service
	// holds those references internally even after we reassign the global
	// vars below.
	bs := api.NewBidirectionalService(origStdin, origStdout)

	state := &runCLIState{payloadCh: make(chan *initPayload, 1)}

	// Drain lintCtx.Done() onto state.signalled so the final exit-code
	// mapping (returns 130 when set) still works even when the signal
	// arrives AFTER the init-handshake select has released sigChInit.
	// We can't tag state.signalled inside signal.NotifyContext (that's
	// a stdlib function), so we observe ctx.Done() here.
	sigDrainStop := make(chan struct{})
	go func() {
		select {
		case <-lintCtx.Done():
			state.signalled.Store(true)
		case <-sigDrainStop:
		}
	}()
	defer close(sigDrainStop)
	bs.SetInboundHandler(&runCLIInboundHandler{state: state})

	// Inbound `log` notification from the runner: forward to our stderr.
	bs.RegisterNotification(api.KindLog, func(_ context.Context, msg *api.Message) {
		var rec logNotification
		if err := api.DecodeData(msg.Data, &rec); err == nil {
			fmt.Fprintf(os.Stderr, "[runner-log:%s:%s] %s\n", rec.Source, rec.Level, rec.Text)
		}
	})

	bs.Start()

	// Disconnect watcher: if the reader goroutine inside BS exits (peer
	// closed its stdin) before init arrives, the init-select would
	// otherwise hang up to `initTimeout` waiting for a payload that
	// will never come. Push a nil into payloadCh so the select sees a
	// disconnect immediately. once.Do guards against double-delivery
	// in the normal-init path (handler already populated the channel).
	go func() {
		bs.Wait()
		state.once.Do(func() {
			select {
			case state.payloadCh <- nil:
			default:
			}
		})
	}()

	// Wait for init handshake or signal. Mandatory regardless of intent —
	// the protocol is uniform; --help and --init both wait for init too.
	// A 60s ceiling guards against the peer crashing or being killed
	// before sending init: without a deadline we'd hang forever on the
	// payloadCh receive (signal.Notify alone doesn't fire on parent EOF).
	const initTimeout = 60 * time.Second
	initTimer := time.NewTimer(initTimeout)
	defer initTimer.Stop()

	var payload *initPayload
	select {
	case payload = <-state.payloadCh:
	case <-sigChInit:
		state.signalled.Store(true)
		_ = bs.Close()
		return 130
	case <-initTimer.C:
		fmt.Fprintf(os.Stderr, "ipc-cli: timed out waiting for init message after %s\n", initTimeout)
		_ = bs.Close()
		return 2
	}
	if payload == nil {
		// Channel closed without payload — peer disconnected.
		_ = bs.Close()
		return 2
	}

	// Apply payload to baseArgs. Working directory, plugin entries, and
	// the positional file set are payload-authoritative; everything else
	// flows from flag.Parse and the payload may additionally override a
	// short list of switches the Node parent prefers to drive (Format,
	// FixMode, ForceColor, SingleThreaded). Each override below is
	// intentionally additive — payload values supplement flag values
	// rather than replacing them, so a flag-true cannot be turned off
	// by an absent payload field.
	if payload.WorkingDirectory != "" {
		// Hard-fail on chdir failure: every downstream path resolution
		// (config discovery, file scope, gap-file matching) anchors at
		// process cwd. Silently continuing in the wrong directory would
		// produce confusing wrong-files-linted output.
		if err := os.Chdir(payload.WorkingDirectory); err != nil {
			fmt.Fprintf(os.Stderr,
				"ipc-cli: chdir to %q failed: %v\n",
				payload.WorkingDirectory, err)
			_ = bs.Close()
			return 2
		}
	}
	rslintconfig.RegisterEslintPluginRules(payload.EslintPluginEntries)

	if len(payload.Files) > 0 {
		baseArgs.AllowFiles, baseArgs.AllowDirs = classifyPaths(payload.Files)
	}
	if payload.Format != "" {
		baseArgs.Format = payload.Format
	}
	if payload.FixMode {
		baseArgs.Fix = true
	}
	if payload.Runtime.ForceColor {
		baseArgs.ForceColor = true
	}
	if payload.Runtime.SingleThreaded {
		baseArgs.SingleThreaded = true
	}
	// ConfigStdin reflects whether the peer has supplied us with configs
	// in the payload (JS/TS config path) or expects the binary to load
	// JSON config from disk itself (ConfigStdin=false).
	baseArgs.ConfigStdin = len(payload.Configs) > 0

	// ── stdout redirect (mandatory for every intent) ─────────────────
	//
	// `--help`, `--init`, and the lint flow all need this redirect: the
	// IPC peer holds the real stdout for frame I/O, so any plain-text
	// write ("Created rslint.config.*", usage banner, lint output)
	// would corrupt the IPC stream if it shared the fd. `--init` only
	// skips the *stdin* reassignment further down (no config payload to
	// read), but its "Created…" line still needs the stdout pipe.
	// Redirect every plain-text stdout write through `output`
	// notifications, which the peer concatenates into its real stdout.
	// stderr is untouched — the binary's stderr is inherited end-to-end
	// via engine.ts's spawn options.
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		_ = bs.Close()
		fmt.Fprintf(os.Stderr, "ipc-cli: stdout pipe: %v\n", err)
		return 2
	}
	stdoutDrainDone := make(chan struct{})
	go drainStdoutToIPC(stdoutR, bs, stdoutDrainDone)
	os.Stdout = stdoutW
	// Helper to flush + restore stdout cleanly before the shutdown handshake.
	finalizeStdout := func() {
		os.Stdout = origStdout
		_ = stdoutW.Close()
		<-stdoutDrainDone
		_ = stdoutR.Close()
	}

	// ── Intent dispatch ──────────────────────────────────────────────

	if help {
		// Usage prints to stderr (inherited), so this works without
		// stdout reassignment, but we keep the redirect active for
		// uniformity.
		fmt.Fprint(os.Stderr, usage)
		finalizeStdout()
		shutdownPeer(bs, state)
		_ = bs.Close()
		return 0
	}

	if baseArgs.Init {
		// executeLintPipeline handles `args.Init` internally:
		// InitDefaultConfig + "Created rslint.config.*" + return 0.
		// The created-file announcement reaches stdout (now the IPC
		// pipe → engine.ts → user terminal). No *stdin* reassignment is
		// needed — ConfigStdin is implicitly false for --init, so the
		// pipeline never tries to read a config payload from stdin.
		//
		// `--init` never needs a compat dispatcher (it doesn't lint
		// anything), so pass nil. Also doesn't honor cancellation —
		// the operation is fast and atomic.
		exitCode := executeLintPipeline(context.Background(), baseArgs, nil)
		finalizeStdout()
		shutdownPeer(bs, state)
		_ = bs.Close()
		if state.signalled.Load() {
			return 130
		}
		return exitCode
	}

	// ── Full lint flow ────────────────────────────────────────────────

	// Synthesize stdin only when configs were provided in the payload.
	// JSON-config path leaves baseArgs.ConfigStdin=false, in which case
	// executeLintPipeline calls LoadConfigurationWithFallback itself.
	if baseArgs.ConfigStdin {
		stdinR, stdinW, err := os.Pipe()
		if err != nil {
			finalizeStdout()
			_ = bs.Close()
			fmt.Fprintf(os.Stderr, "ipc-cli: stdin pipe: %v\n", err)
			return 2
		}
		cfgBytes, mErr := json.Marshal(map[string]interface{}{
			"configs": payload.Configs,
		})
		if mErr != nil {
			finalizeStdout()
			_ = bs.Close()
			_ = stdinR.Close()
			_ = stdinW.Close()
			fmt.Fprintf(os.Stderr, "ipc-cli: marshal config payload: %v\n", mErr)
			return 2
		}
		writerDone := startSynthStdinWriter(lintCtx, stdinW, cfgBytes)
		os.Stdin = stdinR
		defer func() {
			os.Stdin = origStdin
			_ = stdinR.Close()
			<-writerDone // ensure no dangling writer goroutine on return
		}()
	}

	// Build the dispatcher inline so it captures `bs` without leaking to
	// a package-level global. The pipeline's per-Program compat batch
	// reaches this closure for any rule whose IsEslintPluginRule=true.
	// nil dispatcher = compat rules silently skipped (same default the
	// linter uses for paths without plugin support).
	dispatcher := makeIPCDispatcher(bs)

	// ── Run the shared lint pipeline ─────────────────────────────────
	// lintCtx is cancelled by SIGINT/SIGTERM via signal.NotifyContext
	// (set up above). RunLinter's per-file boundary check honors it.
	exitCode := executeLintPipeline(lintCtx, baseArgs, dispatcher)

	finalizeStdout()
	shutdownPeer(bs, state)
	_ = bs.Close()

	if state.signalled.Load() {
		return 130
	}
	return exitCode
}

// shutdownPeer best-effort signals the peer that we're done; on success
// the peer drains its WorkerPool and exits cleanly. Caller must still
// bs.Close() afterwards.
//
// If a SIGINT/SIGTERM already fired (state.signalled), we skip the
// request: the peer will see its own ctx.Done() via stdin disconnect
// anyway, and pushing another frame races the closing pipe. Previously
// the no-op contract was implicit (SendRequest's internal `closed`
// check kicked in once bs.Close had run), but that left a window where
// state.signalled was true but bs.Close hadn't yet been called — we'd
// push the frame, the peer might process it, and the signal cleanup
// would race the shutdown ack. Explicit check closes the window.
func shutdownPeer(bs *api.BidirectionalService, state *runCLIState) {
	if state != nil && state.signalled.Load() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = bs.SendRequest(ctx, api.KindShutdown, struct{}{})
}

// makeIPCDispatcher returns a CompatBatchHandler that reverse-RPCs the
// batch to the Node WorkerPool over BidirectionalService.SendRequest.
//
// On any IPC failure (broken pipe, malformed response, peer-closed)
// the dispatcher returns a non-nil error so the linter marks the
// program's compat output as failed without dropping native
// diagnostics that already streamed through OnDiagnostic.
//
// Cancellation: caller's `ctx` (the lint-level SIGINT context) is
// passed straight through to SendRequest; if the user hits Ctrl-C the
// ctx cancels and SendRequest returns immediately. The Node-side
// WorkerPool already enforces a per-task timeout (default 30s) inside
// each worker_thread, and terminates+respawns on hang, so a stuck
// plugin rule never blocks the dispatcher indefinitely either. There
// is no separate Go-side per-batch deadline — adding one would only
// catch "Node host process alive but its event loop wedged", which
// is extremely rare given audited shutdown / cancel paths, and the
// trade-off (legitimate large-monorepo batches getting killed) was
// not worth it.
func makeIPCDispatcher(bs *api.BidirectionalService) linter.CompatBatchHandler {
	return func(ctx context.Context, batch linter.CompatBatch) ([]linter.CompatFileResult, error) {
		resp, err := bs.SendRequest(ctx, api.KindLintEslintPlugin, batch)
		if err != nil {
			return nil, fmt.Errorf("lintEslintPlugin IPC: %w", err)
		}

		// engine.ts replies with `{ results: CompatFileResult[] }`.
		var decoded struct {
			Results []linter.CompatFileResult `json:"results"`
		}
		if err := api.DecodeData(resp.Data, &decoded); err != nil {
			return nil, fmt.Errorf("decode lintEslintPlugin response: %w", err)
		}
		return decoded.Results, nil
	}
}

// classifyPaths splits a path slice into (files, dirs) by stat'ing each
// entry. Mirrors the positional-arg classification done in
// parseLintFlags so the IPC and CLI-flag entry paths produce identical
// FileScope.{Files,Dirs} downstream.
//
// Two normalization steps, both required to match parseLintFlags:
//
//  1. filepath.Abs — turn relative paths (the common case from cli.ts
//     forwarding `process.argv`-shaped positionals) into absolute. The
//     rest of the pipeline (FileScope, FindNearestConfig via
//     tspath.StartsWithDirectory, programFiles map lookups) keys off
//     absolute paths; a relative path here would silently miss the
//     gap-file detection path and cause the file to be skipped entirely.
//
//  2. tspath.NormalizePath — collapse `./`, trailing slashes etc. so
//     two callers passing the same logical path produce byte-equal
//     strings.
//
// An Abs failure on any entry is rare (current cwd inaccessible) but
// not silently swallowed: the entry is skipped with a stderr warning so
// the user notices instead of getting a "file not in project" error
// later with no explanation.
func classifyPaths(paths []string) (files []string, dirs []string) {
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ipc-cli: filepath.Abs(%q) failed: %v\n", p, err)
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

// startSynthStdinWriter feeds `data` into `w` from a background
// goroutine and returns a channel that's closed when the writer is
// fully joined. Two failure modes are bounded:
//
//  1. The reader (executeLintPipeline / parseConfigPayload) bails
//     before consuming all bytes. Without the ctx watch, a `data`
//     payload larger than the kernel pipe buffer (~64 KiB on Linux)
//     would block the Write forever and leak the goroutine.
//
//  2. ctx fires during the write (SIGINT). Closing `w` from outside
//     unblocks the Write with an error, the goroutine joins, and
//     callers waiting on the returned channel see it close.
//
// Returns immediately (does not block on first byte) — caller can
// install os.Stdin = pipe and proceed; the writer runs concurrently.
func startSynthStdinWriter(ctx context.Context, w io.WriteCloser, data []byte) <-chan struct{} {
	done := make(chan struct{})
	// sync.Once wraps Close so the ctx-cancel branch and the deferred
	// cleanup don't both call Close on `w`. *os.File.Close is
	// idempotent (returns ErrClosed on the second call, which we
	// ignore), but a future change of `w`'s underlying type could
	// break that assumption — e.g. some io.WriteCloser
	// implementations panic on second Close. The Once makes the
	// invariant explicit at the call site.
	var closeOnce sync.Once
	closeW := func() { closeOnce.Do(func() { _ = w.Close() }) }
	go func() {
		defer close(done)
		defer closeW()
		writeDone := make(chan struct{})
		go func() {
			_, _ = w.Write(data)
			close(writeDone)
		}()
		select {
		case <-writeDone:
		case <-ctx.Done():
			// Closing the writer here interrupts the in-flight Write
			// with an error (which the inner goroutine discards). Join
			// it before returning to avoid releasing the goroutine
			// while it still references our local closures. The
			// deferred closeW() above no-ops because closeOnce already
			// fired here.
			closeW()
			<-writeDone
		}
	}()
	return done
}

// stdoutDrainMinFlushBytes is the soft floor for batching IPC `output`
// frames. Reads smaller than this are buffered and combined with later
// reads; reads at-or-above flush immediately. Sized to amortize one
// JSON-frame's fixed overhead (~50 bytes header + per-frame Node-side
// dispatch) over a meaningful chunk of output, without holding so much
// that interactive runs stall.
const stdoutDrainMinFlushBytes = 8 * 1024

// drainStdoutToIPC consumes the stdout-redirect pipe and forwards bytes
// to the IPC peer as `output` notifications. Two correctness concerns
// drive the implementation:
//
//  1. **UTF-8 boundary safety.** `os.Pipe.Read` may return chunks that
//     end in the middle of a multi-byte character; `string(buf[:n])`
//     would then encode replacement-rune bytes (`�`) and the peer
//     would see corrupted non-ASCII output. We hold back any incomplete
//     trailing sequence (1-3 bytes) and prepend it to the next read.
//
//  2. **Frame-count overhead.** Each `bs.SendNotification` is one JSON
//     frame on the IPC stream. For 100K+ diagnostic lines that's 100K
//     frames of CPU + syscall overhead. We coalesce up to
//     stdoutDrainMinFlushBytes worth of bytes per frame; the close path
//     always flushes whatever remains.
//
// On the first SendNotification failure (peer closed its end), we flip
// into discard mode — keep draining `r` so the lint pipeline's writes
// don't block on a full pipe, but drop the bytes silently.
func drainStdoutToIPC(r io.Reader, bs *api.BidirectionalService, done chan<- struct{}) {
	defer close(done)
	buf := make([]byte, 4096)
	var leftover []byte // UTF-8 incomplete tail held back from last read
	var pending []byte  // bytes ready to send, waiting for batch threshold
	discard := false

	flush := func() {
		if len(pending) == 0 || discard {
			// In discard mode we'd accumulate the underlying backing
			// array forever (`pending[:0]` only resets length, not
			// capacity). Drop the reference so future appends start
			// from zero — costs one heap re-alloc the next time we
			// (uselessly) buffer, but lets the GC reclaim everything
			// we've buffered so far. Discard mode fires only when the
			// IPC peer has gone away, so subsequent writes are
			// throwaway anyway.
			pending = nil
			return
		}
		if err := bs.SendNotification(api.KindOutput, map[string]interface{}{
			"stream": "stdout",
			"text":   string(pending),
		}); err != nil {
			discard = true
		}
		pending = pending[:0]
	}

	for {
		n, readErr := r.Read(buf)
		if n > 0 || readErr != nil {
			// Combine leftover (from last iteration's incomplete UTF-8 tail)
			// with this iteration's fresh bytes. The :len:len trick on
			// leftover bounds its capacity so append won't alias into a
			// shared backing array.
			var combined []byte
			if len(leftover) > 0 {
				combined = append(leftover[:len(leftover):len(leftover)], buf[:n]...)
				leftover = nil
			} else if n > 0 {
				combined = buf[:n]
			}

			if readErr != nil {
				// Final read: flush everything, including any genuinely
				// invalid bytes at the tail. Better to send them and let the
				// peer render U+FFFD than to silently drop them.
				pending = append(pending, combined...)
			} else if len(combined) > 0 {
				good, tail := splitAtUTF8Boundary(combined)
				pending = append(pending, good...)
				if len(tail) > 0 {
					// Copy because next iteration's Read will overwrite buf.
					leftover = make([]byte, len(tail))
					copy(leftover, tail)
				}
			}

			// Flush either at EOF/error OR once pending crosses the batch
			// threshold. Holding bytes across a Read block is acceptable
			// here because the lint pipeline produces output in clusters
			// (per-file diagnostic blocks, formatter writes); we'd rather
			// pay one big frame than dozens of small ones.
			if readErr != nil || len(pending) >= stdoutDrainMinFlushBytes {
				flush()
			}
		}
		if readErr != nil {
			return
		}
	}
}

// splitAtUTF8Boundary returns the longest prefix of buf that ends at a
// complete UTF-8 character boundary, plus the trailing 1-3 bytes (if
// any) of an unfinished multi-byte character. Caller prepends
// `incomplete` to the next read so the character isn't split on the
// wire.
//
// On genuinely invalid UTF-8 at the tail (no lead byte in the last 4
// bytes, or an invalid 5+-byte lead byte), returns (buf, nil) — we
// don't try to recover from real corruption; the peer will render
// U+FFFD at the boundary, same as before the fix. This is correct for
// the typical case (chunk boundary in valid UTF-8) without inventing
// behavior for impossible inputs.
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
			// 11111xxx — invalid lead byte. Don't try to recover.
			return buf, nil
		}
		avail := n - i
		if avail >= width {
			return buf, nil // last character is complete
		}
		return buf[:i], buf[i:n]
	}
	// Entire 4-byte tail is continuation bytes (corrupt stream) or buf
	// is empty. Either way, no incomplete-character recovery is possible.
	return buf, nil
}
