//! ECMAScript regular-expression syntax shared by the frozen Rsbuild rules.
//! The owned observations below come from Oxc's RegExp AST, including Annex B,
//! nested Unicode sets, capture groups and lookaround structure. No Rust-regex
//! approximation or JavaScript subprocess participates in lint execution.
use oxc_allocator::Allocator;
use oxc_regular_expression::{ConstructorParser, LiteralParser, Options, ast::*};
use oxc_span::Span;

#[derive(Clone, Debug)]
pub(super) struct Char {
    pub start: usize,
    pub end: usize,
    pub value: u32,
    pub kind: CharacterKind,
    pub class: Option<usize>,
}
#[derive(Debug)]
pub(super) struct Class {
    pub start: usize,
    pub end: usize,
    pub negative: bool,
    pub set_operation: bool,
    /// A range supplies both endpoints separated by a boundary; nested sets and
    /// shorthands interrupt character sequences, as in the frozen detectors.
    pub sequences: Vec<Vec<Char>>,
}
#[derive(Clone, Debug, PartialEq, Eq)]
pub(super) enum NodeKind {
    Root,
    Alternative,
    Group,
    Capture(Option<String>),
    Lookaround { negative: bool, backward: bool },
    Indexed(u32),
    Named(String),
    Other,
}
#[derive(Debug)]
pub(super) struct RxNode {
    pub parent: Option<usize>,
    pub kind: NodeKind,
    pub start: usize,
    pub end: usize,
}
#[derive(Default, Debug)]
pub(super) struct Parsed {
    pub chars: Vec<Char>,
    pub classes: Vec<Class>,
    pub nodes: Vec<RxNode>,
    pub groups: Vec<usize>,
    pub references: Vec<usize>,
}
#[derive(Debug)]
pub(super) struct ParseError {
    pub message: String,
    pub offset: usize,
}
fn flags(flags: &str) -> &'static str {
    match (flags.contains('u'), flags.contains('v')) {
        (true, true) => "uv",
        (true, false) => "u",
        (false, true) => "v",
        _ => "",
    }
}
pub(super) fn parse(pattern: &str, flag_text: &str) -> Result<Parsed, ParseError> {
    let allocator = Allocator::default();
    let tree = LiteralParser::new(
        &allocator,
        pattern,
        Some(flags(flag_text)),
        Options::default(),
    )
    .parse()
    .map_err(|error| ParseError {
        offset: error
            .labels
            .first()
            .map_or(pattern.len(), |label| label.offset() as usize),
        message: error.to_string(),
    })?;
    let mut out = Parsed::default();
    out.disjunction(&tree.body, None, NodeKind::Root);
    // Oxc also understands newer buffer-boundary syntax. The frozen compiler
    // and rule baseline use ES2025: these are invalid identity escapes in u/v.
    if flags(flag_text) != "" {
        for (index, byte) in pattern.bytes().enumerate() {
            if byte == b'\\'
                && matches!(pattern.as_bytes().get(index + 1), Some(b'A' | b'z' | b'Z'))
                && pattern.as_bytes()[..index]
                    .iter()
                    .rev()
                    .take_while(|&&x| x == b'\\')
                    .count()
                    % 2
                    == 0
            {
                return Err(ParseError {
                    message: "Invalid regular expression: Invalid escape".into(),
                    offset: index,
                });
            }
        }
    }
    Ok(out)
}
pub(super) fn parse_constructor(raw: &str, flag_text: &str) -> Result<Parsed, ParseError> {
    let allocator = Allocator::default();
    let raw_flags = serde_json::to_string(flags(flag_text)).unwrap();
    let tree = ConstructorParser::new(&allocator, raw, Some(&raw_flags), Options::default())
        .parse()
        .map_err(|error| ParseError {
            offset: error
                .labels
                .first()
                .map_or(raw.len(), |label| label.offset() as usize),
            message: error.to_string(),
        })?;
    let mut out = Parsed::default();
    out.disjunction(&tree.body, None, NodeKind::Root);
    Ok(out)
}
fn range(span: Span) -> (usize, usize) {
    (span.start as usize, span.end as usize)
}
impl Parsed {
    fn node(&mut self, kind: NodeKind, span: Span, parent: Option<usize>) -> usize {
        let id = self.nodes.len();
        let (start, end) = range(span);
        self.nodes.push(RxNode {
            parent,
            kind,
            start,
            end,
        });
        id
    }
    fn disjunction(
        &mut self,
        body: &Disjunction<'_>,
        parent: Option<usize>,
        kind: NodeKind,
    ) -> usize {
        let id = self.node(kind, body.span, parent);
        for alternative in &body.body {
            let child = self.node(NodeKind::Alternative, alternative.span, Some(id));
            for term in &alternative.body {
                self.term(term, child);
            }
        }
        id
    }
    fn term(&mut self, term: &Term<'_>, parent: usize) {
        match term {
            Term::Character(c) => {
                self.character(c, None);
            }
            Term::CharacterClass(c) => {
                self.class(c);
            }
            Term::CapturingGroup(g) => {
                let id = self.node(
                    NodeKind::Capture(g.name.as_ref().map(|name| name.to_string())),
                    g.span,
                    Some(parent),
                );
                self.groups.push(id);
                self.disjunction(&g.body, Some(id), NodeKind::Group);
            }
            Term::IgnoreGroup(g) => {
                let id = self.node(NodeKind::Group, g.span, Some(parent));
                self.disjunction(&g.body, Some(id), NodeKind::Group);
            }
            Term::LookAroundAssertion(g) => {
                let id = self.node(
                    NodeKind::Lookaround {
                        negative: matches!(
                            g.kind,
                            LookAroundAssertionKind::NegativeLookahead
                                | LookAroundAssertionKind::NegativeLookbehind
                        ),
                        backward: matches!(
                            g.kind,
                            LookAroundAssertionKind::Lookbehind
                                | LookAroundAssertionKind::NegativeLookbehind
                        ),
                    },
                    g.span,
                    Some(parent),
                );
                self.disjunction(&g.body, Some(id), NodeKind::Group);
            }
            Term::Quantifier(q) => {
                let id = self.node(NodeKind::Other, q.span, Some(parent));
                self.term(&q.body, id);
            }
            Term::IndexedReference(r) => {
                let id = self.node(NodeKind::Indexed(r.index), r.span, Some(parent));
                self.references.push(id);
            }
            Term::NamedReference(r) => {
                let id = self.node(NodeKind::Named(r.name.to_string()), r.span, Some(parent));
                self.references.push(id);
            }
            _ => {}
        }
    }
    fn character(&mut self, c: &Character, class: Option<usize>) -> Char {
        let (start, end) = range(c.span);
        let value = Char {
            start,
            end,
            value: c.value,
            kind: c.kind,
            class,
        };
        self.chars.push(value.clone());
        value
    }
    fn class(&mut self, c: &CharacterClass<'_>) -> usize {
        let id = self.classes.len();
        let (start, end) = range(c.span);
        self.classes.push(Class {
            start,
            end,
            negative: c.negative,
            set_operation: c.kind != CharacterClassContentsKind::Union,
            sequences: Vec::new(),
        });
        let mut sequence = Vec::new();
        let mut sequences = Vec::new();
        for element in &c.body {
            match element {
                CharacterClassContents::Character(c) => sequence.push(self.character(c, Some(id))),
                CharacterClassContents::CharacterClassRange(r) => {
                    let mut min = self.character(&r.min, Some(id));
                    min.start = r.span.start as usize;
                    min.end = r.span.end as usize;
                    sequence.push(min);
                    sequences.push(std::mem::take(&mut sequence));
                    let mut max = self.character(&r.max, Some(id));
                    max.start = r.span.start as usize;
                    max.end = r.span.end as usize;
                    sequence.push(max);
                }
                CharacterClassContents::NestedCharacterClass(nested) => {
                    sequences.push(std::mem::take(&mut sequence));
                    self.class(nested);
                }
                CharacterClassContents::ClassStringDisjunction(strings) => {
                    sequences.push(std::mem::take(&mut sequence));
                    // The escape rule visits q-string characters; misleading
                    // character-class deliberately treats the string as one atom.
                    for string in &strings.body {
                        for c in &string.body {
                            self.character(c, Some(id));
                        }
                    }
                }
                _ => {
                    sequences.push(std::mem::take(&mut sequence));
                }
            }
            if c.kind != CharacterClassContentsKind::Union {
                sequences.push(std::mem::take(&mut sequence));
            }
        }
        sequences.push(sequence);
        self.classes[id].sequences = sequences;
        id
    }
    pub(super) fn reference_problems(&self) -> Vec<(usize, usize, &'static str, usize)> {
        let mut out = Vec::new();
        for &reference in &self.references {
            let groups:Vec<_>=match &self.nodes[reference].kind {
                NodeKind::Indexed(n)=>self.groups.get(*n as usize-1).copied().into_iter().collect(),
                NodeKind::Named(name)=>self.groups.iter().copied().filter(|&id|matches!(&self.nodes[id].kind,NodeKind::Capture(Some(value)) if value==name)).collect(),_=>Vec::new(),
            };
            let mut problems = Vec::new();
            for group in groups {
                let Some(kind) = self.classify(reference, group) else {
                    problems.clear();
                    break;
                };
                problems.push((group, kind));
            }
            if problems.iter().any(|(_, kind)| *kind != "disjunctive") {
                problems.retain(|(_, kind)| *kind != "disjunctive");
            }
            if let Some(&(group, kind)) = problems.first() {
                out.push((reference, group, kind, problems.len()));
            }
        }
        out
    }
    fn ancestors(&self, mut id: usize) -> Vec<usize> {
        let mut result = vec![id];
        while let Some(parent) = self.nodes[id].parent {
            result.push(parent);
            id = parent;
        }
        result
    }
    fn classify(&self, reference: usize, group: usize) -> Option<&'static str> {
        let refs = self.ancestors(reference);
        if refs.contains(&group) {
            return Some("nested");
        }
        let groups = self.ancestors(group);
        let common = *groups.iter().find(|id| refs.contains(id))?;
        let backward = self
            .ancestors(common)
            .iter()
            .find_map(|&id| match self.nodes[id].kind {
                NodeKind::Lookaround { backward, .. } => Some(backward),
                _ => None,
            })
            .unwrap_or(false);
        let branch = *groups
            .iter()
            .find(|&&id| self.nodes[id].parent == Some(common))?;
        if self.nodes[branch].kind == NodeKind::Alternative {
            return Some("disjunctive");
        }
        if !backward && self.nodes[reference].end <= self.nodes[group].start {
            return Some("forward");
        }
        if backward && self.nodes[group].end <= self.nodes[reference].start {
            return Some("backward");
        }
        if groups.iter().take_while(|&&id| id != common).any(|&id| {
            matches!(
                self.nodes[id].kind,
                NodeKind::Lookaround { negative: true, .. }
            )
        }) {
            return Some("intoNegativeLookaround");
        }
        None
    }
}

// Unicode 17.0.0 General_Category=Mark data, from the frozen baseline
// toolchain unicode/tables.go. Copyright 2009 The Go Authors.
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
// 3. Neither the name of Google Inc. nor the names of its contributors may be
//    used to endorse or promote products derived from this software without
//    specific prior written permission.
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
pub(super) fn is_mark(value: u32) -> bool {
    MARK_RANGES
        .iter()
        .any(|&(low, high, stride)| low <= value && value <= high && (value - low) % stride == 0)
}
const MARK_RANGES: &[(u32, u32, u32)] = &[
    (0x0300, 0x036f, 1),
    (0x0483, 0x0489, 1),
    (0x0591, 0x05bd, 1),
    (0x05bf, 0x05c1, 2),
    (0x05c2, 0x05c4, 2),
    (0x05c5, 0x05c7, 2),
    (0x0610, 0x061a, 1),
    (0x064b, 0x065f, 1),
    (0x0670, 0x06d6, 102),
    (0x06d7, 0x06dc, 1),
    (0x06df, 0x06e4, 1),
    (0x06e7, 0x06e8, 1),
    (0x06ea, 0x06ed, 1),
    (0x0711, 0x0730, 31),
    (0x0731, 0x074a, 1),
    (0x07a6, 0x07b0, 1),
    (0x07eb, 0x07f3, 1),
    (0x07fd, 0x0816, 25),
    (0x0817, 0x0819, 1),
    (0x081b, 0x0823, 1),
    (0x0825, 0x0827, 1),
    (0x0829, 0x082d, 1),
    (0x0859, 0x085b, 1),
    (0x0897, 0x089f, 1),
    (0x08ca, 0x08e1, 1),
    (0x08e3, 0x0903, 1),
    (0x093a, 0x093c, 1),
    (0x093e, 0x094f, 1),
    (0x0951, 0x0957, 1),
    (0x0962, 0x0963, 1),
    (0x0981, 0x0983, 1),
    (0x09bc, 0x09be, 2),
    (0x09bf, 0x09c4, 1),
    (0x09c7, 0x09c8, 1),
    (0x09cb, 0x09cd, 1),
    (0x09d7, 0x09e2, 11),
    (0x09e3, 0x09fe, 27),
    (0x0a01, 0x0a03, 1),
    (0x0a3c, 0x0a3e, 2),
    (0x0a3f, 0x0a42, 1),
    (0x0a47, 0x0a48, 1),
    (0x0a4b, 0x0a4d, 1),
    (0x0a51, 0x0a70, 31),
    (0x0a71, 0x0a75, 4),
    (0x0a81, 0x0a83, 1),
    (0x0abc, 0x0abe, 2),
    (0x0abf, 0x0ac5, 1),
    (0x0ac7, 0x0ac9, 1),
    (0x0acb, 0x0acd, 1),
    (0x0ae2, 0x0ae3, 1),
    (0x0afa, 0x0aff, 1),
    (0x0b01, 0x0b03, 1),
    (0x0b3c, 0x0b3e, 2),
    (0x0b3f, 0x0b44, 1),
    (0x0b47, 0x0b48, 1),
    (0x0b4b, 0x0b4d, 1),
    (0x0b55, 0x0b57, 1),
    (0x0b62, 0x0b63, 1),
    (0x0b82, 0x0bbe, 60),
    (0x0bbf, 0x0bc2, 1),
    (0x0bc6, 0x0bc8, 1),
    (0x0bca, 0x0bcd, 1),
    (0x0bd7, 0x0c00, 41),
    (0x0c01, 0x0c04, 1),
    (0x0c3c, 0x0c3e, 2),
    (0x0c3f, 0x0c44, 1),
    (0x0c46, 0x0c48, 1),
    (0x0c4a, 0x0c4d, 1),
    (0x0c55, 0x0c56, 1),
    (0x0c62, 0x0c63, 1),
    (0x0c81, 0x0c83, 1),
    (0x0cbc, 0x0cbe, 2),
    (0x0cbf, 0x0cc4, 1),
    (0x0cc6, 0x0cc8, 1),
    (0x0cca, 0x0ccd, 1),
    (0x0cd5, 0x0cd6, 1),
    (0x0ce2, 0x0ce3, 1),
    (0x0cf3, 0x0d00, 13),
    (0x0d01, 0x0d03, 1),
    (0x0d3b, 0x0d3c, 1),
    (0x0d3e, 0x0d44, 1),
    (0x0d46, 0x0d48, 1),
    (0x0d4a, 0x0d4d, 1),
    (0x0d57, 0x0d62, 11),
    (0x0d63, 0x0d81, 30),
    (0x0d82, 0x0d83, 1),
    (0x0dca, 0x0dcf, 5),
    (0x0dd0, 0x0dd4, 1),
    (0x0dd6, 0x0dd8, 2),
    (0x0dd9, 0x0ddf, 1),
    (0x0df2, 0x0df3, 1),
    (0x0e31, 0x0e34, 3),
    (0x0e35, 0x0e3a, 1),
    (0x0e47, 0x0e4e, 1),
    (0x0eb1, 0x0eb4, 3),
    (0x0eb5, 0x0ebc, 1),
    (0x0ec8, 0x0ece, 1),
    (0x0f18, 0x0f19, 1),
    (0x0f35, 0x0f39, 2),
    (0x0f3e, 0x0f3f, 1),
    (0x0f71, 0x0f84, 1),
    (0x0f86, 0x0f87, 1),
    (0x0f8d, 0x0f97, 1),
    (0x0f99, 0x0fbc, 1),
    (0x0fc6, 0x102b, 101),
    (0x102c, 0x103e, 1),
    (0x1056, 0x1059, 1),
    (0x105e, 0x1060, 1),
    (0x1062, 0x1064, 1),
    (0x1067, 0x106d, 1),
    (0x1071, 0x1074, 1),
    (0x1082, 0x108d, 1),
    (0x108f, 0x109a, 11),
    (0x109b, 0x109d, 1),
    (0x135d, 0x135f, 1),
    (0x1712, 0x1715, 1),
    (0x1732, 0x1734, 1),
    (0x1752, 0x1753, 1),
    (0x1772, 0x1773, 1),
    (0x17b4, 0x17d3, 1),
    (0x17dd, 0x180b, 46),
    (0x180c, 0x180d, 1),
    (0x180f, 0x1885, 118),
    (0x1886, 0x18a9, 35),
    (0x1920, 0x192b, 1),
    (0x1930, 0x193b, 1),
    (0x1a17, 0x1a1b, 1),
    (0x1a55, 0x1a5e, 1),
    (0x1a60, 0x1a7c, 1),
    (0x1a7f, 0x1ab0, 49),
    (0x1ab1, 0x1add, 1),
    (0x1ae0, 0x1aeb, 1),
    (0x1b00, 0x1b04, 1),
    (0x1b34, 0x1b44, 1),
    (0x1b6b, 0x1b73, 1),
    (0x1b80, 0x1b82, 1),
    (0x1ba1, 0x1bad, 1),
    (0x1be6, 0x1bf3, 1),
    (0x1c24, 0x1c37, 1),
    (0x1cd0, 0x1cd2, 1),
    (0x1cd4, 0x1ce8, 1),
    (0x1ced, 0x1cf4, 7),
    (0x1cf7, 0x1cf9, 1),
    (0x1dc0, 0x1dff, 1),
    (0x20d0, 0x20f0, 1),
    (0x2cef, 0x2cf1, 1),
    (0x2d7f, 0x2de0, 97),
    (0x2de1, 0x2dff, 1),
    (0x302a, 0x302f, 1),
    (0x3099, 0x309a, 1),
    (0xa66f, 0xa672, 1),
    (0xa674, 0xa67d, 1),
    (0xa69e, 0xa69f, 1),
    (0xa6f0, 0xa6f1, 1),
    (0xa802, 0xa806, 4),
    (0xa80b, 0xa823, 24),
    (0xa824, 0xa827, 1),
    (0xa82c, 0xa880, 84),
    (0xa881, 0xa8b4, 51),
    (0xa8b5, 0xa8c5, 1),
    (0xa8e0, 0xa8f1, 1),
    (0xa8ff, 0xa926, 39),
    (0xa927, 0xa92d, 1),
    (0xa947, 0xa953, 1),
    (0xa980, 0xa983, 1),
    (0xa9b3, 0xa9c0, 1),
    (0xa9e5, 0xaa29, 68),
    (0xaa2a, 0xaa36, 1),
    (0xaa43, 0xaa4c, 9),
    (0xaa4d, 0xaa7b, 46),
    (0xaa7c, 0xaa7d, 1),
    (0xaab0, 0xaab2, 2),
    (0xaab3, 0xaab4, 1),
    (0xaab7, 0xaab8, 1),
    (0xaabe, 0xaabf, 1),
    (0xaac1, 0xaaeb, 42),
    (0xaaec, 0xaaef, 1),
    (0xaaf5, 0xaaf6, 1),
    (0xabe3, 0xabea, 1),
    (0xabec, 0xabed, 1),
    (0xfb1e, 0xfe00, 738),
    (0xfe01, 0xfe0f, 1),
    (0xfe20, 0xfe2f, 1),
    (0x101fd, 0x102e0, 227),
    (0x10376, 0x1037a, 1),
    (0x10a01, 0x10a03, 1),
    (0x10a05, 0x10a06, 1),
    (0x10a0c, 0x10a0f, 1),
    (0x10a38, 0x10a3a, 1),
    (0x10a3f, 0x10ae5, 166),
    (0x10ae6, 0x10d24, 574),
    (0x10d25, 0x10d27, 1),
    (0x10d69, 0x10d6d, 1),
    (0x10eab, 0x10eac, 1),
    (0x10efa, 0x10eff, 1),
    (0x10f46, 0x10f50, 1),
    (0x10f82, 0x10f85, 1),
    (0x11000, 0x11002, 1),
    (0x11038, 0x11046, 1),
    (0x11070, 0x11073, 3),
    (0x11074, 0x1107f, 11),
    (0x11080, 0x11082, 1),
    (0x110b0, 0x110ba, 1),
    (0x110c2, 0x11100, 62),
    (0x11101, 0x11102, 1),
    (0x11127, 0x11134, 1),
    (0x11145, 0x11146, 1),
    (0x11173, 0x11180, 13),
    (0x11181, 0x11182, 1),
    (0x111b3, 0x111c0, 1),
    (0x111c9, 0x111cc, 1),
    (0x111ce, 0x111cf, 1),
    (0x1122c, 0x11237, 1),
    (0x1123e, 0x11241, 3),
    (0x112df, 0x112ea, 1),
    (0x11300, 0x11303, 1),
    (0x1133b, 0x1133c, 1),
    (0x1133e, 0x11344, 1),
    (0x11347, 0x11348, 1),
    (0x1134b, 0x1134d, 1),
    (0x11357, 0x11362, 11),
    (0x11363, 0x11366, 3),
    (0x11367, 0x1136c, 1),
    (0x11370, 0x11374, 1),
    (0x113b8, 0x113c0, 1),
    (0x113c2, 0x113c5, 3),
    (0x113c7, 0x113ca, 1),
    (0x113cc, 0x113d0, 1),
    (0x113d2, 0x113e1, 15),
    (0x113e2, 0x11435, 83),
    (0x11436, 0x11446, 1),
    (0x1145e, 0x114b0, 82),
    (0x114b1, 0x114c3, 1),
    (0x115af, 0x115b5, 1),
    (0x115b8, 0x115c0, 1),
    (0x115dc, 0x115dd, 1),
    (0x11630, 0x11640, 1),
    (0x116ab, 0x116b7, 1),
    (0x1171d, 0x1172b, 1),
    (0x1182c, 0x1183a, 1),
    (0x11930, 0x11935, 1),
    (0x11937, 0x11938, 1),
    (0x1193b, 0x1193e, 1),
    (0x11940, 0x11942, 2),
    (0x11943, 0x119d1, 142),
    (0x119d2, 0x119d7, 1),
    (0x119da, 0x119e0, 1),
    (0x119e4, 0x11a01, 29),
    (0x11a02, 0x11a0a, 1),
    (0x11a33, 0x11a39, 1),
    (0x11a3b, 0x11a3e, 1),
    (0x11a47, 0x11a51, 10),
    (0x11a52, 0x11a5b, 1),
    (0x11a8a, 0x11a99, 1),
    (0x11b60, 0x11b67, 1),
    (0x11c2f, 0x11c36, 1),
    (0x11c38, 0x11c3f, 1),
    (0x11c92, 0x11ca7, 1),
    (0x11ca9, 0x11cb6, 1),
    (0x11d31, 0x11d36, 1),
    (0x11d3a, 0x11d3c, 2),
    (0x11d3d, 0x11d3f, 2),
    (0x11d40, 0x11d45, 1),
    (0x11d47, 0x11d8a, 67),
    (0x11d8b, 0x11d8e, 1),
    (0x11d90, 0x11d91, 1),
    (0x11d93, 0x11d97, 1),
    (0x11ef3, 0x11ef6, 1),
    (0x11f00, 0x11f01, 1),
    (0x11f03, 0x11f34, 49),
    (0x11f35, 0x11f3a, 1),
    (0x11f3e, 0x11f42, 1),
    (0x11f5a, 0x13440, 5350),
    (0x13447, 0x13455, 1),
    (0x1611e, 0x1612f, 1),
    (0x16af0, 0x16af4, 1),
    (0x16b30, 0x16b36, 1),
    (0x16f4f, 0x16f51, 2),
    (0x16f52, 0x16f87, 1),
    (0x16f8f, 0x16f92, 1),
    (0x16fe4, 0x16ff0, 12),
    (0x16ff1, 0x1bc9d, 19628),
    (0x1bc9e, 0x1cf00, 4706),
    (0x1cf01, 0x1cf2d, 1),
    (0x1cf30, 0x1cf46, 1),
    (0x1d165, 0x1d169, 1),
    (0x1d16d, 0x1d172, 1),
    (0x1d17b, 0x1d182, 1),
    (0x1d185, 0x1d18b, 1),
    (0x1d1aa, 0x1d1ad, 1),
    (0x1d242, 0x1d244, 1),
    (0x1da00, 0x1da36, 1),
    (0x1da3b, 0x1da6c, 1),
    (0x1da75, 0x1da84, 15),
    (0x1da9b, 0x1da9f, 1),
    (0x1daa1, 0x1daaf, 1),
    (0x1e000, 0x1e006, 1),
    (0x1e008, 0x1e018, 1),
    (0x1e01b, 0x1e021, 1),
    (0x1e023, 0x1e024, 1),
    (0x1e026, 0x1e02a, 1),
    (0x1e08f, 0x1e130, 161),
    (0x1e131, 0x1e136, 1),
    (0x1e2ae, 0x1e2ec, 62),
    (0x1e2ed, 0x1e2ef, 1),
    (0x1e4ec, 0x1e4ef, 1),
    (0x1e5ee, 0x1e5ef, 1),
    (0x1e6e3, 0x1e6e6, 3),
    (0x1e6ee, 0x1e6ef, 1),
    (0x1e6f5, 0x1e8d0, 475),
    (0x1e8d1, 0x1e8d6, 1),
    (0x1e944, 0x1e94a, 1),
    (0xe0100, 0xe01ef, 1),
];
/// Keep the original constructor-validation error contract. The grammar is
/// parsed first; error classification translates AST parser failures into the
/// baseline compatibility layer's public diagnostics rather than exposing the
/// dependency's own wording.
pub(super) fn legacy_error(pattern: &str, flags: &str) -> Option<String> {
    let parsed = parse(pattern, flags);
    if let Ok(parsed) = &parsed {
        let mut names = std::collections::BTreeSet::new();
        for &id in &parsed.groups {
            if let NodeKind::Capture(Some(name)) = &parsed.nodes[id].kind {
                if !names.insert(name) {
                    return Some(format!(
                        "unsupported regexp syntax: duplicate capture name {}",
                        serde_json::to_string(name).unwrap()
                    ));
                }
            }
        }
        return None;
    }
    let error = parsed.unwrap_err();
    let at = error.offset.min(pattern.len());
    if pattern.ends_with('\\') && pattern.bytes().rev().take_while(|&b| b == b'\\').count() % 2 == 1
    {
        return Some("unsupported regexp syntax: \\ at the end of the pattern".into());
    }
    if error.message.contains("Unterminated character class")
        || error.message.contains("Unterminated class")
    {
        return Some("unsupported regexp syntax: [ that no ] closes".into());
    }
    if error.message.contains("Unterminated group")
        || error.message.contains("Unterminated capturing group")
    {
        return Some(format!("missing closing ) in `{pattern}`"));
    }
    if error.message.contains("modifier")
        || error.message.contains("Modifier")
        || error.message.contains("group specifier")
        || error.message.contains("Group specifier")
    {
        let opening = pattern[..at].rfind("(?").or_else(|| pattern.find("(?"));
        if let Some(opening) = opening {
            let character = pattern[opening + 2..].chars().next().unwrap_or(' ');
            return Some(format!("unsupported regexp syntax: (?{character}"));
        }
    }
    // The original adapter rejects quantification of assertions before calling
    // regexp2. Locate that token from the grammar error's labelled source span.
    if error.message.contains("quantifier") || error.message.contains("Quantifier") {
        let index = (at..pattern.len())
            .find(|&i| matches!(pattern.as_bytes()[i], b'*' | b'+' | b'?' | b'{'))
            .or_else(|| {
                (0..at)
                    .rev()
                    .find(|&i| matches!(pattern.as_bytes()[i], b'*' | b'+' | b'?' | b'{'))
            });
        if let Some(index) = index {
            let before = &pattern[..index];
            if before.ends_with('^')
                || before.ends_with('$')
                || before.ends_with("\\b")
                || before.ends_with("\\B")
                || before.ends_with(')')
            {
                let width = if pattern.as_bytes()[index] == b'{' {
                    pattern[index..].find('}').map_or(1, |p| p + 1)
                } else {
                    1
                };
                return Some(format!(
                    "unsupported regexp syntax: {} repeats an assertion",
                    &pattern[index..index + width]
                ));
            }
            return Some(format!(
                "missing argument to repetition operator: `{}` in `{pattern}`",
                pattern.as_bytes()[index] as char
            ));
        }
    }
    if flags.contains('u') {
        // Oxc identifies the invalid escape; emit the compatibility layer's
        // spelling. Non-unicode reparsing distinguishes identity escapes from
        // malformed structural syntax without guessing a regexp grammar.
        if let Ok(legacy) = parse(pattern, "") {
            for character in legacy.chars {
                let Some(raw) = pattern.get(character.start..character.end) else {
                    continue;
                };
                if let Some(value) = raw.strip_prefix('\\') {
                    if value.len() == 1 && !"^$\\.*+?()[]{}|/bBdfnrstvwDsSW".contains(value) {
                        return Some(format!("unsupported regexp syntax: \\{value}"));
                    }
                }
            }
        }
    }
    if error.message.contains("Invalid escape") || error.message.contains("escape") {
        let index = pattern[..at]
            .rfind('\\')
            .or_else(|| pattern[at..].find('\\').map(|i| at + i));
        if let Some(index) = index {
            let next = pattern[index + 1..].chars().next().unwrap_or(' ');
            let detail = match next {
                'x' | 'u' => format!("\\{next} naming no character"),
                'c' => "\\c naming no control character".into(),
                'k' => "\\k naming no group".into(),
                _ => format!("\\{next}"),
            };
            return Some(format!("unsupported regexp syntax: {detail}"));
        }
    }
    if error.message.contains("range out of order") {
        return Some(format!("[x-y] range in reverse order in `{pattern}`"));
    }
    if error.message.contains("Could not parse the entire pattern") {
        return Some(format!("unexpected ) in `{pattern}`"));
    }
    Some(format!(
        "{} in `{pattern}`",
        error
            .message
            .trim_start_matches("Invalid regular expression: ")
    ))
}
