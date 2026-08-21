package use_isnan

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed use_isnan.schema.json
var schemaJSON []byte

var (
	comparisonWithNaNMessage = rule.RuleMessage{
		Id:          "comparisonWithNaN",
		Description: "Use the isNaN function to compare with NaN.",
	}
	switchNaNMessage = rule.RuleMessage{
		Id:          "switchNaN",
		Description: "'switch(NaN)' can never match a case clause. Use Number.isNaN instead of the switch.",
	}
	caseNaNMessage = rule.RuleMessage{
		Id:          "caseNaN",
		Description: "'case NaN' can never match. Use Number.isNaN before the switch.",
	}
	indexOfNaNMessage = rule.RuleMessage{
		Id:          "indexOfNaN",
		Description: "Array prototype method 'indexOf' cannot find NaN.",
	}
	lastIndexOfNaNMessage = rule.RuleMessage{
		Id:          "indexOfNaN",
		Description: "Array prototype method 'lastIndexOf' cannot find NaN.",
	}
)

// unwrapToValue strips parentheses (any depth), resolves at most one level
// of comma expression (taking the last element), then strips parentheses again.
// This matches ESLint's isNaNIdentifier which does:
//
//	node.type === "SequenceExpression" ? node.expressions.at(-1) : node
//
// In ESTree parentheses are not AST nodes, but in tsgo they are, so we must
// strip them before and after the comma resolution.
func unwrapToValue(node *ast.Node) *ast.Node {
	stripped := ast.SkipParentheses(node)

	if stripped.Kind == ast.KindBinaryExpression {
		binary := stripped.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken {
			return ast.SkipParentheses(binary.Right)
		}
	}

	return stripped
}

type globalReferenceChecker struct {
	refs          *rule.RefStore
	sourceFile    *ast.SourceFile
	bindingsReady bool
	mayShadowNaN  bool
	mayShadowNum  bool
}

// isNaNIdentifier checks if a node represents NaN (either as the identifier
// "NaN" or as the member expression "Number.NaN").
// Parentheses and one sequence-expression level are transparent, matching ESLint.
func isNaNIdentifier(checker *globalReferenceChecker, node *ast.Node) bool {
	if node == nil {
		return false
	}

	nodeToCheck := unwrapToValue(node)

	// Check for bare NaN identifier
	if nodeToCheck.Kind == ast.KindIdentifier {
		return nodeToCheck.AsIdentifier().Text == "NaN" && checker.isGlobalReference(nodeToCheck)
	}

	// Check for Number.NaN / Number['NaN'] / Number?.NaN / Number[`NaN`]
	if !isNaNProperty(nodeToCheck) {
		return false
	}
	object := ast.SkipParentheses(nodeToCheck.Expression())
	return object != nil && object.Kind == ast.KindIdentifier &&
		object.AsIdentifier().Text == "Number" && checker.isGlobalReference(object)
}

func isNaNProperty(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "NaN"
	case ast.KindElementAccessExpression:
		argument := utils.SkipAssertionsAndParens(node.AsElementAccessExpression().ArgumentExpression)
		name, ok := utils.GetStaticExpressionValue(argument)
		return ok && name == "NaN"
	}
	return false
}

func (checker *globalReferenceChecker) isGlobalReference(identifier *ast.Node) bool {
	name := identifier.AsIdentifier().Text
	if isInsideNamedExpression(identifier, name) {
		return false
	}
	if checker.refs == nil {
		return !utils.IsShadowed(identifier, name)
	}
	if !checker.mayShadow(name) {
		return true
	}
	if symbol := checker.refs.Resolve(identifier); symbol != nil {
		return !utils.IsValueSymbolDeclaredInFile(symbol, checker.sourceFile)
	}
	// Namespace-only bindings are outside RefStore's value lookup. Preserve a
	// syntactic fallback for those and for parser-recovered trees.
	return !utils.IsShadowed(identifier, name)
}

func isInsideNamedExpression(identifier *ast.Node, name string) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindFunctionExpression && current.Kind != ast.KindClassExpression {
			continue
		}
		expressionName := current.Name()
		if expressionName != nil && expressionName.Kind == ast.KindIdentifier &&
			expressionName.AsIdentifier().Text == name {
			return true
		}
	}
	return false
}

func (checker *globalReferenceChecker) mayShadow(name string) bool {
	if !checker.bindingsReady {
		checker.bindingsReady = true
		for container := checker.sourceFile.AsNode(); container != nil; {
			data := container.LocalsContainerData()
			if data == nil {
				break
			}
			if utils.IsValueSymbolDeclaredInFile(data.Locals["NaN"], checker.sourceFile) {
				checker.mayShadowNaN = true
			}
			if utils.IsValueSymbolDeclaredInFile(data.Locals["Number"], checker.sourceFile) {
				checker.mayShadowNum = true
			}
			container = data.NextContainer
		}
	}
	if name == "NaN" {
		return checker.mayShadowNaN
	}
	return checker.mayShadowNum
}

func sourceMayUseNaN(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil {
		return true
	}
	if sourceFile.HasIdentifier("NaN") {
		return true
	}
	return sourceFile.HasIdentifier("Number")
}

func isComparisonOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsEqualsToken,
		ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken,
		ast.KindExclamationEqualsEqualsToken,
		ast.KindGreaterThanToken,
		ast.KindGreaterThanEqualsToken,
		ast.KindLessThanToken,
		ast.KindLessThanEqualsToken:
		return true
	}
	return false
}

type useIsNaNOptions struct {
	enforceForSwitchCase bool
	enforceForIndexOf    bool
}

func parseOptions(options []any) useIsNaNOptions {
	result := useIsNaNOptions{
		enforceForSwitchCase: true,
		enforceForIndexOf:    false,
	}
	if len(options) == 0 {
		return result
	}

	optsMap, _ := options[0].(map[string]any)
	if boolVal, ok := optsMap["enforceForSwitchCase"].(bool); ok {
		result.enforceForSwitchCase = boolVal
	}
	if boolVal, ok := optsMap["enforceForIndexOf"].(bool); ok {
		result.enforceForIndexOf = boolVal
	}

	return result
}

// UseIsNaNRule requires calls to isNaN() when checking for NaN
var UseIsNaNRule = rule.Rule{
	Name:   "use-isnan",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !sourceMayUseNaN(ctx.SourceFile) {
			return nil
		}

		opts := parseOptions(options)
		checker := globalReferenceChecker{
			refs:       ctx.Refs,
			sourceFile: ctx.SourceFile,
		}

		listeners := rule.RuleListeners{
			// Check binary expressions for NaN comparisons
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				if !isComparisonOperator(binary.OperatorToken.Kind) {
					return
				}

				if isNaNIdentifier(&checker, binary.Left) || isNaNIdentifier(&checker, binary.Right) {
					ctx.ReportNode(node, comparisonWithNaNMessage)
				}
			},
		}

		// Check switch statements for NaN in discriminant and case clauses
		if opts.enforceForSwitchCase {
			listeners[ast.KindSwitchStatement] = func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				if switchStmt == nil {
					return
				}

				// Check if discriminant is NaN
				if isNaNIdentifier(&checker, switchStmt.Expression) {
					ctx.ReportNode(node, switchNaNMessage)
				}

				// Check case clauses for NaN
				if switchStmt.CaseBlock == nil {
					return
				}
				caseBlock := switchStmt.CaseBlock.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				for _, clause := range caseBlock.Clauses.Nodes {
					if clause.Kind != ast.KindCaseClause {
						continue
					}
					caseClause := clause.AsCaseOrDefaultClause()
					if caseClause == nil || caseClause.Expression == nil {
						continue
					}
					if isNaNIdentifier(&checker, caseClause.Expression) {
						ctx.ReportNode(clause, caseNaNMessage)
					}
				}
			}
		}

		// Check indexOf/lastIndexOf calls with NaN argument
		if opts.enforceForIndexOf {
			listeners[ast.KindCallExpression] = func(node *ast.Node) {
				// Get the callee expression, skipping parentheses for (foo?.indexOf)(NaN)
				callee := ast.SkipParentheses(node.Expression())
				if callee == nil {
					return
				}

				// Extract the method name from dot and bracket notation.
				message, ok := indexOfMethod(callee)
				if !ok {
					return
				}

				// Check if the first argument is NaN, with at most 2 arguments
				args := node.Arguments()
				if len(args) == 0 || len(args) > 2 {
					return
				}

				if isNaNIdentifier(&checker, args[0]) {
					ctx.ReportNode(node, message)
				}
			}
		}

		return listeners
	},
}

func indexOfMethod(callee *ast.Node) (rule.RuleMessage, bool) {
	var methodName string
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return rule.RuleMessage{}, false
		}
		methodName = name.AsIdentifier().Text
	case ast.KindElementAccessExpression:
		argument := utils.SkipAssertionsAndParens(callee.AsElementAccessExpression().ArgumentExpression)
		var ok bool
		methodName, ok = utils.GetStaticExpressionValue(argument)
		if !ok {
			return rule.RuleMessage{}, false
		}
	default:
		return rule.RuleMessage{}, false
	}

	switch methodName {
	case "indexOf":
		return indexOfNaNMessage, true
	case "lastIndexOf":
		return lastIndexOfNaNMessage, true
	default:
		return rule.RuleMessage{}, false
	}
}
