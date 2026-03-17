package no_undef

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-undef

type options struct {
	checkTypeof bool
}

func parseOptions(opts any) options {
	result := options{checkTypeof: false}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if v, ok := optsMap["typeof"].(bool); ok {
			result.checkTypeof = v
		}
	}
	return result
}

var globalCommentPattern = regexp.MustCompile(`/\*\s*global\s+([^*]+)\*/`)
var globalNamePattern = regexp.MustCompile(`(\w+)\s*(?::\s*\w+)?`)

// parseGlobalComments scans source text for /*global ...*/ block comments and
// returns the set of variable names declared in them.
func parseGlobalComments(sourceText string) map[string]bool {
	globals := make(map[string]bool)
	for _, match := range globalCommentPattern.FindAllStringSubmatch(sourceText, -1) {
		if len(match) > 1 {
			for _, nameMatch := range globalNamePattern.FindAllStringSubmatch(match[1], -1) {
				if len(nameMatch) > 1 {
					globals[nameMatch[1]] = true
				}
			}
		}
	}
	return globals
}

var NoUndefRule = rule.Rule{
	Name: "no-undef",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Without TypeChecker, this rule cannot resolve symbols
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		// Parse /*global ...*/ comments to find declared globals
		declaredGlobals := parseGlobalComments(ctx.SourceFile.Text())

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				// Skip identifiers that are not value references
				if shouldSkip(node, opts.checkTypeof) {
					return
				}

				// Try to resolve the symbol
				sym := ctx.TypeChecker.GetSymbolAtLocation(node)
				if sym != nil {
					return // Symbol found, variable is defined
				}

				name := node.Text()

				// Skip identifiers declared via /*global*/ comments
				if declaredGlobals[name] {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "undef",
					Description: fmt.Sprintf("'%s' is not defined.", name),
				})
			},
		}
	},
}

// shouldSkip returns true if the identifier should not be checked for being
// undefined. This includes property names, declaration names, labels, type
// positions, and typeof operands (when checkTypeof is false).
func shouldSkip(node *ast.Node, checkTypeof bool) bool {
	if node.Parent == nil {
		return true
	}

	// Skip declaration names (var x, function x, class x, import x, etc.)
	if ast.IsDeclarationName(node) {
		return true
	}

	parent := node.Parent

	// Skip property names in member access (obj.prop -> skip "prop")
	if parent.Kind == ast.KindPropertyAccessExpression {
		if parent.AsPropertyAccessExpression().Name() == node {
			return true
		}
	}

	// Skip property names in object literals ({key: value} -> skip "key")
	// Note: do NOT skip ShorthandPropertyAssignment names since they are also
	// value references ({x} is both property name and variable reference)
	if parent.Kind == ast.KindPropertyAssignment {
		if parent.Name() == node {
			return true
		}
	}

	// Skip label identifiers
	if parent.Kind == ast.KindLabeledStatement ||
		parent.Kind == ast.KindBreakStatement ||
		parent.Kind == ast.KindContinueStatement {
		return true
	}

	// Skip typeof operands (typeof x -> skip "x" unless checkTypeof is true)
	if !checkTypeof && parent.Kind == ast.KindTypeOfExpression {
		return true
	}

	// Skip identifiers in type-only positions (type annotations, type
	// references, interface/type alias declarations, etc.)
	if isInTypeOnlyPosition(node) {
		return true
	}

	// Skip the right side of qualified names (Namespace.Name)
	if parent.Kind == ast.KindQualifiedName {
		return true
	}

	// Skip enum member names
	if parent.Kind == ast.KindEnumMember && parent.Name() == node {
		return true
	}

	return false
}

// isInTypeOnlyPosition checks whether the identifier is inside a type-only
// context (type annotations, type references, etc.) by walking up the parent
// chain. Identifiers in type-only positions should not be checked because
// no-undef is about value references, not type references.
func isInTypeOnlyPosition(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		// If we hit a type node, the identifier is in a type-only position
		if ast.IsTypeNode(current) {
			return true
		}

		// Type alias declarations (type X = ...)
		if current.Kind == ast.KindTypeAliasDeclaration {
			return true
		}

		// Interface declarations (interface X { ... })
		if current.Kind == ast.KindInterfaceDeclaration {
			return true
		}

		// Heritage clause in 'implements' position is type-only
		if current.Kind == ast.KindHeritageClause {
			return true
		}

		// Stop walking at statement or declaration boundaries
		if isStatementOrDeclarationBoundary(current) {
			break
		}

		current = current.Parent
	}
	return false
}

// isStatementOrDeclarationBoundary returns true for nodes where we should stop
// walking up the parent chain when checking for type-only positions.
func isStatementOrDeclarationBoundary(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindBlock,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindVariableStatement,
		ast.KindExpressionStatement,
		ast.KindReturnStatement,
		ast.KindIfStatement,
		ast.KindSwitchStatement,
		ast.KindForStatement,
		ast.KindForInStatement,
		ast.KindForOfStatement,
		ast.KindWhileStatement,
		ast.KindDoStatement,
		ast.KindTryStatement,
		ast.KindThrowStatement:
		return true
	}
	return false
}
