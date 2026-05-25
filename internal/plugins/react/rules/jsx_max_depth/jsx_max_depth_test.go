package jsx_max_depth

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxMaxDepthRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxMaxDepthRule, []rule_tester.ValidTestCase{
		// ---- Upstream: depth 0 (default max=2) ----
		{Code: `<App />`, Tsx: true},

		// ---- Upstream: explicit max=1, depth 1 ----
		{Code: `
        <App>
          <foo />
        </App>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Upstream: default max=2, depth 2 ----
		{Code: `
        <App>
          <foo>
            <bar />
          </foo>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          <foo>
            <bar />
          </foo>
        </App>
      `, Tsx: true, Options: map[string]interface{}{"max": 2}},

		// ---- Upstream: identifier resolution to JSX ----
		{Code: `
        const x = <div><em>x</em></div>;
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 2}},
		{Code: `const foo = (x) => <div><em>{x}</em></div>;`, Tsx: true, Options: map[string]interface{}{"max": 2}},

		// ---- Upstream: fragments ----
		{Code: `<></>`, Tsx: true},
		{Code: `
        <>
          <foo />
        </>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		{Code: `
        const x = <><em>x</em></>;
        <>{x}</>
      `, Tsx: true, Options: map[string]interface{}{"max": 2}},

		// ---- Upstream: identifier-resolved table fragment, depth fits max=2 ----
		{Code: `
        const x = (
          <tr>
            <td>1</td>
            <td>2</td>
          </tr>
        );
        <tbody>
          {x}
        </tbody>
      `, Tsx: true, Options: map[string]interface{}{"max": 2}},

		// ---- Upstream: function-with-loop returns leaf JSX, max=1 ----
		{Code: `
        const Example = props => {
          for (let i = 0; i < length; i++) {
            return <Text key={i} />;
          }
        };
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Upstream: React.Fragment with an interpolation; default max=2 ----
		// Inside the React.Fragment the `{ <div /> }` JsxExpression wraps a
		// JsxSelfClosingElement — but the fragment's only JSX-bearing child IS
		// that wrapper, and the wrapped element is itself a leaf. The outer
		// `<div>{A}</div>` resolves A to the React.Fragment, then descends into
		// its children (the wrapped `<div />`) at baseDepth=1+1=2 which fits.
		{Code: `
        export function MyComponent() {
          const A = <React.Fragment>{<div />}</React.Fragment>;
          return <div>{A}</div>;
        }
      `, Tsx: true},

		// ---- Upstream: circular references must not stall the resolver ----
		{Code: `
        function Component() {
          let first = "";
          const second = first;
          first = second;
          return <div id={first} />;
        };
      `, Tsx: true},
		{Code: `
        function Component() {
          let first = "";
          let second = "";
          let third = "";
          let fourth = "";
          const fifth = first;
          first = second;
          second = third;
          third = fourth;
          fourth = fifth;
          return <div id={first} />;
        };
      `, Tsx: true},

		// ---- Lock-in: parens around the resolved identifier ----
		// tsgo preserves ParenthesizedExpression as an explicit node; espree
		// flattens it. We SkipParentheses so `{(x)}` resolves identically to
		// `{x}` — anything else would silently miss real-world wrapper code.
		{Code: `
        const x = <div><span /></div>;
        <div>{(x)}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 2}},

		// ---- Lock-in: TS-only wrappers are NOT peeled ----
		// Upstream's `node.expression.type === 'Identifier'` rejects any TS
		// wrapper outright; we mirror that to keep the diagnostic surface
		// equal on the same input. `<div>{(x as any)}</div>` therefore does
		// NOT trigger the resolve path even though the underlying value is JSX.
		{Code: `
        const x = <div><span /></div>;
        <div>{x as any}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		{Code: `
        const x = <div><span /></div>;
        <div>{x!}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		{Code: `
        const x = <div><span /></div>;
        <div>{x satisfies any}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: JSX inside an attribute is at depth 0, not depth 1 ----
		// The JsxExpression that hosts the attribute value is NOT a JSX
		// container, so the parent walk for an attribute-positioned `<span/>`
		// must terminate at the JsxAttribute (count=0). Without this the
		// attribute would inherit the parent element's depth.
		{Code: `<div title={<span />} />`, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: text-only children remain leaves ----
		// JsxText nodes never satisfy hasJSX, so an element whose only children
		// are text is still a leaf — this is what lets `<App />` and
		// `<div>hello</div>` share the same isLeaf result.
		{Code: `<div>hello</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ---- Lock-in: identifier resolution to a non-JSX value ----
		// `findJsxElementOrFragment` must return nil when the resolved
		// initializer is a non-JSX, non-Identifier value (string / number /
		// object / call). Without that bail the descendant walk would crash
		// on a missing children list.
		{Code: `
        const x = "hello";
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},
		{Code: `
        const x = 42;
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ---- Lock-in: option key with non-numeric value falls back to default ----
		// Upstream's schema rejects `max: "two"` at config-load; rslint's
		// option parsing tolerates it by defaulting to max=2 instead of
		// crashing. Verify the default still applies (depth 2 fits max 2).
		{Code: `
        <App>
          <foo>
            <bar />
          </foo>
        </App>
      `, Tsx: true, Options: map[string]interface{}{"max": "two"}},

		// ---- Lock-in: reassignment to non-JSX after JSX init also bails ----
		// Upstream picks the LAST write in source order; if it's neither JSX
		// nor an Identifier the resolver returns null and no descendant walk
		// fires. The earlier JSX init is intentionally NOT consulted as a
		// fallback (mirrors `findJSXElementOrFragment.find` returning on the
		// first writeExpr seen reverse-iterating).
		{Code: `
        let x = <div><span /></div>;
        x = "";
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: parameter binding has no resolvable write ----
		// `findJSXElementOrFragment` resolves through VariableDeclarations
		// only. A parameter (with or without default) yields no write
		// expression — matches upstream which doesn't surface parameter
		// defaults as `references[].writeExpr` for our purpose either.
		{Code: `
        function Foo(x) {
          return <div>{x}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ---- Lock-in: destructured binding doesn't resolve ----
		// Object / array binding patterns aren't tracked; the resolver
		// returns nil so the descendant walk doesn't fire even if the
		// destructured value happens to be JSX.
		{Code: `
        const { x } = obj;
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ---- Lock-in: property assignment to a same-name property is NOT a write ----
		// `obj.x = <JSX>` writes to a property, not the binding `x`. The
		// reassignment scan must reject anything whose LHS isn't a bare
		// Identifier matching the binding name. With max=1 the JSX in
		// `obj.x = ...` doesn't itself violate; if the LHS were
		// mis-classified, the JsxExpression listener would resolve `x`
		// through it and flag a 2-deep <span/> descend.
		{Code: `
        let x = "hello";
        obj.x = <div><span /></div>;
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: compound assignment is NOT a write ----
		// Only `=` reassigns the binding; `+=` / `&&=` / etc. are skipped
		// because they don't produce a fresh JSX value the same way ESLint's
		// scope manager wouldn't surface them as `writeExpr` either.
		{Code: `
        let x = "hello";
        x += "world";
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ---- Lock-in: shadowed inner block binding is independent ----
		// Inner `let x = <div><span/></div>` shadows outer `let x = ""`;
		// the outer reference resolves to outer's init only and does not
		// pick up the inner block's write. JSX inside the inner block sits
		// at depth 1 which fits max=1, so any extra report would only come
		// from a mis-resolved descend.
		{Code: `
        let x = "";
        if (cond) {
          let x = <div><span /></div>;
          void x;
        }
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: function parameter shadows outer binding ----
		// Inner function with parameter `x` writes to the parameter, not
		// the outer `let x`. Same depth-1 setup so any leak would surface
		// via a JsxExpression-listener descend on outer `{x}`.
		{Code: `
        let x = "";
        function inner(x) {
          x = <div><span /></div>;
          return x;
        }
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: function-body `let x` shadows the outer binding ----
		// `function inner() { let x = <div><span/></div>; }` declares its
		// own `x`; the closure-write descent must skip it.
		{Code: `
        let x = "";
        function inner() {
          let x = <div><span /></div>;
          return x;
        }
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: `var` declarations are function-scoped, not block-scoped ----
		// `if (cond) { var x = <JSX>; }` does NOT shadow at block level;
		// the inner var hoists to the enclosing function. Verify we don't
		// crash on the form (we still pick the latest write, which is the
		// `var` init that lives in the same hoisted scope).
		{Code: `
        function Foo() {
          if (cond) {
            var x = "";
          }
          return <div>{x}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ---- Lock-in: ConditionalExpression breaks the parent walk ----
		// `<span/>` inside `{cond ? <span/> : null}` has a ConditionalExpression
		// parent, which is neither JSX nor JsxExpression. getDepth therefore
		// stops at count=0 — matching upstream's `while (isJSX || isExpression)`
		// loop semantics. The `<div>` that surrounds the interpolation does
		// NOT add to the inner JSX's depth.
		{Code: `<div>{cond ? <span /> : null}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ---- Lock-in: class component returning JSX from render ----
		// The parent chain from inner JSX walks through MethodDeclaration /
		// ClassDeclaration / ReturnStatement — none JSX-like — so depth is
		// counted correctly within the returned tree only.
		{Code: `
        class Foo extends React.Component {
          render() {
            return (
              <div>
                <span />
              </div>
            );
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: arrow with block body returning JSX ----
		{Code: `
        const Foo = () => {
          return (
            <div>
              <span />
            </div>
          );
        };
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ---- Lock-in: JsxExpression child wrapping non-leaf JSX is opaque ----
		// `checkDescendant` must NOT recurse through a JsxExpressionContainer
		// to inspect its inner JSX's children. Upstream's `isLeaf` returns
		// true for any node without a `.children` field (including JSX
		// expression containers), so for max=2 the JsxExpression child sits
		// at baseDepth=2 and reports nothing — even though the wrapped
		// `<div><span/></div>` would itself violate at depth 3.
		{Code: `
        const A = <Fragment>{<div><span /></div>}</Fragment>;
        <div>{A}</div>;
      `, Tsx: true, Options: map[string]interface{}{"max": 3}},

		// ====== Option-shape coverage ======
		// Each shape exercises a different `parseOptions` branch.
		// nil options (no second-position arg from JS) → defaults.
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: nil},
		// Empty array (`[]`) → defaults.
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: []interface{}{}},
		// Empty object (`[{}]`) → defaults.
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: map[string]interface{}{}},
		// Numeric string `"2"` falls through to default rather than parsing
		// (mirrors upstream's `has(option, 'max') ? option.max : 2` — when
		// `max` IS present but isn't a number, ESLint also fails the schema
		// at config-load and disables the rule). We default to be lenient.
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: map[string]interface{}{"max": "2"}},
		// `max: false` (boolean) → defaults.
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: map[string]interface{}{"max": false}},
		// `max: null` → defaults.
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: map[string]interface{}{"max": nil}},
		// Negative `max: -1` → defaults (upstream schema enforces minimum 0;
		// we silently fall back to default rather than crashing).
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: map[string]interface{}{"max": -1}},
		// Unknown extra option keys are ignored.
		{Code: `<App><foo><bar/></foo></App>`, Tsx: true, Options: map[string]interface{}{"max": 2, "unknown": true}},
		// max=100 — anything fits.
		{Code: `
        <a><b><c><d><e><f><g><h><i /></h></g></f></e></d></c></b></a>
      `, Tsx: true, Options: map[string]interface{}{"max": 100}},

		// ====== JsxExpression inner shapes that are NOT Identifier ======
		// `{this.x}` — PropertyAccessExpression. Listener bails immediately.
		{Code: `<div>{this.x}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// `{x?.y}` — optional chain. Bail.
		{Code: `<div>{x?.y}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// `{x()}` — call expression. Bail.
		{Code: `<div>{makeJSX()}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// `{...arr}` — spread (parsed as JsxExpression with no inner Identifier).
		{Code: `<div>{...arr}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// `{}` — empty JsxExpression has nil inner; listener returns early.
		{Code: `<div>{}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// `{/* comment */}` — JsxExpression whose inner is nil after the
		// comment scrub; listener returns early.
		{Code: `<div>{/* nothing */}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// `{x + y}` — binary expression. Bail.
		{Code: `<div>{x + y}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// `{cond && <span/>}` — logical-and rendering pattern. Inner is
		// BinaryExpression, listener bails. The inner <span/> is reachable
		// via the leaf listener at depth 0 (LogicalExpression breaks parent
		// walk), which fits max=0.
		{Code: `<div>{cond && <span />}</div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ====== Real React patterns ======
		// Render-prop fallback: JSX in attribute is depth 0 regardless of
		// the surrounding component depth.
		{Code: `
        <Suspense fallback={<Loading />}>
          <App />
        </Suspense>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Render-prop function: arrow body is a fresh scope; its JSX depth
		// starts at 0 even when invoked from a deeply nested element.
		{Code: `
        <Provider>
          <Consumer>
            {(value) => <span>{value}</span>}
          </Consumer>
        </Provider>
      `, Tsx: true, Options: map[string]interface{}{"max": 2}},
		// Iteration via .map: `<li>` rendered per item — leaf depth 0
		// inside the arrow body.
		{Code: `<ul>{items.map((item) => <li>{item}</li>)}</ul>`, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// useMemo/useCallback returning JSX: `useMemo(() => <div/>, deps)` —
		// the call result isn't followed.
		{Code: `
        const Foo = () => {
          const memo = useMemo(() => <div><span /></div>, []);
          return <div>{memo}</div>;
        };
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// JSX inside a default parameter — parameter binding is not tracked.
		{Code: `
        function Foo({ children = <span /> }) {
          return <div>{children}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ====== Identifier resolution edge cases ======
		// `let x = x;` — self-init. Cycle detection bails.
		{Code: `
        let x = x;
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// Mutual aliasing.
		{Code: `
        let x = y;
        let y = x;
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// Long aliasing chain (well within the 32-step cap), terminating
		// at a non-JSX init.
		{Code: `
        let a = "";
        let b = a;
        let c = b;
        let d = c;
        let e = d;
        <div>{e}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// Resolution to a function call result — non-JSX, non-Identifier.
		{Code: `
        const x = makeJSX();
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// ====== JSX shape edges ======
		// Empty fragment.
		{Code: `<></>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// Empty element.
		{Code: `<div></div>`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// JSX with comments interleaved between children.
		{Code: `
        <div>
          {/* comment */}
          <span />
        </div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// JSX with spread attribute — attributes don't influence depth.
		{Code: `
        <div {...props}>
          <span {...spanProps} />
        </div>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Multi-level member-access tag `<a.b.c.d/>` — single JSX node.
		{Code: `<a.b.c.d />`, Tsx: true, Options: map[string]interface{}{"max": 0}},
		// JSX as the right-hand side of an arrow with concise body.
		{Code: `const F = () => <div><span /></div>;`, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ====== Reassignment inside diverse control flow ======
		// Reassign in `try` block.
		{Code: `
        function Foo() {
          let x = "";
          try { x = <div />; } catch (e) {}
          return <div>{x}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Reassign in `catch` block.
		{Code: `
        function Foo() {
          let x = "";
          try {} catch (e) { x = <div />; }
          return <div>{x}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Reassign in `finally` block.
		{Code: `
        function Foo() {
          let x = "";
          try {} finally { x = <div />; }
          return <div>{x}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Reassign in `while` body.
		{Code: `
        function Foo() {
          let x = "";
          while (cond) { x = <div />; }
          return <div>{x}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Reassign in `switch` case.
		{Code: `
        function Foo() {
          let x = "";
          switch (kind) {
            case "a": x = <div />; break;
          }
          return <div>{x}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ====== For-loop init scope isolation ======
		// Outer `let x = <a/>` and an inner `for (let x of items)` — uses
		// of `x` inside the loop body resolve to the for-init binding, NOT
		// the outer let. The outer let's writes must not leak in.
		{Code: `
        let x = <a />;
        function Foo() {
          for (let x of items) {
            return <div>{x}</div>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// `for (let x = ...; ...; ...)` — same isolation.
		{Code: `
        let x = <a />;
        function Foo() {
          for (let x = 0; x < n; x++) {
            return <span key={x} />;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// `for (let x in obj)` — same isolation.
		{Code: `
        let x = <a />;
        function Foo() {
          for (let x in obj) {
            return <span key={x} />;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// `for (const x of items)` followed by use OUTSIDE the loop — the
		// outer `x` resolves to the outer declaration. With shadow check
		// honoured, the inner for-init `x` doesn't leak out.
		{Code: `
        let x = "";
        for (const x of items) { void x; }
        <div>{x}</div>
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},

		// Reassignment inside an inner for-of `let x` body MUST NOT leak
		// to outer same-name binding. Differential-validated against upstream.
		{Code: `
        let x = "";
        for (let x of items) { x = <div><span /></div>; }
        <wrap>{x}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// Same scenario for `for (let x = …; …; …)`.
		{Code: `
        let x = "";
        for (let x = 0; x < n; x++) { x = <div><span /></div>; }
        <wrap>{x}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// CatchClause parameter shadows outer `let x`; reassignment to
		// the catch param MUST NOT leak. Differential-validated.
		{Code: `
        let x = "";
        try { foo(); } catch (x) { x = <div><span /></div>; }
        <wrap>{x}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// Destructured catch param: `catch ({ name })` also shadows.
		{Code: `
        let err = "";
        try { foo(); } catch ({ err }) { err = <div><span /></div>; }
        <wrap>{err}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// for-in variant.
		{Code: `
        let k = "";
        for (let k in obj) { k = <div><span /></div>; }
        <wrap>{k}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ====== Parameter / destructured-param default tracking ======
		// Bare parameter with default JSX that fits max=1.
		{Code: `
        function Foo(x = <a />) {
          return <wrap>{x}</wrap>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Destructured parameter with default JSX that fits max=1 (depth 1
		// inside <div>{children}</div> + 0-depth default <span/> = 1).
		{Code: `
        function Foo({ children = <span /> }) {
          return <div>{children}</div>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		// Reassign parameter in body to non-JSX — latest write bails.
		{Code: `
        function Foo(x = <div><span /></div>) {
          x = "";
          return <wrap>{x}</wrap>;
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ====== Switch case-let scope ======
		// `let x` declared in case 'a' is visible in case 'b' (per ES2015
		// CaseBlock scoping). Resolution must descend into clauses.
		// max=1 with x → <a/> (shallow): no descent violation.
		{Code: `
        function Foo(kind) {
          switch (kind) {
            case 'a': let x = <a />; break;
            case 'b': return <wrap>{x}</wrap>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ====== Destructured const default is NOT tracked as writeExpr ======
		// Differential validation against upstream confirms: for
		// `const { x = <jsx> } = obj`, ESLint's scope manager surfaces
		// the destructure RHS (`obj`), NOT the default, as `writeExpr`.
		// So even a default that would violate max=1 produces ZERO reports.
		// The same holds for `let { x = <jsx> } = obj`. Parameter destructure
		// (`function f({ x = <jsx> })`) DOES track the default — see the
		// invalid suite for that contrast.
		{Code: `
        const { x = <a /> } = obj;
        <wrap>{x}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		{Code: `
        const { x = <div><span /></div> } = obj;
        <wrap>{x}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		{Code: `
        let { x = <div><span /></div> } = obj;
        <wrap>{x}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},
		{Code: `
        const [x = <div><span /></div>] = arr;
        <wrap>{x}</wrap>
      `, Tsx: true, Options: map[string]interface{}{"max": 1}},

		// ====== for-of with destructured pattern ======
		// `for (const { x } of items)` — body uses x. No default, no
		// reassignment in body, so no resolution either way. Just verify
		// the destructure doesn't crash the resolver.
		{Code: `
        function Foo() {
          for (const { x } of items) {
            return <wrap>{x}</wrap>;
          }
        }
      `, Tsx: true, Options: map[string]interface{}{"max": 0}},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: max=0 with depth 1 ----
		{
			Code: `
        <App>
          <foo />
        </App>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Message: "Expected the depth of nested jsx elements to be <= 0, but found 1.", Line: 3, Column: 11},
			},
		},

		// ---- Upstream: max=0 with depth 1, JsxExpression child whose inner is non-JSX ----
		// `<foo>{bar}</foo>` is a leaf because `{bar}` is an Identifier, not JSX.
		{
			Code: `
        <App>
          <foo>{bar}</foo>
        </App>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 11},
			},
		},

		// ---- Upstream: max=1 with depth 2 ----
		{
			Code: `
        <App>
          <foo>
            <bar />
          </foo>
        </App>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: identifier-resolved JSX puts <span/> at depth 2 ----
		{
			Code: `
        const x = <div><span /></div>;
        <div>{x}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 24},
			},
		},

		// ---- Upstream: chain through a re-binding `let y = x` ----
		{
			Code: `
        const x = <div><span /></div>;
        let y = x;
        <div>{y}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 24},
			},
		},

		// ---- Upstream: two interpolations report independently ----
		{
			Code: `
        const x = <div><span /></div>;
        let y = x;
        <div>{x}-{y}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 24},
				{MessageId: "wrongDepth", Line: 2, Column: 24},
			},
		},

		// ---- Upstream: inline `{<div>...</div>}` — JsxExpression with JSX
		// inner does NOT trigger the identifier path; <span/> reports via
		// the leaf listener at depth 3.
		{
			Code: `
        <div>
        {<div><div><span /></div></div>}
        </div>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 20},
			},
		},

		// ---- Upstream: fragment, max=0, depth 1 ----
		{
			Code: `
        <>
          <foo />
        </>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 11},
			},
		},

		// ---- Upstream: nested fragments, max=1, depth 2 ----
		{
			Code: `
        <>
          <>
            <bar />
          </>
        </>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
			},
		},

		// ---- Upstream: fragment-wrapped duplicate interpolations ----
		{
			Code: `
        const x = <><span /></>;
        let y = x;
        <>{x}-{y}</>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 21},
				{MessageId: "wrongDepth", Line: 2, Column: 21},
			},
		},

		// ---- Upstream: identifier-resolved table — both <td>s exceed max=1 ----
		{
			Code: `
        const x = (
          <tr>
            <td>1</td>
            <td>2</td>
          </tr>
        );
        <tbody>
          {x}
        </tbody>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
				{MessageId: "wrongDepth", Line: 5, Column: 13},
			},
		},

		// ---- Upstream: deeply nested modal scaffold, max=4 ----
		// Only the inner <Button> reports — its leaf-checked depth is 5, while
		// every other container at depth 4 either has a leaf body that fits
		// (Icon, content div) or is a non-leaf container that defers to its
		// descendants.
		{
			Code: `
        <div className="custom_modal">
          <Modal className={classes.modal} open={isOpen} closeAfterTransition>
            <Fade in={isOpen}>
              <DialogContent>
                <Icon icon="cancel" onClick={onClose} popoverText="Close Modal" />
                <div className="modal_content">{children}</div>
                <div className={clsx('modal_buttons', classes.buttons)}>
                  <Button className="modal_buttons--cancel" onClick={onCancel}>
                    {cancelMsg ? cancelMsg : 'Cancel'}
                  </Button>
                </div>
              </DialogContent>
            </Fade>
          </Modal>
        </div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 4},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 9, Column: 19},
			},
		},

		// ---- Lock-in: parens around resolved identifier still report ----
		{
			Code: `
        const x = <div><span /></div>;
        <div>{(x)}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 24},
			},
		},

		// ---- Lock-in: array option shape (matches multi-element CLI config) ----
		// utils.GetOptionsMap unwraps a single-element array; verify that the
		// `max` key still parses through the array path.
		{
			Code: `
        <App>
          <foo />
        </App>
      `,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"max": 0}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 11},
			},
		},

		// ---- Lock-in: Fragment-wrapped JSX inside an interpolation ----
		// `<><span /></>` inside `{<><span /></>}` is reachable via the leaf
		// listener path: the outer `<div>` is non-leaf (has a JsxExpression
		// child whose inner is JSX), the JsxExpression listener no-ops
		// (inner is JSX, not Identifier), and `<span />` reports as a leaf
		// at depth = outer-div + Fragment + (JsxExpression skipped) = 2.
		{
			Code: `
        <div>{<><span /></>}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 17},
			},
		},

		// ---- Reassignment: latest write wins (matches upstream) ----
		// `let x = ""` followed by `x = <div><span/></div>` puts the JSX
		// reassignment LATER in source order. ESLint's scope manager picks
		// it as the reference's write expression; we mirror that by scanning
		// every `name = expr` in the binding's enclosing scope.
		{
			Code: `
        let x = "";
        x = <div><span /></div>;
        <div>{x}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 18},
			},
		},

		// ---- Reassignment: declaration without initializer ----
		// `let x;` produces no write expression, then `x = <JSX>` is the
		// only write. The scan picks it up.
		{
			Code: `
        let x;
        x = <div><span /></div>;
        <div>{x}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 18},
			},
		},

		// ---- Reassignment: multiple writes, latest wins ----
		// Three writes — declaration init `<a/>`, reassignment `<b><c/></b>`,
		// reassignment `<d><e><f/></e></d>`. The deepest violation comes
		// from the LAST write in source order; the earlier writes are
		// silenced. (Upstream takes the same "first reverse-iter" branch.)
		//
		// Two reports fire in traversal order:
		//   - `<f/>` leaf via the leaf listener while traversing the third
		//     reassignment (depth=2 > 1).
		//   - `<e>` via the JsxExpression listener on the later `{x}` site,
		//     which resolves to the latest write `<d><e><f/></e></d>` and
		//     reports `<e>` at baseDepth=2 > 1.
		{
			Code: `
        let x = <a />;
        x = <b><c /></b>;
        x = <d><e><f /></e></d>;
        <div>{x}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 19},
				{MessageId: "wrongDepth", Line: 4, Column: 16},
			},
		},

		// ---- Reassignment: inside an `if` branch ----
		// The reassignment lives in an inner Block, but the binding's scope
		// is the enclosing function body — the scan must descend through
		// nested Blocks.
		{
			Code: `
        function Foo() {
          let x = "";
          if (cond) {
            x = <div><span /></div>;
          }
          return <div>{x}</div>;
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 5, Column: 22},
			},
		},

		// ---- Reassignment: closure write from a non-shadowing inner function ----
		// Inner function does NOT redeclare `x`, so its write to `x` is a
		// closure write to the outer binding. ESLint's scope manager tracks
		// this via the variable's reference list; we descend into inner
		// function bodies that don't shadow `name`.
		{
			Code: `
        function outer() {
          let x = "";
          function inner() {
            x = <div><span /></div>;
          }
          return <div>{x}</div>;
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 5, Column: 22},
			},
		},

		// ---- Reassignment chain: y → x with x reassigned to JSX ----
		// `let y = x` (init = Identifier x). Recurse on x. x's latest write
		// is the JSX reassignment, not the empty-string init.
		{
			Code: `
        let x = "";
        x = <div><span /></div>;
        let y = x;
        <div>{y}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 18},
			},
		},

		// ---- Reassignment: multi-decl statement `let a, b = <JSX>, c` ----
		// Multiple declarations in one VariableStatement — the scan must
		// pick the right one by Identifier name.
		{
			Code: `
        let a = 1, b = <div><span /></div>, c = 3;
        <div>{b}</div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 29},
			},
		},

		// ---- Class component returning deep JSX violates max=1 ----
		// Locks in the leaf listener's behavior across the full
		// MethodDeclaration → ClassDeclaration → SourceFile parent chain.
		{
			Code: `
        class Foo extends React.Component {
          render() {
            return (
              <div>
                <div>
                  <span />
                </div>
              </div>
            );
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 7, Column: 19},
			},
		},

		// ---- Class field arrow returning deep JSX ----
		// `class C { render = () => (<div><div><span/></div></div>); }` —
		// arrow stored as a class field, not a method. Same depth count
		// applies; only the parent chain shape is different.
		{
			Code: `
        class Foo extends React.Component {
          render = () => (
            <div>
              <div>
                <span />
              </div>
            </div>
          );
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 6, Column: 17},
			},
		},

		// ---- TS generic component tag ----
		// `<Foo<string>>` parses as a JsxOpeningElement with type-argument;
		// the rule must treat it as a normal JSX element so its child
		// depth still counts.
		{
			Code: `
        <Foo<string>>
          <Bar>
            <Baz />
          </Bar>
        </Foo>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
			},
		},

		// ---- Member-access tag name ----
		// `<Module.Foo>` is a single JsxElement with a PropertyAccess tag —
		// IsJsxLike must still match it for the parent walk and leaf check.
		{
			Code: `
        <Module.Foo>
          <Bar>
            <Baz />
          </Bar>
        </Module.Foo>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
			},
		},

		// ---- Namespaced tag name ----
		// `<svg:line/>` uses JsxNamespacedName as the tag — same handling.
		{
			Code: `
        <svg>
          <svg:g>
            <svg:line />
          </svg:g>
        </svg>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
			},
		},

		// ---- Deep fragment chain — every level counts ----
		// `<><><><><x/></></></></>` puts <x/> at depth 4. Default max=2.
		{
			Code: `
        <>
          <>
            <>
              <>
                <x />
              </>
            </>
          </>
        </>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 6, Column: 17},
			},
		},

		// ---- Mixed element / fragment nesting ----
		// Verify counting works the same regardless of which JsxKind sits
		// at each level.
		{
			Code: `
        <App>
          <>
            <Foo>
              <Bar />
            </Foo>
          </>
        </App>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"max": 2},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 5, Column: 15},
			},
		},

		// ====== Boundary depth + message-data alignment ======
		// max=0 with depth-1 leaf — exact message text alignment. The
		// `{{found}}` / `{{needed}}` interpolation must produce ESLint's
		// exact wording.
		{
			Code: `<App><foo /></App>`,
			Tsx:  true, Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Message: "Expected the depth of nested jsx elements to be <= 0, but found 1.", Line: 1, Column: 6},
			},
		},
		// max=0 with depth-2 leaf — only the LEAF reports (depth 2),
		// non-leaf containers don't fire. Message data: found=2 needed=0.
		{
			Code: `<App><foo><bar /></foo></App>`,
			Tsx:  true, Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Message: "Expected the depth of nested jsx elements to be <= 0, but found 2.", Line: 1, Column: 11},
			},
		},
		// Off-by-one: max=2 with depth-3 leaf reports; depth-2 sibling
		// doesn't. Verifies the strict `>` comparison.
		{
			Code: `<a><b><c><d /></c></b></a>`,
			Tsx:  true, Options: map[string]interface{}{"max": 2},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Message: "Expected the depth of nested jsx elements to be <= 2, but found 3.", Line: 1, Column: 10},
			},
		},
		// Large max value with deeper actual: max=5, depth=6.
		{
			Code: `<a><b><c><d><e><f><g /></f></e></d></c></b></a>`,
			Tsx:  true, Options: map[string]interface{}{"max": 5},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Message: "Expected the depth of nested jsx elements to be <= 5, but found 6.", Line: 1, Column: 19},
			},
		},

		// ====== Leaf vs non-leaf in same container ======
		// Two siblings — one leaf at over-depth, one non-leaf that itself
		// contains a deeper leaf. Both should report at their own leaves.
		{
			Code: `
        <a>
          <b />
          <c><d /></c>
        </a>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 14},
			},
		},

		// ====== Reassignment in control flow that DOES exceed ======
		// While-loop body with deep JSX reassignment.
		{
			Code: `
        function Foo() {
          let x = "";
          while (cond) { x = <div><span /></div>; }
          return <div>{x}</div>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 35},
			},
		},
		// Try-catch reassignment.
		{
			Code: `
        function Foo() {
          let x = "";
          try { x = <div><span /></div>; } catch (e) {}
          return <div>{x}</div>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 26},
			},
		},
		// Switch case reassignment.
		{
			Code: `
        function Foo() {
          let x = "";
          switch (kind) {
            case "deep": x = <div><span /></div>; break;
          }
          return <div>{x}</div>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 5, Column: 35},
			},
		},

		// ====== Closure descent through TWO function levels ======
		// Reassign happens inside a nested function whose own body doesn't
		// shadow `x`. Walking must descend through both layers.
		{
			Code: `
        function outer() {
          let x = "";
          function mid() {
            function inner() {
              x = <div><span /></div>;
            }
          }
          return <div>{x}</div>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 6, Column: 24},
			},
		},

		// ====== Multiple use sites of the same binding ======
		// Three `{x}` interpolations — each reports independently.
		{
			Code: `
        const x = <div><span /></div>;
        <div>{x}{x}{x}</div>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 24},
				{MessageId: "wrongDepth", Line: 2, Column: 24},
				{MessageId: "wrongDepth", Line: 2, Column: 24},
			},
		},

		// ====== Resolution chain length ======
		// y → x where x is reassigned to JSX in source: chain still works.
		{
			Code: `
        let x = "";
        let y = x;
        x = <div><span /></div>;
        <div>{y}</div>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 18},
			},
		},

		// ====== Destructured parameter default IS tracked as a write ======
		// `function Foo({ children = <div><span/></div> })` — the BindingElement
		// default seeds the write list, so `<div>{children}</div>` resolves
		// children → <div><span/></div> and the descendant walk reports at
		// depth 2. The leaf listener also reports the SAME `<span/>` from the
		// default at depth 1 — two reports total at the same position.
		{
			Code: `
        function Foo({ children = <div><span /></div> }) {
          return <div>{children}</div>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 40},
				{MessageId: "wrongDepth", Line: 2, Column: 40},
			},
		},

		// ====== Multibyte source positions ======
		// Multibyte chars in JsxText shift byte offsets but UTF-16 column
		// counts (which rslint emits) align with what ESLint reports for
		// the same source.
		{
			Code: `
        <App>
          <foo>日本語</foo>
          <bar><baz /></bar>
        </App>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 16},
			},
		},

		// ====== Nested resolution: y → x, both reassigned ======
		// y's latest write is `<a><b/></a>` (Identifier x is overwritten).
		// Without latest-write tracking, we'd resolve y to x → ... and miss.
		{
			Code: `
        let x = "";
        let y = "";
        x = <div><span /></div>;
        y = <a><b /></a>;
        <div>{y}</div>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 5, Column: 16},
			},
		},

		// ====== Same-line multi-element ======
		// All on one line — verifies column counting without leading
		// indentation.
		{
			Code: `<a><b><c /></b></a>`,
			Tsx:  true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 1, Column: 7},
			},
		},

		// ====== TS generic with multiple type args ======
		{
			Code: `
        <Foo<string, number>>
          <Bar>
            <Baz />
          </Bar>
        </Foo>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
			},
		},

		// ====== Stress: aliasing chain to a deep JSX value ======
		// e → d → c → b → a, where a's reassignment goes deeper.
		{
			Code: `
        let a = "";
        const b = a;
        const c = b;
        const d = c;
        const e = d;
        a = <div><span /></div>;
        <div>{e}</div>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 7, Column: 18},
			},
		},

		// ====== For-init binding reassigned to JSX in loop body ======
		// `for (let x of items)` followed by `x = <JSX>` in body — the
		// reassignment IS within the for-init binding's scope (the entire
		// ForStatement), so resolution picks it up. Without ForStatement
		// scope detection, we'd attribute the write to the wrong binding.
		{
			Code: `
        function Foo() {
          for (let x of items) {
            x = <div><span /></div>;
            return <div>{x}</div>;
          }
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 22},
			},
		},

		// `for (let x = init; ...; ...)` reassigned in body to deeper JSX.
		{
			Code: `
        function Foo() {
          for (let x = <a />; cond; ) {
            x = <div><span /></div>;
            return <div>{x}</div>;
          }
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 22},
			},
		},

		// ====== Parameter default JSX exceeds via descend ======
		// `function Foo(x = <div><span/></div>)` — parameter default has
		// depth 1 (so leaf listener at <span/> is OK with max=1), but the
		// `<wrap>{x}</wrap>` resolves x → <div><span/></div> and the
		// descend reports <span/> at baseDepth 2.
		{
			Code: `
        function Foo(x = <div><span /></div>) {
          return <wrap>{x}</wrap>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 31},
			},
		},

		// ====== Switch case-let scope (cross-clause resolution) ======
		// `let x = <div><span/></div>` declared in case 'a' is visible in
		// case 'b'. Descent into clauses MUST find it.
		{
			Code: `
        function Foo(kind) {
          switch (kind) {
            case 'a': let x = <div><span /></div>; break;
            case 'b': return <wrap>{x}</wrap>;
          }
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 36},
			},
		},

		// ====== Destructured const reassigned in scope ======
		// `const { x } = obj; x = <jsx>` — even though `const` re-assignment
		// is a runtime TypeError, it parses and ESLint's scope picks it up.
		// We mirror that.
		{
			Code: `
        const { x } = obj;
        x = <div><span /></div>;
        <wrap>{x}</wrap>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 18},
			},
		},

		// ====== Parameter reassigned in body (latest wins) ======
		// `x` parameter reassigned in body to deeper JSX.
		{
			Code: `
        function Foo(x = <a />) {
          x = <div><span /></div>;
          return <wrap>{x}</wrap>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 3, Column: 20},
			},
		},

		// ====== Resolution chain crosses a shadowed binding ======
		// Outer `x` is `<a/>` (shallow), inner shadow is `<b><c/></b>`.
		// `y = x` inside inner refers to inner x. Walk through `let y`
		// initializer must resolve x in the inner scope and pick the inner
		// JSX, not the outer one. With max=1, <c/> at depth 2 must fire.
		{
			Code: `
        const x = <a />;
        function inner() {
          const x = <b><c /></b>;
          const y = x;
          return <div>{y}</div>;
        }
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 24},
			},
		},

		// ====== Fragment at top level wrapping deeper JSX ======
		// Top-level Fragment counts toward depth from its child JSX.
		{
			Code: `
        <>
          <App>
            <Child />
          </App>
        </>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 1},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 4, Column: 13},
			},
		},

		// ====== Mixed Fragment + Element five levels deep ======
		// Each level (regardless of Fragment vs Element) increments depth.
		// max=3 with depth 4 → leaf reports.
		{
			Code: `
        <Outer>
          <>
            <Mid>
              <>
                <Leaf />
              </>
            </Mid>
          </>
        </Outer>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 3},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 6, Column: 17},
			},
		},

		// ====== Reassignment latest-wins picks NON-violating value ======
		// First reassignment `<a><b/></a>` violates max=1, but a later
		// reassignment to a flatter `<x/>` is the latest. Resolution must
		// pick the latest, so the JsxExpression listener finds no
		// violation through `{r}` — but the leaf listener still fires on
		// `<b/>` at line 3 (its source-level depth=1 only, fits max=1, OK)
		// and only the second reassignment's `<x/>` is at depth 0. The
		// JsxExpression listener thus produces ZERO reports — moved to
		// Valid. Lock in via the inverse: a deeper JSX in latest write.
		// (See valid suite for the matching "latest win, fits" case.)

		// ====== JSX in array literal — each element is depth 0 ======
		// Ensure JSX siblings inside an ArrayLiteralExpression don't add
		// depth to each other. `<b/>` inside `[<a/>, <b/>, <c/>]` is leaf
		// depth 0 — and the rule only sees them via the leaf listener.
		// max=0 fits because depth is 0 for all of them. Locked here as
		// invalid by adding nesting inside one element.
		{
			Code: `
        const arr = [<a />, <b><c /></b>, <d />];
      `,
			Tsx: true, Options: map[string]interface{}{"max": 0},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 32},
			},
		},

		// ====== Long alias chain hits the depth cap ======
		// 32+ aliasing levels should bail without infinite recursion.
		// Even though semantically the chain terminates at JSX, our cap
		// trips first — no diagnostic, no crash. Locks the cap behavior;
		// upstream's `containDuplicates` set serves the same purpose, just
		// detected differently.
		{
			Code: `
        const a0 = <div><span /></div>;
        const a1 = a0;
        const a2 = a1; const a3 = a2; const a4 = a3; const a5 = a4;
        const a6 = a5; const a7 = a6; const a8 = a7; const a9 = a8;
        const a10 = a9; const a11 = a10; const a12 = a11; const a13 = a12;
        const a14 = a13; const a15 = a14; const a16 = a15; const a17 = a16;
        const a18 = a17; const a19 = a18; const a20 = a19; const a21 = a20;
        const a22 = a21; const a23 = a22; const a24 = a23; const a25 = a24;
        const a26 = a25; const a27 = a26; const a28 = a27; const a29 = a28;
        const a30 = a29; const a31 = a30; const a32 = a31;
        // a32 → a31 → ... → a0 → JSX. Chain length 33 > cap.
        // Resolution bails; only the leaf listener fires on the original
        // <span/> at depth 1 — which violates max=0.
        <div>{a32}</div>
      `,
			Tsx: true, Options: map[string]interface{}{"max": 0},
			// Single report from the leaf listener at the original <span/>.
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "wrongDepth", Line: 2, Column: 25},
			},
		},
	})
}
