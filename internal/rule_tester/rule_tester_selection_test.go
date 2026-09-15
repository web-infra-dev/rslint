package rule_tester

// cspell:ignore tparallel

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// Check the test binary's exit status: a focused suite must fail the command,
// even when the selected case would otherwise be skipped or pass.
func TestRunRuleTesterRejectsIncompleteSuites(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{"valid", "invalid", "json", "empty"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(os.Args[0], "-test.run=^TestRunRuleTesterIncompleteSuiteSubprocess$")
			cmd.Env = append(os.Environ(), "RSLINT_TEST_SUITE_SELECTION="+mode)
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("incomplete suite returned success:\n%s", output)
			}
			want := "focused rule test cases (Only) are not allowed"
			if mode == "empty" {
				want = "rule test suite must not be empty"
			}
			if !strings.Contains(string(output), want) {
				t.Fatalf("test command failed without the selection guard %q:\n%s", want, output)
			}
		})
	}
}

//nolint:tparallel // RunRuleTester calls t.Parallel; another call here would panic before the guard.
func TestRunRuleTesterIncompleteSuiteSubprocess(t *testing.T) {
	mode := os.Getenv("RSLINT_TEST_SUITE_SELECTION")
	if mode == "" {
		t.Skip("only runs in the selection guard subprocess")
	}
	var valid []ValidTestCase
	var invalid []InvalidTestCase
	switch mode {
	case "valid":
		valid = []ValidTestCase{{Only: true, Skip: true}}
	case "invalid":
		invalid = []InvalidTestCase{{Only: true, Skip: true}}
	case "json":
		var suite ESLintTestSuite
		if err := json.Unmarshal([]byte(`{"valid":[{"code":"","only":true,"skip":true}]}`), &suite); err != nil {
			t.Fatal(err)
		}
		converted := ConvertESLintTestSuite(&suite)
		valid, invalid = converted.Valid, converted.Invalid
	case "empty":
	default:
		t.Fatalf("unexpected selection mode %q", mode)
	}
	RunRuleTester(Root{}, "", t, &rule.Rule{}, valid, invalid)
}
