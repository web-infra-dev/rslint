package function_component_definition

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed function_component_definition.schema.json
var schemaJSON []byte

// The three function forms the rule can enforce. These double as messageIds,
// matching upstream's `messages` map keys.
const (
	typeFunctionDeclaration = "function-declaration"
	typeFunctionExpression  = "function-expression"
	typeArrowFunction       = "arrow-function"
)

// namedFunctionTemplates / unnamedFunctionTemplates mirror upstream's
// NAMED_FUNCTION_TEMPLATES / UNNAMED_FUNCTION_TEMPLATES.
var namedFunctionTemplates = map[string]string{
	typeFunctionDeclaration: "function {name}{typeParams}({params}){returnType} {body}",
	typeArrowFunction:       "{varType} {name}{typeAnnotation} = {typeParams}({params}){returnType} => {body}",
	typeFunctionExpression:  "{varType} {name}{typeAnnotation} = function{typeParams}({params}){returnType} {body}",
}

var unnamedFunctionTemplates = map[string]string{
	typeFunctionExpression: "function{typeParams}({params}){returnType} {body}",
	typeArrowFunction:      "{typeParams}({params}){returnType} => {body}",
}

var messages = map[string]string{
	typeFunctionDeclaration: "Function component is not a function declaration",
	typeFunctionExpression:  "Function component is not a function expression",
	typeArrowFunction:       "Function component is not an arrow function",
}

// templatePartOrder is the key order upstream's `buildFunction` iterates
// (`Object.keys(parts)` over the object literal built in `getFixer`). Each
// key replaces only its FIRST occurrence in the template, so the order
// matters whenever an earlier-substituted value happens to contain the
// literal text of a later placeholder (a component body containing `{name}`,
// for instance). Keeping the order identical to upstream keeps the produced
// fix identical too.
var templatePartOrder = []string{"typeAnnotation", "typeParams", "params", "returnType", "body", "name", "varType"}

// buildFunction mirrors upstream's `buildFunction(template, parts)`.
func buildFunction(template string, parts map[string]string) string {
	out := template
	for _, key := range templatePartOrder {
		out = strings.Replace(out, "{"+key+"}", parts[key], 1)
	}
	return out
}

type options struct {
	named   []string
	unnamed []string
}

func parseOptions(opts []any) options {
	// Upstream: `[].concat(configuration.namedComponents || 'function-declaration')`.
	parsed := options{
		named:   []string{typeFunctionDeclaration},
		unnamed: []string{typeFunctionExpression},
	}
	if len(opts) == 0 {
		return parsed
	}
	optsMap, _ := opts[0].(map[string]any)
	if list, ok := readStringOrList(optsMap["namedComponents"]); ok {
		parsed.named = list
	}
	if list, ok := readStringOrList(optsMap["unnamedComponents"]); ok {
		parsed.unnamed = list
	}
	return parsed
}

// readStringOrList mirrors JS's `[].concat(value)` coercion for the
// `string | string[]` option shape. The bool result says whether the key was
// present in a recognized shape at all: upstream's `configuration.x || default`
// substitutes the default only for `undefined`, and an explicitly configured
// `[]` is truthy in JS and therefore kept as an empty list. So an empty array
// returns `(nil, true)` and must NOT fall back to the default.
func readStringOrList(raw any) ([]string, bool) {
	switch v := raw.(type) {
	case string:
		return []string{v}, true
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out, true
	}
	return nil, false
}

var FunctionComponentDefinitionRule = rule.Rule{
	Name:   "react/function-component-definition",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, opts []any) rule.RuleListeners {
		cfg := parseOptions(opts)
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)

		w := &walker{
			ctx:      ctx,
			text:     ctx.SourceFile.Text(),
			pragma:   pragma,
			wrappers: wrappers,
			cfg:      cfg,
		}
		w.run()
		return rule.RuleListeners{}
	},
}

type functionPair struct {
	node         *ast.Node
	functionType string
}

type walker struct {
	ctx      rule.RuleContext
	text     string
	pragma   string
	wrappers []reactutil.ComponentWrapperEntry
	cfg      options

	pairs []functionPair
	// hasES6OrJsx mirrors upstream's file-level flag: any ES module syntax or
	// any JSX element, plus any `const` / `let` declaration, means the file is
	// modern enough that a synthesized variable should use `const` instead of
	// `var`.
	hasES6OrJsx bool
}

func (w *walker) run() {
	w.collect(w.ctx.SourceFile.AsNode())
	// Upstream defers every report to `Program:exit` so that `fileVarType` is
	// final before any fix text is built.
	for _, pair := range w.pairs {
		w.validate(pair.node, pair.functionType)
	}
}

// collect walks the file once, mirroring upstream's listener set: the three
// function listeners push `[node, functionType]` pairs, and the ES-module /
// JSX / `const`-`let` listeners raise `hasES6OrJsx`.
func (w *walker) collect(root *ast.Node) {
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindFunctionDeclaration:
			w.pairs = append(w.pairs, functionPair{n, typeFunctionDeclaration})
		case ast.KindFunctionExpression:
			w.pairs = append(w.pairs, functionPair{n, typeFunctionExpression})
		case ast.KindArrowFunction:
			w.pairs = append(w.pairs, functionPair{n, typeArrowFunction})
		case ast.KindVariableDeclarationList:
			// ESTree `VariableDeclaration.kind`; tsgo keeps the kind on the
			// declaration list's node flags.
			switch utils.GetVarDeclListKind(n) {
			case "const", "let":
				w.hasES6OrJsx = true
			}
		case ast.KindImportDeclaration,
			ast.KindExportDeclaration,
			ast.KindExportAssignment,
			ast.KindExportSpecifier,
			ast.KindImportEqualsDeclaration,
			ast.KindJsxElement,
			ast.KindJsxSelfClosingElement:
			// tsgo splits ESTree's single JSXElement into JsxElement (with
			// children) and JsxSelfClosingElement; both are JSXElement
			// upstream. JsxFragment is deliberately absent — ESTree models it
			// as JSXFragment, which upstream's selector does not list.
			//
			// ExportDeclaration covers both `export {x}` and `export * from`;
			// ExportAssignment covers `export default <expr>` and `export =`.
			w.hasES6OrJsx = true
		}
		// ESTree hoists `export`/`export default` into wrapper nodes
		// (ExportNamedDeclaration / ExportDefaultDeclaration) that upstream's
		// selector matches; tsgo keeps `export` as a modifier on the declaration
		// itself, so the flag has to be read off the modifier list.
		if hasExportModifier(n) {
			w.hasES6OrJsx = true
		}
		n.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
}

func hasExportModifier(n *ast.Node) bool {
	mods := n.Modifiers()
	if mods == nil {
		return false
	}
	for _, m := range mods.Nodes {
		if m.Kind == ast.KindExportKeyword {
			return true
		}
	}
	return false
}

// fileVarType mirrors upstream's `fileVarType`, resolved at `Program:exit`.
func (w *walker) fileVarType() string {
	if w.hasES6OrJsx {
		return "const"
	}
	return "var"
}

func (w *walker) validate(node *ast.Node, functionType string) {
	if !w.isDetectedComponentNode(node) {
		return
	}
	// Upstream `if (node.parent && node.parent.type === 'Property') return;`.
	// Object-literal shorthand methods reach the same early return upstream
	// through the FunctionExpression listener; in tsgo they are
	// MethodDeclaration nodes, which `collect` never records at all.
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAssignment {
		return
	}

	named := hasName(node)
	// An explicitly configured empty list allows nothing, so upstream reaches
	// `report({ messageId: config[0] })` with `config[0] === undefined` and
	// ESLint aborts the whole lint run with "Missing `message` property in
	// report() call". rslint has no equivalent of that crash; staying silent is
	// the closest safe behavior and, unlike falling back to the default list,
	// it never invents a diagnostic upstream would not have produced.
	if named && len(w.cfg.named) > 0 && !contains(w.cfg.named, functionType) {
		target := w.cfg.named[0]
		w.report(node, target, namedFunctionTemplates[target], w.namedFixRange(node))
	}
	if !named && len(w.cfg.unnamed) > 0 && !contains(w.cfg.unnamed, functionType) {
		target := w.cfg.unnamed[0]
		w.report(node, target, unnamedFunctionTemplates[target], utils.TrimNodeTextRange(w.ctx.SourceFile, node))
	}
}

func (w *walker) report(node *ast.Node, target, template string, fixRange core.TextRange) {
	w.ctx.ReportRangeWithDeferredFixes(w.reportRange(node), rule.RuleMessage{
		Id:          target,
		Description: messages[target],
	}, func() []rule.RuleFix {
		replacement, ok := w.buildFix(node, target, template)
		if !ok {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, replacement)}
	})
}

// reportRange is the ESTree range of the reported function node. tsgo folds
// `export` / `export default` into the declaration's modifier list, whereas
// ESTree lifts them into a wrapper node that is not what upstream reports, so
// those two modifiers are skipped to keep the reported column aligned with
// ESLint's.
func (w *walker) reportRange(node *ast.Node) core.TextRange {
	if node.Kind == ast.KindFunctionDeclaration {
		return core.NewTextRange(reactutil.DeclarationKeywordStart(w.text, node), node.End())
	}
	return utils.TrimNodeTextRange(w.ctx.SourceFile, node)
}

// namedFixRange mirrors upstream's
// `node.type === 'FunctionDeclaration' ? node.range : node.parent.parent.range`.
//
// For the variable form, ESTree's `node.parent.parent` is the
// VariableDeclaration — the `var`/`let`/`const` keyword through the trailing
// semicolon, with any `export` kept on the enclosing wrapper node. tsgo splits
// that into VariableDeclarationList (keyword + declarators, no semicolon)
// inside a VariableStatement (which owns both the modifiers and the
// semicolon), so the equivalent range runs from the list's start to the
// statement's end.
func (w *walker) namedFixRange(node *ast.Node) core.TextRange {
	if node.Kind == ast.KindFunctionDeclaration {
		return core.NewTextRange(reactutil.DeclarationKeywordStart(w.text, node), node.End())
	}
	decl := declaratorParent(node)
	list := decl.Parent
	end := list.End()
	if list.Parent != nil && list.Parent.Kind == ast.KindVariableStatement {
		end = list.Parent.End()
	}
	return core.NewTextRange(utils.TrimNodeTextRange(w.ctx.SourceFile, list).Pos(), end)
}

// buildFix mirrors upstream's `getFixer`. The bool result is upstream's
// "returned undefined" — the diagnostic stands but carries no autofix.
func (w *walker) buildFix(node *ast.Node, target, template string) (string, bool) {
	typeAnnotation := w.typeAnnotationText(node)

	if target == typeFunctionDeclaration && typeAnnotation != "" {
		return "", false
	}
	if target == typeArrowFunction && hasOneUnconstrainedTypeParam(node) {
		return "", false
	}
	if isUnfixableBecauseOfExport(node) {
		return "", false
	}
	if isFunctionExpressionWithName(node) {
		return "", false
	}

	varType := w.fileVarType()
	if node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction {
		if decl := declaratorParent(node); decl != nil && decl.Kind == ast.KindVariableDeclaration {
			if kind := utils.GetVarDeclListKind(decl.Parent); kind != "" {
				varType = kind
			}
		}
	}

	body, ok := w.bodyText(node)
	if !ok {
		return "", false
	}

	return buildFunction(template, map[string]string{
		"typeAnnotation": typeAnnotation,
		"typeParams":     w.typeParamsText(node),
		"params":         w.paramsText(node),
		"returnType":     w.typeNodeTextWithColon(node.Type()),
		"body":           body,
		"name":           getName(node),
		"varType":        varType,
	}), true
}

// hasName mirrors upstream's `hasName`: a function declaration always has a
// name, and a function expression / arrow gets one from the variable
// declarator it initializes.
func hasName(node *ast.Node) bool {
	if node.Kind == ast.KindFunctionDeclaration {
		return true
	}
	decl := declaratorParent(node)
	return decl != nil && decl.Kind == ast.KindVariableDeclaration
}

// declaratorParent returns the node's parent as ESTree would see it, looking
// through parentheses only. ESTree has no ParenthesizedExpression, so
// `var Hello = ((props) => ...)` reaches the declarator directly upstream; TS
// wrappers (`as`, `satisfies`, `!`, `<T>x`) DO exist in TSESTree and are
// deliberately not skipped, because upstream treats a function behind one of
// them as unnamed.
func declaratorParent(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

func getName(node *ast.Node) string {
	if node.Kind == ast.KindFunctionDeclaration {
		return identifierText(node.Name())
	}
	if !hasName(node) {
		return ""
	}
	return identifierText(declaratorParent(node).Name())
}

func identifierText(name *ast.Node) string {
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

// typeAnnotationText mirrors upstream's `getTypeAnnotation`: only the variable
// declarator's own annotation, and never for a function declaration.
func (w *walker) typeAnnotationText(node *ast.Node) string {
	if !hasName(node) || node.Kind == ast.KindFunctionDeclaration {
		return ""
	}
	return w.typeNodeTextWithColon(declaratorParent(node).Type())
}

// typeNodeTextWithColon returns the source text of a type annotation
// INCLUDING its leading colon, matching the range of ESTree's TSTypeAnnotation
// wrapper (which tsgo has no node for — it stores the bare type). tsgo sets a
// node's Pos to the end of the preceding token, so the colon always sits at
// `Type.Pos() - 1`.
func (w *walker) typeNodeTextWithColon(typeNode *ast.Node) string {
	if typeNode == nil {
		return ""
	}
	start := typeNode.Pos() - 1
	if start < 0 || start >= len(w.text) || w.text[start] != ':' {
		return ""
	}
	return w.text[start:typeNode.End()]
}

// typeParamsText returns `<...>` including both angle brackets, matching the
// range of ESTree's TSTypeParameterDeclaration. tsgo's type-parameter NodeList
// spans only the parameters between the brackets.
func (w *walker) typeParamsText(node *ast.Node) string {
	list := typeParameterList(node)
	if list == nil || len(list.Nodes) == 0 {
		return ""
	}
	open := list.Pos() - 1
	if open < 0 || open >= len(w.text) || w.text[open] != '<' {
		return ""
	}
	// The closing `>` follows the last type parameter, possibly after a
	// trailing comma and/or trivia.
	closeIdx := list.End()
	for closeIdx < len(w.text) {
		closeIdx = scanner.SkipTrivia(w.text, closeIdx)
		if closeIdx >= len(w.text) {
			return ""
		}
		if w.text[closeIdx] == '>' {
			return w.text[open : closeIdx+1]
		}
		closeIdx++
	}
	return ""
}

func typeParameterList(node *ast.Node) *ast.NodeList {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return node.TypeParameterList()
	}
	return nil
}

// hasOneUnconstrainedTypeParam mirrors upstream's helper of the same name: one
// type parameter with no `extends` constraint, the shape whose arrow-function
// form (`<T>(props) => …`) is ambiguous with JSX and therefore unfixable.
func hasOneUnconstrainedTypeParam(node *ast.Node) bool {
	list := typeParameterList(node)
	if list == nil || len(list.Nodes) != 1 {
		return false
	}
	return list.Nodes[0].AsTypeParameterDeclaration().Constraint == nil
}

// paramsText mirrors upstream's `getParams`: the source span from the first
// parameter's start to the last parameter's end, excluding the parentheses and
// any trailing comma.
func (w *walker) paramsText(node *ast.Node) string {
	params := node.Parameters()
	if len(params) == 0 {
		return ""
	}
	start := utils.TrimNodeTextRange(w.ctx.SourceFile, params[0]).Pos()
	return w.text[start:params[len(params)-1].End()]
}

// bodyText mirrors upstream's `getBody`: a block body is copied verbatim,
// while an arrow's expression body is wrapped into an explicit block.
//
// Upstream slices `node.body.range`, and ESTree has no node for parentheses,
// so `const Hello = (props) => (cond ? <A /> : null);` yields the body text
// WITHOUT the source parentheses. tsgo keeps a ParenthesizedExpression node,
// so the outer parens have to be peeled back explicitly to produce the same
// fix.
func (w *walker) bodyText(node *ast.Node) (string, bool) {
	body := reactutil.FunctionBody(node)
	if body == nil {
		return "", false
	}
	if body.Kind != ast.KindBlock {
		body = ast.SkipParentheses(body)
	}
	bodyRange := utils.TrimNodeTextRange(w.ctx.SourceFile, body)
	if body.Kind != ast.KindBlock {
		return "{\n  return " + w.text[bodyRange.Pos():bodyRange.End()] + "\n}", true
	}
	return w.text[bodyRange.Pos():bodyRange.End()], true
}

// isUnfixableBecauseOfExport mirrors upstream's helper: a default-exported
// function declaration has no variable form to be rewritten into. tsgo keeps
// `export default` as modifiers on the declaration rather than as an
// ExportDefaultDeclaration parent.
func isUnfixableBecauseOfExport(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionDeclaration {
		return false
	}
	return ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsDefault != 0
}

func isFunctionExpressionWithName(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionExpression && identifierText(node.Name()) != ""
}

// isDetectedComponentNode answers upstream's `components.get(node)` for the
// function node the rule is about to validate — that is, whether the
// `Components.detect` pipeline registered THIS node (rather than nothing, or a
// wrapper call around it) with a non-zero confidence.
func (w *walker) isDetectedComponentNode(node *ast.Node) bool {
	// Confidence 0: `Components.detect`'s FunctionExpression /
	// FunctionDeclaration listeners permanently ban an async generator.
	if node.Kind != ast.KindArrowFunction && reactutil.IsAsyncGeneratorFunction(node) {
		return false
	}
	wrapper := reactutil.OutermostComponentWrapperCall(node, w.pragma, w.wrappers, w.ctx.TypeChecker)
	if wrapper != nil && reactutil.WrapperWrapsKnownSiblingComponent(wrapper, node) {
		return false
	}
	if !reactutil.IsStatelessReactComponentWithWrappers(node, w.pragma, w.ctx.TypeChecker, w.wrappers) {
		return false
	}
	// `getStatelessComponent` redirects a wrapped function to its outer-most
	// wrapper call, so the function node itself never enters the components
	// list — `components.get(node)` is null and the rule stays silent.
	return wrapper == nil
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
