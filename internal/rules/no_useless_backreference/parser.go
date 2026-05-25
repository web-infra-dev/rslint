// Package no_useless_backreference implements ESLint's
// no-useless-backreference rule. The rule walks each regex pattern looking
// for backreferences that always match the empty string regardless of input.
//
// This file contains a minimal ECMAScript regex parser that produces just
// enough of an AST (Pattern → Alternatives → Groups / Lookarounds /
// Backreferences) to run the analysis. It is NOT a complete regex parser —
// it skips over content it doesn't need (characters, character classes,
// quantifiers) but always advances correctly past them.
//
// On any syntax error the parser returns ok=false and the caller skips the
// regex entirely (matching ESLint's behavior of catching parser exceptions).
package no_useless_backreference

import (
	"strconv"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// rxNodeKind is a regex AST node kind.
type rxNodeKind int

const (
	nkPattern rxNodeKind = iota
	nkAlternative
	nkCapturingGroup
	nkGroup // non-capturing (?:...)
	nkLookaround
	nkBackref
)

// rxNode is a node in the lightweight regex AST. Only the fields relevant to
// the rule are populated.
type rxNode struct {
	kind     rxNodeKind
	start    int // byte offset of the construct's first char in the pattern
	end      int // byte offset one past the last char
	raw      string
	parent   *rxNode
	children []*rxNode

	// for capturing groups
	name   string // empty for unnamed groups
	number int    // 1-based; 0 for non-capturing groups

	// for lookarounds
	isAhead bool
	negate  bool

	// for backreferences
	refName  string   // for \k<name>
	refNum   int      // for \N (1-based, 0 if name-based)
	resolved []*rxNode // resolved capture groups (multiple under ES2025 named-duplicate alternatives)
}

func isLookaround(n *rxNode) bool         { return n != nil && n.kind == nkLookaround }
func isLookbehind(n *rxNode) bool         { return n != nil && n.kind == nkLookaround && !n.isAhead }
func isNegativeLookaround(n *rxNode) bool { return n != nil && n.kind == nkLookaround && n.negate }

// parsePattern parses a regex pattern and returns the root Pattern node and a
// flat list of backreference nodes. Returns ok=false on syntax errors.
func parsePattern(pattern string, flags utils.RegexFlags) (root *rxNode, backrefs []*rxNode, ok bool) {
	totalGroups, hasNamedGroups, scanOk := scanGroups(pattern, flags)
	if !scanOk {
		return nil, nil, false
	}

	p := &rxParser{
		pattern:        pattern,
		flags:          flags,
		totalGroups:    totalGroups,
		hasNamedGroups: hasNamedGroups,
		namedGroups:    map[string][]*rxNode{},
	}

	root = &rxNode{kind: nkPattern, start: 0, end: len(pattern), raw: pattern}
	if !p.parseAlternatives(root, false) {
		return nil, nil, false
	}
	if p.pos != len(pattern) {
		return nil, nil, false
	}

	for _, br := range p.backrefs {
		if br.refName != "" {
			br.resolved = p.namedGroups[br.refName]
		} else if br.refNum > 0 && br.refNum <= len(p.numberedGroups) {
			br.resolved = []*rxNode{p.numberedGroups[br.refNum-1]}
		}
		if len(br.resolved) == 0 {
			return nil, nil, false
		}
	}

	return root, p.backrefs, true
}

type rxParser struct {
	pattern        string
	flags          utils.RegexFlags
	pos            int
	totalGroups    int
	hasNamedGroups bool

	numberedGroups []*rxNode
	namedGroups    map[string][]*rxNode
	backrefs       []*rxNode
}

// parseAlternatives consumes alternatives separated by `|`, ending at either
// `)` (when expectClose) or end-of-input.
func (p *rxParser) parseAlternatives(parent *rxNode, expectClose bool) bool {
	alt := &rxNode{kind: nkAlternative, start: p.pos, parent: parent}
	parent.children = append(parent.children, alt)

	for p.pos < len(p.pattern) {
		c := p.pattern[p.pos]
		if c == ')' {
			if !expectClose {
				return false
			}
			alt.end = p.pos
			alt.raw = p.pattern[alt.start:alt.end]
			return true
		}
		if c == '|' {
			alt.end = p.pos
			alt.raw = p.pattern[alt.start:alt.end]
			p.pos++
			alt = &rxNode{kind: nkAlternative, start: p.pos, parent: parent}
			parent.children = append(parent.children, alt)
			continue
		}
		if !p.parseTerm(alt) {
			return false
		}
	}

	if expectClose {
		return false
	}

	alt.end = p.pos
	alt.raw = p.pattern[alt.start:alt.end]
	return true
}

func (p *rxParser) parseTerm(parent *rxNode) bool {
	c := p.pattern[p.pos]
	switch c {
	case '(':
		if !p.parseGroup(parent) {
			return false
		}
	case '[':
		end, ok := utils.ClassEnd(p.pattern, p.pos, p.flags)
		if !ok {
			return false
		}
		p.pos = end
	case '\\':
		if !p.parseEscape(parent) {
			return false
		}
	case '*', '+', '?':
		// Standalone quantifier without an operand — invalid, but for
		// robustness against patterns that fail to parse cleanly we just
		// consume it. (Top-level `parsePattern` rejects on remaining errors.)
		return false
	case '{':
		// Under u/v mode, a standalone `{` (one not part of a `{n}`/`{n,m}`
		// quantifier and not following a target) is a syntax error in
		// regexpp's strict parser. Skip the regex.
		if p.flags.UV() {
			return false
		}
		p.pos++
	default:
		_, w := utf8.DecodeRuneInString(p.pattern[p.pos:])
		if w == 0 {
			return false
		}
		p.pos += w
	}
	p.skipQuantifier()
	return true
}

func (p *rxParser) skipQuantifier() {
	if p.pos >= len(p.pattern) {
		return
	}
	c := p.pattern[p.pos]
	switch c {
	case '?', '*', '+':
		p.pos++
	case '{':
		save := p.pos
		i := p.pos + 1
		nStart := i
		for i < len(p.pattern) && isDigit(p.pattern[i]) {
			i++
		}
		if i == nStart {
			return
		}
		if i < len(p.pattern) && p.pattern[i] == ',' {
			i++
			for i < len(p.pattern) && isDigit(p.pattern[i]) {
				i++
			}
		}
		if i < len(p.pattern) && p.pattern[i] == '}' {
			p.pos = i + 1
		} else {
			p.pos = save
			return
		}
	default:
		return
	}
	if p.pos < len(p.pattern) && p.pattern[p.pos] == '?' {
		p.pos++
	}
}

func (p *rxParser) parseGroup(parent *rxNode) bool {
	start := p.pos
	p.pos++ // consume '('

	var node *rxNode

	if p.pos < len(p.pattern) && p.pattern[p.pos] == '?' {
		if p.pos+1 >= len(p.pattern) {
			return false
		}
		c := p.pattern[p.pos+1]
		switch c {
		case ':':
			p.pos += 2
			node = &rxNode{kind: nkGroup, start: start, parent: parent}
		case '=':
			p.pos += 2
			node = &rxNode{kind: nkLookaround, start: start, parent: parent, isAhead: true, negate: false}
		case '!':
			p.pos += 2
			node = &rxNode{kind: nkLookaround, start: start, parent: parent, isAhead: true, negate: true}
		case '<':
			if p.pos+2 >= len(p.pattern) {
				return false
			}
			switch p.pattern[p.pos+2] {
			case '=':
				p.pos += 3
				node = &rxNode{kind: nkLookaround, start: start, parent: parent, isAhead: false, negate: false}
			case '!':
				p.pos += 3
				node = &rxNode{kind: nkLookaround, start: start, parent: parent, isAhead: false, negate: true}
			default:
				// (?<name>...)
				p.pos += 2 // consume `?<`
				nameEnd := p.pos
				for nameEnd < len(p.pattern) && p.pattern[nameEnd] != '>' {
					nameEnd++
				}
				if nameEnd >= len(p.pattern) || nameEnd == p.pos {
					return false
				}
				name := p.pattern[p.pos:nameEnd]
				p.pos = nameEnd + 1
				p.numberedGroups = append(p.numberedGroups, nil)
				num := len(p.numberedGroups)
				node = &rxNode{kind: nkCapturingGroup, start: start, parent: parent, name: name, number: num}
				p.numberedGroups[num-1] = node
				p.namedGroups[name] = append(p.namedGroups[name], node)
			}
		default:
			return false
		}
	} else {
		p.numberedGroups = append(p.numberedGroups, nil)
		num := len(p.numberedGroups)
		node = &rxNode{kind: nkCapturingGroup, start: start, parent: parent, number: num}
		p.numberedGroups[num-1] = node
	}

	parent.children = append(parent.children, node)

	if !p.parseAlternatives(node, true) {
		return false
	}
	if p.pos >= len(p.pattern) || p.pattern[p.pos] != ')' {
		return false
	}
	p.pos++ // consume ')'

	node.end = p.pos
	node.raw = p.pattern[node.start:node.end]
	return true
}

// parseEscape handles a `\`-prefixed sequence at p.pos. Identifies backrefs;
// otherwise just advances past the escape.
func (p *rxParser) parseEscape(parent *rxNode) bool {
	if p.pos+1 >= len(p.pattern) {
		return false
	}
	next := p.pattern[p.pos+1]

	switch {
	case next == 'k':
		// \k<name> — only treated as a backref under u/v mode OR when at
		// least one named group exists in the pattern. Otherwise it's an
		// identity escape (`k`) followed by literals.
		if p.flags.UV() || p.hasNamedGroups {
			if p.pos+2 < len(p.pattern) && p.pattern[p.pos+2] == '<' {
				nameEnd := p.pos + 3
				for nameEnd < len(p.pattern) && p.pattern[nameEnd] != '>' {
					nameEnd++
				}
				if nameEnd < len(p.pattern) && nameEnd > p.pos+3 {
					name := p.pattern[p.pos+3 : nameEnd]
					bref := &rxNode{
						kind:    nkBackref,
						start:   p.pos,
						end:     nameEnd + 1,
						raw:     p.pattern[p.pos : nameEnd+1],
						parent:  parent,
						refName: name,
					}
					parent.children = append(parent.children, bref)
					p.backrefs = append(p.backrefs, bref)
					p.pos = nameEnd + 1
					return true
				}
				if p.flags.UV() {
					return false
				}
			} else if p.flags.UV() {
				return false
			}
		}
		// Identity escape `\k`
		p.pos += 2
		return true

	case next >= '1' && next <= '9':
		end := p.pos + 1
		for end < len(p.pattern) && isDigit(p.pattern[end]) {
			end++
		}
		n, _ := strconv.Atoi(p.pattern[p.pos+1 : end])
		if p.flags.UV() || n <= p.totalGroups {
			bref := &rxNode{
				kind:   nkBackref,
				start:  p.pos,
				end:    end,
				raw:    p.pattern[p.pos:end],
				parent: parent,
				refNum: n,
			}
			parent.children = append(parent.children, bref)
			p.backrefs = append(p.backrefs, bref)
			p.pos = end
			return true
		}
		// Octal escape — just consume.
		p.pos = end
		return true

	default:
		// All other escapes (\x, \u, \c, \p, \P, \d, identity escapes, etc.)
		// — defer to the shared utility, which knows the precise byte width
		// for each form under each flag mode.
		step, ok := utils.SkipPatternEscape(p.pattern, p.pos, p.flags)
		if !ok {
			return false
		}
		p.pos += step
		return true
	}
}

// scanGroups walks the pattern to count capturing groups and detect whether
// any are named. Skips over character classes and escapes correctly.
func scanGroups(pattern string, flags utils.RegexFlags) (totalGroups int, hasNamedGroups bool, ok bool) {
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		switch c {
		case '\\':
			step, escOk := skipBackrefAwareEscape(pattern, i, flags)
			if !escOk {
				return 0, false, false
			}
			i += step
		case '[':
			end, classOk := utils.ClassEnd(pattern, i, flags)
			if !classOk {
				return 0, false, false
			}
			i = end
		case '(':
			i++
			if i < len(pattern) && pattern[i] == '?' {
				if i+1 >= len(pattern) {
					return 0, false, false
				}
				next := pattern[i+1]
				if next == ':' || next == '=' || next == '!' {
					i += 2
					continue
				}
				if next == '<' {
					if i+2 < len(pattern) {
						after := pattern[i+2]
						if after == '=' || after == '!' {
							i += 3
							continue
						}
					}
					// Named group (?<name>...)
					i += 2 // consume `?<`
					nameEnd := i
					for nameEnd < len(pattern) && pattern[nameEnd] != '>' {
						nameEnd++
					}
					if nameEnd >= len(pattern) || nameEnd == i {
						return 0, false, false
					}
					totalGroups++
					hasNamedGroups = true
					i = nameEnd + 1
					continue
				}
				return 0, false, false
			}
			totalGroups++
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return totalGroups, hasNamedGroups, true
}

// skipBackrefAwareEscape wraps utils.SkipPatternEscape with extra handling
// for `\k<name>` so that a `>` inside the name can't be misread as part of
// surrounding syntax during the group-scan pass.
func skipBackrefAwareEscape(pattern string, i int, flags utils.RegexFlags) (int, bool) {
	if i+2 < len(pattern) && pattern[i+1] == 'k' && pattern[i+2] == '<' {
		nameEnd := i + 3
		for nameEnd < len(pattern) && pattern[nameEnd] != '>' {
			nameEnd++
		}
		if nameEnd < len(pattern) && nameEnd > i+3 {
			return nameEnd + 1 - i, true
		}
	}
	return utils.SkipPatternEscape(pattern, i, flags)
}

// pathToRoot returns [n, n.parent, …, root].
func pathToRoot(n *rxNode) []*rxNode {
	var out []*rxNode
	cur := n
	for cur != nil {
		out = append(out, cur)
		cur = cur.parent
	}
	return out
}

func nodeContains(path []*rxNode, target *rxNode) bool {
	for _, n := range path {
		if n == target {
			return true
		}
	}
	return false
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
