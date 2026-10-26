// Package estree projects the existing Program syntax into the final ESTree
// representation consumed by third-party JavaScript rules. It owns no parser,
// filesystem, configuration, checker, transport, or rule execution.
package estree

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	headerSize          = 40
	tagShift            = 29
	payloadMask         = 1<<tagShift - 1
	tagNull      uint32 = 0
	tagNode      uint32 = 1
	tagList      uint32 = 2
	tagString    uint32 = 3
	tagFalse     uint32 = 4
	tagTrue      uint32 = 5
	tagNumber    uint32 = 6
	tagUndefined uint32 = 7
)

const (
	// Frame lengths and offsets are uint32; slice lengths must also fit int.
	maxWireFrameSize = min(math.MaxUint32, math.MaxInt)
	// Bound speculative reservations, not the size of accepted frames.
	maxInitialFrameBytes = 64 << 20
)

type nodeRecord struct {
	typeID                 uint16
	start, end, valueStart uint32
}
type estreeProjector struct {
	ctx              context.Context
	file             *ast.SourceFile
	text, sourceType string
	offsets          utf16Offsets
	nodes            []nodeRecord
	values, lists    []uint32
	emptyList        uint32
	strings          []string
	stringBytes      int
	stringIDs        map[string]uint32
	numbers          []float64
	tokenCovers      []tokenCover
	triviaStarts     map[int]int
	frameBytes       uint64
	scan             scanner.Scanner
}

// Options are syntax metadata already resolved by the linter.
type Options struct {
	SourceType string
	HasBOM     bool
}

// Builder reuses scratch for sequential projections. Build returns
// independently owned bytes and clears every source-backed reference on return.
// A Builder must not be used concurrently. Reset bounds retained scratch
// before a caller returns the Builder to its own finite cache.
type Builder struct {
	p        estreeProjector
	tokens   []projectedToken
	comments []projectedComment
}

// Reset retains at most 4 MiB of source-cleared columns and a small cleared
// string index. Large dictionaries never enter the caller's finite cache.
// Returned frames never alias these buffers, so callers may reset immediately.
func (b *Builder) Reset() {
	b.p.clearSource()
	// Evict the largest columns first. One oversized column must not discard
	// every other reusable buffer and force the next file to rebuild them all.
	sizes := [...]int{cap(b.p.nodes) * 16, cap(b.p.values) * 4, cap(b.p.lists) * 4,
		cap(b.p.strings) * 16, cap(b.p.numbers) * 8, cap(b.tokens) * 12,
		cap(b.comments) * 12, cap(b.p.tokenCovers) * 12, cap(b.p.offsets.entries) * int(unsafe.Sizeof(utf16Offset{}))}
	total := 0
	for _, size := range sizes {
		total += size
	}
	for total > 4<<20 {
		largest := 0
		for index, size := range sizes {
			if size > sizes[largest] {
				largest = index
			}
		}
		total -= sizes[largest]
		sizes[largest] = 0
		switch largest {
		case 0:
			b.p.nodes = nil
		case 1:
			b.p.values = nil
		case 2:
			b.p.lists = nil
		case 3:
			b.p.strings = nil
		case 4:
			b.p.numbers = nil
		case 5:
			b.tokens = nil
		case 6:
			b.comments = nil
		case 7:
			b.p.tokenCovers = nil
		case 8:
			b.p.offsets.entries = nil
		}
	}
}

func (b *Builder) Build(ctx context.Context, file *ast.SourceFile, options Options) (data []byte, err error) {
	return b.BuildInto(ctx, file, options, nil)
}

// BuildInto asks allocate for exactly the final frame size after projection.
// A nil allocator gives the frame independent ownership. Caller-provided
// storage must remain immutable until every consumer has finished; it never
// aliases the Builder's scratch. allocate is called synchronously at most once.
func (b *Builder) BuildInto(ctx context.Context, file *ast.SourceFile, options Options, allocate func(int) []byte) (data []byte, err error) {
	if options.SourceType != "module" && options.SourceType != "script" && options.SourceType != "commonjs" {
		return nil, fmt.Errorf("invalid ESTree sourceType %q", options.SourceType)
	}
	defer func() {
		b.p.clearSource()
		if failure := recover(); failure != nil {
			data = nil
			if e, ok := failure.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("ESTree projection: %v", failure)
			}
		}
	}()
	p := &b.p
	p.ctx = ctx
	p.check()
	p.file = file
	p.text = file.Text()
	p.sourceType = options.SourceType
	if options.SourceType == "commonjs" {
		p.sourceType = "script"
	}
	if len(p.text) > maxWireFrameSize-headerSize-6 {
		return nil, errors.New("ESTree source does not fit frame encoding")
	}
	// Reserve the maximum padding between columns before growing scratch.
	p.frameBytes = headerSize + uint64(len(p.text)) + 6
	p.offsets.reset(ctx, p.text)
	// The native node count is a sizing hint, not a limit on projected syntax.
	// Reserve once per file instead of copying growing columns at every append.
	p.nodes = slices.Grow(p.nodes[:0], min(file.NodeCount, maxInitialFrameBytes/14))
	p.values = slices.Grow(p.values[:0], min(file.NodeCount*4, maxInitialFrameBytes/4))
	p.lists = p.lists[:0]
	p.emptyList = 0
	p.strings = p.strings[:0]
	p.stringBytes = 0
	p.numbers = p.numbers[:0]
	if p.stringIDs == nil {
		p.stringIDs = make(map[string]uint32)
	}
	p.tokenCovers = p.tokenCovers[:0]
	root := p.convert(file.AsNode(), false)
	initialTokenBytes := maxInitialFrameBytes - int(min(p.frameBytes, maxInitialFrameBytes))
	b.tokens = slices.Grow(b.tokens[:0], min(max(16, file.NodeCount*2), initialTokenBytes/9))
	b.tokens, b.comments = scanTokensAndCommentsInto(ctx, &p.scan, file, p.offsets, p.tokenCovers, b.tokens[:0], b.comments[:0], maxWireFrameSize-int(p.frameBytes))
	return b.encode(root, options.HasBOM, allocate)
}

func (p *estreeProjector) check() {
	if err := p.ctx.Err(); err != nil {
		panic(err)
	}
}

func (p *estreeProjector) clearSource() {
	p.file = nil
	p.text = ""
	p.sourceType = ""
	p.ctx = nil
	p.offsets.entries = p.offsets.entries[:0]
	indexSize := len(p.stringIDs)
	clear(p.stringIDs)
	if indexSize > 4096 {
		p.stringIDs = nil
	}
	clear(p.strings)
	p.strings = p.strings[:0]
	p.tokenCovers = p.tokenCovers[:0]
	p.triviaStarts = nil
	// Numeric caches and token state may contain source-backed strings.
	// Reuse the scanner's storage, never state from a previous file.
	p.scan = scanner.Scanner{}
}

func (p *estreeProjector) reserve(bytes uint64) {
	if bytes > uint64(maxWireFrameSize)-p.frameBytes {
		panic(errors.New("ESTree frame size does not fit its encoding"))
	}
	p.frameBytes += bytes
}
func taggedValue(tag, payload uint32) uint32 {
	if payload > payloadMask {
		panic("ESTree wire payload exceeds 29 bits")
	}
	return tag<<tagShift | payload
}

func nodeValue(id uint32) uint32 {
	if id == 0 {
		return taggedValue(tagNull, 0)
	}
	return taggedValue(tagNode, id)
}

func boolValue(value bool) uint32 {
	if value {
		return taggedValue(tagTrue, 0)
	}
	return taggedValue(tagFalse, 0)
}

func nullValue() uint32      { return taggedValue(tagNull, 0) }
func undefinedValue() uint32 { return taggedValue(tagUndefined, 0) }

func (p *estreeProjector) stringValue(value string) uint32 {
	if id, ok := p.stringIDs[value]; ok {
		return taggedValue(tagString, id)
	}
	id := uint32(len(p.strings))
	p.stringIDs[value] = id
	encodedSize := uint64(8)
	if _, ok := p.sourceStringOffset(value); !ok {
		encodedSize = (4 + uint64(ecmascript.StringCodeUnitCount(value))*2 + 3) &^ 3
	}
	p.reserve(encodedSize)
	p.stringBytes += int(encodedSize)
	p.strings = append(p.strings, value)
	return taggedValue(tagString, id)
}

// sourceStringOffset recognizes immutable slices of this file's source. The
// addresses are only a sizing hint: bounds, equality and UTF-8 validity guard
// the representation. No pointer or source-backed view enters the frame.
// Source-span lengths reserve their high bit to distinguish UTF-16 records.
func (p *estreeProjector) sourceStringOffset(value string) (int, bool) {
	if len(value) < 3 || len(value) > math.MaxInt32 || len(value) > len(p.text) {
		return 0, false
	}
	base := uintptr(unsafe.Pointer(unsafe.StringData(p.text)))
	address := uintptr(unsafe.Pointer(unsafe.StringData(value)))
	if address < base || address-base > uintptr(len(p.text)-len(value)) {
		return 0, false
	}
	offset := int(address - base)
	if p.text[offset:offset+len(value)] != value || !utf8.ValidString(value) {
		return 0, false
	}
	return offset, true
}

// Encode straight into the final frame instead of retaining a second UTF-16
// byte column. clearSource releases all staged string references on return.
func (p *estreeProjector) encodeStrings(out []byte) {
	for _, value := range p.strings {
		if offset, ok := p.sourceStringOffset(value); ok {
			p.check()
			binary.LittleEndian.PutUint32(out, ^uint32(len(value)))
			binary.LittleEndian.PutUint32(out[4:], uint32(offset))
			out = out[8:]
			continue
		}
		cursor := 4
		for index := 0; index < len(value); {
			p.check()
			end := min(len(value), index+4096)
			for index < end {
				r, size := ecmascript.DecodeStringRune(value[index:])
				if r > 0xFFFF {
					high, low := utf16.EncodeRune(r)
					binary.LittleEndian.PutUint16(out[cursor:], uint16(high))
					binary.LittleEndian.PutUint16(out[cursor+2:], uint16(low))
					cursor += 4
				} else {
					binary.LittleEndian.PutUint16(out[cursor:], uint16(r))
					cursor += 2
				}
				index += size
			}
		}
		binary.LittleEndian.PutUint32(out, uint32((cursor-4)/2))
		end := align4(cursor)
		clear(out[cursor:end])
		out = out[end:]
	}
}

func (p *estreeProjector) numberValue(value float64) uint32 {
	p.reserve(8)
	index := uint32(len(p.numbers))
	p.numbers = append(p.numbers, value)
	return taggedValue(tagNumber, index)
}

func (p *estreeProjector) listValue(ids []uint32) uint32 {
	if len(ids) == 0 && p.emptyList != 0 {
		return p.emptyList
	}
	p.reserve((uint64(len(ids)) + 1) * 4)
	offset := uint32(len(p.lists))
	p.lists = append(p.lists, uint32(len(ids)))
	p.lists = append(p.lists, ids...)
	value := taggedValue(tagList, offset)
	if len(ids) == 0 {
		p.emptyList = value
	}
	return value
}

func (p *estreeProjector) convertList(nodes []*ast.Node, pattern bool) uint32 {
	ids := make([]uint32, 0, len(nodes))
	for _, node := range nodes {
		if id := p.convert(node, pattern); id != 0 {
			ids = append(ids, id)
		}
	}
	return p.listValue(ids)
}

func (p *estreeProjector) convertListWithNulls(nodes []*ast.Node, pattern bool) uint32 {
	ids := make([]uint32, len(nodes))
	for index, node := range nodes {
		ids[index] = p.convert(node, pattern)
	}
	return p.listValue(ids)
}

func (p *estreeProjector) add(typeName string, start, end uint32, values ...uint32) uint32 {
	typeID := typeIDs[typeName]
	for index, field := range schemas[typeID].fields {
		switch field {
		case "abstract", "accessibility", "declare", "definite", "exportKind", "importKind", "implements", "override", "readonly", "returnType", "superTypeArguments", "typeAnnotation", "typeArguments", "typeParameters":
			if p.isJavaScriptFile() {
				values[index] = undefinedValue()
			} else if values[index] == undefinedValue() {
				values[index] = nullValue()
			}
		case "optional":
			if p.isJavaScriptFile() && typeName != "MemberExpression" && typeName != "CallExpression" {
				values[index] = undefinedValue()
			}
		case "decorators":
			if p.isJavaScriptFile() && (typeName == "Identifier" || typeName == "ArrayPattern" || typeName == "ObjectPattern" || typeName == "RestElement" || typeName == "AssignmentPattern") {
				values[index] = undefinedValue()
			}
		}
	}
	p.reserve(14 + uint64(len(values))*4)
	id := uint32(len(p.nodes) + 1)
	p.nodes = append(p.nodes, nodeRecord{
		typeID:     typeID,
		start:      start,
		end:        end,
		valueStart: uint32(len(p.values)),
	})
	p.values = append(p.values, values...)
	return id
}

func (p *estreeProjector) tokenStart(node *ast.Node) int {
	// Missing and JSDoc/JSX nodes use different trivia rules. Keep their
	// scanner behavior rather than sharing an ordinary-source cache entry.
	if ast.NodeIsMissing(node) || ast.IsJSDocNode(node) ||
		node.Kind == ast.KindJsxText || node.Flags&ast.NodeFlagsJSDoc != 0 {
		return scanner.GetTokenPosOfNode(node, p.file, false)
	}
	pos := node.Pos()
	if start, ok := p.triviaStarts[pos]; ok {
		return start
	}
	start := scanner.GetTokenPosOfNode(node, p.file, false)
	// Only index long runs: ordinary short spacing needs no extra storage.
	// Every parent in a deep expression can share this same full-start pos.
	if start-pos >= 64 {
		if p.triviaStarts == nil {
			p.triviaStarts = make(map[int]int)
		}
		p.triviaStarts[pos] = start
	}
	return start
}

func (p *estreeProjector) rangeOf(node *ast.Node) (uint32, uint32) {
	if node == nil {
		return 0, 0
	}
	start := node.Pos()
	if node.Kind == ast.KindSourceFile {
		start = 0
	} else if node.Kind != ast.KindJsxText && node.Kind != ast.KindJsxTextAllWhiteSpaces {
		start = p.tokenStart(node)
	}
	return p.offsets.At(start), p.offsets.At(node.End())
}

func (p *estreeProjector) addNode(typeName string, node *ast.Node, values ...uint32) uint32 {
	start, end := p.rangeOf(node)
	return p.add(typeName, start, end, values...)
}

func (p *estreeProjector) hasModifier(node *ast.Node, flag ast.ModifierFlags) bool {
	return node != nil && utils.ESTreeModifierFlags(node)&flag != 0
}

func (p *estreeProjector) decorators(node *ast.Node) uint32 {
	if p.isJavaScriptFile() && (node == nil || len(node.Decorators()) == 0) {
		return undefinedValue()
	}
	if node == nil {
		return p.listValue(nil)
	}
	return p.convertList(node.Decorators(), false)
}

func (p *estreeProjector) accessibility(node *ast.Node) uint32 {
	flags := utils.ESTreeModifierFlags(node)
	switch {
	case flags&ast.ModifierFlagsPrivate != 0:
		return p.stringValue("private")
	case flags&ast.ModifierFlagsProtected != 0:
		return p.stringValue("protected")
	case flags&ast.ModifierFlagsPublic != 0:
		return p.stringValue("public")
	default:
		return undefinedValue()
	}
}

func (p *estreeProjector) typeAnnotation(node *ast.Node) uint32 {
	if node == nil || isJSDocCommentSyntax(node) || p.isJavaScriptFile() {
		return undefinedValue()
	}
	typeID := p.convert(node, false)
	_, end := p.rangeOf(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindTypePredicate {
		inner := p.nodes[typeID-1]
		return nodeValue(p.add("TSTypeAnnotation", inner.start, inner.end, nodeValue(typeID)))
	}
	// The parser's full start follows the authored ':' or '=>' token and
	// includes intervening trivia, so comments cannot confuse a backward scan.
	width := 1
	if parent := node.Parent; parent != nil && (parent.Kind == ast.KindFunctionType || parent.Kind == ast.KindConstructorType) {
		width = 2
	}
	start := p.offsets.At(max(0, node.Pos()-width))
	return nodeValue(p.add("TSTypeAnnotation", start, end, nodeValue(typeID)))
}

func (p *estreeProjector) typeParameters(list *ast.NodeList, declaration bool) uint32 {
	if list == nil || p.isJavaScriptFile() || (len(list.Nodes) > 0 && isJSDocCommentSyntax(list.Nodes[0])) {
		return undefinedValue()
	}
	typeName := "TSTypeParameterInstantiation"
	if declaration {
		typeName = "TSTypeParameterDeclaration"
	}
	start := p.offsets.At(max(0, list.Pos()-1))
	endPos := list.End()
	scan := &p.scan
	scan.SetText(p.text)
	scan.SetSkipTrivia(true)
	scan.ResetTokenState(endPos)
	if scan.Scan() != ast.KindGreaterThanToken {
		panic(errors.New("missing type parameter closing token"))
	}
	endPos = scan.TokenEnd()
	end := p.offsets.At(endPos)
	return nodeValue(p.add(typeName, start, end, p.convertList(list.Nodes, false)))
}

func (p *estreeProjector) identifierWithAnnotation(name, typeNode *ast.Node) uint32 {
	id := p.convert(name, true)
	if typeNode == nil || id == 0 || p.isJavaScriptFile() || isJSDocCommentSyntax(typeNode) {
		return id
	}
	record := p.nodes[id-1]
	fields := schemas[record.typeID].fields
	annotation := p.typeAnnotation(typeNode)
	for index, field := range fields {
		if field == "typeAnnotation" {
			p.values[int(record.valueStart)+index] = annotation
			p.nodes[id-1].end = p.offsets.At(typeNode.End())
			break
		}
	}
	return id
}

// JSDoc kinds can also represent source-authored TypeScript recovery syntax.
// Only the flags identify syntax synthesized from comments.
func isJSDocCommentSyntax(node *ast.Node) bool {
	return node != nil && node.Flags&(ast.NodeFlagsJSDoc|ast.NodeFlagsReparsed) != 0
}

func (p *estreeProjector) isJavaScriptFile() bool {
	return p.file.ScriptKind == core.ScriptKindJS || p.file.ScriptKind == core.ScriptKindJSX
}

func parseNumberLiteral(text string) float64 {
	value, ok := ecmascript.StringToNumber(strings.ReplaceAll(text, "_", ""))
	if !ok {
		panic(fmt.Errorf("invalid numeric literal %q", text))
	}
	return value
}

// The plugin projection owns this reusable sparse index. Do not populate the
// SourceFile's native PositionMap cache: it outlives the projected frame.
// Each sparse 64-byte block records rune ends whose byte/UTF-16 difference
// increases by one (two-byte UTF-8) or two (three/four-byte UTF-8). Prefix
// deltas and population counts preserve queries inside a rune without keeping
// an entry per character in large Unicode strings. No source is retained.
type utf16Offset struct {
	block, delta uint32
	one, two     uint64
}
type utf16Offsets struct {
	entries []utf16Offset
	delta   uint32
}

func (o *utf16Offsets) reset(ctx context.Context, text string) {
	o.entries = o.entries[:0]
	o.delta = 0
	delta, checkAt := 0, 0
	for index := 0; index < len(text); {
		if index >= checkAt {
			if err := ctx.Err(); err != nil {
				panic(err)
			}
			checkAt = index + 4096
		}
		if text[index] < utf8.RuneSelf {
			index++
			continue
		}
		r, size := ecmascript.DecodeStringRune(text[index:])
		units := 1
		if r >= 0x10000 {
			units = 2
		}
		index += size
		step := size - units
		if step == 0 {
			continue
		}
		block := uint32(index) &^ 63
		if len(o.entries) == 0 || o.entries[len(o.entries)-1].block != block {
			o.entries = append(o.entries, utf16Offset{block: block, delta: uint32(delta)})
		}
		entry := &o.entries[len(o.entries)-1]
		if step == 1 {
			entry.one |= uint64(1) << (uint32(index) & 63)
		} else {
			entry.two |= uint64(1) << (uint32(index) & 63)
		}
		delta += step
	}
	o.delta = uint32(delta)
}

func (o utf16Offsets) At(offset int) uint32 {
	if len(o.entries) == 0 {
		return uint32(offset)
	}
	return o.nonASCIIAt(offset)
}

func (o utf16Offsets) nonASCIIAt(offset int) uint32 {
	lo, hi := 0, len(o.entries)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if int(o.entries[mid].block) <= offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo == 0 {
		return uint32(offset)
	}
	entry := o.entries[lo-1]
	within := offset - int(entry.block)
	if within >= 63 {
		// Between populated blocks, the next prefix is the complete delta for
		// this block. The final block uses the total, avoiding bit counts in
		// long ASCII gaps without adding a field to every entry.
		if lo < len(o.entries) {
			return uint32(offset) - o.entries[lo].delta
		}
		return uint32(offset) - o.delta
	}
	mask := ^uint64(0) >> (63 - within)
	delta := entry.delta + uint32(bits.OnesCount64(entry.one&mask)+2*bits.OnesCount64(entry.two&mask))
	return uint32(offset) - delta
}
func align4(n int) int { return (n + 3) &^ 3 }

func (b *Builder) encode(root uint32, hasBOM bool, allocate func(int) []byte) ([]byte, error) {
	p := &b.p
	n := len(p.nodes)
	types := headerSize
	starts := align4(types + n*2)
	ends := starts + n*4
	valueStarts := ends + n*4
	values := valueStarts + n*4
	lists := values + len(p.values)*4
	stringsOffset := lists + len(p.lists)*4
	numbers := stringsOffset + p.stringBytes
	tokenTypes := numbers + len(p.numbers)*8
	tokenStarts := align4(tokenTypes + len(b.tokens))
	tokenEnds := tokenStarts + len(b.tokens)*4
	comments := tokenEnds + len(b.tokens)*4
	source := comments + len(b.comments)*12
	size := source + len(p.text)
	if size > maxWireFrameSize {
		return nil, errors.New("ESTree frame size does not fit its encoding")
	}
	var out []byte
	if allocate == nil {
		out = make([]byte, size)
	} else {
		out = allocate(size)
	}
	// A reused buffer may contain a previous frame's flags and alignment bytes.
	clear(out[:headerSize])
	clear(out[types+n*2 : starts])
	clear(out[tokenTypes+len(b.tokens) : tokenStarts])
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(out[offset:], value) }
	// One private format shared by the matching Go and Node build. Counts
	// locate the columns; the remaining bytes are the source text.
	if hasBOM {
		put(0, 1)
	}
	put(4, root)
	put(8, uint32(n))
	put(12, uint32(len(p.values)))
	put(16, uint32(len(p.lists)))
	put(20, uint32(len(p.strings)))
	put(24, uint32(p.stringBytes))
	put(28, uint32(len(p.numbers)))
	put(32, uint32(len(b.tokens)))
	put(36, uint32(len(b.comments)))
	for index, node := range p.nodes {
		if index&1023 == 0 {
			p.check()
		}
		binary.LittleEndian.PutUint16(out[types+index*2:], node.typeID)
		put(starts+index*4, node.start)
		put(ends+index*4, node.end)
		put(valueStarts+index*4, node.valueStart)
	}
	for index, value := range p.values {
		if index&4095 == 0 {
			p.check()
		}
		put(values+index*4, value)
	}
	for index, value := range p.lists {
		put(lists+index*4, value)
	}
	p.encodeStrings(out[stringsOffset:numbers])
	for index, value := range p.numbers {
		binary.LittleEndian.PutUint64(out[numbers+index*8:], math.Float64bits(value))
	}
	for index, token := range b.tokens {
		if index&1023 == 0 {
			p.check()
		}
		out[tokenTypes+index] = token.typeCode
		put(tokenStarts+index*4, token.start)
		put(tokenEnds+index*4, token.end)
	}
	for index, comment := range b.comments {
		base := comments + index*12
		put(base, comment.kind)
		put(base+4, comment.start)
		put(base+8, comment.end)
	}
	copy(out[source:], p.text)
	p.check()
	return out[:size:size], nil
}
