package jsx_newline

import (
	"slices"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Source: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/jsx-newline.js
// Documentation: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-newline.md
func TestJsxNewlineUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNewlineRule, []rule_tester.ValidTestCase{
		// Upstream valid 1
		{Code: `
        <div>
          <Button>{data.label}</Button>

          <List />

          <Button>
            <IconPreview />
            Button 2

            <span></span>
          </Button>

          {showSomething === true && <Something />}

          <Button>Button 3</Button>

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `, Tsx: true,
		},
		// Upstream valid 2
		{Code: `
        <div>
          <Button>{data.label}</Button>
          <List />
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
          {showSomething === true && <Something />}
          <Button>Button 3</Button>
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
		},
		// Upstream valid 3
		{Code: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `, Tsx: true,
		},
		// Upstream valid 4
		{Code: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
		},
		// Upstream valid 5
		{Code: `
        {/* fake-eslint-disable-next-line react/forbid-component-props */}
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          <Icon f7='gear' />
        </Button>
      `, Tsx: true,
		},
		// Upstream valid 6
		{Code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* fake-eslint-disable-next-line should also work inside a component */}
          <Icon f7='gear' />
        </Button>
      `, Tsx: true,
		},
		// Upstream valid 7
		{Code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* should work inside a component */}
          {/* and it should work when using multiple comments */}
          <Icon f7='gear' />
        </Button>
      `, Tsx: true,
		},
		// Upstream valid 8
		{Code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* this is a multiline
              block comment */}
          <Icon f7='gear' />
        </Button>
      `, Tsx: true,
		},
		// Upstream valid 9
		{Code: `
        <>
          {/* does this */}
          <Icon f7='gear' />

          {/* also work with multiple components and inside a fragment? */}
          <OneLineComponent />
        </>
      `, Tsx: true,
		},
		// Upstream valid 10
		{Code: `
        <>
          <OneLineComponent />
          <AnotherOneLineComponent prop={prop} />

          <MultilineComponent
            prop1={prop1}
            prop2={prop2}
          />

          <OneLineComponent />
        </>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
		},
		// Upstream valid 11
		{Code: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />

          {/* but this one needs one */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
		},
		// Upstream valid 12
		{Code: `
        <div>
          <Button>{data.label}</Button>
          <List />

          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>

          {showSomething === true && <Something />}
          <Button>Button 3</Button>

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}

        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
		},
		// Upstream documentation example 4
		{Code: `<div>
  <Button>{data.label}</Button>

  <List />

  <Button>
    <IconPreview />
    Button 2

    <span></span>
  </Button>

  {showSomething === true && <Something />}

  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": false}},
		},
		// Upstream documentation example 6
		{Code: `<div>
  <Button>{data.label}</Button>
  <List />
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
		},
		// Upstream documentation example 7
		{Code: `<div>
  <Button>{data.label}</Button>
  {showSomething === true && <Something />}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
		},
		// Upstream documentation example 8
		{Code: `<div>
  {showSomething === true && <Something />}
  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
		},
	}, []rule_tester.InvalidTestCase{
		// Upstream invalid 1
		{Code: `
        <div>
          <Button>{data.label}</Button>
          <List />
        </div>
      `, Tsx: true,
			Output: []string{`
        <div>
          <Button>{data.label}</Button>

          <List />
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 4, Column: 11, EndLine: 4, EndColumn: 19},
			},
		},
		// Upstream invalid 2
		{Code: `
        <div>
          <Button>{data.label}</Button>
          {showSomething === true && <Something />}
        </div>
      `, Tsx: true,
			Output: []string{`
        <div>
          <Button>{data.label}</Button>

          {showSomething === true && <Something />}
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 4, Column: 11, EndLine: 4, EndColumn: 52},
			},
		},
		// Upstream invalid 3
		{Code: `
        <div>
          {showSomething === true && <Something />}
          <Button>{data.label}</Button>
        </div>
      `, Tsx: true,
			Output: []string{`
        <div>
          {showSomething === true && <Something />}

          <Button>{data.label}</Button>
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 4, Column: 11, EndLine: 4, EndColumn: 40},
			},
		},
		// Upstream invalid 4
		{Code: `
        <div>
          {showSomething === true && <Something />}
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `, Tsx: true,
			Output: []string{`
        <div>
          {showSomething === true && <Something />}

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 4, Column: 11, EndLine: 8, EndColumn: 13},
			},
		},
		// Upstream invalid 5
		{Code: `
        <div>
          {/* This should however still not work*/}
          <Icon f7='gear' />

          <OneLineComponent />
          {/* Comments between components still need a newLine */}
          <OneLineComponent />
        </div>
      `, Tsx: true,
			Output: []string{`
        <div>
          {/* This should however still not work*/}
          <Icon f7='gear' />

          <OneLineComponent />

          {/* Comments between components still need a newLine */}
          <OneLineComponent />
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 7, Column: 11, EndLine: 7, EndColumn: 67},
			},
		},
		// Upstream invalid 6
		{Code: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />
          {/* but this one needs one */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />

          {/* but this one needs one */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 7, Column: 11, EndLine: 7, EndColumn: 41},
			},
		},
		// Upstream invalid 7
		{Code: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />
          {/* Multiline */}
          {/* Block comments */}
          {/* Stick to MultilineComponent */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />

          {/* Multiline */}
          {/* Block comments */}
          {/* Stick to MultilineComponent */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 7, Column: 11, EndLine: 7, EndColumn: 28},
			},
		},
		// Upstream invalid 8
		{Code: `
        <div>
          <div>
            <button></button>
            <button></button>
          </div>
          <div>
            <span></span>
            <span></span>
          </div>
        </div>
      `, Tsx: true,
			Output: []string{`
        <div>
          <div>
            <button></button>

            <button></button>
          </div>

          <div>
            <span></span>

            <span></span>
          </div>
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				// The Go tester preserves parent-first reporting order; ESLint sorts by location.
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 7, Column: 11, EndLine: 10, EndColumn: 17},
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 5, Column: 13, EndLine: 5, EndColumn: 30},
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 9, Column: 13, EndLine: 9, EndColumn: 26},
			},
		},
		// Upstream invalid 9
		{Code: `
        <div>
          <Button>{data.label}</Button>

          <List />
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`
        <div>
          <Button>{data.label}</Button>
          <List />
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 5, Column: 11, EndLine: 5, EndColumn: 19},
			},
		},
		// Upstream invalid 10
		{Code: `
        <div>
          <Button>{data.label}</Button>

          {showSomething === true && <Something />}
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`
        <div>
          <Button>{data.label}</Button>
          {showSomething === true && <Something />}
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 5, Column: 11, EndLine: 5, EndColumn: 52},
			},
		},
		// Upstream invalid 11
		{Code: `
        <div>
          {showSomething === true && <Something />}

          <Button>{data.label}</Button>
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`
        <div>
          {showSomething === true && <Something />}
          <Button>{data.label}</Button>
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 5, Column: 11, EndLine: 5, EndColumn: 40},
			},
		},
		// Upstream invalid 12
		{Code: `
        <div>
          {showSomething === true && <Something />}

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`
        <div>
          {showSomething === true && <Something />}
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 5, Column: 11, EndLine: 9, EndColumn: 13},
			},
		},
		// Upstream invalid 13
		{Code: `
        <div>
          <div>
            <button></button>

            <button></button>
          </div>

          <div>
            <span></span>

            <span></span>
          </div>
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`
        <div>
          <div>
            <button></button>
            <button></button>
          </div>
          <div>
            <span></span>
            <span></span>
          </div>
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				// The Go tester preserves parent-first reporting order; ESLint sorts by location.
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 9, Column: 11, EndLine: 13, EndColumn: 17},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 6, Column: 13, EndLine: 6, EndColumn: 30},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 12, Column: 13, EndLine: 12, EndColumn: 26},
			},
		},
		// Upstream invalid 14
		{Code: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `, Tsx: true,
			Output: []string{`
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 5, Column: 11, EndLine: 5, EndColumn: 45},
			},
		},
		// Upstream invalid 15
		{Code: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 6, Column: 11, EndLine: 6, EndColumn: 45},
			},
		},
		// Upstream invalid 16
		{Code: `
        <>
          <OneLineComponent />
          <AnotherOneLineComponent prop={prop} />
          <MultilineComponent
            prop1={prop1}
            prop2={prop2}
          />
          <OneLineComponent />
        </>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`
        <>
          <OneLineComponent />
          <AnotherOneLineComponent prop={prop} />

          <MultilineComponent
            prop1={prop1}
            prop2={prop2}
          />

          <OneLineComponent />
        </>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 5, Column: 11, EndLine: 8, EndColumn: 13},
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 9, Column: 11, EndLine: 9, EndColumn: 31},
			},
		},
		// Upstream invalid 17
		{Code: `
        <div>
          {showSomething === true && <Something />}
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`
        <div>
          {showSomething === true && <Something />}

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 4, Column: 11, EndLine: 8, EndColumn: 13},
			},
		},
		// Upstream invalid 18
		{Code: `
        <div>
          <div>
            <button></button>

            <button></button>
          </div>
          <div>
            <span></span>

            <span></span>
          </div>
        </div>
      `, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`
        <div>
          <div>
            <button></button>
            <button></button>
          </div>

          <div>
            <span></span>
            <span></span>
          </div>
        </div>
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				// The Go tester preserves parent-first reporting order; ESLint sorts by location.
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 8, Column: 11, EndLine: 12, EndColumn: 17},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 6, Column: 13, EndLine: 6, EndColumn: 30},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 11, Column: 13, EndLine: 11, EndColumn: 26},
			},
		},
		// Upstream invalid 19
		{Code: "\n        const frag: DocumentFragment = (\n          <Fragment>\n            <sni-sequence-editor-tool\n              name=\"forward\"\n              direction=\"forward\"\n              type=\"control\"\n              onClick={ () => this.onClickNavigate('forward') }\n            />\n            <sni-sequence-editor-tool\n              name=\"rotate\"\n              direction=\"left\"\n              type=\"control\"\n              onClick={ () => this.onClickNavigate('left') }\n            />\n    \n            <sni-sequence-editor-tool\n              name=\"rotate\"\n              direction=\"right\"\n              type=\"control\"\n              onClick={ (): void => this.onClickNavigate('right') }\n            />\n    \n            <div className=\"sni-sequence-editor-control-panel__delete\" data-name=\"delete\" onClick={ this.onDeleteCommand } />\n    \n            {\n              ...Array.from(this.children)\n            }\n          </Fragment>\n        )\n      ", Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			// Upstream offers an unchanged fix after the first pass.
			Output: slices.Repeat([]string{"\n        const frag: DocumentFragment = (\n          <Fragment>\n            <sni-sequence-editor-tool\n              name=\"forward\"\n              direction=\"forward\"\n              type=\"control\"\n              onClick={ () => this.onClickNavigate('forward') }\n            />\n\n            <sni-sequence-editor-tool\n              name=\"rotate\"\n              direction=\"left\"\n              type=\"control\"\n              onClick={ () => this.onClickNavigate('left') }\n            />\n    \n            <sni-sequence-editor-tool\n              name=\"rotate\"\n              direction=\"right\"\n              type=\"control\"\n              onClick={ (): void => this.onClickNavigate('right') }\n            />\n    \n            <div className=\"sni-sequence-editor-control-panel__delete\" data-name=\"delete\" onClick={ this.onDeleteCommand } />\n    \n            {\n              ...Array.from(this.children)\n            }\n          </Fragment>\n        )\n      "}, 10),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 10, Column: 13, EndLine: 15, EndColumn: 15},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 26, Column: 13, EndLine: 28, EndColumn: 14},
			},
		},
		// Upstream documentation example 1
		{Code: `<div>
  <Button>{data.label}</Button>
  <List />
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": false}},
			Output: []string{`<div>
  <Button>{data.label}</Button>

  <List />
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 3, EndLine: 3, EndColumn: 11},
			},
		},
		// Upstream documentation example 2
		{Code: `<div>
  <Button>{data.label}</Button>
  {showSomething === true && <Something />}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": false}},
			Output: []string{`<div>
  <Button>{data.label}</Button>

  {showSomething === true && <Something />}
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 3, EndLine: 3, EndColumn: 44},
			},
		},
		// Upstream documentation example 3
		{Code: `<div>
  {showSomething === true && <Something />}
  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": false}},
			Output: []string{`<div>
  {showSomething === true && <Something />}

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "require", Message: "JSX element should start in a new line", Line: 3, Column: 3, EndLine: 7, EndColumn: 5},
			},
		},
		// Upstream documentation example 5
		{Code: `<div>
  <Button>{data.label}</Button>

  <List />

  <Button>
    <IconPreview />
    Button 2

    <span></span>
  </Button>

  {showSomething === true && <Something />}

  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true}},
			Output: []string{`<div>
  <Button>{data.label}</Button>
  <List />
  <Button>
    <IconPreview />
    Button 2
    <span></span>
  </Button>
  {showSomething === true && <Something />}
  <Button>Button 3</Button>
  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				// The Go tester preserves parent-first reporting order; ESLint sorts by location.
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 4, Column: 3, EndLine: 4, EndColumn: 11},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 6, Column: 3, EndLine: 11, EndColumn: 12},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 13, Column: 3, EndLine: 13, EndColumn: 44},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 15, Column: 3, EndLine: 15, EndColumn: 28},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 17, Column: 3, EndLine: 21, EndColumn: 5},
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 10, Column: 5, EndLine: 10, EndColumn: 18},
			},
		},
		// Upstream documentation example 9
		{Code: `<div>
  {showSomething === true && <Something />}

  <Button>Button 3</Button>
  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`<div>
  {showSomething === true && <Something />}
  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 4, Column: 3, EndLine: 4, EndColumn: 28},
				{MessageId: "allowMultilines", Message: "Multiline JSX elements should start in a new line", Line: 5, Column: 3, EndLine: 9, EndColumn: 5},
			},
		},
		// Upstream documentation example 10
		// Labeled correct upstream, but v7.37.5 reports prevent here.
		{Code: `<div>
  {showSomething === true && <Something />}

  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`, Tsx: true,
			Options: []any{map[string]any{"prevent": true, "allowMultilines": true}},
			Output: []string{`<div>
  {showSomething === true && <Something />}
  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "prevent", Message: "JSX element should not start in a new line", Line: 4, Column: 3, EndLine: 4, EndColumn: 28},
			},
		},
	})
}
