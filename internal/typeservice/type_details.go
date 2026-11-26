package typeservice

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/program"
)

// TypeDetails holds immutable structural metadata for specialized types
// (generic interfaces, tuples, template literals, and conditional roots).
type TypeDetails struct {
	rootNode                          *ast.Node
	labels                            []*ast.Node
	rootRef                           *NodeRef
	labelRefs                         []*NodeRef
	outer, local, elements            Handles
	minPlusOne, fixedPlusOne          uint32
	texts                             []string
	rootCheck, rootExtends, rootAlias uint32
	distributive                      bool
	infer, rootOuter, rootArguments   Handles
}

func (*TypeDetails) result() {}

func (s *Session) typeDetails(t *checker.Type) *TypeDetails {
	d := &TypeDetails{}
	flags := t.Flags()
	if def := t.AsInterfaceType(); def != nil {
		d.outer = s.typeIDs(def.OuterTypeParameters())
		d.local = s.typeIDs(def.LocalTypeParameters())
		if t.ObjectFlags()&checker.ObjectFlagsTuple != 0 {
			tuple := t.AsTupleType()
			d.fixedPlusOne = uint32(tuple.FixedLength() + 1)
			d.minPlusOne = uint32(checker.TupleType_minLength(tuple) + 1)
			d.elements = make(Handles, len(tuple.ElementInfos()))
			d.labels = make([]*ast.Node, len(d.elements))
			for i, el := range tuple.ElementInfos() {
				d.elements[i] = uint32(el.TupleElementFlags())
				d.labels[i] = el.LabeledDeclaration()
			}
		}
	}
	switch {
	case flags&checker.TypeFlagsTemplateLiteral != 0:
		d.texts = t.AsTemplateLiteralType().Texts()
	case flags&checker.TypeFlagsConditional != 0:
		root := checker.ConditionalType_root(t.AsConditionalType())
		d.rootNode = checker.ConditionalRoot_node(root).AsNode()
		d.rootCheck = s.typeID(checker.ConditionalRoot_checkType(root))
		d.rootExtends = s.typeID(checker.ConditionalRoot_extendsType(root))
		d.infer = s.typeIDs(checker.ConditionalRoot_inferTypeParameters(root))
		d.rootOuter = s.typeIDs(checker.ConditionalRoot_outerTypeParameters(root))
		d.distributive = checker.ConditionalRoot_isDistributive(root)
		alias := checker.ConditionalRoot_alias(root)
		d.rootAlias, d.rootArguments = s.symbolID(alias.Symbol()), s.typeIDs(alias.TypeArguments())
	}
	return d
}

// declaredIndexInfos returns an interface/tuple's own index signatures once
// resolved, without forcing resolution or including inherited index infos.
func (s *Session) declaredIndexInfos(t *checker.Type) OptionalIndexInfos {
	def := t.AsInterfaceType()
	if def == nil || !checker.InterfaceType_declaredMembersResolved(def) {
		return OptionalIndexInfos{}
	}
	infos := checker.InterfaceType_declaredIndexInfos(def)
	result := make(IndexInfos, len(infos))
	for i, info := range infos {
		result[i] = s.indexInfo(info)
	}
	return OptionalIndexInfos{Present: true, Infos: result}
}

func nodeRefs(p *program.Program, nodes []*ast.Node) []*NodeRef {
	if nodes == nil {
		return nil
	}
	refs := make([]*NodeRef, len(nodes))
	for i, n := range nodes {
		refs[i] = nodeReference(p, n)
	}
	return refs
}
