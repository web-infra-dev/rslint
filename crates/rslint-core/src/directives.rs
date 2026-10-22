//! Lint comment suppression, ported from the frozen rslint DisableManager.
//!
//! Only parser-owned comments participate. Locations are ECMAScript lines and
//! UTF-16 columns; source ranges are UTF-8 bytes. Unlike ESLint's unused-disable
//! option, the frozen rslint backend neither reports nor tracks unused comments.
//! Other directive categories are retained explicitly for their own consumers.

use std::{
    cell::{Cell, RefCell},
    collections::{BTreeMap, HashMap},
    marker::PhantomData,
    rc::Rc,
};

use tsrs_ast::{Kind, SourceFile};
use tsrs_core::P;

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub(crate) enum AuxiliaryKind {
    InlineConfig,
    Globals,
    Exported,
    Unknown,
}

#[derive(Debug, Eq, PartialEq)]
pub(crate) struct AuxiliaryDirective {
    pub kind: AuxiliaryKind,
    /// Full comment range in UTF-8 bytes, including its delimiters.
    pub range: [usize; 2],
    /// Trimmed content excluding delimiters, retaining options and reasons.
    pub content: String,
    /// Zero-based ECMAScript line and UTF-16 column of the opening slash.
    pub location: [usize; 2],
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub(crate) enum GlobalAccess {
    Off,
    Readonly,
    Writable,
}

impl GlobalAccess {
    pub(crate) fn is_declared(self) -> bool {
        self != Self::Off
    }
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub(crate) struct GlobalDeclaration {
    pub name: String,
    /// Final setting after every directive, including a final `off`.
    pub access: GlobalAccess,
    /// One exact name range per comment, retaining the first duplicate in it.
    pub name_ranges: Vec<[usize; 2]>,
}

struct GlobalsCache {
    scope_id: u64,
    // An opaque equality key only: never converted back into an arena pointer.
    file_identity: usize,
    access: HashMap<String, GlobalAccess>,
    declarations: Vec<GlobalDeclaration>,
}

thread_local! {
    static GLOBAL_SCOPES: RefCell<Vec<GlobalsCache>> = const { RefCell::new(Vec::new()) };
    static NEXT_SCOPE_ID: Cell<u64> = const { Cell::new(0) };
    #[cfg(test)]
    static DIRECTIVE_PARSE_COUNT: Cell<usize> = const { Cell::new(0) };
}

/// Scope-owned copied metadata only; neither the guard nor thread-local entries
/// retain a compiler arena pointer. The guard cannot move to another thread.
/// Its borrow also keeps the parsed directives alive until the visit ends.
pub(crate) struct GlobalsScope<'a> {
    scope_id: u64,
    _directives: PhantomData<&'a Directives>,
    _thread: PhantomData<Rc<()>>,
}

impl Drop for GlobalsScope<'_> {
    fn drop(&mut self) {
        GLOBAL_SCOPES.with(|scopes| {
            let mut scopes = scopes.borrow_mut();
            if let Some(index) = scopes
                .iter()
                .position(|entry| entry.scope_id == self.scope_id)
            {
                scopes.remove(index);
            }
        });
    }
}

fn file_identity(file: P<SourceFile>) -> usize {
    std::ptr::from_ref(&*file) as usize
}

/// Only authored inline settings participate; `None` preserves language defaults.
/// Parser-owned comment trivia excludes lookalikes in strings, regexes and JSX.
pub(crate) fn global_setting(file: P<SourceFile>, name: &str) -> Option<bool> {
    global_access(file, name).map(GlobalAccess::is_declared)
}

pub(crate) fn global_access(file: P<SourceFile>, name: &str) -> Option<GlobalAccess> {
    let identity = file_identity(file);
    if let Some(access) = GLOBAL_SCOPES.with(|scopes| {
        scopes
            .borrow()
            .iter()
            .rev()
            .find(|entry| entry.file_identity == identity)
            .map(|entry| entry.access.get(name).copied())
    }) {
        // A known file without this name is a cache hit as well.
        return access;
    }
    // Direct rule/parser tests may run without the normal file-visit scope.
    Directives::parse(file)
        .ok()?
        .globals
        .iter()
        .find(|entry| entry.name == name)
        .map(|entry| entry.access)
}

pub(crate) fn global_declarations(file: P<SourceFile>) -> Result<Vec<GlobalDeclaration>, String> {
    let identity = file_identity(file);
    if let Some(declarations) = GLOBAL_SCOPES.with(|scopes| {
        scopes
            .borrow()
            .iter()
            .rev()
            .find(|entry| entry.file_identity == identity)
            .map(|entry| entry.declarations.clone())
    }) {
        return Ok(declarations);
    }
    Directives::parse(file).map(|directives| directives.globals)
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
enum KindOfSuppression {
    Disable,
    Enable,
    Line,
    NextLine,
}

#[derive(Debug)]
struct BlockDirective {
    location: [usize; 2],
    disable: bool,
    // Empty means all rules. A literal "*" is not a block wildcard in rslint.
    rules: Vec<String>,
}

#[derive(Debug)]
pub(crate) struct Directives {
    // This owns no source/AST pointer and remains valid after the parser arena
    // is released. Entries are byte offsets; columns are projected on demand.
    line_starts: Vec<usize>,
    blocks: Vec<BlockDirective>,
    lines: BTreeMap<usize, Vec<String>>,
    auxiliary: Vec<AuxiliaryDirective>,
    globals: Vec<GlobalDeclaration>,
}

impl Directives {
    pub(crate) fn parse(file: P<SourceFile>) -> Result<Self, String> {
        #[cfg(test)]
        DIRECTIVE_PARSE_COUNT.with(|count| count.set(count.get() + 1));
        let text = file.text();
        let mut result = Self {
            line_starts: Vec::new(),
            blocks: Vec::new(),
            lines: BTreeMap::new(),
            auxiliary: Vec::new(),
            globals: Vec::new(),
        };
        // A cheap negative check avoids a full token walk for ordinary files.
        // False positives in strings and regexes still use the AST collector.
        if !["eslint", "rslint", "global", "exported"]
            .iter()
            .any(|label| text.contains(label))
        {
            return Ok(result);
        }
        result.line_starts = line_starts(text);
        let mut comments = Vec::new();
        crate::comments::for_each_comment(file, |comment| {
            comments.push(comment);
            Ok(())
        })?;
        // A comment may be both the previous token's trailing trivia and the
        // next token's leading trivia. Traversal order is not source order.
        comments.sort_by_key(|comment| (comment.pos(), comment.end()));
        comments.dedup_by_key(|comment| (comment.pos(), comment.end()));
        for comment in comments {
            let start = comment.pos() as usize;
            let end = comment.end() as usize;
            let Some(raw) = text.get(start..end) else {
                return Err("comment has an invalid source range".into());
            };
            let Some(content) = raw.get(2..) else {
                return Err("comment has no opening delimiter".into());
            };
            let content = if comment.kind == Kind::MultiLineCommentTrivia {
                content.strip_suffix("*/").unwrap_or(content)
            } else {
                content
            };
            let location = result.location(text, start)?;
            // This is deliberately the baseline's Go TrimSpace set, not
            // Rust/JS whitespace. Its U+0085/U+FEFF distinction is observable.
            let suppression_content = content.trim_matches(is_baseline_space);
            if let Some((kind, rules)) = match_suppression(suppression_content) {
                match kind {
                    KindOfSuppression::Disable | KindOfSuppression::Enable => {
                        result.blocks.push(BlockDirective {
                            location,
                            disable: kind == KindOfSuppression::Disable,
                            rules,
                        });
                    }
                    KindOfSuppression::Line | KindOfSuppression::NextLine => {
                        let line = location[0] + usize::from(kind == KindOfSuppression::NextLine);
                        let target = result.lines.entry(line).or_default();
                        if rules.is_empty() {
                            target.push("*".into());
                        } else {
                            target.extend(rules);
                        }
                    }
                }
                continue;
            }
            let content = content.trim_matches(is_js_space);
            if let Some(kind) = auxiliary_kind(comment.kind, content) {
                if kind == AuxiliaryKind::Globals {
                    // Keep offsets into the original source while normalizing
                    // whitespace around the directive's commas and colons.
                    let content_start = start + 2 + raw[2..].find(content).unwrap_or(0);
                    result.add_globals(content, content_start);
                }
                result.auxiliary.push(AuxiliaryDirective {
                    kind,
                    range: [start, end],
                    content: content.into(),
                    location,
                });
            }
        }
        Ok(result)
    }

    pub(crate) fn auxiliary(&self) -> &[AuxiliaryDirective] {
        &self.auxiliary
    }

    #[cfg(test)]
    fn globals(&self) -> &[GlobalDeclaration] {
        &self.globals
    }

    /// Install one file generation for the synchronous rule visit. Nested
    /// visits restore their outer metadata on normal return, error or unwind.
    pub(crate) fn enter_globals(&self, file: P<SourceFile>) -> GlobalsScope<'_> {
        let scope_id = NEXT_SCOPE_ID.with(|next| {
            let id = next
                .get()
                .checked_add(1)
                .expect("directive scope ID overflow");
            next.set(id);
            id
        });
        let cache = GlobalsCache {
            scope_id,
            file_identity: file_identity(file),
            access: self
                .globals
                .iter()
                .map(|entry| (entry.name.clone(), entry.access))
                .collect(),
            declarations: self.globals.clone(),
        };
        GLOBAL_SCOPES.with(|scopes| scopes.borrow_mut().push(cache));
        GlobalsScope {
            scope_id,
            _directives: PhantomData,
            _thread: PhantomData,
        }
    }

    fn add_globals(&mut self, content: &str, content_start: usize) {
        for (name, access, range) in parse_global_names(content, content_start) {
            if let Some(entry) = self.globals.iter_mut().find(|entry| entry.name == name) {
                entry.access = access;
                entry.name_ranges.push(range);
            } else {
                self.globals.push(GlobalDeclaration {
                    name,
                    access,
                    name_ranges: vec![range],
                });
            }
        }
    }

    /// Call until the host has implemented all auxiliary directive consumers.
    /// Suppression itself is supported; configuration must not be discarded.
    pub(crate) fn ensure_supported(&self) -> Result<(), String> {
        // The current host supports only the default file language goals:
        // module, or a JavaScript CommonJS wrapper for .cjs. Exported globals
        // never grant usage in either scope (frozen RuleContext contract).
        // Keep their parsed metadata for a future explicit script consumer.
        if let Some(directive) = self.auxiliary().iter().find(|directive| {
            !matches!(
                directive.kind,
                AuxiliaryKind::Exported | AuxiliaryKind::Globals
            )
        }) {
            return Err(format!(
                "unsupported {:?} lint directive at byte {}: {}",
                directive.kind, directive.range[0], directive.content
            ));
        }
        Ok(())
    }

    /// `position` must be the original report range's UTF-8 byte start. Its
    /// location is checked before any fix or suggestion is exposed/applied.
    pub(crate) fn is_rule_disabled(
        &self,
        source: &str,
        rule: &str,
        position: usize,
    ) -> Result<bool, String> {
        if position > source.len() || !source.is_char_boundary(position) {
            return Err(format!("invalid directive query byte position {position}"));
        }
        if self.blocks.is_empty() && self.lines.is_empty() {
            return Ok(false);
        }
        let [line, column] = self.location(source, position)?;
        Ok(self.is_rule_disabled_at(rule, line, column))
    }

    /// A convenience for already projected diagnostic positions. Both indices
    /// are zero-based and columns count UTF-16 code units, not Unicode scalars.
    pub(crate) fn is_rule_disabled_at(&self, rule: &str, line: usize, column: usize) -> bool {
        let mut all_disabled = false;
        let mut rule_disabled = None;
        for event in &self.blocks {
            if event.location > [line, column] {
                break;
            }
            if event.rules.is_empty() {
                all_disabled = event.disable;
                rule_disabled = None;
            } else if event.rules.iter().any(|name| name == rule) {
                rule_disabled = Some(event.disable);
            }
        }
        rule_disabled.unwrap_or(all_disabled)
            || self
                .lines
                .get(&line)
                .is_some_and(|rules| rules.iter().any(|name| name == rule || name == "*"))
    }

    fn location(&self, text: &str, position: usize) -> Result<[usize; 2], String> {
        if position > text.len() || !text.is_char_boundary(position) {
            return Err(format!("invalid directive query byte position {position}"));
        }
        let line = self.line_starts.partition_point(|&start| start <= position) - 1;
        let column = text[self.line_starts[line]..position]
            .encode_utf16()
            .count();
        Ok([line, column])
    }
}

fn match_suppression(content: &str) -> Option<(KindOfSuppression, Vec<String>)> {
    // Prefix matching (including misspelled suffixes) is pinned behavior:
    // eslint-disabled names a rule "d", and enable-line names rule "-line".
    for prefix in ["rslint", "eslint"] {
        let Some(rest) = content
            .strip_prefix(prefix)
            .and_then(|s| s.strip_prefix('-'))
        else {
            continue;
        };
        if let Some(rest) = rest.strip_prefix("disable") {
            let (kind, rest) = if let Some(rest) = rest.strip_prefix("-line") {
                (KindOfSuppression::Line, rest)
            } else if let Some(rest) = rest.strip_prefix("-next-line") {
                (KindOfSuppression::NextLine, rest)
            } else {
                (KindOfSuppression::Disable, rest)
            };
            return Some((kind, rule_names(rest)));
        }
        if let Some(rest) = rest.strip_prefix("enable") {
            return Some((KindOfSuppression::Enable, rule_names(rest)));
        }
    }
    None
}

fn rule_names(content: &str) -> Vec<String> {
    let content = content.split_once(" -- ").map_or(content, |(head, _)| head);
    content
        .split(',')
        .map(|name| name.trim_matches(is_baseline_space))
        .filter(|name| !name.is_empty())
        .map(str::to_owned)
        .collect()
}

fn auxiliary_kind(comment_kind: Kind, content: &str) -> Option<AuxiliaryKind> {
    if comment_kind == Kind::MultiLineCommentTrivia {
        for (label, kind) in [
            ("globals", AuxiliaryKind::Globals),
            ("global", AuxiliaryKind::Globals),
            ("exported", AuxiliaryKind::Exported),
            ("eslint", AuxiliaryKind::InlineConfig),
        ] {
            if content
                .strip_prefix(label)
                .is_some_and(|rest| rest.is_empty() || rest.chars().next().is_some_and(is_js_space))
            {
                return Some(kind);
            }
        }
    }
    (content.starts_with("eslint-") || content.starts_with("rslint-"))
        .then_some(AuxiliaryKind::Unknown)
}

fn parse_global_names(
    content: &str,
    content_start: usize,
) -> Vec<(String, GlobalAccess, [usize; 2])> {
    let label_len = if content.starts_with("globals") { 7 } else { 6 };
    let mut end = content.len();
    // ESLint splits the justification on /\s-{2,}\s/u, rather than the
    // suppression parser's deliberately narrower literal " -- ".
    for (offset, c) in content[label_len..].char_indices() {
        if !is_js_space(c) {
            continue;
        }
        let start = label_len + offset;
        let hyphens = start + c.len_utf8();
        let after = content[hyphens..]
            .bytes()
            .take_while(|&c| c == b'-')
            .count()
            + hyphens;
        if after - hyphens >= 2 && content[after..].chars().next().is_some_and(is_js_space) {
            end = start;
            break;
        }
    }
    // Whitespace immediately beside ':' or ',' disappears before tokenizing.
    // Original source offsets stay attached to every retained character.
    let raw: Vec<_> = content[label_len..end]
        .char_indices()
        .map(|(offset, c)| (c, label_len + offset, label_len + offset + c.len_utf8()))
        .collect();
    let mut normalized = Vec::with_capacity(raw.len());
    let mut index = 0;
    while index < raw.len() {
        if !is_js_space(raw[index].0) {
            normalized.push(raw[index]);
            index += 1;
            continue;
        }
        let start = index;
        while index < raw.len() && is_js_space(raw[index].0) {
            index += 1;
        }
        let before_delimiter = index < raw.len() && matches!(raw[index].0, ':' | ',');
        let after_delimiter = normalized
            .last()
            .is_some_and(|&(c, _, _)| matches!(c, ':' | ','));
        if !before_delimiter && !after_delimiter {
            normalized.extend_from_slice(&raw[start..index]);
        }
    }
    let mut entries: Vec<(String, Option<GlobalAccess>, [usize; 2])> = Vec::new();
    for token in normalized.split(|&(c, _, _)| c == ',' || is_js_space(c)) {
        if token.is_empty() {
            continue;
        }
        let name_end = token
            .iter()
            .position(|&(c, _, _)| c == ':')
            .unwrap_or(token.len());
        if name_end == 0 {
            continue;
        }
        let name_start = token[0].1;
        let name_finish = token[name_end - 1].2;
        let name = &content[name_start..name_finish];
        let access = if name_end == token.len() {
            Some(GlobalAccess::Readonly)
        } else {
            let setting = &token[name_end + 1..];
            let setting_end = setting
                .iter()
                .position(|&(c, _, _)| c == ':')
                .unwrap_or(setting.len());
            let value = if setting_end == 0 {
                ""
            } else {
                &content[setting[0].1..setting[setting_end - 1].2]
            };
            match value {
                "off" => Some(GlobalAccess::Off),
                "false" | "readonly" | "readable" => Some(GlobalAccess::Readonly),
                "true" | "writable" | "writeable" => Some(GlobalAccess::Writable),
                _ => None,
            }
        };
        // The last setting wins even when invalid, before valid entries are
        // merged across comments. Its range remains the first mention here.
        if let Some(entry) = entries.iter_mut().find(|entry| entry.0 == name) {
            entry.1 = access;
        } else {
            entries.push((
                name.into(),
                access,
                [content_start + name_start, content_start + name_finish],
            ));
        }
    }
    entries
        .into_iter()
        .filter_map(|(name, access, range)| access.map(|access| (name, access, range)))
        .collect()
}

fn is_baseline_space(c: char) -> bool {
    // unicode.IsSpace, as used by strings.TrimSpace in the pinned baseline.
    is_js_space(c) && c != '\u{feff}' || c == '\u{0085}'
}

fn is_js_space(c: char) -> bool {
    matches!(
        c,
        '\u{0009}'..='\u{000d}'
            | '\u{0020}'
            | '\u{00a0}'
            | '\u{1680}'
            | '\u{2000}'..='\u{200a}'
            | '\u{2028}'
            | '\u{2029}'
            | '\u{202f}'
            | '\u{205f}'
            | '\u{3000}'
            | '\u{feff}'
    )
}

fn line_starts(text: &str) -> Vec<usize> {
    let mut lines = vec![0];
    let mut chars = text.char_indices().peekable();
    while let Some((offset, character)) = chars.next() {
        if character == '\r' && chars.peek().is_some_and(|(_, c)| *c == '\n') {
            chars.next();
            lines.push(offset + 2);
        } else if matches!(character, '\r' | '\n' | '\u{2028}' | '\u{2029}') {
            lines.push(offset + character.len_utf8());
        }
    }
    lines
}

#[cfg(test)]
mod tests {
    use super::*;
    use tsrs_ast::SourceFileParseOptions;
    use tsrs_core::{ScriptKind, arena::Region, tspath::Path};

    fn parse(source: &str) -> Directives {
        let region = Region::new(1 << 20);
        let _scope = region.enter();
        let path = "/rslint-directive-tests/case.tsx";
        let file = tsrs_parser::parse_source_file(
            SourceFileParseOptions {
                file_name: path.into(),
                path: Path::new(path),
                ..Default::default()
            },
            source,
            ScriptKind::TSX,
        );
        Directives::parse(file).unwrap()
    }

    fn disabled(directives: &Directives, source: &str, rule: &str, marker: &str) -> bool {
        let pos = source.find(marker).expect("query marker must exist");
        directives.is_rule_disabled(source, rule, pos).unwrap()
    }

    #[test]
    fn same_line_boundaries_use_utf16_and_both_prefixes() {
        for prefix in ["rslint", "eslint"] {
            let source = format!(
                "/* 😀 */ before(); /* {prefix}-disable no-debugger */ inside(); /* {prefix}-enable no-debugger */ after();"
            );
            let directives = parse(&source);
            directives.ensure_supported().unwrap();
            assert!(!disabled(&directives, &source, "no-debugger", "before()"));
            assert!(disabled(&directives, &source, "no-debugger", "inside()"));
            assert!(!disabled(&directives, &source, "no-debugger", "after()"));
            assert!(!disabled(&directives, &source, "no-console", "inside()"));
            let pos = source.find("inside()").unwrap();
            let column = source[..pos].encode_utf16().count();
            assert!(directives.is_rule_disabled_at("no-debugger", 0, column));
            assert_eq!(directives.blocks.len(), 2, "deduplicate token trivia");
        }
    }

    #[test]
    fn wildcard_and_specific_transitions_reset_in_source_order() {
        let source = "/* eslint-disable */ one(); /* eslint-enable a */ two(); /* rslint-disable a */ three(); /* eslint-enable */ four(); /* eslint-disable b */ five();";
        let directives = parse(source);
        for (marker, a, b) in [
            ("one()", true, true),
            ("two()", false, true),
            ("three()", true, true),
            ("four()", false, false),
            ("five()", false, true),
        ] {
            assert_eq!(disabled(&directives, source, "a", marker), a, "{marker}");
            assert_eq!(disabled(&directives, source, "b", marker), b, "{marker}");
        }
    }

    #[test]
    fn line_controls_apply_independently_of_block_enables() {
        let source = "before(); // eslint-disable-line a\n// eslint-disable-next-line a, b -- reason\n/* eslint-enable */ inside();\nafter();\n// rslint-disable-next-line -- all rules\nwildcard();";
        let directives = parse(source);
        assert!(disabled(&directives, source, "a", "before()"));
        assert!(!disabled(&directives, source, "b", "before()"));
        for rule in ["a", "b"] {
            assert!(disabled(&directives, source, rule, "inside()"));
            assert!(!disabled(&directives, source, rule, "after()"));
            assert!(disabled(&directives, source, rule, "wildcard()"));
        }
    }

    #[test]
    fn every_ecmascript_line_terminator_and_block_start_line() {
        for separator in ["\n", "\r", "\r\n", "\u{2028}", "\u{2029}"] {
            let source =
                format!("// eslint-disable-next-line a{separator}inside();{separator}after();");
            let directives = parse(&source);
            assert!(disabled(&directives, &source, "a", "inside()"));
            assert!(!disabled(&directives, &source, "a", "after()"));
        }
        let source =
            "/* eslint-disable-next-line a\n -- starts at line zero */ inside();\nafter();";
        let directives = parse(source);
        assert!(disabled(&directives, source, "a", "inside()"));
        assert!(!disabled(&directives, source, "a", "after()"));
    }

    #[test]
    fn string_regex_template_and_jsx_text_are_not_comments() {
        for source in [
            "const s = '/* eslint-disable */'; after();",
            "const s = `// eslint-disable`; after();",
            "const r = /[/* eslint-disable */]/; after();",
            "const x = <div>/* eslint-disable */</div>; after();",
            "const x = <div>// eslint-disable</div>; after();",
            "const x = <div text='/* eslint-disable */'/>; after();",
        ] {
            let directives = parse(source);
            directives.ensure_supported().unwrap();
            assert!(!disabled(&directives, source, "a", "after()"), "{source}");
        }
        let source = "const x = <div>{/* eslint-disable */}{after()}</div>";
        assert!(disabled(&parse(source), source, "a", "after()"));
    }

    #[test]
    fn malformed_prefixes_retain_baseline_meaning() {
        for (source, rule, expected) in [
            (
                "/* eslint-disableno-debugger */ after();",
                "no-debugger",
                true,
            ),
            (
                "/* eslint-disabled no-debugger */ after();",
                "no-debugger",
                false,
            ),
            (
                "/* eslint-disable */ /* eslint-enable-line */ after();",
                "a",
                true,
            ),
            ("/* ESLINT-DISABLE */ after();", "a", false),
            ("/* tslint-disable */ after();", "a", false),
            ("/* eslint-disable * */ after();", "a", false),
            ("after(); // eslint-disable-line *", "a", true),
            ("/* eslint-disable , , */ after();", "a", true),
        ] {
            assert_eq!(
                disabled(&parse(source), source, rule, "after()"),
                expected,
                "{source}"
            );
        }
    }

    #[test]
    fn suppression_justification_is_the_pinned_literal_separator() {
        assert_eq!(rule_names(" a, b -- why -- more"), ["a", "b"]);
        assert!(rule_names(" -- all rules").is_empty());
        assert_eq!(rule_names(" a\t--\twhy"), ["a\t--\twhy"]);
        assert_eq!(rule_names(" a --- why"), ["a --- why"]);
        assert_eq!(rule_names(" a--why"), ["a--why"]);
        assert_eq!(rule_names(" a b"), ["a b"]);
    }

    #[test]
    fn whitespace_is_explicit_not_rust_unicode_defaults() {
        assert!(is_js_space('\u{feff}'));
        assert!(!is_baseline_space('\u{feff}'));
        assert!(!is_js_space('\u{0085}'));
        assert!(is_baseline_space('\u{0085}'));
        for space in ["\t", "\u{00a0}", "\u{2003}", "\u{0085}"] {
            let source = format!("/*{space}eslint-disable a */ after();");
            assert!(disabled(&parse(&source), &source, "a", "after()"));
        }
        let source = "/*\u{feff}eslint-disable a */ after();";
        let directives = parse(source);
        assert!(!disabled(&directives, source, "a", "after()"));
        assert_eq!(directives.auxiliary()[0].kind, AuxiliaryKind::Unknown);
    }

    #[test]
    fn auxiliary_directives_are_typed_and_never_silently_discarded() {
        for (source, kind) in [
            ("/* eslint no-debugger: 0 */", AuxiliaryKind::InlineConfig),
            ("/* eslint\tno-debugger: 0 */", AuxiliaryKind::InlineConfig),
            ("/* globals value:readonly */", AuxiliaryKind::Globals),
            ("/*\u{feff}global\u{2028}value */", AuxiliaryKind::Globals),
            ("/* exported value -- reason */", AuxiliaryKind::Exported),
            ("/* eslint-check */", AuxiliaryKind::Unknown),
            ("// rslint-unknown", AuxiliaryKind::Unknown),
        ] {
            let directives = parse(source);
            let auxiliary = directives.auxiliary();
            assert_eq!(auxiliary.len(), 1, "{source}");
            assert_eq!(auxiliary[0].kind, kind, "{source}");
            assert_eq!(auxiliary[0].range, [0, source.len()]);
            assert_eq!(auxiliary[0].location, [0, 0]);
            assert_eq!(
                directives.ensure_supported().is_err(),
                !matches!(kind, AuxiliaryKind::Exported | AuxiliaryKind::Globals),
                "{source}"
            );
        }
        for source in [
            "// global value",
            "// exported value",
            "// eslint a:0",
            "/* globally value */",
        ] {
            parse(source).ensure_supported().unwrap();
        }
    }

    #[test]
    fn globals_preserve_aliases_overrides_exact_ranges_and_justifications() {
        let source = "/* globals bare, ro : readonly, alias:readable, rw:writable, rw2:writeable, yes:true, no:false, gone:off, repeated, repeated:writable -- omitted */\n/* global repeated:off, gone, ro:bogus */";
        let directives = parse(source);
        directives.ensure_supported().unwrap();
        let global = |name: &str| {
            directives
                .globals()
                .iter()
                .find(|entry| entry.name == name)
                .unwrap()
        };
        for name in ["bare", "ro", "alias", "no", "gone"] {
            assert_eq!(global(name).access, GlobalAccess::Readonly, "{name}");
        }
        for name in ["rw", "rw2", "yes"] {
            assert_eq!(global(name).access, GlobalAccess::Writable, "{name}");
        }
        assert_eq!(global("repeated").access, GlobalAccess::Off);
        assert_eq!(global("repeated").name_ranges.len(), 2);
        assert_eq!(
            global("ro").name_ranges.len(),
            1,
            "invalid later setting is ignored"
        );
        assert_eq!(global("gone").name_ranges.len(), 2);
        for entry in directives.globals() {
            for &[start, end] in &entry.name_ranges {
                assert_eq!(&source[start..end], entry.name);
            }
        }
        assert!(
            !directives
                .globals()
                .iter()
                .any(|entry| entry.name == "omitted")
        );
        // The final setting within a single comment is selected before it is
        // validated, so a later invalid spelling discards that comment entry.
        let directives = parse(
            "/* global a:off */ /* global a:readonly, a:bogus, empty:, nonsense:other, valid:true */",
        );
        assert_eq!(directives.globals().len(), 2);
        assert_eq!(directives.globals()[0].access, GlobalAccess::Off);
        assert_eq!(directives.globals()[0].name_ranges.len(), 1);
        assert_eq!(directives.globals()[1].name, "valid");
    }

    #[test]
    fn global_settings_use_real_block_comments_and_javascript_whitespace() {
        for source in [
            "const s = '/* global Object:off */';",
            "const s = `/* global Object:off */`;",
            "const r = /[/* global Object:off */]/;",
            "const x = <div>/* global Object:off */</div>;",
            "const x = <div label='/* global Object:off */'/>;",
            "// global Object:off",
            "/* Global Object:off */",
            "/* globally Object:off */",
            "/*\u{0085}global Object:off */",
        ] {
            assert!(parse(source).globals().is_empty(), "{source}");
        }
        let source = "#!/usr/bin/env node\n/*\u{feff}globals\u{2028}😀\u{00a0}:\u{2003}readonly\t,\nObject : off\u{2029}---\tignored */";
        let directives = parse(source);
        assert_eq!(directives.globals().len(), 2);
        assert_eq!(directives.globals()[0].name, "😀");
        assert_eq!(directives.globals()[1].access, GlobalAccess::Off);
        for entry in directives.globals() {
            let [start, end] = entry.name_ranges[0];
            assert_eq!(&source[start..end], entry.name);
        }
        let source = "const x = <div>{/* global Object:off */}</div>;";
        assert_eq!(parse(source).globals()[0].name, "Object");
    }

    #[test]
    fn global_setting_distinguishes_disabled_declared_and_absent() {
        let region = Region::new(1 << 20);
        let _scope = region.enter();
        let path = "/rslint-directive-tests/globals.ts";
        let file = tsrs_parser::parse_source_file(
            SourceFileParseOptions {
                file_name: path.into(),
                path: Path::new(path),
                ..Default::default()
            },
            "/* global Object:off, custom:readonly, writable:true */",
            ScriptKind::TS,
        );
        assert_eq!(global_setting(file, "Object"), Some(false));
        assert_eq!(global_setting(file, "custom"), Some(true));
        assert_eq!(global_setting(file, "writable"), Some(true));
        assert_eq!(global_setting(file, "absent"), None);
    }

    fn cache_test_file(source: &str) -> P<SourceFile> {
        let path = "/rslint-directive-tests/same-path.ts";
        tsrs_parser::parse_source_file(
            SourceFileParseOptions {
                file_name: path.into(),
                path: Path::new(path),
                ..Default::default()
            },
            source,
            ScriptKind::TS,
        )
    }

    fn assert_no_cached_globals() {
        GLOBAL_SCOPES.with(|scopes| assert!(scopes.borrow().is_empty()));
    }

    #[test]
    fn repeated_globals_queries_use_owned_file_scope_without_reparsing() {
        assert_no_cached_globals();
        let region = Region::new(1 << 20);
        let _region = region.enter();
        let file = cache_test_file("/* globals Object:off, custom:writable */");
        let directives = Directives::parse(file).unwrap();
        let before = DIRECTIVE_PARSE_COUNT.with(Cell::get);
        {
            let _globals = directives.enter_globals(file);
            for _ in 0..100 {
                assert_eq!(global_setting(file, "Object"), Some(false));
                assert_eq!(global_access(file, "custom"), Some(GlobalAccess::Writable));
                assert_eq!(global_setting(file, "absent"), None);
                assert_eq!(global_declarations(file).unwrap(), directives.globals);
            }
            assert_eq!(DIRECTIVE_PARSE_COUNT.with(Cell::get), before);
        }
        assert_no_cached_globals();
        // Helpers outside a normal lint visit still parse current metadata.
        assert_eq!(global_setting(file, "Object"), Some(false));
        assert_eq!(DIRECTIVE_PARSE_COUNT.with(Cell::get), before + 1);
    }

    #[test]
    fn nested_global_scopes_restore_outer_and_errors_release_entries() {
        assert_no_cached_globals();
        let region = Region::new(1 << 20);
        let _region = region.enter();
        let outer = cache_test_file("/* global Object:off */");
        let inner = cache_test_file("/* global Object:readonly */");
        let outer_directives = Directives::parse(outer).unwrap();
        let inner_directives = Directives::parse(inner).unwrap();
        let before = DIRECTIVE_PARSE_COUNT.with(Cell::get);
        {
            let _outer = outer_directives.enter_globals(outer);
            {
                let _inner = inner_directives.enter_globals(inner);
                assert_eq!(global_setting(outer, "Object"), Some(false));
                assert_eq!(global_setting(inner, "Object"), Some(true));
                GLOBAL_SCOPES.with(|scopes| assert_eq!(scopes.borrow().len(), 2));
            }
            GLOBAL_SCOPES.with(|scopes| assert_eq!(scopes.borrow().len(), 1));
            assert_eq!(global_setting(outer, "Object"), Some(false));
            assert_eq!(DIRECTIVE_PARSE_COUNT.with(Cell::get), before);
        }
        assert_no_cached_globals();
        fn failing_visit(file: P<SourceFile>) -> Result<(), &'static str> {
            let directives = Directives::parse(file).unwrap();
            let _globals = directives.enter_globals(file);
            Err("stop file visit")
        }
        assert_eq!(failing_visit(outer), Err("stop file visit"));
        assert_no_cached_globals();
    }

    #[test]
    fn same_path_new_arena_generations_do_not_reuse_global_metadata() {
        assert_no_cached_globals();
        for (source, expected) in [
            ("/* global Object:off */", Some(false)),
            ("/* global Object:writable */", Some(true)),
            ("const noDirective = true;", None),
        ] {
            let region = Region::new(1 << 20);
            let _region = region.enter();
            let file = cache_test_file(source);
            let directives = Directives::parse(file).unwrap();
            {
                let _globals = directives.enter_globals(file);
                assert_eq!(global_setting(file, "Object"), expected);
            }
            assert_no_cached_globals();
        }
        assert_no_cached_globals();
    }

    #[test]
    fn unwinding_global_scope_releases_all_owned_metadata() {
        assert_no_cached_globals();
        let region = Region::new(1 << 20);
        let _region = region.enter();
        let file = cache_test_file("/* global Object:off */");
        let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            let directives = Directives::parse(file).unwrap();
            let _globals = directives.enter_globals(file);
            assert_eq!(global_setting(file, "Object"), Some(false));
            panic!("stop nested lint visit");
        }));
        assert!(result.is_err());
        assert_no_cached_globals();
        assert_eq!(global_setting(file, "Object"), Some(false));
    }

    #[test]
    fn comments_only_and_shebang_files_are_supported() {
        let source = "#!/usr/bin/env node\n// rslint-disable a\n";
        let directives = parse(source);
        assert!(
            directives
                .is_rule_disabled(source, "a", source.len())
                .unwrap()
        );
        assert_eq!(directives.blocks.len(), 1);
        assert_eq!(parse("/* eslint-disable */").blocks.len(), 1);
        assert!(parse("").blocks.is_empty());
    }

    #[test]
    fn invalid_byte_queries_are_rejected() {
        let source = "😀";
        let directives = parse(source);
        assert!(directives.is_rule_disabled(source, "a", 1).is_err());
        assert!(
            directives
                .is_rule_disabled(source, "a", source.len() + 1)
                .is_err()
        );
    }
}
