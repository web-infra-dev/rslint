package no_namespace

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoNamespaceRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoNamespaceRule, []rule_tester.ValidTestCase{
		// Global declarations
		{Code: `declare global {}`},
		{Code: `declare module 'foo' {}`},
		
		// With allowDeclarations option
		{
			Code:    `declare module foo {}`,
			Options: map[string]interface{}{"allowDeclarations": true},
		},
		{
			Code:    `declare namespace foo {}`,
			Options: map[string]interface{}{"allowDeclarations": true},
		},
		{
			Code: `
declare global {
  namespace foo {}
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
		},
		{
			Code: `
declare module foo {
  namespace bar {}
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
		},
		{
			Code: `
declare global {
  namespace foo {
    namespace bar {}
  }
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
		},
		{
			Code: `
declare namespace foo {
  namespace bar {
    namespace baz {}
  }
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
		},
		{
			Code: `
export declare namespace foo {
  export namespace bar {
    namespace baz {}
  }
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
		},

		// With allowDefinitionFiles option (default true)
		{
			Code:     `namespace foo {}`,
			Filename: "test.d.ts",
			Options:  map[string]interface{}{"allowDefinitionFiles": true},
		},
		{
			Code:     `module foo {}`,
			Filename: "test.d.ts",
			Options:  map[string]interface{}{"allowDefinitionFiles": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Basic cases
		{
			Code: `module foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `namespace foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},

		// With options explicitly false
		{
			Code:    `module foo {}`,
			Options: map[string]interface{}{"allowDeclarations": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    `namespace foo {}`,
			Options: map[string]interface{}{"allowDeclarations": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},

		// With allowDeclarations true but not declared
		{
			Code:    `module foo {}`,
			Options: map[string]interface{}{"allowDeclarations": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    `namespace foo {}`,
			Options: map[string]interface{}{"allowDeclarations": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Declare but allowDeclarations false
		{
			Code: `declare module foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `declare namespace foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    `declare module foo {}`,
			Options: map[string]interface{}{"allowDeclarations": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:    `declare namespace foo {}`,
			Options: map[string]interface{}{"allowDeclarations": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Definition files with allowDefinitionFiles false
		{
			Code:     `namespace foo {}`,
			Filename: "test.d.ts",
			Options:  map[string]interface{}{"allowDefinitionFiles": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:     `module foo {}`,
			Filename: "test.d.ts",
			Options:  map[string]interface{}{"allowDefinitionFiles": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:     `declare module foo {}`,
			Filename: "test.d.ts",
			Options:  map[string]interface{}{"allowDefinitionFiles": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code:     `declare namespace foo {}`,
			Filename: "test.d.ts",
			Options:  map[string]interface{}{"allowDefinitionFiles": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},

		// Nested namespaces
		{
			Code:    `namespace Foo.Bar {}`,
			Options: map[string]interface{}{"allowDeclarations": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `
namespace Foo.Bar {
  namespace Baz.Bas {
    interface X {}
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      2,
					Column:    1,
				},
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      3,
					Column:    3,
				},
			},
		},

		// Complex nested scenarios with allowDeclarations
		{
			Code: `
namespace A {
  namespace B {
    declare namespace C {}
  }
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      2,
					Column:    1,
				},
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
namespace A {
  namespace B {
    export declare namespace C {}
  }
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      2,
					Column:    1,
				},
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `
namespace A {
  declare namespace B {
    namespace C {}
  }
}`,
			Options: map[string]interface{}{"allowDeclarations": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "moduleSyntaxIsPreferred",
					Line:      2,
					Column:    1,
				},
			},
		},
	})
}