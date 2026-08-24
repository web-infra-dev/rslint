package catch_error_name

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

const messageID = "catch-error-name"

var reservedWords = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "debugger": true, "default": true, "delete": true, "do": true,
	"else": true, "enum": true, "export": true, "extends": true, "false": true,
	"finally": true, "for": true, "function": true, "if": true, "import": true,
	"in": true, "instanceof": true, "new": true, "null": true, "return": true,
	"super": true, "switch": true, "this": true, "throw": true, "true": true,
	"try": true, "typeof": true, "var": true, "void": true, "while": true,
	"with": true, "as": true, "await": true, "implements": true, "interface": true, "let": true,
	"package": true, "private": true, "protected": true, "public": true, "static": true,
	"yield": true, "any": true, "boolean": true, "constructor": true, "declare": true,
	"get": true, "module": true, "require": true, "number": true, "set": true,
	"string": true, "symbol": true, "type": true, "from": true, "of": true, "eval": true,
}

//go:embed catch_error_name.schema.json
var schemaJSON []byte

type options struct {
	name   string
	ignore []*esregexp.RegExp
}

func parseOptions(raw []any) options {
	result := options{name: "error"}
	if len(raw) == 0 {
		return result
	}
	values, _ := raw[0].(map[string]any)
	if name, ok := values["name"].(string); ok {
		result.name = name
	}
	if patterns, ok := values["ignore"].([]any); ok {
		for _, value := range patterns {
			pattern, ok := value.(string)
			if !ok {
				continue
			}
			if re, err := esregexp.Compile(pattern, "u"); err == nil {
				result.ignore = append(result.ignore, re)
			}
		}
	}
	return result
}

func jsxNamespaceReferences(identifier *ast.Node, body *ast.Node, name string) []*ast.Node {
	var references []*ast.Node
	declaration := bindingDeclaration(identifier)
	var boundary *ast.Node
	if declaration != nil {
		boundary = declaration.Parent
	}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindJsxNamespacedName {
			namespacedName := node.AsJsxNamespacedName()
			if namespacedName != nil && namespacedName.Namespace != nil && ast.IsJsxTagName(node) &&
				namespacedName.Namespace.Text() == name &&
				!utils.IsNameShadowedBetween(namespacedName.Namespace, boundary, name) {
				references = append(references, namespacedName.Namespace)
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(body)
	return references
}

func upperFirst(value string) string {
	if value == "" {
		return value
	}
	first, size := utf8.DecodeRuneInString(value)
	// Upstream uppercases `charAt(0)`, a single UTF-16 code unit. For an
	// astral character that unit is only the high surrogate and has no case
	// mapping, so the string stays unchanged.
	if first > 0xffff {
		return value
	}
	return ecmascript.StringToUpperCase(string(first)) + value[size:]
}

func (o options) allows(name string) bool {
	if name == o.name || strings.HasSuffix(name, o.name) || strings.HasSuffix(name, upperFirst(o.name)) {
		return true
	}
	for _, re := range o.ignore {
		if re.TestOrTimeout(name) {
			return true
		}
	}
	return false
}

func bindingDeclaration(identifier *ast.Node) *ast.Node {
	if identifier == nil || identifier.Parent == nil ||
		(identifier.Parent.Kind != ast.KindParameter && identifier.Parent.Kind != ast.KindVariableDeclaration) {
		return nil
	}
	declaration := identifier.Parent
	if declaration.Name() != identifier {
		return nil
	}
	return declaration
}

func isCatchParameter(identifier *ast.Node) bool {
	declaration := bindingDeclaration(identifier)
	if declaration == nil || declaration.Parent == nil || declaration.Parent.Kind != ast.KindCatchClause {
		return false
	}
	clause := declaration.Parent.AsCatchClause()
	return clause != nil && clause.VariableDeclaration == declaration
}

func isPromiseCatchParameter(identifier *ast.Node) bool {
	declaration := bindingDeclaration(identifier)
	if declaration == nil || declaration.Kind != ast.KindParameter || declaration.Parent == nil {
		return false
	}
	function := declaration.Parent
	if function.Kind != ast.KindArrowFunction && function.Kind != ast.KindFunctionExpression {
		return false
	}
	parameters := function.Parameters()
	if len(parameters) == 0 || unicornutil.PlainParameterIdentifier(parameters[0]) != identifier || function.Parent == nil {
		return false
	}

	argumentsLength := 1
	method := "catch"
	if match, ok := unicornutil.MatchDotMethodCall(function.Parent, unicornutil.DotMethodCallOptions{
		Method:              method,
		ArgumentsLength:     &argumentsLength,
		AllowOptionalMember: true,
	}); ok && match.Call.Arguments()[0] == function {
		return true
	}

	argumentsLength = 2
	method = "then"
	match, ok := unicornutil.MatchDotMethodCall(function.Parent, unicornutil.DotMethodCallOptions{
		Method:              method,
		ArgumentsLength:     &argumentsLength,
		AllowOptionalMember: true,
	})
	return ok && match.Call.Arguments()[1] == function
}

func isValidVariableName(name string) bool {
	return scanner.IsValidIdentifier(name) && !reservedWords[name]
}

func handlerBody(identifier *ast.Node) *ast.Node {
	declaration := bindingDeclaration(identifier)
	if declaration == nil || declaration.Parent == nil {
		return nil
	}
	parent := declaration.Parent
	if parent.Kind == ast.KindCatchClause {
		return parent.AsCatchClause().Block
	}
	return parent.Body()
}

func bodyHasExternalReference(ctx rule.RuleContext, body *ast.Node, originalSymbol *ast.Symbol, candidate string) bool {
	conflict := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || conflict {
			return
		}
		if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == candidate &&
			(node.Parent == nil || node.Parent.Kind != ast.KindJsxNamespacedName || ast.IsJsxTagName(node.Parent)) &&
			(!ast.IsJsxTagName(node) || !scanner.IsIntrinsicJsxName(node.Text())) &&
			!utils.IsNonReferenceIdentifier(node) {
			symbol := ctx.Refs.Resolve(node)
			conflict = symbol != originalSymbol && !utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
			if conflict {
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return conflict
		})
	}
	walk(body)
	return conflict
}

func statementsHaveTypeDeclaration(statements []*ast.Node, name string) bool {
	for _, statement := range statements {
		if statement == nil || (statement.Kind != ast.KindTypeAliasDeclaration && statement.Kind != ast.KindInterfaceDeclaration) {
			continue
		}
		if declarationName := statement.Name(); declarationName != nil && declarationName.Text() == name {
			return true
		}
	}
	return false
}

func hasTypeDeclarationInScope(node *ast.Node, name string) bool {
	for current := node; current != nil; current = current.Parent {
		if ast.IsFunctionLikeDeclaration(current) ||
			current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			for _, typeParameter := range current.TypeParameters() {
				if typeParameterName := typeParameter.Name(); typeParameterName != nil && typeParameterName.Text() == name {
					return true
				}
			}
		}
		switch current.Kind {
		case ast.KindBlock:
			block := current.AsBlock()
			if block != nil && block.Statements != nil && statementsHaveTypeDeclaration(block.Statements.Nodes, name) {
				return true
			}
		case ast.KindModuleBlock:
			moduleBlock := current.AsModuleBlock()
			if moduleBlock != nil && moduleBlock.Statements != nil &&
				(utils.HasLocalDeclarationInStatements(moduleBlock.Statements.Nodes, name) ||
					statementsHaveTypeDeclaration(moduleBlock.Statements.Nodes, name) ||
					utils.HasHoistedVarDeclaration(current, name)) {
				return true
			}
		case ast.KindCaseClause, ast.KindDefaultClause:
			clause := current.AsCaseOrDefaultClause()
			if clause != nil && clause.Statements != nil && statementsHaveTypeDeclaration(clause.Statements.Nodes, name) {
				return true
			}
		case ast.KindCaseBlock:
			caseBlock := current.AsCaseBlock()
			if caseBlock != nil && caseBlock.Clauses != nil {
				for _, clauseNode := range caseBlock.Clauses.Nodes {
					clause := clauseNode.AsCaseOrDefaultClause()
					if clause != nil && clause.Statements != nil &&
						statementsHaveTypeDeclaration(clause.Statements.Nodes, name) {
						return true
					}
				}
			}
		case ast.KindSourceFile:
			sourceFile := current.AsSourceFile()
			return sourceFile != nil && sourceFile.Statements != nil &&
				statementsHaveTypeDeclaration(sourceFile.Statements.Nodes, name)
		}
	}
	return false
}

func catchBodyHasUnsafeDeclaration(body *ast.Node, name string) bool {
	block := body.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}
	for _, statement := range block.Statements.Nodes {
		if statement == nil {
			continue
		}
		switch statement.Kind {
		case ast.KindVariableStatement:
			variableStatement := statement.AsVariableStatement()
			if variableStatement != nil && variableStatement.DeclarationList != nil &&
				!utils.IsVarKeyword(variableStatement.DeclarationList) &&
				utils.HasVarDeclListWithName(variableStatement.DeclarationList, name) {
				return true
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if declarationName := statement.Name(); declarationName != nil && declarationName.Text() == name {
				return true
			}
		}
	}
	return false
}

func availableName(ctx rule.RuleContext, identifier *ast.Node, references []*ast.Node, expected string) string {
	candidate := expected
	if !isValidVariableName(candidate) {
		candidate += "_"
		if !isValidVariableName(candidate) {
			return ""
		}
	}
	originalSymbol := bindingDeclaration(identifier).Symbol()
	body := handlerBody(identifier)
	for {
		available := !ctx.Globals.Access(candidate).IsDeclared() &&
			!utils.IsShadowed(identifier, candidate) &&
			!hasTypeDeclarationInScope(identifier, candidate)
		if available && candidate == "arguments" {
			available = false
		}
		if available {
			for _, reference := range references {
				if utils.IsShadowed(reference, candidate) || hasTypeDeclarationInScope(reference, candidate) {
					available = false
					break
				}
			}
		}
		if available && bodyHasExternalReference(ctx, body, originalSymbol, candidate) {
			available = false
		}
		// NOTE: Unlike ESLint, avoid a known unsafe fix when an unused handler
		// parameter would collide with a declaration in its body.
		if available && len(references) == 0 && body != nil && body.Kind == ast.KindBlock {
			if isCatchParameter(identifier) {
				available = !catchBodyHasUnsafeDeclaration(body, candidate)
			} else {
				block := body.AsBlock()
				available = !utils.HasShadowingDeclaration(body, candidate) &&
					!utils.HasHoistedVarDeclaration(body, candidate) &&
					(block == nil || block.Statements == nil || !statementsHaveTypeDeclaration(block.Statements.Nodes, candidate))
			}
		}
		if available {
			return candidate
		}
		candidate += "_"
	}
}

func renameFixes(ctx rule.RuleContext, identifier *ast.Node, references []*ast.Node, name string) []rule.RuleFix {
	fixes := make([]rule.RuleFix, 0, len(references)+1)
	fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, identifier, name))
	for _, reference := range references {
		replacement := name
		if parent := reference.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == reference {
			replacement = reference.Text() + ": " + name
		}
		fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, reference, replacement))
	}
	return fixes
}

var CatchErrorNameRule = rule.Rule{
	Name:   "unicorn/catch-error-name",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		return rule.RuleListeners{
			ast.KindIdentifier: func(identifier *ast.Node) {
				if !isCatchParameter(identifier) && !isPromiseCatchParameter(identifier) {
					return
				}
				originalName := identifier.AsIdentifier().Text
				trimmedName := strings.TrimRight(originalName, "_")
				if opts.allows(originalName) || opts.allows(trimmedName) {
					return
				}
				declaration := bindingDeclaration(identifier)
				if declaration == nil || ctx.Refs == nil || declaration.Symbol() == nil {
					return
				}
				references := append([]*ast.Node(nil), ctx.Refs.References(declaration.Symbol())...)
				references = append(references, jsxNamespaceReferences(identifier, handlerBody(identifier), originalName)...)
				sort.Slice(references, func(i, j int) bool { return references[i].Pos() < references[j].Pos() })
				if originalName == "_" && len(references) == 0 {
					return
				}
				fixedName := availableName(ctx, identifier, references, opts.name)
				messageName := fixedName
				if messageName == "" {
					messageName = opts.name
				}
				message := rule.RuleMessage{
					Id:          messageID,
					Description: fmt.Sprintf("The catch parameter `%s` should be named `%s`.", originalName, messageName),
					Data:        map[string]string{"originalName": originalName, "fixedName": messageName},
				}
				ctx.ReportNodeWithDeferredFixes(identifier, message, func() []rule.RuleFix {
					if fixedName == "" {
						return nil
					}
					return renameFixes(ctx, identifier, references, fixedName)
				})
			},
		}
	},
}
