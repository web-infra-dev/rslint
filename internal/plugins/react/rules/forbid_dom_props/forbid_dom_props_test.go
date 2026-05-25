package forbid_dom_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestForbidDomPropsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForbidDomPropsRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		// `id` forbidden but `<Foo>` is a Component, not a DOM node.
		{
			Code: `
        var First = createReactClass({
          render: function() {
            return <Foo id="foo" />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Multiple forbid entries — none apply to a Component.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo id="bar" style={{color: "red"}} />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"style", "id"}},
		},
		// `<this.Foo>` is a member-expression tag — `typeof tag === 'string'`
		// is false in upstream → skipped. (The user prop `bar` isn't in forbid
		// either, but the upstream-skip is what's locked in.)
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <this.Foo bar="baz" />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// `<this.foo>` — even though the rightmost segment is lowercase, the
		// tag is a member expression and upstream's `parent.name.name` is
		// undefined → skipped. Locks in that member-expression tags are NEVER
		// classified as DOM here.
		{
			Code: `
        class First extends createReactClass {
          render() {
            return <this.foo id="bar" />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Spread on a member-expression tag — no JsxAttribute to match.
		{
			Code: `
        const First = (props) => (
          <this.Foo {...props} />
        );
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// JSX namespaced name `<fbt:param>` — upstream's tag is a
		// JSXIdentifier object (truthy but not a string); the `typeof` check
		// rejects it and the listener exits before checking `name`.
		{
			Code: `
        const First = (props) => (
          <fbt:param name="name">{props.name}</fbt:param>
        );
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// DOM tag, but `name` is not in the forbid list.
		{
			Code: `
        const First = (props) => (
          <div name="foo" />
        );
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// `disallowedFor` doesn't include `div`, so the prop is allowed.
		{
			Code: `
        const First = (props) => (
          <div otherProp="bar" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "otherProp",
					"disallowedFor": []interface{}{"span"},
				},
			}},
		},
		// `disallowedValues: []` — explicit empty array. Upstream's truthy
		// check on the array reference passes; `[].indexOf(value)` always
		// returns -1, so nothing matches.
		{
			Code: `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "someProp",
					"disallowedValues": []interface{}{},
				},
			}},
		},
		// Component (Foo) with disallowedValues — Component is skipped before
		// values are checked.
		{
			Code: `
        const First = (props) => (
          <Foo someProp="someValue" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "someProp",
					"disallowedValues": []interface{}{"someValue"},
				},
			}},
		},
		// DOM tag, but the value doesn't match the disallowedValues list.
		{
			Code: `
        const First = (props) => (
          <div someProp="value" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "someProp",
					"disallowedValues": []interface{}{"someValue"},
				},
			}},
		},
		// `disallowedFor: ['span']` — `<div>` doesn't match, so the value
		// match alone isn't enough to forbid.
		{
			Code: `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "someProp",
					"disallowedValues": []interface{}{"someValue"},
					"disallowedFor":    []interface{}{"span"},
				},
			}},
		},

		// ---- Additional edge cases (Dimension 4 universal edge shapes) ----
		// No options at all → defaults to empty forbid list → no diagnostics.
		{Code: `<div id="x" className="y" />;`, Tsx: true},
		// Explicit empty forbid → no diagnostics, even on a DOM intrinsic.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{}},
		},
		// Spread-only DOM tag — no JsxAttribute, so nothing to report.
		{
			Code:    `<div {...props} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Spread + an unrelated named attribute on a DOM tag.
		{
			Code:    `<div {...props} name="ok" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Boolean shorthand `<div hidden />` — upstream `node.value.value`
		// is `null.value`, which throws TypeError and produces no
		// diagnostic for that attribute. We exit on absent initializer to
		// reach the same observable "no diagnostic" outcome.
		//
		// Variants below cover propName-only forbid, disallowedValues, and
		// `disallowedFor` — all must be silent because the listener exits
		// before any `isForbidden` evaluation.
		{
			Code:    `<div hidden />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"hidden"}},
		},
		{
			Code: `<div hidden />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "hidden",
					"disallowedValues": []interface{}{"true"},
				},
			}},
		},
		{
			Code: `<div hidden />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "hidden", "disallowedFor": []interface{}{"div"}},
			}},
		},
		// Expression initializer on a DOM tag — upstream reads
		// `node.value.value` as `undefined`, so disallowedValues never matches.
		{
			Code: `<div id={someId} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "id",
					"disallowedValues": []interface{}{"x"},
				},
			}},
		},
		// `<Foo.bar>` — member expression with a lowercase rightmost
		// segment. Member-expression tags are skipped before the lowercase
		// check, so the prop is not flagged.
		{
			Code:    `<Foo.bar id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// `<Foo.Bar.Baz>` — deeper member chain, also skipped.
		{
			Code:    `<Foo.Bar.Baz className="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"className"}},
		},
		// Empty-string entries in the forbid list are skipped; a real entry
		// alongside still applies — but `bar` is not the real entry.
		{
			Code:    `<div bar="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"", "id"}},
		},
		// `forbid` not an array (schema-violating): falls back to the empty
		// default → nothing forbidden.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": "id"},
		},
		// Object entry without `propName`/`propNamePattern` — silently
		// skipped (upstream creates an entry keyed by `undefined`, which
		// the listener never queries).
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"disallowedFor": []interface{}{"div"}},
			}},
		},
		// JSX Fragment containing DOM elements — every intrinsic is checked
		// independently of fragment wrapping.
		{
			Code:    `<><div className="ok" /><span className="ok" /></>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Component inside a parent attribute expression — only DOM children
		// are checked; nested `<Inner>` is a Component.
		{
			Code:    `<div data-x={(<Inner id="ok" />)} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Custom message but the rule doesn't fire (Component) — empty
		// message stays unused.
		{
			Code: `<Foo className="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "Avoid className",
				},
			}},
		},
		// Empty custom-message string — falls back to the default messageId
		// path (mirrors upstream's truthy `customMessage || ...`). Component
		// here, so no diagnostic anyway.
		{
			Code: `<Foo className="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "",
				},
			}},
		},
		// Non-string entries (numbers, nulls, nested arrays) — silently
		// skipped. The remaining entry restricts only `id`.
		{
			Code:    `<div className="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{42, nil, []interface{}{"className"}, "id"}},
		},

		// ---- TS-specific syntax & wrappers (Dimension 1) ----
		// TS generic on a Component tag — DOM check skips it.
		{
			Code:    `function W<T>() { return <Foo<T> id="x" />; }`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// TS `as` cast inside a JSX expression — value side, not inspected.
		{
			Code:    `<div data-x={("ok" as string)} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},

		// ---- Whitespace & comment robustness (Dimension 4) ----
		// Tag immediately followed by self-close: `<div/>` (no space).
		{
			Code:    `<div className="ok"/>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Multi-line attributes on a DOM intrinsic — only prop names matter.
		{
			Code: `
        <div
          name="ok"
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Block comment between tag and attribute — should not break parsing.
		{
			Code:    `<div /* hi */ name="ok" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},

		// ---- Defaults & options edge shapes ----
		// `null` options → empty forbid → no diagnostic.
		{Code: `<div id="x" />;`, Tsx: true, Options: nil},
		// Options array shape `[opts]` — exercises GetOptionsMap unwrap path.
		{
			Code:    `<div name="ok" />;`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"forbid": []interface{}{"id"}}},
		},
		// Options object shape `opts` — single-option CLI path.
		{
			Code:    `<div name="ok" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},

		// ---- Real-world configurations ----
		// Forbid `style` on every DOM intrinsic; only Components remain.
		{
			Code: `
        <Foo style={{}}>
          <Bar style={{}} />
        </Foo>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"style"}},
		},
		// JSX namespaced attribute name (`xlink:href`) — upstream's
		// `node.name.name` returns the inner JSXIdentifier object (not a
		// string), so `forbid.get(<object>)` never matches and the rule
		// silently exits. Locks in: a user-supplied `xlink:href` string in
		// `forbid` MUST NOT pair up with a JSXNamespacedName attribute.
		{
			Code:    `<svg><use xlink:href="#icon" /></svg>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"xlink:href"}},
		},
		// `accept` only forbidden on `<form>` — `<input>` is allowed.
		{
			Code: `
        const First = (props) => (
          <input type="file" accept="video/*" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "accept",
					"disallowedFor": []interface{}{"form"},
					"message":       "Avoid using the accept attribute on <form>",
				},
			}},
		},

		// ---- Additional Go-port robustness coverage ----
		// JSX namespaced TAG with lowercase-starting namespace
		// (`<svg:path>`-style). Upstream's `typeof tag === 'string'` is false
		// for JSXNamespacedName regardless of casing, so it's skipped — even
		// if a user puts `"svg:path"` in `forbid`. Locks in: namespaced
		// element names never enter the DOM-check path.
		{
			Code:    `<svg:path id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id", "svg:path"}},
		},
		// JSX namespaced TAG with uppercase-starting namespace (`<X:Y>`).
		// Same path as above — skipped before the case check.
		{
			Code:    `<X:Y id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Object entry with empty-string `propName` — upstream keys the Map
		// on `""`. We skip (treat as malformed) so an attribute literally
		// named `""` (impossible in valid JSX) doesn't accidentally match.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "", "disallowedFor": []interface{}{"div"}},
			}},
		},
		// `disallowedFor: []` explicit empty array — upstream `!disallowList`
		// is false (empty array is truthy), and `[].indexOf(tag)` is -1, so
		// the rule never fires for any tag. Mirror that.
		{
			Code: `<div className="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "className", "disallowedFor": []interface{}{}},
			}},
		},
		// Empty `{}` forbid entry — neither propName nor propNamePattern;
		// silently skipped.
		{
			Code:    `<div className="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{}}},
		},
		// `disallowedValues` is case-sensitive — upstream's `indexOf` uses
		// strict equality. `<div id="X">` does NOT match `disallowedValues:
		// ["x"]`.
		{
			Code: `<div id="X" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"x"}},
			}},
		},
		// Multi-byte tag name (Unicode identifier — legal JSX). The tag is
		// 'デ' which is uncased; `IsCasedLowercaseFirstLetter` returns false
		// → skipped (matches upstream where `'デ'.toUpperCase() === 'デ'`).
		{
			Code:    `<デ id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// `<div>` containing string children + JsxExpression child — neither
		// affects attribute checking.
		{
			Code: `
        <div data-x="ok">
          {someExpression}
          static text
        </div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Conditional render where both branches produce DOM intrinsics
		// without the forbidden prop — neither branch fires.
		{
			Code:    `const cond = true; const x = cond ? <div name="a" /> : <span name="b" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// `Array.prototype.map` arrow body returning a DOM intrinsic with a
		// non-forbidden prop.
		{
			Code:    `const xs: any[] = []; xs.map(x => <div data-x={x} />);`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// JSX Fragment as the root expression (no enclosing element).
		{
			Code:    `<><div name="a" /><span name="b" /></>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// `disallowedValues: [""]` matching attribute with empty-string value
		// is technically possible — but `<div id="" />` is the only way to
		// reach it. Verifies the empty-string round-trips through the
		// disallow check WITHOUT firing when the user value isn't empty.
		{
			Code: `<div id="not-empty" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{""}},
			}},
		},
		// Same prop appears more than once on the same element (legal JSX
		// even though React warns). Each attribute walks the listener
		// independently — neither here matches the forbid list.
		{
			Code:    `<div name="a" name="b" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Attribute names starting with `$` or `_` — not in forbid, no
		// diagnostic. Locks in that the listener doesn't normalize names.
		{
			Code:    `<div $foo="x" _bar="y" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"foo", "bar"}},
		},
		// JsxElement-as-attribute-value (`<div foo=<Bar />/>` form). Rare but
		// legal in some JSX parsers; tsgo accepts it. Initializer is
		// JsxSelfClosingElement, NOT a string literal, so propValue stays
		// undefined and disallowedValues never matches. Locks in graceful
		// degradation rather than crash.
		{
			Code: `<div foo=<Bar /> />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "foo", "disallowedValues": []interface{}{"x"}},
			}},
		},
		// Both `disallowedFor` and `disallowedValues` empty arrays — both
		// conjuncts fail (`[].indexOf(x) === -1`), so the rule never fires
		// even though the propName matches.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "id",
					"disallowedFor":    []interface{}{},
					"disallowedValues": []interface{}{},
				},
			}},
		},

		// ---- Initializer / value-shape coverage ----
		// `KindNoSubstitutionTemplateLiteral` initializer (rare; some JSX
		// parsers reject, tsgo accepts in error recovery). We don't treat
		// it as a string value, so `disallowedValues` doesn't match.
		{
			Code: "<div id={`x`} />;",
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"x"}},
			}},
		},
		// JsxExpression with a numeric literal — `node.value.value` is
		// undefined upstream; disallowedValues never matches.
		{
			Code: `<div tabIndex={0} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "tabIndex", "disallowedValues": []interface{}{"0"}},
			}},
		},
		// JsxExpression containing a string literal — `node.value.value`
		// undefined (JsxExpression has `.expression`, not `.value`); same
		// as above. Locks in that brace-wrapped string is NOT the bare
		// string-attribute path.
		{
			Code: `<div id={"x"} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"x"}},
			}},
		},
		// JsxExpression containing a non-null assertion — TS-specific
		// wrapper inside the value position; still propValue undefined.
		{
			Code: `const v: any = "x"; <div id={v!} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"x"}},
			}},
		},

		// ---- Tag-name AST shape coverage ----
		// `<this id="x" />` — bare ThisKeyword is not KindIdentifier; we
		// don't match it (upstream `node.parent.name.name` would also miss).
		{
			Code:    `<this id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Mixed-namespace lowercase tag with uppercase right segment
		// (`<svg:Symbol>`) — namespaced still skipped.
		{
			Code:    `<svg:Symbol id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},

		// ---- Forbid-entry shape coverage ----
		// `propName` whitespace-only — upstream creates entry keyed by
		// "  ", which never matches a real attribute name.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "  "},
			}},
		},
		// `disallowedValues` containing non-string elements (numbers, null) —
		// upstream JS `[1, null].indexOf(propValue)` for string propValue
		// returns -1; our `stringSlice` filters non-strings, equivalent
		// outcome.
		{
			Code: `<div id="1" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{1, nil, true}},
			}},
		},
		// `forbid` field literally `null` (vs absent) — JS `configuration.forbid || DEFAULTS`
		// falls back to `[]` (empty default for forbid-dom-props). No diagnostic.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": nil},
		},
		// Multiple forbid entries; only one matches the current attr —
		// confirms unrelated entries don't cause false positives.
		{
			Code: `<div className="ok" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				"id", "style",
				map[string]interface{}{"propName": "title", "disallowedFor": []interface{}{"div"}},
			}},
		},
		// Many forbid entries (≥10) — exercises the map dispatch
		// (constant-time per attr) instead of upstream's linear Map iteration.
		{
			Code: `<div data-x="ok" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				"id", "className", "style", "title", "lang", "dir", "role",
				"hidden", "tabIndex", "draggable",
			}},
		},

		// ---- Real-world configurations ----
		// Locale enforcement: forbid `lang` on inline elements; permit it
		// on block-level (here we test the inline case as valid because
		// disallowedFor doesn't contain `span`).
		{
			Code: `<span data-i18n="ok" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "lang",
					"disallowedFor": []interface{}{"em", "strong", "code"},
				},
			}},
		},
		// Multiple Component-vs-DOM transitions in a single tree where
		// neither side carries the forbidden prop.
		{
			Code: `
        <Layout>
          <header>
            <Logo />
            <Nav role="navigation" />
          </header>
          <main>
            <Article id="article-id" />
          </main>
        </Layout>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Library-user pattern: forbid `style` on every DOM tag, but the
		// app actually uses `<canvas style={ctx}/>` correctly per allowance.
		{
			Code: `
        const Painter = () => (
          <canvas width={100} height={100} />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "style",
					"disallowedFor": []interface{}{"div", "span", "section", "article"},
				},
			}},
		},
		// `<img>` with `alt` required-style use case — alt allowed, src
		// forbidden as data-URL (disallowedValues by prefix… but upstream
		// uses exact match, so a literal data-URL value test).
		{
			Code: `<img alt="logo" src="https://example.com/logo.png" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "src",
					"disallowedFor":    []interface{}{"img"},
					"disallowedValues": []interface{}{"data:"},
				},
			}},
		},
		// `<input>` form-control safety: forbid `autoComplete="off"` on
		// password fields specifically, allow elsewhere.
		{
			Code: `<input type="email" autoComplete="email" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "autoComplete",
					"disallowedFor":    []interface{}{"input"},
					"disallowedValues": []interface{}{"off"},
				},
			}},
		},

		// ---- Additional shape coverage ----
		// JsxAttribute.Initializer being a JsxFragment — `<div foo=<></> />`.
		// Initializer kind is JsxFragment, not StringLiteral; propValue
		// undefined; disallowedValues never matches. Listener exits before
		// `isForbidden` evaluates the value path.
		{
			Code: `<div foo=<></> />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "foo", "disallowedValues": []interface{}{"x"}},
			}},
		},
		// `disallowedValues` containing duplicates — upstream's
		// `indexOf` still finds the first hit; `slices.Contains` is
		// equivalent. Locks in: schema's `uniqueItems` violation is
		// silently tolerated.
		{
			Code:    `<div className="ok" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Prop name case-sensitivity: `<div ID="x" />` does NOT match
		// `forbid: ['id']` — exact-string lookup.
		{
			Code:    `<div ID="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Tag `<div>` with TS type-argument-position `<T>` — tsgo parses
		// this as JsxOpeningElement with TypeArguments; tagName remains
		// KindIdentifier "div". Without forbidden prop, no diagnostic.
		// (We can't construct an invalid scenario here without ambiguity,
		// since `<div<T>>` is uncommon real-world; valid case is enough
		// to lock the parse path.)
		{
			Code:    `function W<T>() { return <div<T> name="ok" />; }`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
		// Mixed value-shape on a single element: string literal +
		// expression + boolean shorthand + spread — none in forbid list,
		// all four shapes coexist without false positives.
		{
			Code:    `<div className="a" data-x={1} hidden {...rest} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		// `<div id="bar">` inside a createReactClass.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <div id="bar" />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 25},
			},
		},
		// Same, inside a class component.
		{
			Code: `
        class First extends createReactClass {
          render() {
            return <div id="bar" />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 4, Column: 25},
			},
		},
		// Same, inside an arrow component.
		{
			Code: `
        const First = (props) => (
          <div id="foo" />
        );
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 16},
			},
		},
		// Custom message replaces the default; no messageId emitted.
		{
			Code: `
        const First = (props) => (
          <div className="foo" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "Please use class instead of ClassName",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Please use class instead of ClassName", Line: 3, Column: 16},
			},
		},
		// `disallowedFor` includes `span` — `<span>` flagged.
		{
			Code: `
        const First = (props) => (
          <span otherProp="bar" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "otherProp",
					"disallowedFor": []interface{}{"span"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 17},
			},
		},
		// `disallowedValues` matches — propIsForbiddenWithValue messageId.
		{
			Code: `
        const First = (props) => (
          <div someProp="someValue" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "someProp",
					"disallowedValues": []interface{}{"someValue"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbiddenWithValue", Line: 3, Column: 16},
			},
		},
		// Two custom-message entries on nested elements.
		{
			Code: `
        const First = (props) => (
          <div className="foo">
            <div otherProp="bar" />
          </div>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "Please use class instead of ClassName",
				},
				map[string]interface{}{
					"propName": "otherProp",
					"message":  "Avoid using otherProp",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Please use class instead of ClassName", Line: 3, Column: 16},
				{Message: "Avoid using otherProp", Line: 4, Column: 18},
			},
		},
		// Mixed: one default message (no custom), one custom.
		{
			Code: `
        const First = (props) => (
          <div className="foo">
            <div otherProp="bar" />
          </div>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "className"},
				map[string]interface{}{
					"propName": "otherProp",
					"message":  "Avoid using otherProp",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 16},
				{Message: "Avoid using otherProp", Line: 4, Column: 18},
			},
		},
		// `accept` forbidden only on `<form>` — single error on the form.
		{
			Code: `
        const First = (props) => (
          <form accept='file'>
            <input type="file" id="videoFile" accept="video/*" />
            <input type="hidden" name="fullname" />
          </form>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "accept",
					"disallowedFor": []interface{}{"form"},
					"message":       "Avoid using the accept attribute on <form>",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Avoid using the accept attribute on <form>", Line: 3, Column: 17},
			},
		},
		// `className` disallowed for `div` and `span`; otherProp forbidden
		// universally. Three matches across a nested structure.
		{
			Code: `
        const First = (props) => (
          <div className="foo">
            <input className="boo" />
            <span className="foobar">Foobar</span>
            <div otherProp="bar" />
          </div>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"div", "span"},
					"message":       "Please use class instead of ClassName",
				},
				map[string]interface{}{
					"propName": "otherProp",
					"message":  "Avoid using otherProp",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Please use class instead of ClassName", Line: 3, Column: 16},
				{Message: "Please use class instead of ClassName", Line: 5, Column: 19},
				{Message: "Avoid using otherProp", Line: 6, Column: 18},
			},
		},
		// Complex multi-prop scenario combining propName, disallowedFor,
		// disallowedValues, and a custom message.
		{
			Code: `
        const First = (props) => (
          <div className="foo">
            <input className="boo" />
            <span className="foobar">Foobar</span>
            <div otherProp="bar" />
            <p thirdProp="foo" />
            <div thirdProp="baz" />
            <p thirdProp="bar" />
            <p thirdProp="baz" />
          </div>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"div", "span"},
					"message":       "Please use class instead of ClassName",
				},
				map[string]interface{}{
					"propName": "otherProp",
					"message":  "Avoid using otherProp",
				},
				map[string]interface{}{
					"propName":         "thirdProp",
					"disallowedFor":    []interface{}{"p"},
					"disallowedValues": []interface{}{"bar", "baz"},
					"message":          "Do not use thirdProp with values bar and baz on p",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Please use class instead of ClassName", Line: 3, Column: 16},
				{Message: "Please use class instead of ClassName", Line: 5, Column: 19},
				{Message: "Avoid using otherProp", Line: 6, Column: 18},
				{Message: "Do not use thirdProp with values bar and baz on p", Line: 9, Column: 16},
				{Message: "Do not use thirdProp with values bar and baz on p", Line: 10, Column: 16},
			},
		},

		// ---- Additional lock-in cases ----
		// `propIsForbiddenWithValue` description includes both prop and value
		// in the formatted message. Locks in the substitution path.
		{
			Code:    `<div someProp="someValue" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{map[string]interface{}{"propName": "someProp", "disallowedValues": []interface{}{"someValue"}}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "someProp" with value "someValue" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},
		// `propIsForbidden` description substitutes the prop name.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbidden",
					Message:   `Prop "id" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},
		// Spread + multiple forbidden named attrs on a DOM tag — each
		// reports independently.
		{
			Code:    `<div {...rest} id="x" className="y" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id", "className"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 16},
				{MessageId: "propIsForbidden", Line: 1, Column: 23},
			},
		},
		// Last-write-wins on duplicate `forbid` entries with the same propName.
		// First entry has a custom message; second entry has none — second
		// wins, so the default messageId is reported.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "message": "First"},
				map[string]interface{}{"propName": "id"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// `disallowedValues` set, value matches: prop+value reported. Locks
		// in the data substitution for both placeholders.
		{
			Code: `<div id="bad" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"bad"}},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "id" with value "bad" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},

		// ---- Reporting position robustness (Go-vs-ESLint ECMA columns) ----
		// JSXOpeningElement (`<div>`) variant — listener fires on the
		// attribute, not the opening tag.
		{
			Code:    `<div id="x">child</div>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// Indented multi-line attribute — column counts from the attribute's
		// own line start.
		{
			Code: `
        <div
          id="x"
        />;
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 11},
			},
		},
		// Tab-indented attribute — UTF-16 column counting.
		{
			Code:    "<div\n\tid=\"x\"\n/>;",
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 2},
			},
		},
		// Multi-byte (CJK) child content — column on the attribute line is
		// not affected by the multi-byte payload elsewhere in the file.
		{
			Code: `
        const t = "中文";
        <div id={t} />;
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 14},
			},
		},

		// ---- Additional Go-port robustness cases ----
		// EndLine / EndColumn assertion — confirms the diagnostic range
		// covers the entire JsxAttribute node (`id="x"`), not just the prop
		// identifier. Locks in the report range against future refactors.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbidden",
					Line:      1, Column: 6,
					EndLine: 1, EndColumn: 12,
				},
			},
		},
		// `disallowedValues: [""]` matches `<div id="" />` exactly — locks
		// in that empty-string equality round-trips through the conjunct.
		{
			Code: `<div id="" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{""}},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "id" with value "" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},
		// Same prop attribute appearing twice on the same element — both
		// fire independently. Locks in that the listener doesn't dedupe.
		{
			Code:    `<div id="a" id="b" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
				{MessageId: "propIsForbidden", Line: 1, Column: 13},
			},
		},
		// Multiple `disallowedValues` candidates — first match wins, exact
		// equality.
		{
			Code: `<div id="bar" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "id",
					"disallowedValues": []interface{}{"foo", "bar", "baz"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "id" with value "bar" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},
		// `disallowedFor` includes the same tag listed multiple times — no
		// dedupe needed since `slices.Contains` short-circuits on first hit.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "id",
					"disallowedFor": []interface{}{"div", "div", "span"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// String entry alongside object entry for the SAME propName —
		// upstream Map's last-write-wins applies. Object entry with custom
		// message comes second → its message wins.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				"id",
				map[string]interface{}{"propName": "id", "message": "Object wins"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Object wins", Line: 1, Column: 6},
			},
		},
		// Reverse order: object first, string second → string entry overrides
		// the message. Locks in upstream Map semantics on heterogeneous
		// duplicates.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "message": "Object loses"},
				"id",
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// Conditional render with one branch producing a forbidden DOM tag.
		{
			Code:    `const cond = true; const x = cond ? <div id="a" /> : <span name="b" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 42},
			},
		},
		// `Array.prototype.map` arrow body — JSX inside still walks the
		// listener.
		{
			Code:    `const xs: any[] = []; xs.map(x => <div id={x} />);`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 40},
			},
		},
		// JSX Fragment as the root — DOM intrinsics inside still get
		// reported.
		{
			Code:    `<><div id="a" /><span id="b" /></>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 8},
				{MessageId: "propIsForbidden", Line: 1, Column: 23},
			},
		},
		// Real-world: design system enforcing `style` ban on most DOM tags
		// while allowing `<canvas>` (where inline styles are sometimes
		// unavoidable). Mixed result over many siblings.
		{
			Code: `
        <section>
          <div style={{}} />
          <canvas style={{}} />
          <span style={{}} />
        </section>
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "style",
					"disallowedFor": []interface{}{"div", "span", "section"},
					"message":       "Inline style banned on layout primitives",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Inline style banned on layout primitives", Line: 3, Column: 16},
				{Message: "Inline style banned on layout primitives", Line: 5, Column: 17},
			},
		},
		// Real-world: forbid `target="_blank"` on anchors specifically (a
		// security-style rule via disallowedValues). Confirms value match
		// reports propIsForbiddenWithValue.
		{
			Code: `
        <a href="/x" target="_blank">click</a>;
        <a href="/y" target="_self">click</a>;
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "target",
					"disallowedFor":    []interface{}{"a"},
					"disallowedValues": []interface{}{"_blank"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "target" with value "_blank" is forbidden on DOM Nodes`,
					Line:      2, Column: 22,
				},
			},
		},
		// Component-as-attribute-value: `<div data-x={(<div id="x" />)} />`.
		// The OUTER `<div data-x>` is OK (data-x not forbidden), but the
		// INNER nested `<div id>` IS reported. Locks in that nested DOM
		// elements inside attribute expressions still walk the listener.
		{
			Code:    `<div data-x={(<div id="x" />)} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 20},
			},
		},
		// Many sibling DOM intrinsics under a Component parent: each DOM
		// child reports independently; the Component parent is skipped even
		// if it carries the same forbidden prop.
		{
			Code: `
        <Group className="parent">
          <Foo className="ok-component" />
          <span className="bad" />
          <Bar className="ok-component-2" />
          <i className="bad-i" />
          <Baz className="ok-component-3" />
        </Group>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"className"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 4, Column: 17},
				{MessageId: "propIsForbidden", Line: 6, Column: 14},
			},
		},
		// `<div hidden={true} />` — boolean expression form. Initializer
		// is JsxExpression containing a BooleanLiteral; propValue is
		// undefined per upstream's `node.value.value` (JsxExpression has
		// `.expression`, not `.value`). propName-only forbid still fires.
		{
			Code:    `<div hidden={true} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"hidden"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// Custom message with `{{prop}}` placeholder — upstream
		// `report(context, message, ...)` only runs ESLint's interpolation
		// engine on the messages-table path (i.e. when resolved via
		// messageId). A directly-passed `customMessage` stays literal —
		// placeholders are NOT substituted. Locks in this exact behavior.
		{
			Code: `<div className="foo" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "Prop {{prop}} is not allowed",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Message: "Prop {{prop}} is not allowed",
					Line:    1, Column: 6,
				},
			},
		},
		// Custom message with `{{prop}}` and `{{propValue}}` placeholders
		// AND a triggering disallowedValues — placeholders still stay
		// literal because the custom path bypasses the messages table.
		{
			Code: `<div id="bad" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "id",
					"disallowedValues": []interface{}{"bad"},
					"message":          "{{prop}}={{propValue}} forbidden",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Message: "{{prop}}={{propValue}} forbidden",
					Line:    1, Column: 6,
				},
			},
		},
		// Mixed `disallowedFor` with disallowedValues — only the row matching
		// BOTH conjuncts fires.
		{
			Code: `
        <input type="text" />;
        <input type="password" />;
        <input type="hidden" />;
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "type",
					"disallowedFor":    []interface{}{"input"},
					"disallowedValues": []interface{}{"password", "hidden"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "type" with value "password" is forbidden on DOM Nodes`,
					Line:      3, Column: 16,
				},
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "type" with value "hidden" is forbidden on DOM Nodes`,
					Line:      4, Column: 16,
				},
			},
		},

		// ---- More invalid edge / real-world cases ----
		// disallowedValues on the same element — multi-attribute, only the
		// matching value triggers `propIsForbiddenWithValue`.
		{
			Code: `<input type="hidden" name="csrf" autoComplete="off" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "autoComplete",
					"disallowedFor":    []interface{}{"input"},
					"disallowedValues": []interface{}{"off"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "autoComplete" with value "off" is forbidden on DOM Nodes`,
					Line:      1, Column: 34,
				},
			},
		},
		// `<a target="_blank" rel="noopener" href="..."/>` real-world
		// security pattern — flag _blank, allow the rest. Confirms the
		// listener visits attributes in source order and only reports the
		// matching one.
		{
			Code: `<a target="_blank" rel="noopener" href="/x">click</a>;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "target",
					"disallowedFor":    []interface{}{"a"},
					"disallowedValues": []interface{}{"_blank"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "target" with value "_blank" is forbidden on DOM Nodes`,
					Line:      1, Column: 4,
				},
			},
		},
		// `<img>` with literal `src` matching disallowed value — used as
		// a stand-in for image-asset policy.
		{
			Code: `<img alt="x" src="data:" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "src",
					"disallowedFor":    []interface{}{"img"},
					"disallowedValues": []interface{}{"data:"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "src" with value "data:" is forbidden on DOM Nodes`,
					Line:      1, Column: 14,
				},
			},
		},
		// Many forbid entries; only the matching one fires. Locks in that
		// the dispatch is by-key, not linear-scan-with-side-effects.
		{
			Code: `<div title="bad" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				"id", "className", "style", "lang", "dir", "role",
				map[string]interface{}{"propName": "title", "message": "Use aria-label instead"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Use aria-label instead", Line: 1, Column: 6},
			},
		},
		// disallowedValues with a non-trivial value containing special
		// chars — locks in exact-string matching (no regex / glob on values).
		{
			Code: `<div data-tracking="user/123 click" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "data-tracking",
					"disallowedValues": []interface{}{"user/123 click"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "data-tracking" with value "user/123 click" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},
		// disallowedValues with Unicode value content.
		{
			Code: `<div data-l10n="中文" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "data-l10n",
					"disallowedValues": []interface{}{"中文"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "data-l10n" with value "中文" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},
		// Custom message with empty disallowedValues string match.
		{
			Code: `<div className="" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "className",
					"disallowedValues": []interface{}{""},
					"message":          "Empty className is suspicious",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Empty className is suspicious", Line: 1, Column: 6},
			},
		},
		// JsxFragment NESTED inside a Component, with DOM intrinsics inside
		// — confirms the listener walks through both Component and Fragment
		// boundaries.
		{
			Code: `
        const T = () => (
          <Wrap>
            <>
              <div id="a" />
              <span id="b" />
            </>
          </Wrap>
        );
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 20},
				{MessageId: "propIsForbidden", Line: 6, Column: 21},
			},
		},
		// Switch-case-style logic emitting different DOM intrinsics in each
		// arm — locks in that nested ternary chains all walk independently.
		{
			Code: `
        const x: any = "A";
        const node =
          x === "A" ? <div id="a" />
          : x === "B" ? <span id="b" />
          : <p id="c" />;
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 4, Column: 28},
				{MessageId: "propIsForbidden", Line: 5, Column: 31},
				{MessageId: "propIsForbidden", Line: 6, Column: 16},
			},
		},
		// JSX inside a default-export arrow returning DOM directly.
		{
			Code:    `export default () => <div id="root" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 27},
			},
		},
		// Conditional with logical-AND short-circuit returning a DOM tag.
		{
			Code:    `const cond = true; const x = cond && <div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 43},
			},
		},
		// JSX inside an IIFE.
		{
			Code:    `const x = (() => <div id="iife" />)();`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 23},
			},
		},
		// Same propName as both a string entry and an object entry with
		// disallowedValues — last entry wins (Map semantics). The object
		// requires a specific value to trigger.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				"id",
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"x"}},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "id" with value "x" is forbidden on DOM Nodes`,
					Line:      1, Column: 6,
				},
			},
		},
		// Reverse: object first with disallowedValues, string second wins.
		// The string entry has no value constraint, so any value triggers
		// `propIsForbidden` (NOT WithValue).
		{
			Code: `<div id="other" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"x"}},
				"id",
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// `disallowedFor` listing many tags; multiple-DOM intrinsics in a
		// realistic markup chunk.
		{
			Code: `
        <article>
          <header className="a" />
          <section className="b" />
          <p>plain</p>
          <footer className="c" />
        </article>
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"header", "section", "footer"},
					"message":       "Layout primitives must be unstyled",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Layout primitives must be unstyled", Line: 3, Column: 19},
				{Message: "Layout primitives must be unstyled", Line: 4, Column: 20},
				{Message: "Layout primitives must be unstyled", Line: 6, Column: 19},
			},
		},
		// JSX with attribute values containing escape sequences — the
		// cooked text is what `node.value.value` returns. Locks in we use
		// the cooked, not raw, text.
		{
			Code: `<div id="a\nb" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{"a\\nb"}},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Line:      1, Column: 6,
				},
			},
		},
		// Unicode escape in attribute value (`"a"` cooks to `"a"`).
		// JSX attribute string literals don't process JS-style escapes —
		// JSX preserves the source verbatim, so cooked == raw. Locks in
		// that we use whatever tsgo gives us as `.Text`.
		{
			Code: `<div id="a" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "id", "disallowedValues": []interface{}{`a`}},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbiddenWithValue", Line: 1, Column: 6},
			},
		},
		// Truth table: disallowedFor + disallowedValues both set, both
		// match → forbidden with `propIsForbiddenWithValue`. Locks in the
		// AND-gate (both conjuncts must pass).
		{
			Code: `<span data-track="click" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "data-track",
					"disallowedFor":    []interface{}{"div", "span"},
					"disallowedValues": []interface{}{"click", "hover"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "propIsForbiddenWithValue",
					Message:   `Prop "data-track" with value "click" is forbidden on DOM Nodes`,
					Line:      1, Column: 7,
				},
			},
		},
		// Three independent forbid entries × three matching props on the
		// same element — verifies the listener runs once per attribute and
		// dispatches by propName.
		{
			Code: `<div id="x" className="y" data-test="z" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				"id", "className",
				map[string]interface{}{"propName": "data-test"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
				{MessageId: "propIsForbidden", Line: 1, Column: 13},
				{MessageId: "propIsForbidden", Line: 1, Column: 27},
			},
		},
		// `disallowedValues` with duplicate entries — schema-violation but
		// `slices.Contains` short-circuits on first hit. Locks in tolerance.
		{
			Code: `<div id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":         "id",
					"disallowedValues": []interface{}{"x", "x", "x"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbiddenWithValue", Line: 1, Column: 6},
			},
		},
		// Self-closing vs paired-element position equivalence: same
		// attribute on `<div id="x" />` and `<div id="x"></div>` should
		// report at the same column.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<div id="x"></div>;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// Component depth ≥4 with DOM intrinsics interleaved at every
		// level — locks in that the listener fires regardless of
		// ancestor-Component depth.
		{
			Code: `
        <A>
          <B>
            <C>
              <D>
                <div id="deep" />
              </D>
            </C>
          </B>
        </A>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 6, Column: 22},
			},
		},

		// ---- Deeply nested DOM intrinsics ----
		// Component / DOM mixed: only DOM children are reported.
		{
			Code: `
        <Outer id="ok">
          <div id="bad">
            <Inner id="ok2">
              <span id="bad2" />
            </Inner>
          </div>
        </Outer>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 16},
				{MessageId: "propIsForbidden", Line: 5, Column: 21},
			},
		},
	})
}

// TestForbidDomPropsOptionParsing exercises the JSON shapes that the CLI and
// JS rule tester send. Hand-rolled fallbacks that only handle one shape
// silently drift on the others — `GetOptionsMap` is the canonical extractor,
// but it only works if every option-typed field round-trips through it.
func TestForbidDomPropsOptionParsing(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForbidDomPropsRule, []rule_tester.ValidTestCase{
		// Bare object (single-option CLI shape).
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"className"}},
		},
		// Array-wrapped (multi-element rule_tester shape).
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"forbid": []interface{}{"className"}}},
		},
		// Nil options.
		{Code: `<div id="x" />;`, Tsx: true, Options: nil},
		// Empty options array.
		{Code: `<div id="x" />;`, Tsx: true, Options: []interface{}{}},
		// Malformed `forbid` (string, not array) — silently ignored, defaults apply.
		{Code: `<div id="x" />;`, Tsx: true, Options: map[string]interface{}{"forbid": "id"}},
	}, []rule_tester.InvalidTestCase{
		// Bare object with a real forbid list.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"id"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// Array-wrapped with a real forbid list.
		{
			Code:    `<div id="x" />;`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"forbid": []interface{}{"id"}}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
	})
}
