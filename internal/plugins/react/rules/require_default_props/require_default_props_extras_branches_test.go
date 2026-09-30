// TestRequireDefaultPropsExtrasBranches covers declaration ordering, type boundaries,
// and approved corrections to upstream behavior. Other edge shapes live in extras_test.go.
package require_default_props

import (
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestRequireDefaultPropsExtrasBranches(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &RequireDefaultPropsRule, []rule_tester.ValidTestCase{
		// ---- Approved correction: React.memo ignore ----
		{Code: "const C=React.memo((props)=><div/>); C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "ignore"}}, Settings: map[string]any{}},
		// ---- Approved correction: React.memo default argument ----
		{Code: "const C=React.memo(({x=1})=><div/>); C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Approved correction: React.forwardRef ignore ----
		{Code: "const C=React.forwardRef((props)=><div/>); C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "ignore"}}, Settings: map[string]any{}},
		// ---- Approved correction: React.forwardRef default argument ----
		{Code: "const C=React.forwardRef(({x=1})=><div/>); C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		// ---- Approved correction: explicit toString ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={toString:P.string}; C.defaultProps={toString:null};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Approved correction: required toString ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={toString:P.string.isRequired};", Tsx: true, Options: []any{map[string]any{"forbidDefaultForRequired": true}}, Settings: map[string]any{}},
		// ---- Approved correction: explicit constructor ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={constructor:P.string}; C.defaultProps={constructor:null};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Approved correction: required constructor ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={constructor:P.string.isRequired};", Tsx: true, Options: []any{map[string]any{"forbidDefaultForRequired": true}}, Settings: map[string]any{}},
		// ---- Approved correction: explicit hasOwnProperty ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={hasOwnProperty:P.string}; C.defaultProps={hasOwnProperty:null};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Approved correction: required hasOwnProperty ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={hasOwnProperty:P.string.isRequired};", Tsx: true, Options: []any{map[string]any{"forbidDefaultForRequired": true}}, Settings: map[string]any{}},
		// ---- Branch: declaration before class ----
		{Code: "C.propTypes={x:P.string}; class C extends React.Component{static propTypes={x:P.string.isRequired}}", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Branch: ES5 shorthand ----
		{Code: "const C=createReactClass({propTypes:{x:P.string},getDefaultProps(){return {x:1}},render(){return <div/>}});", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Branch: ES5 paren return ----
		{Code: "const C=createReactClass({propTypes:{x:P.string},getDefaultProps(){return ({x:1})},render(){return <div/>}});", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Branch: const component alias ----
		{Code: "const C=props=><div/>; const D=C; D.propTypes={x:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Branch: TS quoted defaults type ----
		{Code: "function C(props:{x?:string}){return <div/>}; C.defaultProps={x:1};", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Branch: nullable props ----
		{Code: "function C(props:{x?:string}|null){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}},
		// ---- Regression: wrapper function options apply to plain return values ----
		{Code: `const C = React.memo(({ x = "ok" }) => x); C.propTypes = { x: PropTypes.string };`, Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
		{Code: `const C = React.memo(({ x }) => x); C.propTypes = { x: PropTypes.string };`, Tsx: true, Options: []any{map[string]any{"functions": "ignore"}}, Settings: map[string]any{}},
		// ---- Regression: JSDoc this is not the authored props parameter ----
		{Code: "/** @this {void} */\nfunction C({ x = \"ok\" }) { return <div>{x}</div>; }\nC.propTypes = { x: PropTypes.string };", FileName: "review.jsx", TSConfig: "tsconfig.allow-js.json", Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}},
	}, []rule_tester.InvalidTestCase{
		// ---- Approved correction: React.memo missing argument ----
		{Code: "const C=React.memo(({x})=><div/>); C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23}}},
		// ---- Approved correction: React.forwardRef missing argument ----
		{Code: "const C=React.forwardRef(({x})=><div/>); C.propTypes={x:P.string};", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 28, EndLine: 1, EndColumn: 29}}},
		// ---- Approved correction: inherited toString ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={toString:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"toString\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 81}}},
		// ---- Approved correction: inherited constructor ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={constructor:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"constructor\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 84}}},
		// ---- Approved correction: inherited hasOwnProperty ----
		{Code: "function C(props) { return <div>{props.x}</div> } C.propTypes={hasOwnProperty:P.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"hasOwnProperty\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 64, EndLine: 1, EndColumn: 87}}},
		// ---- Branch: inline class type (approved correction) ----
		{Code: "class C extends React.Component<{x?:string}> {render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 34, EndLine: 1, EndColumn: 43}}},
		// ---- Branch: class alias ----
		{Code: "type Props={x?:string}; class C extends React.Component<Props> {render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}}},
		// ---- Branch: local type (approved correction) ----
		{Code: "function outer(){type Props={x?:string}; function C(props:Props){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 30, EndLine: 1, EndColumn: 39}}},
		// ---- Branch: interface method ----
		{Code: "interface Props{x?():void}; function C(props:Props){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 17, EndLine: 1, EndColumn: 26}}},
		// ---- Branch: typed empty key ----
		{Code: "function C(props:{\"\"?:string}){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 19, EndLine: 1, EndColumn: 29}}},
		// ---- Branch: FC type ----
		{Code: "import React from 'react'; const C: React.FC<{x?:string}> = props => <div/>;", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 47, EndLine: 1, EndColumn: 56}}},
		// ---- Branch: typed props alias ----
		{Code: "type Props={x?:string}; type Alias=Props; function C(props:Alias){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}}},
		// ---- Branch: class typed props field ----
		{Code: "class C extends React.Component {props:{x?:string};render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 41, EndLine: 1, EndColumn: 50}}},
		// ---- Branch: anonymous default export ----
		{Code: "export default function (props:{x?:string}){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 33, EndLine: 1, EndColumn: 42}}},
		// ---- Branch: declaration before function ----
		{Code: "C.propTypes={x:P.string}; function C(){return <div/>}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24}}},
		// ---- Branch: class propTypes alias ----
		{Code: "const types={x:P.string};class C extends React.Component{static propTypes=types;render(){return <div/>}}", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 14, EndLine: 1, EndColumn: 24}}},
		// ---- Branch: assignment RHS read (approved correction) ----
		{Code: "function C(){return <div/>}; C.propTypes={x:P.string}; foo=C.defaultProps;", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 43, EndLine: 1, EndColumn: 53}}},
		// ---- Regression: named function expressions use their containing property path ----
		{Code: "const ns = { C: function Named(props) { return <div />; } }; ns.C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 81, EndLine: 1, EndColumn: 100}}},
		// ---- Regression: named class expressions use their containing property path ----
		{Code: "const ns = { C: class Named extends React.Component { render() { return <div />; } } }; ns.C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 108, EndLine: 1, EndColumn: 127}}},
		// ---- Regression: repeated component assignments share one prop contract ----
		{Code: "let C = () => <div />; C = () => <span />; C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 60, EndLine: 1, EndColumn: 79}}},
		// ---- Regression: conditional component assignments retain every candidate without duplicate reports ----
		{Code: "let C; if (flag) C = () => <div />; else C = () => <span />; C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 78, EndLine: 1, EndColumn: 97}}},
		// ---- Regression: repeated assignments deduplicate defaultArguments diagnostics ----
		{Code: "let C = ({ x }) => <div />; C = ({ x }) => <span />; C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 12, EndLine: 1, EndColumn: 13}}},
		// ---- Regression: separate inline contracts on one binding remain independent ----
		{Code: "let C; if (flag) { C = (props: { x?: string }) => <div />; } else { C = (props: { x?: string }) => <span />; }", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 34, EndLine: 1, EndColumn: 44}, {MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 83, EndLine: 1, EndColumn: 93}}},
		// ---- Regression: separate inline static propTypes remain independent ----
		{Code: "let C = class extends React.Component { static propTypes = { x: PropTypes.string }; render() { return <div />; } }; C = class extends React.Component { static propTypes = { x: PropTypes.string }; render() { return <span />; } };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 62, EndLine: 1, EndColumn: 81}, {MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 174, EndLine: 1, EndColumn: 193}}},
		// ---- Regression: repeated properties in one parameter are separate diagnostics ----
		{Code: "function C({ x, x: y }) { return <div />; } C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15}, {MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21}}},
		// ---- Regression: reused TypeScript contracts stay scoped to each component ----
		{Code: "type Props = { x?: string }; function A({ x }: Props) { return <div>{x}</div>; } function B({ x }: Props) { return <span>{x}</span>; }", Tsx: true, Options: []any{map[string]any{"functions": "defaultArguments"}}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 43, EndLine: 1, EndColumn: 44}, {MessageId: "shouldAssignObjectDefault", Message: "propType \"x\" is not required, but has no corresponding default argument value.", Line: 1, Column: 95, EndLine: 1, EndColumn: 96}}},
		// ---- Regression: reused runtime contracts stay scoped to each component ----
		{Code: "const props = { x: PropTypes.string }; function A() { return <div />; } A.propTypes = props; function B() { return <span />; } B.propTypes = props;", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 17, EndLine: 1, EndColumn: 36}, {MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 17, EndLine: 1, EndColumn: 36}}},
		// ---- Regression: declaration markers inside component namespaces are preserved ----
		{Code: "const ns = { propTypes: { C: () => <div /> } }; ns.propTypes.C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 78, EndLine: 1, EndColumn: 97}}},
		// ---- Intentional difference: member-assigned object containers retain nested paths ----
		{Code: "const ns={};ns.registry={C:()=> <div/>};ns.registry.C.propTypes={x:PropTypes.string};", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 66, EndLine: 1, EndColumn: 84}}},
		// ---- Regression: a prop named propTypes is not mistaken for a declaration marker ----
		{Code: "function C() { return <div />; } C.propTypes = { propTypes: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"propTypes\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 50, EndLine: 1, EndColumn: 77}}},
		// ---- Regression: distinct declarations on reassigned components remain distinct ----
		{Code: "let C = () => <div />; C.propTypes = { x: PropTypes.string }; C = () => <span />; C.propTypes = { y: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 40, EndLine: 1, EndColumn: 59}, {MessageId: "shouldHaveDefault", Message: "propType \"y\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 99, EndLine: 1, EndColumn: 118}}},
		// ---- Intentional difference: a later component remains visible after a non-component value ----
		{Code: "let C = () => 1; C = () => <div />; C.propTypes = { x: PropTypes.string };", Tsx: true, Options: []any{}, Settings: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "shouldHaveDefault", Message: "propType \"x\" is not required, but has no corresponding defaultProps declaration.", Line: 1, Column: 53, EndLine: 1, EndColumn: 72}}},
	})
}
