package no_restricted_syntax

import (
	"math"
	"math/big"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/languages/js/estree"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

var (
	isDeclarationType []bool
	isStatementType   []bool
	isFunctionType    []bool
	isExpressionType  []bool
	isPatternType     []bool

	typeIdentifierID   estree.TypeID
	typeMetaPropertyID estree.TypeID
)

func init() {
	count := estree.TypeCount()
	isDeclarationType = make([]bool, count)
	isStatementType = make([]bool, count)
	isFunctionType = make([]bool, count)
	isExpressionType = make([]bool, count)
	isPatternType = make([]bool, count)

	typeIdentifierID, _ = estree.TypeIDByName("Identifier")
	typeMetaPropertyID, _ = estree.TypeIDByName("MetaProperty")

	for i := range count {
		name := estree.TypeID(i).String()
		decl := strings.HasSuffix(name, "Declaration")
		isDeclarationType[i] = decl
		isStatementType[i] = strings.HasSuffix(name, "Statement") || decl
		isFunctionType[i] = name == "FunctionDeclaration" ||
			name == "FunctionExpression" ||
			name == "ArrowFunctionExpression"
		isExpressionType[i] = strings.HasSuffix(name, "Expression") ||
			strings.HasSuffix(name, "Literal") ||
			name == "MetaProperty"
		isPatternType[i] = strings.HasSuffix(name, "Pattern")
	}
}

type memoKey struct {
	id    uint32
	node  estree.NodeID
	scope estree.NodeID
}

type evalMemo struct {
	cache map[memoKey]bool
}

func (m *evalMemo) get(id uint32, node, scope estree.NodeID) (bool, bool) {
	if m == nil || m.cache == nil {
		return false, false
	}
	v, ok := m.cache[memoKey{id: id, node: node, scope: scope}]
	return v, ok
}

func (m *evalMemo) set(id uint32, node, scope estree.NodeID, value bool) {
	if m == nil {
		return
	}
	if m.cache == nil {
		m.cache = make(map[memoKey]bool)
	}
	m.cache[memoKey{id: id, node: node, scope: scope}] = value
}

func matchesSelector(sel selector, tree *estree.Tree, id, scopeRoot estree.NodeID, memo *evalMemo) bool {
	if id == estree.NoNode {
		return false
	}
	switch v := sel.(type) {
	case subjectSelector:
		return matchesSelector(v.Inner, tree, id, scopeRoot, memo)
	case identifierSelector:
		if v.Wildcard {
			return true
		}
		return v.TypeID != estree.UnknownType && tree.Type(id) == v.TypeID
	case compoundSelector:
		return matchesSelector(v.Right, tree, id, scopeRoot, memo) &&
			matchesSelector(v.Left, tree, id, scopeRoot, memo)
	case classSelector:
		if !matchesSelector(v.Inner, tree, id, scopeRoot, memo) {
			return false
		}
		return matchesClass(tree, id, v.Path, scopeRoot)
	case attrSelector:
		if !matchesSelector(v.Inner, tree, id, scopeRoot, memo) {
			return false
		}
		return matchesAttr(v, tree, id)
	case combinatorSelector:
		return matchesCombinator(v, tree, id, scopeRoot, memo)
	case combinedPseudo:
		return matchesSelector(v.Inner, tree, id, scopeRoot, memo) &&
			matchesPseudo(v.Pseudo, tree, id, scopeRoot, memo)
	case pseudoSelector:
		return matchesPseudo(v, tree, id, scopeRoot, memo)
	case unionSelector:
		for _, child := range v.Selectors {
			if matchesSelector(child, tree, id, scopeRoot, memo) {
				return true
			}
		}
		return false
	}
	return false
}

func parentInScope(tree *estree.Tree, id, scopeRoot estree.NodeID) estree.NodeID {
	if id == estree.NoNode || (scopeRoot != estree.NoNode && id == scopeRoot) {
		return estree.NoNode
	}
	return tree.Parent(id)
}

func matchesCombinator(c combinatorSelector, tree *estree.Tree, id, scopeRoot estree.NodeID, memo *evalMemo) bool {
	switch c.Kind {
	case combChild:
		if !matchesSelector(c.Right, tree, id, scopeRoot, memo) {
			return false
		}
		parent := parentInScope(tree, id, scopeRoot)
		return parent != estree.NoNode && matchesSelector(c.Left, tree, parent, scopeRoot, memo)
	case combDescendant:
		if !matchesSelector(c.Right, tree, id, scopeRoot, memo) {
			return false
		}
		return matchesDescendantAncestor(c, tree, id, scopeRoot, memo)
	case combAdjacent, combSibling:
		if scopeRoot != estree.NoNode && id == scopeRoot {
			return false
		}
		siblings, idx, ok := tree.SiblingList(id)
		if !ok {
			return false
		}
		if idx > 0 && matchesSelector(c.Right, tree, id, scopeRoot, memo) {
			if c.Kind == combAdjacent {
				// Sparse array holes (NoNode) break adjacency.
				if prev := siblings.At(idx - 1); prev != estree.NoNode && matchesSelector(c.Left, tree, prev, scopeRoot, memo) {
					return true
				}
			} else if matchesPrevSibling(c, tree, siblings, idx, scopeRoot, memo) {
				return true
			}
		}
		// Reverse-direction branches mirror esquery 1.7.0's subject handling.
		if c.Kind == combSibling && isSubjectSelector(c.Left) && matchesSelector(c.Left, tree, id, scopeRoot, memo) {
			if matchesNextSibling(c, tree, siblings, idx, scopeRoot, memo) {
				return true
			}
		}
		if c.Kind == combAdjacent && isSubjectSelector(c.Right) && matchesSelector(c.Left, tree, id, scopeRoot, memo) && idx+1 < siblings.Len() {
			if next := siblings.At(idx + 1); next != estree.NoNode && matchesSelector(c.Right, tree, next, scopeRoot, memo) {
				return true
			}
		}
	}
	return false
}

func matchesDescendantAncestor(c combinatorSelector, tree *estree.Tree, id, scopeRoot estree.NodeID, memo *evalMemo) bool {
	if memo == nil || c.memoID == 0 {
		for parent := parentInScope(tree, id, scopeRoot); parent != estree.NoNode; parent = parentInScope(tree, parent, scopeRoot) {
			if matchesSelector(c.Left, tree, parent, scopeRoot, memo) {
				return true
			}
		}
		return false
	}
	var pathBuf [16]estree.NodeID
	visited := pathBuf[:0]
	matched := false
	for parent := parentInScope(tree, id, scopeRoot); parent != estree.NoNode; parent = parentInScope(tree, parent, scopeRoot) {
		if hit, ok := memo.get(c.memoID, parent, scopeRoot); ok {
			matched = hit
			break
		}
		visited = append(visited, parent)
		if matchesSelector(c.Left, tree, parent, scopeRoot, memo) {
			matched = true
			break
		}
	}
	for _, node := range visited {
		memo.set(c.memoID, node, scopeRoot, matched)
	}
	return matched
}

func matchesPrevSibling(c combinatorSelector, tree *estree.Tree, siblings estree.List, idx int, scopeRoot estree.NodeID, memo *evalMemo) bool {
	if memo == nil || c.memoID == 0 {
		for i := idx - 1; i >= 0; i-- {
			if prev := siblings.At(i); prev != estree.NoNode && matchesSelector(c.Left, tree, prev, scopeRoot, memo) {
				return true
			}
		}
		return false
	}
	var visitedBuf [16]estree.NodeID
	visited := visitedBuf[:0]
	matched := false
	for i := idx - 1; i >= 0; i-- {
		prev := siblings.At(i)
		if prev == estree.NoNode {
			continue
		}
		if hit, ok := memo.get(c.memoID, prev, scopeRoot); ok {
			matched = hit
			break
		}
		visited = append(visited, prev)
		if matchesSelector(c.Left, tree, prev, scopeRoot, memo) {
			matched = true
			break
		}
	}
	for _, node := range visited {
		memo.set(c.memoID, node, scopeRoot, matched)
	}
	return matched
}

func matchesNextSibling(c combinatorSelector, tree *estree.Tree, siblings estree.List, idx int, scopeRoot estree.NodeID, memo *evalMemo) bool {
	if memo == nil || c.memoID == 0 {
		for i := idx + 1; i < siblings.Len(); i++ {
			if next := siblings.At(i); next != estree.NoNode && matchesSelector(c.Right, tree, next, scopeRoot, memo) {
				return true
			}
		}
		return false
	}
	revID := c.memoID + 1
	var visitedBuf [16]estree.NodeID
	visited := visitedBuf[:0]
	matched := false
	for i := idx + 1; i < siblings.Len(); i++ {
		next := siblings.At(i)
		if next == estree.NoNode {
			continue
		}
		if hit, ok := memo.get(revID, next, scopeRoot); ok {
			matched = hit
			break
		}
		visited = append(visited, next)
		if matchesSelector(c.Right, tree, next, scopeRoot, memo) {
			matched = true
			break
		}
	}
	for _, node := range visited {
		memo.set(revID, node, scopeRoot, matched)
	}
	return matched
}

func isSubjectSelector(sel selector) bool {
	_, ok := sel.(subjectSelector)
	return ok
}

func hasReverseSiblingSubject(sel combinatorSelector) bool {
	return sel.Kind == combSibling && isSubjectSelector(sel.Left) ||
		sel.Kind == combAdjacent && isSubjectSelector(sel.Right)
}

func matchesClass(tree *estree.Tree, id estree.NodeID, path []pathSegment, scopeRoot estree.NodeID) bool {
	if len(path) == 0 {
		return false
	}
	cur := id
	for i := len(path) - 1; i >= 0; i-- {
		seg := path[i]
		if seg.fieldID == estree.NoField {
			return false
		}
		if scopeRoot != estree.NoNode && cur == scopeRoot {
			return false
		}
		if tree.ParentField(cur) != seg.fieldID {
			return false
		}
		cur = tree.Parent(cur)
		if cur == estree.NoNode {
			return false
		}
	}
	return true
}

func matchesPseudo(p pseudoSelector, tree *estree.Tree, id, scopeRoot estree.NodeID, memo *evalMemo) bool {
	switch p.Name {
	case "is", "matches":
		for _, arg := range p.Args {
			if matchesSelector(arg, tree, id, scopeRoot, memo) {
				return true
			}
		}
		return false
	case "not":
		for _, arg := range p.Args {
			if matchesSelector(arg, tree, id, scopeRoot, memo) {
				return false
			}
		}
		return true
	case "has":
		if memo != nil && p.memoID != 0 {
			if hit, ok := memo.get(p.memoID, id, estree.NoNode); ok {
				return hit
			}
		}
		matched := false
		for _, arg := range p.Args {
			if matchesHas(tree, id, arg, memo) {
				matched = true
				break
			}
		}
		if memo != nil && p.memoID != 0 {
			memo.set(p.memoID, id, estree.NoNode, matched)
		}
		return matched
	case "nth-child":
		if scopeRoot != estree.NoNode && id == scopeRoot {
			return false
		}
		_, idx, ok := tree.SiblingList(id)
		return ok && idx == p.N-1
	case "nth-last-child":
		if scopeRoot != estree.NoNode && id == scopeRoot {
			return false
		}
		list, idx, ok := tree.SiblingList(id)
		return ok && list.Len()-idx == p.N
	case "statement", "declaration", "function", "expression", "pattern":
		return matchesNodeClass(tree, id, p.Name, scopeRoot)
	}
	return false
}

func matchesNodeClass(tree *estree.Tree, id estree.NodeID, class string, scopeRoot estree.NodeID) bool {
	tid := tree.Type(id)
	switch class {
	case "statement":
		return isStatementType[tid]
	case "declaration":
		return isDeclarationType[tid]
	case "function":
		return isFunctionType[tid]
	case "expression":
		return isExpressionNode(tree, id, tid, scopeRoot)
	case "pattern":
		return isPatternType[tid] || isExpressionNode(tree, id, tid, scopeRoot)
	}
	return false
}

func isExpressionNode(tree *estree.Tree, id estree.NodeID, tid estree.TypeID, scopeRoot estree.NodeID) bool {
	if isExpressionType[tid] {
		return true
	}
	if tid == typeIdentifierID {
		parent := parentInScope(tree, id, scopeRoot)
		return parent == estree.NoNode || tree.Type(parent) != typeMetaPropertyID
	}
	return false
}

func matchesHas(tree *estree.Tree, root estree.NodeID, sel selector, memo *evalMemo) bool {
	if relative, ok := sel.(relativeSelector); ok {
		if relative.Kind != combChild {
			return false
		}
		var childBuf [16]estree.NodeID
		children := tree.AppendChildren(root, childBuf[:0])
		for _, child := range children {
			if matchesSelector(relative.Inner, tree, child, root, memo) {
				return true
			}
		}
		return false
	}
	var stackBuf [32]estree.NodeID
	var childBuf [16]estree.NodeID
	stack := append(stackBuf[:0], root)
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if matchesSelector(sel, tree, cur, root, memo) {
			return true
		}
		children := tree.AppendChildren(cur, childBuf[:0])
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, children[i])
		}
	}
	return false
}

func matchesAttr(a attrSelector, tree *estree.Tree, id estree.NodeID) bool {
	val := lookupAttrPath(tree, id, a.Path)
	if a.Op == attrPresent {
		return val.Kind != estree.ValueUndefined && val.Kind != estree.ValueNull
	}
	return compareAttr(tree, val, a.Op, a.Value)
}

func lookupAttrPath(tree *estree.Tree, id estree.NodeID, path []pathSegment) estree.Value {
	cur := estree.Value{Kind: estree.ValueNode, Node: id}
	for _, seg := range path {
		cur = stepAttrPath(tree, cur, seg)
		if cur.Kind == estree.ValueNull || cur.Kind == estree.ValueUndefined {
			// esquery's getPath short-circuits on null or undefined and returns
			// that exact value without converting null to undefined.
			return cur
		}
	}
	return cur
}

func stepAttrPath(tree *estree.Tree, current estree.Value, seg pathSegment) estree.Value {
	switch current.Kind {
	case estree.ValueNode:
		switch seg.special {
		case pathType:
			return estree.Value{Kind: estree.ValueString, Str: tree.Type(current.Node).String()}
		case pathParent:
			// Program.parent is null (not undefined) in ESLint's AST.
			parent := tree.Parent(current.Node)
			if parent == estree.NoNode {
				return estree.Value{Kind: estree.ValueNull}
			}
			return estree.Value{Kind: estree.ValueNode, Node: parent}
		case pathLength:
			return estree.Value{Kind: estree.ValueUndefined}
		}
		return tree.Field(current.Node, seg.fieldID)
	case estree.ValueList:
		list := current.List()
		if seg.special == pathLength {
			return estree.Value{Kind: estree.ValueNumber, Num: float64(list.Len())}
		}
		if seg.index >= 0 && seg.index < list.Len() {
			child := list.At(seg.index)
			if child == estree.NoNode {
				return estree.Value{Kind: estree.ValueNull}
			}
			return estree.Value{Kind: estree.ValueNode, Node: child}
		}
		return estree.Value{Kind: estree.ValueUndefined}
	case estree.ValueString:
		if seg.special == pathLength {
			return estree.Value{Kind: estree.ValueNumber, Num: float64(ecmascript.StringCodeUnitCount(current.Str))}
		}
		if seg.index >= 0 {
			if ch, ok := stringCodeUnitAt(current.Str, seg.index); ok {
				return estree.Value{Kind: estree.ValueString, Str: ch}
			}
		}
		return estree.Value{Kind: estree.ValueUndefined}
	case estree.ValueRegexObject:
		pattern, flags, ok := tree.RegexInfo(current.Node)
		if !ok {
			return estree.Value{Kind: estree.ValueUndefined}
		}
		switch seg.raw {
		case "pattern":
			return estree.Value{Kind: estree.ValueString, Str: pattern}
		case "flags":
			return estree.Value{Kind: estree.ValueString, Str: flags}
		}
		return estree.Value{Kind: estree.ValueUndefined}
	case estree.ValueRegExp:
		pattern, flags, ok := tree.RegexInfo(current.Node)
		if !ok {
			return estree.Value{Kind: estree.ValueUndefined}
		}
		return readRegExpProp(pattern, flags, seg.raw)
	case estree.ValueTemplateValue:
		cooked, cookedNull, raw, ok := tree.TemplateValue(current.Node)
		if !ok {
			return estree.Value{Kind: estree.ValueUndefined}
		}
		switch seg.raw {
		case "cooked":
			if cookedNull {
				return estree.Value{Kind: estree.ValueNull}
			}
			return estree.Value{Kind: estree.ValueString, Str: cooked}
		case "raw":
			return estree.Value{Kind: estree.ValueString, Str: raw}
		}
		return estree.Value{Kind: estree.ValueUndefined}
	}
	return estree.Value{Kind: estree.ValueUndefined}
}

// stringCodeUnitAt returns the UTF-16 code unit at index as a single-unit
// JavaScript string (preserving unpaired surrogates in WTF-8 rather than
// replacing them with U+FFFD).
func stringCodeUnitAt(s string, index int) (string, bool) {
	if index < 0 {
		return "", false
	}
	unitIdx := 0
	for i := 0; i < len(s); {
		if s[i] < utf8.RuneSelf {
			if unitIdx == index {
				return s[i : i+1], true
			}
			unitIdx++
			i++
			continue
		}
		r, size := ecmascript.DecodeStringRune(s[i:])
		if r <= 0xFFFF {
			if unitIdx == index {
				return s[i : i+size], true
			}
			unitIdx++
		} else {
			r1, r2 := utf16.EncodeRune(r)
			if unitIdx == index {
				return ecmascript.StringFromCodeUnits([]uint16{uint16(r1)}), true
			}
			if unitIdx+1 == index {
				return ecmascript.StringFromCodeUnits([]uint16{uint16(r2)}), true
			}
			unitIdx += 2
		}
		i += size
	}
	return "", false
}

func readRegExpProp(pattern, flags, prop string) estree.Value {
	switch prop {
	case "source":
		return estree.Value{Kind: estree.ValueString, Str: regExpSource(pattern)}
	case "flags":
		return estree.Value{Kind: estree.ValueString, Str: sortRegExpFlags(flags)}
	case "lastIndex":
		return estree.Value{Kind: estree.ValueNumber, Num: 0}
	case "hasIndices":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 'd') >= 0}
	case "global":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 'g') >= 0}
	case "ignoreCase":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 'i') >= 0}
	case "multiline":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 'm') >= 0}
	case "dotAll":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 's') >= 0}
	case "unicode":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 'u') >= 0}
	case "unicodeSets":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 'v') >= 0}
	case "sticky":
		return estree.Value{Kind: estree.ValueBool, Bool: strings.IndexByte(flags, 'y') >= 0}
	}
	return estree.Value{Kind: estree.ValueUndefined}
}

func regExpSource(pattern string) string {
	if pattern == "" {
		return "(?:)"
	}
	return pattern
}

// sortRegExpFlags returns ECMAScript's canonical flag order: dgimsuvy.
func sortRegExpFlags(flags string) string {
	const order = "dgimsuvy"
	var buf [8]byte
	n := 0
	for i := range len(order) {
		if strings.IndexByte(flags, order[i]) >= 0 {
			buf[n] = order[i]
			n++
		}
	}
	return string(buf[:n])
}

func compareAttr(tree *estree.Tree, left estree.Value, op attrOp, right attrValue) bool {
	switch op {
	case attrEqual:
		if right.Kind == attrValueRegex {
			return left.Kind == estree.ValueString && attrStringMatchesRegex(left.Str, right)
		}
		return attrEquals(tree, left, right)
	case attrNotEqual:
		if right.Kind == attrValueRegex {
			return !attrStringMatchesRegex(valueAsString(tree, left), right)
		}
		return !attrEquals(tree, left, right)
	case attrLess, attrLessOrEqual, attrGreater, attrGreaterOrEqual:
		cmp, ok := compareRelational(tree, left, right)
		if !ok {
			return false
		}
		switch op {
		case attrLess:
			return cmp < 0
		case attrLessOrEqual:
			return cmp <= 0
		case attrGreater:
			return cmp > 0
		case attrGreaterOrEqual:
			return cmp >= 0
		}
	}
	return false
}

func compareRelational(tree *estree.Tree, left estree.Value, right attrValue) (int, bool) {
	if left.Kind == estree.ValueBigInt {
		leftInt, valid := ecmascript.StringToBigInt(left.Str)
		if !valid {
			return 0, false
		}
		if right.Kind == attrValueString || right.Kind == attrValueIdent {
			text := right.Str
			if right.Kind == attrValueIdent {
				text = right.Ident
			}
			rightInt, valid := ecmascript.StringToBigInt(text)
			if !valid {
				return 0, false
			}
			return leftInt.Cmp(rightInt), true
		}
		if right.Kind != attrValueNumber || math.IsNaN(right.Num) {
			return 0, false
		}
		return new(big.Float).SetInt(leftInt).Cmp(new(big.Float).SetFloat64(right.Num)), true
	}

	leftText, leftIsString, leftNumber, leftOK := leftRelationalPrimitive(tree, left)
	rightText, rightIsString, rightNumber, rightOK := rightRelationalPrimitive(right)
	if !leftOK || !rightOK {
		return 0, false
	}
	if leftIsString && rightIsString {
		return ecmascript.CompareStrings(leftText, rightText), true
	}
	if leftIsString {
		leftNumber, leftOK = ecmascript.StringToNumber(leftText)
	}
	if rightIsString {
		rightNumber, rightOK = ecmascript.StringToNumber(rightText)
	}
	if !leftOK || !rightOK || math.IsNaN(leftNumber) || math.IsNaN(rightNumber) {
		return 0, false
	}
	switch {
	case leftNumber < rightNumber:
		return -1, true
	case leftNumber > rightNumber:
		return 1, true
	default:
		return 0, true
	}
}

func leftRelationalPrimitive(tree *estree.Tree, v estree.Value) (text string, isString bool, number float64, ok bool) {
	switch v.Kind {
	case estree.ValueString:
		return v.Str, true, 0, true
	case estree.ValueNumber:
		return "", false, v.Num, true
	case estree.ValueBool:
		if v.Bool {
			return "", false, 1, true
		}
		return "", false, 0, true
	case estree.ValueNull:
		return "", false, 0, true
	case estree.ValueUndefined:
		return "", false, 0, false
	case estree.ValueNode, estree.ValueList, estree.ValueRegExp, estree.ValueRegexObject, estree.ValueTemplateValue:
		return valueAsString(tree, v), true, 0, true
	}
	return "", false, 0, false
}

func rightRelationalPrimitive(v attrValue) (text string, isString bool, number float64, ok bool) {
	switch v.Kind {
	case attrValueString:
		return v.Str, true, 0, true
	case attrValueIdent:
		return v.Ident, true, 0, true
	case attrValueNumber:
		return "", false, v.Num, true
	}
	return "", false, 0, false
}

func attrEquals(tree *estree.Tree, left estree.Value, right attrValue) bool {
	switch right.Kind {
	case attrValueString:
		return valueEqualsString(tree, left, right.Str)
	case attrValueNumber:
		if left.Kind == estree.ValueNumber {
			return left.Num == right.Num || (math.IsNaN(left.Num) && math.IsNaN(right.Num))
		}
		return valueEqualsString(tree, left, right.numStr)
	case attrValueIdent:
		return valueEqualsString(tree, left, right.Ident)
	case attrValueType:
		return valueTypeOf(left) == right.Ident
	}
	return false
}

func valueEqualsString(tree *estree.Tree, left estree.Value, expected string) bool {
	if left.Kind == estree.ValueList {
		n := left.List().Len()
		if n == 0 {
			return expected == ""
		}
		if len(expected) < n-1 || len(expected) > n*len("[object Object],")-1 {
			return false
		}
	}
	return valueAsString(tree, left) == expected
}

func attrStringMatchesRegex(left string, right attrValue) bool {
	if right.regexPrefix != "" && !strings.HasPrefix(left, right.regexPrefix) {
		return false
	}
	re := right.compiledRegex
	if re == nil {
		var err error
		re, err = esregexp.Compile(right.Regex, right.Flags)
		if err != nil {
			return false
		}
	}
	if !utf8.ValidString(left) {
		return re.TestRunes(decodeWTF8Runes(left))
	}
	return re.Test(left)
}

func decodeWTF8Runes(s string) []rune {
	runes := make([]rune, 0, len(s))
	for i := 0; i < len(s); {
		if s[i] < utf8.RuneSelf {
			runes = append(runes, rune(s[i]))
			i++
			continue
		}
		r, size := ecmascript.DecodeStringRune(s[i:])
		runes = append(runes, r)
		i += size
	}
	return runes
}

func valueAsString(tree *estree.Tree, v estree.Value) string {
	switch v.Kind {
	case estree.ValueString, estree.ValueBigInt:
		return v.Str
	case estree.ValueBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case estree.ValueNumber:
		return ecmascript.NumberToString(v.Num)
	case estree.ValueNull:
		return "null"
	case estree.ValueUndefined:
		return "undefined"
	case estree.ValueNode, estree.ValueRegexObject, estree.ValueTemplateValue:
		return "[object Object]"
	case estree.ValueRegExp:
		pattern, flags, ok := tree.RegexInfo(v.Node)
		if !ok {
			return "null"
		}
		return "/" + regExpSource(pattern) + "/" + sortRegExpFlags(flags)
	case estree.ValueList:
		list := v.List()
		n := list.Len()
		if n == 0 {
			return ""
		}
		var sb strings.Builder
		sb.Grow(n * len("[object Object],"))
		for i := range n {
			if i > 0 {
				sb.WriteByte(',')
			}
			if list.At(i) != estree.NoNode {
				sb.WriteString("[object Object]")
			}
		}
		return sb.String()
	}
	return ""
}

func valueTypeOf(v estree.Value) string {
	switch v.Kind {
	case estree.ValueUndefined:
		return "undefined"
	case estree.ValueBool:
		return "boolean"
	case estree.ValueNumber:
		return "number"
	case estree.ValueString:
		return "string"
	case estree.ValueBigInt:
		return "bigint"
	case estree.ValueNull, estree.ValueNode, estree.ValueList, estree.ValueRegExp, estree.ValueRegexObject, estree.ValueTemplateValue:
		return "object"
	}
	return "undefined"
}
