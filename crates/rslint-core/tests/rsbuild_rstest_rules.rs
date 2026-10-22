//! Frozen Rstest rule inputs plus exact diagnostic and edit contracts.
use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, text: &str, fix: bool) -> Value {
    serde_json::from_str(
        &lint_json(
            &json!({
                "schemaVersion": 1, "cwd": "/rslint-rstest-tests", "fix": fix,
                "files": [{"path": "test.ts", "text": text}], "rules": [{"name": rule}]
            })
            .to_string(),
        )
        .unwrap(),
    )
    .unwrap()
}
#[test]
fn upstream_rule_inputs_and_regressions() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-rstest.json"
    ))
    .unwrap();
    for suite in corpus.as_array().unwrap() {
        let rule = suite["rule"].as_str().unwrap();
        for category in ["valid", "invalid"] {
            let cases = suite[category].as_array().unwrap();
            assert!(!cases.is_empty());
            for case in cases {
                let source = case.as_str().or_else(|| case["text"].as_str()).unwrap();
                let result = lint(rule, source, false);
                let file = &result["files"][0];
                assert_eq!(result["files"].as_array().unwrap().len(), 1);
                if case["expectSyntaxErrors"] == true {
                    assert_eq!(file["hasSyntaxErrors"], true, "{rule}: {source}");
                    assert_eq!(file["appliedRules"], json!([]));
                    assert!(file["messages"].as_array().unwrap().iter().all(|message| {
                        message["ruleId"]
                            .as_str()
                            .unwrap()
                            .starts_with("TypeScript(TS")
                    }));
                } else {
                    assert_eq!(file["appliedRules"], json!([rule]), "{rule}: {source}");
                    let messages = file["messages"].as_array().unwrap();
                    assert_eq!(
                        messages.is_empty(),
                        category == "valid",
                        "{rule}: {source}\n{result}"
                    );
                    for message in messages {
                        assert_eq!(message["ruleId"], rule);
                        assert_eq!(message["severity"], 2);
                    }
                }
            }
        }
    }
}
#[test]
fn imports_preserve_quotes_and_prefer_existing_rstack_spelling() {
    let source =
        "import { expect } from 'rstack/test';\nimport { test as run } from \"node:test\";";
    let result = lint("rstest/no-import-node-test", source, false);
    let diagnostic = &result["files"][0]["messages"][0];
    assert_eq!(diagnostic["messageId"], "noImportNodeTest");
    assert_eq!(
        diagnostic["message"],
        "Do not import the Node test runner in Rstest test files"
    );
    assert_eq!(diagnostic["range"], json!([38, 78]));
    assert_eq!(diagnostic["fix"]["text"], "\"rstack/test\"");
    let fixed = lint("rstest/no-import-node-test", source, true);
    assert_eq!(
        fixed["files"][0]["output"],
        source.replace("node:test", "rstack/test")
    );
    assert_eq!(fixed["files"][0]["messages"], json!([]));
}
#[test]
fn disabled_registration_can_report_both_skip_and_missing_body() {
    let result = lint("rstest/no-disabled-tests", "test.skip('pending');", false);
    let diagnostics = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(diagnostics.len(), 2);
    assert_eq!(diagnostics[0]["range"], json!([0, 20]));
    assert_eq!(diagnostics[1]["range"], json!([0, 20]));
    let ids: Vec<_> = diagnostics
        .iter()
        .map(|message| message["messageId"].as_str().unwrap())
        .collect();
    assert!(ids.contains(&"skippedTest") && ids.contains(&"missingFunction"));
    assert!(
        diagnostics
            .iter()
            .all(|message| message.get("fix").is_none() && message.get("suggestions").is_none())
    );
}
#[test]
fn alias_focus_suggestions_deduplicate_the_original_accessor() {
    let source = "const focused = test.only; focused('one', fn); focused('two', fn);";
    let result = lint("rstest/no-focused-tests", source, false);
    let diagnostics = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(diagnostics.len(), 2);
    assert_eq!(diagnostics[0]["messageId"], "focusedTest");
    assert_eq!(diagnostics[0]["range"], json!([27, 34]));
    assert_eq!(diagnostics[1]["range"], json!([47, 54]));
    let suggestion = &diagnostics[0]["suggestions"][0];
    assert_eq!(suggestion["messageId"], "suggestRemoveFocus");
    assert_eq!(suggestion["fix"]["range"], json!([20, 25]));
    assert_eq!(suggestion["fix"]["text"], "");
    assert!(diagnostics[1].get("suggestions").is_none());
    let fixed = lint("rstest/no-focused-tests", source, true);
    assert!(
        fixed["files"][0].get("output").is_none(),
        "suggestions must not be applied as fixes"
    );
}
#[test]
fn inline_snapshot_selects_only_the_runtime_snapshot_argument() {
    let source = "expect(value).toMatchInlineSnapshot({}, `${snapshot}`, `case ${id}`);";
    let result = lint("rstest/no-interpolation-in-snapshots", source, false);
    let diagnostics = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(diagnostics.len(), 1);
    assert_eq!(diagnostics[0]["messageId"], "noInterpolation");
    assert_eq!(
        diagnostics[0]["message"],
        "Do not use string interpolation inside of snapshots"
    );
    assert_eq!(diagnostics[0]["range"], json!([40, 53]));
    assert!(diagnostics[0].get("fix").is_none());
    assert!(diagnostics[0].get("suggestions").is_none());
}
