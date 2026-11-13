package consistent_return

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	base "github.com/web-infra-dev/rslint/internal/rules/consistent_return"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_return.schema.json
var schemaJSON []byte

// https://github.com/typescript-eslint/typescript-eslint/blob/v8.67.0/packages/eslint-plugin/src/rules/consistent-return.ts
var ConsistentReturnRule = rule.CreateRule(rule.Rule{
	Name:             "consistent-return",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run:              run,
})

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	// Several bare returns in the same function need only one signature query.
	// Keep the cache per file/run, never across checker or program lifetimes.
	voidReturns := make(map[*ast.Node]bool)
	return base.NewListeners(ctx, options, base.ReturnChecks{
		SkipReturn: func(root, statement *ast.Node) bool {
			if statement.Expression() != nil || root.Kind == ast.KindSourceFile || ctx.TypeChecker == nil {
				return false
			}
			if result, ok := voidReturns[root]; ok {
				return result
			}
			result := returnsVoid(ctx.TypeChecker, root)
			voidReturns[root] = result
			return result
		},
		IsUndefined: func(expression *ast.Node) bool {
			if ctx.TypeChecker == nil {
				return false
			}
			t := ctx.TypeChecker.GetTypeAtLocation(expression)
			// The upstream check is exact: a union merely containing undefined
			// must still be treated as a specified return value.
			return t != nil && checker.Type_flags(t) == checker.TypeFlagsUndefined
		},
	})
}

func returnsVoid(typeChecker *checker.Checker, node *ast.Node) bool {
	functionType := typeChecker.GetTypeAtLocation(node)
	if functionType == nil {
		return false
	}
	for _, signature := range utils.GetCallSignatures(typeChecker, functionType) {
		returnType := checker.Checker_getReturnTypeOfSignature(typeChecker, signature)
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
			if isPromiseVoid(typeChecker, node, returnType) {
				return true
			}
		} else if utils.IsTypeFlagSetWithUnion(returnType, checker.TypeFlagsVoid) {
			return true
		}
	}
	return false
}

func isPromiseVoid(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	if !utils.IsTypeReference(t) || !utils.IsThenableType(typeChecker, node, t) {
		return false
	}
	arguments := checker.Checker_getTypeArguments(typeChecker, t)
	if len(arguments) == 0 {
		return false
	}
	awaited := arguments[0]
	if utils.IsTypeFlagSetWithUnion(awaited, checker.TypeFlagsVoid) {
		return true
	}
	return isPromiseVoid(typeChecker, node, awaited)
}
