package interactive_supports_focus

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// recommendedOptions / strictOptions mirror the values that
// `eslint-plugin-jsx-a11y`'s shared configs hand to the rule. The upstream
// `interactive-supports-focus:recommended` and `:strict` test suites parametrize
// every generated case with these.
//
// Source: eslint-plugin-jsx-a11y/src/index.js — `configs.recommended.rules`
// and `configs.strict.rules` entries for `jsx-a11y/interactive-supports-focus`.
var (
	recommendedTabbable = []string{
		"button", "checkbox", "link", "searchbox", "spinbutton", "switch", "textbox",
	}
	strictTabbable = []string{
		"button", "checkbox", "link", "progressbar", "searchbox", "slider", "spinbutton", "switch", "textbox",
	}
)

func recommendedOptions() map[string]interface{} {
	return map[string]interface{}{"tabbable": stringSliceToAny(recommendedTabbable)}
}

func strictOptions() map[string]interface{} {
	return map[string]interface{}{"tabbable": stringSliceToAny(strictTabbable)}
}

func stringSliceToAny(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

// interactiveRolesForTests mirrors upstream's `interactiveRoles` constant in
// `__tests__/src/rules/interactive-supports-focus-test.js` (a hand-picked
// subset of widget-descendant roles — NOT every role in
// `interactiveRolesSet`). Order matches upstream byte-for-byte.
var interactiveRolesForTests = []string{
	"button",
	"checkbox",
	"link",
	"gridcell",
	"menuitem",
	"menuitemcheckbox",
	"menuitemradio",
	"option",
	"radio",
	"searchbox",
	"slider",
	"spinbutton",
	"switch",
	"tab",
	"textbox",
	"treeitem",
}

// triggeringHandlers mirrors upstream's `triggeringHandlers` (mouse +
// keyboard event handlers — the same set the rule itself inspects via
// `hasAnyProp(attributes, interactiveProps)`).
var triggeringHandlers = []string{
	"onClick", "onContextMenu", "onDblClick", "onDoubleClick",
	"onDrag", "onDragEnd", "onDragEnter", "onDragExit",
	"onDragLeave", "onDragOver", "onDragStart", "onDrop",
	"onMouseDown", "onMouseEnter", "onMouseLeave", "onMouseMove",
	"onMouseOut", "onMouseOver", "onMouseUp",
	"onKeyDown", "onKeyPress", "onKeyUp",
}

// allEventHandlers mirrors `jsx-ast-utils`' top-level `eventHandlers`
// (every category flat-mapped together). Used by upstream's "non-triggering
// handler keeps the element valid" assertion — we walk the difference
// `allEventHandlers - triggeringHandlers` to generate the non-mouse /
// non-keyboard event names that should NOT trip the rule.
var allEventHandlers = []string{
	// clipboard
	"onCopy", "onCut", "onPaste",
	// composition
	"onCompositionEnd", "onCompositionStart", "onCompositionUpdate",
	// keyboard
	"onKeyDown", "onKeyPress", "onKeyUp",
	// focus
	"onFocus", "onBlur",
	// form
	"onChange", "onInput", "onSubmit",
	// mouse
	"onClick", "onContextMenu", "onDblClick", "onDoubleClick",
	"onDrag", "onDragEnd", "onDragEnter", "onDragExit",
	"onDragLeave", "onDragOver", "onDragStart", "onDrop",
	"onMouseDown", "onMouseEnter", "onMouseLeave", "onMouseMove",
	"onMouseOut", "onMouseOver", "onMouseUp",
	// selection
	"onSelect",
	// touch
	"onTouchCancel", "onTouchEnd", "onTouchMove", "onTouchStart",
	// ui
	"onScroll",
	// wheel
	"onWheel",
	// media
	"onAbort", "onCanPlay", "onCanPlayThrough", "onDurationChange",
	"onEmptied", "onEncrypted", "onEnded", "onError", "onLoadedData",
	"onLoadedMetadata", "onLoadStart", "onPause", "onPlay", "onPlaying",
	"onProgress", "onRateChange", "onSeeked", "onSeeking", "onStalled",
	"onSuspend", "onTimeUpdate", "onVolumeChange", "onWaiting",
	// image
	"onLoad", // onError already in media
	// animation
	"onAnimationStart", "onAnimationEnd", "onAnimationIteration",
	// transition
	"onTransitionEnd",
}

func nonTriggeringHandlers() []string {
	out := make([]string, 0, len(allEventHandlers))
	for _, h := range allEventHandlers {
		if !slices.Contains(triggeringHandlers, h) {
			out = append(out, h)
		}
	}
	return out
}

// componentsSettings mirrors upstream's `componentsSettings` constant —
// `Div` resolves to the lowercase HTML `div` so that `<Div ... />` is
// treated as a div by `getElementType`.
var componentsSettings = map[string]interface{}{
	"jsx-a11y": map[string]interface{}{
		"components": map[string]interface{}{
			"Div": "div",
		},
	},
}

func codeTemplate(element, role, handler string) string {
	return fmt.Sprintf("<%s role=\"%s\" %s={() => void 0} />", element, role, handler)
}

func fixedTemplate(element, tabIndex, role, handler string) string {
	return fmt.Sprintf("<%s tabIndex={%s} role=\"%s\" %s={() => void 0} />", element, tabIndex, role, handler)
}

func tabindexTemplate(element, role, handler string) string {
	return fmt.Sprintf("<%s role=\"%s\" %s={() => void 0} tabIndex=\"0\" />", element, role, handler)
}

func tabbableMessage(role string) string {
	return fmt.Sprintf("Elements with the '%s' interactive role must be tabbable.", role)
}

func focusableMessage(role string) string {
	return fmt.Sprintf("Elements with the '%s' interactive role must be focusable.", role)
}

func failCases(roles, handlers []string, message func(string) string) []rule_tester.InvalidTestCase {
	var out []rule_tester.InvalidTestCase
	for _, element := range []string{"div"} {
		for _, role := range roles {
			for _, handler := range handlers {
				code := codeTemplate(element, role, handler)
				suggestions := []rule_tester.InvalidTestCaseSuggestion{
					{
						MessageId: "tabIndexZero",
						Output:    fixedTemplate(element, "0", role, handler),
					},
				}
				msgId := "tabbable"
				if isFocusable := !messageIsTabbable(message); isFocusable {
					suggestions = append(suggestions, rule_tester.InvalidTestCaseSuggestion{
						MessageId: "tabIndexNegOne",
						Output:    fixedTemplate(element, "-1", role, handler),
					})
					msgId = "focusable"
				}
				out = append(out, rule_tester.InvalidTestCase{
					Code: code,
					Tsx:  true,
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId:   msgId,
						Message:     message(role),
						Line:        1,
						Column:      1,
						Suggestions: suggestions,
					}},
				})
			}
		}
	}
	return out
}

// messageIsTabbable detects which template was passed to failCases without
// having to plumb a second argument. We inspect the rendered text — tabbable
// vs focusable messages differ by exactly one word in a fixed position.
func messageIsTabbable(message func(string) string) bool {
	sample := message("button")
	return strings.Contains(sample, "must be tabbable")
}

func passCases(roles, handlers []string, codeFn func(string, string, string) string) []rule_tester.ValidTestCase {
	var out []rule_tester.ValidTestCase
	for _, element := range []string{"div"} {
		for _, role := range roles {
			for _, handler := range handlers {
				out = append(out, rule_tester.ValidTestCase{
					Code: codeFn(element, role, handler),
					Tsx:  true,
				})
			}
		}
	}
	return out
}

// alwaysValid mirrors upstream's `alwaysValid` array verbatim — these cases
// are valid under EVERY option combination.
var alwaysValid = []rule_tester.ValidTestCase{
	{Code: `<div />`, Tsx: true},
	{Code: `<div aria-hidden onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={true == true} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={true === true} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={hidden !== false} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={hidden != false} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={1 < 2} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={1 <= 2} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={2 > 1} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div aria-hidden={2 >= 1} onClick={() => void 0} />`, Tsx: true},
	{Code: `<div onClick={() => void 0} />;`, Tsx: true},
	{Code: `<div onClick={() => void 0} tabIndex={undefined} />;`, Tsx: true},
	{Code: `<div onClick={() => void 0} tabIndex="bad" />;`, Tsx: true},
	{Code: `<div onClick={() => void 0} role={undefined} />;`, Tsx: true},
	{Code: `<div role="section" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div onClick={() => void 0} aria-hidden={false} />;`, Tsx: true},
	{Code: `<div onClick={() => void 0} {...props} />;`, Tsx: true},
	{Code: `<input type="text" onClick={() => void 0} />`, Tsx: true},
	{Code: `<input type="hidden" onClick={() => void 0} tabIndex="-1" />`, Tsx: true},
	{Code: `<input type="hidden" onClick={() => void 0} tabIndex={-1} />`, Tsx: true},
	{Code: `<input onClick={() => void 0} />`, Tsx: true},
	{Code: `<input onClick={() => void 0} role="combobox" />`, Tsx: true},
	{Code: `<button onClick={() => void 0} className="foo" />`, Tsx: true},
	{Code: `<option onClick={() => void 0} className="foo" />`, Tsx: true},
	{Code: `<select onClick={() => void 0} className="foo" />`, Tsx: true},
	{Code: `<area href="#" onClick={() => void 0} className="foo" />`, Tsx: true},
	{Code: `<area onClick={() => void 0} className="foo" />`, Tsx: true},
	{Code: `<summary onClick={() => void 0} />`, Tsx: true},
	{Code: `<textarea onClick={() => void 0} className="foo" />`, Tsx: true},
	{Code: `<a onClick="showNextPage();">Next page</a>`, Tsx: true},
	{Code: `<a onClick="showNextPage();" tabIndex={undefined}>Next page</a>`, Tsx: true},
	{Code: `<a onClick="showNextPage();" tabIndex="bad">Next page</a>`, Tsx: true},
	{Code: `<a onClick={() => void 0} />`, Tsx: true},
	{Code: `<a tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<a tabIndex={dynamicTabIndex} onClick={() => void 0} />`, Tsx: true},
	{Code: `<a tabIndex={0} onClick={() => void 0} />`, Tsx: true},
	{Code: `<a role="button" href="#" onClick={() => void 0} />`, Tsx: true},
	{Code: `<a onClick={() => void 0} href="http://x.y.z" />`, Tsx: true},
	{Code: `<a onClick={() => void 0} href="http://x.y.z" tabIndex="0" />`, Tsx: true},
	{Code: `<a onClick={() => void 0} href="http://x.y.z" tabIndex={0} />`, Tsx: true},
	{Code: `<a onClick={() => void 0} href="http://x.y.z" role="button" />`, Tsx: true},
	{Code: `<TestComponent onClick={doFoo} />`, Tsx: true},
	{Code: `<input onClick={() => void 0} type="hidden" />;`, Tsx: true},
	{Code: `<span onClick="submitForm();">Submit</span>`, Tsx: true},
	{Code: `<span onClick="submitForm();" tabIndex={undefined}>Submit</span>`, Tsx: true},
	{Code: `<span onClick="submitForm();" tabIndex="bad">Submit</span>`, Tsx: true},
	{Code: `<span onClick="doSomething();" tabIndex="0">Click me!</span>`, Tsx: true},
	{Code: `<span onClick="doSomething();" tabIndex={0}>Click me!</span>`, Tsx: true},
	{Code: `<span onClick="doSomething();" tabIndex="-1">Click me too!</span>`, Tsx: true},
	{Code: `<a href="javascript:void(0);" onClick="doSomething();">Click ALL the things!</a>`, Tsx: true},
	{Code: `<section onClick={() => void 0} />;`, Tsx: true},
	{Code: `<main onClick={() => void 0} />;`, Tsx: true},
	{Code: `<article onClick={() => void 0} />;`, Tsx: true},
	{Code: `<header onClick={() => void 0} />;`, Tsx: true},
	{Code: `<footer onClick={() => void 0} />;`, Tsx: true},
	{Code: `<div role="button" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="checkbox" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="link" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="menuitem" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="menuitemcheckbox" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="menuitemradio" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="option" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="radio" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="spinbutton" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="switch" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="tablist" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="tab" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="textbox" tabIndex="0" onClick={() => void 0} />`, Tsx: true},
	{Code: `<div role="textbox" aria-disabled="true" onClick={() => void 0} />`, Tsx: true},
	{Code: `<Foo.Bar onClick={() => void 0} aria-hidden={false} />;`, Tsx: true},
	{Code: `<Input onClick={() => void 0} type="hidden" />;`, Tsx: true},
	{Code: `<Div onClick={() => void 0} role="button" tabIndex="0" />`, Tsx: true, Settings: componentsSettings},
}

// neverValid mirrors upstream's `neverValid` — invalid under every option
// combination. The single entry leverages the `components` setting to map
// `<Div>` → `div` so a `role="button"` on it trips the rule.
func neverValid(options map[string]interface{}) []rule_tester.InvalidTestCase {
	return []rule_tester.InvalidTestCase{
		{
			Code:     `<Div onClick={() => void 0} role="button" />`,
			Tsx:      true,
			Settings: componentsSettings,
			Options:  options,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "tabbable",
				Message:   tabbableMessage("button"),
				Line:      1,
				Column:    1,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "tabIndexZero",
					Output:    `<Div tabIndex={0} onClick={() => void 0} role="button" />`,
				}},
			}},
		},
	}
}

// applyOptionsValid clones the input slice with the given `Options` field
// populated on each case. Mirrors upstream's `ruleOptionsMapperFactory`.
func applyOptionsValid(cases []rule_tester.ValidTestCase, options map[string]interface{}) []rule_tester.ValidTestCase {
	out := make([]rule_tester.ValidTestCase, len(cases))
	for i, c := range cases {
		c.Options = options
		out[i] = c
	}
	return out
}

func applyOptionsInvalid(cases []rule_tester.InvalidTestCase, options map[string]interface{}) []rule_tester.InvalidTestCase {
	out := make([]rule_tester.InvalidTestCase, len(cases))
	for i, c := range cases {
		c.Options = options
		out[i] = c
	}
	return out
}

// filterRoles returns `roles` minus any entries also present in `exclude`.
// Mirrors upstream's `.filter(role => !includes(exclude, role))`.
func filterRoles(roles, exclude []string) []string {
	out := make([]string, 0, len(roles))
	for _, r := range roles {
		if !slices.Contains(exclude, r) {
			out = append(out, r)
		}
	}
	return out
}

// TestInteractiveSupportsFocusUpstreamRecommended mirrors upstream's
// `interactive-supports-focus:recommended` suite — the `recommended` shared
// config's `tabbable` list, fed through the same reducer pipeline that
// generates every (element × role × handler) combination.
//
// The four reducer arms exercise the four observable diagnostic branches:
//
//   - alwaysValid                                                — pass on every config
//   - non-triggering handlers on every interactive role          — pass (rule short-circuits at hasInteractiveProps)
//   - non-recommended interactive roles with tabIndex="0"        — pass (hasTabindex true)
//   - recommendedRoles with triggering handlers                  — fail w/ tabbable message
//   - non-recommended interactiveRoles with triggering handlers  — fail w/ focusable message
//   - neverValid (Div + componentsSettings + role="button")      — fail
func TestInteractiveSupportsFocusUpstreamRecommended(t *testing.T) {
	opts := recommendedOptions()
	valid := append([]rule_tester.ValidTestCase{}, alwaysValid...)
	valid = append(valid, passCases(interactiveRolesForTests, nonTriggeringHandlers(), codeTemplate)...)
	valid = append(valid, passCases(filterRoles(interactiveRolesForTests, recommendedTabbable), triggeringHandlers, tabindexTemplate)...)
	valid = applyOptionsValid(valid, opts)

	invalid := append([]rule_tester.InvalidTestCase{}, neverValid(opts)...)
	invalid = append(invalid, applyOptionsInvalid(failCases(recommendedTabbable, triggeringHandlers, tabbableMessage), opts)...)
	invalid = append(invalid, applyOptionsInvalid(failCases(filterRoles(interactiveRolesForTests, recommendedTabbable), triggeringHandlers, focusableMessage), opts)...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &InteractiveSupportsFocusRule, valid, invalid)
}

// TestInteractiveSupportsFocusUpstreamStrict mirrors upstream's
// `interactive-supports-focus:strict` suite — same shape as the
// recommended suite with the `strict` config's broader `tabbable` list.
func TestInteractiveSupportsFocusUpstreamStrict(t *testing.T) {
	opts := strictOptions()
	valid := append([]rule_tester.ValidTestCase{}, alwaysValid...)
	valid = append(valid, passCases(interactiveRolesForTests, nonTriggeringHandlers(), codeTemplate)...)
	valid = append(valid, passCases(filterRoles(interactiveRolesForTests, strictTabbable), triggeringHandlers, tabindexTemplate)...)
	valid = applyOptionsValid(valid, opts)

	invalid := append([]rule_tester.InvalidTestCase{}, neverValid(opts)...)
	invalid = append(invalid, applyOptionsInvalid(failCases(strictTabbable, triggeringHandlers, tabbableMessage), opts)...)
	invalid = append(invalid, applyOptionsInvalid(failCases(filterRoles(interactiveRolesForTests, strictTabbable), triggeringHandlers, focusableMessage), opts)...)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &InteractiveSupportsFocusRule, valid, invalid)
}
