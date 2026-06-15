//go:build !js

// hangdiag.go — opt-in hang diagnostics for hunting the intermittent
// windows-latest CI hang (a tiny type-aware lint occasionally never exits;
// see the disable-comments.test.ts SIGINT-after-30min symptom).
//
// Everything here is gated behind env vars and is a NO-OP unless they are
// set, so production CLI runs are unaffected. The point is: when the hang
// recurs under the debug workflow, ONE run captures *where* the Go side is
// wedged.
//
//	RSLINT_HANG_TRACE=1            timestamped lifecycle markers to stderr,
//	                              so the last marker before a hang localizes
//	                              the wedged phase (init / lint / output /
//	                              shutdown).
//	RSLINT_HANG_WATCHDOG_MS=45000 arm a watchdog: if the process has not
//	                              finished within the deadline, dump ALL
//	                              goroutine stacks and exit(hangDiagExitCode).
//	                              A normal run finishes in ~2s, so a 45s
//	                              deadline only ever fires on a true hang.
//	RSLINT_HANG_DUMP_DIR=<dir>    also write each dump to a file in <dir>
//	                              (for CI artifact upload); stderr always gets
//	                              it regardless (inherited → CI log).
//
// On-demand: while either trace or watchdog is enabled we also install a
// dump-on-signal handler (SIGQUIT on unix, SIGBREAK/Ctrl-Break on Windows —
// Windows has no SIGQUIT) so the parent or a human can extract live stacks
// from a wedged child without killing it.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// hangDiagExitCode is the exit code the watchdog forces after dumping. Distinct
// from rslint's own 0/1/2/130 so a watchdog kill is unambiguous in CI logs and
// in the Node host (engine.ts maps it straight through from the child).
const hangDiagExitCode = 99

const (
	envHangTrace    = "RSLINT_HANG_TRACE"
	envHangWatchdog = "RSLINT_HANG_WATCHDOG_MS"
	envHangDumpDir  = "RSLINT_HANG_DUMP_DIR"
)

// hangDiagStart anchors trace timestamps; set once at first use.
var hangDiagStart = time.Now()

// dumpSeq disambiguates multiple dumps from one process (watchdog + signals).
var dumpSeq atomic.Int64

func hangTraceEnabled() bool    { return os.Getenv(envHangTrace) != "" }
func hangWatchdogEnabled() bool { return os.Getenv(envHangWatchdog) != "" }
func hangDiagEnabled() bool     { return hangTraceEnabled() || hangWatchdogEnabled() }

// tracePhase prints a timestamped lifecycle marker (elapsed since process
// start + pid) when RSLINT_HANG_TRACE is set. The last marker emitted before a
// hang tells us which phase wedged.
func tracePhase(format string, args ...any) {
	if !hangTraceEnabled() {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[hangdiag +%dms pid=%d] %s\n",
		time.Since(hangDiagStart).Milliseconds(), os.Getpid(), msg)
}

// captureAllStacks returns a snapshot of every goroutine's stack, growing the
// buffer until it fits (runtime.Stack truncates silently to the buffer).
func captureAllStacks() []byte {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

// dumpAllGoroutines writes every goroutine stack to stderr (always — inherited
// by the Node parent, so it lands in the CI log) and, when RSLINT_HANG_DUMP_DIR
// is set, to a file there for artifact upload.
func dumpAllGoroutines(reason string) {
	seq := dumpSeq.Add(1)
	var b strings.Builder
	fmt.Fprintf(&b,
		"\n===== RSLINT HANG DIAG: goroutine dump #%d (pid=%d, +%dms) =====\nreason: %s\nGOMAXPROCS=%d NumGoroutine=%d\n",
		seq, os.Getpid(), time.Since(hangDiagStart).Milliseconds(),
		reason, runtime.GOMAXPROCS(0), runtime.NumGoroutine())
	b.Write(captureAllStacks())
	fmt.Fprintf(&b, "===== end goroutine dump #%d =====\n", seq)
	content := b.String()

	// stderr is the inherited fd in IPC mode (only stdout is redirected), so
	// this lands in the Node parent's stderr → the CI log.
	fmt.Fprint(os.Stderr, content)

	if dir := os.Getenv(envHangDumpDir); dir != "" {
		name := filepath.Join(dir, fmt.Sprintf("goroutines-pid%d-%d.txt", os.Getpid(), seq))
		if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "[hangdiag] could not write dump file %s: %v\n", name, err)
		} else {
			fmt.Fprintf(os.Stderr, "[hangdiag] dump #%d written to %s\n", seq, name)
		}
	}
}

// armHangWatchdog starts the watchdog when RSLINT_HANG_WATCHDOG_MS is set, and
// returns a cancel func to call on normal completion. No-op (cancel is a noop)
// when disabled. Scope the call to the lint run, NOT process-global, so the
// long-lived LSP/API modes are never killed.
func armHangWatchdog() (cancel func()) {
	raw := os.Getenv(envHangWatchdog)
	if raw == "" {
		return func() {}
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms <= 0 {
		fmt.Fprintf(os.Stderr, "[hangdiag] ignoring invalid %s=%q\n", envHangWatchdog, raw)
		return func() {}
	}
	timer := time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
		dumpAllGoroutines(fmt.Sprintf("watchdog: no exit within %dms (deadlock suspected)", ms))
		// Force termination so the parent's `await childExit` resolves and the
		// CI loop fails fast WITH the dump above, instead of hanging the job.
		os.Exit(hangDiagExitCode)
	})
	return func() { timer.Stop() }
}

// installDumpSignalHandler wires an on-demand goroutine dump to the platform's
// dump signal (unix: SIGQUIT; windows: SIGBREAK). Only installed when hang diag
// is enabled, so default signal behavior is untouched in production. The
// handler dumps and KEEPS RUNNING (unlike Go's built-in SIGQUIT crash), so the
// parent can sample a wedged child repeatedly.
func installDumpSignalHandler() {
	if !hangDiagEnabled() {
		return
	}
	sigs := dumpSignals()
	if len(sigs) == 0 {
		return
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	go func() {
		for s := range ch {
			dumpAllGoroutines(fmt.Sprintf("on-demand signal %v", s))
		}
	}()
}
