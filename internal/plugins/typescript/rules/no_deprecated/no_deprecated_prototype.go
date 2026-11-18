package no_deprecated

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/utils"
	"strings"
)

type prototypeMemberKey struct {
	constructor *ast.Node
	name        string
}

type prototypeResult struct {
	deprecationInfo
	declaration *ast.Node
}

// The checker does not yet model prototype members of JS constructor
// functions. Retain their authored declarations by constructor binding, never
// by the spelling of a method or constructor. Each relevant source is indexed
// at most once during this rule's execution.
type prototypeReferences struct {
	checker       *checker.Checker
	source        *ast.SourceFile
	scanned       map[*ast.SourceFile]bool
	writes        map[*ast.Node]bool
	members       map[prototypeMemberKey][]*ast.Node
	replacements  map[*ast.Node][]*ast.Node
	returnsObject map[*ast.Node]bool
	results       map[prototypeMemberKey]prototypeResult
}

func newPrototypeReferences(c *checker.Checker, source *ast.SourceFile) *prototypeReferences {
	return &prototypeReferences{
		checker: c, source: source,
		scanned: make(map[*ast.SourceFile]bool), writes: make(map[*ast.Node]bool),
		members: make(map[prototypeMemberKey][]*ast.Node), replacements: make(map[*ast.Node][]*ast.Node),
		returnsObject: make(map[*ast.Node]bool), results: make(map[prototypeMemberKey]prototypeResult),
	}
}

func accessParts(c *checker.Checker, node *ast.Node) (*ast.Node, string, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil, "", false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.Expression(), node.Name().Text(), true
	case ast.KindElementAccessExpression:
		argument := node.AsElementAccessExpression().ArgumentExpression
		typ := c.GetTypeAtLocation(argument)
		if typ != nil && checker.Type_flags(typ)&checker.TypeFlagsBigIntLiteral != 0 {
			// Index actual JS property keys; distinct BigInts must not share
			// the upstream PseudoBigInt display string used by the main path.
			return node.Expression(), strings.TrimSuffix(checker.ValueToString(typ.AsLiteralType().Value()), "n"), true
		}
		name, known := elementAccessPropertyName(c, argument)
		return node.Expression(), name, known
	}
	return nil, "", false
}

func (p *prototypeReferences) scan(source *ast.SourceFile) {
	if source == nil || p.scanned[source] {
		return
	}
	p.scanned[source] = true
	clear(p.results)
	if source.IsDeclarationFile {
		return
	}
	var definitions, replacements []*ast.Node
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if (node.Kind == ast.KindIdentifier || ast.IsAccessExpression(node)) && utils.IsWriteReference(node) {
			if symbol := p.checker.GetSymbolAtLocation(node); symbol != nil {
				for _, declaration := range symbol.Declarations {
					p.writes[declaration] = true
				}
			}
		}
		if ast.IsInJSFile(node) && ast.IsAssignmentExpression(node, true) {
			assignment := node.AsBinaryExpression()
			receiver, property, known := accessParts(p.checker, assignment.Left)
			if known && property == "prototype" {
				replacements = append(replacements, node)
			}
			if known {
				_, name, prototype := accessParts(p.checker, receiver)
				if prototype && name == "prototype" {
					definitions = append(definitions, node)
				}
			}
		}
		return node.ForEachChild(visit)
	}
	visit(source.AsNode())
	// Resolve bindings only after collecting writes throughout the source.
	for _, definition := range definitions {
		receiver, name, _ := accessParts(p.checker, definition.AsBinaryExpression().Left)
		constructor, _, _ := accessParts(p.checker, receiver)
		origin := p.constructor(constructor, map[*ast.Symbol]bool{})
		if origin != nil {
			key := prototypeMemberKey{origin, name}
			p.members[key] = append(p.members[key], definition)
		}
	}
	for _, replacement := range replacements {
		constructor, _, _ := accessParts(p.checker, replacement.AsBinaryExpression().Left)
		if origin := p.constructor(constructor, map[*ast.Symbol]bool{}); origin != nil {
			p.replacements[origin] = append(p.replacements[origin], replacement)
		}
	}

}

func (p *prototypeReferences) binding(node *ast.Node, seen map[*ast.Symbol]bool) *ast.Node {
	symbol := p.checker.GetSymbolAtLocation(node)
	if symbol == nil || seen[symbol] {
		return nil
	}
	seen[symbol] = true
	p.scan(ast.GetSourceFileOfNode(node))
	for _, declaration := range symbol.Declarations {
		if p.writes[declaration] {
			return nil
		}
	}
	symbol = p.checker.SkipAlias(symbol)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	p.scan(ast.GetSourceFileOfNode(declaration))
	if p.writes[declaration] {
		return nil
	}
	return declaration
}

func (p *prototypeReferences) constructor(node *ast.Node, seen map[*ast.Symbol]bool) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindClassExpression:
		return node
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		declaration := p.binding(node, seen)
		if declaration == nil {
			return nil
		}
		switch declaration.Kind {
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindFunctionExpression, ast.KindClassExpression:
			return declaration
		case ast.KindVariableDeclaration:
			return p.constructor(declaration.Initializer(), seen)
		}
	}
	return nil
}

func (p *prototypeReferences) receiver(node *ast.Node, seen map[*ast.Symbol]bool) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}
	if object, name, known := accessParts(p.checker, node); known && name == "prototype" {
		return p.constructor(object, seen)
	}
	if node.Kind == ast.KindNewExpression {
		constructor := p.constructor(node.AsNewExpression().Expression, seen)
		if constructor != nil && !p.mayReturnObject(constructor) {
			return constructor
		}
		return nil
	}
	if node.Kind == ast.KindIdentifier {
		declaration := p.binding(node, seen)
		if declaration != nil && declaration.Kind == ast.KindVariableDeclaration {
			return p.receiver(declaration.Initializer(), seen)
		}
	}
	// Actual class instance types can also establish the receiver's identity.
	// In particular, never use an `any` receiver to guess a constructor.
	typ := p.checker.GetTypeAtLocation(node)
	if typ != nil && typ.Symbol() != nil && typ.Symbol().Flags&ast.SymbolFlagsClass != 0 {
		for _, declaration := range typ.Symbol().Declarations {
			if declaration.Kind == ast.KindClassDeclaration || declaration.Kind == ast.KindClassExpression {
				return declaration
			}
		}
	}
	return nil
}

// A constructor may replace its instance with an explicitly returned object.
// Do not attribute that object's members to the constructor's prototype.
func (p *prototypeReferences) mayReturnObject(constructor *ast.Node) bool {
	if result, ok := p.returnsObject[constructor]; ok {
		return result
	}
	var body *ast.Node
	if ast.IsClassLike(constructor) {
		for _, member := range constructor.Members() {
			if member.Kind == ast.KindConstructor && member.Body() != nil {
				body = member.Body()
			}
		}
	} else {
		body = constructor.Body()
	}
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) || ast.IsClassLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			expression := ast.SkipParentheses(node.AsReturnStatement().Expression)
			if expression == nil || expression.Kind == ast.KindThisKeyword {
				return false
			}
			typ := p.checker.GetTypeAtLocation(expression)
			return typ == nil || checker.Type_flags(typ)&checker.TypeFlagsPrimitive == 0
		}
		return node.ForEachChild(visit)
	}
	result := body != nil && visit(body)
	p.returnsObject[constructor] = result
	return result
}

func (p *prototypeReferences) declarations(access *ast.Node) (prototypeMemberKey, []*ast.Node) {
	receiver, name, known := accessParts(p.checker, access)
	if !known {
		return prototypeMemberKey{}, nil
	}
	p.scan(p.source)
	origin := p.receiver(receiver, map[*ast.Symbol]bool{})
	if origin == nil {
		return prototypeMemberKey{}, nil
	}
	p.scan(ast.GetSourceFileOfNode(origin))
	key := prototypeMemberKey{origin, name}
	if _, cached := p.results[key]; cached {
		return key, nil
	}
	definitions := p.members[key]
	if len(p.replacements[origin]) == 0 {
		return key, definitions
	}
	// A replaced prototype no longer carries earlier members. Cross-file
	// replacement order is not known here, so do not guess that association.
	var retained []*ast.Node
	for _, definition := range definitions {
		invalidated := false
		for _, replacement := range p.replacements[origin] {
			if ast.GetSourceFileOfNode(replacement) != ast.GetSourceFileOfNode(definition) || replacement.Pos() > definition.Pos() {
				invalidated = true
				break
			}
		}
		if !invalidated {
			retained = append(retained, definition)
		}
	}
	return key, retained
}

// An assignment carrying its own deprecation tag documents a definition.
// Ordinary subsequent writes still refer to the deprecated member.
func (r *deprecationResolver) isDeprecatedAssignmentDefinition(access *ast.Node) bool {
	receiver, _, member := accessParts(r.checker, access)
	constructor, name, prototype := accessParts(r.checker, receiver)
	if !member || !prototype || name != "prototype" {
		return false
	}
	current := access
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || !ast.IsInJSFile(parent) || !ast.IsAssignmentExpression(parent, true) || parent.AsBinaryExpression().Left != current {
		return false
	}
	deprecated, _ := r.declarationDeprecation(parent)
	return deprecated && r.prototypes != nil && r.prototypes.constructor(constructor, map[*ast.Symbol]bool{}) != nil
}

func (r *deprecationResolver) prototypeDeprecation(access *ast.Node) (bool, string) {
	if r.prototypes == nil {
		return false, ""
	}
	key, declarations := r.prototypes.declarations(access)
	if key.constructor == nil {
		return false, ""
	}
	if info, cached := r.prototypes.results[key]; cached {
		return info.deprecated, info.reason
	}
	var info prototypeResult
	for _, declaration := range declarations {
		if deprecated, reason := r.declarationDeprecation(declaration); deprecated {
			info.deprecated, info.reason, info.declaration = true, reason, declaration
			break
		}
	}
	r.prototypes.results[key] = info
	return info.deprecated, info.reason
}

// Use the actual assigned value's type for value-based package allowances.
func (r *deprecationResolver) prototypeValueType(access *ast.Node) *checker.Type {
	if r.prototypes == nil {
		return nil
	}
	key, _ := r.prototypes.declarations(access)
	result := r.prototypes.results[key]
	if result.declaration == nil {
		return nil
	}
	return r.checker.GetTypeAtLocation(result.declaration.AsBinaryExpression().Right)
}
