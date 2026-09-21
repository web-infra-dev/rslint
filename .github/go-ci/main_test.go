// cspell:words toolexec
package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	for _, tt := range []struct {
		cpu, memory, fallback string
		wantCPU               int
		wantMemory            int64
	}{
		{"10", "40", "16", 10, 32768},
		{"32", "128", "16", 32, 104857},
		{"20", "80", "16", 20, 65536},
		{"2.5", "4.5", "16", 2, 3686},
		{"0.5", "1", "16", 1, 819},
		{"", "", "16", 16, 0},
		{"bad", "-1", "8", 8, 0},
		{"NaN", "+Inf", "16", 16, 0},
		{"-4", "40", "16", 16, 32768},
		{"64", "bad", "16", 64, 0},
	} {
		t.Run(tt.cpu+"-"+tt.memory, func(t *testing.T) {
			values := map[string]string{"cpu_limit": tt.cpu, "memory_Limit": tt.memory, "GOMAXPROCS": tt.fallback}
			c := readConfig(func(key string) string { return values[key] })
			if c.CPU != tt.wantCPU || c.BudgetMiB != tt.wantMemory {
				t.Fatalf("got %+v; want CPU=%d memory=%d", c, tt.wantCPU, tt.wantMemory)
			}
		})
	}
}

func unlimitedCommit() (int64, error) { return 1 << 30, nil }

func TestMixedReservations(t *testing.T) {
	b := newBudget(config{CPU: 10, BudgetMiB: 32768, HeadroomMiB: 8192}, unlimitedCommit)
	var first *reservation
	for _, phase := range []string{"compile", "compile", "test", "test", "link", "link"} {
		r, err := b.tryAcquire(phase)
		if r == nil || err != nil {
			t.Fatalf("could not admit %s: %v", phase, err)
		}
		if first == nil {
			first = r
		}
	}
	if r, _ := b.tryAcquire("test"); r != nil {
		t.Fatal("mixed phases exceeded the shared 32 GiB budget")
	}
	b.release(first, 0, 0)
	if r, _ := b.tryAcquire("test"); r == nil {
		t.Fatal("released memory was not available to another phase")
	}
}

func TestAtomicAdmission(t *testing.T) {
	b := newBudget(config{CPU: 32, BudgetMiB: 32768}, unlimitedCommit)
	var group sync.WaitGroup
	for range 32 {
		group.Go(func() { _, _ = b.tryAcquire("link") })
	}
	group.Wait()
	if len(b.active) != 4 {
		t.Fatalf("concurrent admission granted %d linkers, want 4", len(b.active))
	}
}

func TestCommitHeadroomAndPeakGrowth(t *testing.T) {
	available := int64(20 * 1024)
	b := newBudget(config{CPU: 10, BudgetMiB: 32768, HeadroomMiB: 8192}, func() (int64, error) { return available, nil })
	first, err := b.tryAcquire("link")
	if err != nil || first == nil {
		t.Fatal("first linker should fit")
	}
	if second, _ := b.tryAcquire("link"); second != nil {
		t.Fatal("ignored uncommitted memory promised to the first linker")
	}
	// A process exceeding its estimate grows its reservation immediately.
	available = 1 << 30
	b.observe(first, usage{CurrentMiB: 25 * 1024, PeakMiB: 25 * 1024})
	if r, _ := b.tryAcquire("compile"); r != nil {
		t.Fatal("admitted more work after an observed peak exceeded the budget")
	}
	b.release(first, 0, 0)
	if b.stats["link"].EstimateMiB != 38400 {
		t.Fatal("did not retain peak plus 50% headroom")
	}
	// Smaller later processes must not reduce the forecast for large packages.
	next, _ := b.tryAcquire("link")
	b.observe(next, usage{PeakMiB: 100})
	if b.stats["link"].EstimateMiB != 38400 {
		t.Fatal("reduced the estimate after a small package")
	}
}

func TestFallbackAndCPUQuota(t *testing.T) {
	b := newBudget(config{CPU: 24}, func() (int64, error) { t.Fatal("queried memory without a quota"); return 0, nil })
	for range 24 {
		if r, _ := b.tryAcquire("test"); r == nil || r.MiB != 0 {
			t.Fatal("unexpected memory restriction or legacy 16-CPU cap")
		}
	}
	if r, _ := b.tryAcquire("test"); r != nil {
		t.Fatal("exceeded CPU quota")
	}
	b = newBudget(config{CPU: 2, BudgetMiB: 32768}, func() (int64, error) { return 0, errors.New("probe failed") })
	if _, err := b.tryAcquire("test"); err == nil {
		t.Fatal("failed open when memory accounting was unavailable")
	}
}

func TestWrapperIntegration(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "go-ci.exe") // Exercise quoted executable paths.
	if output, err := exec.Command("go", "build", "-p=1", "-o", exe, ".").CombinedOutput(); err != nil {
		t.Fatalf("build wrapper: %v\n%s", err, output)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	b := newBudget(config{CPU: 2}, unlimitedCommit)
	var handlers sync.WaitGroup
	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			handlers.Go(func() { serve(conn, "test-token", b) })
		}
	}()
	defer func() { listener.Close(); <-accepted; handlers.Wait() }()
	t.Setenv("RSLINT_GO_SCHEDULER", listener.Addr().String())
	t.Setenv("RSLINT_GO_TOKEN", "test-token")
	t.Setenv("GOFLAGS", "")
	// This is a tiny independent module, not the repository's test suite.
	for name, source := range map[string]string{
		"go.mod":     "module fixture\n\ngo 1.26.0\n",
		"fixture.go": "package fixture\nfunc Value() int { return 42 }\n",
		"fixture_test.go": `package fixture
import ("os"; "testing")
func TestFixture(t *testing.T) {
 if Value() != 42 { t.Fatal("wrong value") }
 if _, err := os.Stat("go.mod"); err != nil { t.Fatal("working directory changed", err) }
 if os.Getenv("FAIL_FIXTURE") == "yes" { t.Fatal("requested failure") }
}
`,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"test", "-p=2", "-toolexec=" + wrapperFlag(exe, "tool"), "-exec=" + wrapperFlag(exe, "test"), "-count=1", "-run=TestFixture", "."}
	for _, fail := range []bool{false, true} {
		cmd := exec.Command("go", args...)
		cmd.Dir = dir
		cmd.Env = os.Environ()
		if fail {
			cmd.Env = append(cmd.Env, "FAIL_FIXTURE=yes")
		}
		out, err := cmd.CombinedOutput()
		if fail != (err != nil) || (fail && !bytes.Contains(out, []byte("requested failure"))) {
			t.Fatalf("failure=%v err=%v output=%s", fail, err, out)
		}
	}
	// Also ensure arbitrary child exit codes and arguments survive the wrapper.
	self, _ := os.Executable()
	cmd := exec.Command(exe, "test", self, "-test.run=^TestWrapperChild$", "--", "a b", `x\y`)
	cmd.Env = append(os.Environ(), "GO_CI_CHILD=yes")
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 7 || !strings.Contains(string(out), `a b|x\y`) {
		t.Fatalf("child semantics changed: err=%v output=%s", err, out)
	}
	listener.Close()
	<-accepted
	handlers.Wait()
	if b.stats["test"].Processes != 3 || b.stats["compile"].Processes == 0 || b.stats["link"].Processes == 0 {
		t.Fatalf("missing wrapped phases: %+v", b.stats)
	}
	if runtime.GOOS == "windows" && b.stats["test"].PeakMiB < 64 {
		t.Fatalf("Windows process tree peak was not measured: %+v", b.stats["test"])
	}
}

func TestWrapperChild(t *testing.T) {
	if os.Getenv("GO_CI_CHILD") != "yes" {
		return
	}
	data := make([]byte, 64*mib)
	for i := range data {
		data[i] = byte(i)
	}
	time.Sleep(300 * time.Millisecond)
	fmt.Println(strings.Join(os.Args[len(os.Args)-2:], "|"))
	runtime.KeepAlive(data)
	os.Exit(7)
}
