// This file contains rslint-specific AST-shape, real-user, and upstream-branch
// lock-in tests. Upstream-migrated cases live in hook_use_state_upstream_test.go.
package hook_use_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestHookUseStateExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HookUseStateRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized receiver and callee ----
		{Code: `import React from 'react'; const [value, setValue] = (React).useState()`, Tsx: true},
		{Code: `import { useState } from 'react'; const [value, setValue] = (useState)()`, Tsx: true},
		// ESTree elides parentheses around the call expression itself too.
		{Code: `import { useState } from 'react'; const [value, setValue] = (useState())`, Tsx: true},
		// ---- Dimension 4: element access does not match the identifier-property gate ----
		{Code: `import React from 'react'; const result = React['useState']()`, Tsx: true},
		// ---- Dimension 4: TS wrappers are explicit and remain non-destructured ----
		{Code: `import { useState } from 'react'; const [value, setValue] = useState<number>()`, Tsx: true},
		// N/A: declarations, classes, and nesting are not inspected by this call/declarator rule.
		// ---- Dimension 4: empty binding pattern degrades without a panic ----
		{Code: `import { useState } from 'react'; const [value, setValue] = useState()`, Tsx: true},
		// ---- Real-user: aliased React default import ----
		{Code: `import R from 'react'; const [data, setData] = R.useState({})`, Tsx: true},
		// ---- Real-user: an alias imported from react ----
		{Code: `import { useState as useStore } from 'react'; const [data, setData] = useStore({})`, Tsx: true},
		// Components#isReactHookCall does not recognize renamed named imports.
		{Code: `import { useState as state } from 'react'; const [color] = state(value)`, Tsx: true},
		// Components sees only the first default React import.
		{Code: `import A from 'react'; import B from 'react'; const result = B.useState()`, Tsx: true},
		// Components has not seen an import declared after the call yet.
		{Code: `const result = useState(); import { useState } from 'react'`, Tsx: true},
		{Code: `const result = React.useState(); import React from 'react'`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// Locks in upstream CallExpression arm 1: an immediate return is ignored; non-return expression reports.
		hookUseStateError(`import { useState } from 'react';
useState()`, useStateErrorText, 2, 1),
		// A hook-shaped local alias remains recognized upstream.
		hookUseStateError(`import { useState as useStore } from 'react';
const result = useStore()`, useStateErrorText, 2, 16),
		// An optional member access is a ChainExpression upstream.
		hookUseStateError(`import React from 'react';
const [value, setValue] = React?.useState()`, useStateErrorText, 2, 27),
		// Locks in upstream symmetric-pair arm: a transposed acronym setter is rejected.
		{
			Code: `import { useState } from 'react';
const [rgb, setRBG] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [rgb, setRgb] = useState()`}}}}, Tsx: true,
		},
		// Locks in upstream allowDestructuredState arm: setter destructuring is never allowed.
		{
			Code: `import { useState } from 'react';
const [value, {setValue}] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [value, setValue] = useState()`}}}}, Tsx: true,
		},
		// Optional calls are still React hook calls upstream.
		hookUseStateError(`import { useState } from 'react';
const [value, setValue] = useState?.()`, useStateErrorText, 2, 27),
		// ESTree represents either defaulted binding element as an AssignmentPattern.
		hookUseStateError(`import { useState } from 'react';
const [value = initialValue, setValue] = useState()`, useStateErrorText, 2, 7),
		{
			Code: `import { useState } from 'react';
const [value, setValue = initialSetter] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [value, setValue] = useState()`}},
			}}, Tsx: true,
		},
		// ---- Dimension 4: nested array value has option-specific diagnostic ----
		hookUseStateError(`import { useState } from 'react';
const [[first], setFirst] = useState([1])`, destructuredStateErrorText, 2, 7),
		// Components chooses the first default React import for the memo suggestion.
		{
			Code: `import A from 'react'; import B from 'react'; import { useState } from 'react'; const [color] = useState(value)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import A from 'react'; import B from 'react'; import { useState, useMemo } from 'react'; const color = A.useMemo(() => value, [])`},
					{MessageId: "suggestPair", Output: `import A from 'react'; import B from 'react'; import { useState } from 'react'; const [color, setColor] = useState(value)`},
				},
			}},
		},
		// Upstream inserts useMemo even when the default React import is used.
		{
			Code: `import React, { useState } from 'react'; const [color] = useState(value)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import React, { useState, useMemo } from 'react'; const color = React.useMemo(() => value, [])`},
					{MessageId: "suggestPair", Output: `import React, { useState } from 'react'; const [color, setColor] = useState(value)`},
				},
			}},
		},
	})
}

func hookUseStateError(code, message string, line, column int) rule_tester.InvalidTestCase {
	messageID := "useStateErrorMessage"
	if message == destructuredStateErrorText {
		messageID = "useStateErrorMessageOrAddOption"
	}
	return rule_tester.InvalidTestCase{
		Code: code,
		Tsx:  true,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: messageID,
			Message:   message,
			Line:      line,
			Column:    column,
		}},
	}
}
