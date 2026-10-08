package default_props_match_prop_types

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestDefaultPropsMatchPropTypesExtras checks branch boundaries and syntax
// variations beyond the upstream suite in default_props_match_prop_types_upstream_test.go.
// N/A: fixes and suggestions; this rule only emits diagnostics.
func TestDefaultPropsMatchPropTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DefaultPropsMatchPropTypesRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: key [a] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes={[a]:P.string.isRequired};C.defaultProps={[a]:1};", Tsx: true},
			// ---- Dimension 4: key [Symbol.iterator] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes={[Symbol.iterator]:P.string.isRequired};C.defaultProps={[Symbol.iterator]:1};", Tsx: true},
			// ---- Dimension 4: propTypes access [propTypes] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C[propTypes]={a:P.string.isRequired};C.defaultProps={a:1};", Tsx: true},
			// ---- Dimension 4: defaultProps access [defaultProps] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C[defaultProps]={a:1};", Tsx: true},
			// ---- Dimension 4: individual key [a] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={};C.defaultProps[a]=1;", Tsx: true},
			// ---- Dimension 4: optional validator ----
			{Code: "function C(props) { return <div/>; }C.propTypes={a:P?.string.isRequired};C.defaultProps={a:1};", Tsx: true},
			// ---- Dimension 4: optional wrapper ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps=wrap?.({a:1});", Tsx: true, Options: []any{}, Settings: map[string]any{"propWrapperFunctions": []any{"wrap"}}},
			// ---- Dimension 4: container async function* C(props){return <div/>;} ----
			{Code: "async function* C(props){return <div/>;};C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true},
			// ---- Dimension 4: shadowed assignment ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };function f(C){C.defaultProps={a:1};}", Tsx: true},
			// ---- Dimension 4: empty and bodyless ----
			{Code: "declare class C extends React.Component {static defaultProps: unknown; render(): unknown;} function Empty(){};", Tsx: true},
			// ---- TypeScript: union opaque ----
			{Code: "type Props={a:string}|{b:string};function C(p:Props){return <div/>;}C.defaultProps={a:1};", Tsx: true},
			// ---- Real-user: #1908 quoted optional defaults ----
			{Code: "function C(){return <div/>;}C.propTypes={\"firstProperty\":P.string.isRequired,\"secondProperty\":P.bool,\"thirdProperty\":P.func};C.defaultProps={\"secondProperty\":false,\"thirdProperty\":()=>undefined};", Tsx: true},
			// ---- Real-user: #3138 runtime spread ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "const testShape={isDeleted:P.bool};const C=({custom,isDeleted})=><div/>;C.propTypes={custom:P.bool.isRequired,...testShape};C.defaultProps={isDeleted:false};", Tsx: true},
			// ---- TypeScript: shadowed alias ----
			// NOTE: Intentional: lexical type bindings and unknown imported declarations prevent false missing/required-prop reports.
			{Code: "type Props={a:string};function outer(){type Props={a?:string};const C=(props:Props)=><div/>;C.defaultProps={a:1};}", Tsx: true},
			// ---- Real-user: #3665 imported intersection ----
			// NOTE: Intentional: lexical type bindings and unknown imported declarations prevent false missing/required-prop reports.
			{Code: "import type {External} from \"./types\";type Props=External & {known?:string};function C(props:Props){return <div/>;}C.defaultProps={unknown:1};", Tsx: true},
			// ---- Regression: a static props field is class metadata, not a component props contract ----
			{Code: "type Metadata={a:string};class C extends React.Component{static props:Metadata;static defaultProps={a:1};render(){return <div/>;}}", Tsx: true},
			// ---- Regression: runtime propTypes must be a static class field ----
			{Code: "class C extends React.Component{propTypes={a:P.string.isRequired};static defaultProps={a:1};render(){return <div/>;}}", Tsx: true},
			// ---- Regression: a same-file React namespace shadows the ambient/default pragma ----
			{Code: "namespace React{export interface FC<P>{(props:P):unknown}}type Props={a:string};const C:React.FC<Props>=(props)=><div/>;C.defaultProps={a:1};", Tsx: true},
			// ---- Regression: a local configured-pragma receiver is not React.forwardRef ----
			{Code: "const R={forwardRef:<T,P>(fn:(props:P)=>unknown)=>fn};type Props={a:string};const C=R.forwardRef<HTMLDivElement,Props>((props)=><div/>);C.defaultProps={a:1};", Tsx: true, Settings: map[string]any{"react": map[string]any{"pragma": "R"}}},
			// ---- Regression: React.FC without type arguments is safe ----
			{Code: "import React from \"react\";const C:React.FC=(props)=><div/>;", Tsx: true},
			// ---- Regression control: implicit children remains optional ----
			{Code: "import React from \"react\";const C:React.FC<{name:string}>=(props)=><div/>;C.defaultProps={children:\"x\"};", Tsx: true},
			// ---- Regression control: explicit optional children remains optional ----
			{Code: "import React from \"react\";const C:React.FC<{children?:string}>=(props)=><div/>;C.defaultProps={children:\"x\"};", Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receivers ----
			{Code: "function C(props) { return <div/>; }((C)).propTypes=({a:(P.string).isRequired});((C)).defaultProps=({a:1});", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 102, EndLine: 1, EndColumn: 105}}},
			// ---- Dimension 4: receiver C! ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C!.propTypes={a:P.string.isRequired};C!.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 91, EndLine: 1, EndColumn: 94}}},
			// ---- Dimension 4: receiver (C as any) ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }(C as any).propTypes={a:P.string.isRequired};(C as any).defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 107, EndLine: 1, EndColumn: 110}}},
			// ---- Dimension 4: receiver (C satisfies any) ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }(C satisfies any).propTypes={a:P.string.isRequired};(C satisfies any).defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 121, EndLine: 1, EndColumn: 124}}},
			// ---- Dimension 4: key a ----
			{Code: "function C(props) { return <div/>; }C.propTypes={a:P.string.isRequired};C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 89, EndLine: 1, EndColumn: 92}}},
			// ---- Dimension 4: key 'a' ----
			{Code: "function C(props) { return <div/>; }C.propTypes={'a':P.string.isRequired};C.defaultProps={'a':1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 91, EndLine: 1, EndColumn: 96}}},
			// ---- Dimension 4: key 0 ----
			{Code: "function C(props) { return <div/>; }C.propTypes={0:P.string.isRequired};C.defaultProps={0:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"0\" defined for isRequired propType.", Line: 1, Column: 89, EndLine: 1, EndColumn: 92}}},
			// ---- Dimension 4: key 0x10 ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes={0x10:P.string.isRequired};C.defaultProps={0x10:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"16\" defined for isRequired propType.", Line: 1, Column: 92, EndLine: 1, EndColumn: 98}}},
			// ---- Dimension 4: key ['a'] ----
			{Code: "function C(props) { return <div/>; }C.propTypes={['a']:P.string.isRequired};C.defaultProps={['a']:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 93, EndLine: 1, EndColumn: 100}}},
			// ---- Dimension 4: key [`a`] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes={[`a`]:P.string.isRequired};C.defaultProps={[`a`]:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 93, EndLine: 1, EndColumn: 100}}},
			// ---- Dimension 4: propTypes access ['propTypes'] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C['propTypes']={a:P.string.isRequired};C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 92, EndLine: 1, EndColumn: 95}}},
			// ---- Dimension 4: propTypes access [`propTypes`] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C[`propTypes`]={a:P.string.isRequired};C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 92, EndLine: 1, EndColumn: 95}}},
			// ---- Dimension 4: defaultProps access ['defaultProps'] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C['defaultProps']={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 110, EndLine: 1, EndColumn: 113}}},
			// ---- Dimension 4: defaultProps access [`defaultProps`] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C[`defaultProps`]={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 110, EndLine: 1, EndColumn: 113}}},
			// ---- Dimension 4: individual key .a ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={};C.defaultProps.a=1;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 109, EndLine: 1, EndColumn: 127}}},
			// ---- Dimension 4: individual key ['a'] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={};C.defaultProps['a']=1;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 109, EndLine: 1, EndColumn: 130}}},
			// ---- Dimension 4: individual key [`a`] ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={};C.defaultProps[`a`]=1;", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 109, EndLine: 1, EndColumn: 130}}},
			// ---- Dimension 4: shorthand method accessor defaults ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };const a=1;C.defaultProps={a,b(){},get c(){return 1}};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 117, EndLine: 1, EndColumn: 118}, {MessageId: "defaultHasNoType", Message: "defaultProp \"c\" has no corresponding propTypes declaration.", Line: 1, Column: 125, EndLine: 1, EndColumn: 142}}},
			// ---- Dimension 4: container const C = function(props){return <div/>;} ----
			{Code: "const C = function(props){return <div/>;};C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 113, EndLine: 1, EndColumn: 116}}},
			// ---- Dimension 4: container const C = (props)=> <div/>; ----
			{Code: "const C = (props)=> <div/>;;C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 99, EndLine: 1, EndColumn: 102}}},
			// ---- Dimension 4: container async function C(props){return <div/>;} ----
			{Code: "async function C(props){return <div/>;};C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 111, EndLine: 1, EndColumn: 114}}},
			// ---- Dimension 4: container function* C(props){return <div/>;} ----
			{Code: "function* C(props){return <div/>;};C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 106, EndLine: 1, EndColumn: 109}}},
			// ---- Dimension 4: container const C = class extends React.Component {}; ----
			{Code: "const C = class extends React.Component {};;C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 115, EndLine: 1, EndColumn: 118}}},
			// ---- Dimension 4: container class C extends React.Component { #a=1; } ----
			{Code: "class C extends React.Component { #a=1; };C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 113, EndLine: 1, EndColumn: 116}}},
			// ---- Dimension 4: nested component ----
			{Code: "function C(props) { return <div/>; }C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};function Outer(){const C=()=> <div/>;C.propTypes={a:P.string};C.defaultProps={a:1};return <C/>;}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 107, EndLine: 1, EndColumn: 110}}},
			// ---- Dimension 4: rest parameter ----
			{Code: "function C({a,...rest}) {return <div/>;}C.propTypes = { a: P.string.isRequired, b: P.string };C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 111, EndLine: 1, EndColumn: 114}}},
			// ---- TypeScript: interfaces ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "interface A { a:string };interface Props extends A {b?:string};function C(p:Props){return <div/>;}C.defaultProps={a:1,b:1,z:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 115, EndLine: 1, EndColumn: 118}, {MessageId: "defaultHasNoType", Message: "defaultProp \"z\" has no corresponding propTypes declaration.", Line: 1, Column: 123, EndLine: 1, EndColumn: 126}}},
			// ---- TypeScript: merging ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "interface Props{a?:string}interface Props{b:string}function C(p:Props){return <div/>;}C.defaultProps={a:1,b:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"b\" defined for isRequired propType.", Line: 1, Column: 107, EndLine: 1, EndColumn: 110}}},
			// ---- TypeScript: recursive ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "type Props = Props & {a:string};function C(p:Props){return <div/>;}C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 84, EndLine: 1, EndColumn: 87}}},
			// ---- TypeScript: return type ----
			// NOTE: Intentional: normalize static keys, keep dynamic declarations opaque, and analyze TypeScript props independently of the parameter name.
			{Code: "type Props=ReturnType<()=>{a:string}>;function C(p:Props){return <div/>;}C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 90, EndLine: 1, EndColumn: 93}}},
			// ---- TypeScript: imported FC ----
			{Code: "import {FC} from \"react\";const C:FC<{a:string}>=(props)=><div/>;C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 81, EndLine: 1, EndColumn: 84}}},
			// ---- TypeScript: namespace FC ----
			{Code: "import React from \"react\";const C:React.FC<{a:string}>=(props)=><div/>;C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 88, EndLine: 1, EndColumn: 91}}},
			// ---- TypeScript: forwardRef ----
			{Code: "import React from \"react\";const C=React.forwardRef<HTMLDivElement,{a:string}>((props,ref)=><div/>);C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 116, EndLine: 1, EndColumn: 119}}},
			// ---- Locations: astral CRLF ----
			{Code: "// 😀\r\nfunction C(props) { return <div/>; }\r\nC.propTypes = { a: P.string.isRequired, b: P.string };\r\nC.defaultProps={ /* 😀 */ a:1,c:2};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 4, Column: 27, EndLine: 4, EndColumn: 30}, {MessageId: "defaultHasNoType", Message: "defaultProp \"c\" has no corresponding propTypes declaration.", Line: 4, Column: 31, EndLine: 4, EndColumn: 34}}},
			// ---- Binding: assignment before declaration ----
			{Code: "C.propTypes={a:P.string.isRequired};C.defaultProps={a:1};function C(props) { return <div/>; }", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 53, EndLine: 1, EndColumn: 56}}},
			// ---- Binding: named class expression ----
			{Code: "const C=class Inner extends React.Component {static propTypes={a:P.string.isRequired};static defaultProps={a:1}};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 108, EndLine: 1, EndColumn: 111}}},
			// ---- Binding: namespace component ----
			{Code: "const Views={};Views.C=()=> <div/>;Views.C.propTypes={a:P.string.isRequired};Views.C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 100, EndLine: 1, EndColumn: 103}}},
			// ---- Pragma: custom ----
			{Code: "class C extends R.Component {static propTypes={a:P.string.isRequired};static defaultProps={a:1};}", Tsx: true, Options: []any{}, Settings: map[string]any{"react": map[string]any{"pragma": "R"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 92, EndLine: 1, EndColumn: 95}}},
			// ---- Regression: ambient React.FC is stable across typed and source-only Programs ----
			{Code: "/// <reference path=\"./ambient-react.d.ts\" />\ntype Props={a:string};const C:React.FC<Props>=(props)=><div/>;C.defaultProps={a:1};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 2, Column: 79, EndLine: 2, EndColumn: 82}}},
			// ---- Regression: configured pragma namespace forwardRef contributes generic props ----
			{Code: "import * as R from \"react\";type Props={a:string};const C=R.forwardRef<HTMLDivElement,Props>((props)=><div/>);C.defaultProps={a:1};", Tsx: true, Settings: map[string]any{"react": map[string]any{"pragma": "R"}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 126, EndLine: 1, EndColumn: 129}}},
			// ---- Regression control: an instance props field is the class props contract ----
			{Code: "type Props={a:string};class C extends React.Component{props:Props;static defaultProps={a:1};render(){return <div/>;}}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 88, EndLine: 1, EndColumn: 91}}},
			// ---- Regression control: static propTypes is the runtime props contract ----
			{Code: "class C extends React.Component{static propTypes={a:P.string.isRequired};static defaultProps={a:1};render(){return <div/>;}}", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"a\" defined for isRequired propType.", Line: 1, Column: 95, EndLine: 1, EndColumn: 98}}},
			// ---- Regression: explicit required children is not overwritten by React.FC ----
			{Code: "import React from \"react\";const C:React.FC<{children:string}>=(props)=><div/>;C.defaultProps={children:\"x\"};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"children\" defined for isRequired propType.", Line: 1, Column: 95, EndLine: 1, EndColumn: 107}}},
			// ---- Regression: PropsWithChildren does not overwrite explicit required children ----
			{Code: "import React from \"react\";const C:React.FC<React.PropsWithChildren<{children:string}>>=(props)=><div/>;C.defaultProps={children:\"x\"};", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "requiredHasDefault", Message: "defaultProp \"children\" defined for isRequired propType.", Line: 1, Column: 120, EndLine: 1, EndColumn: 132}}},
		},
	)
}

func TestDefaultPropsMatchPropTypesSourceOnlyAmbientReact(t *testing.T) {
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "ambient-react-source-only.tsx")
	code := "type Props={a:string};const C:React.FC<Props>=(props)=><div/>;C.defaultProps={a:1};"
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{fileName},
		Host:          utils.CreateCompilerHost(root.Dir, fs),
		CompilerOptions: &core.CompilerOptions{
			Jsx:    core.JsxEmitPreserve,
			Module: core.ModuleKindESNext,
			Target: core.ScriptTargetESNext,
		},
		SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	var diagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program: program,
		File:    fileName,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     DefaultPropsMatchPropTypesRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					if ctx.TypeChecker != nil {
						t.Fatal("expected source-only linting without a TypeChecker")
					}
					return DefaultPropsMatchPropTypesRule.Run(ctx, nil)
				},
			}}
		},
		Consumer: rule.DiagnosticConsumer{
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		},
	})
	if len(diagnostics) != 1 || diagnostics[0].Message.Id != "requiredHasDefault" {
		t.Fatalf("diagnostics = %+v, want one requiredHasDefault", diagnostics)
	}
}
