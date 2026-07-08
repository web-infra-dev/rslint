package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestGate_BypassedTypeAwareRuleOnGapFile_Crashes is the REVERSE assertion that
// nails the type-aware gate (GetActiveRulesForFile) as the *only* line of
// defense for gap files.
//
// A gap file has no tsconfig coverage, so RunLinter passes its rules a nil
// TypeChecker (linter.go: TypeInfoFiles is non-nil and does not contain it). A
// RequiresTypeInfo rule dereferences that checker without a nil guard, so if
// such a rule is ever allowed to run on a gap file — i.e. the gate is bypassed
// — the whole process crashes (the rule runs in a worker goroutine with no
// recover). The forward safety case is TestHandleLint_GapFile_TypeAwareRuleGatedOff.
//
// This runs the bypassed-gate scenario (rules taken from GetEnabledRules, which
// does NOT gate, instead of GetActiveRulesForFile, which does) in a subprocess
// and asserts it does NOT exit cleanly. If a future change makes a bypassed
// gate "safe" (e.g. a nil-TypeChecker guard inside rules, or a second filtering
// layer), this fails — forcing a deliberate re-examination of whether the gate
// is still load-bearing.
func TestGate_BypassedTypeAwareRuleOnGapFile_Crashes(t *testing.T) {
	if os.Getenv("RSLINT_GATE_CRASH_CHILD") == "1" {
		runBypassedGateScenario()
		// Reaching here means the type-aware rule did NOT crash on the nil
		// TypeChecker — exit 0 so the parent's "expected a crash" fails loudly.
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestGate_BypassedTypeAwareRuleOnGapFile_Crashes$")
	cmd.Env = append(os.Environ(), "RSLINT_GATE_CRASH_CHILD=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected the bypassed-gate scenario to crash (nil TypeChecker dereference): "+
			"the type-aware gate is the only thing preventing this crash and must not be "+
			"silently neutralized by a second defense layer.\nchild stderr:\n%s", stderr.String())
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected the subprocess to terminate with a non-zero exit (crash), got %T: %v", err, err)
	}
	// Pin the crash to its CAUSE — a nil-TypeChecker dereference inside the
	// type-aware rule — so an unrelated panic elsewhere cannot masquerade as a
	// passing gate-regression. The Go runtime prints the nil-pointer signature
	// and a stack frame in the rule package.
	out := stderr.String()
	if !strings.Contains(out, "nil pointer dereference") {
		t.Fatalf("expected a nil-pointer-dereference crash signature, got child stderr:\n%s", out)
	}
	if !strings.Contains(out, "no_unsafe_member_access") {
		t.Fatalf("expected the crash to originate in the type-aware rule (no_unsafe_member_access), got child stderr:\n%s", out)
	}
}

// runBypassedGateScenario reproduces a gap-file lint with the type-aware gate
// REMOVED: rules come from GetEnabledRules (no gate) instead of
// GetActiveRulesForFile (gate), so a RequiresTypeInfo rule is handed to a file
// that RunLinter gives a nil TypeChecker. This MUST crash; it exists only to be
// driven by the subprocess above.
func runBypassedGateScenario() {
	tmpDir := tspath.NormalizePath(os.TempDir())
	gapFile := tspath.NormalizePath(filepath.Join(tmpDir, "rslint-gate-crash-gap.ts"))

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	fs = utils.NewOverlayVFS(fs, map[string]string{gapFile: "let a: any = 10;\na.b = 20;\n"})

	parseCache := utils.NewParseCache()
	fallback, _ := createFallbackProgram([]string{gapFile}, false, tmpDir, fs, parseCache, "", "")
	if fallback == nil {
		return // could not build the program → cannot reproduce; parent fails
	}

	rslintconfig.RegisterAllRules()
	cfg := rslintconfig.RslintConfig{
		rslintconfig.ConfigEntry{
			Files:   []string{"**/*.ts"},
			Rules:   rslintconfig.Rules{"@typescript-eslint/no-unsafe-member-access": "error"},
			Plugins: []string{"@typescript-eslint"},
		},
	}
	// GetEnabledRules is the NON-gated path: it returns the RequiresTypeInfo
	// rule for the gap file unconditionally (GetActiveRulesForFile would have
	// filtered it out via typeInfoFiles).
	rules, _ := rslintconfig.GlobalRuleRegistry.GetEnabledRules(cfg, gapFile, tmpDir, false)

	// TypeInfoFiles is non-nil and excludes the gap file, so RunLinter passes a
	// nil TypeChecker for it — exactly the HandleLint setup, minus the gate.
	typeInfoFiles := map[string]struct{}{tspath.NormalizePath(filepath.Join(tmpDir, "unrelated.ts")): {}}
	_, _ = linter.RunLinter(linter.RunLinterOptions{
		Programs:      []*compiler.Program{fallback},
		Scope:         linter.FileScope{Files: []string{gapFile}},
		TypeInfoFiles: typeInfoFiles,
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return rules
		},
		OnDiagnostic: func(rule.RuleDiagnostic) {},
	})
}
