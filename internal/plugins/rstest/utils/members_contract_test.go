package utils_test

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func probeMessage(id string, description string) rule.RuleMessage {
	return rule.RuleMessage{Id: id, Description: description}
}

// misalignedMembersProbe reports whenever a parsed call's Members and
// MemberEntries disagree, so any such call shows up as an error on a test case
// declared valid.
var misalignedMembersProbe = rule.Rule{
	Name:             "rstest/members-alignment-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				aligned := len(parsed.Members) == len(parsed.MemberEntries)
				if aligned {
					for i := range parsed.Members {
						if parsed.Members[i] != parsed.MemberEntries[i].Name {
							aligned = false
							break
						}
					}
				}
				if !aligned {
					ctx.ReportNode(node, probeMessage(
						"misaligned", "Members and MemberEntries disagree"))
				}
			},
		}
	},
}

// Members and MemberEntries are the syntactic view of a call: both cover exactly
// the members written at the call site, so Members[i] == MemberEntries[i].Name
// holds by construction, and neither carries members consumed while following a
// const alias. Rules index one by the other, so feeding only Members would
// misreport or go out of range.
func TestMembersStayAlignedWithMemberEntries(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &misalignedMembersProbe,
		[]rule_tester.ValidTestCase{
			{Code: `test.skip.each([1])("case", cb);`},
			{Code: `const skipped = test.skip; skipped("case", cb);`},
			{Code: `const forCase = test.for([1]); forCase("case", cb);`},
			{Code: `const appTest = test.extend({}); appTest.concurrent("case", cb);`},
			{Code: `const base = test;
const skipped = base.skip;
skipped.each([1])("case", cb);`},
			{Code: `import * as rstest from "@rstest/core";
const skippedSuite = rstest.describe.skip;
skippedSuite("suite", cb);`},
			{Code: `describe.only("suite", cb);`},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

// semanticFlagProbe reports whenever a parsed call carries a semantic flag, so a
// test case listed as invalid asserts that the flag survived alias resolution
// even though Members is empty there.
var semanticFlagProbe = rule.Rule{
	Name:             "rstest/semantic-flag-probe",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				if parsed.Skipped || parsed.Todo || parsed.IsParameterized() {
					ctx.ReportNode(node, probeMessage("flagged", "call carries a semantic flag"))
				}
			},
		}
	},
}

// The semantic fields carry across an alias what Members deliberately drops:
// every invalid case below has an empty Members yet must still be flagged.
func TestSemanticFieldsSurviveAliases(t *testing.T) {
	flagged := []rule_tester.InvalidTestCaseError{{MessageId: "flagged"}}
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &semanticFlagProbe,
		[]rule_tester.ValidTestCase{
			{Code: `test("case", cb);`},
			{Code: `describe("suite", cb);`},
			{Code: `const plain = test; plain("case", cb);`},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `const skipped = test.skip; skipped("case", cb);`, Errors: flagged},
			{Code: `const todoTest = test.todo; todoTest("case");`, Errors: flagged},
			{Code: `const forCase = test.for([1]); forCase("case", cb);`, Errors: flagged},
			{Code: `const base = test;
const skipped = base.skip;
skipped.each([1])("case", cb);`, Errors: flagged},
			{Code: `import * as rstest from "@rstest/core";
const skippedSuite = rstest.describe.skip;
skippedSuite("suite", cb);`, Errors: flagged},
		},
	)
}
