package typesnapshot

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildFixtureSnapshot(t *testing.T) Snapshot {
	return buildSnapshotFromFixture(t, "fixture.ts")
}

func buildSnapshotFromFixture(t *testing.T, filename string) Snapshot {
	t.Helper()
	dir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	currentDirectory := tspath.NormalizePath(dir)
	host := utils.CreateCompilerHost(currentDirectory, fs)
	program, err := utils.CreateProgram(true, fs, currentDirectory, filepath.Join(dir, "tsconfig.json"), host)
	if err != nil {
		t.Fatalf("create program: %v", err)
	}
	tc, done := program.GetTypeChecker(context.Background())
	defer done()
	var file *ast.SourceFile
	for _, f := range program.GetSourceFiles() {
		if strings.HasSuffix(f.FileName(), filename) {
			file = f
			break
		}
	}
	if file == nil {
		t.Fatalf("%s not found in program", filename)
	}
	return Build(tc, file)
}

func containsID(ids []TypeID, id TypeID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func findByName(s Snapshot, name string) (TypeBlock, bool) {
	for _, b := range s.Types {
		if b.Name == name {
			return b, true
		}
	}
	return TypeBlock{}, false
}

// TestNode2TypeResolvesSpanCollisions pins the M1 fix: real code produces
// (tokenStart,end) span collisions (a for-of's VariableDeclaration shares its
// span with the inner ArrayBindingPattern; an `extends Base<T>` EWTA with its
// Identifier). Build must dedup them — keeping the DEEPER node — so the worker's
// binary search never picks an entry arbitrarily.
func TestNode2TypeResolvesSpanCollisions(t *testing.T) {
	s := buildSnapshotFromFixture(t, "collision.ts")

	// 1. Every (tokenStart,end) key is unique — no surviving collision.
	seen := map[[2]int]TypeID{}
	for _, e := range s.Node2Type {
		key := [2]int{e.TokenStart, e.End}
		if prev, dup := seen[key]; dup {
			t.Errorf("unresolved span collision at (%d,%d): typeIds %d and %d", e.TokenStart, e.End, prev, e.TypeID)
		}
		seen[key] = e.TypeID
	}

	// 2. The for-of `[k, v]` resolves to its tuple type (the deeper
	//    ArrayBindingPattern), NOT the wrapper VariableDeclaration's `any`. So a
	//    node2type entry must point at a tuple; had dedup kept the shallow wrapper,
	//    the [k,v] entry would be `any` and no tuple would anchor any node.
	foundTuple := false
	for _, e := range s.Node2Type {
		if b, ok := s.Types[e.TypeID]; ok && b.IsTuple {
			foundTuple = true
			break
		}
	}
	if !foundTuple {
		t.Error("for-of [k,v] tuple type missing from node2type — dedup kept the `any` wrapper over the deeper ArrayBindingPattern")
	}

	// 3. bare `extends Base` → `typeof Base` (the deeper Identifier = ts-eslint's
	//    superClass anchor), NOT the EWTA's instance type `Base`. The deeper
	//    Identifier must win, else the EWTA/Identifier collision picked wrong.
	foundTypeofBase := false
	for _, e := range s.Node2Type {
		if b, ok := s.Types[e.TypeID]; ok && b.Name == "typeof Base" {
			foundTypeofBase = true
			break
		}
	}
	if !foundTypeofBase {
		var names []string
		for _, e := range s.Node2Type {
			if b, ok := s.Types[e.TypeID]; ok {
				names = append(names, b.Name)
			}
		}
		t.Errorf("extends `typeof Base` missing from node2type — dedup kept the EWTA instance type over the deeper Identifier; node2type type names: %v", names)
	}
}

func TestBuildSnapshot(t *testing.T) {
	s := buildFixtureSnapshot(t)

	// Dump exported names once (visible with -v) to ease debugging name asserts.
	names := make([]string, 0, len(s.Types))
	for _, b := range s.Types {
		names = append(names, b.Name)
	}
	t.Logf("undefined id=%d; exported type names: %v", s.PrimTypes.Undefined, names)

	undef := s.PrimTypes.Undefined
	if undef == 0 {
		t.Fatal("PrimTypes.Undefined is 0")
	}
	if b, ok := s.Types[undef]; !ok || b.Name != "undefined" {
		t.Fatalf("undefined type block missing/misnamed: ok=%v block=%+v", ok, s.Types[undef])
	}

	// const a: undefined -> at least one node resolves directly to undefined.
	direct := 0
	for _, e := range s.Node2Type {
		if e.TypeID == undef {
			direct++
		}
	}
	if direct == 0 {
		t.Error("no node resolved directly to undefined type")
	}

	// c: string | undefined -> union members contain undefined.
	if b, ok := findByName(s, "string | undefined"); !ok {
		t.Error("union type 'string | undefined' not exported")
	} else if !containsID(b.MemberTypes, undef) {
		t.Errorf("union members %v missing undefined(%d)", b.MemberTypes, undef)
	}

	// e: { a: number } & { b: string } -> intersection members exported. The
	// adapter exposes isIntersection(); its members must have data (ty.Types()
	// covers intersections via TypeFlagsUnionOrIntersection).
	intersectionFound := false
	for _, b := range s.Types {
		if strings.Contains(b.Name, "&") && b.Flags&(1<<28) != 0 { // 1<<28 = tsgo Intersection
			intersectionFound = true
			if len(b.MemberTypes) == 0 {
				t.Errorf("intersection %q exported with empty memberTypes", b.Name)
			}
		}
	}
	if !intersectionFound {
		t.Error("no intersection type exported from fixture")
	}

	// d: undefined[] -> marked IsArray, typeArgs contain undefined.
	if b, ok := findByName(s, "undefined[]"); !ok {
		t.Error("array type 'undefined[]' not exported")
	} else if !b.IsArray {
		t.Error("'undefined[]' not marked IsArray")
	} else if !containsID(b.TypeArgs, undef) {
		t.Errorf("array typeArgs %v missing undefined(%d)", b.TypeArgs, undef)
	}

	// g(): undefined -> callSigReturns contain undefined.
	if b, ok := findByName(s, "() => undefined"); !ok {
		t.Error("function type '() => undefined' not exported")
	} else if !containsID(b.CallSigReturns, undef) {
		t.Errorf("fn callSigReturns %v missing undefined(%d)", b.CallSigReturns, undef)
	}

	// No (tokenStart, end) key may map to two DIFFERENT type-ids — the invariant
	// the worker relies on (it matches an oxc node by (range[0], range[1]) and
	// must get the one right type). The fixture deliberately includes a generic
	// call `id(undefined)` + named-type references, which produce type-annotation
	// nodes that share a span with their inner Identifier (resolving to `any`);
	// Build skips those (ast.IsPartOfTypeNode), so this asserts the skip actually
	// eliminates the collision.
	seen := map[[2]int]TypeID{}
	for _, e := range s.Node2Type {
		k := [2]int{e.TokenStart, e.End}
		if prev, dup := seen[k]; dup && prev != e.TypeID {
			t.Errorf("node2type key (tokenStart=%d,end=%d) maps to two type-ids %d and %d", e.TokenStart, e.End, prev, e.TypeID)
		}
		seen[k] = e.TypeID
	}

	// Transitive closure: every referenced type-id must be present in Types.
	for _, b := range s.Types {
		refs := append(append(append([]TypeID{}, b.MemberTypes...), b.TypeArgs...), b.CallSigReturns...)
		for _, ref := range refs {
			if _, ok := s.Types[ref]; !ok {
				t.Errorf("dangling type-id %d referenced by %q(%d)", ref, b.Name, b.ID)
			}
		}
	}
}

// TestNode2TypeIdentifierKeyUsesDecodedNameLength pins the escaped-identifier
// span fix: `esc` has a DECODED name "esc" (3 UTF-16 units) but an 8-char
// SOURCE token. oxc keys the worker lookup on range[0]+name.length (the decoded
// length), so the node2type entry must end at tokenStart+3, NOT tokenStart+8
// (tsgo's source End) — else the worker misses an escaped binding. snapshot.go
// ends an Identifier key at the decoded name length to match.
func TestNode2TypeIdentifierKeyUsesDecodedNameLength(t *testing.T) {
	s := buildSnapshotFromFixture(t, "escaped.ts")
	const tsgoUnion = 134217728 // 1<<27, raw tsgo TypeFlagsUnion
	sawDecodedWidth := false
	for _, e := range s.Node2Type {
		blk, ok := s.Types[e.TypeID]
		if !ok || blk.Flags&tsgoUnion == 0 || len(blk.MemberTypes) != 2 {
			continue // only the string|undefined binding/init is a 2-member union
		}
		switch e.End - e.TokenStart {
		case 3:
			sawDecodedWidth = true
		case 8:
			t.Errorf("escaped identifier keyed on SOURCE width 8, not decoded width 3 (span regressed): %+v", e)
		}
	}
	if !sawDecodedWidth {
		t.Error("no string|undefined entry with decoded width 3 — escaped binding `\\u0065sc` not keyed on its decoded name length")
	}
}
