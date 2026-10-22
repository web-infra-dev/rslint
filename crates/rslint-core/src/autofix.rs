//! Bounded, in-memory apply-and-observe pipeline. Files are never written here.

use std::collections::{BTreeMap, BTreeSet};

use crate::{FileObservation, FileResult, edits};

const MAX_ROUNDS: usize = 10;

pub(super) struct Outcome {
    pub files: Vec<FileResult>,
    /// Accepted diagnostic groups across every round, including no-op edits.
    pub fixed_issues: usize,
}

pub(super) fn run(
    initial: Vec<FileObservation>,
    mut observe: impl FnMut(&BTreeMap<String, String>) -> Result<Vec<FileObservation>, String>,
) -> Result<Outcome, String> {
    let original: Vec<_> = initial.iter().map(|file| file.result.clone()).collect();
    let original_sources: BTreeMap<_, _> = initial
        .iter()
        .map(|file| (file.result.file_path.clone(), file.source.clone()))
        .collect();
    let mut sources = original_sources.clone();
    let mut current = initial;
    let mut ever_fixed = BTreeSet::new();
    let mut fixed_issues = 0usize;
    for _ in 0..MAX_ROUNDS {
        let mut applied_any = false;
        for file in &current {
            let groups: Vec<_> = file
                .result
                .messages
                .iter()
                .map(|m| m.edits.clone())
                .collect();
            let applied = edits::apply_edits(&file.source, &groups)?;
            debug_assert_eq!(
                applied.applied.len() + applied.unapplied.len(),
                groups.len()
            );
            if applied.did_apply() {
                fixed_issues = fixed_issues
                    .checked_add(applied.applied.len())
                    .ok_or("applied diagnostic count exceeds the supported range")?;
                applied_any = true;
                ever_fixed.insert(file.result.file_path.clone());
                sources.insert(file.result.file_path.clone(), applied.output);
            }
        }
        if !applied_any {
            break;
        }
        // This includes applied no-op edits. The original API returns output
        // for every ever-fixed file even when all texts return to their origin.
        if sources == original_sources {
            return Ok(Outcome {
                files: outputs(original, &sources, &ever_fixed),
                fixed_issues,
            });
        }
        // Rebuild/check *all* admitted files, including unchanged importers.
        // The tenth application is followed by one final observation as well.
        current = observe(&sources)?;
        if current.len() != original.len()
            || current.iter().zip(&original).any(|(file, origin)| {
                file.result.file_path != origin.file_path
                    || sources.get(&file.result.file_path) != Some(&file.source)
            })
        {
            return Err(
                "autofix observation changed the admitted files or source generation".into(),
            );
        }
    }
    Ok(Outcome {
        files: outputs(
            current.into_iter().map(|f| f.result).collect(),
            &sources,
            &ever_fixed,
        ),
        fixed_issues,
    })
}

fn outputs(
    mut files: Vec<FileResult>,
    sources: &BTreeMap<String, String>,
    ever_fixed: &BTreeSet<String>,
) -> Vec<FileResult> {
    for file in &mut files {
        if ever_fixed.contains(&file.file_path) {
            file.output = Some(sources[&file.file_path].clone());
        }
    }
    files
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::Diagnostic;

    fn file(name: &str, source: &str, after: Option<&str>) -> FileObservation {
        FileObservation {
            result: FileResult {
                file_path: name.into(),
                messages: after
                    .map(|after| Diagnostic {
                        rule_id: "test-only-state-machine".into(),
                        message_id: "replace".into(),
                        message: source.into(),
                        severity: 2,
                        line: 1,
                        column: 1,
                        end_line: 1,
                        end_column: source.len() + 1,
                        range: [0, source.len()],
                        fix: None,
                        suggestions: vec![],
                        edits: vec![edits::Edit::replace([0, source.len()], after)],
                    })
                    .into_iter()
                    .collect(),
                applied_rules: vec![],
                has_type_information: None,
                output: None,
                has_syntax_errors: false,
            },
            source: source.into(),
        }
    }

    #[test]
    fn no_edits_does_not_reobserve_or_add_output() {
        let result = run(vec![file("a", "original", None)], |_| {
            panic!("unexpected observation")
        })
        .unwrap();
        assert!(result.files[0].output.is_none());
        assert_eq!(result.fixed_issues, 0);
    }

    #[test]
    fn counts_atomic_diagnostics_once_and_defers_overlaps() {
        for apply_deferred in [false, true] {
            let mut initial = file("a", "abcd", Some("unused"));
            initial.result.messages[0].edits = vec![
                edits::Edit::replace([0, 1], "A"),
                edits::Edit::replace([3, 4], "D"),
            ];
            let mut overlapping = initial.result.messages[0].clone();
            overlapping.edits = vec![edits::Edit::replace([1, 2], "B")];
            initial.result.messages.push(overlapping);
            let mut observations = 0;
            let result = run(vec![initial], |sources| {
                observations += 1;
                let mut next = file("a", &sources["a"], None);
                if observations == 1 && apply_deferred {
                    let mut diagnostic = file("a", &sources["a"], Some("unused"))
                        .result
                        .messages
                        .pop()
                        .unwrap();
                    diagnostic.edits = vec![edits::Edit::replace([1, 2], "B")];
                    next.result.messages.push(diagnostic);
                }
                Ok(vec![next])
            })
            .unwrap();
            // Two edits in the first diagnostic count once. The conflicting
            // diagnostic counts only if it is actually applied in a later pass.
            assert_eq!(result.fixed_issues, if apply_deferred { 2 } else { 1 });
            assert_eq!(
                result.files[0].output.as_deref(),
                Some(if apply_deferred { "ABcD" } else { "AbcD" })
            );
        }
    }

    #[test]
    fn no_op_and_empty_output_still_record_application() {
        let result = run(vec![file("a", "same", Some("same"))], |_| {
            panic!("restored initial")
        })
        .unwrap();
        assert_eq!(result.files[0].output.as_deref(), Some("same"));
        assert_eq!(result.fixed_issues, 1);
        assert_eq!(result.files[0].messages[0].message, "same");
        let mut observations = 0;
        let result = run(vec![file("a", "x", Some(""))], |sources| {
            observations += 1;
            assert_eq!(sources["a"], "");
            Ok(vec![file("a", "", None)])
        })
        .unwrap();
        assert_eq!(observations, 1);
        assert_eq!(result.files[0].output.as_deref(), Some(""));
        assert_eq!(result.fixed_issues, 1);
        assert!(result.files[0].messages.is_empty());
    }

    #[test]
    fn true_multipass_stops_after_no_edits_and_returns_final_diagnostics() {
        let mut observations = 0;
        let result = run(
            vec![file("a", "a", Some("b")), file("quiet", "keep", None)],
            |sources| {
                observations += 1;
                assert_eq!(sources["quiet"], "keep");
                let source = &sources["a"];
                Ok(vec![
                    file("a", source, if source == "b" { Some("c") } else { None }),
                    file("quiet", "keep", None),
                ])
            },
        )
        .unwrap();
        assert_eq!(observations, 2);
        assert_eq!(result.files[0].output.as_deref(), Some("c"));
        assert_eq!(result.fixed_issues, observations);
        assert!(result.files[0].messages.is_empty());
        assert!(result.files[1].output.is_none());
    }

    #[test]
    fn initial_cycle_restores_original_observation_without_extra_check() {
        let mut observations = 0;
        let result = run(vec![file("a", "a", Some("b"))], |sources| {
            observations += 1;
            Ok(vec![file("a", &sources["a"], Some("a"))])
        })
        .unwrap();
        assert_eq!(observations, 1);
        assert_eq!(result.files[0].output.as_deref(), Some("a"));
        assert_eq!(result.fixed_issues, 2);
        assert_eq!(result.files[0].messages[0].message, "a");
    }

    #[test]
    fn partial_restore_does_not_stop_other_files() {
        let mut observations = 0;
        let result = run(
            vec![file("a", "a", Some("b")), file("b", "1", Some("2"))],
            |sources| {
                observations += 1;
                Ok(vec![
                    file("a", &sources["a"], (observations == 1).then_some("a")),
                    file("b", &sources["b"], (observations == 1).then_some("3")),
                ])
            },
        )
        .unwrap();
        assert_eq!(observations, 2);
        assert_eq!(result.files[0].output.as_deref(), Some("a"));
        assert!(result.files[0].messages.is_empty());
        assert_eq!(result.files[1].output.as_deref(), Some("3"));
        assert_eq!(result.fixed_issues, 4);
    }

    #[test]
    fn non_initial_cycle_uses_ten_applications_and_final_observation() {
        let mut observations = 0;
        let result = run(vec![file("a", "start", Some("b"))], |sources| {
            observations += 1;
            let source = &sources["a"];
            Ok(vec![file(
                "a",
                source,
                Some(if source == "b" { "c" } else { "b" }),
            )])
        })
        .unwrap();
        assert_eq!(observations, 10);
        assert_eq!(result.files[0].output.as_deref(), Some("c"));
        assert_eq!(result.fixed_issues, observations);
        assert_eq!(result.files[0].messages[0].message, "c");
    }

    #[test]
    fn errors_or_incomplete_observations_never_return_partial_success() {
        assert_eq!(
            run(vec![file("a", "a", Some("b"))], |_| Err(
                "compiler failed".into()
            ))
            .err()
            .as_deref(),
            Some("compiler failed")
        );
        for next in [
            vec![],
            vec![file("other", "b", None)],
            vec![file("a", "stale", None)],
        ] {
            let mut next = Some(next);
            assert!(
                run(vec![file("a", "a", Some("b"))], |_| Ok(next
                    .take()
                    .unwrap()))
                .is_err()
            );
        }
    }
}
