package no_invalid_remove_event_listener_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_invalid_remove_event_listener "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_invalid_remove_event_listener"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoInvalidRemoveEventListenerExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch, Dimension 4 row, or real-user
// scenario it covers.
func TestNoInvalidRemoveEventListenerExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_invalid_remove_event_listener.NoInvalidRemoveEventListenerRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: static and dynamic element-access method keys are excluded ----
			tsValid(`target['removeEventListener']('click', () => {})`),
			tsValid("target[`removeEventListener`]('click', () => {})"),
			tsValid(`target[0]('click', () => {})`),
			tsValid(`target[Symbol.removeEventListener]('click', () => {})`),

			// ---- Dimension 4: optional outer call and parenthesized optional-chain callee are excluded ----
			tsValid(`target.removeEventListener?.('click', () => {})`),
			tsValid(`(target.removeEventListener)?.('click', () => {})`),
			tsValid(`(target?.removeEventListener)('click', () => {})`),

			// ---- Dimension 4: TypeScript wrappers around the listener remain distinct expressions upstream ----
			tsValid(`target.removeEventListener('click', (() => {}) as EventListener)`),
			tsValid(`target.removeEventListener('click', handler.bind(target) as EventListener)`),
			tsValid(`target.removeEventListener('click', handler.bind(target)!)`),

			// ---- Dimension 4: optional/computed bind forms are excluded ----
			tsValid(`target.removeEventListener('click', handler.bind?.(target))`),
			tsValid(`target.removeEventListener('click', handler?.bind(target))`),
			tsValid(`target.removeEventListener('click', handler?.foo.bind(target))`),
			tsValid(`target.removeEventListener('click', handler?.foo.bar.bind(target))`),
			tsValid(`target.removeEventListener('click', handler['bind'](target))`),
			tsValid("target.removeEventListener('click', handler[`bind`](target))"),
			tsValid(`target.removeEventListener('click', handler[bind](target))`),

			// ---- Dimension 4: spread and empty argument forms degrade without reporting ----
			tsValid(`target.removeEventListener(...arguments, () => {})`),
			tsValid(`target.removeEventListener()`),

			// Locks in upstream isMethodCall() rejection branches for non-call and wrong-name shapes.
			tsValid(`new target.removeEventListener('click', () => {})`),
			tsValid(`target.removeListener('click', () => {})`),

			// N/A: declaration/container key forms, class/function container variants, body-absent declarations, and ancestor walks; the rule only inspects call expressions and their first two arguments.
			// N/A: autofix boundaries; the rule has no fix or suggestion.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized listener and bind callees are transparent ----
			tsInvalid(`target.removeEventListener('click', ((event) => {}))`, `=>`, 1),
			tsInvalid(`target.removeEventListener('click', (handler.bind(target)))`, `bind`, 1),
			tsInvalid(`target.removeEventListener('click', (handler.bind)(target))`, `bind`, 1),
			tsInvalid(`target.removeEventListener('click', ((handler).bind(target)))`, `bind`, 1),
			tsInvalid(`target.removeEventListener('click', (handler?.foo).bind(target))`, `bind`, 1),

			// ---- Dimension 4: parenthesized and TypeScript-wrapped receivers remain reportable ----
			tsInvalid(`(target).removeEventListener('click', () => {})`, `=>`, 1),
			tsInvalid(`((target)).removeEventListener('click', () => {})`, `=>`, 1),
			tsInvalid(`target!.removeEventListener('click', () => {})`, `=>`, 1),
			tsInvalid(`(target as EventTarget).removeEventListener('click', () => {})`, `=>`, 1),
			tsInvalid(`(target satisfies EventTarget).removeEventListener('click', () => {})`, `=>`, 1),

			// ---- Dimension 4: optional receiver member remains reportable ----
			// ---- Real-user: #2253 optional receiver regression ----
			tsInvalid(`this.input?.removeEventListener(event, this.onInputEvent.bind(this), true)`, `bind`, 1),

			// ---- Dimension 4: function expression variants all report their function heads ----
			tsInvalid(`target.removeEventListener('click', function* () {})`, `function* `, 1),
			tsInvalid(`target.removeEventListener('click', async function* named() {})`, `async function* named`, 1),

			// ---- Dimension 4: later spread arguments do not hide the listener ----
			tsInvalid(`target.removeEventListener('click', () => {}, ...options)`, `=>`, 1),

			// ---- Dimension 4: nested matching calls report independently without traversal bleed ----
			{
				Code:     `outer.removeEventListener('outer', () => { inner.removeEventListener('inner', () => {}) })`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(`outer.removeEventListener('outer', () => { inner.removeEventListener('inner', () => {}) })`, `=>`, 1),
					expectedError(`outer.removeEventListener('outer', () => { inner.removeEventListener('inner', () => {}) })`, `=>`, 2),
				},
			},

			// ---- Real-user: #682 custom-element lifecycle removal with a fresh bound listener ----
			tsInvalid(`class MyClass extends HTMLElement {
	connectedCallback() {
		this.addEventListener('click', this.handleClick.bind(this));
	}
	disconnectedCallback() {
		this.removeEventListener('click', this.handleClick.bind(this));
	}
	handleClick() {}
}`, `bind`, 2),

			// Locks in upstream listener arm 1: an arrow reports at its arrow token.
			tsInvalid(`target.removeEventListener('click', async event => event)`, `=>`, 1),

			// Locks in upstream listener arm 2: a function expression reports its head, not its body.
			tsInvalid(`target.removeEventListener('click', function named(event) { return event })`, `function named`, 1),

			// Locks in upstream listener arm 3: a direct non-optional dot-property bind call reports only `bind`.
			tsInvalid(`target.removeEventListener('click', createHandler().bind(target))`, `bind`, 1),
		},
	)
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func tsInvalid(code string, target string, occurrence int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.ts",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, occurrence),
		},
	}
}
