package mouse_events_have_key_events

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Upstream `mouseOverError` / `mouseOutError` / `pointerEnterError` /
// `pointerLeaveError` shape the expected diagnostic. We mirror with
// `Message` exact-string assertions so the rule's user-visible text is
// pinned. Position assertions are added on top of upstream because the
// upstream report writes `node: getProp(attributes, handler)` — i.e. the
// JsxAttribute, which always starts at column 6 (after `<div ` /
// `<{tag} `).
var (
	mouseOverErrorAtCol6 = rule_tester.InvalidTestCaseError{
		MessageId: "mouseOver",
		Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
		Line:      1,
		Column:    6,
	}
	mouseOutErrorAtCol6 = rule_tester.InvalidTestCaseError{
		MessageId: "mouseOut",
		Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
		Line:      1,
		Column:    6,
	}
)

// TestMouseEventsHaveKeyEventsUpstream covers the full valid/invalid
// suite migrated 1:1 from upstream eslint-plugin-jsx-a11y's
// `__tests__/src/rules/mouse-events-have-key-events-test.js`. Order
// inside each group mirrors the upstream file so a future audit can
// grep across both side-by-side.
//
// Anything NOT in upstream's test file — TS wrappers, position
// assertions on multi-element / multi-line code, extra spread shapes,
// extra edge cases — lives in mouse_events_have_key_events_extras_test.go.
func TestMouseEventsHaveKeyEventsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MouseEventsHaveKeyEventsRule,
		[]rule_tester.ValidTestCase{
			// ---- Default config — onMouseOver + onFocus pair, onMouseOut + onBlur pair. ----
			{Code: `<div onMouseOver={() => void 0} onFocus={() => void 0} />;`, Tsx: true},
			{Code: `<div onMouseOver={() => void 0} onFocus={() => void 0} {...props} />;`, Tsx: true},
			{Code: `<div onMouseOver={handleMouseOver} onFocus={handleFocus} />;`, Tsx: true},
			{Code: `<div onMouseOver={handleMouseOver} onFocus={handleFocus} {...props} />;`, Tsx: true},

			// ---- No mouse event handlers → no pairing required. ----
			{Code: `<div />;`, Tsx: true},
			{Code: `<div onBlur={() => {}} />`, Tsx: true},
			{Code: `<div onFocus={() => {}} />`, Tsx: true},

			// ---- onMouseOut + onBlur pair (mirror of the onMouseOver/onFocus group). ----
			{Code: `<div onMouseOut={() => void 0} onBlur={() => void 0} />`, Tsx: true},
			{Code: `<div onMouseOut={() => void 0} onBlur={() => void 0} {...props} />`, Tsx: true},
			{Code: `<div onMouseOut={handleMouseOut} onBlur={handleOnBlur} />`, Tsx: true},
			{Code: `<div onMouseOut={handleMouseOut} onBlur={handleOnBlur} {...props} />`, Tsx: true},

			// ---- Custom (non-DOM) component — rule short-circuits on
			//      `dom.get(name)` check; mouse events on custom components
			//      are out of scope (we don't know the rendered low-level DOM). ----
			{Code: `<MyElement />`, Tsx: true},
			{Code: `<MyElement onMouseOver={() => {}} />`, Tsx: true},
			{Code: `<MyElement onMouseOut={() => {}} />`, Tsx: true},
			{Code: `<MyElement onBlur={() => {}} />`, Tsx: true},
			{Code: `<MyElement onFocus={() => {}} />`, Tsx: true},
			{Code: `<MyElement onMouseOver={() => {}} {...props} />`, Tsx: true},
			{Code: `<MyElement onMouseOut={() => {}} {...props} />`, Tsx: true},
			{Code: `<MyElement onBlur={() => {}} {...props} />`, Tsx: true},
			{Code: `<MyElement onFocus={() => {}} {...props} />`, Tsx: true},

			// ---- Passing in empty options doesn't check any event handlers. ----
			{
				Code:    `<div onMouseOver={() => {}} onMouseOut={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{}, "hoverOutHandlers": []interface{}{}}},
			},

			// ---- Passing in custom handlers — explicit lists override the
			//      defaults. A handler listed here still requires its pair
			//      (onFocus for hover-in, onBlur for hover-out). ----
			{
				Code:    `<div onMouseOver={() => {}} onFocus={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{"onMouseOver"}}},
			},
			{
				Code:    `<div onMouseEnter={() => {}} onFocus={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{"onMouseEnter"}}},
			},
			{
				Code:    `<div onMouseOut={() => {}} onBlur={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverOutHandlers": []interface{}{"onMouseOut"}}},
			},
			{
				Code:    `<div onMouseLeave={() => {}} onBlur={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverOutHandlers": []interface{}{"onMouseLeave"}}},
			},
			{
				Code:    `<div onMouseOver={() => {}} onMouseOut={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{"onPointerEnter"}, "hoverOutHandlers": []interface{}{"onPointerLeave"}}},
			},

			// ---- Custom options only check the handlers passed in — the
			//      element's other mouse handlers are ignored for pairing. ----
			{
				Code:    `<div onMouseLeave={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverOutHandlers": []interface{}{"onPointerLeave"}}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Plain <div> with onMouseOver and no onFocus → reports. ----
			{
				Code:   `<div onMouseOver={() => void 0} />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			// ---- Plain <div> with onMouseOut and no onBlur → reports. ----
			{
				Code:   `<div onMouseOut={() => void 0} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},

			// ---- onFocus={undefined} — `getPropValue` extracts `undefined`,
			//      `!= null` is false → onFocus counted as absent → reports. ----
			{
				Code:   `<div onMouseOver={() => void 0} onFocus={undefined} />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			{
				Code:   `<div onMouseOut={() => void 0} onBlur={undefined} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},

			// ---- {...props} spread is opaque under FindAttributeByName's
			//      semantics — it cannot supply onFocus / onBlur. ----
			{
				Code:   `<div onMouseOver={() => void 0} {...props} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			{
				Code:   `<div onMouseOut={() => void 0} {...props} />`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},

			// ---- Custom options enabling both pairings: hover-in and hover-out
			//      both fire. Each report sits on its own attribute. ----
			{
				Code:    `<div onMouseOver={() => {}} onMouseOut={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{"onMouseOver"}, "hoverOutHandlers": []interface{}{"onMouseOut"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "mouseOver",
						Message:   "onMouseOver must be accompanied by onFocus for accessibility.",
						Line:      1,
						Column:    6,
					},
					{
						MessageId: "mouseOut",
						Message:   "onMouseOut must be accompanied by onBlur for accessibility.",
						Line:      1,
						Column:    29,
					},
				},
			},
			// ---- Custom pointer-event handlers — same flow, different names. ----
			{
				Code:    `<div onPointerEnter={() => {}} onPointerLeave={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{"onPointerEnter"}, "hoverOutHandlers": []interface{}{"onPointerLeave"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "mouseOver",
						Message:   "onPointerEnter must be accompanied by onFocus for accessibility.",
						Line:      1,
						Column:    6,
					},
					{
						MessageId: "mouseOut",
						Message:   "onPointerLeave must be accompanied by onBlur for accessibility.",
						Line:      1,
						Column:    32,
					},
				},
			},

			// ---- Custom options activating only one side — only that side reports. ----
			{
				Code:    `<div onMouseOver={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{"onMouseOver"}}},
				Errors:  []rule_tester.InvalidTestCaseError{mouseOverErrorAtCol6},
			},
			{
				Code:    `<div onPointerEnter={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverInHandlers": []interface{}{"onPointerEnter"}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOver",
					Message:   "onPointerEnter must be accompanied by onFocus for accessibility.",
					Line:      1,
					Column:    6,
				}},
			},
			{
				Code:    `<div onMouseOut={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverOutHandlers": []interface{}{"onMouseOut"}}},
				Errors:  []rule_tester.InvalidTestCaseError{mouseOutErrorAtCol6},
			},
			{
				Code:    `<div onPointerLeave={() => {}} />`,
				Tsx:     true,
				Options: []interface{}{map[string]interface{}{"hoverOutHandlers": []interface{}{"onPointerLeave"}}},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "mouseOut",
					Message:   "onPointerLeave must be accompanied by onBlur for accessibility.",
					Line:      1,
					Column:    6,
				}},
			},
		},
	)
}
