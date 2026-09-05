// This file contains rslint-specific AST-shape, real-user, and upstream-branch
// lock-in tests. Upstream-migrated cases live in hook_use_state_upstream_test.go.
package hook_use_state

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestHookUseStateEditDemand(t *testing.T) {
	const code = `import {useState} from 'react'; const [value]: [number] = useState(initial);`
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.tsx", Path: "/test.tsx",
	}, code, core.ScriptKindTSX)
	for _, testCase := range []struct {
		name   string
		demand rule.EditDemand
		count  int
	}{
		{"diagnostics", 0, 0},
		{"autofix", rule.EditDemandAutofix, 0},
		{"suggestions", rule.EditDemandSuggestion, 2},
		{"all", rule.EditDemandAll, 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var diagnostics []rule.RuleDiagnostic
			ctx := (rule.RuleContext{SourceFile: file}).WithDiagnosticConsumer(HookUseStateRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
				Demand: testCase.demand,
				Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) },
			})
			listeners := HookUseStateRule.Run(ctx, nil)
			utils.VisitDescendants(file.AsNode(), func(node *ast.Node) bool {
				if listener := listeners[node.Kind]; listener != nil {
					listener(node)
				}
				return true
			})
			if len(diagnostics) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(diagnostics))
			}
			diagnostic := diagnostics[0]
			if got := code[diagnostic.Range.Pos():diagnostic.Range.End()]; got != "[value]: [number]" {
				t.Errorf("diagnostic covers %q, want the typed pattern", got)
			}
			if diagnostic.FixesPtr != nil {
				t.Error("suggestions must not become automatic fixes")
			}
			count := 0
			if diagnostic.Suggestions != nil {
				count = len(*diagnostic.Suggestions)
			}
			if count != testCase.count {
				t.Errorf("got %d suggestions, want %d", count, testCase.count)
			}
		})
	}
}

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

// Verified against eslint-plugin-react 7.37.5 and master c99d3b27 with
// @typescript-eslint/parser 8.69.0, including complete suggestion output.
func TestHookUseStateCompatibility(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HookUseStateRule, []rule_tester.ValidTestCase{
		{Code: "import useOther, { useState } from 'react'; useState; const result=useOther();", Tsx: true},
		{Code: "import React, { useMemo as useState } from 'react'; const result = React.useState();", Tsx: true},
		{Code: "import {useState as use9} from 'react';const result=use9();", Tsx: true},
		{Code: "import {use9 as useState} from 'react';const result=useState();", Tsx: true},
		{Code: "import {useEffect} from 'react'; type T<A>=typeof useEffect; const result=useState();", Tsx: true},
		{Code: "import {useEffect} from 'react'; interface T<A> {x:typeof useEffect}; const result=useState();", Tsx: true},
		{Code: "import {useEffect} from 'react'; type T = { [K in keyof typeof useEffect]: K }; const result=useState();", Tsx: true},
		{Code: "import {useEffect} from 'react'; type T = () => typeof useEffect; const result=useState();", Tsx: true},
		{Code: "import {useEffect} from 'react'; type T = typeof useEffect extends any ? 1 : 2; const result=useState();", Tsx: true},
		{Code: "import {useEffect} from 'react'; type T = any extends any ? typeof useEffect : 2; const result=useState();", Tsx: true},
		{Code: "import {useEffect} from 'react'; type T<X> = typeof useEffect extends X ? 1 : 2; const result=useState();", Tsx: true},
		{Code: "import React from 'react'; const [x,setX]=React[useState]();", Tsx: true},
		{Code: "import React from 'react'; function f() {return React[useState]();}", Tsx: true},
		{Code: "import React, { useState, useEffect } from 'react';\nconst result=(React?.useState)();", Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "import React from 'react'; const result = React[useState]();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 43, EndLine: 1, EndColumn: 60,
				},
			},
		},
		{
			Code: "import React, { useState as useStore } from 'react'; const result = React.useStore();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 69, EndLine: 1, EndColumn: 85,
				},
			},
		},
		{
			Code: "import React from 'react'; const result=React[(useState)]();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 41, EndLine: 1, EndColumn: 60,
				},
			},
		},
		{
			Code: "import React from 'react'; const [x]=React[useState](initial);", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 34, EndLine: 1, EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import React from 'react'; const x=React.useMemo(() => initial, []);"},
						{MessageId: "suggestPair", Output: "import React from 'react'; const [x, setX]=React[useState](initial);"},
					},
				},
			},
		},
		{
			Code: "import React from 'react'; const [x,setX]=React?.[useState]();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 43, EndLine: 1, EndColumn: 62,
				},
			},
		},
		{
			Code: "import { useState } from 'react'; const [x]: [number] = useState(initial);", Tsx: true,
			Options: map[string]any{"allowDestructuredState": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 41, EndLine: 1, EndColumn: 54,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import { useState, useMemo } from 'react'; const x = useMemo(() => initial, []);"},
						{MessageId: "suggestPair", Output: "import { useState } from 'react'; const [x, setX] = useState(initial);"},
					},
				},
			},
		},
		{
			Code: "import {useState} from 'react'; const [x,wrong] /*note*/ : [number,any] = useState(initial);", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 39, EndLine: 1, EndColumn: 72,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestPair", Output: "import {useState} from 'react'; const [x, setX] = useState(initial);"},
					},
				},
			},
		},
		{
			Code: "import {useState} from 'react'; const [{x},setX]:[any,any]=useState(initial);", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessageOrAddOption", Message: destructuredStateErrorText,
					Line: 1, Column: 39, EndLine: 1, EndColumn: 59,
				},
			},
		},
		{
			Code: "import {useEffect} from 'react'; type T=typeof useEffect; const result=useState();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 72, EndLine: 1, EndColumn: 82,
				},
			},
		},
		{
			Code: "import {useEffect} from 'react'; interface T {x:typeof useEffect}; const result=useState();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 81, EndLine: 1, EndColumn: 91,
				},
			},
		},
		{
			Code: "import {useEffect} from 'react'; function f() { type T=typeof useEffect; const result=useState(); }", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 87, EndLine: 1, EndColumn: 97,
				},
			},
		},
		{
			Code: "declare module 'custom' { import React from 'react'; const result=React.useState(); }", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 67, EndLine: 1, EndColumn: 83,
				},
			},
		},
		{
			Code: "declare module 'custom' { import {useState as useStore} from 'react'; const result=useStore(); }", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 84, EndLine: 1, EndColumn: 94,
				},
			},
		},
		{
			Code: "import React from 'react'; declare module 'custom' { import React from 'other'; const result=React.useState(); }", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 94, EndLine: 1, EndColumn: 110,
				},
			},
		},
		{
			Code: "import {useState} from 'react'; useState(); declare module 'custom' { import {useState as useStore} from 'react'; const result=useStore(); }", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 33, EndLine: 1, EndColumn: 43,
				},
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 128, EndLine: 1, EndColumn: 138,
				},
			},
		},
		{
			Code: "import {useState} from 'react'; useState(); import {useState as useStore} from 'react'; const result=useStore();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 33, EndLine: 1, EndColumn: 43,
				},
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 102, EndLine: 1, EndColumn: 112,
				},
			},
		},
		{
			Code: "import React from 'react'; React.useState(); declare module 'custom' { import {useState} from 'react'; const result=useState(); }", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 28, EndLine: 1, EndColumn: 44,
				},
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 117, EndLine: 1, EndColumn: 127,
				},
			},
		},
		{
			Code: "import {useState} from 'react'; const result=useState(); import React from 'react'; const result2=React.useState();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 46, EndLine: 1, EndColumn: 56,
				},
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 99, EndLine: 1, EndColumn: 115,
				},
			},
		},
		{
			Code: "import {useState as useStore} from 'react'; useStore(); declare module 'custom' { import {useEffect as useStore} from 'react'; const result=useStore(); }", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 45, EndLine: 1, EndColumn: 55,
				},
			},
		},
		{
			Code: "import {useEffect} from 'react'; type T = any extends any ? 1 : typeof useEffect; const result=useState();", Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 96, EndLine: 1, EndColumn: 106,
				},
			},
		},
	})
}

// JSDoc wrappers must be exercised as JavaScript so tsgo reparses them.
func TestHookUseStateJSDoc(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &HookUseStateRule, []rule_tester.ValidTestCase{
		{Code: "import React, {useState, useEffect} from 'react'; const [x,setX] = /** @type {any} */ (useState(initial));", FileName: "audit.js"},
		{Code: "import React, {useState, useEffect} from 'react'; function f() { return /** @satisfies {any} */ (useState(initial)); }", FileName: "audit.js"},
		{Code: "import React, {useState, useEffect} from 'react'; const result = (/** @type {any} */ (React?.useState))(initial);", FileName: "audit.js"},
		{Code: "import React, {useState, useEffect} from 'react'; const result = (/** @satisfies {any} */ (React?.useState))(initial);", FileName: "audit.js"},
		{Code: "import {useEffect} from 'react'; /** @type {typeof useEffect} */ const value = initial; const result = useState();", FileName: "audit.js"},
		{Code: "import {useEffect} from 'react'; /** @param {typeof useEffect} value */ function f(value) { const result = useState(); }", FileName: "audit.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "import React, {useState, useEffect} from 'react'; const [x] = (/** @type {any} */ (useState))(initial);", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 57, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import React, {useState, useMemo, useEffect} from 'react'; const x = React.useMemo(() => initial, []);"},
						{MessageId: "suggestPair", Output: "import React, {useState, useEffect} from 'react'; const [x, setX] = (/** @type {any} */ (useState))(initial);"},
					},
				},
			}},
		{Code: "import React, {useState, useEffect} from 'react'; const result = (/** @satisfies {any} */ (useState))(initial);", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 66, EndLine: 1, EndColumn: 111,
				},
			}},
		{Code: "import React, {useState, useEffect} from 'react'; const [x] = (/** @type {any} */ (React.useState))(initial);", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 57, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import React, {useState, useMemo, useEffect} from 'react'; const x = React.useMemo(() => initial, []);"},
						{MessageId: "suggestPair", Output: "import React, {useState, useEffect} from 'react'; const [x, setX] = (/** @type {any} */ (React.useState))(initial);"},
					},
				},
			}},
		{Code: "import React, {useState, useEffect} from 'react'; const result = (/** @type {any} */ (React).useState)(initial);", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 66, EndLine: 1, EndColumn: 112,
				},
			}},
		{Code: "import React, {useState, useEffect} from 'react'; const [x] = (React[/** @type {any} */ (useState)])(initial);", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 57, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import React, {useState, useMemo, useEffect} from 'react'; const x = React.useMemo(() => initial, []);"},
						{MessageId: "suggestPair", Output: "import React, {useState, useEffect} from 'react'; const [x, setX] = (React[/** @type {any} */ (useState)])(initial);"},
					},
				},
			}},
		{Code: "import React, {useState, useEffect} from 'react'; const [x] = /** @type {any} */ (useState(initial));", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 57, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import React, {useState, useMemo, useEffect} from 'react'; const x = /** @type {any} */ (React.useMemo(() => initial, []));"},
						{MessageId: "suggestPair", Output: "import React, {useState, useEffect} from 'react'; const [x, setX] = /** @type {any} */ (useState(initial));"},
					},
				},
			}},
		{Code: "import React, {useState, useEffect} from 'react'; const [x] = /** @satisfies {any} */ (React.useState(initial));", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 57, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import React, {useState, useMemo, useEffect} from 'react'; const x = /** @satisfies {any} */ (React.useMemo(() => initial, []));"},
						{MessageId: "suggestPair", Output: "import React, {useState, useEffect} from 'react'; const [x, setX] = /** @satisfies {any} */ (React.useState(initial));"},
					},
				},
			}},
		{Code: "import React, {useState, useEffect} from 'react'; const [x] = useState(/** @type {any} */ (initial));", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 57, EndLine: 1, EndColumn: 60,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import React, {useState, useMemo, useEffect} from 'react'; const x = React.useMemo(() => initial, []);"},
						{MessageId: "suggestPair", Output: "import React, {useState, useEffect} from 'react'; const [x, setX] = useState(/** @type {any} */ (initial));"},
					},
				},
			}},
		{Code: "import {useState} from 'react'; /** @type {[number]} */ const [x] = useState(initial);", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 63, EndLine: 1, EndColumn: 66,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestMemo", Output: "import {useState, useMemo} from 'react'; /** @type {[number]} */ const x = useMemo(() => initial, []);"},
						{MessageId: "suggestPair", Output: "import {useState} from 'react'; /** @type {[number]} */ const [x, setX] = useState(initial);"},
					},
				},
			}},
		{Code: "import {useState} from 'react'; /** @type {[number, any]} */ const [x,wrong] = useState(initial);", FileName: "audit.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useStateErrorMessage", Message: useStateErrorText,
					Line: 1, Column: 68, EndLine: 1, EndColumn: 77,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestPair", Output: "import {useState} from 'react'; /** @type {[number, any]} */ const [x, setX] = useState(initial);"},
					},
				},
			}},
	})
}
