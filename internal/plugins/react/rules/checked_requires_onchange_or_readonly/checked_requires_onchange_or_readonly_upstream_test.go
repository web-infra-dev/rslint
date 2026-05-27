// TestCheckedRequiresOnchangeOrReadonlyUpstream migrates the full valid/invalid
// suite from upstream tests/lib/rules/checked-requires-onchange-or-readonly.js
// 1:1. Position assertions cover line/column for every invalid case. rslint-
// specific lock-in cases (edge shapes, branch lock-ins, real-user shapes) live
// in the checked_requires_onchange_or_readonly_extras_test.go file.
package checked_requires_onchange_or_readonly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCheckedRequiresOnchangeOrReadonlyUpstream(t *testing.T) {
	ignoreMissing := map[string]interface{}{"ignoreMissingProperties": true}
	ignoreExclusive := map[string]interface{}{"ignoreExclusiveCheckedAttribute": true}
	bothIgnore := map[string]interface{}{"ignoreMissingProperties": true, "ignoreExclusiveCheckedAttribute": true}
	bothFalse := map[string]interface{}{"ignoreMissingProperties": false, "ignoreExclusiveCheckedAttribute": false}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &CheckedRequiresOnchangeOrReadonlyRule, []rule_tester.ValidTestCase{
		{Code: `<input type="checkbox" />`, Tsx: true},
		{Code: `<input type="checkbox" onChange={noop} />`, Tsx: true},
		{Code: `<input type="checkbox" readOnly />`, Tsx: true},
		{Code: `<input type="checkbox" checked onChange={noop} />`, Tsx: true},
		{Code: `<input type="checkbox" checked={true} onChange={noop} />`, Tsx: true},
		{Code: `<input type="checkbox" checked={false} onChange={noop} />`, Tsx: true},
		{Code: `<input type="checkbox" checked readOnly />`, Tsx: true},
		{Code: `<input type="checkbox" checked={true} readOnly />`, Tsx: true},
		{Code: `<input type="checkbox" checked={false} readOnly />`, Tsx: true},
		{Code: `<input type="checkbox" defaultChecked />`, Tsx: true},
		{Code: `React.createElement('input')`, Tsx: true},
		{Code: `React.createElement('input', { checked: true, onChange: noop })`, Tsx: true},
		{Code: `React.createElement('input', { checked: false, onChange: noop })`, Tsx: true},
		{Code: `React.createElement('input', { checked: true, readOnly: true })`, Tsx: true},
		{Code: `React.createElement('input', { checked: true, onChange: noop, readOnly: true })`, Tsx: true},
		{Code: `React.createElement('input', { checked: foo, onChange: noop, readOnly: true })`, Tsx: true},
		{
			Code:    `<input type="checkbox" checked />`,
			Options: ignoreMissing,
			Tsx:     true,
		},
		{
			Code:    `<input type="checkbox" checked={true} />`,
			Options: ignoreMissing,
			Tsx:     true,
		},
		{
			Code:    `<input type="checkbox" onChange={noop} checked defaultChecked />`,
			Options: ignoreExclusive,
			Tsx:     true,
		},
		{
			Code:    `<input type="checkbox" onChange={noop} checked={true} defaultChecked />`,
			Options: ignoreExclusive,
			Tsx:     true,
		},
		{
			Code:    `<input type="checkbox" onChange={noop} checked defaultChecked />`,
			Options: bothIgnore,
			Tsx:     true,
		},
		{Code: `<span/>`, Tsx: true},
		{Code: `React.createElement('span')`, Tsx: true},
		{Code: `(()=>{})()`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `<input type="radio" checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="radio" checked={true} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="checkbox" checked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="checkbox" checked={true} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="checkbox" checked={condition ? true : false} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `<input type="checkbox" checked defaultChecked />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exclusiveCheckedAttribute", Line: 1, Column: 1},
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("input", { checked: false })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code: `React.createElement("input", { checked: true, defaultChecked: true })`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exclusiveCheckedAttribute", Line: 1, Column: 1},
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<input type="checkbox" checked defaultChecked />`,
			Options: ignoreMissing,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exclusiveCheckedAttribute", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<input type="checkbox" checked defaultChecked />`,
			Options: ignoreExclusive,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<input type="checkbox" checked defaultChecked />`,
			Options: bothFalse,
			Tsx:     true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "exclusiveCheckedAttribute", Line: 1, Column: 1},
				{MessageId: "missingProperty", Line: 1, Column: 1},
			},
		},
	})
}
