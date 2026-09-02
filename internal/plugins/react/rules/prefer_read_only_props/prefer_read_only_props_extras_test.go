// TestPreferReadOnlyPropsExtras covers tsgo/ESTree edge shapes, real-user
// regressions, and branch lock-ins. The upstream mirror is kept separately in
// prefer_read_only_props_upstream_test.go.
package prefer_read_only_props

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferReadOnlyPropsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferReadOnlyPropsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized type wrappers ----
		{Code: `type Props = (({ readonly name: string })); class Hello extends React.Component<Props> { render() { return <div/>; } }`, Tsx: true},
		{Code: `const Hello = (props: ((({ readonly name: string })))) => <div/>;`, Tsx: true},

		// ---- Dimension 4: property key equivalence classes ----
		{Code: `class Hello extends React.Component<{ readonly "name": string; readonly 0: string; readonly [key: string]: string }> { render() { return <div/>; } }`, Tsx: true},
		{Code: `class Hello extends React.Component<{ [name: string]: string; method(): void; readonly name?: string }> { render() { return <div/>; } }`, Tsx: true},

		// ---- Dimension 4: declaration/container forms ----
		{Code: `const Hello = class extends React.Component<{ readonly name: string }> { render() { return <div/>; } };`, Tsx: true},
		{Code: `const Hello = class extends React.Component<{ name: string }> { render() { return <div/>; } };`, Tsx: true},
		{Code: `type Props = { name: string }; const Hello = forwardRef<HTMLDivElement, Props>((props, ref) => <div/>);`, Tsx: true},
		{Code: `abstract class Hello extends React.Component<{ readonly name: string }> { abstract render(): React.ReactNode; }`, Tsx: true},
		{Code: `class Hello extends React.Component<{ readonly name: string }> { static readonly value = 1; render() { return <div/>; } }`, Tsx: true},

		// ---- Dimension 4: graceful degradation ----
		{Code: `type Props = { readonly name: string } & OtherProps; const Hello = (props: Props) => <div/>;`, Tsx: true},
		{Code: `type Props = {}; const Hello = ({}: Props) => <div/>;`, Tsx: true},
		{Code: `class Hello extends React.Component {} const value = (props: { name: string }) => null;`, Tsx: true},
		{Code: `declare function Hello(props: { name: string }): JSX.Element;`, Tsx: true},
		{Code: `type Props = { name: string }; const Hello = React.forwardRef<HTMLDivElement, Props>(((props, ref) => <div ref={ref} />)!);`, Tsx: true},

		// ---- Branch lock-ins: component classification and type selection ----
		// Locks in upstream isSuperTypeParameterPropsDeclaration()'s no-type-args arm.
		{Code: `class Hello extends React.Component { render() { return <div/>; } }`, Tsx: true},
		// Locks in upstream isES6Component()'s bare Component / PureComponent arms.
		{Code: `class Hello extends Component<{ readonly name: string }> { render() { return <div/>; } }`, Tsx: true},
		{Code: `class Hello extends React.PureComponent<{ readonly name: string }> { render() { return <div/>; } }`, Tsx: true},
		// Locks in upstream isValidReactGenericTypeAnnotation()'s rejected generic arm.
		{Code: `import React from "react"; type Props = { name: string }; const notAComponent: MyFC<Props> = (props) => <div/>;`, Tsx: true},
		// The released upstream rule requires a React generic import for this arm.
		// Real-user #3650: the modern JSX runtime leaves React unimported, so the
		// released rule intentionally produces no diagnostic for this shape.
		{Code: `type Props = { name: string }; const Chip: React.FC<Props> = ({name}) => <div>{name}</div>;`, Tsx: true},
		// Locks in the TS visitor's union fallback: unions are not traversed.
		{Code: `import React from "react"; type Props = { name: string } | { other: string }; const Hello: React.FC<Props> = (props) => <div/>;`, Tsx: true},
		// Locks in the function-without-props branch.
		{Code: `const Hello = () => <div/>;`, Tsx: true},

		// ---- Real-user: #3786 namespace-qualified props type ----
		// v7.37.5 searches only top-level declarations, so a namespace member is
		// not resolved and remains silent, matching the released implementation.
		{Code: `namespace ItemsListElementSkeleton { export interface Props { name?: string } } function ItemsListElementSkeleton({name}: ItemsListElementSkeleton.Props) { return <div>{name}</div>; }`, Tsx: true},
		{Code: `namespace ItemsListElementSkeleton { export interface Props { name?: string } } type Props = { top: string }; function ItemsListElementSkeleton({name}: ItemsListElementSkeleton.Props) { return <div>{name}</div>; }`, Tsx: true},
		// ---- Real-user: #3650 implicit React reference ----
		{Code: `interface ChipProps { chipColor: string; label: string } const Chip: React.FC<ChipProps> = ({chipColor, label}) => <div>{chipColor}{label}</div>;`, Tsx: true},
		// Bare wrapper names are not React components unless imported from React.
		{Code: `type Props = { name: string }; const Hello = memo((props: Props) => <div/>);`, Tsx: true},
		{Code: `type Props = { name: string }; const Hello = forwardRef((props: Props) => <div/>);`, Tsx: true},

		// N/A: receiver/member access and element access — the rule inspects type
		// declarations, not JavaScript expressions.
		// N/A: RestElement bindings — function parameter binding shape is irrelevant
		// once its type annotation has been collected.
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: optional property and multiline diagnostic range ----
		{
			Code: `type Props = {
  name?: string;
  title: string;
};
function Hello(props: Props) {
  return <div>{props.name}{props.title}</div>;
}`,
			Tsx: true,
			Output: []string{`type Props = {
  readonly name?: string;
  readonly title: string;
};
function Hello(props: Props) {
  return <div>{props.name}{props.title}</div>;
}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only.", Line: 2, Column: 3, EndLine: 2, EndColumn: 17},
				{MessageId: "readOnlyProp", Message: "Prop 'title' should be read-only.", Line: 3, Column: 3, EndLine: 3, EndColumn: 17},
			},
		},
		// ---- Branch lock-in: imported named generic arm ----
		{
			Code:   `import { FC } from "react"; type Props = { name: string }; const Hello: FC<Props> = (props) => <div/>;`,
			Tsx:    true,
			Output: []string{`import { FC } from "react"; type Props = { readonly name: string }; const Hello: FC<Props> = (props) => <div/>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `import { FC } from "react"; function Hello(props: FC<{ name: string }>) { return <div/>; }`,
			Tsx:    true,
			Output: []string{`import { FC } from "react"; function Hello(props: FC<{ readonly name: string }>) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `import { PropsWithChildren } from "react"; function Hello(props: PropsWithChildren<{ name: string }>) { return <div/>; }`,
			Tsx:    true,
			Output: []string{`import { PropsWithChildren } from "react"; function Hello(props: PropsWithChildren<{ readonly name: string }>) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `import { ComponentPropsWithRef } from "react"; class Hello extends React.Component<ComponentPropsWithRef<{ name: string }>> { render() { return <div/>; } }`,
			Tsx:    true,
			Output: []string{`import { ComponentPropsWithRef } from "react"; class Hello extends React.Component<ComponentPropsWithRef<{ readonly name: string }>> { render() { return <div/>; } }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:     `type Props = { name: string }; const Hello = customMemo((props: Props) => <div/>);`,
			Settings: map[string]interface{}{"componentWrapperFunctions": []interface{}{"customMemo"}},
			Tsx:      true,
			Output:   []string{`type Props = { readonly name: string }; const Hello = customMemo((props: Props) => <div/>);`},
			Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		// ---- Real-user: #3653 async server component ----
		{
			Code:   `type Props = { enabled: boolean }; const AsyncComponent = async ({enabled}: Props) => <div>{enabled}</div>;`,
			Tsx:    true,
			Output: []string{`type Props = { readonly enabled: boolean }; const AsyncComponent = async ({enabled}: Props) => <div>{enabled}</div>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'enabled' should be read-only."}},
		},
		// ---- Branch lock-in: forwardRef props type argument ----
		{
			Code:   `import React from "react"; type Props = { name: string }; const Hello = React.forwardRef<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.name}</div>);`,
			Tsx:    true,
			Output: []string{`import React from "react"; type Props = { readonly name: string }; const Hello = React.forwardRef<HTMLDivElement, Props>((props, ref) => <div ref={ref}>{props.name}</div>);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `type Props = { ['name']: string; [name]: number; [1]: boolean; [+1]: unknown }; function Hello(props: Props) { return <div/>; }`,
			Tsx:    true,
			Output: []string{`type Props = { ['name']: string; readonly [name]: number; readonly [1]: boolean; readonly [+1]: unknown }; function Hello(props: Props) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."},
				{MessageId: "readOnlyProp", Message: "Prop '1' should be read-only."},
				{MessageId: "readOnlyProp", Message: "Prop 'undefined' should be read-only."},
			},
		},
		{
			Code:   `type Props = { name: string }; const obj = { Hello(props: Props) { return <div/>; } };`,
			Tsx:    true,
			Output: []string{`type Props = { readonly name: string }; const obj = { Hello(props: Props) { return <div/>; } };`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `import { ForwardRefRenderFunction } from "react"; type Props = { name: string }; const Hello: ForwardRefRenderFunction<HTMLDivElement, Props> = (props, ref) => <div/>;`,
			Tsx:    true,
			Output: []string{`import { ForwardRefRenderFunction } from "react"; type Props = { readonly name: string }; const Hello: ForwardRefRenderFunction<HTMLDivElement, Props> = (props, ref) => <div/>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `import { forwardRef } from "react"; type Props = { name: string }; const Hello: forwardRef<HTMLDivElement, Props> = (props, ref) => <div/>;`,
			Tsx:    true,
			Output: []string{`import { forwardRef } from "react"; type Props = { readonly name: string }; const Hello: forwardRef<HTMLDivElement, Props> = (props, ref) => <div/>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `type Props = { name: string }; class Hello extends React.Component<Props> { props: Props; render() { return <div/>; } }`,
			Tsx:    true,
			Output: []string{`type Props = { readonly name: string }; class Hello extends React.Component<Props> { props: Props; render() { return <div/>; } }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `import React from "react"; type Props = { name: string }; const Hello: React.FC<Props> = (props: Props) => <div/>;`,
			Tsx:    true,
			Output: []string{`import React from "react"; type Props = { readonly name: string }; const Hello: React.FC<Props> = (props: Props) => <div/>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `type Props = { [true]: string }; function Hello(props: Props) { return <div/>; }`,
			Tsx:    true,
			Output: []string{`type Props = { readonly [true]: string }; function Hello(props: Props) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'true' should be read-only."}},
		},
		{
			Code:   `type Props = { 0x1: string }; function Hello(props: Props) { return <div/>; }`,
			Tsx:    true,
			Output: []string{`type Props = { readonly 0x1: string }; function Hello(props: Props) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop '0x1' should be read-only."}},
		},
		{
			Code:   `type Props = { [0x1]: string }; function Hello(props: Props) { return <div/>; }`,
			Tsx:    true,
			Output: []string{`type Props = { readonly [0x1]: string }; function Hello(props: Props) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop '0x1' should be read-only."}},
		},
		{
			Code:   `interface A { name: string } interface B extends A { title: string } class Hello extends React.Component<B> { render() { return <div/>; } }`,
			Tsx:    true,
			Output: []string{`interface A { readonly name: string } interface B extends A { readonly title: string } class Hello extends React.Component<B> { render() { return <div/>; } }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."},
				{MessageId: "readOnlyProp", Message: "Prop 'title' should be read-only."},
			},
		},
		{
			Code:   `type Props = { name: string } & { name: number }; function Hello(props: Props) { return <div/>; }`,
			Tsx:    true,
			Output: []string{`type Props = { name: string } & { readonly name: number }; function Hello(props: Props) { return <div/>; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `const Hello = (props: ReturnType<() => { name: string }>) => <div/>;`,
			Tsx:    true,
			Output: []string{`const Hello = (props: ReturnType<() => { readonly name: string }>) => <div/>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `interface Props extends ReturnType<() => { name: string }> {} const Hello = (props: Props) => <div/>;`,
			Tsx:    true,
			Output: []string{`interface Props extends ReturnType<() => { readonly name: string }> {} const Hello = (props: Props) => <div/>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code:   `import React from "react"; type P = { a: string }; type Q = { b: string }; const Hello: React.FC<P> = (props: Q) => <div/>;`,
			Tsx:    true,
			Output: []string{`import React from "react"; type P = { a: string }; type Q = { readonly b: string }; const Hello: React.FC<P> = (props: Q) => <div/>;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'b' should be read-only."}},
		},
		{
			Code:   `type Props = { name: string }; class Hello extends React.Component<(Props)> { render() { return <div/>; } }`,
			Tsx:    true,
			Output: []string{`type Props = { readonly name: string }; class Hello extends React.Component<(Props)> { render() { return <div/>; } }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
	})
}

func TestPreferReadOnlyPropsLongAliasChain(t *testing.T) {
	var source strings.Builder
	for i := 0; i < 34; i++ {
		name := string(rune('A' + i))
		next := string(rune('A' + i + 1))
		if i == 25 {
			name = "Z"
			next = "AA"
		}
		if i > 25 {
			name = "A" + string(rune('A'+i-26))
			next = "A" + string(rune('A'+i-25))
		}
		source.WriteString("type " + name + " = " + next + "; ")
	}
	source.WriteString("type AI = { name: string }; function Hello(props: A) { return <div/>; }")
	code := source.String()
	output := strings.Replace(code, "type AI = { name: string }", "type AI = { readonly name: string }", 1)
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferReadOnlyPropsRule, nil, []rule_tester.InvalidTestCase{{
		Code:   code,
		Tsx:    true,
		Output: []string{output},
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: "readOnlyProp",
			Message:   "Prop 'name' should be read-only.",
		}},
	}})
}
