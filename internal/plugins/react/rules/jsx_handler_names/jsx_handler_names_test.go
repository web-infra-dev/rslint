package jsx_handler_names

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxHandlerNamesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxHandlerNamesRule, []rule_tester.ValidTestCase{
		// ---- Upstream: standard `onX` / `handleX` patterns ----
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true},
		// upstream comment: "TODO: make this an invalid test"; we keep it valid for parity
		{Code: `var x = <TestComponent onChange={this.handle123Change} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this.props.onChange} />`, Tsx: true},

		// ---- Upstream: multiline member access (whitespace stripped before regex) ----
		{Code: `
        var x = <TestComponent
          onChange={
            this
              .handleChange
          } />
      `, Tsx: true},
		{Code: `
        var x = <TestComponent
          onChange={
            this
              .props
              .handleChange
          } />
      `, Tsx: true},

		// ---- Upstream: checkLocalVariables ----
		{Code: `var x = <TestComponent onChange={handleChange} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true}},
		{Code: `var x = <TestComponent onChange={takeCareOfChange} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": false}},

		// ---- Upstream: checkInlineFunction ----
		{Code: `var x = <TestComponent onChange={event => window.alert(event.target.value)} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": false}},
		{Code: `var x = <TestComponent onChange={() => handleChange()} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true, "checkLocalVariables": true}},
		{Code: `var x = <TestComponent onChange={() => this.handleChange()} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true}},
		// Default options skip arrow inline handlers entirely.
		{Code: `var x = <TestComponent onChange={() => 42} />`, Tsx: true},

		// ---- Upstream: well-named props with no special handler shape ----
		{Code: `var x = <TestComponent onChange={this.props.onFoo} />`, Tsx: true},
		{Code: `var x = <TestComponent isSelected={this.props.isSelected} />`, Tsx: true},
		{Code: `var x = <TestComponent shouldDisplay={this.state.shouldDisplay} />`, Tsx: true},
		{Code: `var x = <TestComponent shouldDisplay={arr[0].prop} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={props.onChange} />`, Tsx: true},

		// ---- Upstream: ref is always allowed ----
		{Code: `var x = <TestComponent ref={this.handleRef} />`, Tsx: true},
		{Code: `var x = <TestComponent ref={this.somethingRef} />`, Tsx: true},

		// ---- Upstream: custom prefixes (matching prop key, not handler) ----
		{Code: `var x = <TestComponent test={this.props.content} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "on", "eventHandlerPropPrefix": "on"}},

		// ---- Upstream: bind operator forms — TypeScript does not parse `::`,
		// so these are skipped. SKIP: rslint's tsgo parser does not implement
		// the stage-1 `::` bind operator proposal.
		{Code: `var x = <TestComponent onChange={props::handleChange} />`, Tsx: true, Skip: true},
		{Code: `var x = <TestComponent onChange={::props.onChange} />`, Tsx: true, Skip: true},
		{Code: `var x = <TestComponent onChange={props.foo::handleChange} />`, Tsx: true, Skip: true},
		{Code: `var x = <TestComponent onChange={() => props::handleChange()} />`, Tsx: true, Skip: true, Options: map[string]interface{}{"checkInlineFunction": true}},
		{Code: `var x = <TestComponent onChange={() => ::props.onChange()} />`, Tsx: true, Skip: true, Options: map[string]interface{}{"checkInlineFunction": true}},
		{Code: `var x = <TestComponent onChange={() => props.foo::handleChange()} />`, Tsx: true, Skip: true, Options: map[string]interface{}{"checkInlineFunction": true}},

		// ---- Upstream: prop with no `on` prefix (no handler shape on either side) ----
		{Code: `var x = <TestComponent only={this.only} />`, Tsx: true},

		// ---- Upstream: disabled prefixes ----
		// `eventHandlerPrefix: false` shuts off the handler-name half — only
		// prop-key violations can fire. The handler text is never checked.
		{Code: `var x = <TestComponent onChange={this.someChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": false, "eventHandlerPropPrefix": "on"}},
		{Code: `var x = <TestComponent somePrefixChange={this.someChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": false, "eventHandlerPropPrefix": "somePrefix"}},
		// `eventHandlerPropPrefix: false` shuts off the prop-key half — only
		// handler-name violations can fire (and only for matching prop keys,
		// which never match because the prop regex is null).
		{Code: `var x = <TestComponent someProp={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": false}},
		{Code: `var x = <TestComponent someProp={this.somePrefixChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "somePrefix", "eventHandlerPropPrefix": false}},
		{Code: `var x = <TestComponent someProp={props.onChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": false}},

		// ---- Upstream: ignoreComponentNames ----
		{Code: `var x = <ComponentFromOtherLibraryBar customPropNameBar={handleSomething} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"ComponentFromOtherLibraryBar"}}},
		{Code: `
        function App() {
          return (
            <div>
              <MyLibInput customPropNameBar={handleSomething} />
              <MyLibCheckbox customPropNameBar={handleSomething} />
              <MyLibButtom customPropNameBar={handleSomething} />
            </div>
          )
        }
      `, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"MyLib*"}}},
		{Code: `var x = <A.TestComponent customPropNameBar={handleSomething} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"A.TestComponent"}}},
		{Code: `
        function App() {
          return (
            <div>
              <A.MyLibInput customPropNameBar={handleSomething} />
              <A.MyLibCheckbox customPropNameBar={handleSomething} />
              <A.MyLibButtom customPropNameBar={handleSomething} />
            </div>
          )
        }
      `, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"A.MyLib*"}}},

		// ---- tsgo edge: parenthesized member access (Dimension 4) ----
		// In ESTree, `(this.handleChange)` flattens to a MemberExpression. In
		// tsgo, it's a ParenthesizedExpression wrapping a PropertyAccessExpression;
		// the rule unwraps before the gating check and before reading the text.
		{Code: `var x = <TestComponent onChange={(this.handleChange)} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={((this.handleChange))} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={(this.props.onChange)} />`, Tsx: true},

		// ---- tsgo edge: optional chain on receiver (Dimension 4) ----
		// `this?.handleChange` is a PropertyAccessExpression with the optional
		// flag in tsgo; in ESTree it's a MemberExpression with `optional: true`.
		// Both have a `.object` / Expression, so the gating treats them the
		// same as `this.handleChange`.
		{Code: `var x = <TestComponent onChange={this?.handleChange} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this?.props?.onChange} />`, Tsx: true},

		// ---- Universal edge: deep / chained member access (Dimension 4) ----
		// `(.*\.)?handle` in upstream's regex handles arbitrarily deep chains;
		// lock that in so a future helper extraction doesn't accidentally
		// truncate the dotted prefix.
		{Code: `var x = <TestComponent onChange={a.b.c.handleChange} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this.props.deep.nested.handleChange} />`, Tsx: true},

		// ---- Universal edge: array-index member access on the receiver ----
		// `arr[0].handleChange` — ElementAccessExpression at the base, dotted
		// access on top. Both ESLint and rslint treat the outer dot as the
		// MemberExpression and accept it.
		{Code: `var x = <TestComponent onChange={arr[0].handleChange} />`, Tsx: true},

		// ---- Universal edge: container forms (Dimension 4) ----
		// async / generator / class-field arrow / method shapes: the rule
		// listens on JsxAttribute regardless of where the JSX sits, so verify
		// it does the right thing inside each container.
		{Code: `class C { onClick = <Foo onChange={this.handleChange} /> }`, Tsx: true},
		{Code: `function App() { return <Foo onChange={this.handleChange} />; }`, Tsx: true},
		{Code: `const App = () => <Foo onChange={this.handleChange} />`, Tsx: true},
		{Code: `const App = async () => <Foo onChange={this.handleChange} />`, Tsx: true},

		// ---- Universal edge: same-kind nesting (Dimension 4) ----
		// JSX-in-JSX: each JsxAttribute should be checked independently, with
		// no leakage across boundaries. The outer + inner attributes both pass
		// here, so the lack of diagnostics confirms the listener visits both.
		{Code: `var x = <Outer onChange={this.handleChange}><Inner onClick={this.handleClick} /></Outer>`, Tsx: true},

		// ---- Universal edge: async inline arrow body (Dimension 1) ----
		// `async () => this.handleChange()` — async modifier doesn't change
		// the body shape; CallExpression with member callee still matches.
		{Code: `var x = <TestComponent onChange={async () => this.handleChange()} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true}},

		// ---- Universal edge: graceful degradation (Dimension 4) ----
		// Spread JSX attribute is JsxSpreadAttribute, not JsxAttribute, so
		// the listener never fires on it. Verify the rule doesn't crash on
		// elements that ONLY have spreads.
		{Code: `var x = <TestComponent {...props} />`, Tsx: true},

		// ---- Lock-in: empty-string prefix falls back to default ----
		// Upstream's `configuration.eventHandlerPrefix || 'handle'` falls back
		// to the default for any falsy value, including `""`. The default
		// `handle` regex must therefore still apply with explicit `""`.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": ""}},
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": ""}},

		// ---- Lock-in: invalid-regex prefix doesn't panic ----
		// User-provided prefixes are concatenated into the regex source. If
		// the prefix contains unbalanced metacharacters the resulting regex
		// fails to compile — `regexp.MustCompile` would panic and crash the
		// linter. We use `regexp.Compile` and silently disable the half of
		// the rule whose regex couldn't compile, mirroring upstream's
		// "rule loading error" outcome (the rule effectively becomes a
		// no-op for that side instead of taking down the process).
		//
		// Each row exercises a different category of regex-compile failure;
		// none should panic, and with the failing side disabled the
		// otherwise-valid `this.handleChange` should produce no diagnostic.
		// `(`            — unbalanced open paren.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "("}},
		// `[`            — unbalanced open class.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "["}},
		// `)`            — stray close paren.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": ")"}},
		// `*`            — quantifier with no atom (becomes `?*` in the
		//                  emitted regex, which RE2 rejects as nested).
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "*"}},
		// `+`            — same as `*`.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "+"}},
		// `\`            — trailing escape with no following char.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "\\"}},
		// `[a-Z]`        — inverted character-class range (RE2 rejects).
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "[a-Z]"}},
		// `(?`           — unbalanced + Perl-syntax fragment.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "(?"}},
		// `[[:foo:]]`    — bogus POSIX class.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "[[:foo:]]"}},

		// Symmetric coverage on the prop side. Note: when the prop regex
		// fails to compile, `propIsEventHandler` is permanently false, so
		// `<X onChange={this.handleChange}>` produces no diagnostic (the
		// matching half of the rule is disabled).
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": "("}},
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": ")"}},
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": "*"}},
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": "+"}},
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": "?"}},
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPropPrefix": "[a-Z]"}},

		// Both prefixes invalid — neither half can fire. No diagnostic
		// regardless of input.
		{Code: `var x = <TestComponent onChange={this.bad} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "(", "eventHandlerPropPrefix": "("}},

		// ---- Lock-in: invalid glob in ignoreComponentNames doesn't panic ----
		// `MatchGlob` returns `false` for patterns that fail to compile, so
		// invalid globs silently fail to ignore — matching upstream
		// minimatch's "no match on bad pattern" behavior. The rule then
		// runs as usual and fires on the bad handler.
		// `Foo[`         — unbalanced class.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"TestComponent[", "TestComponent"}}},
		// Empty class `[]`.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"TestComponent[]", "TestComponent"}}},
		// Reverse range `[z-a]`.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"[z-a]", "TestComponent"}}},

		// ---- Glob alignment: `!` (empty negation) matches everything ----
		// `!pattern` strips one `!` and matches text that does NOT match
		// the rest. With `pattern == ""` and a non-empty text, the empty
		// regex matches only the empty string → false → negate → true.
		{Code: `var x = <TestComponent onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"!"}}},

		// ---- Glob alignment: `**` collapses to `*` (noglobstar) ----
		// Component-name patterns never carry path separators, so `**` and
		// `*` are equivalent in practice — both match anything.
		{Code: `var x = <TestComponent onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"**"}}},

		// ---- Lock-in: regex meta in prefix is interpreted as regex (NOT escaped) ----
		// Upstream concatenates the user prefix into the regex source without
		// escaping, so `.+` means "any one-or-more chars", `[a-z]` is a class.
		// Escaping with `regexp.QuoteMeta` would diverge from upstream on
		// these inputs — we deliberately don't escape.
		// With `eventHandlerPrefix: '.+'`, the rule's "is this a valid handler
		// name" regex is `((.*\.)?.+)[0-9]*[A-Z].*` — propValue "aXY" matches
		// (`.+` consumes "a", then `[A-Z]` consumes "X"), so no violation.
		{Code: `var x = <TestComponent onChange={this.aXY} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": ".+"}},
		// With `eventHandlerPrefix: '[a-z]'`, propValue "aXY" matches similarly
		// (`[a-z]` consumes "a", `[A-Z]` consumes "X").
		{Code: `var x = <TestComponent onChange={this.aXY} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": "[a-z]"}},

		// ---- tsgo edge: TS-only wrappers don't unwrap (Dimension 4) ----
		// AsExpression / NonNullExpression / SatisfiesExpression don't carry
		// a `.object` field, so under default `checkLocalVariables: false`
		// they fail the member-access gate and the rule produces no diagnostic.
		{Code: `var x = <TestComponent onChange={this.someChange as any} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this.someChange!} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this.someChange satisfies any} />`, Tsx: true},

		// ---- tsgo edge: empty / boolean / string-literal attribute initializers ----
		{Code: `var x = <TestComponent onChange />`, Tsx: true},
		{Code: `var x = <TestComponent onChange="literal" />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={} />`, Tsx: true},

		// ---- tsgo edge: JsxNamespacedName attribute (Dimension 4) ----
		// `<X ns:on={...} />` — namespaced JSX attribute. Upstream's `propKey`
		// is then a node object that matches no regex; the effective behavior
		// is "no diagnostic". Mirrored by skipping non-Identifier names.
		{Code: `var x = <TestComponent ns:onChange={this.doSomethingOnChange} />`, Tsx: true},

		// ---- Lock-in: option-shape coverage (CLI/JSON path through GetOptionsMap) ----
		// Bare-object form (single-element CLI shape) and array-wrapped form
		// (rule-tester / multi-element shape) must both reach the parser.
		{Code: `var x = <TestComponent onChange={takeCareOfChange} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": false}},
		{Code: `var x = <TestComponent onChange={takeCareOfChange} />`, Tsx: true, Options: []interface{}{map[string]interface{}{"checkLocalVariables": false}}},
		// Defaults: empty object → identical behavior to no options.
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{}},
		{Code: `var x = <TestComponent onChange={this.handleChange} />`, Tsx: true, Options: []interface{}{}},

		// ---- Real-world: full ES6 class component with handler ----
		// The classic React class-component pattern. Verifies the listener
		// fires on JsxAttribute regardless of which container hosts the JSX
		// (here: a method `render`).
		{Code: `
        class Hello extends React.Component {
          handleClick() { this.setState({ clicked: true }); }
          render() {
            return <button type="button" onClick={this.handleClick}>Click</button>;
          }
        }
      `, Tsx: true},

		// ---- Real-world: class field arrow handler ----
		// `handleClick = () => {...}` — the JSX consumer reads it as
		// `this.handleClick`. Common autobinding alternative.
		{Code: `
        class Hello extends React.Component {
          handleClick = () => this.setState({ clicked: true });
          render() {
            return <button onClick={this.handleClick} />;
          }
        }
      `, Tsx: true},

		// ---- Real-world: functional component with useCallback ----
		// `useCallback(() => {}, [])` returns the local handler; under default
		// options the rule skips local-variable refs (gate: !member-access).
		{Code: `
        function App() {
          const handleClick = React.useCallback(() => {}, []);
          return <button onClick={handleClick} />;
        }
      `, Tsx: true},

		// ---- Real-world: prop forwarding via `props.onX` ----
		// The canonical form for forwarding a `onX` prop. propValue
		// `props.onChange` matches `props\.on[0-9]*[A-Z]` directly.
		{Code: `
        function Wrapper({ onChange, handleChange }) {
          return <Inner onChange={handleChange} />;
        }
      `, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true}},

		// ---- Real-world: render prop / function-as-children pattern ----
		// Each inner JSX element is checked independently; outer container's
		// JSX expressions don't bleed into rule scope.
		{Code: `
        <Provider>
          {(value) => <Consumer onChange={this.handleChange} />}
        </Provider>
      `, Tsx: true},

		// ---- Real-world: rendering a list with map ----
		// Inline arrow at the JSX child position, NOT as a JSX attribute —
		// rule does not fire on JSX child expressions.
		{Code: `
        function App() {
          return <ul>{items.map(item => <Item key={item.id} onClick={this.handleClick} />)}</ul>;
        }
      `, Tsx: true},

		// ---- Real-world: JSX fragment with handlers ----
		// `<>...</>` shorthand fragments — the rule visits any JsxAttribute
		// regardless of fragment wrapping.
		{Code: `<><button onClick={this.handleClick} /><input onChange={this.handleChange} /></>`, Tsx: true},
		{Code: `<React.Fragment><button onClick={this.handleClick} /></React.Fragment>`, Tsx: true},

		// ---- Real-world: spread + named-attribute mix (matching pair valid) ----
		// Spread doesn't fire JsxAttribute; the named attribute is the only
		// thing the rule considers.
		{Code: `var x = <TestComponent {...props} onChange={this.handleChange} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this.handleChange} {...props} />`, Tsx: true},

		// ---- Real-world: handler-method invocation chain ----
		// `this.handleChange.bind(this)` — CallExpression at the top.
		// Default options: !checkLocal && !memberAccess → skip.
		{Code: `var x = <TestComponent onChange={this.handleChange.bind(this)} />`, Tsx: true},
		// With checkLocal: true the gate doesn't apply, but the regex still
		// matches because the trimmed text contains `(.*\.)?handle...[A-Z]`.
		{Code: `var x = <TestComponent onChange={this.handleChange.bind(this)} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true}},

		// ---- Real-world: getter-style handler reference ----
		// `getHandler()` is a CallExpression — under default options skipped.
		{Code: `var x = <TestComponent onChange={getHandler()} />`, Tsx: true},

		// ---- Real-world: Unicode whitespace inside expression ----
		// Non-breaking space (U+00A0) between identifier tokens — JS's
		// `\s` regex strips this; we must too via unicode.IsSpace.
		{Code: "var x = <TestComponent onChange={this . handleChange} />", Tsx: true},
		// Line separator (U+2028) inside the expression — same coverage.
		{Code: "var x = <TestComponent onChange={this .handleChange} />", Tsx: true},

		// ---- Real-world: TypeScript prop with explicit generic ----
		// Type assertions on JSX tag names don't reach JsxAttribute parsing.
		{Code: `var x = <TestComponent<string> onChange={this.handleChange} />`, Tsx: true},

		// ---- Real-world: conditional / logical / nullish handler value ----
		// ConditionalExpression / BinaryExpression don't have `.object`; under
		// default options they fail the member-access gate and skip.
		{Code: `var x = <TestComponent onChange={cond ? this.handleA : this.handleB} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this.cond && this.handleChange} />`, Tsx: true},
		{Code: `var x = <TestComponent onChange={this.handler ?? defaultHandler} />`, Tsx: true},

		// ---- Real-world: ignore wildcard '*' matches every component ----
		// A common shorthand to bypass the rule for an entire JSX file.
		{Code: `var x = <AnyComponent randomProp={someFunc} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"*"}}},

		// ---- Real-world: multiple ignore patterns, last one matches ----
		// Iteration must visit every pattern, not short-circuit on first miss.
		{Code: `var x = <Special prop={handleX} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Foo", "Bar*", "Special"}}},

		// ---- Glob alignment: brace expansion `{a,b,c}` (minimatch parity) ----
		// Upstream uses minimatch which expands `{1,2}` into "1 or 2". rslint's
		// `MatchGlob` was upgraded to support this. Each branch must be honored
		// independently — a future regression would either over-match or
		// fail to match.
		{Code: `var x = <Foo1 onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Foo{1,2}"}}},
		{Code: `var x = <Foo2 onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Foo{1,2}"}}},

		// ---- Glob alignment: nested brace `{a,{b,c}}` ----
		{Code: `var x = <X.a onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"X.{a,{b,c}}"}}},
		{Code: `var x = <X.c onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"X.{a,{b,c}}"}}},

		// ---- Glob alignment: character class `[ABC]` ----
		{Code: `var x = <TestA onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test[ABC]"}}},
		{Code: `var x = <TestC onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test[ABC]"}}},

		// ---- Glob alignment: negated character class `[!A]` and `[^A]` ----
		{Code: `var x = <TestB onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test[!A]"}}},
		{Code: `var x = <TestB onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test[^A]"}}},

		// ---- Glob alignment: leading `!Foo` whole-pattern negation ----
		// `!Foo` matches everything except "Foo" — so `<Bar>` is ignored,
		// `<Foo>` is not.
		{Code: `var x = <Bar onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"!Foo"}}},

		// ---- Glob alignment: extglob `+(a|b)` and `@(a|b)` ----
		{Code: `var x = <Testa onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test+(a|b)"}}},
		{Code: `var x = <Testab onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test+(a|b)"}}},
		{Code: `var x = <Testa onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test@(a|b)"}}},

		// ---- Glob alignment: extglob `?(a)` and `*(ab)` ----
		// `?(a)` is "zero or one a" — both `Test` and `Testa` match.
		{Code: `var x = <Test onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test?(a)"}}},
		{Code: `var x = <Testa onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test?(a)"}}},
		// `*(ab)` is "zero or more ab" — `Test`, `Testab`, `Testabab` match.
		{Code: `var x = <Testabab onChange={whateverHandler} />`, Tsx: true, Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test*(ab)"}}},

		// ---- Configuration: both prefixes disabled = rule fully no-op ----
		// `eventHandlerPrefix: false, eventHandlerPropPrefix: false` means
		// neither half of the rule can fire. Even patterns that would
		// normally be invalid (`handleChange={this.bad}`, `onChange={x}`)
		// must produce no diagnostics.
		{Code: `var x = <X handleChange={this.handleChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": false, "eventHandlerPropPrefix": false}},
		{Code: `var x = <X onChange={this.doSomethingOnChange} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": false, "eventHandlerPropPrefix": false}},
		{Code: `var x = <X anything={anyValue} />`, Tsx: true, Options: map[string]interface{}{"eventHandlerPrefix": false, "eventHandlerPropPrefix": false, "checkLocalVariables": true}},

		// ---- Real-world: HOC + ref + complex render ----
		// Combines several patterns: ref skip, valid handler, deep chain.
		// Verifies no interference between attributes on the same element.
		{Code: `
        const Enhanced = withRouter(
          class extends React.Component {
            handleClick = () => this.handleAction();
            render() {
              return (
                <button
                  ref={this.btnRef}
                  onClick={this.handleClick}
                  onSubmit={this.props.onSubmit}
                />
              );
            }
          }
        );
      `, Tsx: true},

		// ---- Real-world: typed event handler ----
		// TypeScript event-handler prop with explicit type annotation —
		// type info doesn't affect the runtime expression shape; the rule
		// still sees PropertyAccessExpression as the value.
		{Code: `var x = <Button onClick={(e: React.MouseEvent<HTMLButtonElement>) => this.handleClick(e)} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true}},

		// ---- tsgo edge: trailing comment doesn't affect propValue ----
		// `getText` includes trailing comments only when they sit inside the
		// node's range. Trailing block comment after the property name does
		// fall inside `this.handleChange /* trailing */` — but it's stripped
		// out as whitespace-adjacent comment characters from the regex's
		// perspective, leaving `(.*\.)?handle...[A-Z]` matching.
		{Code: `var x = <TestComponent onChange={this.handleChange /* trailing */} />`, Tsx: true},

		// ---- tsgo edge: paren-wrapped optional chain (Dimension 4) ----
		// Both layers (paren + optional flag) need to be respected.
		{Code: `var x = <TestComponent onChange={(this?.handleChange)} />`, Tsx: true},

		// ---- tsgo edge: optional-call inline handler ----
		// `() => this?.handleChange?.()` — outer body is a CallExpression with
		// optional flag. The callee is also optional. With checkInlineFunction
		// + default checkLocal, the gate skips because callee is an optional
		// chain.
		{Code: `var x = <TestComponent onChange={() => this?.handleChange?.()} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true}},

		// ---- tsgo edge: arrow with block body, default options ----
		// Block-body arrow has no `body.callee`; under default options the
		// gate skips before propValue is read.
		{Code: `var x = <TestComponent onChange={() => { return this.handleChange(); }} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true}},

		// ---- tsgo edge: paren around arrow inline handler ----
		// `(() => this.handleChange())` — outer ParenthesizedExpression must
		// unwrap to ArrowFunction before isInline detection.
		{Code: `var x = <TestComponent onChange={(() => this.handleChange())} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true}},

		// ---- tsgo edge: paren around the call inside an arrow body ----
		// `() => (this.handleChange())` — inner paren wraps the CallExpression,
		// outer body is ParenthesizedExpression. SkipParentheses must unwrap
		// at every step.
		{Code: `var x = <TestComponent onChange={() => (this.handleChange())} />`, Tsx: true, Options: map[string]interface{}{"checkInlineFunction": true}},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: bad handler name (default options) ----
		{
			Code: `var x = <TestComponent onChange={this.doSomethingOnChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Message: "Handler function for onChange prop key must be a camelCase name beginning with 'handle' only", Line: 1, Column: 24},
			},
		},
		{
			Code: `var x = <TestComponent onChange={this.handlerChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},
		{
			Code: `var x = <TestComponent onChange={this.handle} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},
		{
			Code: `var x = <TestComponent onChange={this.handle2} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},
		{
			Code: `var x = <TestComponent onChange={this.handl3Change} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},
		{
			Code: `var x = <TestComponent onChange={this.handle4change} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: checkLocalVariables (local var with wrong name) ----
		{
			Code:    `var x = <TestComponent onChange={takeCareOfChange} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: checkInlineFunction (inline arrow with bad callee) ----
		{
			Code:    `var x = <TestComponent onChange={() => this.takeCareOfChange()} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkInlineFunction": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: bad prop key — non-`on` key paired with `handle*` ----
		// First case asserts the exact message text (covers upstream parity).
		{
			Code: `var x = <TestComponent only={this.handleChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Message: "Prop key for handleChange must begin with 'on'", Line: 1, Column: 24},
			},
		},
		{
			Code: `var x = <TestComponent2 only={this.handleChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Message: "Prop key for handleChange must begin with 'on'", Line: 1, Column: 25},
			},
		},
		{
			Code: `var x = <TestComponent handleChange={this.handleChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: checkLocalVariables + bad prop key ----
		{
			Code:    `var x = <TestComponent whenChange={handleChange} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: checkInlineFunction + checkLocalVariables + bad prop key ----
		{
			Code:    `var x = <TestComponent whenChange={() => handleChange()} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkInlineFunction": true, "checkLocalVariables": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: custom prefix `when` for prop, default `handle` for fn ----
		{
			Code:    `var x = <TestComponent onChange={handleChange} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "eventHandlerPrefix": "handle", "eventHandlerPropPrefix": "when"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Message: "Prop key for handleChange must begin with 'when'", Line: 1, Column: 24},
			},
		},
		{
			Code:    `var x = <TestComponent onChange={() => handleChange()} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkInlineFunction": true, "checkLocalVariables": true, "eventHandlerPrefix": "handle", "eventHandlerPropPrefix": "when"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: handler named `onChange` for `onChange` prop fails handle prefix ----
		{
			Code: `var x = <TestComponent onChange={this.onChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Upstream: bind-operator invalids — SKIP, TS does not parse `::` ----
		// SKIP: rslint's tsgo parser does not implement the `::` proposal.
		{
			Code: `var x = <TestComponent onChange={props::onChange} />`,
			Tsx:  true,
			Skip: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName"},
			},
		},
		{
			Code: `var x = <TestComponent onChange={props.foo::onChange} />`,
			Tsx:  true,
			Skip: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName"},
			},
		},

		// ---- Upstream: ignoreComponentNames non-matching pattern fires for each child ----
		{
			Code: `
        function App() {
          return (
            <div>
              <MyLibInput customPropNameBar={handleInput} />
              <MyLibCheckbox customPropNameBar={handleCheckbox} />
              <MyLibButtom customPropNameBar={handleButton} />
            </div>
          )
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"MyLibrary*"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badPropKey", Line: 5, Column: 27},
				{MessageId: "badPropKey", Line: 6, Column: 30},
				{MessageId: "badPropKey", Line: 7, Column: 28},
			},
		},

		// ---- Upstream: namespaced component, ignoreComponentNames patterns don't match ----
		{
			Code:    `var x = <A.TestComponent onChange={onChange} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"B.TestComponent", "TestComponent", "Test*"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 26},
			},
		},

		// ---- Lock-in: tsgo flat-AST equivalence with parens ----
		// Locks in the SkipParentheses unwrap before reading propValue / gating.
		// `(this.doSomethingOnChange)` must produce the same diagnostic as
		// `this.doSomethingOnChange` would — otherwise paren-wrapped code drifts.
		{
			Code: `var x = <TestComponent onChange={(this.doSomethingOnChange)} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},
		// Multi-paren: same expectation, locks recursive SkipParentheses unwrap.
		{
			Code: `var x = <TestComponent onChange={((this.doSomethingOnChange))} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Lock-in: element access (computed member) — text includes
		// brackets, so the handler regex never matches `handle`, and the
		// `onX` prop key fails the handler-name check. Both ESLint and
		// rslint behave identically here (the gate `expression.object` is
		// truthy on either MemberExpression flavor).
		{
			Code: `var x = <TestComponent onChange={this['handleChange']} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Lock-in: option-shape coverage for CLI / JSON path ----
		// Both the bare-object option form (single-option CLI shape) and the
		// array-wrapped form (rule-tester / multi-element shape) must reach
		// the rule's parser. A regression that only handles `[]interface{}`
		// would silently fall back to defaults on every CLI invocation.
		{
			Code:    `var x = <TestComponent onChange={takeCareOfChange} />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"checkLocalVariables": true}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Lock-in: nested JSX, multiple violations across siblings ----
		// Verifies the listener visits every JsxAttribute — not just the
		// outermost — and that each is reported independently.
		{
			Code: `
        function App() {
          return (
            <Outer onChange={this.bad1}>
              <Inner onClick={this.bad2} />
              <Sibling onSubmit={this.bad3} />
            </Outer>
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 4, Column: 20},
				{MessageId: "badHandlerName", Line: 5, Column: 22},
				{MessageId: "badHandlerName", Line: 6, Column: 24},
			},
		},

		// ---- Lock-in: multiple attributes on one element fire independently ----
		// `onChange` is bad, `onClick` is bad, `disabled` is irrelevant. Two
		// reports, one per offending attribute.
		{
			Code: `var x = <TestComponent disabled onChange={this.bad1} onClick={this.bad2} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 33},
				{MessageId: "badHandlerName", Line: 1, Column: 54},
			},
		},

		// ---- Lock-in: empty-string prefix falls back to default (invalid side) ----
		// Same input as the corresponding default-options case — empty string
		// must NOT change behavior, otherwise the option-parsing path drifts.
		{
			Code:    `var x = <TestComponent onChange={this.doSomethingOnChange} />`,
			Tsx:     true,
			Options: map[string]interface{}{"eventHandlerPrefix": ""},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Lock-in: optional chain still fires under checkLocalVariables ----
		// With checkLocal disabling the member-access gate, optional-chain
		// receivers don't get a free pass — the propValue regex still applies,
		// so a bad name reports. This mirrors ESLint's behavior: the gate
		// only skips when checkLocal is OFF; with it ON, the propValue test
		// runs regardless of the receiver shape.
		{
			Code:    `var x = <TestComponent onChange={this?.takeCareOfChange} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Lock-in: comment between `this.` and property name fires ----
		// `this. /* c */ handleChange` — the comment stays between `this.`
		// and `handleChange` after whitespace strip, breaking the
		// `(.*\.)?handle` greedy match (no further `.` to land on). Both
		// ESLint and rslint report; locks in that we mirror the upstream
		// regex's treatment of embedded comments.
		{
			Code: `var x = <TestComponent onChange={this. /* c */ handleChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Lock-in: bare local `onChange` identifier fails handler regex ----
		// Real-world prop-forwarding pitfall: `<Inner onChange={onChange} />`
		// with checkLocalVariables: true. The local Identifier `onChange`
		// matches the prop-key regex but NOT the handler regex (no `handle`,
		// no `props.on` prefix). Both ESLint and rslint report.
		{
			Code:    `var x = <Inner onChange={onChange} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 16},
			},
		},

		// ---- Lock-in: spread + bad handler — only the bad attribute fires ----
		// JsxSpreadAttribute is invisible to the rule; the named handler is
		// the only candidate.
		{
			Code: `var x = <TestComponent {...props} onChange={this.bad} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 35},
			},
		},

		// ---- Lock-in: deeply nested JSX hits inner attributes ----
		// 3-level nesting: outermost passes, innermost fails. Verifies the
		// listener doesn't stop at the first JsxOpeningElement.
		{
			Code: `
        function App() {
          return (
            <Outer onClick={this.handleClick}>
              <div>
                <span>
                  <Inner onChange={this.bad} />
                </span>
              </div>
            </Outer>
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 7, Column: 26},
			},
		},

		// ---- Lock-in: ignoreComponentNames last-pattern match wins ----
		// Iteration is "any pattern matches" — earlier non-matching patterns
		// must not short-circuit. Locks in the loop's "skip on any match"
		// semantics by putting the matching pattern at the end.
		{
			Code:    `var x = <Foo onChange={this.bad} />`,
			Tsx:     true,
			Options: map[string]interface{}{"ignoreComponentNames": []interface{}{"Bar", "Baz", "NeverMatches"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 14},
			},
		},

		// ---- Lock-in: rendering inside list map — inner JSX checked ----
		// Common iteration pattern: each rendered <Item> still goes through
		// the JsxAttribute listener.
		{
			Code: `
        function App() {
          return (
            <ul>
              {items.map(item => <Item key={item.id} onClick={this.bad} />)}
            </ul>
          );
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 5, Column: 54},
			},
		},

		// ---- Glob alignment: brace expansion DOES NOT match outside set ----
		// `Foo{1,2}` matches Foo1/Foo2 but NOT Foo3 — the rule then runs
		// normally on `<Foo3>` and fires badHandlerName because
		// `whateverHandler` doesn't satisfy the handler regex.
		{
			Code:    `var x = <Foo3 onChange={whateverHandler} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Foo{1,2}"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 15},
			},
		},

		// ---- Glob alignment: char class DOES NOT match outside set ----
		// `Test[ABC]` matches TestA/TestB/TestC but NOT TestD.
		{
			Code:    `var x = <TestD onChange={whateverHandler} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test[ABC]"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 16},
			},
		},

		// ---- Glob alignment: negated char class respects negation ----
		// `Test[!A]` matches TestB but NOT TestA.
		{
			Code:    `var x = <TestA onChange={whateverHandler} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test[!A]"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 16},
			},
		},

		// ---- Glob alignment: leading `!Foo` doesn't match Foo itself ----
		// `!Foo` ignores everything EXCEPT Foo, so `<Foo>` should still fire.
		{
			Code:    `var x = <Foo onChange={whateverHandler} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"!Foo"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 14},
			},
		},

		// ---- Glob alignment: extglob `+(a|b)` requires at least one ----
		// `Test+(a|b)` doesn't match bare "Test".
		{
			Code:    `var x = <Test onChange={whateverHandler} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkLocalVariables": true, "ignoreComponentNames": []interface{}{"Test+(a|b)"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 15},
			},
		},

		// ---- Lock-in: arrow with block body, checkLocal: true, checkInline: true ----
		// Block-body arrow has no `body.callee`. Upstream's `getText(undefined)`
		// returns the WHOLE source — a buggy code path that can produce
		// inconsistent results depending on what surrounds the JSX. We instead
		// treat "no callee" as empty propValue, which deterministically fires
		// `badHandlerName` (empty string can't satisfy the handler regex).
		// Locks in our saner-than-upstream behavior at this corner.
		{
			Code:    `var x = <TestComponent onChange={() => { return this.handleChange(); }} />`,
			Tsx:     true,
			Options: map[string]interface{}{"checkInlineFunction": true, "checkLocalVariables": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 24},
			},
		},

		// ---- Lock-in: full position assertion (single-line) ----
		// Locks in the JsxAttribute's full range as the diagnostic span. A
		// regression that reports on just the property name (or the whole
		// element) would surface here.
		{
			Code: `var x = <TestComponent onChange={this.bad} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "badHandlerName",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 43,
				},
			},
		},

		// ---- Lock-in: full position assertion (multi-line) ----
		// The JsxAttribute spans 4 lines (start of `onChange` on line 4,
		// closing `}` on line 6). Verifies multi-line range computation
		// matches ESLint's location semantics.
		{
			Code: `
        function App() {
          return <Outer
            onChange={
              this.bad
            } />;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "badHandlerName",
					Line:      4,
					Column:    13,
					EndLine:   6,
					EndColumn: 14,
				},
			},
		},

		// ---- Lock-in: real-world false-positive — `data.onChange` style ----
		// User passes `data.onChange` to `onChange` prop; the rule fires
		// because only `props.on*` is whitelisted (not `<any-name>.on*`).
		// This is a known upstream design choice; rslint must replicate.
		{
			Code: `var x = <Button onChange={data.onChange} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badHandlerName", Line: 1, Column: 17},
			},
		},

		// ---- Lock-in: data passed via Data field for downstream tooling ----
		// Upstream's `report({ data: { propKey, handlerPrefix } })` lets
		// IDE/CLI consumers re-format messages. We populate `Data` so the
		// same template variables are accessible.
		{
			Code: `var x = <X onChange={this.bad} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "badHandlerName",
					Message:   "Handler function for onChange prop key must be a camelCase name beginning with 'handle' only",
					Line:      1,
					Column:    12,
				},
			},
		},
	})
}
