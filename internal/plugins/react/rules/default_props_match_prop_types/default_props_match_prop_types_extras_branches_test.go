package default_props_match_prop_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestDefaultPropsMatchPropTypesExtrasBranches checks branch boundaries and syntax
// variations beyond the upstream suite in default_props_match_prop_types_upstream_test.go.
// N/A: fixes and suggestions; this rule only emits diagnostics.
func TestDefaultPropsMatchPropTypesExtrasBranches(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DefaultPropsMatchPropTypesRule,
		[]rule_tester.ValidTestCase{
			// ---- Branch: no defaults ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };", Tsx: true},
			// ---- Branch: no declared types ----
			{Code: "function C(props) { return <div/>; }C.defaultProps={a:1};", Tsx: true},
			// ---- Branch: empty declared types ----
			{Code: "function C(props) { return <div/>; }C.propTypes={};C.defaultProps={a:1};", Tsx: true},
			// ---- Branch: empty defaults ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={};", Tsx: true},
			// ---- Branch: optional default ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={b:1};", Tsx: true},
			// ---- Branch: unresolved sticks ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps=external;C.defaultProps={a:1};C.defaultProps.z=1;", Tsx: true},
			// ---- Branch: default spread ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={...other,a:1};", Tsx: true},
			// ---- Branch: no wrapper argument ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps=wrap();", Tsx: true, Options: []any{}, Settings: map[string]any{"propWrapperFunctions": []any{"wrap"}}},
			// ---- Branch: default property wrapper not resolved ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps=Object.freeze({a:1});", Tsx: true, Options: []any{}, Settings: map[string]any{"propWrapperFunctions": []any{"Object.freeze"}}},
			// ---- Branch: default alias chain ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };const a={a:1};const b=a;C.defaultProps=b;", Tsx: true},
			// ---- Branch: required validator alias ----
			{Code: "function C(props) { return <div/>; }const validator=P.string.isRequired;C.propTypes={a:validator};C.defaultProps={a:1};", Tsx: true},
			// ---- Branch: empty getter ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static get defaultProps(){}}", Tsx: true},
			// ---- Branch: nonstatic defaults ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};defaultProps={a:1};get defaultProps(){return {a:1}}}", Tsx: true},
			// ---- Branch: legacy alias not resolved ----
			{Code: "const defaults={a:1};const C=React.createClass({propTypes:{a:P.string.isRequired},getDefaultProps(){return defaults},render(){return <div/>}});", Tsx: true},
			// ---- Branch: legacy object method ----
			{Code: "const C=React.createClass({propTypes:{a:P.string.isRequired},getDefaultProps(){return {a:1}},render(){return <div/>}});", Tsx: true},
			// ---- Branch: empty switch ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static get defaultProps(){switch(x){}}}", Tsx: true},
			// ---- Branch: switch no return ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static get defaultProps(){return {a:1};switch(x){default:break;}}}", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Branch: required default ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 107, EndLine: 1, EndColumn: 110}}},
			// ---- Branch: missing default ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={c:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "defaultHasNoType", Message: "defaultProp \"c\" has no corresponding propTypes declaration.", Line: 1, Column: 107, EndLine: 1, EndColumn: 110}}},
			// ---- Branch: options [{}] ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1,c:1};", Tsx: true, Options: []any{map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 107, EndLine: 1, EndColumn: 110}, {MessageId: "defaultHasNoType", Message: "defaultProp \"c\" has no corresponding propTypes declaration.", Line: 1, Column: 111, EndLine: 1, EndColumn: 114}}},
			// ---- Branch: options [{"allowRequiredDefaults":false}] ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1,c:1};", Tsx: true, Options: []any{map[string]any{"allowRequiredDefaults": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 107, EndLine: 1, EndColumn: 110}, {MessageId: "defaultHasNoType", Message: "defaultProp \"c\" has no corresponding propTypes declaration.", Line: 1, Column: 111, EndLine: 1, EndColumn: 114}}},
			// ---- Branch: options [{"allowRequiredDefaults":true}] ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1,c:1};", Tsx: true, Options: []any{map[string]any{"allowRequiredDefaults": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "defaultHasNoType", Message: "defaultProp \"c\" has no corresponding propTypes declaration.", Line: 1, Column: 111, EndLine: 1, EndColumn: 114}}},
			// ---- Branch: default wrapper ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps=wrap({a:1});", Tsx: true, Options: []any{}, Settings: map[string]any{"propWrapperFunctions": []any{"wrap"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 112, EndLine: 1, EndColumn: 115}}},
			// ---- Branch: object wrapper setting ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps=wrap({a:1});", Tsx: true, Options: []any{}, Settings: map[string]any{"propWrapperFunctions": []any{map[string]any{"property": "wrap"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 112, EndLine: 1, EndColumn: 115}}},
			// ---- Branch: property wrapper ----
			{Code: "function C(props) { return <div/>; }C.propTypes=Object.freeze({a:P.string.isRequired});C.defaultProps={a:1};", Tsx: true, Options: []any{}, Settings: map[string]any{"propWrapperFunctions": []any{map[string]any{"object": "Object", "property": "freeze"}}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 104, EndLine: 1, EndColumn: 107}}},
			// ---- Branch: runtime alias chain ----
			{Code: "function C(props) { return <div/>; }const a={a:P.string.isRequired};const b=a;C.propTypes=b;C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 109, EndLine: 1, EndColumn: 112}}},
			// ---- Branch: duplicate defaults ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1,a:2};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 111, EndLine: 1, EndColumn: 114}}},
			// ---- Branch: successive declarations ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};C.propTypes={a:P.string};C.defaultProps={c:2};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "defaultHasNoType", Message: "defaultProp \"c\" has no corresponding propTypes declaration.", Line: 1, Column: 153, EndLine: 1, EndColumn: 156}}},
			// ---- Branch: class default getters ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static get defaultProps(){if(x)return {z:1};return {a:1};}}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 127, EndLine: 1, EndColumn: 130}}},
			// ---- Branch: opaque getter ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static get defaultProps(){return unknown;}}C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 134, EndLine: 1, EndColumn: 137}}},
			// ---- Branch: static getDefaultProps field ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static getDefaultProps={a:1};}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 99, EndLine: 1, EndColumn: 102}}},
			// ---- Branch: switch last case ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static get defaultProps(){switch(x){case 0:return {b:1};default:return {a:1};}}}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 147, EndLine: 1, EndColumn: 150}}},
			// ---- Branch: nested switch ----
			{Code: "class C extends React.Component {static propTypes={a:P.string.isRequired};static get defaultProps(){switch(x){case 0:switch(y){case 1:return {a:1};}}}}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 143, EndLine: 1, EndColumn: 146}}},
		},
	)
}
