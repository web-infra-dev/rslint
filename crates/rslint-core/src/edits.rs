//! Project one diagnostic's atomic edits into the public JavaScript fix shape.
//!
//! Ordinary edit ranges index the original tsrs source in UTF-8 bytes, including
//! a leading BOM when present. They must address the text after that BOM; removing
//! the mark is the explicit `RemoveBom` operation. Public fix ranges instead use
//! BOM-free UTF-16 offsets, with `[-1, 0]` reserved for removing the mark.
//! This module constructs message payloads. It does not apply automatic fixes.

use serde::Serialize;

#[derive(Clone, Debug, PartialEq, Eq)]
pub(crate) enum Edit {
    Replace {
        range: [usize; 2],
        text: String,
    },
    #[cfg_attr(
        not(test),
        expect(
            dead_code,
            reason = "The BOM rule is not yet migrated; its edit contract is tested here."
        )
    )]
    RemoveBom,
}

impl Edit {
    pub(crate) fn replace(range: [usize; 2], text: impl Into<String>) -> Self {
        Self::Replace {
            range,
            text: text.into(),
        }
    }

    pub(crate) fn remove(range: [usize; 2]) -> Self {
        Self::replace(range, String::new())
    }
}

#[derive(Debug, PartialEq, Eq, Serialize)]
pub(crate) struct Fix {
    pub(crate) range: [isize; 2],
    pub(crate) text: String,
}

struct SourceEdit<'a> {
    // BOM-free byte positions; only RemoveBom may use the -1 sentinel.
    start: isize,
    end: isize,
    text: &'a str,
}

/// Merge an atomic edit group, preserving untouched source between its edits.
///
/// Unlike the legacy serializer's defensive overlap skip, an invalid native
/// edit group is an error: returning only part of a rule's edit would change its
/// meaning. Equal-position insertions retain their authored order. Adjacent
/// edits are valid within this group; conflict selection between diagnostics
/// belongs to the eventual automatic fixer.
pub(crate) fn merge_edits(source: &str, edits: &[Edit]) -> Result<Option<Fix>, String> {
    if edits.is_empty() {
        return Ok(None);
    }
    // tsrs can represent JavaScript string values using WTF-8 internally. The
    // current public JSON boundary accepts UTF-8 source/replacements only.
    std::str::from_utf8(source.as_bytes())
        .map_err(|_| "fix source is not valid UTF-8".to_owned())?;
    let bom_len = if source.starts_with('\u{feff}') { 3 } else { 0 };
    let text = &source[bom_len..];
    let mut ordered = Vec::with_capacity(edits.len());
    for edit in edits {
        let edit = match edit {
            Edit::RemoveBom => {
                if bom_len == 0 {
                    return Err("cannot remove a BOM from source without one".into());
                }
                SourceEdit {
                    start: -1,
                    end: 0,
                    text: "",
                }
            }
            Edit::Replace {
                range: [start, end],
                text: replacement,
            } => {
                if start > end || *end > source.len() {
                    return Err(format!("invalid fix byte range [{start}, {end}]"));
                }
                if !source.is_char_boundary(*start) || !source.is_char_boundary(*end) {
                    return Err(format!(
                        "fix byte range [{start}, {end}] splits a UTF-8 character"
                    ));
                }
                if *start < bom_len {
                    return Err(
                        "ordinary fix ranges must start after the BOM; use RemoveBom to remove it"
                            .into(),
                    );
                }
                std::str::from_utf8(replacement.as_bytes())
                    .map_err(|_| "fix replacement is not valid UTF-8".to_owned())?;
                SourceEdit {
                    start: byte_position(*start - bom_len)?,
                    end: byte_position(*end - bom_len)?,
                    text: replacement,
                }
            }
        };
        ordered.push(edit);
    }
    // Stable sorting matters for multiple inserts at exactly the same point.
    ordered.sort_by_key(|edit| (edit.start, edit.end));
    for pair in ordered.windows(2) {
        if pair[1].start < pair[0].end {
            return Err("overlapping fixes in one atomic edit group".into());
        }
    }

    let start = ordered[0].start;
    let end = ordered[ordered.len() - 1].end;
    let mut merged = String::new();
    let mut previous_end = start;
    for edit in ordered {
        // There is no source text preceding position zero for RemoveBom.
        merged.push_str(&text[previous_end.max(0) as usize..edit.start.max(0) as usize]);
        merged.push_str(edit.text);
        previous_end = edit.end;
    }
    Ok(Some(Fix {
        range: [utf16_position(text, start)?, utf16_position(text, end)?],
        text: merged,
    }))
}

fn byte_position(offset: usize) -> Result<isize, String> {
    isize::try_from(offset).map_err(|_| "fix byte offset exceeds the supported range".into())
}

fn utf16_position(text: &str, offset: isize) -> Result<isize, String> {
    if offset < 0 {
        return Ok(offset);
    }
    isize::try_from(text[..offset as usize].encode_utf16().count())
        .map_err(|_| "fix UTF-16 offset exceeds the supported range".into())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn merged(source: &str, edits: &[Edit]) -> Fix {
        merge_edits(source, edits).unwrap().unwrap()
    }

    #[test]
    fn empty_group_has_no_fix_payload() {
        assert_eq!(merge_edits("source", &[]).unwrap(), None);
        assert_eq!(merge_edits("", &[]).unwrap(), None);
    }

    #[test]
    fn single_edits_use_flat_utf16_offsets_across_emoji_and_crlf() {
        let fix = merged("😀\r\nalpha", &[Edit::replace([6, 11], "β")]);
        assert_eq!(
            serde_json::to_value(fix).unwrap(),
            serde_json::json!({"range": [4, 9], "text": "β"})
        );
        assert_eq!(
            merged("😀", &[Edit::remove([0, 4])]),
            Fix {
                range: [0, 2],
                text: String::new()
            }
        );
    }

    #[test]
    fn multiple_edits_sort_and_preserve_exact_source_gaps() {
        let source = "😀 first\r\n + second";
        let first = source.find("first").unwrap();
        let second = source.find("second").unwrap();
        assert_eq!(
            merged(
                source,
                &[
                    Edit::replace([second, second + 6], "RIGHT"),
                    Edit::replace([first, first + 5], "LEFT"),
                ]
            ),
            Fix {
                range: [3, source.encode_utf16().count() as isize],
                text: "LEFT\r\n + RIGHT".into(),
            }
        );
    }

    #[test]
    fn equal_point_insertions_keep_input_order() {
        assert_eq!(
            merged(
                "ab",
                &[
                    Edit::replace([1, 1], "second"),
                    Edit::replace([1, 1], "first")
                ]
            ),
            Fix {
                range: [1, 1],
                text: "secondfirst".into(),
            }
        );
    }

    #[test]
    fn adjacent_edits_inside_one_group_remain_atomic() {
        assert_eq!(
            merged(
                "abc",
                &[
                    Edit::replace([1, 2], "B"),
                    Edit::replace([2, 2], "!"),
                    Edit::replace([1, 1], "<"),
                    Edit::replace([2, 3], "C"),
                ]
            ),
            Fix {
                range: [1, 3],
                text: "<B!C".into(),
            }
        );
    }

    #[test]
    fn bom_coordinates_and_removal_are_explicit() {
        let source = "\u{feff}😀\r\nvalue";
        assert_eq!(
            merged(source, &[Edit::RemoveBom]),
            Fix {
                range: [-1, 0],
                text: String::new(),
            }
        );
        assert_eq!(
            merged(source, &[Edit::replace([3, 3], "prefix")]),
            Fix {
                range: [0, 0],
                text: "prefix".into(),
            }
        );
        assert_eq!(
            merged(source, &[Edit::replace([9, 14], "name")]),
            Fix {
                range: [4, 9],
                text: "name".into(),
            }
        );
        assert_eq!(
            merged(source, &[Edit::replace([9, 14], "name"), Edit::RemoveBom]),
            Fix {
                range: [-1, 9],
                text: "😀\r\nname".into(),
            }
        );
        assert_eq!(
            merged("\u{feff}", &[Edit::replace([3, 3], "x"), Edit::RemoveBom]),
            Fix {
                range: [-1, 0],
                text: "x".into(),
            }
        );
    }

    #[test]
    fn malformed_ranges_and_overlapping_atomic_groups_fail() {
        for edits in [
            vec![Edit::remove([2, 1])],
            vec![Edit::remove([0, 4])],
            vec![Edit::remove([0, 2]), Edit::remove([1, 3])],
            vec![Edit::remove([0, 3]), Edit::replace([1, 1], "x")],
            vec![Edit::RemoveBom],
        ] {
            assert!(merge_edits("abc", &edits).is_err(), "{edits:?}");
        }
        for range in [[0, 1], [1, 4], [2, 2]] {
            assert!(merge_edits("😀", &[Edit::remove(range)]).is_err());
        }
        for range in [[0, 0], [0, 3], [0, 4], [1, 3], [2, 3]] {
            assert!(merge_edits("\u{feff}x", &[Edit::remove(range)]).is_err());
        }
        assert!(merge_edits("\u{feff}x", &[Edit::RemoveBom, Edit::RemoveBom]).is_err());
    }

    #[test]
    fn empty_source_allows_insertions_and_empty_replacements() {
        assert_eq!(
            merged("", &[Edit::replace([0, 0], "😀")]),
            Fix {
                range: [0, 0],
                text: "😀".into(),
            }
        );
        assert_eq!(
            merged("", &[Edit::remove([0, 0])]),
            Fix {
                range: [0, 0],
                text: String::new(),
            }
        );
    }
}
