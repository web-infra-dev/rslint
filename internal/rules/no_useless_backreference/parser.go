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
	kind   rxNodeKind
	start  int // byte offset of the construct's first char in the pattern
	end    int // byte offset one past the last char
	raw    string
	parent *rxNode

	// for lookarounds
	isAhead bool
	negate  bool

	// for backreferences
	refName  string    // for \k<name>
	refNum   int       // for \N (1-based, 0 if name-based)
	resolved []*rxNode // resolved capture groups (multiple under ES2025 named-duplicate alternatives)
}

func isLookaround(n *rxNode) bool         { return n != nil && n.kind == nkLookaround }
func isLookbehind(n *rxNode) bool         { return n != nil && n.kind == nkLookaround && !n.isAhead }
func isNegativeLookaround(n *rxNode) bool { return n != nil && n.kind == nkLookaround && n.negate }

// parsePattern parses a regex pattern and returns the root Pattern node and a
// flat list of backreference nodes. Returns ok=false on syntax errors.
func parsePattern(pattern string, flags utils.RegexFlags) (root *rxNode, backrefs []*rxNode, ok bool) {
	totalGroups, hasNamedGroups, nodeCapacity, backrefCapacity, scanOk := scanGroups(pattern, flags)
	if !scanOk {
		return nil, nil, false
	}

	p := &rxParser{
		pattern:        pattern,
		flags:          flags,
		totalGroups:    totalGroups,
		hasNamedGroups: hasNamedGroups,
		nodes:          make([]rxNode, 0, nodeCapacity),
		numberedGroups: make([]*rxNode, 0, totalGroups),
		backrefs:       make([]*rxNode, 0, backrefCapacity),
	}
	if hasNamedGroups {
		p.namedGroups = make(map[string][]*rxNode, totalGroups)
	}

	root = p.newNode(rxNode{kind: nkPattern})
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
			br.resolved = p.numberedGroups[br.refNum-1 : br.refNum]
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
	nodes          []rxNode
}

// newNode allocates parser nodes from one pre-sized backing array. scanGroups
// computes an upper bound before parsing, so pointers into nodes stay stable
// and a regex pays for one node allocation instead of one allocation per
// structural construct. The fallback keeps pointer stability if a future
// parser extension creates a node the scan does not yet count.
func (p *rxParser) newNode(node rxNode) *rxNode {
	if len(p.nodes) == cap(p.nodes) {
		result := new(rxNode)
		*result = node
		return result
	}
	p.nodes = p.nodes[:len(p.nodes)+1]
	result := &p.nodes[len(p.nodes)-1]
	*result = node
	return result
}

// parseAlternatives consumes alternatives separated by `|`, ending at either
// `)` (when expectClose) or end-of-input.
func (p *rxParser) parseAlternatives(parent *rxNode, expectClose bool) bool {
	alt := p.newNode(rxNode{kind: nkAlternative, parent: parent})

	for p.pos < len(p.pattern) {
		c := p.pattern[p.pos]
		if c == ')' {
			if !expectClose {
				return false
			}
			return true
		}
		if c == '|' {
			p.pos++
			alt = p.newNode(rxNode{kind: nkAlternative, parent: parent})
			continue
		}
		if !p.parseTerm(alt) {
			return false
		}
	}

	if expectClose {
		return false
	}

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
			node = p.newNode(rxNode{kind: nkGroup, start: start, parent: parent})
		case '=':
			p.pos += 2
			node = p.newNode(rxNode{kind: nkLookaround, start: start, parent: parent, isAhead: true})
		case '!':
			p.pos += 2
			node = p.newNode(rxNode{kind: nkLookaround, start: start, parent: parent, isAhead: true, negate: true})
		case '<':
			if p.pos+2 >= len(p.pattern) {
				return false
			}
			switch p.pattern[p.pos+2] {
			case '=':
				p.pos += 3
				node = p.newNode(rxNode{kind: nkLookaround, start: start, parent: parent})
			case '!':
				p.pos += 3
				node = p.newNode(rxNode{kind: nkLookaround, start: start, parent: parent, negate: true})
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
				node = p.newNode(rxNode{kind: nkCapturingGroup, start: start, parent: parent})
				p.numberedGroups[num-1] = node
				p.namedGroups[name] = append(p.namedGroups[name], node)
			}
		default:
			return false
		}
	} else {
		p.numberedGroups = append(p.numberedGroups, nil)
		num := len(p.numberedGroups)
		node = p.newNode(rxNode{kind: nkCapturingGroup, start: start, parent: parent})
		p.numberedGroups[num-1] = node
	}

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
					bref := p.newNode(rxNode{
						kind:    nkBackref,
						start:   p.pos,
						end:     nameEnd + 1,
						raw:     p.pattern[p.pos : nameEnd+1],
						parent:  parent,
						refName: name,
					})
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
			bref := p.newNode(rxNode{
				kind:   nkBackref,
				start:  p.pos,
				end:    end,
				raw:    p.pattern[p.pos:end],
				parent: parent,
				refNum: n,
			})
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
func scanGroups(pattern string, flags utils.RegexFlags) (
	totalGroups int,
	hasNamedGroups bool,
	nodeCapacity int,
	backrefCapacity int,
	ok bool,
) {
	// Pattern and its first alternative always exist. Each group contributes
	// its own node and first alternative; each `|` contributes another
	// alternative; and each backreference-looking escape can contribute at
	// most one backreference node. This is an upper bound because some numeric
	// escapes are later classified as octal.
	nodeCapacity = 2
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		switch c {
		case '\\':
			if i+1 < len(pattern) &&
				(pattern[i+1] == 'k' || pattern[i+1] >= '1' && pattern[i+1] <= '9') {
				nodeCapacity++
				backrefCapacity++
			}
			step, escOk := skipBackrefAwareEscape(pattern, i, flags)
			if !escOk {
				return 0, false, 0, 0, false
			}
			i += step
		case '[':
			end, classOk := utils.ClassEnd(pattern, i, flags)
			if !classOk {
				return 0, false, 0, 0, false
			}
			i = end
		case '(':
			nodeCapacity += 2
			i++
			if i < len(pattern) && pattern[i] == '?' {
				if i+1 >= len(pattern) {
					return 0, false, 0, 0, false
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
						return 0, false, 0, 0, false
					}
					totalGroups++
					hasNamedGroups = true
					i = nameEnd + 1
					continue
				}
				return 0, false, 0, 0, false
			}
			totalGroups++
		case '|':
			nodeCapacity++
			i++
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return totalGroups, hasNamedGroups, nodeCapacity, backrefCapacity, true
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

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
