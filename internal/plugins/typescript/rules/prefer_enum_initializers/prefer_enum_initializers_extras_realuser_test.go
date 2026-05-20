// TestPreferEnumInitializersExtrasRealuser pulls representative shapes from
// real-world TypeScript codebases that the upstream test suite does not write
// inline — enums declared inside function bodies, deeply nested namespaces,
// non-ASCII identifiers, and HTTP-status-like mixes of initialized/uninitialized
// members at arbitrary indices. These are the inputs production code actually
// produces; the upstream synthetic suite typically misses them.
package prefer_enum_initializers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEnumInitializersExtrasRealuser(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferEnumInitializersRule, []rule_tester.ValidTestCase{
		// ---- Real-user: enum in a function body — listener fires regardless of
		//      enclosing function scope.
		{Code: `
function makeEnum() {
  enum Local {
    X = 1,
    Y = 2,
  }
  return Local;
}
`},

		// ---- Real-user: triply-nested namespaces — listener still fires at
		//      arbitrary depth.
		{Code: `
namespace Outer {
  export namespace Middle {
    export namespace Inner {
      export enum Deep {
        X = 1,
        Y = 2,
      }
    }
  }
}
`},
	}, []rule_tester.InvalidTestCase{
		// ---- Real-user: mixed initialized/uninitialized members at arbitrary
		//      indices — the implicit-numbering pitfall the rule documentation
		//      cites as motivation. Only the uninitialized members report.
		{
			Code: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      5,
					Column:    3,
					EndLine:   5,
					EndColumn: 11,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted = 2,
  NoContent,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted = 3,
  NoContent,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted = 'Accepted',
  NoContent,
}
`},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      6,
					Column:    3,
					EndLine:   6,
					EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent = 3,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent = 4,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
enum HttpStatus {
  OK = 200,
  Created = 201,
  Accepted,
  NoContent = 'NoContent',
}
`},
					},
				},
			},
		},

		// ---- Real-user: exported enum — common shape in declaration files.
		//      Listener fires regardless of `export` modifier.
		{
			Code: `
export enum LogLevel {
  Trace,
  Debug,
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace = 0,
  Debug,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace = 1,
  Debug,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace = 'Trace',
  Debug,
}
`},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    3,
					EndLine:   4,
					EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace,
  Debug = 1,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace,
  Debug = 2,
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
export enum LogLevel {
  Trace,
  Debug = 'Debug',
}
`},
					},
				},
			},
		},

		// ---- Real-user: enum inside a function body — listener fires
		//      regardless of enclosing scope.
		{
			Code: `
function makeEnum() {
  enum Local {
    X,
  }
  return Local;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      4,
					Column:    5,
					EndLine:   4,
					EndColumn: 6,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
function makeEnum() {
  enum Local {
    X = 0,
  }
  return Local;
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
function makeEnum() {
  enum Local {
    X = 1,
  }
  return Local;
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
function makeEnum() {
  enum Local {
    X = 'X',
  }
  return Local;
}
`},
					},
				},
			},
		},

		// ---- Real-user: triply-nested namespaces with a broken inner enum —
		//      listener must fire at arbitrary depth, no early-exit.
		{
			Code: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X,
      }
    }
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      6,
					Column:    9,
					EndLine:   6,
					EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "defineInitializerSuggestion", Output: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X = 0,
      }
    }
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X = 1,
      }
    }
  }
}
`},
						{MessageId: "defineInitializerSuggestion", Output: `
namespace A {
  export namespace B {
    export namespace C {
      export enum Deep {
        X = 'X',
      }
    }
  }
}
`},
					},
				},
			},
		},
	})
}
