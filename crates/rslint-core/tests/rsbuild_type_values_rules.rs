//! Frozen rsbuild defaults: builtin declaration identity, template interpolation,
//! and rejection reasons. Each oracle input retains its original module scope.

use rslint_core::lint_json;
use serde_json::{Value, json};

const TEMPLATE: &str = "@typescript-eslint/restrict-template-expressions";
const REJECT: &str = "@typescript-eslint/prefer-promise-reject-errors";

fn lint(rule: &str, sources: &[(String, String)], options: Option<Value>) -> Vec<Value> {
    // Preserve each oracle input's script/module status and global scope. Each
    // file receives a separate program; unrelated fixtures must not merge globals.
    sources
        .iter()
        .enumerate()
        .map(|(index, (text, extension))| {
            let mut request = json!({
                "schemaVersion": 1, "cwd": "/rslint-rsbuild-type-values-tests",
                "files": [{"path": format!("case-{index}{extension}"), "text": text}],
                "rules": [{"name": rule}],
            });
            if let Some(options) = &options {
                request["compilerOptions"] = options.clone();
            }
            let result: Value =
                serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
            let files = result["files"].as_array().unwrap();
            assert_eq!(files.len(), 1, "silent files remain admitted");
            assert_eq!(
                files[0]["filePath"],
                format!("/rslint-rsbuild-type-values-tests/case-{index}{extension}")
            );
            assert_eq!(files[0]["appliedRules"], json!([rule]));
            files[0]["messages"].clone()
        })
        .collect()
}

fn one(rule: &str, source: &str, options: Option<Value>) -> Vec<Value> {
    lint(rule, &[(source.to_string(), ".ts".to_string())], options)[0]
        .as_array()
        .unwrap()
        .clone()
}

fn verify_cases(rule: &str) {
    let cases: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-type-values.json"
    ))
    .unwrap();
    let group = cases
        .as_array()
        .unwrap()
        .iter()
        .find(|group| group["rule"] == rule)
        .unwrap();
    for category in ["valid", "invalid"] {
        let entries = group[category].as_array().unwrap();
        assert!(!entries.is_empty());
        for entry in entries {
            let (source, extension) = if let Some(text) = entry.as_str() {
                (text.to_string(), ".ts".to_string())
            } else {
                (
                    entry["text"].as_str().unwrap().to_string(),
                    entry["extension"].as_str().unwrap_or(".ts").to_string(),
                )
            };
            let options = entry.get("compilerOptions").map(|overrides| {
                let mut options = json!({"strict":true, "allowJs":true, "target":9, "module":99, "moduleResolution":100});
                options.as_object_mut().unwrap().extend(overrides.as_object().unwrap().clone());
                options
            });
            let messages = lint(rule, &[(source.clone(), extension)], options).remove(0);
            let messages = messages.as_array().unwrap();
            if category == "valid" {
                assert!(messages.is_empty(), "{rule}: {source}: {messages:?}");
            } else {
                assert!(!messages.is_empty(), "{rule}: {source}");
            }
            for diagnostic in messages {
                assert_eq!(diagnostic["ruleId"], rule);
                assert_eq!(diagnostic["severity"], 2);
                assert!(diagnostic["fix"].is_null());
                assert!(
                    diagnostic["suggestions"].is_null(),
                    "these type-value rules do not offer edits"
                );
            }
        }
    }
}

#[test]
fn templates_preserve_default_primitive_flags_union_intersection_and_lib_identity() {
    verify_cases(TEMPLATE);
}

#[test]
fn rejects_preserve_promise_identity_executor_symbols_and_readonly_error_semantics() {
    verify_cases(REJECT);
}

#[test]
fn template_reports_each_bad_expression_with_its_exact_type_and_utf16_location() {
    let source = "declare const a: unknown; declare const b: object; const msg = `😀 ${a} ${b}`;";
    let messages = one(TEMPLATE, source, None);
    assert_eq!(messages.len(), 2);
    for (message, marker, type_name) in [
        (&messages[0], "${a}", "unknown"),
        (&messages[1], "${b}", "object"),
    ] {
        let byte_start = source.find(marker).unwrap() + 2;
        let start = source[..byte_start].encode_utf16().count();
        assert_eq!(message["messageId"], "invalidType");
        assert_eq!(
            message["message"],
            format!("Invalid type \"{type_name}\" of template literal expression.")
        );
        assert_eq!(message["range"], json!([start, start + 1]));
        assert_eq!(message["line"], 1);
        assert_eq!(message["column"], start + 1);
        assert_eq!(message["endLine"], 1);
        assert_eq!(message["endColumn"], start + 2);
    }
}

#[test]
fn promise_report_covers_the_whole_call_including_parenthesized_callee() {
    let source = "/* 😀 */ (Promise.reject)( /* reason */ 'bad' );";
    let messages = one(REJECT, source, None);
    assert_eq!(messages.len(), 1);
    let start = source[..source.find("(Promise").unwrap()]
        .encode_utf16()
        .count();
    let end = source.encode_utf16().count() - 1;
    assert_eq!(messages[0]["messageId"], "rejectAnError");
    assert_eq!(
        messages[0]["message"],
        "Expected the Promise rejection reason to be an Error."
    );
    assert_eq!(messages[0]["range"], json!([start, end]));
    assert_eq!(messages[0]["column"], start + 1);
    assert_eq!(messages[0]["endColumn"], end + 1);
}

#[test]
fn nested_templates_are_checked_even_when_the_outer_template_is_tagged() {
    let source = "declare const value: object; const x = tag`outer ${`inner ${value}`}`;";
    let messages = one(TEMPLATE, source, None);
    assert_eq!(messages.len(), 1);
    assert_eq!(
        messages[0]["message"],
        "Invalid type \"object\" of template literal expression."
    );
    let start = source.rfind("value").unwrap();
    assert_eq!(messages[0]["range"], json!([start, start + 5]));
}

#[test]
fn reject_checks_only_the_first_reason_and_tracks_shadowed_executor_parameters() {
    let source = "new Promise((resolve, reject) => { reject(new Error(), 'ignored'); function inner(reject: (value: unknown) => void) { reject('ignored'); } reject('bad', new Error()); });";
    let messages = one(REJECT, source, None);
    assert_eq!(messages.len(), 1);
    let start = source.find("reject('bad'").unwrap();
    let end = start + "reject('bad', new Error())".len();
    assert_eq!(messages[0]["range"], json!([start, end]));
    assert_eq!(messages[0]["messageId"], "rejectAnError");
}
