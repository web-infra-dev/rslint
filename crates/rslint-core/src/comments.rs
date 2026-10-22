//! Discover comments through parsed token owners, never by scanning raw source
//! for comment markers. Adapted from `internal/utils/ts_api_utils.go`, originally
//! ported from ts-api-utils `comments.ts` at
//! 491c0374725a5dd64632405efea101f20ed5451f (MIT).
//!
//! Comment consumers interpret directives after this lexical traversal.

use tsrs_ast::{self as ast, CommentRange, Kind, Node, NodeFlags, SourceFile};
use tsrs_core::{LanguageVariant, P};
use tsrs_scanner::{self as scanner, Scanner};

pub(crate) fn for_each_comment(
    file: P<SourceFile>,
    mut callback: impl FnMut(CommentRange) -> Result<(), String>,
) -> Result<(), String> {
    let text = file.text();
    let not_jsx = file.language_variant() != LanguageVariant::JSX;
    let mut saw_token = false;
    for_each_token_range(file, |kind, pos, end, parent| {
        if pos == end {
            return Ok(());
        }
        saw_token = true;
        if kind != Kind::JsxText {
            let comment_pos = if pos == 0 {
                scanner::get_shebang(text).len() as u32
            } else {
                pos
            };
            for comment in scanner::get_leading_comment_ranges(text, comment_pos) {
                callback(comment)?;
            }
        }
        if not_jsx || can_have_trailing_trivia(kind, end, parent) {
            for comment in scanner::get_trailing_comment_ranges(text, end) {
                callback(comment)?;
            }
        }
        Ok(())
    })?;
    if !saw_token {
        for comment in
            scanner::get_leading_comment_ranges(text, scanner::get_shebang(text).len() as u32)
        {
            callback(comment)?;
        }
    }
    Ok(())
}

fn is_jsx_element_or_fragment(node: Option<P<Node>>) -> bool {
    node.is_some_and(|node| matches!(node.kind(), Kind::JsxElement | Kind::JsxFragment))
}

fn can_have_trailing_trivia(kind: Kind, end: u32, parent: Option<P<Node>>) -> bool {
    let Some(parent) = parent else {
        return true;
    };
    match kind {
        Kind::CloseBraceToken => {
            parent.kind() != Kind::JsxExpression || !is_jsx_element_or_fragment(parent.parent())
        }
        Kind::GreaterThanToken => match parent.kind() {
            // Both closing forms are excluded inside a containing JSX body.
            // The old Go switch accidentally omitted the closing-element arm.
            Kind::JsxClosingElement | Kind::JsxClosingFragment => {
                !is_jsx_element_or_fragment(parent.parent().and_then(|node| node.parent()))
            }
            Kind::JsxOpeningElement => parent.end().source_offset() != Some(end),
            Kind::JsxOpeningFragment => false,
            Kind::JsxSelfClosingElement => {
                parent.end().source_offset() != Some(end)
                    || !is_jsx_element_or_fragment(parent.parent())
            }
            _ => true,
        },
        _ => true,
    }
}

enum Item {
    Node(P<Node>),
    Token {
        kind: Kind,
        pos: u32,
        end: u32,
        parent: P<Node>,
    },
}

fn scan_gap(
    scanner: &mut Scanner,
    file: P<SourceFile>,
    parent: P<Node>,
    pos: &mut u32,
    end: u32,
    queue: &mut Vec<Item>,
) -> Result<(), String> {
    if *pos >= end {
        return Ok(());
    }
    // Reset every gap: lexical state from a template/regex/JSX region must not
    // leak into the next gap. AST token nodes own those regions in their entirety.
    scanner.reset();
    scanner.set_text(file.text());
    scanner.set_language_variant(file.language_variant());
    scanner.reset_pos(*pos);
    scanner.scan();
    while *pos < end {
        let next = scanner.token_end();
        if next <= *pos {
            return Err(format!(
                "comment traversal did not advance in {} at byte {}",
                file.file_name(),
                pos
            ));
        }
        queue.push(Item::Token {
            kind: scanner.token(),
            pos: scanner.token_full_start(),
            end: next,
            parent,
        });
        *pos = next;
        scanner.scan();
    }
    Ok(())
}

pub(crate) fn for_each_token_range(
    file: P<SourceFile>,
    mut callback: impl FnMut(Kind, u32, u32, Option<P<Node>>) -> Result<(), String>,
) -> Result<(), String> {
    let mut queue = vec![Item::Node(file.as_node())];
    let mut scanner = Scanner::new();
    while let Some(item) = queue.pop() {
        let node = match item {
            Item::Token {
                kind,
                pos,
                end,
                parent,
            } => {
                callback(kind, pos, end, Some(parent))?;
                continue;
            }
            Item::Node(node) => node,
        };
        let mut pos = source_position(node.pos().source_offset(), file)?;
        let end = source_position(node.end().source_offset(), file)?;
        if ast::is_token_kind(node.kind()) {
            callback(node.kind(), pos, end, node.parent())?;
            continue;
        }
        let start = queue.len();
        let mut error = None;
        node.for_each_child(&mut |child| {
            if child.flags().intersects(NodeFlags::Reparsed) {
                return false;
            }
            let result = (|| {
                let child_pos = source_position(child.pos().source_offset(), file)?;
                let child_end = source_position(child.end().source_offset(), file)?;
                scan_gap(&mut scanner, file, node, &mut pos, child_pos, &mut queue)?;
                queue.push(Item::Node(child));
                pos = child_end;
                Ok::<_, String>(())
            })();
            if let Err(message) = result {
                error = Some(message);
                return true;
            }
            false
        });
        if let Some(error) = error {
            return Err(error);
        }
        scan_gap(&mut scanner, file, node, &mut pos, end, &mut queue)?;
        queue[start..].reverse();
    }
    Ok(())
}

fn source_position(position: Option<u32>, file: P<SourceFile>) -> Result<u32, String> {
    position.ok_or_else(|| {
        format!(
            "comment traversal encountered a synthetic node in {}",
            file.file_name()
        )
    })
}
