package restrict_template_expressions

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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

// parseAllow decodes the `allow` option — type specifiers written either in
// the string shorthand or in object form — through TypeOrValueSpecifier's own
// UnmarshalJSON rather than re-deriving both shapes by hand.
func parseAllow(raw any) []utils.TypeOrValueSpecifier {
	rawJSON, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var allow []utils.TypeOrValueSpecifier
	if err := json.Unmarshal(rawJSON, &allow); err != nil {
		return nil
	}
	return allow
}

func parseOptions(options []any) RestrictTemplateExpressionsOptions {
	opts := RestrictTemplateExpressionsOptions{
		Allow: []utils.TypeOrValueSpecifier{{
			From: utils.TypeOrValueSpecifierFromLib,
			Name: utils.NameList{"Error", "URL", "URLSearchParams"},
		}},
		AllowAny:     true,
		AllowBoolean: true,
		AllowNullish: true,
		AllowNumber:  true,
		AllowRegExp:  true,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, ok := options[0].(map[string]interface{})
	if !ok {
		return opts
	}
	if raw, ok := optsMap["allow"]; ok {
		opts.Allow = parseAllow(raw)
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
		Description: fmt.Sprintf("Invalid type \"%v\" of template literal expression.", t),
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

// matchesTypeOrBaseType reports whether the type or any of its base types
// satisfies the matcher.
func matchesTypeOrBaseType(
	typeChecker *checker.Checker,
	matcher func(t *checker.Type) bool,
	t *checker.Type,
	seen map[*checker.Type]struct{},
) bool {
	if _, ok := seen[t]; ok {
		return false
	}
	seen[t] = struct{}{}

	if matcher(t) {
		return true
	}

	for _, base := range getBaseTypesForType(typeChecker, t) {
		if matchesTypeOrBaseType(typeChecker, matcher, base, seen) {
			return true
		}
	}
	return false
}

var RestrictTemplateExpressionsRule = rule.CreateRule(rule.Rule{
	Name:             "restrict-template-expressions",
	RequiresTypeInfo: true,
	Schema:           rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		var recursivelyCheckType func(innerType *checker.Type) bool
		recursivelyCheckType = func(innerType *checker.Type) bool {
			if innerType == nil {
				return false
			}
			if utils.IsUnionType(innerType) {
				return utils.Every(innerType.Types(), recursivelyCheckType)
			}
			if utils.IsIntersectionType(innerType) {
				return utils.Some(innerType.Types(), recursivelyCheckType)
			}

			if utils.IsTypeFlagSet(innerType, checker.TypeFlagsStringLike) {
				return true
			}

			if len(opts.Allow) > 0 && matchesTypeOrBaseType(ctx.TypeChecker, func(t *checker.Type) bool {
				return utils.TypeMatchesSomeSpecifier(t, opts.Allow, nil, ctx.Program)
			}, innerType, map[*checker.Type]struct{}{}) {
				return true
			}

			if opts.AllowAny && utils.IsTypeAnyType(innerType) {
				return true
			}
			if opts.AllowArray &&
				(checker.Checker_isArrayType(ctx.TypeChecker, innerType) || checker.IsTupleType(innerType)) &&
				recursivelyCheckType(utils.GetNumberIndexType(ctx.TypeChecker, innerType)) {
				return true
			}
			if opts.AllowBoolean && utils.IsTypeFlagSet(innerType, checker.TypeFlagsBooleanLike) {
				return true
			}
			if opts.AllowNullish && utils.IsTypeFlagSet(innerType, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
				return true
			}
			if opts.AllowNumber && utils.IsTypeFlagSet(innerType, checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
				return true
			}
			if opts.AllowRegExp && utils.GetTypeName(ctx.TypeChecker, innerType) == "RegExp" {
				return true
			}
			if opts.AllowNever && utils.IsTypeFlagSet(innerType, checker.TypeFlagsNever) {
				return true
			}

			return false
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
					if !recursivelyCheckType(expressionType) {
						ctx.ReportNode(expression, buildInvalidTypeMessage(ctx.TypeChecker.TypeToString(expressionType)))
					}
				}
			},
		}
	},
})
