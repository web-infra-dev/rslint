package jsx_no_leaked_render

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoLeakedRender(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoLeakedRenderRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases (default options) ----
		{
			Code: `
        const Component = () => {
          return <div>{customTitle || defaultTitle}</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements}</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>There are {elements.length} elements</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{!count && 'No results found'}</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{!!elements.length && <List elements={elements}/>}</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{Boolean(elements.length) && <List elements={elements}/>}</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements.length > 0 && <List elements={elements}/>}</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements.length ? <List elements={elements}/> : null}</div>
        }
      `,
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : null}</div>
        }
      `,
			Tsx: true,
		},

		// ---- Upstream valid: ternary-only strategy ----
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : null}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Tsx:     true,
		},

		// ---- Upstream valid: coerce-only strategy ----
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{!!count && <List elements={elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Upstream valid: explicit [coerce, ternary] (same as default) ----
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : null}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Tsx:     true,
		},
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{!!count && <List elements={elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Tsx:     true,
		},

		// ---- Upstream valid: should not delete valid alternates from ternaries ----
		// https://github.com/jsx-eslint/eslint-plugin-react/issues/3292
		// https://github.com/jsx-eslint/eslint-plugin-react/issues/3297
		{
			Code: `
        const Component = ({ elements, count }) => {
          return (
            <div>
              <div> {direction ? (direction === "down" ? "▼" : "▲") : ""} </div>
              <div>{ containerName.length > 0 ? "Loading several stuff" : "Loading" }</div>
            </div>
          )
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{direction ? (direction === "down" ? "▼" : "▲") : ""}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Tsx:     true,
		},

		// ---- Upstream valid: nested logical expressions are coerce-valid when leaves are valid ----
		{
			Code: `
        const Component = ({ direction }) => {
          return (
            <div>
              <div>{!!direction && direction === "down" && "▼"}</div>
              <div>{direction === "down" && !!direction && "▼"}</div>
              <div>{direction === "down" || !!direction && "▼"}</div>
              <div>{(!display || display === DISPLAY.WELCOME) && <span>foo</span>}</div>
            </div>
          )
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Upstream valid: ternary with JSX alternate ----
		// https://github.com/jsx-eslint/eslint-plugin-react/issues/3354
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : <EmptyList />}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Tsx:     true,
		},
		{
			Code: `
        const Component = ({ elements, count }) => {
          return <div>{count ? <List elements={elements}/> : <EmptyList />}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Upstream valid: const-bound boolean Identifier resolves through TypeChecker ----
		{
			Code: `
        const isOpen = true;
        const Component = () => {
          return <Popover open={isOpen && items.length > 0} />
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},
		{
			Code: `
        const isOpen = false;
        const Component = () => {
          return <Popover open={isOpen && items.length > 0} />
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Upstream valid: ignoreAttributes ----
		// https://github.com/jsx-eslint/eslint-plugin-react/issues/3292
		{
			Code: `
        const Component = ({ enabled, checked }) => {
          return <CheckBox checked={enabled && checked} />
        }
      `,
			Options: map[string]interface{}{"ignoreAttributes": true},
			Tsx:     true,
		},

		// ---- React 18+: empty-string left side is safe ----
		{
			Code: `
        const Example = () => {
          return <>{'' && <Something/>}</>
        }
      `,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.0.0"}},
			Tsx:      true,
		},
		// Default react version (latest) → React 18+ behavior.
		{
			Code: `
        const Example = () => {
          return <>{'' && <Something/>}</>
        }
      `,
			Tsx: true,
		},

		// ---- Edge: JsxFragment as alternate is treated as non-JSX (matches upstream
		// strict 'JSXElement' check) — but with default ['ternary', 'coerce'] the
		// conditional listener does not fire at all, so the case is valid. ----
		{
			Code: `
        const Example = ({ count }) => {
          return <div>{count ? <Foo/> : null}</div>
        }
      `,
			Tsx: true,
		},

		// ---- Edge: spread JsxExpression must not crash. ----
		{
			Code: `
        const Component = ({ rest }) => <div {...rest} />
      `,
			Tsx: true,
		},

		// ---- Edge: parenthesized && expression flattens transparently
		// (tsgo-specific: ESTree paren-flattens). With the coerce-valid left side
		// `Boolean(...)`, the case stays valid. ----
		{
			Code: `
        const Component = ({ x }) => {
          return <div>{(Boolean(x) && <Foo />)}</div>
        }
      `,
			Tsx: true,
		},

		// ---- Edge (tsgo-specific): multiple paren wrappers around a logical
		// chain whose leaves are all coerce-valid — the recursive
		// SkipParentheses inside isCoerceValidNestedLogical drills through.
		// Locks in upstream's intent (parens transparent) under tsgo's preserved
		// ParenthesizedExpression. ----
		{
			Code: `
        const Component = () => <div>{(((!a && b > 0))) && <Foo/>}</div>
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Edge (tsgo-specific): TS-only wrappers `as` / `satisfies` /
		// non-null assertion on a coerce-valid leaf — skipTypeAssertions sees
		// past them so no `!!` is needed. ----
		{
			Code: `
        const Component = ({ arr }: { arr: number[] }) => (
          <div>{(arr.length as number) > 0 && <Foo/>}</div>
        )
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},
		{
			Code: `
        const Component = ({ arr }: { arr?: number[] }) => (
          <div>{!!arr! && <Foo/>}</div>
        )
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Edge: JsxFragment as alternate is NOT JSXElement under upstream's
		// strict check — but with default validStrategies the conditional
		// listener is short-circuited, so the case stays valid. Locks in the
		// "default options skip ternaries unconditionally" path. ----
		{
			Code: `
        const Component = ({ count }) => <div>{count ? <Foo/> : <></>}</div>
      `,
			Tsx: true,
		},

		// ---- Edge: nested JsxExpression containing a `&&` whose own right side
		// contains another JsxExpression with a `&&`. Without coerce-valid
		// guards the OUTER `&&` reports; under coerce strategy with valid
		// guards on both, neither reports. ----
		{
			Code: `
        const Component = ({ a, b }) => (
          <div>
            {!!a && <Foo>{!!b && <Bar/>}</Foo>}
          </div>
        )
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Edge: empty JsxExpression `{}` must not crash. ----
		{
			Code: `
        const Component = () => <div>{}</div>
      `,
			Tsx: true,
		},

		// ---- Edge: Identifier resolves through `let`-bound variable whose
		// initializer is a boolean literal. Upstream's `defs[0].node.init.value`
		// path triggers regardless of `let` vs `const`; tsgo's GetDeclaration
		// yields the same VariableDeclaration. ----
		{
			Code: `
        let isOpen = true;
        const Component = () => <Popover open={isOpen && items.length > 0} />
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Tsx:     true,
		},

		// ---- Edge: spread JsxExpression inside JSX attribute (e.g.
		// `{...spread}`) sets dotDotDotToken. The listener's early-return on
		// dotDotDotToken protects against false-positive reports on the spread
		// argument. ----
		{
			Code: `
        const Component = ({ rest, count }) =>
          <Foo {...rest} bar={count ? <Bar/> : null} />
      `,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid: React 17 — '' is leaked ----
		{
			Code: `
        const Example = () => {
          return (
            <>
              {0 && <Something/>}
              {'' && <Something/>}
              {NaN && <Something/>}
            </>
          )
        }
      `,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "17.999.999"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noPotentialLeakedRender", Line: 5, Column: 16},
				{MessageId: "noPotentialLeakedRender", Line: 6, Column: 16},
				{MessageId: "noPotentialLeakedRender", Line: 7, Column: 16},
			},
			Output: []string{
				`
        const Example = () => {
          return (
            <>
              {0 ? <Something/> : null}
              {'' ? <Something/> : null}
              {NaN ? <Something/> : null}
            </>
          )
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: React 18 — '' is safe but 0 / NaN still leak ----
		{
			Code: `
        const Example = () => {
          return (
            <>
              {0 && <Something/>}
              {'' && <Something/>}
              {NaN && <Something/>}
            </>
          )
        }
      `,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "18.0.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noPotentialLeakedRender", Line: 5, Column: 16},
				{MessageId: "noPotentialLeakedRender", Line: 7, Column: 16},
			},
			Output: []string{
				`
        const Example = () => {
          return (
            <>
              {0 ? <Something/> : null}
              {'' && <Something/>}
              {NaN ? <Something/> : null}
            </>
          )
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: default options (ternary fix) ----
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{count && title}</div>
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count }) => {
          return <div>{count && <span>There are {count} results</span>}</div>
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count }) => {
          return <div>{count ? <span>There are {count} results</span> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements.length && <List elements={elements}/>}</div>
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ elements }) => {
          return <div>{elements.length ? <List elements={elements}/> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length ? <List elements={nestedCollection.elements}/> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements[0] && <List elements={elements}/>}</div>
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ elements }) => {
          return <div>{elements[0] ? <List elements={elements}/> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) ? <Results>{numberA+numberB}</Results> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		// Same as above, but with [coerce, ternary]: first strategy = coerce.
		{
			Code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ numberA, numberB }) => {
          return <div>{!!(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: ternary-only ----
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{count && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count }) => {
          return <div>{count && <span>There are {count} results</span>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count }) => {
          return <div>{count ? <span>There are {count} results</span> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements.length && <List elements={elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ elements }) => {
          return <div>{elements.length ? <List elements={elements}/> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length ? <List elements={nestedCollection.elements}/> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements[0] && <List elements={elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ elements }) => {
          return <div>{elements[0] ? <List elements={elements}/> : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) ? <Results>{numberA+numberB}</Results> : null}</div>
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: ternary-only — boolean-coerce on the left is not a free pass ----
		{
			Code: `
        const Component = ({ someCondition, title }) => {
          return <div>{!someCondition && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ someCondition, title }) => {
          return <div>{!someCondition ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{!!count && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{count > 0 && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{count > 0 ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{0 != count && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{0 != count ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count, total, title }) => {
          return <div>{count < total && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, total, title }) => {
          return <div>{count < total ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},
		// trimDoubleNot: `!!(count && somethingElse)` → ternary leaves the inner
		// LogicalExpression as-is (no `!!`), and tsgo's ParenthesizedExpression is
		// invisible in the emitted text.
		{
			Code: `
        const Component = ({ count, title, somethingElse }) => {
          return <div>{!!(count && somethingElse) && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title, somethingElse }) => {
          return <div>{count && somethingElse ? title : null}</div>
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: coerce-only ----
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{count && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{!!count && title}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count }) => {
          return <div>{count && <span>There are {count} results</span>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count }) => {
          return <div>{!!count && <span>There are {count} results</span>}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements.length && <List elements={elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ elements }) => {
          return <div>{!!elements.length && <List elements={elements}/>}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ nestedCollection }) => {
          return <div>{nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ nestedCollection }) => {
          return <div>{!!nestedCollection.elements.length && <List elements={nestedCollection.elements}/>}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ elements }) => {
          return <div>{elements[0] && <List elements={elements}/>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ elements }) => {
          return <div>{!!elements[0] && <List elements={elements}/>}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ numberA, numberB }) => {
          return <div>{(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ numberA, numberB }) => {
          return <div>{!!(numberA || numberB) && <Results>{numberA+numberB}</Results>}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ connection, hasError, hasErrorUpdate}) => {
          return <div>{connection && (hasError || hasErrorUpdate)}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ connection, hasError, hasErrorUpdate}) => {
          return <div>{!!connection && (hasError || hasErrorUpdate)}</div>
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: ternary not allowed when validStrategies=['coerce'] ----
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{count ? title : null}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{!!count && title}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{!count ? title : null}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{!count && title}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ count, somethingElse, title }) => {
          return <div>{count && somethingElse ? title : null}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, somethingElse, title }) => {
          return <div>{!!count && !!somethingElse && title}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const Component = ({ items, somethingElse, title }) => {
          return <div>{items.length > 0 && somethingElse && title}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ items, somethingElse, title }) => {
          return <div>{items.length > 0 && !!somethingElse && title}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const MyComponent = () => {
          const items = []
          const breakpoint = { phones: true }

          return <div>{items.length > 0 && breakpoint.phones && <span />}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 6, Column: 24}},
			Output: []string{
				`
        const MyComponent = () => {
          const items = []
          const breakpoint = { phones: true }

          return <div>{items.length > 0 && !!breakpoint.phones && <span />}</div>
        }
      `,
			},
			Tsx: true,
		},
		{
			Code: `
        const MyComponent = () => {
          return <div>{maybeObject && (isFoo ? <Aaa /> : <Bbb />)}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const MyComponent = () => {
          return <div>{!!maybeObject && (isFoo ? <Aaa /> : <Bbb />)}</div>
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: inside JSX attribute, default options ----
		// https://github.com/jsx-eslint/eslint-plugin-react/issues/3292
		{
			Code: `
        const Component = ({ enabled, checked }) => {
          return <CheckBox checked={enabled && checked} />
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 37}},
			Output: []string{
				`
        const Component = ({ enabled, checked }) => {
          return <CheckBox checked={enabled ? checked : null} />
        }
      `,
			},
			Tsx: true,
		},
		// ---- Upstream invalid: inverse ternary (`cond ? false : alt`) under coerce ----
		{
			Code: `
        const MyComponent = () => {
          return <Something checked={isIndeterminate ? false : isChecked} />
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 38}},
			Output: []string{
				`
        const MyComponent = () => {
          return <Something checked={!isIndeterminate && isChecked} />
        }
      `,
			},
			Tsx: true,
		},
		// ---- Inverse ternary with logical-and test (preserves `? false : alt`) ----
		{
			Code: `
        const MyComponent = () => {
          return <Something checked={cond && isIndeterminate ? false : isChecked} />
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 38}},
			Output: []string{
				`
        const MyComponent = () => {
          return <Something checked={!!cond && !!isIndeterminate ? false : isChecked} />
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: multi-line JSX right side, coerce strategy ----
		{
			Code: `
        const MyComponent = () => {
          return (
            <>
              {someCondition && (
                <div>
                  <p>hello</p>
                </div>
              )}
            </>
          )
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 5, Column: 16}},
			Output: []string{
				`
        const MyComponent = () => {
          return (
            <>
              {!!someCondition && (
                <div>
                  <p>hello</p>
                </div>
              )}
            </>
          )
        }
      `,
			},
			Tsx: true,
		},
		// ---- Multi-line JSX self-closing right side ----
		{
			Code: `
        const MyComponent = () => {
          return (
            <>
              {someCondition && (
                <SomeComponent
                  prop1={val1}
                  prop2={val2}
                />
              )}
            </>
          )
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce", "ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 5, Column: 16}},
			Output: []string{
				`
        const MyComponent = () => {
          return (
            <>
              {!!someCondition && (
                <SomeComponent
                  prop1={val1}
                  prop2={val2}
                />
              )}
            </>
          )
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: const-bound non-boolean Identifier still leaks ----
		{
			Code: `
        const isOpen = 0;
        const Component = () => {
          return <Popover open={isOpen && items.length > 0} />
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 4, Column: 33}},
			Output: []string{
				`
        const isOpen = 0;
        const Component = () => {
          return <Popover open={!!isOpen && items.length > 0} />
        }
      `,
			},
			Tsx: true,
		},

		// ---- Upstream invalid: ignoreAttributes only mutes attribute-level reports;
		// children expressions still report. ----
		{
			Code: `
        const Component = ({ enabled }) => {
          return (
            <Foo bar={
              <Something>{enabled && <MuchWow />}</Something>
            } />
          )
        }
      `,
			Options: map[string]interface{}{"ignoreAttributes": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 5, Column: 27}},
			Output: []string{
				`
        const Component = ({ enabled }) => {
          return (
            <Foo bar={
              <Something>{enabled ? <MuchWow /> : null}</Something>
            } />
          )
        }
      `,
			},
			Tsx: true,
		},

		// ---- Edge: paren-wrapped && at the JsxExpression root.
		// In ESTree this is paren-flattened; tsgo wraps it in
		// ParenthesizedExpression, but the listener treats the inner BinaryExpression
		// transparently and reports on the LogicalExpression's start. ----
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{(count && title)}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"ternary"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 25}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{(count ? title : null)}</div>
        }
      `,
			},
			Tsx: true,
		},

		// ---- Edge: `undefined` Identifier as ternary alternate (coerce-only) — value
		// of `undefined` Identifier is undefined → IN [undefined, null, false] →
		// invalid alternate → report. Locks in upstream's `node.alternate.value` arm
		// for non-Literal alternates. ----
		{
			Code: `
        const Component = ({ count, title }) => {
          return <div>{count ? title : undefined}</div>
        }
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 3, Column: 24}},
			Output: []string{
				`
        const Component = ({ count, title }) => {
          return <div>{!!count && title}</div>
        }
      `,
			},
			Tsx: true,
		},

		// ---- Edge: deep nesting — JsxExpression inside JsxExpression
		// recursion. Both inner and outer `&&` report independently (each is
		// its own JSXExpressionContainer > LogicalExpression). The outer fix
		// covers a range that contains the inner `&&`, so the two fixes
		// overlap; the linter applies them across two autofix iterations
		// (outer first, then inner). ----
		{
			Code: `
        const Component = ({ a, b }) => (
          <div>
            {a && <Foo>{b && <Bar/>}</Foo>}
          </div>
        )
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noPotentialLeakedRender", Line: 4, Column: 14},
				{MessageId: "noPotentialLeakedRender", Line: 4, Column: 25},
			},
			Output: []string{
				`
        const Component = ({ a, b }) => (
          <div>
            {a ? <Foo>{b && <Bar/>}</Foo> : null}
          </div>
        )
      `,
				`
        const Component = ({ a, b }) => (
          <div>
            {a ? <Foo>{b ? <Bar/> : null}</Foo> : null}
          </div>
        )
      `,
			},
			Tsx: true,
		},

		// ---- Edge: nested ConditionalExpression as right of `&&` — under
		// coerce, outer wraps right in parens, inner ternary preserved. ----
		{
			Code: `
        const Component = ({ a, isFoo }) => <div>{a && (isFoo ? <X/> : <Y/>)}</div>
      `,
			Options: map[string]interface{}{"validStrategies": []interface{}{"coerce"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noPotentialLeakedRender", Line: 2, Column: 51}},
			Output: []string{
				`
        const Component = ({ a, isFoo }) => <div>{!!a && (isFoo ? <X/> : <Y/>)}</div>
      `,
			},
			Tsx: true,
		},
	})
}
