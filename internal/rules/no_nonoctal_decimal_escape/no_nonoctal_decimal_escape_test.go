package no_nonoctal_decimal_escape

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestScanDecimalEscapes locks in the byte-level scanner used by the rule.
// The rule_tester invalid suite also covers behavior end-to-end, but the
// scanner has many edge cases (multi-byte chars, line continuations, escape
// pair handling) that are easier to assert as pure-function inputs/outputs.
func TestScanDecimalEscapes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []decimalEscapeHit
	}{
		// ---- No \8 or \9 ----
		{name: "empty", raw: ``, want: nil},
		{name: "no escape", raw: `'foo'`, want: nil},
		{name: "plain digit 8", raw: `'8'`, want: nil},
		{name: "plain digit 9", raw: `'9'`, want: nil},
		{name: "octal escape \\1", raw: `'\1'`, want: nil},
		{name: "null escape", raw: `'\0'`, want: nil},
		{name: "escaped backslash + 8", raw: `'\\8'`, want: nil},
		{name: "escaped backslash + 9", raw: `'\\9'`, want: nil},
		{name: "many backslash pairs + 8", raw: `'\\\\\\8'`, want: nil},

		// ---- Single \8 ----
		{name: "single \\8", raw: `'\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "single \\9", raw: `'\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\9`, previousEscapeStart: -1},
		}},
		{name: "double-quoted \\8", raw: `"\8"`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "leading text", raw: `'foo\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\9`, previousEscapeStart: -1},
		}},

		// ---- Adjacent NULL escape (\0\8 / \0\9) ----
		{name: "adjacent \\0\\8", raw: `'\0\8'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},
		{name: "adjacent \\0\\9", raw: `'\0\9'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\9`},
		}},
		{name: "leading text + \\0\\8", raw: `'foo\0\8'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 4, decimalEscapeStart: 6, decimalEscapeEnd: 8, decimalEscape: `\8`},
		}},

		// ---- Non-adjacent NULL escape — anything between \0 and \X clears it ----
		{name: "\\0 space \\8", raw: `'\0 \8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "\\0\\1\\8", raw: `'\0\1\8'`, want: []decimalEscapeHit{
			{previousEscape: `\1`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`},
		}},
		{name: "\\01\\8 (octal between)", raw: `'\01\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "0 (no \\) \\8", raw: `'0\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 2, decimalEscapeEnd: 4, decimalEscape: `\8`, previousEscapeStart: -1},
		}},

		// ---- Escaped backslash before \X — escape pair should NOT be the NULL pair ----
		{name: "\\\\\\8 (3 backslashes + 8)", raw: `'\\\8'`, want: []decimalEscapeHit{
			{previousEscape: `\\`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},
		{name: "\\\\0\\8 (escaped backslash, 0, then \\8)", raw: `'\\0\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\8`, previousEscapeStart: -1},
		}},

		// ---- Multiple decimal escapes ----
		{name: "\\8\\8", raw: `'\8\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "\\9\\8", raw: `'\9\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\9`, previousEscapeStart: -1},
			{decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "\\0\\8\\9 — first hit special, second hit standard", raw: `'\0\8\9'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
			{decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\9`, previousEscapeStart: -1},
		}},
		{name: "\\8\\0\\9 — first standard, second special", raw: `'\8\0\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
			{previousEscape: `\0`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\9`},
		}},

		// ---- Multi-byte UTF-8 between escapes ----
		// \👍\8 — backslash + emoji + \8. The byte after \ is the leading byte
		// of the 4-byte 👍 sequence; the scanner must consume \X as a 2-byte
		// pair without misreading the continuation bytes as backslashes.
		{name: "\\👍\\8", raw: "'\\\xf0\x9f\x91\x8d\\8'", want: []decimalEscapeHit{
			{decimalEscapeStart: 6, decimalEscapeEnd: 8, decimalEscape: `\8`, previousEscapeStart: -1},
		}},

		// ---- Line continuations — \<LF> consumed as a 2-byte escape pair.
		// Tracks ESLint's `\\.` semantics under the `s` flag: `previousEscape`
		// captures the literal backslash + LF; the rule's reporter still treats
		// it as non-`\0` and falls through to the standard refactor.
		{name: "\\<LF>\\8", raw: "'\\\n\\8'", want: []decimalEscapeHit{
			{previousEscape: "\\\n", previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},
		{name: "\\<CRLF>\\9 — CR consumed with backslash, LF resets prev", raw: "'\\\r\n\\9'", want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\9`, previousEscapeStart: -1},
		}},
		{name: "\\<CR>\\8", raw: "'\\\r\\8'", want: []decimalEscapeHit{
			{previousEscape: "\\\r", previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},

		// ---- Upstream ESLint invalid suite (parity coverage) ----
		// The full invalid set from https://github.com/eslint/eslint/blob/main/tests/lib/rules/no-nonoctal-decimal-escape.js
		// Names mirror upstream code so a future audit can grep both files.
		{name: "upstream: 'f\\9'", raw: `'f\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 2, decimalEscapeEnd: 4, decimalEscape: `\9`, previousEscapeStart: -1},
		}},
		{name: "upstream: 'fo\\9'", raw: `'fo\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\9`, previousEscapeStart: -1},
		}},
		{name: "upstream: '\\\\\\\\\\9' — 5 backslashes + 9", raw: `'\\\\\9'`, want: []decimalEscapeHit{
			{previousEscape: `\\`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\9`},
		}},
		{name: "upstream: 'foo\\\\\\8' — text + escaped \\\\ + \\8", raw: `'foo\\\8'`, want: []decimalEscapeHit{
			{previousEscape: `\\`, previousEscapeStart: 4, decimalEscapeStart: 6, decimalEscapeEnd: 8, decimalEscape: `\8`},
		}},
		{name: "upstream: '\\ \\8' — \\<space> then \\8", raw: `'\ \8'`, want: []decimalEscapeHit{
			// `\<space>` consumed as a 2-byte escape pair → previousEscape captures it.
			{previousEscape: `\ `, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},
		{name: "upstream: '\\1\\9' — octal then \\9", raw: `'\1\9'`, want: []decimalEscapeHit{
			{previousEscape: `\1`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\9`},
		}},
		{name: "upstream: 'foo\\1\\9' — text + octal + \\9", raw: `'foo\1\9'`, want: []decimalEscapeHit{
			{previousEscape: `\1`, previousEscapeStart: 4, decimalEscapeStart: 6, decimalEscapeEnd: 8, decimalEscape: `\9`},
		}},
		{name: "upstream: '\\n\\n\\8\\n' — escape sequences around \\8", raw: `'\n\n\8\n'`, want: []decimalEscapeHit{
			{previousEscape: `\n`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`},
		}},
		{name: "upstream: '\\👍\\8' — escaped emoji adjacent to \\8", raw: "'\\\xf0\x9f\x91\x8d\\8'", want: []decimalEscapeHit{
			// `\<emoji>` consumed as 2-byte pair (\\ + first emoji byte). The
			// remaining 3 continuation bytes reset previousEscape to "". So when
			// \8 is reached, prev is empty — diverges from ESLint where the `u`
			// flag makes `.` match the whole code point. Documented in the rule
			// .md as a tsgo-byte-scanner side effect (no observable behavior
			// change: both yield the same standard-path suggestion).
			{decimalEscapeStart: 6, decimalEscapeEnd: 8, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "upstream: '\\\\8\\9' — escaped \\\\ + 8 then \\9", raw: `'\\8\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\9`, previousEscapeStart: -1},
		}},
		{name: "upstream: '\\8\\\\9' — \\8 then escaped \\\\ + 9", raw: `'\8\\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "upstream: '\\8 \\\\9' — \\8 then space + escaped \\\\ + 9", raw: `'\8 \\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "upstream: 'foo\\8bar\\9baz' — two hits separated", raw: `'foo\8bar\9baz'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 9, decimalEscapeEnd: 11, decimalEscape: `\9`, previousEscapeStart: -1},
		}},
		{name: "upstream: '\\8\\1\\9' — between octal", raw: `'\8\1\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
			{previousEscape: `\1`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\9`},
		}},
		{name: "upstream: '\\8\\\\\\9' — \\8, escaped \\\\, then \\9", raw: `'\8\\\9'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
			{previousEscape: `\\`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\9`},
		}},
		{name: "upstream: '\\1\\0\\8' — octal + \\0 + \\8 → adjacent NULL", raw: `'\1\0\8'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`},
		}},

		// ---- Real-user / TS-context boundary cases (extra coverage) ----
		// These exercise scanner inputs that come up in real .ts source but
		// aren't directly tested by upstream's unit suite.
		{name: "extra: empty raw string", raw: ``, want: nil},
		{name: "extra: single quote alone", raw: `'`, want: nil},
		{name: "extra: '\\' (truncated trailing backslash)", raw: `'\`, want: nil},
		{name: "extra: '\\\\' alone (escaped backslash, no digit)", raw: `'\\'`, want: nil},
		{name: "extra: '\\u8888' (unicode escape, digits 8 in hex)", raw: `'袈'`, want: nil},
		{name: "extra: '\\xFF' (hex escape, no octal/decimal)", raw: `'\xFF'`, want: nil},
		{name: "extra: '\\xa8\\8' (hex escape adjacent to \\8)", raw: `'\xa8\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "extra: plain digit-8 chars before \\8", raw: `'88\8'`, want: []decimalEscapeHit{
			// '88' before the escape are non-backslash bytes; \8 still picked up.
			{decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "extra: '\\\"\\8' — escaped double-quote then \\8", raw: `'\"\8'`, want: []decimalEscapeHit{
			{previousEscape: `\"`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},
		{name: "extra: \"\\'\\8\" — escaped single-quote then \\8", raw: `"\'\8"`, want: []decimalEscapeHit{
			{previousEscape: `\'`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},
		{name: "extra: '\\\\\\0\\8' — escaped \\\\ + \\0 + \\8 → adjacent NULL", raw: `'\\\0\8'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`},
		}},
		{name: "extra: '\\\\\\\\\\0\\8' — even backslash count + \\0 + \\8", raw: `'\\\\\0\8'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 5, decimalEscapeStart: 7, decimalEscapeEnd: 9, decimalEscape: `\8`},
		}},
		{name: "extra: '\\8' at very end of long string", raw: `'aaaaaaaaaaaa\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 13, decimalEscapeEnd: 15, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "extra: 8 hits in one string", raw: `'\8\8\8\8\8\8\8\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 1, decimalEscapeEnd: 3, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 7, decimalEscapeEnd: 9, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 9, decimalEscapeEnd: 11, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 11, decimalEscapeEnd: 13, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 13, decimalEscapeEnd: 15, decimalEscape: `\8`, previousEscapeStart: -1},
			{decimalEscapeStart: 15, decimalEscapeEnd: 17, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "extra: '\\\\\\\\\\8' (4 backslashes + \\8 → standard prev=\\\\)", raw: `'\\\\\8'`, want: []decimalEscapeHit{
			{previousEscape: `\\`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`},
		}},
		{name: "extra: BOM-like leading bytes don't trip scanner", raw: "'\xef\xbb\xbf\\8'", want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "extra: ascii control \\x01 then \\8", raw: "'\x01\\8'", want: []decimalEscapeHit{
			{decimalEscapeStart: 2, decimalEscapeEnd: 4, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "extra: tab + \\8", raw: "'\t\\8'", want: []decimalEscapeHit{
			{decimalEscapeStart: 2, decimalEscapeEnd: 4, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
		{name: "extra: long path-like prefix + \\8", raw: `'/usr/local/bin/foo\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 19, decimalEscapeEnd: 21, decimalEscape: `\8`, previousEscapeStart: -1},
		}},

		// ---- Adjacency contract — anything (even another \X but not \0) clears special path ----
		{name: "extra: '\\t\\8' — \\t then \\8 → standard", raw: `'\t\8'`, want: []decimalEscapeHit{
			{previousEscape: `\t`, previousEscapeStart: 1, decimalEscapeStart: 3, decimalEscapeEnd: 5, decimalEscape: `\8`},
		}},
		{name: "extra: '\\n\\0\\8' — escape, then \\0, then \\8 → adjacent NULL", raw: `'\n\0\8'`, want: []decimalEscapeHit{
			{previousEscape: `\0`, previousEscapeStart: 3, decimalEscapeStart: 5, decimalEscapeEnd: 7, decimalEscape: `\8`},
		}},
		{name: "extra: '\\0a\\8' — \\0, then plain a, then \\8 → standard", raw: `'\0a\8'`, want: []decimalEscapeHit{
			{decimalEscapeStart: 4, decimalEscapeEnd: 6, decimalEscape: `\8`, previousEscapeStart: -1},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanDecimalEscapes(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scanDecimalEscapes(%q) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}

// TestNoNonoctalDecimalEscapeRule runs the rule against TypeScript fixtures.
// Note: TypeScript's parser rejects \8 / \9 as syntax errors in strict-mode
// modules (which is the rule_tester default). The full diagnostic-level
// behavior — including positions and suggestion outputs — is exercised by
// the JS test suite under packages/rslint-test-tools, which runs rslint via
// the lenient parse path that the production binary uses for files outside
// any tsconfig. Here we only assert that valid code (no \8/\9) is silent.
func TestNoNonoctalDecimalEscapeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNonoctalDecimalEscapeRule,
		[]rule_tester.ValidTestCase{
			// ---- Upstream ESLint valid cases that are TS-parseable ----
			{Code: `var x = 8;`},
			{Code: `var x = "袈";`},
			{Code: `var x = /\8/;`},
			{Code: `var x = /\9/;`},
			{Code: `''`},
			{Code: `""`},
			{Code: `'foo'`},
			{Code: `'8'`},
			{Code: `'9'`},
			{Code: `'foo8'`},
			{Code: `'foo9bar'`},
			{Code: `'\\8'`},
			{Code: `'\\9'`},
			{Code: `'\\8\\9'`},
			{Code: `'\\ \\8'`},
			{Code: `'\\\\9'`},
			{Code: `'\\9bar'`},
			{Code: `'a\\8'`},
			{Code: `'foo\\8'`},
			{Code: `'foo\\8bar'`},
			{Code: `'9\\9'`},
			{Code: `'n\n8'`},
			{Code: `'n\nn\n8'`},
			{Code: `'\\\\\\x38'`},
			{Code: `'\n'`},
			{Code: `'\t'`},
			{Code: `'\0'`},

			// ---- Real TS contexts the rule MUST stay silent on ----
			// String literals containing \8 / \9 are TS1488 syntax errors and
			// can't appear here, but unrelated TS surface MUST not trigger.
			{Code: `const x: string = "hello"`},
			{Code: `function f(x: string = "default") { return x; }`},
			{Code: `type T = "8" | "9"`},
			{Code: `enum E { A = "a", B = "b" }`},
			{Code: `class C { static name = "8"; method() { return "9"; } }`},
			{Code: `const obj = { "8": 1, "9": 2 }`},
			{Code: `const arr = ["8", "9", "foo"]`},
			{Code: `import x from "./mod"; export { x };`},
			// Template literals — out of scope for the rule (KindStringLiteral
			// only). Strict-TS rejects `\8`/`\9` even inside templates, so we
			// only assert that valid templates remain silent.
			{Code: "const t = `foo${'bar'}baz`"},
			{Code: "const t = `\\\\8`"},
			// Tagged template literal — same exclusion.
			{Code: "tag`hello\\\\8`"},
			// JSX strings (TSX) — covered by the same KindStringLiteral filter,
			// but JSX text nodes aren't string literals so they're untouched.
			{Code: `const e = <div title="hello">8 9</div>`, Tsx: true},
			// String inside type-only position.
			{Code: `type S = '\\8' extends '\\8' ? true : false`},
			{Code: `declare const x: '\\8' | '\\9'`},
			// Regex with \8 / \9 (regex backreferences) — KindRegularExpressionLiteral, not KindStringLiteral.
			{Code: `const re1 = /(\w)\1\8/`},
			{Code: `const re2 = /\9/g`},
		},
		// Invalid cases cannot be tested through the strict rule_tester because
		// the TypeScript parser reports `\8` / `\9` as syntactic errors (TS1488),
		// preventing program creation. The detector is exercised in
		// TestScanDecimalEscapes above, and the diagnostic-emission pipeline —
		// positions and suggestion outputs — is exercised in
		// TestNoNonoctalDecimalEscapeDiagnostics below using a lenient program.
		[]rule_tester.InvalidTestCase{},
	)
}

// expectedSuggestion mirrors rule_tester.InvalidTestCaseSuggestion shape so we
// can describe expected suggestion outputs against a hand-built test source.
type expectedSuggestion struct {
	MessageId string
	Output    string
}

// expectedDiagnostic locks in the diagnostic shape (position + message id +
// suggestion outputs) for one report. Line/Column are 1-indexed in
// ECMAScript-compatible UTF-16 columns, matching ESLint's position scheme.
type expectedDiagnostic struct {
	MessageId   string
	Description string
	Line        int
	Column      int
	EndLine     int
	EndColumn   int
	Suggestions []expectedSuggestion
}

// TestNoNonoctalDecimalEscapeDiagnostics exercises the full diagnostic pipeline
// (position computation, message text, suggestion outputs) using a *lenient*
// program — the same code path that runs in production for gap files.
//
// Each case mirrors the upstream ESLint test of the same input, with positions
// and suggestion outputs copied verbatim. We have to drive the linter manually
// here because the strict rule_tester refuses to construct a Program when the
// parser emits TS1488 ("escape sequence '\8' is not allowed").
func TestNoNonoctalDecimalEscapeDiagnostics(t *testing.T) {
	tests := []struct {
		name string
		code string
		tsx  bool
		want []expectedDiagnostic
	}{
		// ---- Plain \8 / \9 in single-/double-quoted strings ----
		{
			name: "single \\8",
			code: `var s = '\8';`,
			want: []expectedDiagnostic{{
				MessageId:   "decimalEscape",
				Description: `Don't use '\8' escape sequence.`,
				Line:        1, Column: 10, EndLine: 1, EndColumn: 12,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: `var s = '8';`},
					{MessageId: "escapeBackslash", Output: `var s = '\\8';`},
				},
			}},
		},
		{
			name: "single \\9",
			code: `var s = '\9';`,
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 10, EndLine: 1, EndColumn: 12,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: `var s = '9';`},
					{MessageId: "escapeBackslash", Output: `var s = '\\9';`},
				},
			}},
		},
		{
			name: "double-quoted \\8",
			code: `var s = "\8";`,
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 10, EndLine: 1, EndColumn: 12,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: `var s = "8";`},
					{MessageId: "escapeBackslash", Output: `var s = "\\8";`},
				},
			}},
		},

		// ---- Mid-string \8 / \9 ----
		{
			name: "foo\\8bar",
			code: `var s = 'foo\8bar';`,
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 13, EndLine: 1, EndColumn: 15,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: `var s = 'foo8bar';`},
					{MessageId: "escapeBackslash", Output: `var s = 'foo\\8bar';`},
				},
			}},
		},

		// ---- Multi-byte char before the escape — column should count UTF-16 units ----
		{
			name: "thumbs-up before \\8",
			code: "var s = '\xf0\x9f\x91\x8d\\8';",
			want: []expectedDiagnostic{{
				// 👍 is a surrogate pair → occupies 2 UTF-16 code units, so \8 lands at column 12.
				MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '\xf0\x9f\x91\x8d8';"},
					{MessageId: "escapeBackslash", Output: "var s = '\xf0\x9f\x91\x8d\\\\8';"},
				},
			}},
		},

		// ---- Multiple escapes in the same string ----
		{
			name: "\\8\\8 — two reports",
			code: `var s = '\8\8';`,
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 10, EndLine: 1, EndColumn: 12,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: `var s = '8\8';`},
						{MessageId: "escapeBackslash", Output: `var s = '\\8\8';`},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: `var s = '\88';`},
						{MessageId: "escapeBackslash", Output: `var s = '\8\\8';`},
					},
				},
			},
		},

		// ---- Multi-line / line-2 reporting ----
		{
			name: "\\9 on line 2",
			code: "var foo = '8'\n  bar = '\\9'",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 2, Column: 10, EndLine: 2, EndColumn: 12,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var foo = '8'\n  bar = '9'"},
					{MessageId: "escapeBackslash", Output: "var foo = '8'\n  bar = '\\\\9'"},
				},
			}},
		},
		{
			name: "line continuation followed by \\8",
			code: "var s = '\\\n\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 2, Column: 1, EndLine: 2, EndColumn: 3,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '\\\n8';"},
					{MessageId: "escapeBackslash", Output: "var s = '\\\n\\\\8';"},
				},
			}},
		},

		// ---- Adjacent NULL escape (special path: 2 refactors + 1 escapeBackslash) ----
		// Outputs use Go interpreted strings so the escape pairs in the
		// post-fix source ('\0', '8', '\\8', etc.) survive verbatim.
		{
			name: "\\0\\8 — adjacent NULL",
			code: "var s = '\\0\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '\\u00008';"},
					{MessageId: "refactor", Output: "var s = '\\0\\u0038';"},
					{MessageId: "escapeBackslash", Output: "var s = '\\0\\\\8';"},
				},
			}},
		},
		{
			name: "foo\\0\\9bar — adjacent NULL with surrounding text",
			code: "var s = 'foo\\0\\9bar';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 15, EndLine: 1, EndColumn: 17,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = 'foo\\u00009bar';"},
					{MessageId: "refactor", Output: "var s = 'foo\\0\\u0039bar';"},
					{MessageId: "escapeBackslash", Output: "var s = 'foo\\0\\\\9bar';"},
				},
			}},
		},
		{
			name: "\\0 space \\8 — NOT adjacent (standard path)",
			code: "var s = '\\0 \\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 13, EndLine: 1, EndColumn: 15,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '\\0 8';"},
					{MessageId: "escapeBackslash", Output: "var s = '\\0 \\\\8';"},
				},
			}},
		},
		{
			name: "\\0\\8\\9 — first special, second standard",
			code: "var s = '\\0\\8\\9';",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "var s = '\\u00008\\9';"},
						{MessageId: "refactor", Output: "var s = '\\0\\u0038\\9';"},
						{MessageId: "escapeBackslash", Output: "var s = '\\0\\\\8\\9';"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 14, EndLine: 1, EndColumn: 16,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "var s = '\\0\\89';"},
						{MessageId: "escapeBackslash", Output: "var s = '\\0\\8\\\\9';"},
					},
				},
			},
		},

		// ---- Upstream parity: every remaining unique invalid case ----
		{
			name: "upstream: 'f\\9' position",
			code: "'f\\9';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 3, EndLine: 1, EndColumn: 5,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'f9';"},
					{MessageId: "escapeBackslash", Output: "'f\\\\9';"},
				},
			}},
		},
		{
			name: "upstream: 'fo\\9' position",
			code: "'fo\\9';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 4, EndLine: 1, EndColumn: 6,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'fo9';"},
					{MessageId: "escapeBackslash", Output: "'fo\\\\9';"},
				},
			}},
		},
		{
			name: "upstream: 4 escaped backslashes + \\9 — column 6",
			code: "'\\\\\\\\\\9';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 6, EndLine: 1, EndColumn: 8,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\\\\\\\9';"},
					{MessageId: "escapeBackslash", Output: "'\\\\\\\\\\\\9';"},
				},
			}},
		},
		{
			name: "upstream: 'foo\\\\\\8' — text + escaped \\\\ + \\8",
			code: "'foo\\\\\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 7, EndLine: 1, EndColumn: 9,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'foo\\\\8';"},
					{MessageId: "escapeBackslash", Output: "'foo\\\\\\\\8';"},
				},
			}},
		},
		{
			name: "upstream: '\\ \\8' — \\<space> then \\8",
			code: "'\\ \\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 4, EndLine: 1, EndColumn: 6,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\ 8';"},
					{MessageId: "escapeBackslash", Output: "'\\ \\\\8';"},
				},
			}},
		},
		{
			name: "upstream: '\\1\\9' — octal then \\9 (standard path)",
			code: "'\\1\\9';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 4, EndLine: 1, EndColumn: 6,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\19';"},
					{MessageId: "escapeBackslash", Output: "'\\1\\\\9';"},
				},
			}},
		},
		{
			name: "upstream: '\\n\\n\\8\\n'",
			code: "'\\n\\n\\8\\n';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 6, EndLine: 1, EndColumn: 8,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\n\\n8\\n';"},
					{MessageId: "escapeBackslash", Output: "'\\n\\n\\\\8\\n';"},
				},
			}},
		},
		{
			name: "upstream: '\\\\8\\9'",
			code: "'\\\\8\\9';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 5, EndLine: 1, EndColumn: 7,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\\\89';"},
					{MessageId: "escapeBackslash", Output: "'\\\\8\\\\9';"},
				},
			}},
		},
		{
			name: "upstream: '\\8\\\\9'",
			code: "'\\8\\\\9';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 2, EndLine: 1, EndColumn: 4,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'8\\\\9';"},
					{MessageId: "escapeBackslash", Output: "'\\\\8\\\\9';"},
				},
			}},
		},
		{
			name: "upstream: 'foo\\8bar\\9baz' — two reports far apart",
			code: "'foo\\8bar\\9baz';",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 5, EndLine: 1, EndColumn: 7,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "'foo8bar\\9baz';"},
						{MessageId: "escapeBackslash", Output: "'foo\\\\8bar\\9baz';"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 10, EndLine: 1, EndColumn: 12,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "'foo\\8bar9baz';"},
						{MessageId: "escapeBackslash", Output: "'foo\\8bar\\\\9baz';"},
					},
				},
			},
		},
		{
			name: "upstream: '\\8\\1\\9' — \\1 between, second hit gets prev=\\1",
			code: "'\\8\\1\\9';",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 2, EndLine: 1, EndColumn: 4,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "'8\\1\\9';"},
						{MessageId: "escapeBackslash", Output: "'\\\\8\\1\\9';"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 6, EndLine: 1, EndColumn: 8,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "'\\8\\19';"},
						{MessageId: "escapeBackslash", Output: "'\\8\\1\\\\9';"},
					},
				},
			},
		},
		{
			name: "upstream: '\\1\\0\\8' — \\0 immediately precedes \\8 → special",
			code: "'\\1\\0\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 6, EndLine: 1, EndColumn: 8,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\1\\u00008';"},
					{MessageId: "refactor", Output: "'\\1\\0\\u0038';"},
					{MessageId: "escapeBackslash", Output: "'\\1\\0\\\\8';"},
				},
			}},
		},
		{
			name: "upstream: '\\8\\0\\9' — first standard, second special",
			code: "'\\8\\0\\9';",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 2, EndLine: 1, EndColumn: 4,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "'8\\0\\9';"},
						{MessageId: "escapeBackslash", Output: "'\\\\8\\0\\9';"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 6, EndLine: 1, EndColumn: 8,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "'\\8\\u00009';"},
						{MessageId: "refactor", Output: "'\\8\\0\\u0039';"},
						{MessageId: "escapeBackslash", Output: "'\\8\\0\\\\9';"},
					},
				},
			},
		},
		{
			name: "upstream: '0\\8' — plain 0 (not \\0) before \\8 → standard",
			code: "'0\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 3, EndLine: 1, EndColumn: 5,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'08';"},
					{MessageId: "escapeBackslash", Output: "'0\\\\8';"},
				},
			}},
		},
		{
			name: "upstream: '\\\\0\\8' — escaped \\\\ + 0 + \\8 → standard",
			code: "'\\\\0\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 5, EndLine: 1, EndColumn: 7,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\\\08';"},
					{MessageId: "escapeBackslash", Output: "'\\\\0\\\\8';"},
				},
			}},
		},
		{
			name: "upstream: '\\01\\8' — octal \\01 between → standard",
			code: "'\\01\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 5, EndLine: 1, EndColumn: 7,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\018';"},
					{MessageId: "escapeBackslash", Output: "'\\01\\\\8';"},
				},
			}},
		},
		{
			name: "upstream: '\\0\\1\\8' — \\1 between → standard",
			code: "'\\0\\1\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 6, EndLine: 1, EndColumn: 8,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\0\\18';"},
					{MessageId: "escapeBackslash", Output: "'\\0\\1\\\\8';"},
				},
			}},
		},
		{
			name: "upstream: '\\👍\\8' — escaped emoji adjacent to \\8 (UTF-16 surrogate pair)",
			code: "'\\\xf0\x9f\x91\x8d\\8';",
			want: []expectedDiagnostic{{
				// '\👍' = '(1) \(2) <surrogate-pair occupies cols 3-4> \(5) 8(6)
				MessageId: "decimalEscape", Line: 1, Column: 5, EndLine: 1, EndColumn: 7,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\\xf0\x9f\x91\x8d8';"},
					{MessageId: "escapeBackslash", Output: "'\\\xf0\x9f\x91\x8d\\\\8';"},
				},
			}},
		},

		// ---- Multi-line strings (line continuations) — line 2 reports ----
		{
			name: "upstream: '\\\\\\r\\n\\9' — CRLF then \\9 on line 2",
			code: "'\\\r\n\\9';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 2, Column: 1, EndLine: 2, EndColumn: 3,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'\\\r\n9';"},
					{MessageId: "escapeBackslash", Output: "'\\\r\n\\\\9';"},
				},
			}},
		},
		{
			name: "upstream: 'foo\\\\\\nbar\\9baz' — line continuation in middle, \\9 on line 2",
			code: "'foo\\\nbar\\9baz';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 2, Column: 4, EndLine: 2, EndColumn: 6,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "'foo\\\nbar9baz';"},
					{MessageId: "escapeBackslash", Output: "'foo\\\nbar\\\\9baz';"},
				},
			}},
		},

		// ---- Multiple distinct strings in one program ----
		{
			name: "two separate strings each with one decimal escape",
			code: "var foo = '\\8'; bar('\\9');",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "var foo = '8'; bar('\\9');"},
						{MessageId: "escapeBackslash", Output: "var foo = '\\\\8'; bar('\\9');"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 22, EndLine: 1, EndColumn: 24,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "var foo = '\\8'; bar('9');"},
						{MessageId: "escapeBackslash", Output: "var foo = '\\8'; bar('\\\\9');"},
					},
				},
			},
		},

		// ---- Real TS / JS contexts ----
		{
			name: "real: object literal property value",
			code: "const obj = { msg: '\\8' };",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 21, EndLine: 1, EndColumn: 23,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "const obj = { msg: '8' };"},
					{MessageId: "escapeBackslash", Output: "const obj = { msg: '\\\\8' };"},
				},
			}},
		},
		{
			name: "real: array literal element",
			code: "const arr = ['\\8', '\\9'];",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 15, EndLine: 1, EndColumn: 17,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "const arr = ['8', '\\9'];"},
						{MessageId: "escapeBackslash", Output: "const arr = ['\\\\8', '\\9'];"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 21, EndLine: 1, EndColumn: 23,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "const arr = ['\\8', '9'];"},
						{MessageId: "escapeBackslash", Output: "const arr = ['\\8', '\\\\9'];"},
					},
				},
			},
		},
		{
			name: "real: function default parameter",
			code: "function f(x = '\\9') { return x; }",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 17, EndLine: 1, EndColumn: 19,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "function f(x = '9') { return x; }"},
					{MessageId: "escapeBackslash", Output: "function f(x = '\\\\9') { return x; }"},
				},
			}},
		},
		{
			name: "real: nested function body, class method",
			code: "class C { m() { return '\\8'; } }",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 25, EndLine: 1, EndColumn: 27,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "class C { m() { return '8'; } }"},
					{MessageId: "escapeBackslash", Output: "class C { m() { return '\\\\8'; } }"},
				},
			}},
		},
		{
			name: "real: deeply nested in arrow + ternary + call",
			code: "const f = (b: boolean) => b ? foo('\\8') : bar('\\9');",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 36, EndLine: 1, EndColumn: 38,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "const f = (b: boolean) => b ? foo('8') : bar('\\9');"},
						{MessageId: "escapeBackslash", Output: "const f = (b: boolean) => b ? foo('\\\\8') : bar('\\9');"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 48, EndLine: 1, EndColumn: 50,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "const f = (b: boolean) => b ? foo('\\8') : bar('9');"},
						{MessageId: "escapeBackslash", Output: "const f = (b: boolean) => b ? foo('\\8') : bar('\\\\9');"},
					},
				},
			},
		},
		{
			name: "real: TS as-expression around a string",
			code: "const x = '\\8' as string;",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "const x = '8' as string;"},
					{MessageId: "escapeBackslash", Output: "const x = '\\\\8' as string;"},
				},
			}},
		},
		{
			name: "real: TS satisfies operator",
			code: "const x = '\\9' satisfies string;",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "const x = '9' satisfies string;"},
					{MessageId: "escapeBackslash", Output: "const x = '\\\\9' satisfies string;"},
				},
			}},
		},
		{
			name: "real: parenthesized + as + non-null assertion stack",
			code: "const x = ((('\\8'! as any) as string));",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 15, EndLine: 1, EndColumn: 17,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "const x = ((('8'! as any) as string));"},
					{MessageId: "escapeBackslash", Output: "const x = ((('\\\\8'! as any) as string));"},
				},
			}},
		},
		{
			name: "real: import source with \\8 — also a string literal",
			code: "import x from './a\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 19, EndLine: 1, EndColumn: 21,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "import x from './a8';"},
					{MessageId: "escapeBackslash", Output: "import x from './a\\\\8';"},
				},
			}},
		},
		{
			name: "real: enum value with \\8",
			code: "enum E { A = '\\8' }",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 15, EndLine: 1, EndColumn: 17,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "enum E { A = '8' }"},
					{MessageId: "escapeBackslash", Output: "enum E { A = '\\\\8' }"},
				},
			}},
		},
		{
			name: "real: string concatenation, both sides have decimal escape",
			code: "const x = '\\8' + '\\9';",
			want: []expectedDiagnostic{
				{
					MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "const x = '8' + '\\9';"},
						{MessageId: "escapeBackslash", Output: "const x = '\\\\8' + '\\9';"},
					},
				},
				{
					MessageId: "decimalEscape", Line: 1, Column: 19, EndLine: 1, EndColumn: 21,
					Suggestions: []expectedSuggestion{
						{MessageId: "refactor", Output: "const x = '\\8' + '9';"},
						{MessageId: "escapeBackslash", Output: "const x = '\\8' + '\\\\9';"},
					},
				},
			},
		},
		{
			name: "real: JSX attribute string value (TSX)",
			code: "const e = <div title='\\8'></div>;",
			tsx:  true,
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 23, EndLine: 1, EndColumn: 25,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "const e = <div title='8'></div>;"},
					{MessageId: "escapeBackslash", Output: "const e = <div title='\\\\8'></div>;"},
				},
			}},
		},

		// ---- Tabs & varying indentation — column counts must reflect actual UTF-16 columns ----
		{
			name: "tab-indented \\8 — tab is 1 column",
			code: "\tvar s = '\\8';",
			want: []expectedDiagnostic{{
				MessageId: "decimalEscape", Line: 1, Column: 11, EndLine: 1, EndColumn: 13,
				Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "\tvar s = '8';"},
					{MessageId: "escapeBackslash", Output: "\tvar s = '\\\\8';"},
				},
			}},
		},
		// NOTE: BOM-prefixed files are intentionally NOT tested here — tsgo
		// strips the BOM from `SourceFile.Text()` while the original `code`
		// passed to ApplyRuleFixes still has it, so positions misalign by 3
		// bytes. This is a framework-level concern (every rule's fixes hit
		// the same offset shift), not specific to no-nonoctal-decimal-escape.

		// ---- Stress: many hits in one literal ----
		{
			name: "stress: 4 \\8 in one string — 4 reports",
			code: "var s = '\\8\\8\\8\\8';",
			want: []expectedDiagnostic{
				{MessageId: "decimalEscape", Line: 1, Column: 10, EndLine: 1, EndColumn: 12, Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '8\\8\\8\\8';"},
					{MessageId: "escapeBackslash", Output: "var s = '\\\\8\\8\\8\\8';"},
				}},
				{MessageId: "decimalEscape", Line: 1, Column: 12, EndLine: 1, EndColumn: 14, Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '\\88\\8\\8';"},
					{MessageId: "escapeBackslash", Output: "var s = '\\8\\\\8\\8\\8';"},
				}},
				{MessageId: "decimalEscape", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '\\8\\88\\8';"},
					{MessageId: "escapeBackslash", Output: "var s = '\\8\\8\\\\8\\8';"},
				}},
				{MessageId: "decimalEscape", Line: 1, Column: 16, EndLine: 1, EndColumn: 18, Suggestions: []expectedSuggestion{
					{MessageId: "refactor", Output: "var s = '\\8\\8\\88';"},
					{MessageId: "escapeBackslash", Output: "var s = '\\8\\8\\8\\\\8';"},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := runRuleLeniently(t, tt.code, tt.tsx)
			if len(diags) != len(tt.want) {
				t.Fatalf("got %d diagnostics, want %d (diags=%+v)", len(diags), len(tt.want), summarizeDiagnostics(diags))
			}
			for i, d := range diags {
				want := tt.want[i]
				if d.Message.Id != want.MessageId {
					t.Errorf("diag %d: messageId %q, want %q", i, d.Message.Id, want.MessageId)
				}
				if want.Description != "" && d.Message.Description != want.Description {
					t.Errorf("diag %d: description %q, want %q", i, d.Message.Description, want.Description)
				}
				lineIdx, colIdx := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, d.Range.Pos())
				endLineIdx, endColIdx := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, d.Range.End())
				gotLine, gotCol := lineIdx+1, int(colIdx)+1
				gotEndLine, gotEndCol := endLineIdx+1, int(endColIdx)+1
				if want.Line != gotLine || want.Column != gotCol || want.EndLine != gotEndLine || want.EndColumn != gotEndCol {
					t.Errorf("diag %d: pos (%d:%d-%d:%d), want (%d:%d-%d:%d)",
						i, gotLine, gotCol, gotEndLine, gotEndCol,
						want.Line, want.Column, want.EndLine, want.EndColumn)
				}
				var gotSuggestions []rule.RuleSuggestion
				if d.Suggestions != nil {
					gotSuggestions = *d.Suggestions
				}
				if len(gotSuggestions) != len(want.Suggestions) {
					t.Errorf("diag %d: got %d suggestions, want %d", i, len(gotSuggestions), len(want.Suggestions))
					continue
				}
				for j, sug := range gotSuggestions {
					if sug.Message.Id != want.Suggestions[j].MessageId {
						t.Errorf("diag %d sug %d: messageId %q, want %q", i, j, sug.Message.Id, want.Suggestions[j].MessageId)
					}
					gotOutput, _, _ := linter.ApplyRuleFixes(tt.code, []rule.RuleSuggestion{sug})
					if gotOutput != want.Suggestions[j].Output {
						t.Errorf("diag %d sug %d: output %q, want %q", i, j, gotOutput, want.Suggestions[j].Output)
					}
				}
			}
		})
	}
}

func runRuleLeniently(t *testing.T, code string, tsx bool) []rule.RuleDiagnostic {
	t.Helper()
	tmpDir := t.TempDir()
	fileName := "file.ts"
	if tsx {
		fileName = "file.tsx"
	}
	filePath := tspath.NormalizePath(filepath.Join(tmpDir, fileName))
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:       core.ScriptTargetESNext,
		Module:       core.ModuleKindESNext,
		SkipLibCheck: core.TSTrue,
	}, []string{filePath}, host)
	if err != nil {
		t.Fatalf("CreateProgramFromOptionsLenient: %v", err)
	}

	configured := linter.ConfiguredRule{
		Name:     NoNonoctalDecimalEscapeRule.Name,
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return NoNonoctalDecimalEscapeRule.Run(ctx, nil)
		},
	}

	var diags []rule.RuleDiagnostic
	linter.RunLinterInProgram(program, nil, nil, utils.ExcludePaths,
		func(sf *ast.SourceFile) []linter.ConfiguredRule {
			if sf.FileName() != filePath {
				return nil
			}
			return []linter.ConfiguredRule{configured}
		},
		false,
		func(d rule.RuleDiagnostic) { diags = append(diags, d) },
		nil,
		nil,
	)
	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Range.Pos() < diags[j].Range.Pos()
	})
	return diags
}

func summarizeDiagnostics(diags []rule.RuleDiagnostic) string {
	parts := make([]string, len(diags))
	for i, d := range diags {
		parts[i] = d.Message.Id + ":" + d.Message.Description
	}
	return strings.Join(parts, " | ")
}
