package no_restricted_syntax

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// parseSelector parses an ESLint selector string into a selector tree.
// It supports the subset of esquery used by `no-restricted-syntax` test
// fixtures plus realistic real-world usage: identifiers, wildcard `*`,
// class selectors (`.field`), attribute selectors (`[a]`, `[a=v]`,
// `[a>v]`, …), pseudo-classes (`:is`, `:not`, `:matches`, `:has`,
// `:first-child`, `:last-child`, `:nth-child(N)`, `:nth-last-child(N)`),
// combinators `>`, `+`, `~` and descendant whitespace, and unions `,`.
//
// Returns the parsed selector and a non-nil error on malformed input.
func parseSelector(input string) (selector, error) {
	p := &parser{src: input, pos: 0}
	p.skipSpaces()
	sel, err := p.parseUnion()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	if p.pos != len(p.src) {
		return nil, fmt.Errorf("unexpected character %q at position %d", p.src[p.pos], p.pos)
	}
	return sel, nil
}

type parser struct {
	src string
	pos int
}

func (p *parser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *parser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) skipSpaces() {
	for !p.eof() && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t' || p.src[p.pos] == '\n' || p.src[p.pos] == '\r') {
		p.pos++
	}
}

// hasSpaceBefore reports whether at least one whitespace char preceded
// the current position. Used to distinguish descendant combinator
// (whitespace) from an end-of-selector situation.
func (p *parser) hasSpaceBefore(prevPos int) bool {
	if prevPos >= p.pos {
		return false
	}
	for i := prevPos; i < p.pos; i++ {
		c := p.src[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return true
		}
	}
	return false
}

func (p *parser) parseUnion() (selector, error) {
	first, err := p.parseCompound()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	if p.eof() || p.peek() != ',' {
		return first, nil
	}
	parts := []selector{first}
	for !p.eof() && p.peek() == ',' {
		p.pos++ // ,
		p.skipSpaces()
		next, err := p.parseCompound()
		if err != nil {
			return nil, err
		}
		parts = append(parts, next)
		p.skipSpaces()
	}
	return unionSelector{Selectors: parts}, nil
}

// parseCompound parses a compound selector, i.e. a chain of sequences
// joined by combinators (` `, `>`, `+`, `~`).
func (p *parser) parseCompound() (selector, error) {
	left, err := p.parseSequence()
	if err != nil {
		return nil, err
	}
	for {
		startPos := p.pos
		p.skipSpaces()
		if p.eof() || p.peek() == ',' || p.peek() == ')' {
			return left, nil
		}
		var kind combinatorKind
		switch p.peek() {
		case '>':
			kind = combChild
			p.pos++
			p.skipSpaces()
		case '+':
			kind = combAdjacent
			p.pos++
			p.skipSpaces()
		case '~':
			kind = combSibling
			p.pos++
			p.skipSpaces()
		default:
			if !p.hasSpaceBefore(startPos) {
				return left, nil
			}
			kind = combDescendant
		}
		right, err := p.parseSequence()
		if err != nil {
			return nil, err
		}
		left = combinatorSelector{Kind: kind, Left: left, Right: right}
	}
}

// parseSequence parses a single selector with optional filters chained on
// the same node (e.g. `Identifier[name="foo"]:not(...)`).
func (p *parser) parseSequence() (selector, error) {
	var head selector
	c := p.peek()
	switch {
	case c == 0:
		return nil, errors.New("unexpected end of selector")
	case c == '*':
		p.pos++
		head = identifierSelector{Name: "*"}
	case c == '.' || c == '[' || c == ':':
		// Filters without an explicit head — equivalent to `*` followed by
		// the filter.
		head = identifierSelector{Name: "*"}
	case isIdentStart(c):
		name, err := p.parseIdent()
		if err != nil {
			return nil, err
		}
		head = identifierSelector{Name: name}
	default:
		return nil, fmt.Errorf("unexpected character %q at position %d", c, p.pos)
	}

	for !p.eof() {
		c := p.peek()
		switch c {
		case '.':
			p.pos++
			cls, err := p.parseIdent()
			if err != nil {
				return nil, err
			}
			head = classSelector{Inner: head, Class: cls}
		case '[':
			attr, err := p.parseAttr(head)
			if err != nil {
				return nil, err
			}
			head = attr
		case ':':
			ps, err := p.parsePseudo(head)
			if err != nil {
				return nil, err
			}
			head = ps
		default:
			return head, nil
		}
	}
	return head, nil
}

func (p *parser) parseIdent() (string, error) {
	start := p.pos
	if p.eof() || !isIdentStart(p.src[p.pos]) {
		return "", fmt.Errorf("expected identifier at position %d", p.pos)
	}
	p.pos++
	for !p.eof() && isIdentCont(p.src[p.pos]) {
		p.pos++
	}
	return p.src[start:p.pos], nil
}

// parseAttr handles `[path]`, `[path op value]`. The leading `[` has been
// peeked but not consumed.
func (p *parser) parseAttr(inner selector) (selector, error) {
	if p.peek() != '[' {
		return nil, fmt.Errorf("expected [ at position %d", p.pos)
	}
	p.pos++ // [
	p.skipSpaces()
	pathParts, err := p.parseAttrPath()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	if !p.eof() && p.peek() == ']' {
		p.pos++
		return attrSelector{Inner: inner, Path: pathParts, Op: attrPresent}, nil
	}
	op, err := p.parseAttrOp()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	val, err := p.parseAttrValue()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	if p.eof() || p.peek() != ']' {
		return nil, fmt.Errorf("expected ] at position %d", p.pos)
	}
	p.pos++
	return attrSelector{Inner: inner, Path: pathParts, Op: op, Value: val}, nil
}

func (p *parser) parseAttrPath() ([]string, error) {
	var parts []string
	for {
		ident, err := p.parseAttrPathSegment()
		if err != nil {
			return nil, err
		}
		parts = append(parts, ident)
		if p.eof() || p.peek() != '.' {
			break
		}
		p.pos++
	}
	return parts, nil
}

// parseAttrPathSegment accepts a regular identifier, optionally followed by
// hyphens (paths like `foo-bar` are not used by ESTree but esquery accepts
// them).
func (p *parser) parseAttrPathSegment() (string, error) {
	start := p.pos
	if p.eof() || !isIdentStart(p.peek()) {
		return "", fmt.Errorf("expected identifier in attribute path at position %d", p.pos)
	}
	p.pos++
	for !p.eof() {
		c := p.peek()
		if isIdentCont(c) {
			p.pos++
			continue
		}
		break
	}
	return p.src[start:p.pos], nil
}

func (p *parser) parseAttrOp() (attrOp, error) {
	if p.eof() {
		return 0, fmt.Errorf("expected operator at position %d", p.pos)
	}
	switch p.peek() {
	case '=':
		p.pos++
		return attrEqual, nil
	case '!':
		if p.pos+1 < len(p.src) && p.src[p.pos+1] == '=' {
			p.pos += 2
			return attrNotEqual, nil
		}
	case '<':
		if p.pos+1 < len(p.src) && p.src[p.pos+1] == '=' {
			p.pos += 2
			return attrLessOrEqual, nil
		}
		p.pos++
		return attrLess, nil
	case '>':
		if p.pos+1 < len(p.src) && p.src[p.pos+1] == '=' {
			p.pos += 2
			return attrGreaterOrEqual, nil
		}
		p.pos++
		return attrGreater, nil
	}
	return 0, fmt.Errorf("unsupported operator at position %d", p.pos)
}

func (p *parser) parseAttrValue() (attrValue, error) {
	if p.eof() {
		return attrValue{}, fmt.Errorf("expected value at position %d", p.pos)
	}
	c := p.peek()
	switch {
	case c == '"' || c == '\'':
		s, err := p.parseString(c)
		if err != nil {
			return attrValue{}, err
		}
		return attrValue{Kind: attrValueString, Str: s}, nil
	case c == '/':
		pat, flags, err := p.parseRegex()
		if err != nil {
			return attrValue{}, err
		}
		return attrValue{Kind: attrValueRegex, Regex: pat, Flags: flags}, nil
	case c == '-' || (c >= '0' && c <= '9'):
		num, err := p.parseNumber()
		if err != nil {
			return attrValue{}, err
		}
		return attrValue{Kind: attrValueNumber, Num: num}, nil
	case isIdentStart(c):
		ident, err := p.parseIdent()
		if err != nil {
			return attrValue{}, err
		}
		switch ident {
		case "true":
			return attrValue{Kind: attrValueBool, Bool: true}, nil
		case "false":
			return attrValue{Kind: attrValueBool, Bool: false}, nil
		case "null":
			return attrValue{Kind: attrValueNull}, nil
		}
		return attrValue{Kind: attrValueIdent, Ident: ident}, nil
	}
	return attrValue{}, fmt.Errorf("unexpected attribute value at position %d", p.pos)
}

func (p *parser) parseString(quote byte) (string, error) {
	if p.peek() != quote {
		return "", fmt.Errorf("expected %c at position %d", quote, p.pos)
	}
	p.pos++ // opening quote
	var sb strings.Builder
	for !p.eof() {
		c := p.src[p.pos]
		if c == '\\' && p.pos+1 < len(p.src) {
			next := p.src[p.pos+1]
			sb.WriteByte(next)
			p.pos += 2
			continue
		}
		if c == quote {
			p.pos++
			return sb.String(), nil
		}
		sb.WriteByte(c)
		p.pos++
	}
	return "", fmt.Errorf("unterminated string starting at quote %c", quote)
}

func (p *parser) parseRegex() (string, string, error) {
	if p.peek() != '/' {
		return "", "", fmt.Errorf("expected / at position %d", p.pos)
	}
	p.pos++ // opening /
	var pat strings.Builder
	for !p.eof() {
		c := p.src[p.pos]
		if c == '\\' && p.pos+1 < len(p.src) {
			pat.WriteByte(c)
			pat.WriteByte(p.src[p.pos+1])
			p.pos += 2
			continue
		}
		if c == '/' {
			p.pos++ // closing /
			flagsStart := p.pos
			for !p.eof() {
				fc := p.src[p.pos]
				if fc >= 'a' && fc <= 'z' || fc >= 'A' && fc <= 'Z' {
					p.pos++
					continue
				}
				break
			}
			return pat.String(), p.src[flagsStart:p.pos], nil
		}
		pat.WriteByte(c)
		p.pos++
	}
	return "", "", errors.New("unterminated regex literal")
}

func (p *parser) parseNumber() (float64, error) {
	start := p.pos
	if p.peek() == '-' {
		p.pos++
	}
	for !p.eof() && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
		p.pos++
	}
	if !p.eof() && p.src[p.pos] == '.' {
		p.pos++
		for !p.eof() && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
			p.pos++
		}
	}
	val, err := strconv.ParseFloat(p.src[start:p.pos], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number at position %d: %w", start, err)
	}
	return val, nil
}

func (p *parser) parsePseudo(inner selector) (selector, error) {
	if p.peek() != ':' {
		return nil, fmt.Errorf("expected : at position %d", p.pos)
	}
	p.pos++ // :
	name, err := p.parsePseudoName()
	if err != nil {
		return nil, err
	}
	switch name {
	case "first-child":
		return wrapPseudo(inner, pseudoSelector{Name: "nth-child", N: 1}), nil
	case "last-child":
		return wrapPseudo(inner, pseudoSelector{Name: "nth-last-child", N: 1}), nil
	case "nth-child", "nth-last-child":
		n, err := p.parsePseudoNumberArg()
		if err != nil {
			return nil, err
		}
		return wrapPseudo(inner, pseudoSelector{Name: name, N: n}), nil
	case "is", "matches", "not", "has":
		args, err := p.parsePseudoSelectorArgs()
		if err != nil {
			return nil, err
		}
		return wrapPseudo(inner, pseudoSelector{Name: name, Args: args}), nil
	case "statement", "expression", "declaration", "function", "pattern":
		// Treat known esquery class pseudos as a no-op match (we don't
		// model these category sets). The combination with a head
		// selector still narrows by the head, so this is conservative.
		return inner, nil
	}
	return nil, fmt.Errorf("unsupported pseudo class %q at position %d", name, p.pos)
}

func wrapPseudo(inner selector, p pseudoSelector) selector {
	if id, ok := inner.(identifierSelector); ok && id.Name == "*" {
		// `*:not(X)` is just `:not(X)` — collapse to keep selector trees tidy
		// but still report the constraint.
		return p
	}
	return combinedPseudo{Inner: inner, Pseudo: p}
}

// combinedPseudo is `inner` AND `pseudo` applied to the same node.
type combinedPseudo struct {
	Inner  selector
	Pseudo pseudoSelector
}

func (combinedPseudo) isSelector() {}

func (p *parser) parsePseudoName() (string, error) {
	start := p.pos
	for !p.eof() {
		c := p.src[p.pos]
		if isIdentCont(c) {
			p.pos++
			continue
		}
		break
	}
	if p.pos == start {
		return "", fmt.Errorf("expected pseudo name at position %d", p.pos)
	}
	return p.src[start:p.pos], nil
}

func (p *parser) parsePseudoNumberArg() (int, error) {
	if p.eof() || p.peek() != '(' {
		return 0, fmt.Errorf("expected ( at position %d", p.pos)
	}
	p.pos++ // (
	p.skipSpaces()
	start := p.pos
	for !p.eof() && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
		p.pos++
	}
	if start == p.pos {
		return 0, fmt.Errorf("expected number at position %d", p.pos)
	}
	n, err := strconv.Atoi(p.src[start:p.pos])
	if err != nil {
		return 0, err
	}
	p.skipSpaces()
	if p.eof() || p.peek() != ')' {
		return 0, fmt.Errorf("expected ) at position %d", p.pos)
	}
	p.pos++
	return n, nil
}

func (p *parser) parsePseudoSelectorArgs() ([]selector, error) {
	if p.eof() || p.peek() != '(' {
		return nil, fmt.Errorf("expected ( at position %d", p.pos)
	}
	p.pos++ // (
	p.skipSpaces()
	var args []selector
	first, err := p.parseCompound()
	if err != nil {
		return nil, err
	}
	args = append(args, first)
	p.skipSpaces()
	for !p.eof() && p.peek() == ',' {
		p.pos++
		p.skipSpaces()
		next, err := p.parseCompound()
		if err != nil {
			return nil, err
		}
		args = append(args, next)
		p.skipSpaces()
	}
	if p.eof() || p.peek() != ')' {
		return nil, fmt.Errorf("expected ) at position %d", p.pos)
	}
	p.pos++
	return args, nil
}

func isIdentStart(c byte) bool {
	return c == '_' || c == '#' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '$'
}

func isIdentCont(c byte) bool {
	if isIdentStart(c) {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	if c == '-' {
		return true
	}
	return false
}
