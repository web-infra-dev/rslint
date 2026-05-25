package jsx_closing_bracket_location

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxClosingBracketLocationRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxClosingBracketLocationRule, []rule_tester.ValidTestCase{
		{
			Code: `
        <App />
      `,
			Tsx: true,
		},
		{
			Code: `
        <App foo />
      `,
			Tsx: true,
		},
		{
			Code: `
        <App
          foo
        />
      `,
			Tsx: true,
		},
		{
			Code: `
        <App foo />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "after-props"},
		},
		{
			Code: `
        <App foo />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        <App foo />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <App
          foo />
      `,
			Tsx:     true,
			Options: "after-props",
		},
		{
			Code: `
        <App
          foo
          />
      `,
			Tsx:     true,
			Options: "props-aligned",
		},
		{
			Code: `
        <App
          foo />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "after-props"},
		},
		{
			Code: `
        <App
          foo
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        <App
          foo
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <App
          foo
          />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "props-aligned"},
		},
		{
			Code: `
        <App foo></App>
      `,
			Tsx: true,
		},
		{
			Code: `
        <App
          foo
        ></App>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        <App
          foo
        ></App>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <App
          foo
          ></App>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "props-aligned"},
		},
		{
			Code: `
        <App
          foo={function() {
            console.log('bar');
          }} />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "after-props"},
		},
		{
			Code: `
        <App
          foo={function() {
            console.log('bar');
          }}
          />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "props-aligned"},
		},
		{
			Code: `
        <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <App foo={function() {
          console.log('bar');
        }}/>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "after-props"},
		},
		{
			Code: `
         <App foo={function() {
                console.log('bar');
              }}
              />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "props-aligned"},
		},
		{
			Code: `
        <App foo={function() {
          console.log('bar');
        }}
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        <App foo={function() {
          console.log('bar');
        }}
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <Provider store>
          <App
            foo />
        </Provider>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"selfClosing": "after-props"},
		},
		{
			Code: `
        <Provider
          store
        >
          <App
            foo />
        </Provider>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"selfClosing": "after-props"},
		},
		{
			Code: `
        <Provider
          store>
          <App
            foo
          />
        </Provider>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"nonEmpty": "after-props"},
		},
		{
			Code: `
        <Provider store>
          <App
            foo
            />
        </Provider>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"selfClosing": "props-aligned"},
		},
		{
			Code: `
        <Provider
          store
          >
          <App
            foo
          />
        </Provider>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"nonEmpty": "props-aligned"},
		},
		{
			Code: `
        var x = function() {
          return <App
            foo
                 >
              bar
                 </App>
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        var x = function() {
          return <App
            foo
                 />
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        var x = <App
          foo
                />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		{
			Code: `
        var x = function() {
          return <App
            foo={function() {
              console.log('bar');
            }}
          />
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        var x = <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <Provider
          store
        >
          <App
            foo={function() {
              console.log('bar');
            }}
          />
        </Provider>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <Provider
          store
        >
          {baz && <App
            foo={function() {
              console.log('bar');
            }}
          />}
        </Provider>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		{
			Code: `
        <App>
          <Foo
            bar
          >
          </Foo>
          <Foo
            bar />
        </App>
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"nonEmpty":    false,
				"selfClosing": "after-props",
			},
		},
		{
			Code: `
        <App>
          <Foo
            bar>
          </Foo>
          <Foo
            bar
          />
        </App>
      `,
			Tsx: true,
			Options: map[string]interface{}{
				"nonEmpty":    "after-props",
				"selfClosing": false,
			},
		},
		{
			Code: `
        <div className={[
          "some",
          "stuff",
          2 ]}
        >
          Some text
        </div>
      `,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		// ---- TS / non-Identifier tag forms (rslint-extra) ----
		// JsxMemberExpression tag (`<Foo.Bar />`) — after-tag, no props.
		{
			Code: `<Foo.Bar />`,
			Tsx:  true,
		},
		// JsxMemberExpression tag, single-line with prop — after-props.
		{
			Code: `<Foo.Bar foo />`,
			Tsx:  true,
		},
		// JsxMemberExpression tag, multi-line — default tag-aligned.
		{
			Code: `<Foo.Bar
  foo
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		// JsxMemberExpression tag, non-self-closing multi-line.
		{
			Code: `<Foo.Bar
  foo
></Foo.Bar>`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		// JsxNamespacedName tag (`<svg:circle />`) — after-tag.
		{
			Code: `<svg:circle />`,
			Tsx:  true,
		},
		// JsxNamespacedName tag, multi-line tag-aligned.
		{
			Code: `<svg:circle
  cx={1}
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		// TS type arguments (`<App<T> foo />`) — single line, after-props.
		{
			Code: `<App<T> foo />`,
			Tsx:  true,
		},
		// TS type arguments, multi-line tag-aligned.
		{
			Code: `<App<T>
  foo
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		// Expression containing '>' as a JSX attribute value — closing-bracket
		// detection must not be confused by the expression-internal '>'.
		{
			Code: `<App foo={x > y} />`,
			Tsx:  true,
		},
		// Nested JSX as attribute value — both elements are valid.
		{
			Code: `<App content={<Foo />} />`,
			Tsx:  true,
		},
		// Nested JSX as attribute value, multi-line outer.
		{
			Code: `<App
  content={<Foo />}
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		// ---- JsxSpreadAttribute coverage (rslint-extra) ----
		// Spread is the only attribute, single line — after-props.
		{
			Code: `<App {...props} />`,
			Tsx:  true,
		},
		// Spread mixed with regular attribute, single line.
		{
			Code: `<App foo {...rest} />`,
			Tsx:  true,
		},
		// Spread is the only attribute, multi-line tag-aligned default.
		{
			Code: `<App
  {...spread}
/>`,
			Tsx: true,
		},
		// Multi-line with spread as last attribute — tag-aligned default.
		{
			Code: `<App
  foo
  {...rest}
/>`,
			Tsx: true,
		},
		// Spread in the middle, regular attribute at the tail.
		{
			Code: `<App
  foo
  {...mid}
  bar
/>`,
			Tsx: true,
		},
		// Spread is last attribute, props-aligned (closing aligned with spread).
		{
			Code: `<App
  foo
  {...rest}
  />`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "props-aligned"},
		},
		// ---- Common attribute shapes ----
		// Empty string attribute.
		{
			Code: `<App foo="" />`,
			Tsx:  true,
		},
		// Boolean (shorthand) attribute.
		{
			Code: `<App disabled />`,
			Tsx:  true,
		},
		// TS `as` expression as attribute value — '>' inside the expression must
		// not be mistaken for the closing bracket.
		{
			Code: `<App foo={x as T} />`,
			Tsx:  true,
		},
		// TS non-null assertion as attribute value.
		{
			Code: `<App foo={x!} />`,
			Tsx:  true,
		},
		// Optional chain as attribute value.
		{
			Code: `<App foo={x?.y} />`,
			Tsx:  true,
		},
		// `satisfies` expression as attribute value (TS 4.9+).
		{
			Code: `<App foo={x satisfies T} />`,
			Tsx:  true,
		},
		// ---- Fragment / nesting / arrow-fn returning JSX ----
		// Rule must not crash on JsxFragment / JsxOpeningFragment;
		// the inner element should still be checked correctly.
		{
			Code: `<>
  <App foo />
</>`,
			Tsx: true,
		},
		// Fragment with multiple sibling elements.
		{
			Code: `<>
  <A />
  <B />
</>`,
			Tsx: true,
		},
		// Multi-level nested elements — every OpeningElement processed
		// independently, no boundary leak.
		{
			Code: `<App>
  <B>
    <C />
  </B>
</App>`,
			Tsx: true,
		},
		// Arrow function that returns multi-line JSX (line-aligned).
		{
			Code: `var f = () => <App
  foo
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		// JSX inside an array literal.
		{
			Code: `var arr = [<A />, <B />];`,
			Tsx:  true,
		},
		// ---- options JSON path robustness ----
		// Array-wrapped string (matches rule_tester / multi-element CLI shape).
		{
			Code:    `<App foo />`,
			Tsx:     true,
			Options: []interface{}{"after-props"},
		},
		// Array-wrapped map (matches rule_tester / multi-element CLI shape).
		{
			Code:    `<App foo />`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"location": "after-props"}},
		},
		// Empty array of options should fall back to defaults.
		{
			Code:    `<App />`,
			Tsx:     true,
			Options: []interface{}{},
		},
		// nonEmpty:false alone — closing tag of `<App\n>...</App>` is unchecked.
		{
			Code: `<App
  foo
  ></App>`,
			Tsx:     true,
			Options: map[string]interface{}{"nonEmpty": false},
		},
		// selfClosing:false alone — closing tag of `<App\n  foo\n  />` is unchecked.
		{
			Code: `<App
  foo
  />`,
			Tsx:     true,
			Options: map[string]interface{}{"selfClosing": false},
		},
		// nonEmpty:false AND selfClosing:false — rule fully disabled for any
		// shape (sanity: rule should not crash, no diagnostics emitted).
		{
			Code: `<>
  <App
  foo
  />
  <App
  foo
  ></App>
</>`,
			Tsx:     true,
			Options: map[string]interface{}{"nonEmpty": false, "selfClosing": false},
		},
		// Unknown enum string — ESLint schema rejects, but rslint should
		// degrade gracefully (default behavior; do not panic).
		{
			Code:    `<App />`,
			Tsx:     true,
			Options: "unknown-location-value",
		},
		// Numeric / boolean options — invalid per schema, must not crash.
		{
			Code:    `<App />`,
			Tsx:     true,
			Options: 42,
		},
		{
			Code:    `<App />`,
			Tsx:     true,
			Options: true,
		},
		// Map with unknown key — ignored, defaults take effect.
		{
			Code:    `<App foo />`,
			Tsx:     true,
			Options: map[string]interface{}{"unknownKey": "tag-aligned"},
		},
		// nonEmpty / selfClosing as non-string non-false (true) — ignored.
		{
			Code:    `<App foo />`,
			Tsx:     true,
			Options: map[string]interface{}{"nonEmpty": true, "selfClosing": true},
		},
		// ---- CRLF line terminators (Windows) ----
		// CRLF-terminated source, default tag-aligned — valid.
		{
			Code: "<App\r\n  foo\r\n/>",
			Tsx:  true,
		},
		// CRLF, options object, nested with provider.
		{
			Code:    "<Provider\r\n  store\r\n>\r\n  <App\r\n    foo\r\n  />\r\n</Provider>",
			Tsx:     true,
			Options: map[string]interface{}{"location": "tag-aligned"},
		},
		// ---- JSX expression-internal comments ----
		// `/* */` comment inside an attribute value expression — closing
		// detection must not be confused.
		{
			Code: `<App foo={/* leading */ 1} />`,
			Tsx:  true,
		},
		// `//` line comment inside an attribute value expression spanning
		// multiple lines.
		{
			Code: `<App
  foo={
    // line comment
    1
  }
/>`,
			Tsx: true,
		},
		// ---- More TS generic forms ----
		// Generic with multiple type arguments.
		{
			Code: `<App<T, U> foo />`,
			Tsx:  true,
		},
		// Generic with multiple type arguments, multi-line.
		{
			Code: `<App<T, U>
  foo
/>`,
			Tsx: true,
		},
		// ---- Non-self-closing symmetry (no props) ----
		// `<App></App>` and `<App\n></App>` — after-tag rule applies the same
		// way to non-self-closing elements with zero attributes.
		{
			Code: `<App></App>`,
			Tsx:  true,
		},
		// ---- Large realistic React component lock-in ----
		// Verifies the rule handles a deeply nested, multi-prop, mixed
		// self-closing / non-self-closing tree without false positives.
		{
			Code: `<Provider
  value={{
    user,
    onLogin: () => {},
  }}
>
  <App
    title="Home"
    subtitle={` + "`Welcome ${user.name}`" + `}
    onClose={handleClose}
  />
  <Footer
    links={[
      { href: '/a', label: 'A' },
      { href: '/b', label: 'B' },
    ]}
  />
</Provider>`,
			Tsx: true,
		},
		// ---- Spread expression internal trivia ----
		// Comment leading the spread argument.
		{
			Code: `<App {/* leading */ ...rest} />`,
			Tsx:  true,
		},
		// Comment trailing the spread argument.
		{
			Code: `<App {...rest /* trailing */} />`,
			Tsx:  true,
		},
		// Spread argument is itself a multi-line object expression.
		{
			Code: `<App
  {...{
    ...defaults,
    extra: 1,
  }}
/>`,
			Tsx: true,
		},
		// ---- Attribute values: fragment / ternary / template ----
		// JSX fragment as attribute value.
		{
			Code: `<App content={<>foo</>} />`,
			Tsx:  true,
		},
		// Multi-line ternary as attribute value (closing of outer JSX must
		// not be confused by the ternary's contents).
		{
			Code: `<App
  className={
    x
      ? 'a'
      : 'b'
  }
/>`,
			Tsx: true,
		},
		// Template literal with embedded expression as attribute value.
		{
			Code: `<App foo={` + "`hello ${name} world`" + `} />`,
			Tsx:  true,
		},
		// JSX as logical-AND right operand (line-aligned: closing at line indent).
		{
			Code: `var x = cond && <App
  foo
/>`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		// JSX as ternary branch (line-aligned).
		{
			Code: `var x = cond ? <App
  foo
/> : null`,
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
		},
		// JSX as logical-AND right operand, tag-aligned (closing at `<` column).
		{
			Code: `var x = cond && <App
                  foo
                />`,
			Tsx: true,
		},
		// ---- Multi-byte / unicode attribute name lock-in ----
		// Lock-in: rule does not crash on JSX attribute / tag names that
		// contain non-ASCII identifier characters. Closing detection works.
		{
			Code: `<App 中文Prop="x" />`,
			Tsx:  true,
		},
		// ---- Spread BEFORE regular attribute (reverse order) ----
		// Symmetry check: I already cover `foo {...rest}`; reverse order
		// `{...rest} foo` should behave identically (lastProp = foo).
		{
			Code: `<App {...rest} foo />`,
			Tsx:  true,
		},
		// Multi-line variant: spread first, regular tail, default tag-aligned.
		{
			Code: `<App
  {...rest}
  foo
/>`,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- after-tag (no props) ----
		{
			Code: `
        <App
        />
      `,
			Tsx: true,
			Output: []string{`
        <App />
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the opening tag",
				},
			},
		},
		// ---- after-props (default, lastProp on opening line) ----
		{
			Code: `
        <App foo
        />
      `,
			Tsx: true,
			Output: []string{`
        <App foo/>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: `
        <App foo
        ></App>
      `,
			Tsx: true,
			Output: []string{`
        <App foo></App>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		// ---- props-aligned vs current after-props -> needs newline ----
		{
			Code: `
        <App
          foo />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
          />
      `},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 11 on the next line)",
					Line:      3,
					Column:    15,
				},
			},
		},
		{
			Code: `
        <App
          foo />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
        />
      `},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
					Line:      3,
					Column:    15,
				},
			},
		},
		{
			Code: `
        <App
          foo />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
        />
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9 on the next line)",
					Line:      3,
					Column:    15,
				},
			},
		},
		// ---- after-props: collapse closing onto last prop line ----
		{
			Code: `
        <App
          foo
        />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo/>
      `},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: `
        <App
          foo
        />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
          />
      `},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 11)",
					Line:      4,
					Column:    9,
				},
			},
		},
		{
			Code: `
        <App
          foo
          />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo/>
      `},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: `
        <App
          foo
          />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
        />
      `},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 9)",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
        <App
          foo
          />
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
        />
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
					Line:      4,
					Column:    11,
				},
			},
		},
		// ---- non-self-closing equivalents ----
		{
			Code: `
        <App
          foo
        ></App>
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo></App>
      `},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: `
        <App
          foo
        ></App>
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
          ></App>
      `},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 11)",
					Line:      4,
					Column:    9,
				},
			},
		},
		{
			Code: `
        <App
          foo
          ></App>
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo></App>
      `},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: `
        <App
          foo
          ></App>
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
        ></App>
      `},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 9)",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
        <App
          foo
          ></App>
      `,
			Tsx: true,
			Output: []string{`
        <App
          foo
        ></App>
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
					Line:      4,
					Column:    11,
				},
			},
		},
		// ---- nested elements with selfClosing/nonEmpty ----
		{
			Code: `
        <Provider
          store>
          <App
            foo
            />
        </Provider>
      `,
			Tsx: true,
			Output: []string{`
        <Provider
          store
        >
          <App
            foo
            />
        </Provider>
      `},
			Options: map[string]interface{}{"selfClosing": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
					Line:      3,
					Column:    16,
				},
			},
		},
		// ---- props-aligned with very-far closing bracket ----
		{
			Code: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
                                            >
              Button Text
            </Button>
          );
        };
      `,
			Tsx: true,
			Output: []string{`
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
              >
              Button Text
            </Button>
          );
        };
      `},
			Options: "props-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 15)",
					Line:      7,
					Column:    45,
				},
			},
		},
		{
			Code: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
                                            >
              Button Text
            </Button>
          );
        };
      `,
			Tsx: true,
			Output: []string{`
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
            >
              Button Text
            </Button>
          );
        };
      `},
			Options: "tag-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 13)",
					Line:      7,
					Column:    45,
				},
			},
		},
		{
			Code: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
                                            >
              Button Text
            </Button>
          );
        };
      `,
			Tsx: true,
			Output: []string{`
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
            >
              Button Text
            </Button>
          );
        };
      `},
			Options: "line-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 13)",
					Line:      7,
					Column:    45,
				},
			},
		},
		// ---- nonEmpty=props-aligned: nested App fix ----
		{
			Code: `
        <Provider
          store
          >
          <App
            foo
            />
        </Provider>
      `,
			Tsx: true,
			Output: []string{`
        <Provider
          store
          >
          <App
            foo
          />
        </Provider>
      `},
			Options: map[string]interface{}{"nonEmpty": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
					Line:      7,
					Column:    13,
				},
			},
		},
		{
			Code: `
        <Provider
          store>
          <App
            foo />
        </Provider>
      `,
			Tsx: true,
			Output: []string{`
        <Provider
          store
        >
          <App
            foo />
        </Provider>
      `},
			Options: map[string]interface{}{"selfClosing": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
					Line:      3,
					Column:    16,
				},
			},
		},
		{
			Code: `
        <Provider
          store>
          <App
            foo
            />
        </Provider>
      `,
			Tsx: true,
			Output: []string{`
        <Provider
          store>
          <App
            foo
          />
        </Provider>
      `},
			Options: map[string]interface{}{"nonEmpty": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
					Line:      6,
					Column:    13,
				},
			},
		},
		// ---- line-aligned vs tag-aligned distinction ----
		{
			Code: `
        var x = function() {
          return <App
            foo
                />
        }
      `,
			Tsx: true,
			Output: []string{`
        var x = function() {
          return <App
            foo
          />
        }
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 11)",
					Line:      5,
					Column:    17,
				},
			},
		},
		{
			Code: `
        var x = <App
          foo
                />
      `,
			Tsx: true,
			Output: []string{`
        var x = <App
          foo
        />
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 9)",
					Line:      4,
					Column:    17,
				},
			},
		},
		{
			Code: `
        var x = (
          <div
            className="MyComponent"
            {...props} />
        )
      `,
			Tsx: true,
			Output: []string{`
        var x = (
          <div
            className="MyComponent"
            {...props}
          />
        )
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 11 on the next line)",
					Line:      5,
					Column:    24,
				},
			},
		},
		{
			Code: `
        var x = (
          <Something
            content={<Foo />} />
        )
      `,
			Tsx: true,
			Output: []string{`
        var x = (
          <Something
            content={<Foo />}
          />
        )
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 11 on the next line)",
					Line:      4,
					Column:    31,
				},
			},
		},
		{
			Code: `
        var x = (
          <Something
            />
        )
      `,
			Tsx: true,
			Output: []string{`
        var x = (
          <Something />
        )
      `},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the opening tag",
				},
			},
		},
		// ---- multi-line array attribute -> tag-aligned forces newline ----
		{
			Code: `
        <div className={[
          "some",
          "stuff",
          2 ]}>
          Some text
        </div>
      `,
			Tsx: true,
			Output: []string{`
        <div className={[
          "some",
          "stuff",
          2 ]}
        >
          Some text
        </div>
      `},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 9 on the next line)",
					Line:      5,
					Column:    15,
				},
			},
		},
		// ---- tab-indented variants (props-aligned / tag-aligned / line-aligned) ----
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo />\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 6 on the next line)",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo />\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo />\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5 on the next line)",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 6)",
					Line:      4,
					Column:    5,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 5)",
					Line:      4,
					Column:    6,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5)",
					Line:      4,
					Column:    6,
				},
			},
		},
		// ---- non-self-closing tab variants ----
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo></App>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 6)",
					Line:      4,
					Column:    5,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo></App>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 5)",
					Line:      4,
					Column:    6,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t\t></App>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t></App>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5)",
					Line:      4,
					Column:    6,
				},
			},
		},
		// ---- nested with selfClosing/nonEmpty (tab) ----
		{
			Code: "\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			},
			Options: map[string]interface{}{"selfClosing": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
					Line:      3,
					Column:    11,
				},
			},
		},
		// ---- props-aligned/tag-aligned/line-aligned with very-far closing bracket (tab) ----
		{
			Code: "\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			},
			Options: "props-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 8)",
					Line:      7,
					Column:    23,
				},
			},
		},
		{
			Code: "\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			},
			Options: "tag-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 7)",
					Line:      7,
					Column:    23,
				},
			},
		},
		{
			Code: "\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tconst Button = function(props) {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Button\n\t\t\t\t\t\t\tsize={size}\n\t\t\t\t\t\t\tonClick={onClick}\n\t\t\t\t\t\t>\n\t\t\t\t\t\t\tButton Text\n\t\t\t\t\t\t</Button>\n\t\t\t\t\t);\n\t\t\t\t};\n\t\t\t",
			},
			Options: "line-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 7)",
					Line:      7,
					Column:    23,
				},
			},
		},
		// ---- nonEmpty=props-aligned: nested App fix (tab) ----
		{
			Code: "\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			},
			Options: map[string]interface{}{"nonEmpty": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 6)",
					Line:      7,
					Column:    7,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo />\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<Provider\n\t\t\t\t\tstore\n\t\t\t\t>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo />\n\t\t\t\t</Provider>\n\t\t\t",
			},
			Options: map[string]interface{}{"selfClosing": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: "\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<Provider\n\t\t\t\t\tstore>\n\t\t\t\t\t<App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t\t</Provider>\n\t\t\t",
			},
			Options: map[string]interface{}{"nonEmpty": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 6)",
					Line:      6,
					Column:    7,
				},
			},
		},
		{
			Code: "\n\t\t\t\tvar x = function() {\n\t\t\t\t\treturn <App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t\t\t\t/>\n\t\t\t\t}\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tvar x = function() {\n\t\t\t\t\treturn <App\n\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t\t}\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 6)",
					Line:      5,
					Column:    9,
				},
			},
		},
		{
			Code: "\n\t\t\t\tvar x = <App\n\t\t\t\t\tfoo\n\t\t\t\t\t\t\t\t/>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tvar x = <App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 5)",
					Line:      4,
					Column:    9,
				},
			},
		},
		{
			Code: "\n\t\t\t\tvar x = (\n\t\t\t\t\t<div\n\t\t\t\t\t\tclassName=\"MyComponent\"\n\t\t\t\t\t\t{...props} />\n\t\t\t\t)\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tvar x = (\n\t\t\t\t\t<div\n\t\t\t\t\t\tclassName=\"MyComponent\"\n\t\t\t\t\t\t{...props}\n\t\t\t\t\t/>\n\t\t\t\t)\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 6 on the next line)",
					Line:      5,
					Column:    18,
				},
			},
		},
		{
			Code: "\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something\n\t\t\t\t\t\tcontent={<Foo />} />\n\t\t\t\t)\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something\n\t\t\t\t\t\tcontent={<Foo />}\n\t\t\t\t\t/>\n\t\t\t\t)\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 6 on the next line)",
					Line:      4,
					Column:    25,
				},
			},
		},
		{
			Code: "\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something\n\t\t\t\t\t\t/>\n\t\t\t\t)\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\tvar x = (\n\t\t\t\t\t<Something />\n\t\t\t\t)\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the opening tag",
				},
			},
		},
		{
			Code: "\n\t\t\t\t<div className={[\n\t\t\t\t\t\"some\",\n\t\t\t\t\t\"stuff\",\n\t\t\t\t\t2 ]}>\n\t\t\t\t\tSome text\n\t\t\t\t</div>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t<div className={[\n\t\t\t\t\t\"some\",\n\t\t\t\t\t\"stuff\",\n\t\t\t\t\t2 ]}\n\t\t\t\t>\n\t\t\t\t\tSome text\n\t\t\t\t</div>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 5 on the next line)",
					Line:      5,
					Column:    10,
				},
			},
		},
		{
			Code: "\n\t\t\t\t\t\t\t<div\n\t\t\t\t\t\t\t\tclassName={styles}\n\t\t\t\t\t >\n\t\t\t\t\t\t\t\t{props}\n\t\t\t\t\t\t\t</div>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t\t\t\t<div\n\t\t\t\t\t\t\t\tclassName={styles}\n\t\t\t\t\t\t\t>\n\t\t\t\t\t\t\t\t{props}\n\t\t\t\t\t\t\t</div>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 8)",
					Line:      4,
					Column:    7,
				},
			},
		},
		{
			Code: `
          <div
            className={styles}
            >
            {props}
          </div>
      `,
			Tsx: true,
			Output: []string{`
          <div
            className={styles}
          >
            {props}
          </div>
      `},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
					Line:      4,
					Column:    13,
				},
			},
		},
		{
			Code: `
          <App
            foo
            />
      `,
			Tsx: true,
			Output: []string{`
          <App
            foo
          />
      `},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 11)",
					Line:      4,
					Column:    13,
				},
			},
		},
		// Mixed-tab edge case: opening line uses 6 tabs, prop line 7 tabs, closing line 5 tabs
		{
			Code: "\n\t\t\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t\t/>\n\t\t\t",
			Tsx:  true,
			Output: []string{
				"\n\t\t\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t\t\t/>\n\t\t\t",
			},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 7)",
					Line:      4,
					Column:    6,
				},
			},
		},
		// ---- TS / non-Identifier tag forms (rslint-extra) ----
		// JsxMemberExpression tag with multi-line prop — column accounting must
		// use the '<' position, not anywhere inside the dotted name.
		{
			Code: `<Foo.Bar
  foo />`,
			Tsx: true,
			Output: []string{`<Foo.Bar
  foo
/>`},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 1 on the next line)",
					Line:      2,
					Column:    7,
				},
			},
		},
		// JsxNamespacedName tag with multi-line prop.
		{
			Code: `<svg:circle
  cx={1} />`,
			Tsx: true,
			Output: []string{`<svg:circle
  cx={1}
/>`},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 1 on the next line)",
					Line:      2,
					Column:    10,
				},
			},
		},
		// Expression containing '>' on a later line — closing detection must
		// pick the JSX-element '>' / '/' even when '>' chars appear in
		// attribute expressions earlier in the source.
		{
			Code: `<App foo={x > y}
/>`,
			Tsx: true,
			Output: []string{`<App foo={x > y}/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		// Nested JSX as attribute value, multi-line outer with misaligned
		// closing — locks in that the inner `<Foo />` doesn't disturb the
		// outer closing's column accounting.
		{
			Code: `<App
  content={<Foo />}
  />`,
			Tsx: true,
			Output: []string{`<App
  content={<Foo />}
/>`},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
					Line:      3,
					Column:    3,
				},
			},
		},
		// ---- Spread attribute as last prop ----
		// Single spread attribute as lastProp, after-tag → after-props collapse.
		{
			Code: `<App {...props}
/>`,
			Tsx:    true,
			Output: []string{`<App {...props}/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		// Multi-line, spread is last prop, default tag-aligned, misaligned `/>`.
		{
			Code: `<App
  foo
  {...rest}
  />`,
			Tsx: true,
			Output: []string{`<App
  foo
  {...rest}
/>`},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
					Line:      4,
					Column:    3,
				},
			},
		},
		// Spread is last prop, on the same line as `/>`, after-props expects collapse.
		{
			Code: `<App foo {...rest}
/>`,
			Tsx:    true,
			Output: []string{`<App foo {...rest}/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the last prop",
				},
			},
		},
		// Spread is last prop, props-aligned with `/>` on next line at wrong column.
		{
			Code: `<App
  foo
  {...rest} />`,
			Tsx: true,
			Output: []string{`<App
  foo
  {...rest}
  />`},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the last prop (expected column 3 on the next line)",
					Line:      3,
					Column:    13,
				},
			},
		},
		// ---- Multi-level nested elements ----
		// Inner element misalign — verify outer/inner are independently checked.
		{
			Code: `<App>
  <B
    bar
    />
</App>`,
			Tsx: true,
			Output: []string{`<App>
  <B
    bar
  />
</App>`},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 3)",
					Line:      4,
					Column:    5,
				},
			},
		},
		// ---- Arrow function returning multi-line JSX ----
		// `() => <App\n  foo\n />` line-aligned — `/>` should align with the
		// line containing `() =>`, not the `<App` column.
		{
			Code: `var f = () => <App
  foo
 />`,
			Tsx: true,
			Output: []string{`var f = () => <App
  foo
/>`},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 1)",
					Line:      3,
					Column:    2,
				},
			},
		},
		// ---- Conditional rendering inside parent JSX ----
		// `{cond && <App\n  foo />}` line-aligned — `/>` should align with the
		// expression's outer line indent.
		{
			Code: `<Wrapper>
  {cond && <App
    foo />}
</Wrapper>`,
			Tsx: true,
			Output: []string{`<Wrapper>
  {cond && <App
    foo
  />}
</Wrapper>`},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 3 on the next line)",
					Line:      3,
					Column:    9,
				},
			},
		},
		// ---- Options-error fallback still triggers default behavior ----
		// Even with non-string non-map options, default tag-aligned applies.
		{
			Code: `<App
  foo
  />`,
			Tsx:     true,
			Options: 42,
			Output: []string{`<App
  foo
/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
					Line:      3,
					Column:    3,
				},
			},
		},
		// Map with unknown key — defaults still apply.
		{
			Code: `<App
  foo
  />`,
			Tsx:     true,
			Options: map[string]interface{}{"unknownKey": "props-aligned"},
			Output: []string{`<App
  foo
/>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
					Line:      3,
					Column:    3,
				},
			},
		},
		// ---- CRLF line terminators (Windows) — invalid ----
		// CRLF source with misaligned `/>` should report and fix correctly.
		{
			Code: "<App\r\n  foo\r\n  />",
			Tsx:  true,
			Output: []string{
				"<App\r\n  foo\n/>",
			},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
					Line:      3,
					Column:    3,
				},
			},
		},
		// ---- Non-self-closing symmetry: `<App\n></App>` after-tag invalid ----
		// Mirror of upstream invalid case 1 (`<App\n/>`) for non-self-closing form.
		{
			Code: `<App
></App>`,
			Tsx:    true,
			Output: []string{`<App ></App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the opening tag",
				},
			},
		},
		// ---- Multi-byte character column lock-in ----
		// Opening line contains multi-byte identifier characters (中文 = 6 bytes
		// UTF-8 / 2 UTF-16 code units). The "expected column N" in the message
		// MUST match the UTF-16 column shown in the user's IDE — not the byte
		// offset within the line.
		//
		// `var 中文 = <App` → `<` is at:
		//   - UTF-16 column 10 (1-based): "var 中文 = " = 4+2+3 = 9 units, '<' at 10
		//   - byte column 14 (1-based): "var " 4 + "中文" 6 + " = " 3 = 13 bytes, '<' at 14
		// We assert the message contains "expected column 10" (UTF-16, matches IDE).
		{
			Code: `var x = (() => {
  return <中文Comp
    foo />
})`,
			Tsx: true,
			Output: []string{`var x = (() => {
  return <中文Comp
    foo
  />
})`},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					// Line 2 (`  return <中文Comp`) is the opening line.
					// `line-aligned` indent = 2 spaces (line 2 leading whitespace).
					// "expected column 3" (1-based UTF-16) — same as IDE shows.
					Message: "The closing bracket must be aligned with the line containing the opening tag (expected column 3 on the next line)",
					Line:    3,
					Column:  9,
				},
			},
		},
		// Tag-aligned with multi-byte chars before opening — expected column
		// must reflect UTF-16 position of '<', NOT byte offset.
		{
			Code: `var 中文Var = <App
  foo />`,
			Tsx: true,
			Output: []string{`var 中文Var = <App
  foo
            />`},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					// `var 中文Var = <App` — '<' at UTF-16 col 13 (0-based 12).
					// "expected column 13" (UTF-16); byte-based would be col 15.
					Message: "The closing bracket must be aligned with the opening tag (expected column 13 on the next line)",
					Line:    2,
					Column:  7,
				},
			},
		},
		// ---- After-tag with whitespace-only line between opening tag and `/>` ----
		// Verifies `tag.line === closing.line` check fires correctly even when
		// the gap is filled by a whitespace-only line (no attributes, just
		// trivia). Fix collapses to single-line `<App />`.
		{
			Code: `<App

/>`,
			Tsx:    true,
			Output: []string{`<App />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					Message:   "The closing bracket must be placed after the opening tag",
				},
			},
		},
		// ---- Three-level nested element: props-aligned column lock-in ----
		// Outermost is `<App>` (no attrs), middle is `<B foo>...</B>`, innermost
		// is `<C\n      bar\n      />` (props-aligned, expected at bar's column).
		// Misalign the innermost `/>` and verify column is computed from the
		// innermost lastProp (`bar`), not from outer elements.
		{
			Code: `<App>
  <B foo>
    <C
      bar
        />
  </B>
</App>`,
			Tsx: true,
			Output: []string{`<App>
  <B foo>
    <C
      bar
      />
  </B>
</App>`},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "bracketLocation",
					// `bar` at UTF-16 col 7 (1-based) → "expected column 7".
					Message: "The closing bracket must be aligned with the last prop (expected column 7)",
					Line:    5,
					Column:  9,
				},
			},
		},
	})
}
