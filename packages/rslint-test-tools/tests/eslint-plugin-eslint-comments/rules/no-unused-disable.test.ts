/**
 * @author Toru Nagashima <https://github.com/mysticatea>
 *
 * Source: @eslint-community/eslint-plugin-eslint-comments v4.7.2
 *   tests/lib/rules/no-unused-disable.js
 *
 * ===========================================================================
 * ENTIRE RULE = KNOWN GAP — not portable to the rslint eslint-plugin runner.
 * ===========================================================================
 *
 * `no-unused-disable` is structurally different from every other rule in this
 * plugin. Upstream documents it directly (rule header):
 *
 *   "This rule is special. This rule patches `Linter#verify` method and:
 *      1. enables `reportUnusedDisableDirectives` option.
 *      2. verifies the code.
 *      3. converts `reportUnusedDisableDirectives` errors to `no-unused-disable`
 *         errors.
 *    So it cannot test with `eslint.RuleTester`."
 *
 * Accordingly the upstream *test* does NOT use `RuleTester` either: it spawns
 * the real `eslint` CLI over stdin with
 *   --plugin @eslint-community/eslint-comments
 *   --rule  @eslint-community/eslint-comments/no-unused-disable:error
 *   [--report-unused-disable-directives]
 * and asserts on the merged CLI message stream. The rule's behaviour depends on:
 *   • ESLint's core unused-disable-directive engine producing the base reports
 *     (the rule has NO `meta.messages` of its own — `messages: {}` — it relabels
 *     ESLint's own directive reports), and
 *   • inline `/*eslint no-undef:off*​/` config comments turning core rules on/off
 *     so a disable directive becomes "unused".
 *
 * rslint's eslint-plugin runner provides neither: it loads plugin rules into a
 * Node worker pool over a Go-parsed AST, does NOT patch `Linter#verify`, exposes
 * no `--report-unused-disable-directives` plumbing into plugin rules, and does
 * not let an inline `/*eslint <core-rule>:off*​/` comment drive ESLint core rule
 * execution. Measured directly through the rslint CLI + `plugins`, every
 * upstream case — valid and invalid alike — yields ZERO
 * `eslint-comments/no-unused-disable` diagnostics, because the mechanism the
 * rule hooks into is absent.
 *
 * No upstream case is ported into a `ruleTester.run()` block: porting the VALID
 * cases would assert 0-diagnostic greens that pass for the WRONG reason (rslint
 * never ran the rule, vs. upstream where the directive was genuinely used), and
 * porting the INVALID cases would demand diagnostics rslint never emits.
 *
 * Instead, the single test below GUARDS the gap: it asserts rslint emits 0
 * diagnostics for a representative upstream INVALID case (where ESLint reports
 * "'no-undef' rule is disabled but never reported."). If rslint ever implements
 * the `Linter#verify` patch / `reportUnusedDisableDirectives` bridge, this test
 * flips red and signals that the upstream cases (preserved verbatim below) must
 * be ported for real. The expectation is measured, not fabricated.
 *
 * ---------------------------------------------------------------------------
 * Upstream VALID (expected 0 messages from the full ESLint CLI run):
 *   /*eslint no-undef:error*​/\nvar a = b //eslint-disable-line
 *   /*eslint no-undef:error*​/\nvar a = b /*eslint-disable-line*​/
 *   /*eslint no-undef:error*​/\nvar a = b //eslint-disable-line no-undef
 *   /*eslint no-undef:error*​/\nvar a = b /*eslint-disable-line no-undef*​/
 *   /*eslint no-undef:error, no-unused-vars:error*​/\nvar a = b //eslint-disable-line no-undef,no-unused-vars
 *   /*eslint no-undef:error, no-unused-vars:error*​/\nvar a = b /*eslint-disable-line no-undef,no-unused-vars*​/
 *   /*eslint no-undef:error*​/\n//eslint-disable-next-line\nvar a = b
 *   /*eslint no-undef:error*​/\n/*eslint-disable-next-line*​/\nvar a = b
 *   /*eslint no-undef:error*​/\n//eslint-disable-next-line no-undef\nvar a = b
 *   /*eslint no-undef:error*​/\n/*eslint-disable-next-line no-undef*​/\nvar a = b
 *   /*eslint no-undef:error, no-unused-vars:error*​/\n//eslint-disable-next-line no-undef,no-unused-vars\nvar a = b
 *   /*eslint no-undef:error, no-unused-vars:error*​/\n/*eslint-disable-next-line no-undef,no-unused-vars*​/\nvar a = b
 *   /*eslint no-undef:error*​/\n/*eslint-disable*​/\nvar a = b
 *   /*eslint no-undef:error*​/\n/*eslint-disable no-undef*​/\nvar a = b
 *   /*eslint no-undef:error, no-unused-vars:error*​/\n/*eslint-disable no-undef,no-unused-vars*​/\nvar a = b
 *   /*eslint no-undef:error*​/\n/*eslint-disable*​/\nvar a = b\n/*eslint-enable*​/
 *   /*eslint no-undef:error*​/\n/*eslint-disable no-undef*​/\nvar a = b\n/*eslint-enable no-undef*​/
 *   /*eslint no-undef:error, no-unused-vars:error*​/\n/*eslint-disable no-undef,no-unused-vars*​/\nvar a = b\n/*eslint-enable no-undef*​/
 *   (no-shadow inside nested functions, two `//eslint-disable-line no-shadow` — both block & line forms)
 *   /*eslint no-undef:error*​/\nvar a = b //eslint-disable-line -- description   (gated >=7.0.0)
 *
 * Upstream INVALID (each `message` is generated by the patched verify path, not
 * by `meta.messages`):
 *   "ESLint rules are disabled but never reported."          (whole disable, no ruleId)
 *   "'<rule>' rule is disabled but never reported."          (per-rule disable)
 *   "Unused eslint-disable directive (no problems were reported)."            (core, --report-unused-disable-directives)
 *   "Unused eslint-disable directive (no problems were reported from '<rule>')." (core, per-rule)
 *   "Parsing error: Unexpected token c"                      (parse-error passthrough)
 *   …with line/column/endLine/endColumn and `suggestions` ("Remove `eslint-disable` comment.")
 *   exactly as in the upstream file. None are reproducible without the patched
 *   `Linter#verify` + core `reportUnusedDisableDirectives` engine.
 * ---------------------------------------------------------------------------
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// The whole rule is a KNOWN GAP (see header). rslint's eslint-plugin runner
// does not patch `Linter#verify` / `reportUnusedDisableDirectives`, so EVERY
// upstream case — valid and invalid alike — yields 0
// `eslint-comments/no-unused-disable` diagnostics. We guard that absence through
// the SHARED RuleTester (no private CLI driver): representative upstream cases
// are listed as `valid` (the shared tester asserts 0 diagnostics for valid). It
// is an honest, measured gap-guard, NOT a correctness claim — the day rslint
// bridges the mechanism these flip to reporting and the `valid` assertion fails,
// forcing a real port of the upstream expectations preserved in the header.
ruleTester.run('no-unused-disable', null as never, {
  valid: [
    // upstream INVALID — whole-file disable with a rule turned off; ESLint's
    // patched verify reports "'no-undef' rule is disabled but never reported.",
    // rslint emits nothing.
    '/*eslint no-undef:off*/\nvar a = b /*eslint-disable-line no-undef*/',
    '/*eslint no-undef:error*/\n/*eslint-disable*/\nvar a = b\n/*eslint-enable*/',
  ],
  invalid: [],
});
