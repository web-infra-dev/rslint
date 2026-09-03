// TestPreferReadOnlyPropsUpstream migrates the TypeScript cases from
// eslint-plugin-react v7.37.5's prefer-read-only-props suite.
// The upstream file also contains Flow/parser cases; those are not part of
// this TypeScript-only mirror because rslint does not support Flow.
// rslint-specific edge and branch cases live in prefer_read_only_props_extras_test.go.
package prefer_read_only_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferReadOnlyPropsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferReadOnlyPropsRule, []rule_tester.ValidTestCase{
		{
			Code: `
        import React from "react";

        interface Props {
          readonly name: string;
        }

        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
			Tsx: true,
		},
		{
			Code: `
        import React from "react";
        type Props = {
          readonly firstName: string;
          readonly lastName: string;
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
			Tsx: true,
		},
		{
			Code: `
        import React from "react";
        type Props = {
          readonly name: string;
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
			Tsx: true,
		},
		{
			Code: `
        import React from "react";
        type Props = {
          readonly name: string[];
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
			Tsx: true,
		},
		{
			Code: `
        import React from "react";
        type Props = {
          readonly name: string[];
        }
        const MyComponent: React.FC<Props> = async ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
			Tsx: true,
		},
		{
			Code: `
        import React from "react";
        type Props = {
          readonly person: {
            name: string;
          }
        }
        const MyComponent: React.FC<Props> = ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        type Props = {
          name: string,
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
			Output: []string{`
        type Props = {
          readonly name: string,
        }

        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code: `
        interface Props {
          name: string;
        }
        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
			Output: []string{`
        interface Props {
          readonly name: string;
        }
        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code: `
        import React from "react";
        type Props = {
          name: string[];
        }
        const MyComponent: React.FC<Props> = async ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `,
			Output: []string{`
        import React from "react";
        type Props = {
          readonly name: string[];
        }
        const MyComponent: React.FC<Props> = async ({ name }) => {
          return <div>{name}</div>;
        };
        export default MyComponent;
      `},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'name' should be read-only."}},
		},
		{
			Code: `
        type Props = {
          readonly firstName: string;
          lastName: string;
        }
        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `,
			Output: []string{`
        type Props = {
          readonly firstName: string;
          readonly lastName: string;
        }
        class Hello extends React.Component<Props> {
          render () {
            return <div>Hello {this.props.name}</div>;
          }
        }
      `},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "readOnlyProp", Message: "Prop 'lastName' should be read-only."}},
		},
	})
}
