package consistent_json_file_read

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed consistent_json_file_read.schema.json
var schemaJSON []byte

const name = "consistent-json-file-read"

var (
	stringMessage = rule.RuleMessage{
		Id:          name + "/string",
		Description: "Prefer reading the JSON file as a string.",
	}
	bufferMessage = rule.RuleMessage{
		Id:          name + "/buffer",
		Description: "Prefer reading the JSON file as a buffer.",
	}
)

// ConsistentJSONFileReadRule enforces a consistent encoding before JSON.parse.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/consistent-json-file-read.js
var ConsistentJSONFileReadRule = rule.Rule{
	Name:   "unicorn/" + name,
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferBuffer := len(options) != 0 && options[0] == "buffer"
		one := 1
		parseOptions := unicornutil.DotMethodCallOptions{Method: "parse", ArgumentsLength: &one, RejectSpreadElement: true}
		var evaluator *utils.StaticStringEvaluator
		getEvaluator := func() *utils.StaticStringEvaluator {
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
				evaluator.GlobalAccess = ctx.Globals.Access
			}
			return evaluator
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parse, ok := unicornutil.MatchDotMethodCall(node, parseOptions)
				if !ok {
					return
				}
				object := utils.ESTreeRuntimeExpression(parse.Object)
				if !ast.IsIdentifier(object) || object.Text() != "JSON" ||
					!ctx.Globals.Access("JSON").IsDeclared() || ctx.Refs.IsDefinedInFile(object) {
					return
				}
				read := readExpression(&ctx, node.Arguments()[0])
				if read == nil || !ast.IsCallExpression(read) || ast.IsOptionalChain(read) {
					return
				}
				args := read.Arguments()
				if len(args) == 0 || len(args) > 2 || (preferBuffer && len(args) != 2) ||
					slices.ContainsFunc(args, func(arg *ast.Node) bool { return ast.IsSpreadElement(arg) || ast.NodeIsMissing(arg) }) {
					return
				}
				callee := utils.ESTreeCallCallee(read.Expression())
				if callee == nil || !isReadFileCallee(&ctx, callee, getEvaluator) {
					return
				}
				if len(args) == 1 {
					ctx.ReportNodeWithDeferredFixes(read, stringMessage, func() []rule.RuleFix {
						return insertEncoding(ctx.SourceFile, read)
					})
					return
				}

				encoding := utils.ESTreeRuntimeExpression(args[1])
				if !usesOppositeEncoding(&ctx, encoding, preferBuffer, getEvaluator) {
					return
				}
				message := stringMessage
				if preferBuffer {
					message = bufferMessage
				}
				ctx.ReportNodeWithDeferredFixes(encoding, message, func() []rule.RuleFix {
					// Folding can treat Object.freeze and similar calls as pass-through.
					// Their effects still matter when the argument itself is removed.
					if evaluator.HasSideEffect(encoding, false) {
						return nil
					}
					if preferBuffer {
						return removeEncoding(&ctx, args)
					}
					return replaceEncoding(&ctx, encoding)
				})
			},
		}
	},
}

// Follow initializers only while the parsed value is the binding's sole use.
// Resolving at each identifier keeps the initializer in its original scope.
func readExpression(ctx *rule.RuleContext, node *ast.Node) *ast.Node {
	for node = utils.ESTreeRuntimeExpression(node); node != nil; node = utils.ESTreeRuntimeExpression(node) {
		switch node.Kind {
		case ast.KindAwaitExpression:
			node = node.Expression()
		case ast.KindIdentifier:
			symbol := ctx.Refs.Resolve(node)
			if symbol == nil || len(symbol.Declarations) != 1 {
				return nil
			}
			declaration := symbol.Declarations[0]
			if !ast.IsVariableDeclaration(declaration) || !ast.IsIdentifier(declaration.Name()) ||
				ast.GetSourceFileOfNode(declaration) != ctx.SourceFile ||
				ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 ||
				ast.IsVarUsing(declaration) || ast.IsVarAwaitUsing(declaration) ||
				ctx.IsExportedGlobalBinding(utils.BindingNameSymbol(declaration.Name()), node.Text()) {
				return nil
			}
			initializer := declaration.Initializer()
			if initializer == nil {
				return nil
			}
			references := ctx.Refs.References(symbol)
			if len(references) != 1 || references[0] != node {
				return nil
			}
			node = initializer
		default:
			return node
		}
	}
	return nil
}

func isReadFileCallee(ctx *rule.RuleContext, callee *ast.Node, getEvaluator func() *utils.StaticStringEvaluator) bool {
	var method string
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		method, _ = utils.AccessExpressionStaticName(callee)
	case ast.KindElementAccessExpression:
		method, _ = getEvaluator().EvalAccessExpressionName(callee)
	case ast.KindIdentifier:
		_, method = fsImport(ctx.Refs.ResolveInFile(callee))
	}
	return method == "readFile" || method == "readFileSync"
}

func usesOppositeEncoding(ctx *rule.RuleContext, node *ast.Node, preferBuffer bool, getEvaluator func() *utils.StaticStringEvaluator) bool {
	node = utils.SkipAssertionsAndParens(node)
	if text, ok := utils.GetStaticStringLiteralValue(node); ok {
		return preferBuffer && isUTF8Encoding(text)
	}
	evaluator := getEvaluator()
	value, known := evaluator.EvalValueIfNoSideEffects(node)
	if !known {
		return false
	}
	property, isOptionsObject := utils.StaticPlainObjectOnlyProperty(value, "encoding")
	if isOptionsObject {
		value = property
	}
	matches := utils.IsNullishStaticValue(value)
	if preferBuffer {
		text, ok := value.(string)
		matches = ok && isUTF8Encoding(text)
	}
	return matches && (!isOptionsObject || optionsStayLocal(ctx, evaluator, node))
}

func isUTF8Encoding(text string) bool {
	switch ecmascript.StringToLowerCase(text) {
	case "utf8", "utf-8":
		return true
	default:
		return false
	}
}
