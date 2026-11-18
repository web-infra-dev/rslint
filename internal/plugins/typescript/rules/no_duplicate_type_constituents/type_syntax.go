package no_duplicate_type_constituents

import (
	"hash/maphash"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// typeSyntaxIndex lives for one source file. Hashes only select candidates;
// equality always checks the syntax, including in the event of a collision.
type typeSyntaxIndex struct {
	source *ast.SourceFile
	seed   maphash.Seed
	hashes map[*ast.Node]uint64
}

func newTypeSyntaxIndex(source *ast.SourceFile) *typeSyntaxIndex {
	return &typeSyntaxIndex{source: source, seed: maphash.MakeSeed(), hashes: make(map[*ast.Node]uint64)}
}

func syntaxNode(node *ast.Node) *ast.Node {
	return ast.SkipTypeParentheses(ast.SkipParentheses(node))
}

// Child traversal omits nil fields and joins lists. The header preserves field
// presence and list boundaries so that, for example, a mapped key cannot match
// a value type. Both hashing and equality use this same representation.
type typeSyntaxHeader struct {
	identity   *ast.Node
	kind       ast.Kind
	text       string
	operator   ast.Kind
	fields     uint64
	modifiers  int
	typeParams int
	params     int
	typeArgs   int
}

func fieldPresence(nodes ...*ast.Node) uint64 {
	var mask uint64
	for i, node := range nodes {
		if node != nil {
			mask |= 1 << i
		}
	}
	return mask
}

func syntaxListLength(list *ast.NodeList) int {
	if list == nil {
		return -1
	}
	return len(list.Nodes)
}

func (s *typeSyntaxIndex) header(node *ast.Node) typeSyntaxHeader {
	h := typeSyntaxHeader{kind: node.Kind}
	if modifiers := node.Modifiers(); modifiers != nil {
		h.modifiers = len(modifiers.Nodes)
	}
	if fn := node.FunctionLikeData(); fn != nil {
		h.typeParams = syntaxListLength(fn.TypeParameters)
		h.params = syntaxListLength(fn.Parameters)
		h.fields = fieldPresence(fn.Type, node.Body())
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindRegularExpressionLiteral,
		ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
		r := utils.TrimNodeTextRange(s.source, node)
		h.text = s.source.Text()[r.Pos():r.End()]
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		h.text = node.Text()
	case ast.KindMappedType:
		n := node.AsMappedTypeNode()
		h.fields = fieldPresence(n.ReadonlyToken, n.TypeParameter, n.NameType, n.QuestionToken, n.Type)
	case ast.KindTypeParameter:
		n := node.AsTypeParameterDeclaration()
		h.fields = fieldPresence(n.Constraint, n.Expression, n.DefaultType)
	case ast.KindTypePredicate:
		n := node.AsTypePredicateNode()
		h.fields = fieldPresence(n.AssertsModifier, n.ParameterName, n.Type)
	case ast.KindParameter:
		n := node.AsParameterDeclaration()
		h.fields = fieldPresence(n.DotDotDotToken, n.Name(), n.QuestionToken, n.Type, n.Initializer)
	case ast.KindPropertySignature:
		n := node.AsPropertySignatureDeclaration()
		h.fields = fieldPresence(n.Name(), n.PostfixToken, n.Type, n.Initializer)
	case ast.KindNamedTupleMember:
		n := node.AsNamedTupleMember()
		h.fields = fieldPresence(n.DotDotDotToken, n.Name(), n.QuestionToken, n.Type)
	case ast.KindImportType:
		n := node.AsImportTypeNode()
		h.fields = fieldPresence(n.Argument, n.Attributes, n.Qualifier)
		if n.IsTypeOf {
			h.fields |= 1 << 3
		}
		h.typeArgs = syntaxListLength(n.TypeArguments)
	case ast.KindTypeReference, ast.KindTypeQuery:
		h.typeArgs = syntaxListLength(node.TypeArgumentList())
	case ast.KindTypeOperator:
		h.operator = node.AsTypeOperatorNode().Operator
		// Each authored unique-symbol type introduces its own nominal
		// identity. Identical spelling cannot establish equivalence between
		// separate declarations, including inside object or function types.
		if h.operator == ast.KindUniqueKeyword {
			h.identity = node
		}
	case ast.KindPrefixUnaryExpression:
		h.operator = node.AsPrefixUnaryExpression().Operator
	case ast.KindPostfixUnaryExpression:
		h.operator = node.AsPostfixUnaryExpression().Operator
	case ast.KindMetaProperty:
		h.operator = node.AsMetaProperty().KeywordToken
	}
	return h
}

func (s *typeSyntaxIndex) hash(node *ast.Node) uint64 {
	node = syntaxNode(node)
	if hash, ok := s.hashes[node]; ok {
		return hash
	}
	hash := maphash.Comparable(s.seed, s.header(node))
	var count uint64
	node.ForEachChild(func(child *ast.Node) bool {
		hash = (hash ^ s.hash(child)) * 1099511628211
		count++
		return false
	})
	hash = (hash ^ count) * 1099511628211
	s.hashes[node] = hash
	return hash
}

func (s *typeSyntaxIndex) find(candidates []*ast.Node, node *ast.Node) *ast.Node {
	for _, previous := range candidates {
		if s.equal(previous, node) {
			return previous
		}
	}
	return nil
}

func (s *typeSyntaxIndex) equal(left, right *ast.Node) bool {
	left, right = syntaxNode(left), syntaxNode(right)
	if left == right {
		return true
	}
	if s.header(left) != s.header(right) {
		return false
	}
	// Most type nodes fit in the stack buffers. Large member lists allocate
	// only after the hash has selected an actual comparison candidate.
	var leftBuffer, rightBuffer [8]*ast.Node
	a, b := leftBuffer[:0], rightBuffer[:0]
	left.ForEachChild(func(node *ast.Node) bool { a = append(a, node); return false })
	right.ForEachChild(func(node *ast.Node) bool { b = append(b, node); return false })
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !s.equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
