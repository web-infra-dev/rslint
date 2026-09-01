// This file migrates the upstream eslint-plugin-react v7.37.5 hook-use-state
// suite. rslint-specific shape and branch tests live in hook_use_state_extras_test.go.
package hook_use_state

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestHookUseStateUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &HookUseStateRule, []rule_tester.ValidTestCase{
		{Code: `import { useState } from 'react'; const [color, setColor] = useState()`, Tsx: true},
		{Code: `import { useState } from 'react'; const [rgb, setRGB] = useState()`, Tsx: true},
		{Code: `import { useState } from 'react'; const [rgbValue, setRGBValue] = useState()`, Tsx: true},
		{Code: `import { useState } from 'react'; const [customColorValue, setCustomColorValue] = useState()`, Tsx: true},
		{Code: `import { useState } from 'react'; const [color, setColor] = useState('#ffffff')`, Tsx: true},
		{Code: `import { useState } from 'react'; const [color1, setColor1] = useState()`, Tsx: true},
		{Code: `import React from 'react'; const [color, setColor] = React.useState()`, Tsx: true},
		{Code: `import React from 'react'; import useState from 'some-other-use-state'; const [color, setFlavor] = useState()`, Tsx: true},
		{Code: `import { useRef } from 'react'; const result = useState()`, Tsx: true},
		{Code: `const result = React.useState()`, Tsx: true},
		{Code: `useState()`, Tsx: true},
		{Code: `const result = useState()`, Tsx: true},
		{Code: `const [color, setFlavor] = useState()`, Tsx: true},
		{Code: `import { useState as alternative } from 'react'; const [color, setColor] = alternative()`, Tsx: true},
		{Code: `import { useState } from 'react'; function useColor() { return useState() }`, Tsx: true},
		{Code: `import { useState } from 'react'; function useColor() { function useState() {} const result = useState() }`, Tsx: true},
		{Code: `import React from 'react'; function useColor() { const React = {useState(){}}; const result = React.useState() }`, Tsx: true},
		{Code: `import { useState } from 'react'; const [color, setColor] = useState<string>()`, Tsx: true},
		{Code: `import { useState } from 'react'; const [color, setColor] = useState<string>('#ffffff')`, Tsx: true},
		{Code: `import { useState } from 'react'; const [{foo}, setFoo] = useState({foo: 1})`, Tsx: true, Options: []any{map[string]any{"allowDestructuredState": true}}},
		{Code: `import { useState } from 'react'; const [[index], setIndex] = useState([0])`, Tsx: true, Options: []any{map[string]any{"allowDestructuredState": true}}},
		{Code: `import { useState } from 'react'; const [{foo, bar, baz}, setFooBarBaz] = useState({foo: 'bbb', bar: 'aaa', baz: 'qqq'})`, Tsx: true, Options: []any{map[string]any{"allowDestructuredState": true}}},
		{Code: `import { useState } from 'react'; const [[index, value], setValueWithIndex] = useState([0, 'hello'])`, Tsx: true, Options: []any{map[string]any{"allowDestructuredState": true}}},
	}, []rule_tester.InvalidTestCase{
		hookUseStateError(`import { useState } from 'react';
const result = useState()`, useStateErrorText, 2, 16),
		hookUseStateError(`import { useState as alternative } from 'react';
const result = alternative()`, useStateErrorText, 2, 16),
		hookUseStateError(`import { useState } from 'react';
function useColor() { const result = useState(); return result }`, useStateErrorText, 2, 38),
		hookUseStateError(`import { useState as alternative } from 'react';
function useColor() { const result = alternative(); return result }`, useStateErrorText, 2, 38),
		hookUseStateError(`import React from 'react';
const result = React.useState()`, useStateErrorText, 2, 16),
		hookUseStateError(`import React from 'react';
function useColor() { const result = React.useState(); return result }`, useStateErrorText, 2, 38),
		hookUseStateError(`import ReactAlternative from 'react';
const result = ReactAlternative.useState()`, useStateErrorText, 2, 16),
		hookUseStateError(`import ReactAlternative from 'react';
function useColor() { const result = ReactAlternative.useState(); return result }`, useStateErrorText, 2, 38),
		hookUseStateError(`import { useState } from 'react';
const [, setColor] = useState()`, useStateErrorText, 2, 7),
		hookUseStateError(`import { useState } from 'react';
const { color } = useState()`, useStateErrorText, 2, 19),
		hookUseStateError(`import { useState } from 'react';
const [] = useState()`, useStateErrorText, 2, 7),
		hookUseStateError(`import { useState } from 'react';
const [, , extra] = useState()`, useStateErrorText, 2, 7),
		hookUseStateError(`import { useState } from 'react';
const [, , , ,] = useState()`, useStateErrorText, 2, 7),
		hookUseStateError(`import { useState } from 'react';
const [, makeColor] = useState()`, useStateErrorText, 2, 7),
		{
			Code: `import { useState } from 'react';
const [color] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState()`}},
			}}, Tsx: true,
		},
		{
			Code: `import { useState } from 'react';
const [color] = useState(initialColor)`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import { useState, useMemo } from 'react';
const color = useMemo(() => initialColor, [])`},
					{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState(initialColor)`},
				},
			}}, Tsx: true,
		},
		{
			Code: `import { useState, useMemo as memo } from 'react';
const [color] = useState(initialColor)`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import { useState, useMemo as memo } from 'react';
const color = memo(() => initialColor, [])`},
					{MessageId: "suggestPair", Output: `import { useState, useMemo as memo } from 'react';
const [color, setColor] = useState(initialColor)`},
				},
			}}, Tsx: true,
		},
		{
			Code: `import React from 'react';
const [color] = React.useState(initialColor)`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{
					{MessageId: "suggestMemo", Output: `import React from 'react';
const color = React.useMemo(() => initialColor, [])`},
					{MessageId: "suggestPair", Output: `import React from 'react';
const [color, setColor] = React.useState(initialColor)`},
				},
			}}, Tsx: true,
		},
		{
			Code: `import { useState } from 'react';
const [color, , extra] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState()`}},
			}}, Tsx: true,
		},
		{
			Code: `import { useState } from 'react';
const [color, setColor, extra1, extra2, extra3] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState()`}}}}, Tsx: true,
		},
		{
			Code: `import { useState } from 'react';
const [color, setFlavor] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState()`}}}}, Tsx: true,
		},
		{
			Code: `import { useState } from 'react';
const [color, setFlavor, extra] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState()`}}}}, Tsx: true,
		},
		hookUseStateError(`import { useState } from 'react';
const [{foo}, setFoo] = useState({foo: 1})`, destructuredStateErrorText, 2, 7),
		{
			Code: `import { useState } from 'react';
const [{foo}, {setFoo}] = useState({foo: 1})`,
			Options: []any{map[string]any{"allowDestructuredState": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7}},
			Tsx:     true,
		},
		hookUseStateError(`import { useState } from 'react';
const [{foo}, {setFoo}] = useState({foo: 1})`, useStateErrorText, 2, 7),
		{
			Code: `import { useState } from 'react';
const [color, setFlavor] = useState<string>()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState<string>()`}}}}, Tsx: true,
		},
		{
			Code: `import { useState } from 'react';
function useColor() { const [color, setFlavor] = useState<string>('#ffffff'); return [color, setFlavor] }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
function useColor() { const [color, setColor] = useState<string>('#ffffff'); return [color, setFlavor] }`}}}}, Tsx: true,
		},
		{
			Code: `import React from 'react';
function useColor() { const [color, setFlavor] = React.useState<string>('#ffffff'); return [color, setFlavor] }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import React from 'react';
function useColor() { const [color, setColor] = React.useState<string>('#ffffff'); return [color, setFlavor] }`}}}}, Tsx: true,
		},
	})
}

func hookUseStateError(code, message string, line, column int) rule_tester.InvalidTestCase {
	messageID := "useStateErrorMessage"
	if message == destructuredStateErrorText {
		messageID = "useStateErrorMessageOrAddOption"
	}
	return rule_tester.InvalidTestCase{Code: code, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: messageID, Message: message, Line: line, Column: column}}}
}
