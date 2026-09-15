package prefer_readonly_parameter_types

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/typescriptutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prefer_readonly_parameter_types.schema.json
var schemaJSON []byte

type PreferReadonlyParameterTypesOptions struct {
	CheckParameterProperties bool `json:"checkParameterProperties"`
	IgnoreInferredTypes      bool `json:"ignoreInferredTypes"`
	TreatMethodsAsReadonly   bool
	Allow                    []utils.TypeOrValueSpecifier
}

func buildShouldBeReadonlyMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "shouldBeReadonly",
		Description: "Parameter should be a read only type.",
	}
}

func parseOptions(options []any) PreferReadonlyParameterTypesOptions {
	opts := PreferReadonlyParameterTypesOptions{
		CheckParameterProperties: true,
		IgnoreInferredTypes:      false,
		TreatMethodsAsReadonly:   false,
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	if v, ok := optsMap["checkParameterProperties"].(bool); ok {
		opts.CheckParameterProperties = v
	}
	if v, ok := optsMap["ignoreInferredTypes"].(bool); ok {
		opts.IgnoreInferredTypes = v
	}
	if v, ok := optsMap["treatMethodsAsReadonly"].(bool); ok {
		opts.TreatMethodsAsReadonly = v
	}
	if v, ok := optsMap["allow"]; ok {
		opts.Allow = utils.ParseTypeOrValueSpecifiers(v)
	}
	return opts
}

// checkParameter validates a parameter node
func checkParameter(ctx rule.RuleContext, param *ast.Node, opts PreferReadonlyParameterTypesOptions) {
	paramDecl := param.AsParameterDeclaration()
	if paramDecl == nil {
		return
	}

	// typescript-estree represents a defaulted parameter as an
	// AssignmentPattern. Its annotation belongs to the left-hand node rather
	// than the AssignmentPattern itself, which is the node upstream checks.
	if opts.IgnoreInferredTypes && (paramDecl.Type == nil || paramDecl.Initializer != nil) {
		return
	}

	// Reading an authored annotation through GetTypeFromTypeNode preserves its
	// alias symbol, which is required for allowlist matching around unions.
	paramType := ctx.TypeChecker.GetTypeAtLocation(param)
	if paramDecl.Type != nil {
		paramType = ctx.TypeChecker.GetTypeFromTypeNode(paramDecl.Type)
	}
	if paramType == nil {
		return
	}

	readonly := typescriptutil.IsTypeReadonly(ctx.TypeChecker, paramType, typescriptutil.ReadonlynessOptions{
		Allow:                  opts.Allow,
		TreatMethodsAsReadonly: opts.TreatMethodsAsReadonly,
	}, ctx.Program())
	if !readonly && !typescriptutil.IsTypeBrandedLiteralLike(ctx.TypeChecker, paramType) {
		// ESTree excludes parameter decorators and TSParameterProperty modifiers
		// from the reported parameter node, but keeps a RestElement's `...`.
		startNode := paramDecl.Name()
		if paramDecl.DotDotDotToken != nil {
			startNode = paramDecl.DotDotDotToken
		}
		start := scanner.SkipTrivia(ctx.SourceFile.Text(), startNode.Pos())
		ctx.ReportRange(core.NewTextRange(start, param.End()), buildShouldBeReadonlyMessage())
	}
}

var PreferReadonlyParameterTypesRule = rule.CreateRule(rule.Rule{
	Name:             "prefer-readonly-parameter-types",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		checkParameters := func(node *ast.Node) {
			params := node.Parameters()
			if params == nil {
				return
			}

			for _, param := range params {
				if !opts.CheckParameterProperties && ast.IsParameterPropertyDeclaration(param, node) {
					continue
				}
				checkParameter(ctx, param, opts)
			}
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: checkParameters,
			ast.KindFunctionExpression:  checkParameters,
			ast.KindArrowFunction:       checkParameters,
			ast.KindMethodDeclaration:   checkParameters,
			ast.KindConstructor:         checkParameters,
			ast.KindGetAccessor:         checkParameters,
			ast.KindSetAccessor:         checkParameters,
			ast.KindCallSignature:       checkParameters,
			ast.KindConstructSignature:  checkParameters,
			ast.KindFunctionType:        checkParameters,
			ast.KindMethodSignature:     checkParameters,
		}
	},
})
