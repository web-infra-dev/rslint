//! Project one diagnostic's atomic edits into the public JavaScript fix shape.
//!
//! Ordinary edit ranges index the original tsrs source in UTF-8 bytes, including
//! a leading BOM when present. They must address the text after that BOM; removing
//! the mark is the explicit `RemoveBom` operation. Public fix ranges instead use
//! BOM-free UTF-16 offsets, with `[-1, 0]` reserved for removing the mark.
//! Each diagnostic supplies an atomic edit group. Automatic application selects
//! non-conflicting groups without applying suggestion-only edits.

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

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
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
/// edits are valid within this group; `apply_edits` selects conflicts between
/// diagnostic groups.
pub(crate) fn merge_edits(source: &str, edits: &[Edit]) -> Result<Option<Fix>, String> {
    if edits.is_empty() {
        return Ok(None);
    }
    let bom_len = leading_bom_len(source);
    let text = &source[bom_len..];
    let ordered = ordered_edits(source, edits, false)?;

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

/// The original diagnostic indices accepted or deferred by a single fix pass.
///
/// Applying an edit with identical text still counts as an application. Callers
/// must not infer `did_apply` by comparing the input and output strings.
#[derive(Debug, PartialEq, Eq)]
pub(crate) struct AppliedEdits {
    pub(crate) output: String,
    pub(crate) applied: Vec<usize>,
    pub(crate) unapplied: Vec<usize>,
}

impl AppliedEdits {
    pub(crate) fn did_apply(&self) -> bool {
        !self.applied.is_empty()
    }
}

/// Apply automatic fix groups using the baseline source-code fixer's selection.
///
/// Groups are ordered by their first edit's start and final edit's end. A group
/// occupies that entire span, including untouched gaps between its edits. A
/// conflict defers the entire group; adjacent replacements are accepted, while
/// an insertion at the boundary of another group conflicts. Edits within one
/// group may be adjacent and equal-position insertions retain authored order.
///
/// Pass only each diagnostic's automatic edits here, never its suggestions.
/// Empty groups are returned as unapplied in input order, followed by conflicting
/// groups in selection order. Invalid native ranges and overlapping edits within
/// a group fail the pass before any result is returned.
pub(crate) fn apply_edits(source: &str, groups: &[Vec<Edit>]) -> Result<AppliedEdits, String> {
    let mut result = AppliedEdits {
        output: String::with_capacity(source.len()),
        applied: Vec::new(),
        unapplied: Vec::new(),
    };
    let mut ordered = Vec::with_capacity(groups.len());
    for (index, group) in groups.iter().enumerate() {
        if group.is_empty() {
            result.unapplied.push(index);
        } else {
            // The baseline considers removing an absent BOM an applied no-op.
            // Serialization remains strict so a rule cannot emit that payload.
            ordered.push((index, ordered_edits(source, group, true)?));
        }
    }
    ordered.sort_by_key(|(_, group)| (group[0].start, group[group.len() - 1].end));

    let bom_len = leading_bom_len(source);
    let text = &source[bom_len..];
    let mut preserve_bom = bom_len != 0;
    // Below every normal edit and equal to the explicit BOM sentinel. Unlike a
    // position-zero insertion, the first RemoveBom must never self-conflict.
    let mut last_end = -1;
    let mut last_was_insertion = false;
    for (index, group) in ordered {
        let first = &group[0];
        let starts_with_insertion = first.start == first.end;
        let overlaps = last_end > first.start;
        let adjacent_conflict = result.did_apply()
            && last_end == first.start
            && (starts_with_insertion || last_was_insertion);
        if overlaps || adjacent_conflict {
            result.unapplied.push(index);
            continue;
        }
        for edit in group {
            if edit.start < 0 || (edit.start == 0 && edit.text.starts_with('\u{feff}')) {
                preserve_bom = false;
            }
            result
                .output
                .push_str(&text[last_end.max(0) as usize..edit.start.max(0) as usize]);
            result.output.push_str(edit.text);
            last_end = edit.end;
            last_was_insertion = edit.start == edit.end;
        }
        result.applied.push(index);
    }
    result.output.push_str(&text[last_end.max(0) as usize..]);
    if preserve_bom {
        result.output.insert(0, '\u{feff}');
    }
    Ok(result)
}

fn leading_bom_len(source: &str) -> usize {
    if source.starts_with('\u{feff}') { 3 } else { 0 }
}

fn ordered_edits<'a>(
    source: &str,
    edits: &'a [Edit],
    allow_absent_bom: bool,
) -> Result<Vec<SourceEdit<'a>>, String> {
    // tsrs can represent JavaScript string values using WTF-8 internally. The
    // current public JSON boundary accepts UTF-8 source/replacements only.
    std::str::from_utf8(source.as_bytes())
        .map_err(|_| "fix source is not valid UTF-8".to_owned())?;
    let bom_len = leading_bom_len(source);
    let mut ordered = Vec::with_capacity(edits.len());
    for edit in edits {
        let edit = match edit {
            Edit::RemoveBom => {
                if bom_len == 0 && !allow_absent_bom {
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

    Ok(ordered)
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

    #[test]
    fn apply_orders_groups_and_edits_without_mutating_them() {
        let groups = vec![
            vec![Edit::replace([8, 11], "CCC")],
            vec![Edit::replace([4, 7], "BBB"), Edit::replace([0, 3], "AAA")],
        ];
        let original_groups = groups.clone();
        let result = apply_edits("aaa bbb ccc", &groups).unwrap();
        assert_eq!(result.output, "AAA BBB CCC");
        assert_eq!(result.applied, [1, 0]);
        assert!(result.unapplied.is_empty());
        assert!(result.did_apply());
        assert_eq!(groups, original_groups);
    }

    #[test]
    fn automatic_groups_reserve_untouched_gaps_and_defer_atomically() {
        let result = apply_edits(
            "function foo() { return bar(); }",
            &[
                vec![
                    Edit::replace([0, 0], "async "),
                    Edit::replace([24, 24], "await "),
                ],
                vec![
                    Edit::replace([10, 10], "CONFLICT"),
                    Edit::replace([32, 32], "!"),
                ],
                vec![Edit::replace([32, 32], " trailing")],
            ],
        )
        .unwrap();
        assert_eq!(
            result.output,
            "async function foo() { return await bar(); } trailing"
        );
        assert_eq!(result.applied, [0, 2]);
        assert_eq!(result.unapplied, [1]);
    }

    #[test]
    fn adjacent_replacements_apply_but_boundary_insertions_conflict() {
        for (groups, expected, applied, unapplied) in [
            (
                vec![
                    vec![Edit::replace([0, 3], "XXX")],
                    vec![Edit::replace([3, 6], "YYY")],
                ],
                "XXXYYY",
                vec![0, 1],
                vec![],
            ),
            (
                vec![
                    vec![Edit::replace([0, 3], "XXX")],
                    vec![Edit::replace([3, 3], "YYY")],
                ],
                "XXXdef",
                vec![0],
                vec![1],
            ),
            (
                vec![
                    vec![Edit::replace([0, 0], "XXX")],
                    vec![Edit::replace([0, 3], "YYY")],
                ],
                "XXXabcdef",
                vec![0],
                vec![1],
            ),
            (
                vec![
                    vec![Edit::replace([0, 3], "XXX")],
                    vec![Edit::replace([0, 0], "YYY")],
                ],
                "YYYabcdef",
                vec![1],
                vec![0],
            ),
            (
                vec![
                    vec![Edit::replace([3, 3], "XXX")],
                    vec![Edit::replace([3, 3], "YYY")],
                ],
                "abcXXXdef",
                vec![0],
                vec![1],
            ),
        ] {
            let result = apply_edits("abcdef", &groups).unwrap();
            assert_eq!(result.output, expected, "{groups:?}");
            assert_eq!(result.applied, applied, "{groups:?}");
            assert_eq!(result.unapplied, unapplied, "{groups:?}");
        }
    }

    #[test]
    fn atomic_group_keeps_adjacent_and_equal_position_edits() {
        let result = apply_edits(
            "abc",
            &[vec![
                Edit::replace([2, 2], "!"),
                Edit::replace([1, 2], "B"),
                Edit::replace([1, 1], "<"),
                Edit::replace([2, 3], "C"),
                Edit::replace([2, 2], "?"),
            ]],
        )
        .unwrap();
        assert_eq!(result.output, "a<B!?C");
        assert_eq!(result.applied, [0]);
        assert!(result.unapplied.is_empty());
    }

    #[test]
    fn group_priority_uses_final_end_then_stable_input_order() {
        let result = apply_edits(
            "abcdef",
            &[
                vec![
                    Edit::replace([0, 1], "FIRST"),
                    Edit::replace([4, 6], "LAST"),
                ],
                vec![Edit::replace([0, 2], "short")],
                vec![Edit::replace([0, 2], "equal")],
            ],
        )
        .unwrap();
        assert_eq!(result.output, "shortcdef");
        assert_eq!(result.applied, [1]);
        assert_eq!(result.unapplied, [2, 0]);
    }

    #[test]
    fn unapplied_indices_put_empty_diagnostics_before_conflicting_groups() {
        let result = apply_edits(
            "abc",
            &[
                vec![Edit::replace([1, 1], "first")],
                vec![],
                vec![Edit::replace([1, 1], "second")],
                vec![],
                vec![Edit::replace([1, 2], "third")],
            ],
        )
        .unwrap();
        assert_eq!(result.output, "afirstbc");
        assert_eq!(result.applied, [0]);
        assert_eq!(result.unapplied, [1, 3, 2, 4]);
    }

    #[test]
    fn no_op_fixes_are_applications_even_when_output_is_empty_or_identical() {
        for (source, groups) in [
            ("", vec![vec![Edit::remove([0, 0])]]),
            ("same", vec![vec![Edit::replace([0, 4], "same")]]),
            ("same", vec![vec![Edit::RemoveBom]]),
            (
                "\u{feff}same",
                vec![vec![Edit::replace([3, 3], "\u{feff}")]],
            ),
        ] {
            let result = apply_edits(source, &groups).unwrap();
            assert_eq!(result.output, source);
            assert_eq!(result.applied, [0]);
            assert!(result.did_apply());
        }
        for source in ["", "same", "\u{feff}"] {
            let result = apply_edits(source, &[vec![]]).unwrap();
            assert_eq!(result.output, source);
            assert!(!result.did_apply());
            assert_eq!(result.unapplied, [0]);
            assert!(!apply_edits(source, &[]).unwrap().did_apply());
        }
    }

    #[test]
    fn automatic_edits_use_byte_boundaries_and_preserve_unicode_line_endings() {
        let source = "😀\r\nα + β\u{2028}end";
        let alpha = source.find('α').unwrap();
        let beta = source.find('β').unwrap();
        let result = apply_edits(
            source,
            &[
                vec![Edit::replace([beta, beta + 'β'.len_utf8()], "🐈")],
                vec![Edit::replace([alpha, alpha + 'α'.len_utf8()], "λ")],
                vec![Edit::replace([source.len(), source.len()], "!")],
            ],
        )
        .unwrap();
        assert_eq!(result.output, "😀\r\nλ + 🐈\u{2028}end!");
        assert_eq!(result.applied, [1, 0, 2]);
    }

    #[test]
    fn bom_is_preserved_removed_or_replaced_exactly_once() {
        let source = "\u{feff}const x = 1";
        for (groups, expected) in [
            (
                vec![vec![Edit::replace([14, 14], ";")]],
                "\u{feff}const x = 1;",
            ),
            (vec![vec![Edit::RemoveBom]], "const x = 1"),
            (
                vec![vec![Edit::replace([3, 3], "prefix ")]],
                "\u{feff}prefix const x = 1",
            ),
            (
                vec![vec![Edit::replace([3, 3], "\u{feff}prefix ")]],
                "\u{feff}prefix const x = 1",
            ),
            (
                vec![vec![Edit::replace([3, 8], "\u{feff}let")]],
                "\u{feff}let x = 1",
            ),
            (
                vec![
                    vec![Edit::RemoveBom],
                    vec![Edit::replace([9, 10], "y")],
                    vec![Edit::replace([14, 14], ";")],
                ],
                "const y = 1;",
            ),
        ] {
            let result = apply_edits(source, &groups).unwrap();
            assert_eq!(result.output, expected, "{groups:?}");
            assert!(result.did_apply());
            assert!(result.unapplied.is_empty());
        }
    }

    #[test]
    fn bom_removal_conflicts_with_separate_start_insertion_but_not_own_group() {
        let source = "\u{feff}x";
        let separate = apply_edits(
            source,
            &[vec![Edit::replace([3, 3], "!")], vec![Edit::RemoveBom]],
        )
        .unwrap();
        assert_eq!(separate.output, "x");
        assert_eq!(separate.applied, [1]);
        assert_eq!(separate.unapplied, [0]);
        let atomic =
            apply_edits(source, &[vec![Edit::replace([3, 3], "!"), Edit::RemoveBom]]).unwrap();
        assert_eq!(atomic.output, "!x");
        assert_eq!(atomic.applied, [0]);
        assert!(atomic.unapplied.is_empty());
    }

    #[test]
    fn invalid_native_groups_fail_instead_of_applying_a_partial_fix() {
        for group in [
            vec![Edit::remove([2, 1])],
            vec![Edit::remove([0, 4])],
            vec![Edit::remove([0, 2]), Edit::remove([1, 3])],
            vec![Edit::remove([0, 3]), Edit::replace([1, 1], "x")],
        ] {
            // Validate the conflicting group too, even though an earlier valid
            // diagnostic would otherwise cause it to be deferred.
            assert!(apply_edits("abc", &[vec![Edit::remove([0, 3])], group]).is_err());
        }
        assert!(apply_edits("😀", &[vec![Edit::remove([0, 2])]]).is_err());
        assert!(apply_edits("\u{feff}x", &[vec![Edit::remove([0, 3])]]).is_err());
        assert!(apply_edits("\u{feff}x", &[vec![Edit::RemoveBom, Edit::RemoveBom]]).is_err());
    }
}
