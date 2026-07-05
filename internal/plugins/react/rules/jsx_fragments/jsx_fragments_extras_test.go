package jsx_fragments

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestJsxFragmentsExtras locks in behavior not covered by the upstream suite:
// nested JSX, multiple fragments in one file, fragment-like lookalikes,
// declaration resolution, binding patterns, and tsgo-specific AST shapes.
func TestJsxFragmentsExtras(t *testing.T) {
	settings := reactSettings("16.2", "Act", "Frag")
	settingsOld := reactSettings("16.1", "Act", "Frag")

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFragmentsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 3: attributes suppress shorthand conversion ----
		{Code: `<Act.Frag {...props}><Foo /></Act.Frag>`, Tsx: true, Settings: settings},
		{Code: `<Act.Frag key={key} />`, Tsx: true, Settings: settings},

		// ---- Dimension 4: access / key forms that are not React fragments ----
		{Code: `<Act.Other><Foo /></Act.Other>`, Tsx: true, Settings: settings},
		{Code: `<Other.Frag><Foo /></Other.Frag>`, Tsx: true, Settings: settings},
		{Code: `<act.Frag><Foo /></act.Frag>`, Tsx: true, Settings: settings},
		{Code: `<ns:Frag><Foo /></ns:Frag>`, Tsx: true, Settings: settings},
		{Code: "const Frag = Act.Other;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "const Frag = require('react').Frag;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},

		// ---- Dimension 4: import source and imported-name forms ----
		{Code: "import { Other as Frag } from 'react';\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "import { Frag } from 'not-react';\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "import { Fragment } from 'react';\n<Fragment><Foo /></Fragment>;", Tsx: true, Settings: settings},

		// ---- Dimension 2: local shadowing boundary ----
		{
			Code: "const Frag = Act.Frag;\nfunction render(Frag) { return <Frag><Foo /></Frag>; }",
			Tsx:  true, Settings: settings,
		},

		// ---- Dimension 4: TS-only expression wrappers stay opaque ----
		{Code: "const Frag = Other.Frag as any;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "const Frag = Act.Frag as any;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "const Frag = Act.Frag satisfies unknown;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "const Frag = Act.Frag!;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "const Frag = require('react') as any;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "const Frag = require('react' as any);\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},

		// ---- Dimension 4: optional-chain initializers stay opaque ----
		{Code: "const Frag = Act?.Frag;\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},
		{Code: "const Frag = require?.('react');\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},

		// ---- Dimension 4: dynamic initializers should not crash or report ----
		{Code: "const { Frag } = require(moduleName);\n<Frag><Foo /></Frag>;", Tsx: true, Settings: settings},

		// ---- Dimension 4: empty children with attributes remain valid ----
		{Code: `<Act.Frag key="key"></Act.Frag>`, Tsx: true, Settings: settings},

		// N/A: Receiver optional chains do not apply to JSX tag names.
		// N/A: Computed JSX tag names are not legal TSX syntax.
		// N/A: Private identifiers do not apply to JSX tag names.
		// N/A: Function/class container forms are not target nodes for this rule.
		// N/A: Autofix side-effect suppression is not needed; fixes only replace tag tokens.
	}, []rule_tester.InvalidTestCase{
		// ---- Config: bare string option shape ----
		{
			Code:     `<><Foo /></>`,
			Output:   []string{`<Act.Frag><Foo /></Act.Frag>`},
			Tsx:      true,
			Options:  modeElement,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferPragma", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: nested shorthand fragment in JSX expression ----
		{
			Code:     `<Box child={<><Foo /></>} />`,
			Output:   []string{`<Box child={<Act.Frag><Foo /></Act.Frag>} />`},
			Tsx:      true,
			Options:  modeElement,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferPragma", Line: 1, Column: 13},
			},
		},

		// ---- Dimension 2: nested shorthand fragments converge over repeated fixes ----
		{
			Code: "<>\n  <>\n    <Foo />\n  </>\n</>",
			Output: []string{
				"<Act.Frag>\n  <>\n    <Foo />\n  </>\n</Act.Frag>",
				"<Act.Frag>\n  <Act.Frag>\n    <Foo />\n  </Act.Frag>\n</Act.Frag>",
			},
			Tsx:      true,
			Options:  modeElement,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferPragma", Line: 1, Column: 1},
				{MessageId: "preferPragma", Line: 2, Column: 3},
			},
		},

		// ---- Branch: React version gate runs before syntax-mode attribute suppression ----
		{
			Code:     `<Act.Frag key="key"><Foo /></Act.Frag>`,
			Tsx:      true,
			Settings: settingsOld,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "fragmentsNotSupported", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: nested standard fragment inside another JSX container ----
		{
			Code:     `<Box><Act.Frag><Foo /></Act.Frag></Box>`,
			Output:   []string{`<Box><><Foo /></></Box>`},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 1, Column: 6},
			},
		},

		// ---- Dimension 3: multi-line fixer preserves children and comments ----
		{
			Code:     "<Act.Frag>\n  {/* keep */}\n  <Foo />\n</Act.Frag>",
			Output:   []string{"<>\n  {/* keep */}\n  <Foo />\n</>"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 1, Column: 1},
			},
		},

		// ---- Dimension 2: multiple standard fragments in one file ----
		{
			Code:     `const x = [<Act.Frag />, <Act.Frag><Foo /></Act.Frag>];`,
			Output:   []string{`const x = [<></>, <><Foo /></>];`},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 1, Column: 12},
				{MessageId: "preferFragment", Line: 1, Column: 26},
			},
		},

		// ---- Dimension 4: parenthesized initializer matching ----
		{
			Code:     "const Frag = (Act.Frag);\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = (Act.Frag);\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "const Frag = (Act).Frag;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = (Act).Frag;\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "const Frag = (require)('react');\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = (require)('react');\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
		{
			Code:     "const Frag = require(('react'));\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = require(('react'));\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream JSXFragment arm 1: default syntax mode does not report shorthand.
		// The invalid counterpart here uses element mode to prove the branch boundary.
		{
			Code:     `<></>`,
			Output:   []string{`<Act.Frag></Act.Frag>`},
			Tsx:      true,
			Options:  []interface{}{modeElement},
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferPragma", Line: 1, Column: 1},
			},
		},

		// Locks in upstream getFixerToShort() arm 1: paired JSXElement.
		{
			Code:     `<Act.Frag></Act.Frag>`,
			Output:   []string{`<></>`},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 1, Column: 1},
			},
		},

		// Locks in upstream getFixerToShort() arm 2: self-closing JSXElement.
		{
			Code:     `<Act.Frag/>`,
			Output:   []string{`<></>`},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 1, Column: 1},
			},
		},

		// Locks in upstream ImportDeclaration arm: renamed import local is tracked.
		{
			Code:     "import { Frag as LocalFrag } from 'react';\n<LocalFrag><Foo /></LocalFrag>;",
			Output:   []string{"import { Frag as LocalFrag } from 'react';\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream ImportDeclaration arm: import aliases are tracked by text, not shadow-aware scope.
		{
			Code:     "import { Frag } from 'react';\nfunction render(Frag) { return <Frag><Foo /></Frag>; }",
			Output:   []string{"import { Frag } from 'react';\nfunction render(Frag) { return <><Foo /></>; }"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 32},
			},
		},

		// Locks in upstream Program:exit behavior: a later variable declaration still resolves.
		{
			Code:     "<Frag><Foo /></Frag>;\nconst Frag = Act.Frag;",
			Output:   []string{"<><Foo /></>;\nconst Frag = Act.Frag;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 1, Column: 1},
			},
		},

		// Locks in upstream refersToReactFragment() arm 1: variable init is the pragma identifier.
		{
			Code:     "const Frag = Act;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = Act;\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream refersToReactFragment() arm 1 with nested object binding.
		{
			Code:     "const { nested: { Frag } } = Act;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const { nested: { Frag } } = Act;\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream refersToReactFragment() arm 1 with array binding.
		{
			Code:     "const [Frag] = Act;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const [Frag] = Act;\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream refersToReactFragment() arm 2: variable init is pragma.fragment.
		{
			Code:     "const Frag = Act.Frag;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = Act.Frag;\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream refersToReactFragment() arm 3: variable init is require('react').
		{
			Code:     "const Frag = require('react');\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = require('react');\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream refersToReactFragment() arm 3: only the first require argument is inspected.
		{
			Code:     "const Frag = require('react', extra);\n<Frag><Foo /></Frag>;",
			Output:   []string{"const Frag = require('react', extra);\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream destructuring quirk: property name is not checked once init is the pragma.
		{
			Code:     "const { Other: Frag } = Act;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const { Other: Frag } = Act;\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// Locks in upstream destructuring quirk: member init is accepted for a binding element.
		{
			Code:     "const { Other: Frag } = Act.Frag;\n<Frag><Foo /></Frag>;",
			Output:   []string{"const { Other: Frag } = Act.Frag;\n<><Foo /></>;"},
			Tsx:      true,
			Settings: settings,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// ---- Real-user: #2420 asks about enforcing bare <Fragment>; current rule treats imported Fragment as standard-form fragment ----
		{
			Code:   "import { Fragment } from 'react';\n<Fragment><Foo /></Fragment>;",
			Output: []string{"import { Fragment } from 'react';\n<><Foo /></>;"},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},

		// ---- Real-user: #1661 proposed const Fragment = React.Fragment as a supported standard form ----
		{
			Code:   "const Fragment = React.Fragment;\n<Fragment><Foo /></Fragment>;",
			Output: []string{"const Fragment = React.Fragment;\n<><Foo /></>;"},
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferFragment", Line: 2, Column: 1},
			},
		},
	})
}
