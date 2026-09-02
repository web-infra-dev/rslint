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
		// ESTree keeps an optional member callee in ChainExpression when it is
		// parenthesized, so the upstream hook matcher does not recognize it.
		{Code: `import React from 'react'; const result = (React?.useState)()`, Tsx: true},
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
		// The imported hook name controls matching even when its local alias is
		// named useState.
		{Code: `import { useFoo as useState } from 'react'; const result = useState()`, Tsx: true},
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
		// Direct optional member calls are also React hook calls upstream.
		hookUseStateError(`import React from 'react';
const result = React?.useState()`, useStateErrorText, 2, 16),
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
		// Components#detect keeps the first matching named import for insertion.
		{
			Code: `import { useState as useFirst, useState as useSecond } from 'react'; const [value] = useSecond(initial)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import { useState as useFirst, useMemo, useState as useSecond } from 'react'; const value = useMemo(() => initial, [])`},
					{MessageId: "suggestPair", Output: `import { useState as useFirst, useState as useSecond } from 'react'; const [value, setValue] = useSecond(initial)`},
				},
			}},
		},
		// Components#detect also keeps the first matching useMemo alias.
		{
			Code: `import { useState, useMemo as memo1, useMemo as memo2 } from 'react'; const [value] = useState(initial)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import { useState, useMemo as memo1, useMemo as memo2 } from 'react'; const value = memo1(() => initial, [])`},
					{MessageId: "suggestPair", Output: `import { useState, useMemo as memo1, useMemo as memo2 } from 'react'; const [value, setValue] = useState(initial)`},
				},
			}},
		},
		// ESTree drops redundant argument parentheses before getText is called.
		{
			Code: `import { useState } from 'react'; const [value] = useState((initial))`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import { useState, useMemo } from 'react'; const value = useMemo(() => initial, [])`},
					{MessageId: "suggestPair", Output: `import { useState } from 'react'; const [value, setValue] = useState((initial))`},
				},
			}},
		},
		// Components selects the first hook-shaped reference in the current scope,
		// even when the current callee is shadowed by a later local declaration.
		hookUseStateError(`import { useFoo, useState } from 'react';
	function f() { useFoo(); function useState() {} const result = useState() }`, useStateErrorText, 2, 65),
		// Components falls back to the bare callee spelling after selecting an
		// earlier hook-shaped reference in the scope.
		hookUseStateError(`import { useEffect } from 'react';
	useEffect();
	const result = useState()`, useStateErrorText, 3, 17),
		// Components uses the first default React reference in the function
		// scope, even when a later body declaration shadows the same spelling.
		{
			Code:   `import React from 'react'; function f(x = React) { const React = {}; const result = React.useState(); }`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText}},
		},
		// An omitted first binding is falsy in the upstream truthiness check and
		// must not receive the single-getter memo suggestion.
		{
			Code:   `import { useState } from 'react'; const [,] = useState(initial)`,
			Tsx:    true,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText}},
		},
		// A single non-identifier binding still receives the upstream memo suggestion.
		{
			Code: `import { useState } from 'react'; const [value = fallback] = useState(initial)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestMemo",
					Output:    `import { useState, useMemo } from 'react'; const undefined = useMemo(() => initial, [])`,
				}},
			}},
		},
		{
			Code: `import { useState } from 'react'; const [...value] = useState(initial)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "suggestMemo",
					Output:    `import { useState, useMemo } from 'react'; const undefined = useMemo(() => initial, [])`,
				}},
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
