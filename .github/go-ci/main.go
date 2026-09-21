// go-ci keeps Go's normal build, vet, and test behavior while admitting its
// toolchain processes and test binaries against one shared Windows budget.
// cspell:words toolexec TOOLEXEC IMPORTPATH
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type request struct {
	Token   string `json:"token"`
	Phase   string `json:"phase"`
	Package string `json:"package"`
}

type grant struct {
	ReservedMiB int64  `json:"reserved_mib"`
	SoftMiB     int64  `json:"soft_limit_mib"`
	Error       string `json:"error,omitempty"`
}

var logMu sync.Mutex

func record(kind string, value any) {
	data, _ := json.Marshal(value)
	logMu.Lock()
	defer logMu.Unlock()
	fmt.Printf("Go scheduler %s: %s\n", kind, data)
}

func main() {
	var err error
	if len(os.Args) > 2 && (os.Args[1] == "tool" || os.Args[1] == "test") {
		err = wrap(os.Args[1], os.Args[2:])
	} else if len(os.Args) == 2 && os.Args[1] == "run" {
		err = run()
	} else {
		err = errors.New("usage: go-ci run | tool command [arguments] | test binary [arguments]")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() > 0 {
			os.Exit(exit.ExitCode())
		}
		os.Exit(1)
	}
	// On Windows this closes the wrapper's Job Object, killing any descendants
	// left behind by a child. Do not close that handle while the wrapper is alive.
	os.Exit(0)
}

func command(args []string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd
}

func toolPhase(mode string, args []string) string {
	if mode == "test" {
		return "test"
	}
	phase := strings.TrimSuffix(filepath.Base(args[0]), ".exe")
	if phase == "compile" || phase == "link" {
		return phase
	}
	return "other"
}

func wrap(mode string, args []string) error {
	// Forward the tool's identity unchanged so wrapping does not invalidate
	// Go's build cache. Version probes do not compile or link anything.
	if len(args) == 2 && args[1] == "-V=full" {
		return command(args).Run()
	}
	conn, err := net.DialTimeout("tcp", os.Getenv("RSLINT_GO_SCHEDULER"), 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Minute))
	encoder, decoder := json.NewEncoder(conn), json.NewDecoder(conn)
	pkg := os.Getenv("TOOLEXEC_IMPORTPATH")
	if mode == "test" {
		pkg, _ = os.Getwd()
	}
	if err = encoder.Encode(request{os.Getenv("RSLINT_GO_TOKEN"), toolPhase(mode, args), pkg}); err != nil {
		return err
	}
	var g grant
	if err = decoder.Decode(&g); err != nil {
		return fmt.Errorf("waiting for memory budget: %w", err)
	}
	if g.Error != "" {
		return errors.New(g.Error)
	}
	_ = conn.SetDeadline(time.Time{})
	if g.SoftMiB > 0 {
		_ = os.Setenv("GOMEMLIMIT", fmt.Sprintf("%dMiB", g.SoftMiB))
	}
	// Establish the process tree before spawning the actual tool. Its children
	// inherit this nested job; even a very short-lived child's peak is counted.
	sample, err := trackTree()
	if err != nil {
		return err
	}
	cmd := command(args)
	if err = cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		var finished bool
		select {
		case err = <-done:
			finished = true
		case <-ticker.C:
		}
		u, sampleErr := sample()
		if sampleErr != nil {
			_ = cmd.Process.Kill()
			return sampleErr
		}
		u.Finished = finished
		if finished && err != nil {
			u.ExitCode = 1
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				u.ExitCode = exit.ExitCode()
			}
		}
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if sendErr := encoder.Encode(u); sendErr != nil {
			_ = cmd.Process.Kill()
			return sendErr
		}
		if finished {
			// Wait until the controller has recorded the final peak before Go
			// can finish this command and print the aggregate summary.
			_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			var ack bool
			if ackErr := decoder.Decode(&ack); ackErr != nil {
				return ackErr
			}
			return err
		}
	}
}

func serve(conn net.Conn, token string, b *budget) {
	defer conn.Close()
	decoder, encoder := json.NewDecoder(conn), json.NewEncoder(conn)
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var req request
	if decoder.Decode(&req) != nil || req.Token != token {
		return
	}
	_ = conn.SetReadDeadline(time.Time{})
	updates := make(chan usage)
	closed := make(chan struct{})
	stopReader := make(chan struct{})
	defer close(stopReader)
	go func() {
		defer close(closed)
		for {
			var u usage
			if decoder.Decode(&u) != nil {
				return
			}
			select {
			case updates <- u:
			case <-stopReader:
				return
			}
		}
	}()
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var r *reservation
	for r == nil {
		var err error
		r, err = b.tryAcquire(req.Phase)
		if err != nil || time.Since(start) > 9*time.Minute {
			if err == nil {
				err = errors.New("no memory admission within 9 minutes; refusing to overcommit")
			}
			_ = encoder.Encode(grant{Error: err.Error()})
			return
		}
		if r == nil {
			select {
			case <-closed:
				return
			case <-ticker.C:
			}
		}
	}
	admitted := time.Now()
	soft := int64(0)
	if r.MiB > 0 {
		soft = max(1, min(4096, r.MiB/2))
	}
	defer func() { b.release(r, admitted.Sub(start), time.Since(admitted)) }()
	if encoder.Encode(grant{ReservedMiB: r.MiB, SoftMiB: soft}) != nil {
		return
	}
	for {
		select {
		case u := <-updates:
			b.observe(r, u)
			if u.Finished {
				record("process", map[string]any{"phase": req.Phase, "package": req.Package, "reserved_mib": r.MiB, "soft_limit_mib": soft, "peak_commit_mib": u.PeakMiB, "wait_ms": admitted.Sub(start).Milliseconds(), "duration_ms": time.Since(admitted).Milliseconds(), "exit_code": u.ExitCode})
				_ = encoder.Encode(true)
				return
			}
		case <-closed:
			return
		}
	}
}

// Go's flag parser accepts a quoted executable followed by wrapper arguments.
// Forward slashes avoid introducing shell-style backslash escaping on Windows.
func wrapperFlag(executable, mode string) string {
	return `"` + strings.ReplaceAll(executable, `\`, "/") + `" ` + mode
}

func run() (result error) {
	started := time.Now()
	c := readConfig(os.Getenv)
	record("config", c)
	_ = os.Setenv("GOMAXPROCS", strconv.Itoa(c.CPU))
	b := newBudget(c, availableCommitMiB)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	secret := make([]byte, 32)
	if _, err = rand.Read(secret); err != nil {
		_ = listener.Close()
		return err
	}
	token := hex.EncodeToString(secret)
	_ = os.Setenv("RSLINT_GO_SCHEDULER", listener.Addr().String())
	_ = os.Setenv("RSLINT_GO_TOKEN", token)
	var handlers sync.WaitGroup
	acceptDone := make(chan struct{})
	go func() {
		defer close(acceptDone)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			handlers.Add(1)
			go func() { defer handlers.Done(); serve(conn, token, b) }()
		}
	}()
	defer func() {
		_ = listener.Close()
		<-acceptDone
		handlers.Wait()
		record("summary", map[string]any{"config": c, "phases": b.stats, "max_active": b.maxActive, "max_reserved_mib": b.maxReservedMiB, "duration_ms": time.Since(started).Milliseconds(), "success": result == nil})
	}()
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	flags := []string{"-p=" + strconv.Itoa(c.CPU)}
	if c.BudgetMiB > 0 {
		flags = append(flags, "-toolexec="+wrapperFlag(executable, "tool"))
	}
	testFlags := append(append([]string{}, flags...), "-count=1")
	if c.BudgetMiB > 0 {
		testFlags = append(testFlags, "-exec="+wrapperFlag(executable, "test"))
	}
	runGo := func(phase string, args []string) error {
		start := time.Now()
		err := command(append([]string{"go"}, args...)).Run()
		record("stage", map[string]any{"stage": phase, "duration_ms": time.Since(start).Milliseconds(), "success": err == nil})
		return err
	}
	if err = runGo("shim", append(append([]string{"test"}, testFlags...), "github.com/microsoft/TypeScript/tsc/shim/checker")); err != nil {
		return err
	}
	if err = runGo("build", append(append([]string{"build"}, flags...), "./cmd/...", "./internal/...")); err != nil {
		return err
	}
	list := exec.Command("go", "list", "-p="+strconv.Itoa(c.CPU), "./cmd/...", "./internal/...")
	list.Stderr = os.Stderr
	output, err := list.Output()
	if err != nil {
		return err
	}
	packages := strings.Fields(string(output))
	if len(packages) == 0 {
		return errors.New("go list returned no packages")
	}
	batchCount := 0
	for offset := 0; offset < len(packages); offset += 32 {
		batch := packages[offset:min(offset+32, len(packages))]
		batchCount++
		record("batch", map[string]any{"index": batchCount, "packages": len(batch), "final": offset+len(batch) == len(packages)})
		if err = runGo(fmt.Sprintf("batch-%d", batchCount), append(append([]string{"test"}, testFlags...), batch...)); err != nil {
			return err
		}
	}
	record("complete", map[string]int{"packages": len(packages), "batches": batchCount})
	return nil
}
