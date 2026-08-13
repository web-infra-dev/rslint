package switch_exhaustiveness_check

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed switch_exhaustiveness_check.schema.json
var schemaJSON []byte

type SwitchExhaustivenessCheckOptions struct {
	AllowDefaultCaseForExhaustiveSwitch bool   `json:"allowDefaultCaseForExhaustiveSwitch"`
	RequireDefaultForNonUnion           bool   `json:"requireDefaultForNonUnion"`
	ConsiderDefaultExhaustiveForUnions  bool   `json:"considerDefaultExhaustiveForUnions"`
	DefaultCaseCommentPattern           string `json:"defaultCaseCommentPattern"`
}

// SwitchExhaustivenessCheckRule implements the switch-exhaustiveness-check rule
// Require exhaustive switch statements
var SwitchExhaustivenessCheckRule = rule.CreateRule(rule.Rule{
	Name:             "switch-exhaustiveness-check",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run:              run,
})

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := SwitchExhaustivenessCheckOptions{
		AllowDefaultCaseForExhaustiveSwitch: false,
		RequireDefaultForNonUnion:           false,
		ConsiderDefaultExhaustiveForUnions:  false,
		DefaultCaseCommentPattern:           "",
	}

	// Parse options
	if len(options) > 0 {
		optsMap, _ := options[0].(map[string]interface{})
		if allowDefault, ok := optsMap["allowDefaultCaseForExhaustiveSwitch"].(bool); ok {
			opts.AllowDefaultCaseForExhaustiveSwitch = allowDefault
		}
		if requireDefault, ok := optsMap["requireDefaultForNonUnion"].(bool); ok {
			opts.RequireDefaultForNonUnion = requireDefault
		}
		if considerDefault, ok := optsMap["considerDefaultExhaustiveForUnions"].(bool); ok {
			opts.ConsiderDefaultExhaustiveForUnions = considerDefault
		}
		if pattern, ok := optsMap["defaultCaseCommentPattern"].(string); ok {
			opts.DefaultCaseCommentPattern = pattern
		}
	}

	return rule.RuleListeners{
		ast.KindSwitchStatement: func(node *ast.Node) {
			// This rule requires type information
			if ctx.TypeChecker == nil {
				return
			}

			switchStmt := node.AsSwitchStatement()
			if switchStmt == nil {
				return
			}

			// TODO: Implement exhaustiveness checking
			// This requires:
			// 1. Getting the type of switchStmt.Expression
			// 2. Checking if it's a union type
			// 3. Comparing covered cases against all union members
			// 4. Checking for default clause
			// 5. Reporting missing cases or dangerous default
			//
			// Example structure:
			// typ := ctx.TypeChecker.GetTypeAtLocation(switchStmt.Expression)
			// if isUnionType(typ) {
			//     coveredCases := getCoveredCases(switchStmt.CaseBlock)
			//     missingCases := findMissingCases(typ, coveredCases)
			//     if len(missingCases) > 0 && !hasDefaultClause(switchStmt.CaseBlock) {
			//         ctx.ReportNode(node, rule.RuleMessage{
			//             Id:          "switchIsNotExhaustive",
			//             Description: "Switch is not exhaustive",
			//         })
			//     }
			// }
		},
	}
}
