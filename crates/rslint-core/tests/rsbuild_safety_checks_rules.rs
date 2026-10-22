use rslint_core::lint_projects_json;
use serde_json::{Value, json};

fn lint(rule: &str, text: &str, fix: bool) -> Value {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-safety","fix":fix,
        "files":[{"path":"input.ts","text":text,"rules":[{"name":rule}],"policy":{"kind":"explicit","configFiles":["/rslint-safety/tsconfig.json"]}}],
        "sources":[{"path":"tsconfig.json","text":r#"{"files":["input.ts"],"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler"}}"#}]});
    let result: Value = serde_json::from_str(
        &lint_projects_json(&request.to_string()).unwrap_or_else(|error| panic!("{error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = result["files"][0].clone();
    assert_eq!(file["appliedRules"], json!([rule]), "{text}");
    assert!(
        file.get("hasSyntaxErrors")
            .is_none_or(|value| value == &Value::Bool(false)),
        "{text}\n{file}"
    );
    file
}

fn corpus(rule: &str) {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-safety-checks.json"
    ))
    .unwrap();
    let group = groups
        .as_array()
        .unwrap()
        .iter()
        .find(|g| g["rule"] == rule)
        .unwrap();
    for (category, valid) in [("valid", true), ("invalid", false)] {
        let cases = group[category].as_array().unwrap();
        assert!(!cases.is_empty());
        for case in cases {
            let text = case["text"].as_str().unwrap();
            let file = lint(rule, text, false);
            let messages = file["messages"].as_array().unwrap();
            assert_eq!(
                messages.is_empty(),
                valid,
                "{}\n{text}\n{file}",
                case["origin"]
            );
            if let Some(expected) = case.get("expected").and_then(Value::as_array) {
                assert_eq!(messages.len(), expected.len(), "{text}\n{file}");
                let mut expected = expected.clone();
                if expected
                    .iter()
                    .all(|v| v.get("line").is_some() && v.get("column").is_some())
                {
                    expected.sort_by_key(|v| (v["line"].as_u64(), v["column"].as_u64()));
                }
                for (actual, expected) in messages.iter().zip(expected) {
                    for key in [
                        "messageId",
                        "message",
                        "line",
                        "column",
                        "endLine",
                        "endColumn",
                    ] {
                        if let Some(value) = expected.get(key) {
                            assert_eq!(&actual[key], value, "{key}: {text}\n{file}");
                        }
                    }
                }
            }
            if let Some(steps) = case.get("outputSteps").and_then(Value::as_array) {
                assert!(!steps.is_empty());
                let fixed = lint(rule, text, true);
                assert_eq!(&fixed["output"], steps.last().unwrap(), "{text}\n{fixed}");
            }
        }
    }
}

#[test]
fn frozen_unused_label_cases() {
    corpus("no-unused-labels");
}
#[test]
fn frozen_optional_chain_cases() {
    corpus("no-unsafe-optional-chaining");
}

#[test]
fn label_fixes_preserve_comments_directives_and_nested_usage() {
    for text in ["label: /* comment */ work();", "label: 'use strict';"] {
        let file = lint("no-unused-labels", text, true);
        assert!(!file["messages"].as_array().unwrap().is_empty());
        assert!(file.get("output").is_none(), "{file}");
    }
    // The outer label contains another label, so removing it cannot create a
    // directive. The remaining inner label still protects the string-like body.
    let file = lint(
        "no-unused-labels",
        "function f() { outer: inner: `directive`; }",
        true,
    );
    assert_eq!(file["output"], "function f() { inner: `directive`; }");
    assert_eq!(
        file["messages"],
        json!([{
            "ruleId": "no-unused-labels",
            "severity": 2,
            "messageId": "unused",
            "message": "'inner:' is defined but never used.",
            "line": 1,
            "column": 16,
            "endLine": 1,
            "endColumn": 21,
            "range": [15, 20],
        }])
    );
    assert!(
        lint("no-unused-labels", file["output"].as_str().unwrap(), true)
            .get("output")
            .is_none()
    );
    let file = lint(
        "no-unused-labels",
        "/* 😀 */ outer: { unused: while (true) { break outer; } }",
        true,
    );
    assert_eq!(
        file["output"],
        "/* 😀 */ outer: { while (true) { break outer; } }"
    );
    assert_eq!(file["messages"], json!([]));
}

#[test]
fn optional_chain_reports_both_conditional_branches_with_exact_utf16_spans() {
    let file = lint(
        "no-unsafe-optional-chaining",
        "/* 😀 */ (ok ? a?.b : c?.d)();",
        false,
    );
    assert_eq!(file["messages"].as_array().unwrap().len(), 2);
    for (message, range) in file["messages"]
        .as_array()
        .unwrap()
        .iter()
        .zip([[15, 19], [22, 26]])
    {
        assert_eq!(message["range"], json!(range));
        assert_eq!(message["messageId"], "unsafeOptionalChain");
        assert!(message.get("fix").is_none());
        assert!(message.get("suggestions").is_none());
    }
    assert_eq!(
        lint(
            "no-unsafe-optional-chaining",
            "(a?.b ?? fallback)(); const value = a?.b + 1;",
            false
        )["messages"],
        json!([])
    );
}
