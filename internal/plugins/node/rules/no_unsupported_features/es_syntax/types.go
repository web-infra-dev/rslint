package es_syntax

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Like es-x, obvious receivers take precedence over optional project types.
// Source-only lint continues to use the small syntax-based inference above.
func (c *syntaxChecker) matchesPrototype(member, object *ast.Node, inferred, class string, aggressive bool) bool {
	object = utils.ESTreeRuntimeExpression(object)
	switch object.Kind {
	case ast.KindArrayLiteralExpression, ast.KindRegularExpressionLiteral, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression, ast.KindFunctionExpression, ast.KindArrowFunction:
		return inferred == class
	}
	if c.ctx.TypeChecker == nil {
		return inferred == class || inferred == "" && aggressive
	}
	matches := func(t *checker.Type) bool { return c.typeMatches(t, class, aggressive, map[*checker.Type]bool{}) }
	if symbol := c.ctx.TypeChecker.GetSymbolAtLocation(propertyKey(member)); symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Parent != nil && matches(c.ctx.TypeChecker.GetTypeAtLocation(declaration.Parent)) {
				return true
			}
		}
	}
	return matches(c.ctx.TypeChecker.GetTypeAtLocation(object))
}
func (c *syntaxChecker) typeMatches(t *checker.Type, class string, aggressive bool, seen map[*checker.Type]bool) bool {
	// es-x accepts any matching union/intersection arm and named interfaces.
	// IsBuiltinSymbolLike instead requires every union arm and a library symbol.
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	symbol := checker.Type_symbol(t)
	if symbol != nil && symbol.Flags&(ast.SymbolFlagsFunction|ast.SymbolFlagsMethod) != 0 || len(utils.GetCallSignatures(c.ctx.TypeChecker, t)) > 0 {
		return class == "Function"
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return aggressive
	}
	flags := checker.Type_objectFlags(t)
	if utils.IsObjectType(t) && flags&checker.ObjectFlagsAnonymous != 0 {
		return false
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike) {
		return class == "String"
	}
	if utils.IsObjectType(t) && flags&(checker.ObjectFlagsArrayLiteral|checker.ObjectFlagsEvolvingArray|checker.ObjectFlagsTuple) != 0 {
		return class == "Array"
	}
	if utils.IsTypeReference(t) && t.Target() != t {
		return c.typeMatches(t.Target(), class, aggressive, seen)
	}
	if constraint, parameter := utils.GetConstraintInfo(c.ctx.TypeChecker, t); parameter {
		return constraint != nil && constraint != t && c.typeMatches(constraint, class, aggressive, seen)
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		for _, part := range t.Types() {
			if c.typeMatches(part, class, aggressive, seen) {
				return true
			}
		}
		return false
	}
	if utils.IsObjectType(t) && flags&checker.ObjectFlagsClassOrInterface != 0 && symbol != nil {
		if strings.Contains(class, ".") {
			return c.ctx.TypeChecker.GetFullyQualifiedName(symbol) == class
		}
		return symbol.Name == class || symbol.Name == "Readonly"+class || class == "Function" && symbol.Name == "CallableFunction"
	}
	return c.ctx.TypeChecker.TypeToString(t) == class
}
