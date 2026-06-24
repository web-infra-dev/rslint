// Package typesnapshot builds a random-access type snapshot for a single
// source file: a node→type-id table keyed by (tokenStart, end) plus the
// transitive closure of the referenced type blocks. It is shipped to the
// Node worker so third-party type-aware ESLint plugin rules can reconstruct
// `parserServices` (program/checker/getTypeAtLocation) without a second
// TypeScript implementation.
//
// node2type key = (tokenStart, end), both UTF-16. tokenStart matches an oxc ESTree
// node's range[0] (oxc.start == tsgo GetTokenPosOfNode). The END does NOT always
// line up: oxc includes a type annotation in a binding/param Identifier's range[1],
// but tsgo's Identifier End() stops at the name — so the worker's getTypeAtLocation
// strips the annotation off the range before lookup (see parser-services-from-snapshot.ts).
// Type-ANNOTATION nodes are intentionally SKIPPED
// (ast.IsPartOfTypeNode): a TypeReference and its inner type-name Identifier
// occupy the SAME span but resolve to different type-ids (the inner Identifier
// to `any`), which would collide on the (tokenStart,end) key. Type-aware rules
// only anchor on VALUE nodes, never on annotation nodes, so skipping annotations
// both removes that collision and shrinks the table; across the remaining
// value/declaration nodes the (tokenStart,end) key is unique.
//
// The walk/record algorithm mirrors cmd/tsgo/semantic.go (the CBOR CLI
// exporter); kept separate for now because that one lives in a main package.
//
// Depth policy: full transitive closure, no bounded depth — bounded by
// construction. recordType records each type-id at most once (the s.Types[id]
// seen guard) and writes a placeholder before recursing, so reference cycles
// terminate and the closure size is bounded by the program's distinct type
// count, not by node count. (tsgo additionally interns structurally-identical
// types to one id and flattens union-of-union into a single member list, so
// unions don't multiply.) The volume driver is node2type instead — one entry
// per value node, linear in node count — addressed by the binary wire (M1; see
// encode_binary.go), not by the type closure, so no depth/size cap is needed.
// TODO(type-aware): unify with cmd/tsgo/semantic.go once the binary wire format
// stabilizes.
package typesnapshot

import (
	"unicode/utf16"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// TypeID is a tsgo type id. It is only stable within a single checker, which is
// why Build must use ONE consistent checker for every file (see Build's CONTRACT
// — CLI checkers[0], LSP query checker).
type TypeID = int

// Snapshot is the per-file type data Build produces. The production wire to the
// worker is EncodeBinary (a compact random-access binary buffer; see
// encode_binary.go) — NOT encoding/json. The json tags below are retained only
// for ad-hoc debug dumps; nothing marshals Snapshot to JSON on the hot path.
type Snapshot struct {
	Node2Type []NodeTypeEntry      `json:"node2type"`
	Types     map[TypeID]TypeBlock `json:"types"`
	PrimTypes PrimTypes            `json:"primTypes"`
}

// NodeTypeEntry maps one value/declaration node to its type. The worker matches
// an oxc ESTree node by (tokenStart == range[0], end == range[1]).
type NodeTypeEntry struct {
	TokenStart int    `json:"tokenStart"` // UTF-16 (== oxc node.range[0])
	End        int    `json:"end"`        // UTF-16 (== oxc node.range[1])
	TypeID     TypeID `json:"typeId"`
}

// TypeBlock is one type's exported shape. memberTypes/typeArgs/callSigReturns
// reference other type-ids that are guaranteed to also be present in the
// snapshot (transitive closure).
type TypeBlock struct {
	ID             TypeID   `json:"id"`
	Flags          int      `json:"flags"`                 // raw tsgo TypeFlags (worker maps to runtime ts)
	Name           string   `json:"name"`                  // checker.TypeToString
	IsArray        bool     `json:"isArray,omitempty"`     // checker.isArrayType — element type(s) in TypeArgs
	IsTuple        bool     `json:"isTuple,omitempty"`     // checker.IsTupleType — element type(s) in TypeArgs
	MemberTypes    []TypeID `json:"memberTypes,omitempty"` // union OR intersection members (ty.Types())
	TypeArgs       []TypeID `json:"typeArgs,omitempty"`
	CallSigReturns []TypeID `json:"callSigReturns,omitempty"`
}

// PrimTypes carries the well-known intrinsic type-ids the rules compare
// against (e.g. `type === checker.getUndefinedType()`).
type PrimTypes struct {
	Undefined TypeID `json:"undefined"`
	Null      TypeID `json:"null"`
}

// Build walks every node of file and produces a snapshot.
//
// CONTRACT: tc MUST be a SINGLE checker, used consistently (type-ids are
// per-checker, so mixing checkers would corrupt the snapshot's id references).
// The CLI passes checkers[0] (program.GetTypeChecker); the LSP passes the project
// pool's default query checker — either works, the snapshot is self-contained
// within one checker. Build's access must not race a concurrent read of the SAME
// checker: the CLI runs Build serially BEFORE the native pass (checkers[0] access
// is UNLOCKED — getCheckerNonExclusive returns a noop release); the LSP runs it on
// the main dispatch loop where the project pool's GetChecker holds the checker
// LOCKED for the whole Build. Both guarantee no concurrent read of the in-use checker.
func Build(tc *checker.Checker, file *ast.SourceFile) Snapshot {
	s := Snapshot{Types: map[TypeID]TypeBlock{}}
	if tc == nil || file == nil {
		return s
	}
	s.PrimTypes = PrimTypes{
		Undefined: int(tc.GetUndefinedType().Id()),
		Null:      int(tc.GetNullType().Id()),
	}
	positionMap := file.GetPositionMap()

	// recordType records ty (and its transitive references) into s.Types and
	// returns its id. A placeholder is inserted before recursing so cyclic
	// references terminate. Members that resolve to id 0 (nil type) are dropped
	// so the worker never sees a dangling 0 in a member list.
	var recordType func(ty *checker.Type) TypeID
	recordType = func(ty *checker.Type) TypeID {
		if ty == nil {
			return 0
		}
		id := int(ty.Id())
		if _, ok := s.Types[id]; ok {
			return id
		}
		block := TypeBlock{ID: id, Flags: int(ty.Flags()), Name: tc.TypeToString(ty)}
		block.IsArray = checker.Checker_isArrayType(tc, ty)
		block.IsTuple = checker.IsTupleType(ty)
		s.Types[id] = block // placeholder to break recursion

		// Union OR intersection members — ty.Types() is valid for both
		// (TypeFlagsUnionOrIntersection); the worker's isUnion()/isIntersection()
		// both read memberTypes.
		if ty.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
			for _, m := range ty.Types() {
				if mid := recordType(m); mid != 0 {
					block.MemberTypes = append(block.MemberTypes, mid)
				}
			}
		}
		if checker.Type_objectFlags(ty)&checker.ObjectFlagsReference != 0 {
			for _, a := range checker.Checker_getTypeArguments(tc, ty) {
				if aid := recordType(a); aid != 0 {
					block.TypeArgs = append(block.TypeArgs, aid)
				}
			}
		}
		for _, sig := range tc.GetCallSignatures(ty) {
			if rt := checker.Checker_getReturnTypeOfSignature(tc, sig); rt != nil {
				if rid := recordType(rt); rid != 0 {
					block.CallSigReturns = append(block.CallSigReturns, rid)
				}
			}
		}

		s.Types[id] = block // finalize with closure filled in
		return id
	}

	// node2type is collected into a map keyed by (tokenStart, end) so a span
	// COLLISION resolves deterministically, instead of emitting two entries the
	// worker's binary search would pick between arbitrarily. Collisions are real
	// on production code: a for-of's VariableDeclaration shares its span with the
	// inner BindingPattern; an `extends X<T>` ExpressionWithTypeArguments shares it
	// with the inner Identifier; a namespace ImportClause with its NamespaceImport.
	// On collision keep the DEEPER node — typescript-eslint maps an ESTree anchor
	// to the most specific TS node, which is always the deeper one
	// (ArrayPattern→ArrayBindingPattern, superClass→Identifier,
	// ImportNamespaceSpecifier→NamespaceImport), so the deeper node carries the
	// type a rule actually queries; the shallower wrapper usually resolves to an
	// `any`/typeof that is never the anchor.
	type n2tRecord struct {
		typeID TypeID
		depth  int
	}
	byKey := map[[2]int]n2tRecord{}
	var visit func(node *ast.Node, depth int)
	visit = func(node *ast.Node, depth int) {
		if node == nil {
			return
		}
		// Record value/declaration nodes only. Skip synthetic nodes, type
		// declarations (GetTypeAtLocation panics on them), and type-annotation
		// nodes (their span collides with the inner Identifier — see package doc).
		if node.Pos() >= 0 && node.End() >= 0 &&
			!ast.IsTypeDeclaration(node) && !ast.IsPartOfTypeNode(node) {
			if ty := tc.GetTypeAtLocation(node); ty != nil {
				if id := recordType(ty); id != 0 {
					tokenStart := positionMap.UTF8ToUTF16(scanner.GetTokenPosOfNode(node, file, false))
					end := positionMap.UTF8ToUTF16(node.End())
					// oxc gives an Identifier its DECODED name and keys the worker
					// lookup on range[0]+name.length (UTF-16 of the decoded name).
					// tsgo's End() is the SOURCE end, which for an ESCAPED identifier
					// (e.g. `esc`) runs past the decoded name. End the key at the
					// decoded length so both sides agree on every identifier — escaped
					// or not. For a non-escaped identifier this equals End() (no-op).
					if node.Kind == ast.KindIdentifier {
						end = tokenStart + len(utf16.Encode([]rune(node.AsIdentifier().Text)))
					}
					key := [2]int{tokenStart, end}
					if existing, ok := byKey[key]; !ok || depth > existing.depth {
						byKey[key] = n2tRecord{typeID: id, depth: depth}
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, depth+1)
			return false
		})
	}
	visit(file.AsNode(), 0)
	for key, rec := range byKey {
		s.Node2Type = append(s.Node2Type, NodeTypeEntry{
			TokenStart: key[0],
			End:        key[1],
			TypeID:     rec.typeID,
		})
	}
	return s
}
