package forward_ref_uses_ref

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestForwardRefUsesRefUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/forward-ref-uses-ref.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in the forward_ref_uses_ref_extras_test.go file.
func TestForwardRefUsesRefUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForwardRefUsesRefRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{Code: `
        import { forwardRef } from 'react'
        forwardRef((props, ref) => {
          return null;
        });
      `, Tsx: true},
		{Code: `
        import { forwardRef } from 'react'
        forwardRef((props, ref) => null);
      `, Tsx: true},
		{Code: `
        import { forwardRef } from 'react'
        forwardRef(function (props, ref) {
          return null;
        });
      `, Tsx: true},
		{Code: `
        import { forwardRef } from 'react'
        forwardRef(function Component(props, ref) {
          return null;
        });
      `, Tsx: true},
		{Code: `
        import * as React from 'react'
        React.forwardRef((props, ref) => {
          return null;
        });
      `, Tsx: true},
		{Code: `
        import * as React from 'react'
        React.forwardRef((props, ref) => null);
      `, Tsx: true},
		{Code: `
        import * as React from 'react'
        React.forwardRef(function (props, ref) {
          return null;
        });
      `, Tsx: true},
		{Code: `
        import * as React from 'react'
        React.forwardRef(function Component(props, ref) {
          return null;
        });
      `, Tsx: true},
		{Code: `
        import * as React from 'react'
        function Component(props) {
          return null;
        };
      `, Tsx: true},
		{Code: `
        import * as React from 'react'
        (props) => null;
      `, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `
        import { forwardRef } from 'react'
        forwardRef((props) => {
          return null;
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(3, 20,
				`
        import { forwardRef } from 'react'
        forwardRef((props, ref) => {
          return null;
        });
      `,
				`
        import { forwardRef } from 'react'
        (props) => {
          return null;
        };
      `)},
			Tsx: true,
		},
		{
			Code: `
        import { forwardRef } from 'react'
        forwardRef(props => {
          return null;
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(3, 20,
				`
        import { forwardRef } from 'react'
        forwardRef((props, ref) => {
          return null;
        });
      `,
				`
        import { forwardRef } from 'react'
        props => {
          return null;
        };
      `)},
			Tsx: true,
		},
		{
			Code: `
        import * as React from 'react'
        React.forwardRef((props) => null);
      `,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(3, 26,
				`
        import * as React from 'react'
        React.forwardRef((props, ref) => null);
      `,
				`
        import * as React from 'react'
        (props) => null;
      `)},
			Tsx: true,
		},
		{
			Code: `
        import { forwardRef } from 'react'
        const Component = forwardRef(function (props) {
          return null;
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(3, 38,
				`
        import { forwardRef } from 'react'
        const Component = forwardRef(function (props, ref) {
          return null;
        });
      `,
				`
        import { forwardRef } from 'react'
        const Component = function (props) {
          return null;
        };
      `)},
			Tsx: true,
		},
		{
			Code: `
        import * as React from 'react'
        React.forwardRef(function Component(props) {
          return null;
        });
      `,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(3, 26,
				`
        import * as React from 'react'
        React.forwardRef(function Component(props, ref) {
          return null;
        });
      `,
				`
        import * as React from 'react'
        function Component(props) {
          return null;
        };
      `)},
			Tsx: true,
		},
	})
}

func missingRefError(line int, column int, addOutput string, removeOutput string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "missingRefParameter",
		Message:   missingRefParameterText,
		Line:      line,
		Column:    column,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{
				MessageId: "addRefParameter",
				Output:    addOutput,
			},
			{
				MessageId: "removeForwardRef",
				Output:    removeOutput,
			},
		},
	}
}
