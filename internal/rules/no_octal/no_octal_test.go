package no_octal

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIsOctalLiteralRaw(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected bool
	}{
		// ---- ESLint upstream invalid cases (legacy octal + leading-zero decimal) ----
		{name: "01234", raw: "01234", expected: true},
		{name: "07", raw: "07", expected: true},
		{name: "00", raw: "00", expected: true},
		{name: "08 (leading-zero decimal)", raw: "08", expected: true},
		{name: "09.1", raw: "09.1", expected: true},
		{name: "09e1", raw: "09e1", expected: true},
		{name: "09.1e1", raw: "09.1e1", expected: true},
		{name: "018", raw: "018", expected: true},
		{name: "019.1", raw: "019.1", expected: true},
		{name: "019e1", raw: "019e1", expected: true},
		{name: "019.1e1", raw: "019.1e1", expected: true},

		// ---- ESLint upstream valid cases ----
		{name: "0", raw: "0", expected: false},
		{name: "0.1", raw: "0.1", expected: false},
		{name: "0.5e1", raw: "0.5e1", expected: false},
		{name: "0x1234", raw: "0x1234", expected: false},
		{name: "0X5", raw: "0X5", expected: false},

		// ---- Modern literal forms (correctly excluded) ----
		{name: "0o17", raw: "0o17", expected: false},
		{name: "0O17", raw: "0O17", expected: false},
		{name: "0b101", raw: "0b101", expected: false},
		{name: "0B101", raw: "0B101", expected: false},

		// ---- Plain decimals ----
		{name: "1", raw: "1", expected: false},
		{name: "123", raw: "123", expected: false},
		{name: "1.5", raw: "1.5", expected: false},
		{name: "1e5", raw: "1e5", expected: false},
		{name: ".5", raw: ".5", expected: false},

		// ---- Boundary: too short to be 0-prefixed ----
		{name: "empty", raw: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOctalLiteralRaw(tt.raw); got != tt.expected {
				t.Errorf("isOctalLiteralRaw(%q) = %v, want %v", tt.raw, got, tt.expected)
			}
		})
	}
}

func TestNoOctalRule(t *testing.T) {
	// Invalid cases cannot be exercised via the tsconfig-bound rule_tester because
	// the TypeScript parser rejects octal literals (TS1121) and leading-zero
	// decimals (TS1489) as syntactic errors, preventing program creation.
	// Detection logic is covered by TestIsOctalLiteralRaw; production files use the
	// lenient fallback program, where the listener fires normally.
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoOctalRule,
		[]rule_tester.ValidTestCase{
			// ---- ESLint upstream valid cases ----
			{Code: `var a = 'hello world';`},
			{Code: `0x1234`},
			{Code: `0X5;`},
			{Code: `a = 0;`},
			{Code: `0.1`},
			{Code: `0.5e1`},

			// ---- Modern literal forms that share a leading zero ----
			{Code: `0o17`},
			{Code: `0O17`},
			{Code: `0b101`},
			{Code: `0B101`},
			{Code: `0n`},

			// ---- Plain decimals ----
			{Code: `123`},
			{Code: `1.5`},
			{Code: `1e5`},
			{Code: `.5`},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
