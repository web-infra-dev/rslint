package state_in_constructor

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestStateInConstructorExtras(t *testing.T) {
	root := fixtures.GetRootDir()
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig-js.json"): `{"extends":"./tsconfig.json","compilerOptions":{"allowJs":true}}`,
	})
	rule_tester.RunRuleTester(root, "tsconfig-js.json", t, &StateInConstructorRule, []rule_tester.ValidTestCase{
		// Unrelated classes and aliased imports are not recognized by name.
		{
			Code: `import { Component as Base } from 'react'; class Plain { state = {}; } const Helper = class extends Other { state = {}; }; class Aliased extends Base { state = {}; }`,
			Tsx:  true,
		},
		// Static fields, literal keys, methods, and accessors are not instance state fields.
		{
			Code: `class C extends React.Component { static state = {}; static #state = {}; 'state' = {}; ['state'] = {}; [state + ''] = {}; state() {} get state() { return {}; } set state(value) {} accessor state = {}; }`,
			Tsx:  true,
		},
		// A non-component class blocks the enclosing component.
		{
			Code: `class Outer extends React.Component { render() { return class Inner { state = {}; }; } }`,
			Tsx:  true,
		},
		// Computed string, optional, and asserted base expressions are not recognized.
		{
			Code: `class A extends React['Component'] { state = {}; } class B extends (React?.Component) { state = {}; } class C extends (React.Component as any) { state = {}; }`,
			Tsx:  true,
		},
		// Type signatures and constructor parameter properties are not class fields.
		{
			Code: `interface Props { state: {}; } type T = { state: {} }; class C extends React.Component { constructor(public state: {}) { super(); } }`,
			Tsx:  true,
		},
		// Braced JSDoc does not explicitly mark a component.
		{
			Code: `/** @extends {React.Component} */
class C extends Base { state = {}; }`,
			FileName: `component.js`,
		},
		// Configured pragma replaces the default React namespace.
		{
			Code:     `class C extends React.Component { state = {}; }`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
		},
		// No state initialization is required, and never allows fields and methods.
		{
			Code:    `class C extends Component { state = {}; reset() { this.state = {}; } }`,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Never excludes literal keys, nested properties, other receivers, updates, and patterns.
		{
			Code:    `class C extends Component { constructor() { super(); this['state'] = {}; this[state + ''] = {}; this.state.value = 1; other.state = {}; this.state++; ++this.state; ({ state: this.state } = other); [this.state] = other; this.state === other; } }`,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Never does not apply to constructors outside class components.
		{
			Code:    `class Plain { constructor() { this.state = {}; } } function constructor() { this.state = {}; } const obj = { constructor() { this.state = {}; } };`,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Static and computed methods called constructor are not constructors.
		{
			Code:    `class C extends React.Component { static constructor() { this.state = {}; } [constructor]() { this.state = {}; } }`,
			Tsx:     true,
			Options: []any{"never"},
		},
		// A non-component nested class blocks constructor assignments.
		{
			Code:    `class Outer extends React.Component { constructor() { super(); class Inner { constructor() { this.state = {}; } } } }`,
			Tsx:     true,
			Options: []any{"never"},
		},
		// Authored TypeScript receiver and target wrappers remain significant.
		{
			Code:    `class C extends React.Component { constructor() { super(); (this as C).state = {}; this!.state = {}; (this.state as any) = {}; } }`,
			Tsx:     true,
			Options: []any{"never"},
		},
		// genericHeritage; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class C extends (React.Component<{}, {}>) { state = {}; }`,
			FileName: "case.tsx",
			Options:  []any{"always"},
		},
	}, []rule_tester.InvalidTestCase{
		// Always explicitly reports each identifier-shaped instance state field.
		{
			Code:    `class C extends Component { state; [state] = {}; [(state)] = {}; #state = {}; }`,
			Tsx:     true,
			Options: []any{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 50, EndLine: 1, EndColumn: 65},
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 66, EndLine: 1, EndColumn: 78},
			},
		},
		// PureComponent, computed identifier heritage, and class expressions are recognized.
		{
			Code: `class A extends PureComponent { state = {}; } class B extends React[Component] { state = {}; } const C = class extends ((React).PureComponent) { state = {}; };`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 33, EndLine: 1, EndColumn: 44},
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 82, EndLine: 1, EndColumn: 93},
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 146, EndLine: 1, EndColumn: 157},
			},
		},
		// A real nested component is recognized independently.
		{
			Code: `class Outer { create() { return class Inner extends React.Component { state = {}; }; } }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 71, EndLine: 1, EndColumn: 82},
			},
		},
		// Default exports and TypeScript fields keep their complete declaration ranges.
		{
			Code: `export default class C extends React.Component<{}, {}> { @decorate public state!: {}; declare other: {}; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 58, EndLine: 1, EndColumn: 86},
			},
		},
		// Declared fields are checked, but abstract fields use a different ESTree node.
		{
			Code: `abstract class C extends React.Component { abstract state: {}; } class D extends React.Component { declare state: {}; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 100, EndLine: 1, EndColumn: 118},
			},
		},
		// JSDoc explicitly marks otherwise unrelated classes as components.
		{
			Code: `/** @extends React.Component */
class C extends Base { state = {}; }`,
			FileName: `component.js`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 2, Column: 24, EndLine: 2, EndColumn: 35},
			},
		},
		// JSDoc also attaches to a class expression declaration.
		{
			Code: `/** @augments React.PureComponent */
const C = class extends Base { state = {}; };`,
			FileName: `component.js`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 2, Column: 32, EndLine: 2, EndColumn: 43},
			},
		},
		// Configured pragma supports component fields.
		{
			Code:     `class C extends Preact.Component { state = {}; }`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 36, EndLine: 1, EndColumn: 47},
			},
		},
		// The JSX annotation overrides the configured pragma.
		{
			Code: `/** @jsx Custom.h */
class C extends Custom.Component { state = {}; }`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"pragma": "Preact"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 2, Column: 36, EndLine: 2, EndColumn: 47},
			},
		},
		// Multiline fields preserve UTF-16 columns, comments, and semicolons.
		{
			Code: `class C extends React.Component { /* 😀 */ state = {
  label: '你好'
};
}`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 44, EndLine: 3, EndColumn: 3},
			},
		},
		// Never includes compound, logical, computed identifier, and private assignments.
		{
			Code:    `class C extends Component { #state; constructor() { super(); this.state += 1; this.state ??= {}; this[state] = {}; this.#state = {}; } }`,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 62, EndLine: 1, EndColumn: 77},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 79, EndLine: 1, EndColumn: 96},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 98, EndLine: 1, EndColumn: 114},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 116, EndLine: 1, EndColumn: 132},
			},
		},
		// Parentheses around receivers, keys, targets, and assignments stay transparent.
		{
			Code:    `class C extends React.Component { constructor() { super(); ((this).state) = {}; ((this[(state)] = {})); } }`,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 60, EndLine: 1, EndColumn: 79},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 83, EndLine: 1, EndColumn: 101},
			},
		},
		// Constructor scopes include parameter defaults and nested functions.
		{
			Code:    `class C extends React.Component { constructor(value = (this.state = {})) { super(); const arrow = () => { this.state = {}; }; function inner() { this.state = {}; } run(function () { this.state = {}; }); } }`,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 56, EndLine: 1, EndColumn: 71},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 107, EndLine: 1, EndColumn: 122},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 146, EndLine: 1, EndColumn: 161},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 183, EndLine: 1, EndColumn: 198},
			},
		},
		// Constructor search crosses inner component methods, fields, and static blocks.
		{
			Code:    `class Outer extends React.Component { constructor() { super(); class Inner extends Component { reset() { this.state = {}; } value = (this.state = {}); static { this.state = {}; } } } }`,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 106, EndLine: 1, EndColumn: 121},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 134, EndLine: 1, EndColumn: 149},
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 161, EndLine: 1, EndColumn: 176},
			},
		},
		// String-named constructors count as constructors.
		{
			Code:    `class C extends React.Component { 'constructor'() { super(); this.state = {}; } }`,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 62, EndLine: 1, EndColumn: 77},
			},
		},
		// Constructor assignment ranges include multiline values and UTF-16 columns.
		{
			Code: `class C extends React.Component { constructor() { super(); /* 😀 */ this.state = {
  label: '你好'
}; } }`,
			Tsx:     true,
			Options: []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 69, EndLine: 3, EndColumn: 2},
			},
		},
		// JavaScript JSDoc casts do not create runtime expression wrappers.
		{
			Code:     `class C extends React.Component { constructor() { super(); (/** @type {C} */ (this)).state = {}; } }`,
			FileName: `component.js`,
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 60, EndLine: 1, EndColumn: 96},
			},
		},
		// decoratedParameter; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class C extends React.Component { constructor(@dec(this.state = {}) value) { super(); } }`,
			FileName: "case.tsx",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 52, EndLine: 1, EndColumn: 67},
			},
		},
		// computedMethodInConstructor; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class Outer extends React.Component { constructor() { super(); class C extends Component { [this.state = {}]() {} } } }`,
			FileName: "case.tsx",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 93, EndLine: 1, EndColumn: 108},
			},
		},
		// privateBase; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class React { static #Component; static make() { return class C extends React.#Component { state = {}; }; } }`,
			FileName: "case.tsx",
			Options:  []any{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 92, EndLine: 1, EndColumn: 103},
			},
		},
		// jsdocBase; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class C extends (/** @type {any} */ (React.Component)) { state = {}; }`,
			FileName: "case.js",
			Options:  []any{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 58, EndLine: 1, EndColumn: 69},
			},
		},
		// jsdocComputedKey; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class C extends Component { [/** @type {string} */ (state)] = {}; }`,
			FileName: "case.js",
			Options:  []any{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 29, EndLine: 1, EndColumn: 66},
			},
		},
		// jsdocComputedAccess; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class C extends Component { constructor() { super(); this[/** @type {string} */ (state)] = {}; } }`,
			FileName: "case.js",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 54, EndLine: 1, EndColumn: 94},
			},
		},
		// optionalRhs; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `class C extends Component { constructor() { super(); this.state = source?.state; } }`,
			FileName: "case.tsx",
			Options:  []any{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitClassProp", Message: "State initialization should be in a class property", Line: 1, Column: 54, EndLine: 1, EndColumn: 80},
			},
		},
		// exportJsdoc; checked against eslint-plugin-react v7.37.5.
		{
			Code: `/** @extends React.Component */
export default class C { state = {}; }`,
			FileName: "case.js",
			Options:  []any{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 2, Column: 26, EndLine: 2, EndColumn: 37},
			},
		},
		// declaredConstructor; checked against eslint-plugin-react v7.37.5.
		{
			Code:     `declare class C extends React.Component { constructor(value?: {}); state: {}; }`,
			FileName: "case.tsx",
			Options:  []any{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "stateInitConstructor", Message: "State initialization should be in a constructor", Line: 1, Column: 68, EndLine: 1, EndColumn: 78},
			},
		},
	})
}
