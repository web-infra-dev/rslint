// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package unicornutil

import (
	"iter"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (c *domNodeTypeClassifier) bindingAnnotation(declaration *ast.Node) (*ast.Node, bool) {
	annotation := ast.GetTypeAnnotationNode(declaration)
	if annotation != nil && !utils.IsJSDocSyntaxNode(annotation) {
		return annotation, ast.HasQuestionToken(declaration)
	}
	if ast.IsParameterDeclaration(declaration) {
		return c.parameterAnnotation(declaration)
	}
	tag := c.uniqueJSDocTag(declaration.Name(), ast.KindJSDocTypeTag)
	if tag == nil || hasNativeTypeAssertion(declaration.Initializer()) {
		return nil, false
	}
	annotation = c.jsDocType(tag, declaration)
	return annotation, annotation != nil && ast.IsJSDocOptionalType(annotation)
}

func (c *domNodeTypeClassifier) parameterAnnotation(declaration *ast.Node) (*ast.Node, bool) {
	name := declaration.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return nil, false
	}
	var annotation *ast.Node
	var optional, found bool
	for tag := range c.jsDocTags(declaration.Parent, ast.KindJSDocParameterTag) {
		parameter := tag.AsJSDocParameterOrPropertyTag()
		if parameter.Name() == nil || !ast.IsIdentifier(parameter.Name()) {
			continue
		}
		var recovered *ast.Node
		if parameter.Name().Text() == "" {
			// The parser stops at `function` before reaching a type-first tag's name.
			var parsed *ast.Node
			recovered, parsed = c.recoverClosureFunction(tag, declaration)
			if parsed == nil {
				continue
			}
			parameter = parsed.AsJSDocParameterOrPropertyTag()
		}
		if parameter.Name() == nil || !ast.IsIdentifier(parameter.Name()) || parameter.Name().Text() != name.Text() {
			continue
		}
		if found {
			annotation, optional = nil, false
			break
		}
		found = true
		annotation = recovered
		if annotation == nil {
			annotation = c.jsDocType(tag, declaration)
		}
		optional = parameter.IsBracketed || (annotation != nil && ast.IsJSDocOptionalType(annotation))
	}
	if inline := c.jsDocType(c.uniqueJSDocTag(declaration, ast.KindJSDocTypeTag), declaration); inline != nil {
		annotation = inline
	}
	return annotation, optional || (annotation != nil && ast.IsJSDocOptionalType(annotation))
}

func (c *domNodeTypeClassifier) uniqueJSDocTag(host *ast.Node, kind ast.Kind) *ast.Node {
	var match *ast.Node
	for tag := range c.jsDocTags(host, kind) {
		if match != nil {
			return nil
		}
		match = tag
	}
	return match
}

func (c *domNodeTypeClassifier) jsDocTags(host *ast.Node, kind ast.Kind) iter.Seq[*ast.Node] {
	return func(yield func(*ast.Node) bool) {
		for owner := host; owner != nil; owner = ast.GetNextJSDocCommentLocation(owner) {
			// Statement-level JSDoc only describes single bindings.
			if ast.IsVariableDeclarationList(owner) && len(owner.AsVariableDeclarationList().Declarations.Nodes) > 1 {
				break
			}
			docs := owner.JSDoc(c.ctx.SourceFile)
			if len(docs) == 0 {
				continue
			}
			// Only tags in the closest documentation can conflict.
			if tags := docs[len(docs)-1].AsJSDoc().Tags; tags != nil {
				for _, tag := range tags.Nodes {
					if tag.Kind == kind && !yield(tag) {
						return
					}
				}
			}
			return
		}
	}
}

func (c *domNodeTypeClassifier) jsDocType(tag, declaration *ast.Node) *ast.Node {
	if tag == nil || tag.TypeExpression() == nil {
		return nil
	}
	annotation := tag.TypeExpression().Type()
	if containsClosureFunctionType(annotation) {
		if callable, _ := c.recoverClosureFunction(tag, declaration); callable != nil {
			return callable
		}
	}
	return annotation
}

func (c *domNodeTypeClassifier) recoverClosureFunction(tag, declaration *ast.Node) (*ast.Node, *ast.Node) {
	expression := tag.TypeExpression()
	if expression == nil || !containsClosureFunctionType(expression.Type()) {
		return nil, nil
	}
	raw := c.ctx.SourceFile.Text()[expression.Pos():tag.End()]
	end := closureTypeEnd(raw)
	if end == 0 {
		return nil, nil
	}
	normalized, changed := normalizeClosureFunctionTypes(raw[1 : end-1])
	if !changed {
		return nil, nil
	}
	// A synthetic @param lets tsgo read the complete type, parameter name and
	// optional marker together after adapting Closure signatures to call signatures.
	name := " value"
	if ast.IsJSDocParameterTag(tag) && tag.Name() != nil && ast.IsIdentifier(tag.Name()) && tag.Name().Text() == "" {
		name = raw[end:]
	}
	text := "/** @param {" + normalized + "}" + name + " */ function annotated(value) {}"
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/annotation.js", Path: "/annotation.js"}, text, core.ScriptKindJS)
	parsed := file.Statements.Nodes[0].JSDoc(file)[0].AsJSDoc().Tags.Nodes[0]
	expression = parsed.TypeExpression()
	if expression == nil {
		return nil, nil
	}
	annotation := expression.Type()
	if annotation == nil || len(file.JSDocDiagnostics()) != 0 {
		return nil, nil
	}
	annotation.Parent = declaration
	return annotation, parsed
}

func containsClosureFunctionType(node *ast.Node) bool {
	return node != nil && (isClosureFunctionType(node) || node.ForEachChild(containsClosureFunctionType))
}

func closureTypeEnd(raw string) int {
	if len(raw) == 0 || raw[0] != '{' {
		return 0
	}
	tokens := scanner.NewScanner()
	tokens.SetText(raw)
	tokens.SetSkipJSDocLeadingAsterisks(true)
	depth := 0
	for token := tokens.Scan(); token != ast.KindEndOfFile; token = tokens.Scan() {
		switch token {
		case ast.KindOpenBraceToken:
			depth++
		case ast.KindCloseBraceToken:
			depth--
			if depth == 0 {
				return tokens.TokenEnd()
			}
		}
	}
	return 0
}
