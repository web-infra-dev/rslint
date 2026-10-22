//! Frozen rsbuild's default type-aware options: positive and negative cases,
//! exact locations, diagnostic text, and suggestion edit payloads.

use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, sources: &[&str]) -> Vec<Value> {
    sources
        .iter()
        .enumerate()
        .map(|(i, text)| {
            // Keep each fixture's script/module identity exactly as written. Separate
            // sessions avoid cross-case global declarations without appending exports.
            let request = json!({
                "schemaVersion": 1, "cwd": "/rslint-rsbuild-typed-tests",
                "files": [{ "path": format!("case-{i}.ts"), "text": text }],
                "rules": [{ "name": rule }]
            });
            let result: Value =
                serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
            let files = result["files"].as_array().unwrap();
            assert_eq!(files.len(), 1, "silent files remain admitted");
            let file = &files[0];
            assert_eq!(
                file["filePath"],
                format!("/rslint-rsbuild-typed-tests/case-{i}.ts")
            );
            assert_eq!(file["appliedRules"], json!([rule]));
            file["messages"].clone()
        })
        .collect()
}

fn case_sources(rule: &str, key: &str) -> Vec<String> {
    let cases: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-typed.json"
    ))
    .unwrap();
    cases
        .as_array()
        .unwrap()
        .iter()
        .find(|case| case["rule"] == rule)
        .unwrap()[key]
        .as_array()
        .unwrap()
        .iter()
        .map(|value| value.as_str().unwrap().to_string())
        .collect()
}

fn verify_cases(rule: &str) {
    for key in ["valid", "invalid"] {
        let sources = case_sources(rule, key);
        assert!(!sources.is_empty());
        let messages = lint(
            rule,
            &sources.iter().map(String::as_str).collect::<Vec<_>>(),
        );
        for (source, messages) in sources.iter().zip(messages) {
            let messages = messages.as_array().unwrap();
            if key == "valid" {
                assert!(messages.is_empty(), "{rule}: {source}: {messages:?}");
            } else {
                assert!(!messages.is_empty(), "{rule}: {source}");
            }
            for diagnostic in messages {
                assert_eq!(diagnostic["ruleId"], rule);
                assert_eq!(diagnostic["severity"], 2);
                assert!(
                    diagnostic["fix"].is_null(),
                    "these rules must never autofix"
                );
            }
        }
    }
}

#[test]
fn for_in_array_resolves_union_intersection_constraint_and_structural_length() {
    verify_cases("@typescript-eslint/no-for-in-array");
}

#[test]
fn array_delete_resolves_arrays_tuples_constraints_and_nullable_unions() {
    verify_cases("@typescript-eslint/no-array-delete");
}

#[test]
fn plus_operands_resolves_types_and_keeps_default_permissive_options() {
    verify_cases("@typescript-eslint/restrict-plus-operands");
}

#[test]
fn only_throw_error_resolves_library_identity_inheritance_and_rejection_handlers() {
    verify_cases("@typescript-eslint/only-throw-error");
}

#[test]
fn array_delete_preserves_comments_parentheses_and_exact_suggestion() {
    let source =
        "declare const a: number[]; delete /* 😀 */ (a /* left */ [ /* middle */ 0 /* right */ ]);";
    let messages = lint("@typescript-eslint/no-array-delete", &[source]);
    let diagnostic = &messages[0][0];
    assert_eq!(diagnostic["messageId"], "noArrayDelete");
    assert_eq!(
        diagnostic["message"],
        "Using the `delete` operator with an array expression is unsafe."
    );
    assert_eq!(
        diagnostic["range"],
        json!([27, source.encode_utf16().count() - 1])
    );
    assert_eq!(
        diagnostic["suggestions"],
        json!([{
            "messageId":"useSplice", "desc":"Use `array.splice()` instead.",
            "fix":{"range":[27, source.encode_utf16().count() - 2], "text":" /* 😀 */ (a /* left */ .splice( /* middle */ 0 /* right */ , 1)"}
        }])
    );
}

#[test]
fn for_in_range_ends_before_body_trivia() {
    let source = "/* 😀 */ for (const key\n in (([1, 2]))\n )\n /* body */ {}";
    let messages = lint("@typescript-eslint/no-for-in-array", &[source]);
    let diagnostic = &messages[0][0];
    let end = source.find("\n /* body */").unwrap();
    assert_eq!(
        diagnostic["range"],
        json!([9, source[..end].encode_utf16().count()])
    );
    assert_eq!(diagnostic["line"], 1);
    assert_eq!(diagnostic["column"], 10);
    assert_eq!(diagnostic["endLine"], 3);
    assert_eq!(diagnostic["endColumn"], 3);
    assert_eq!(diagnostic["messageId"], "forInViolation");
}

#[test]
fn plus_reports_each_object_union_part_and_bigint_number_pair() {
    let sources = [
        "function run(a: string | { a: 1 } | { b: 2 }) { return a + 'a'; }",
        "const value = 1 + 2n;",
    ];
    let messages = lint("@typescript-eslint/restrict-plus-operands", &sources);
    assert_eq!(messages[0].as_array().unwrap().len(), 2);
    let prefix = "Invalid operand for a '+' operation. Operands must each be a number or string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`. Got ";
    assert_eq!(messages[0][0]["message"], format!("{prefix}`{{ a: 1; }}`."));
    assert_eq!(messages[0][1]["message"], format!("{prefix}`{{ b: 2; }}`."));
    assert_eq!(messages[1][0]["messageId"], "bigintAndNumber");
    assert_eq!(
        messages[1][0]["message"],
        "Numeric '+' operations must either be both bigints or both numbers. Got `number` + `bigint`."
    );
    assert_eq!(messages[1][0]["range"], json!([14, 20]));
}

#[test]
fn only_throw_reports_expression_including_parentheses() {
    let messages = lint(
        "@typescript-eslint/only-throw-error",
        &["throw (undefined);", "throw 'oops';"],
    );
    assert_eq!(messages[0][0]["messageId"], "undef");
    assert_eq!(messages[0][0]["message"], "Do not throw undefined.");
    assert_eq!(messages[0][0]["range"], json!([6, 17]));
    assert_eq!(messages[1][0]["messageId"], "object");
    assert_eq!(
        messages[1][0]["message"],
        "Expected an error object to be thrown."
    );
    assert_eq!(messages[1][0]["range"], json!([6, 12]));
}
