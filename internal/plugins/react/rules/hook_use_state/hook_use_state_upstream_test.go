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
		{Code: `import React from 'react'; const [color, setColor] = React.useState()`, Tsx: true},
		{Code: `useState()`, Tsx: true},
		{Code: `const result = useState()`, Tsx: true},
		{Code: `import { useState as alternative } from 'react'; const [color, setColor] = alternative()`, Tsx: true},
		{Code: `import { useState } from 'react'; function useColor() { return useState() }`, Tsx: true},
		{Code: `import { useState } from 'react'; function useColor() { function useState() {} const result = useState() }`, Tsx: true},
		{Code: `import React from 'react'; function useColor() { const React = {useState(){}}; const result = React.useState() }`, Tsx: true},
		{Code: `import { useState } from 'react'; const [color, setColor] = useState<string>()`, Tsx: true},
		{Code: `import { useState } from 'react'; const [{foo}, setFoo] = useState({foo: 1})`, Tsx: true, Options: []any{map[string]any{"allowDestructuredState": true}}},
		{Code: `import { useState } from 'react'; const [[index], setIndex] = useState([0])`, Tsx: true, Options: []any{map[string]any{"allowDestructuredState": true}}},
	}, []rule_tester.InvalidTestCase{
		hookUseStateError(`import { useState } from 'react';
const result = useState()`, useStateErrorText, 2, 16),
		hookUseStateError(`import { useState as alternative } from 'react';
const result = alternative()`, useStateErrorText, 2, 16),
		hookUseStateError(`import React from 'react';
const result = React.useState()`, useStateErrorText, 2, 16),
		hookUseStateError(`import { useState } from 'react';
const [, setColor] = useState()`, useStateErrorText, 2, 7),
		hookUseStateError(`import { useState } from 'react';
const { color } = useState()`, useStateErrorText, 2, 19),
		hookUseStateError(`import { useState } from 'react';
const [] = useState()`, useStateErrorText, 2, 7),
		hookUseStateError(`import { useState } from 'react';
const [, , extra] = useState()`, useStateErrorText, 2, 7),
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
			Code: `import { useState } from 'react';
const [color, setFlavor, extra] = useState()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState()`}}}}, Tsx: true,
		},
		hookUseStateError(`import { useState } from 'react';
const [{foo}, setFoo] = useState({foo: 1})`, destructuredStateErrorText, 2, 7),
		hookUseStateError(`import { useState } from 'react';
const [{foo}, {setFoo}] = useState({foo: 1})`, useStateErrorText, 2, 7),
		{
			Code: `import { useState } from 'react';
const [color, setFlavor] = useState<string>()`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useStateErrorMessage", Message: useStateErrorText, Line: 2, Column: 7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestPair", Output: `import { useState } from 'react';
const [color, setColor] = useState<string>()`}}}}, Tsx: true,
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
