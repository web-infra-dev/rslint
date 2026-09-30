// TestRequireDefaultPropsExtras covers upstream branches, AST boundaries,
// and real-user issue shapes beyond the upstream suite.
// N/A: edit boundaries; this rule has no fixes or suggestions.
package require_default_props

import (
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestRequireDefaultPropsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireDefaultPropsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: default-parenthesized ----
		{Code: "function C(props) { return <div>{props.x}</div> } (C).propTypes = ({x:P.string}); ((C)).defaultProps = (({x:0}));", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: quoted ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={\"x\":P.string}; C.defaultProps={\"x\":0};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: numeric ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={0:P.string}; C.defaultProps={0:0};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: computed-name ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={[x]:P.string}; C.defaultProps={[x]:0};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: computed-string ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={[\"x\"]:P.string}; C.defaultProps={[\"x\"]:0};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: computed-member ----
		{Code: "function C(props) { return <div>{props.x}</div> } C[\"propTypes\"]={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: required-identifier ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string[isRequired]};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: non-null-validator ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string!.isRequired};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: asserted-validator ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:(P.string as any).isRequired};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: satisfies-object ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes=({x:P.string} satisfies T);", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: spread-defaults ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string}; C.defaultProps={...unknown};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: unresolved-sticky ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string}; C.defaultProps=unknown; C.defaultProps={};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: default-alias ----
		{Code: "function C(props) { return <div>{props.x}</div> } const d={x:0}; C.propTypes={x:P.string}; C.defaultProps=d;", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: cyclic-default ----
		{Code: "function C(props) { return <div>{props.x}</div> } const a=b,b=a; C.propTypes={x:P.string}; C.defaultProps=a;", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: cyclic-props ----
		{Code: "function C(props) { return <div>{props.x}</div> } const a=b,b=a; C.propTypes=a;", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: repeated-props ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string}; C.propTypes={x:P.string.isRequired};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: property-required ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={}; C.propTypes.x=P.string.isRequired;", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: defaults-before ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.defaultProps={x:0}; C.propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: get-default-props ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string}; C.getDefaultProps={x:0};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: typed-wrapper ----
		{Code: "function C(props) { return <div>{props.x}</div> } (C as any).propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: unresolved-props ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes=external;", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Locks in defaultArguments: props = {} ----
		{Code: "function C(props = {}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in defaultArguments: {x = 1} ----
		{Code: "function C({x = 1}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in defaultArguments: {"x": local} ----
		{Code: "function C({\"x\": local}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in defaultArguments: {...rest} ----
		{Code: "function C({...rest}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in defaultArguments: {} ----
		{Code: "function C({}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in defaultArguments: [x] ----
		{Code: "function C([x]) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in defaultArguments: ...props ----
		{Code: "function C(...props) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in no-parameter ----
		{Code: "function C() { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Locks in option-precedence ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string}; C.defaultProps={};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments", "ignoreFunctionalComponents": true}}, Settings: map[string]any{}},
		// ---- Locks in functions-ignore ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "ignore"}}, Settings: map[string]any{}},
		// ---- Dimension 4: class static getDefaultProps={x:0}; ----
		{Code: "class C extends React.Component { static propTypes={x:P.string}; static getDefaultProps={x:0}; render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: class expression ----
		{Code: "const C=class extends React.Component {static propTypes={x:P.string}}; C.defaultProps={x:0};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: async-generator ----
		{Code: "async function* C(props){yield 1;return <div/>} C.propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: typed-union ----
		{Code: "function C(props:{x?:string}|{y?:string}){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Real-user: issue 2856 forwardRef (intentional correction: honor functions option) ----
		{Code: "import React from 'react'; const C=React.forwardRef((props,ref)=><div ref={ref}/>); C.propTypes={text:P.string};", Tsx: true, Options: []any{map[string]any{"ignoreFunctionalComponents": true}}, Settings: map[string]any{}},
		// ---- Locks in configured-prop-wrapper ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes=wrap({x:P.string}); C.defaultProps=wrap({x:0});", Tsx: true, Options: []any{}, Settings: map[string]any{"propWrapperFunctions": []any{"wrap"}}},
		// ---- Dimension 4: getter switch ----
		{Code: "class C extends React.Component {static propTypes={x:P.string}; static get defaultProps(){switch(a){default:return {x:0}}} render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: quoted class fields ----
		{Code: "class C extends React.Component {static \"propTypes\"={x:P.string};render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Dimension 4: nested wrapper ----
		{Code: "const C=React.memo(React.forwardRef(({x=1}, ref)=><div/>)); C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: parenthesized ----
		{Code: "function C(props) { return <div>{props.x}</div> } (C).propTypes = (({x: (P.string)}));", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 70, EndLine: 1, EndColumn: 83}}},
		// ---- Dimension 4: computed-identifier-member ----
		{Code: "function C(props) { return <div>{props.x}</div> } C[propTypes]={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 65, EndLine: 1, EndColumn: 75}}},
		// ---- Dimension 4: required-computed ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string[\"isRequired\"]};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 88}}},
		// ---- Dimension 4: optional-validator ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P?.string.isRequired};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 86}}},
		// ---- Dimension 4: spread-props ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={...unknown,x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 75, EndLine: 1, EndColumn: 85}}},
		// ---- Dimension 4: alias-required ----
		{Code: "function C(props) { return <div>{props.x}</div> } const required=P.string.isRequired; C.propTypes={x:required};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 100, EndLine: 1, EndColumn: 110}}},
		// ---- Dimension 4: prototype-key (intentional correction: explicit defaults only) ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={toString:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"toString\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 81}}},
		// ---- Dimension 4: empty-name ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={\"\":P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 75}}},
		// ---- Dimension 4: escaped-default ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string}; C.defaultProps={\"\\x78\":0};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 74}}},
		// ---- Dimension 4: nested-prop-write ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.shape({a:P.string})}; C.propTypes.x.a=P.string;", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 87}}},
		// ---- Dimension 4: object-method ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x(){}};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 69}}},
		// ---- Dimension 4: object-getter ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={get x(){return P.string}};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 88}}},
		// ---- Locks in defaultArguments: props ----
		{Code: "function C(props) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "destructureInSignature", Message: "Must destructure props in the function signature to initialize an optional prop.", Line: 1, Column: 12, EndLine: 1, EndColumn: 17}}},
		// ---- Locks in defaultArguments: {x} ----
		{Code: "function C({x}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 13, EndLine: 1, EndColumn: 14}}},
		// ---- Locks in defaultArguments: {x: local} ----
		{Code: "function C({x: local}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}}},
		// ---- Locks in defaultArguments: {[x]: local} ----
		{Code: "function C({[x]: local}) { return <div/> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23}}},
		// ---- Locks in no-defaults-with-function ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={}; C.defaultProps={};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noDefaultPropsWithFunction", Message: "Don’t use defaultProps with function components.", Line: 1, Column: 1, EndLine: 1, EndColumn: 50}}},
		// ---- Locks in explicit-false ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"ignoreFunctionalComponents": false, "forbidDefaultForRequired": false}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 74}}},
		// ---- Locks in empty-options ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 74}}},
		// ---- Locks in required-default-independent ----
		{Code: "function C({x=1}){return <div/>} C.propTypes={x:P.string.isRequired};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments", "forbidDefaultForRequired": false}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noDefaultWithRequired", Message: "propType \"x\" is required and should not have a defaultProps declaration.", Line: 1, Column: 13, EndLine: 1, EndColumn: 16}}},
		// ---- Dimension 4: class static defaultProps=unknown; ----
		{Code: "class C extends React.Component { static propTypes={x:P.string}; static defaultProps=unknown; render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}}},
		// ---- Dimension 4: class static get defaultProps(){return unknown} ----
		{Code: "class C extends React.Component { static propTypes={x:P.string}; static get defaultProps(){return unknown} render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}}},
		// ---- Dimension 4: class static get defaultProps(){} ----
		{Code: "class C extends React.Component { static propTypes={x:P.string}; static get defaultProps(){} render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}}},
		// ---- Dimension 4: class static defaultProps(){return {x:0}} ----
		{Code: "class C extends React.Component { static propTypes={x:P.string}; static defaultProps(){return {x:0}} render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}}},
		// ---- Dimension 4: class defaultProps={x:0}; ----
		{Code: "class C extends React.Component { static propTypes={x:P.string}; defaultProps={x:0}; render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}}},
		// ---- Dimension 4: class #defaultProps={x:0}; ----
		{Code: "class C extends React.Component { static propTypes={x:P.string}; #defaultProps={x:0}; render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}}},
		// ---- Dimension 4: abstract bodyless ----
		{Code: "abstract class C extends React.Component {static propTypes={x:P.string}; abstract render(): any;}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 61, EndLine: 1, EndColumn: 71}}},
		// ---- Dimension 4: generator ----
		{Code: "function* C(props){yield 1;return <div/>} C.propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 56, EndLine: 1, EndColumn: 66}}},
		// ---- Dimension 4: named-expression ----
		{Code: "const C=function Named(props){return <div/>}; C.propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 60, EndLine: 1, EndColumn: 70}}},
		// ---- Dimension 4: nested-bindings ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string}; function outer(){function C(){return <div/>} C.propTypes={y:P.string}; C.defaultProps={y:0};}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 74}}},
		// ---- Dimension 4: object-component ----
		{Code: "const ns={C(props){return <div/>}}; ns.C.propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 63}}},
		// ---- Dimension 4: typed-optional ----
		{Code: "type Props={x?:string}; function C(props:Props){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}}},
		// ---- Dimension 4: typed-parameter-range ----
		{Code: "function C(props: {x?:string}){return <div/>}", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "destructureInSignature", Message: "Must destructure props in the function signature to initialize an optional prop.", Line: 1, Column: 12, EndLine: 1, EndColumn: 30}}},
		// ---- Dimension 4: typed-intersection ----
		{Code: "type P={x?:string}; function C(props:P & {y?:number}){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "shouldHaveDefault", Message: "propType \"y\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 43, EndLine: 1, EndColumn: 52}}},
		// ---- Dimension 4: recursive-type ----
		{Code: "interface P extends P {x?:string} function C(props:P){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 24, EndLine: 1, EndColumn: 33}}},
		// ---- Dimension 4: interface-extends ----
		{Code: "interface A{x?:string} interface B extends A {y?:number} function C(props:B){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}, {MessageId: "shouldHaveDefault", Message: "propType \"y\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 47, EndLine: 1, EndColumn: 56}}},
		// ---- Dimension 4: ReturnType-function ----
		{Code: "function C(props:ReturnType<()=>{x?:string}>){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 34, EndLine: 1, EndColumn: 43}}},
		// ---- Dimension 4: ReturnType-value ----
		{Code: "const f=()=>({x:1}); function C(props:ReturnType<typeof f>){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 15, EndLine: 1, EndColumn: 18}}},
		// ---- Dimension 4: React-children ----
		{Code: "import React from 'react'; function C(props:React.PropsWithChildren<{}>){return <div/>}", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "destructureInSignature", Message: "Must destructure props in the function signature to initialize an optional prop.", Line: 1, Column: 39, EndLine: 1, EndColumn: 72}}},
		// ---- Dimension 4: unicode-crlf ----
		{Code: "// 😀\r\nfunction C(props){return <div/>}\r\nC.propTypes={\"名字\":P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"名字\" is not required, but has no corresponding defaultProps declaration.", Line: 3, Column: 14, EndLine: 3, EndColumn: 27}}},
		// ---- Real-user: issue 3660 required and nullable defaults ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={x:P.string,y:P.string.isRequired}; C.defaultProps={x:null,y:1};", Tsx: true, Options: []any{map[string]any{"forbidDefaultForRequired": true}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noDefaultWithRequired", Message: "propType \"y\" is required and should not have a defaultProps declaration.", Line: 1, Column: 75, EndLine: 1, EndColumn: 96}}},
		// ---- Locks in configured-component-wrapper ----
		{Code: "const C=observer(props=><div/>); C.propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{"componentWrapperFunctions": []any{"observer"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 47, EndLine: 1, EndColumn: 57}}},
		// ---- Dimension 4: quoted defaults field ----
		{Code: "class C extends React.Component {static propTypes={x:P.string};static \"defaultProps\"={x:1};render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 52, EndLine: 1, EndColumn: 62}}},
	})
}
