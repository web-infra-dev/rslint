package void_dom_elements_no_children

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestVoidDomElementsNoChildrenRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &VoidDomElementsNoChildrenRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div>Children</div>`,
			Tsx:  true,
		},
		{
			Code: `var x = <img src="foo.png" />`,
			Tsx:  true,
		},
		{
			Code: `var x = <br />`,
			Tsx:  true,
		},
		{
			Code: `var x = <input type="text" />`,
			Tsx:  true,
		},
		{
			Code: `var x = <Component children="test" />`,
			Tsx:  true,
		},
		{
			Code: `React.createElement('div')`,
			Tsx:  true,
		},
		{
			Code: `React.createElement('div', { children: 'test' })`,
			Tsx:  true,
		},
		{
			Code: `React.createElement('br')`,
			Tsx:  true,
		},
		{
			// div with dangerouslySetInnerHTML is valid (non-void element)
			Code: `var x = <div dangerouslySetInnerHTML={{ __html: "Foo" }} />`,
			Tsx:  true,
		},
		{
			// React.createElement with no args
			Code: `React.createElement()`,
			Tsx:  true,
		},
		{
			// Non-void createElement with children prop
			Code: `React.createElement('div', { children: 'test' })`,
			Tsx:  true,
		},
		{
			// Void element in createElement but no props object (just element name)
			Code: `React.createElement('img')`,
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `var x = <br>text</br>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			Code: `var x = <img children="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			Code: `var x = <hr dangerouslySetInnerHTML={{ __html: "test" }} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			Code: `var x = <input><span /></input>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			// Third argument with null as second arg (previously missed)
			Code: `React.createElement('br', null, 'text')`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `React.createElement('br', {}, 'Foo')`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `React.createElement('br', { children: 'Foo' })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `React.createElement('img', { dangerouslySetInnerHTML: { __html: 'test' } })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			// Spread plus children prop on void element
			Code: `var x = <img {...props} children="Foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			// Other void elements: area
			Code: `var x = <area children="test" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    9,
				},
			},
		},
		{
			// Other void elements: embed
			Code: `var x = <embed>content</embed>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noChildrenInVoidEl",
					Line:      1,
					Column:    9,
				},
			},
		},
	})
}
