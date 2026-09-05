package restrict_template_expressions

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed restrict_template_expressions.schema.json
var schemaJSON []byte

type RestrictTemplateExpressionsOptions struct {
	Allow        []utils.TypeOrValueSpecifier
	AllowAny     bool
	AllowArray   bool
	AllowBoolean bool
	AllowNever   bool
	AllowNullish bool
	AllowNumber  bool
	AllowRegExp  bool
}

var defaultAllowedTypes = []utils.TypeOrValueSpecifier{{
	From: utils.TypeOrValueSpecifierFromLib,
	Name: utils.NameList{"Error", "URL", "URLSearchParams"},
}}

func parseOptions(options []any) RestrictTemplateExpressionsOptions {
	opts := RestrictTemplateExpressionsOptions{
		Allow:        defaultAllowedTypes,
		AllowAny:     true,
		AllowBoolean: true,
		AllowNullish: true,
		AllowNumber:  true,
		AllowRegExp:  true,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if raw, ok := optsMap["allow"]; ok {
		opts.Allow = utils.ParseTypeOrValueSpecifiers(raw)
	}
	if v, ok := optsMap["allowAny"].(bool); ok {
		opts.AllowAny = v
	}
	if v, ok := optsMap["allowArray"].(bool); ok {
		opts.AllowArray = v
	}
	if v, ok := optsMap["allowBoolean"].(bool); ok {
		opts.AllowBoolean = v
	}
	if v, ok := optsMap["allowNever"].(bool); ok {
		opts.AllowNever = v
	}
	if v, ok := optsMap["allowNullish"].(bool); ok {
		opts.AllowNullish = v
	}
	if v, ok := optsMap["allowNumber"].(bool); ok {
		opts.AllowNumber = v
	}
	if v, ok := optsMap["allowRegExp"].(bool); ok {
		opts.AllowRegExp = v
	}
	return opts
}

func buildInvalidTypeMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidType",
		Description: "Invalid type \"" + t + "\" of template literal expression.",
	}
}

// getBaseTypesForType returns the base types of a class or interface type, and
// nothing for any other type.
func getBaseTypesForType(typeChecker *checker.Checker, t *checker.Type) []*checker.Type {
	if !utils.IsObjectType(t) {
		return nil
	}
	target := t
	if utils.IsTypeReference(t) {
		target = t.Target()
	}
	if checker.Type_objectFlags(target)&checker.ObjectFlagsClassOrInterface == 0 {
		return nil
	}
	return checker.Checker_getBaseTypes(typeChecker, target)
}

// matchesAllowedTypeOrBaseType reports whether the type or any of its base
// types satisfies one of the configured allow specifiers. It only allocates a
// visited set when there are base types to traverse.
func matchesAllowedTypeOrBaseType(
	typeChecker *checker.Checker,
	t *checker.Type,
	allow []utils.TypeOrValueSpecifier,
	program *program.Program,
) bool {
	if utils.TypeMatchesSomeSpecifier(t, allow, nil, program) {
		return true
	}

	baseTypes := getBaseTypesForType(typeChecker, t)
	if len(baseTypes) == 0 {
		return false
	}

	seen := make(map[*checker.Type]struct{}, len(baseTypes)+1)
	seen[t] = struct{}{}
	for _, base := range baseTypes {
		if matchesAllowedBaseType(typeChecker, base, allow, program, seen) {
			return true
		}
	}
	return false
}

func matchesAllowedBaseType(
	typeChecker *checker.Checker,
	t *checker.Type,
	allow []utils.TypeOrValueSpecifier,
	program *program.Program,
	seen map[*checker.Type]struct{},
) bool {
	if _, ok := seen[t]; ok {
		return false
	}
	seen[t] = struct{}{}
	if utils.TypeMatchesSomeSpecifier(t, allow, nil, program) {
		return true
	}
	for _, base := range getBaseTypesForType(typeChecker, t) {
		if matchesAllowedBaseType(typeChecker, base, allow, program, seen) {
			return true
		}
	}
	return false
}

type templateExpressionTypeChecker struct {
	typeChecker *checker.Checker
	program     *program.Program
	options     RestrictTemplateExpressionsOptions
}

func (c *templateExpressionTypeChecker) isAllowed(innerType *checker.Type, arrayPath []*checker.Type) bool {
	if innerType == nil {
		return false
	}
	if utils.IsUnionType(innerType) {
		for _, part := range innerType.Types() {
			if !c.isAllowed(part, arrayPath) {
				return false
			}
		}
		return true
	}
	if utils.IsIntersectionType(innerType) {
		for _, part := range innerType.Types() {
			if c.isAllowed(part, arrayPath) {
				return true
			}
		}
		return false
	}

	if utils.IsTypeFlagSet(innerType, checker.TypeFlagsStringLike) {
		return true
	}

	// These checks are pure alternatives to the allow list. Checking them first
	// avoids symbol and declaration work for the common primitive paths.
	if c.options.AllowAny && utils.IsTypeAnyType(innerType) {
		return true
	}
	if c.options.AllowArray &&
		(checker.Checker_isArrayType(c.typeChecker, innerType) || checker.IsTupleType(innerType)) {
		if !slices.Contains(arrayPath, innerType) {
			if arrayPath == nil {
				arrayPath = make([]*checker.Type, 0, 4)
			}
			if c.isAllowed(
				utils.GetNumberIndexType(c.typeChecker, innerType),
				append(arrayPath, innerType),
			) {
				return true
			}
		}
	}
	if c.options.AllowBoolean && utils.IsTypeFlagSet(innerType, checker.TypeFlagsBooleanLike) {
		return true
	}
	if c.options.AllowNullish && utils.IsTypeFlagSet(innerType, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
		return true
	}
	if c.options.AllowNumber && utils.IsTypeFlagSet(innerType, checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
		return true
	}
	if c.options.AllowNever && utils.IsTypeFlagSet(innerType, checker.TypeFlagsNever) {
		return true
	}

	if len(c.options.Allow) > 0 && matchesAllowedTypeOrBaseType(
		c.typeChecker,
		innerType,
		c.options.Allow,
		c.program,
	) {
		return true
	}
	return c.options.AllowRegExp && utils.GetTypeName(c.typeChecker, innerType) == "RegExp"
}

var RestrictTemplateExpressionsRule = rule.CreateRule(rule.Rule{
	Name:             "restrict-template-expressions",
	RequiresTypeInfo: true,
	Schema:           rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		typeChecker := templateExpressionTypeChecker{
			typeChecker: ctx.TypeChecker,
			program:     ctx.Program(),
			options:     parseOptions(options),
		}

		return rule.RuleListeners{
			ast.KindTemplateExpression: func(node *ast.Node) {
				// don't check tagged template literals
				if node.Parent != nil && ast.IsTaggedTemplateExpression(node.Parent) {
					return
				}

				for _, span := range node.AsTemplateExpression().TemplateSpans.Nodes {
					expression := span.AsTemplateSpan().Expression
					expressionType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)
					// Keep the overwhelmingly common direct-string path out of the
					// recursive checker.
					if expressionType != nil && utils.IsTypeFlagSet(expressionType, checker.TypeFlagsStringLike) {
						continue
					}
					if !typeChecker.isAllowed(expressionType, nil) {
						ctx.ReportNode(expression, buildInvalidTypeMessage(ctx.TypeChecker.TypeToString(expressionType)))
					}
				}
			},
		}
	},
})
