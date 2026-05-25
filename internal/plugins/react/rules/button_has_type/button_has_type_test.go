package button_has_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestButtonHasTypeRule(t *testing.T) {
	resetFalse := map[string]interface{}{"reset": false}
	pragmaFoo := map[string]interface{}{
		"react": map[string]interface{}{"pragma": "Foo"},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ButtonHasTypeRule, []rule_tester.ValidTestCase{
		// ---- JSX ----
		{Code: `var x = <span/>`, Tsx: true},
		{Code: `var x = <span type="foo"/>`, Tsx: true},
		{Code: `var x = <button type="button"/>`, Tsx: true},
		{Code: `var x = <button type="submit"/>`, Tsx: true},
		{Code: `var x = <button type="reset"/>`, Tsx: true},
		{Code: `var x = <button type={"button"}/>`, Tsx: true},
		{Code: `var x = <button type={'button'}/>`, Tsx: true},
		{Code: "var x = <button type={`button`}/>", Tsx: true},
		{Code: `var x = <button type={condition ? "button" : "submit"}/>`, Tsx: true},
		{Code: `var x = <button type={condition ? 'button' : 'submit'}/>`, Tsx: true},
		{Code: "var x = <button type={condition ? `button` : `submit`}/>", Tsx: true},
		{
			Code:    `var x = <button type="button"/>`,
			Options: resetFalse,
			Tsx:     true,
		},
		// Paren-wrapped expressions (rslint edge case)
		{Code: `var x = <button type={("button")}/>`, Tsx: true},
		{Code: `var x = <button type={(condition ? "button" : "submit")}/>`, Tsx: true},
		{Code: `var x = <button type={((("button")))}/>`, Tsx: true},
		// Nested ternary
		{Code: `var x = <button type={a ? "button" : b ? "submit" : "reset"}/>`, Tsx: true},
		// Button with children
		{Code: `var x = <button type="button">Click</button>`, Tsx: true},
		{Code: `var x = <button type="button"><span>hi</span></button>`, Tsx: true},
		// Spread after explicit type — type wins
		{Code: `var x = <button type="button" {...props}/>`, Tsx: true},
		{Code: `var x = <button {...props} type="button"/>`, Tsx: true},
		// Not a button tag
		{Code: `var x = <Button/>`, Tsx: true},
		{Code: `var x = <Button.Primary/>`, Tsx: true},
		{Code: `var x = <div/>`, Tsx: true},
		{Code: `var x = <input/>`, Tsx: true},
		{Code: `var x = <button-like type="foo"/>`, Tsx: true},
		// Namespaced attr "ns:type" is not "type"
		{Code: `var x = <button ns:type="foo" type="button"/>`, Tsx: true},
		// Fragment / array of buttons
		{Code: `var x = <>{[<button key="a" type="button"/>, <button key="b" type="submit"/>]}</>`, Tsx: true},

		// ---- React.createElement ----
		{Code: `React.createElement("span")`, Tsx: true},
		{Code: `React.createElement("span", {type: "foo"})`, Tsx: true},
		{Code: `React.createElement("button", {type: "button"})`, Tsx: true},
		{Code: `React.createElement("button", {type: 'button'})`, Tsx: true},
		{Code: "React.createElement(\"button\", {type: `button`})", Tsx: true},
		{Code: `React.createElement("button", {type: "submit"})`, Tsx: true},
		{Code: `React.createElement("button", {type: 'submit'})`, Tsx: true},
		{Code: "React.createElement(\"button\", {type: `submit`})", Tsx: true},
		{Code: `React.createElement("button", {type: "reset"})`, Tsx: true},
		{Code: `React.createElement("button", {type: 'reset'})`, Tsx: true},
		{Code: "React.createElement(\"button\", {type: `reset`})", Tsx: true},
		{Code: `React.createElement("button", {type: condition ? "button" : "submit"})`, Tsx: true},
		{Code: `React.createElement("button", {type: condition ? 'button' : 'submit'})`, Tsx: true},
		{Code: "React.createElement(\"button\", {type: condition ? `button` : `submit`})", Tsx: true},
		{
			Code:    `React.createElement("button", {type: "button"})`,
			Options: resetFalse,
			Tsx:     true,
		},
		// Non-React createElement
		{Code: `document.createElement("button")`, Tsx: true},
		// createElement where first arg isn't "button"
		{Code: `React.createElement("div", {})`, Tsx: true},
		{Code: `React.createElement(Component, {type: "foo"})`, Tsx: true},
		// createElement first arg is template literal without substitutions — not a StringLiteral,
		// so rslint does not treat it as "button" (matches ESLint, which requires Literal with string).
		{Code: "React.createElement(`button`, {type: \"button\"})", Tsx: true},
		// Empty args
		{Code: `React.createElement()`, Tsx: true},
		// Spread type prop — type found statically
		{Code: `React.createElement("button", {...extraProps, type: "button"})`, Tsx: true},
		// Paren-wrapped callee / args — ESTree-flattening parity
		{Code: `(React).createElement("button", {type: "button"})`, Tsx: true},
		{Code: `(React.createElement)("button", {type: "button"})`, Tsx: true},
		{Code: `React.createElement(("button"), {type: "button"})`, Tsx: true},
		{Code: `React.createElement("button", ({type: "button"}))`, Tsx: true},
		{Code: `((React).createElement)(("button"), ({type: "button"}))`, Tsx: true},
		// Foo.createElement without pragma — not recognized, so no report.
		{Code: `Foo.createElement("span")`, Tsx: true},
		{Code: `Foo.createElement("button")`, Tsx: true},
		// With pragma: "Foo" — first arg isn't "button" so nothing to check.
		{
			Code:     `Foo.createElement("span")`,
			Settings: pragmaFoo,
			Tsx:      true,
		},
		// With pragma: "Foo" — Foo.createElement("button", …) with valid type.
		{
			Code:     `Foo.createElement("button", {type: "button"})`,
			Settings: pragmaFoo,
			Tsx:      true,
		},
		// With pragma: "Foo" — default React.createElement(…) is no longer recognized.
		{
			Code:     `React.createElement("button")`,
			Settings: pragmaFoo,
			Tsx:      true,
		},
		// Spread at JSX side: we don't inspect spread values
		{Code: `var x = <button type="button" {...props}/>`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- JSX: missing (self-closing, position with End*) ----
		{
			Code: `var x = <button/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingType",
					Line:      1, Column: 9,
					EndLine: 1, EndColumn: 18,
				},
			},
		},
		// ---- JSX: missing with children (JsxElement span with End*) ----
		{
			Code: `var x = <button>Click</button>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingType",
					Line:      1, Column: 9,
					EndLine: 1, EndColumn: 31,
				},
			},
		},
		// ---- JSX: missing, multi-line with children ----
		{
			Code: "var x = (\n\t\t\t\t<button>\n\t\t\t\t\tClick\n\t\t\t\t</button>\n\t\t\t)",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingType",
					Line:      2, Column: 5,
					EndLine: 4, EndColumn: 14,
				},
			},
		},
		// ---- JSX: invalid literal value ----
		{
			Code: `var x = <button type="foo"/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"foo" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		// ---- JSX: complex identifier expression ----
		{
			Code: `var x = <button type={foo}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "complexType",
					Message:   "The button type attribute must be specified by a static string or a trivial ternary expression",
					Line:      1, Column: 23,
				},
			},
		},
		{
			Code: `var x = <button type={"foo"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		{
			Code: `var x = <button type={'foo'}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		{
			Code: "var x = <button type={`foo`}/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		// ---- JSX: template with substitution ----
		{
			Code: "var x = <button type={`button${foo}`}/>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		// ---- JSX: forbidden value ----
		{
			Code:    `var x = <button type="reset"/>`,
			Options: resetFalse,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "forbiddenValue",
					Message:   `"reset" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		// ---- JSX: conditional expression ----
		{
			Code: `var x = <button type={condition ? "button" : foo}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 46},
			},
		},
		{
			Code: `var x = <button type={condition ? "button" : "foo"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		{
			Code:    `var x = <button type={condition ? "button" : "reset"}/>`,
			Options: resetFalse,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenValue", Line: 1, Column: 9},
			},
		},
		{
			Code: `var x = <button type={condition ? foo : "button"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 35},
			},
		},
		{
			Code: `var x = <button type={condition ? "foo" : "button"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		// ---- JSX: valueless attribute ----
		{
			Code: `var x = <button type/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"true" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		{
			Code:    `var x = <button type={condition ? "reset" : "button"}/>`,
			Options: resetFalse,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenValue", Line: 1, Column: 9},
			},
		},
		// ---- React.createElement (positions with End*) ----
		{
			Code: `React.createElement("button")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingType",
					Line:      1, Column: 1,
					EndLine: 1, EndColumn: 30,
				},
			},
		},
		{
			Code: `React.createElement("button", {type: foo})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "complexType",
					Line:      1, Column: 38,
					EndLine: 1, EndColumn: 41,
				},
			},
		},
		{
			Code: `React.createElement("button", {type: "foo"})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"foo" is an invalid value for button type attribute`,
					Line:      1, Column: 1,
				},
			},
		},
		{
			Code:    `React.createElement("button", {type: "reset"})`,
			Options: resetFalse,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenValue", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("button", {type: condition ? "button" : foo})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 61},
			},
		},
		{
			Code: `React.createElement("button", {type: condition ? "button" : "foo"})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 1},
			},
		},
		{
			Code:    `React.createElement("button", {type: condition ? "button" : "reset"})`,
			Options: resetFalse,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenValue", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("button", {type: condition ? foo : "button"})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 50},
			},
		},
		{
			Code: `React.createElement("button", {type: condition ? "foo" : "button"})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 1},
			},
		},
		{
			Code:    `React.createElement("button", {type: condition ? "reset" : "button"})`,
			Options: resetFalse,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "forbiddenValue", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("button", {...extraProps})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		// With pragma: "Foo" — Foo.createElement("button") reports missingType (upstream parity).
		{
			Code:     `Foo.createElement("button")`,
			Settings: pragmaFoo,
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		{
			Code:     `Foo.createElement("button", {type: "foo"})`,
			Settings: pragmaFoo,
			Tsx:      true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"foo" is an invalid value for button type attribute`,
					Line:      1, Column: 1,
				},
			},
		},
		{
			Code: `function Button({ type, ...extraProps }) { const button = type; return <button type={button} {...extraProps} />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 86},
			},
		},
		// ---- Multi-line ----
		{
			Code: `var x = (
				<button
					type={foo}
				/>
			)`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 3, Column: 12},
			},
		},
		// ---- Numeric / boolean / null literal values (not string type values) ----
		{
			Code: `var x = <button type={0}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"0" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		{
			// Numeric normalization: ESLint's String(0x1) === "1"
			Code: `var x = <button type={0x1}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"1" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		{
			Code: `var x = <button type={true}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"true" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		{
			Code: `var x = <button type={null}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		// ---- JSX spread — no explicit type ----
		{
			Code: `var x = <button {...props}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 9},
			},
		},
		{
			Code: `var x = <button {...props} type/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"true" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		// ---- LogicalOr fallback — complex ----
		{
			Code: `var x = <button type={foo || "button"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		// ---- Nested ternary, inner alternate invalid ----
		{
			Code: `var x = <button type={a ? "button" : b ? "submit" : "foo"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"foo" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		// ---- Multiple reports: ternary with both sides invalid ----
		{
			Code: `var x = <button type={a ? "foo" : "bar"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		// ---- Duplicate type attribute: first wins (rslint matches ESLint's jsx-ast-utils getProp) ----
		{
			Code: `var x = <button type="foo" type="button"/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidValue", Line: 1, Column: 9},
			},
		},
		// ---- Shorthand object property ----
		{
			Code: `React.createElement("button", {type})`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 32},
			},
		},
		// ---- createElement with template literal substitution ----
		{
			Code: "React.createElement(\"button\", {type: `button${foo}`})",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 38},
			},
		},
		// ---- Complex expression types → complexType ----
		{
			Code: `var x = <button type={getType()}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		{
			Code: `var x = <button type={obj.type}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		{
			Code: `var x = <button type={arr[0]}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		// ---- TS assertions / non-null / satisfies → complexType (matches ESLint with @typescript-eslint-parser) ----
		{
			Code: `var x = <button type={foo as "button"}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		{
			Code: `var x = <button type={foo!}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		{
			Code: `var x = <button type={"button" satisfies string}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "complexType", Line: 1, Column: 23},
			},
		},
		// ---- BigInt literal value — normalized to decimal ----
		{
			Code: `var x = <button type={1n}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"1" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		// ---- Regex literal value ----
		{
			Code: `var x = <button type={/foo/}/>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"/foo/" is an invalid value for button type attribute`,
					Line:      1, Column: 9,
				},
			},
		},
		// ---- createElement with non-object second argument → missingType ----
		{
			Code: `React.createElement("button", null)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("button", "foo")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("button", undefined)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		// ---- Paren-wrapped callee / args — ESTree-flattening parity ----
		{
			Code: `(React).createElement("button")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		{
			Code: `(React.createElement)("button")`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement(("button"))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingType", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("button", ({type: "foo"}))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidValue",
					Message:   `"foo" is an invalid value for button type attribute`,
					Line:      1, Column: 1,
				},
			},
		},
	})
}
