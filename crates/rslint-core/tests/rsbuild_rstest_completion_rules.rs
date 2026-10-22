//! Frozen Rstest contracts: full diagnostic fields, edits and suggestion outputs.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};
const CWD: &str = "/rslint-rstest-completion-tests";
fn run(rule: &str, entry: &Value, fix: bool) -> Value {
    let text = entry["text"].as_str().unwrap();
    let name = format!("input{}", entry["extension"].as_str().unwrap_or(".ts"));
    let options=entry.get("compilerOptions").cloned().unwrap_or_else(||json!({"strict":true,"target":"ESNext","module":"CommonJS","moduleResolution":"Node10","jsx":"preserve"}));
    let request = json!({"schemaVersion":1,"cwd":CWD,"fix":fix,
        "files":[{"path":name,"text":text,"rules":[{"name":rule}],"policy":{"kind":"explicit","configFiles":[format!("{CWD}/tsconfig.json")]}}],
        "sources":[{"path":"tsconfig.json","text":json!({"compilerOptions":options,"files":[name]}).to_string()}]});
    let output = lint_projects_json(&request.to_string())
        .unwrap_or_else(|error| panic!("{rule}: {}\n{text}\n{error}", entry["origin"]));
    let result: Value = serde_json::from_str(&output).unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = result["files"][0].clone();
    assert_eq!(file["appliedRules"], json!([rule]), "{entry}\n{file}");
    assert!(file["hasSyntaxErrors"].is_null(), "{entry}\n{file}");
    file
}
fn apply(text: &str, fix: &Value) -> String {
    let units: Vec<_> = text.encode_utf16().collect();
    let start = fix["range"][0].as_u64().unwrap() as usize;
    let end = fix["range"][1].as_u64().unwrap() as usize;
    String::from_utf16(&units[..start]).unwrap()
        + fix["text"].as_str().unwrap()
        + &String::from_utf16(&units[end..]).unwrap()
}
fn corpus(rule: &str) {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-rstest-completion.json"
    ))
    .unwrap();
    let suite = corpus
        .as_array()
        .unwrap()
        .iter()
        .find(|g| g["rule"] == rule)
        .unwrap();
    for category in ["valid", "invalid"] {
        let cases = suite[category].as_array().unwrap();
        assert!(!cases.is_empty());
        for case in cases {
            let file = run(rule, case, false);
            let messages = file["messages"].as_array().unwrap();
            assert_eq!(
                messages.is_empty(),
                category == "valid",
                "{}\n{}\n{file}",
                case["origin"],
                case["text"]
            );
            if let Some(expected) = case["expected"].as_array() {
                assert_eq!(messages.len(), expected.len(), "{case}\n{file}");
                for (message, expected) in messages.iter().zip(expected) {
                    assert_eq!(message["ruleId"], rule);
                    assert_eq!(message["severity"], 2);
                    for (key, value) in expected.as_object().unwrap() {
                        if key != "suggestions" {
                            assert_eq!(&message[key], value, "{case}\n{file}");
                        }
                    }
                    let empty = vec![];
                    let suggestions = message["suggestions"].as_array().unwrap_or(&empty);
                    let expected_suggestions = expected["suggestions"].as_array().unwrap_or(&empty);
                    assert_eq!(
                        suggestions.len(),
                        expected_suggestions.len(),
                        "{case}\n{file}"
                    );
                    for (suggestion, expected) in suggestions.iter().zip(expected_suggestions) {
                        assert_eq!(suggestion["messageId"], expected["messageId"]);
                        assert_eq!(
                            apply(case["text"].as_str().unwrap(), &suggestion["fix"]),
                            expected["output"].as_str().unwrap(),
                            "{case}\n{file}"
                        );
                    }
                }
            }
            if let Some(expected) = case.get("expectedOutput") {
                let fixed = run(rule, case, true);
                assert_eq!(&fixed["output"], expected, "{case}\n{fixed}");
            }
            if rule != "rstest/valid-expect" {
                assert!(
                    messages.iter().all(|m| m.get("fix").is_none()),
                    "{case}\n{file}"
                );
            }
        }
    }
}
#[test]
fn tests_require_assertions() {
    corpus("rstest/expect-expect");
}
#[test]
fn unconditional_assertions() {
    corpus("rstest/no-conditional-expect");
}
#[test]
fn concurrent_snapshot_context() {
    corpus("rstest/require-local-test-context-for-concurrent-snapshots");
}
#[test]
fn valid_assertion_calls() {
    corpus("rstest/valid-expect");
}
