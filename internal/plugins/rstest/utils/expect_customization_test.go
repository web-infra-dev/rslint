package utils_test

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var expectCustomizationProbe = rule.Rule{
	Name: "rstest/expect-customization-probe",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := node.Expression()
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "probe" {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id: "customization",
					Description: fmt.Sprintf(
						"toBe=%t toBeDefined=%t equality=%t",
						analysis.IsExpectMatcherOverridden("toBe"),
						analysis.IsExpectMatcherOverridden("toBeDefined"),
						analysis.HasCustomEqualityTesters(),
					),
				})
			},
		}
	},
}

func TestRstestExpectCustomizationAnalysis(t *testing.T) {
	want := func(message string) []rule_tester.InvalidTestCaseError {
		return []rule_tester.InvalidTestCaseError{{MessageId: "customization", Message: message}}
	}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &expectCustomizationProbe,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{Code: `expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; check['extend']({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: "const check = expect; check[`addEqualityTesters`]([tester]); probe();", Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `let check = expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; const again = (check as any)!; again['extend']({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; const extend = check.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; const { extend } = check; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import { expect as original } from '@rstest/core'; const check = original; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const check = core['expect']; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const core = require('rstack/test'); const check = core.expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = import.meta.rstest.expect; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `test('case', ctx => { const check = ctx.expect; check.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const check = expect; check.addEqualityTesters([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `const check = expect; const { addEqualityTesters: add } = check; add([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `const check = expect; check.addEqualityTesters([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let check = expect; check = other; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = other; const other = check; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let check; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = expect(); check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = expect.soft; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from '@rstest/core'; const expect = 'other'; const check = core[expect]; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = import.meta.rstest.expect(); check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import { expect as original } from 'vitest'; const check = original; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `function configure(expect: any) { const check = expect; check.extend({ toBe() {} }); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const check = expect; { const check = other; check.extend({ toBe() {} }); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `expect.extend({ ["toBe"]() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `expect.extend({ ...matchers }); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `expect.extend(matchers); probe();`, Errors: want("toBe=true toBeDefined=true equality=false")},
			{Code: `expect.addEqualityTesters([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `expect.addEqualityTesters([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `import { expect as check } from "@rstest/core"; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from "rstack/test"; core.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import.meta.rstest.expect.extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const extend = expect.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend = expect.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import * as core from "@rstest/core"; const extend = core.expect.extend; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `const add = import.meta.rstest.expect.addEqualityTesters; add([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `const { extend } = expect; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let { extend } = expect; extend({ toBe() {} }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `let extend = expect.extend; extend = other; extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `let { extend } = expect; extend = other; extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `import * as core from "@rstest/core"; const { addEqualityTesters: add } = core.expect; add([]); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `test("case", ({ expect }) => { expect.extend({ toBe() {} }); }); probe();`, Errors: want("toBe=true toBeDefined=false equality=false")},
			{Code: `import { expect as check } from "@rstest/core"; check.addEqualityTesters([tester]); probe();`, Errors: want("toBe=false toBeDefined=false equality=true")},
			{Code: `import { expect as check } from "vitest"; check.extend({ toBe() {} }); probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
			{Code: `function configure(expect: any) { expect.extend({ toBe() {} }); } probe();`, Errors: want("toBe=false toBeDefined=false equality=false")},
		},
	)
}
