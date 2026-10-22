//! Frozen variable-rule corpus through the configured tsrs Program API. The
//! private-member rule's suggestion edits are applied independently to the
//! original text; none of these rules may automatically modify a source file.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

fn lint(rule: &str, case: &Value, fix: bool) -> Result<Value, String> {
    let cwd = "/rslint-rsbuild-variable-usage";
    let extension = case["extension"].as_str().unwrap_or(".ts");
    let name = case["fileName"]
        .as_str()
        .map(str::to_owned)
        .unwrap_or_else(|| format!("case{extension}"));
    let config = format!("{cwd}/tsconfig.json");
    let mut compiler_options = json!({"strict":true,"allowJs":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler","jsx":"preserve"});
    if let Some(options) = case["compilerOptions"].as_object() {
        compiler_options
            .as_object_mut()
            .unwrap()
            .extend(options.clone());
    }
    let request = json!({"schemaVersion":1,"cwd":cwd,
        "files":[{"path":name,"text":case["text"],"rules":[{"name":rule,"options":[]}],"policy":{"kind":"explicit","configFiles":[config]}}],
        "sources":[{"path":config,"text":json!({"files":[name],"compilerOptions":compiler_options}).to_string()}],"fix":fix});
    let result: Value = serde_json::from_str(&lint_projects_json(&request.to_string())?).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = result["files"][0].clone();
    assert_eq!(file["appliedRules"], json!([rule]));
    assert!(
        file.get("output").is_none(),
        "{rule} must not automatically edit text"
    );
    for message in file["messages"].as_array().unwrap() {
        assert_eq!(message["ruleId"], rule);
        assert_eq!(message["severity"], 2);
        assert!(message.get("fix").is_none());
        if rule != "no-unused-private-class-members" {
            assert!(
                message
                    .get("suggestions")
                    .is_none_or(|s| s.as_array().is_some_and(Vec::is_empty))
            );
        }
    }
    Ok(file)
}
fn apply_suggestion(source: &str, fix: &Value) -> String {
    let utf16: Vec<_> = source.encode_utf16().collect();
    let start = fix["range"][0].as_u64().unwrap() as usize;
    let end = fix["range"][1].as_u64().unwrap() as usize;
    assert!(start <= end && end <= utf16.len());
    let mut result = String::from_utf16(&utf16[..start]).unwrap();
    result.push_str(fix["text"].as_str().unwrap());
    result.push_str(&String::from_utf16(&utf16[end..]).unwrap());
    result
}
fn corpus(rule: &str) {
    let suites: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-variable-usage.json"
    ))
    .unwrap();
    let suite = suites
        .as_array()
        .unwrap()
        .iter()
        .find(|suite| suite["rule"] == rule)
        .unwrap();
    let mut failures = Vec::new();
    for category in ["valid", "invalid"] {
        assert!(!suite[category].as_array().unwrap().is_empty());
        for case in suite[category].as_array().unwrap() {
            let file = lint(rule, case, true).unwrap_or_else(|error| {
                panic!("{rule} {}: {error}\n{}", case["origin"], case["text"])
            });
            let messages = file["messages"].as_array().unwrap();
            let mut expected = case["expected"].as_array().cloned().unwrap_or_default();
            expected.sort_by_key(|message| {
                (
                    message["line"].as_u64().unwrap_or(1),
                    message["column"].as_u64().unwrap_or(1),
                )
            });
            if messages.len() != expected.len() {
                failures.push(format!(
                    "{category}: {}\n{}\nexpected:{expected:?}\nactual:{messages:?}",
                    case["origin"], case["text"]
                ));
            }
            for (actual, expected) in messages.iter().zip(expected) {
                for (key, value) in expected.as_object().unwrap() {
                    if key == "suggestions" {
                        let actual_suggestions =
                            actual[key].as_array().cloned().unwrap_or_default();
                        let expected_suggestions = value.as_array().unwrap();
                        if actual_suggestions.len() != expected_suggestions.len() {
                            failures.push(format!(
                                "{}: suggestion count: {} != {}\n{}",
                                case["origin"],
                                actual_suggestions.len(),
                                expected_suggestions.len(),
                                case["text"]
                            ));
                        }
                        for (actual, expected) in
                            actual_suggestions.iter().zip(expected_suggestions)
                        {
                            if actual["messageId"] != expected["messageId"] {
                                failures.push(format!(
                                    "{}: suggestion messageId {actual} != {expected}",
                                    case["origin"]
                                ));
                            }
                            let output =
                                apply_suggestion(case["text"].as_str().unwrap(), &actual["fix"]);
                            if output != expected["output"].as_str().unwrap() {
                                failures.push(format!(
                                    "{}: suggestion output\nexpected:{}\nactual:{output}",
                                    case["origin"], expected["output"]
                                ));
                            }
                        }
                    } else if &actual[key] != value {
                        failures.push(format!(
                            "{}: {key}: expected {value}, actual {}\n{}\n{actual}",
                            case["origin"], actual[key], case["text"]
                        ));
                    }
                }
            }
        }
    }
    assert!(
        failures.is_empty(),
        "{} preserved failures for {rule}:\n{}",
        failures.len(),
        failures.join("\n\n")
    );
}
#[test]
fn frozen_redeclaration_scope_and_binding_ranges() {
    corpus("no-redeclare");
}
#[test]
fn frozen_unassigned_symbol_and_runtime_reference_cases() {
    corpus("no-unassigned-vars");
}
#[test]
fn frozen_private_member_reads_writes_and_suggestion_outputs() {
    corpus("no-unused-private-class-members");
}
#[test]
fn frozen_useless_assignment_control_flow_and_bitset_boundaries() {
    corpus("no-useless-assignment");
}

#[test]
fn preserved_global_directives_match_frozen_diagnostics_and_ranges() {
    let suites: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-variable-usage.json"
    ))
    .unwrap();
    let mut pending = 0;
    for suite in suites.as_array().unwrap() {
        for case in suite["pendingCompatibility"].as_array().unwrap() {
            let result = lint(suite["rule"].as_str().unwrap(), case, false).unwrap();
            let messages = result["messages"].as_array().unwrap();
            let expected = case["expected"].as_array().cloned().unwrap_or_default();
            assert_eq!(messages.len(), expected.len(), "{case}\n{result}");
            for (message, expected) in messages.iter().zip(expected) {
                for (key, value) in expected.as_object().unwrap() {
                    assert_eq!(&message[key], value, "{case}\n{result}");
                }
            }
            assert!(case["origin"].is_string() && case["reason"].is_string());
            if case["category"] == "invalid" {
                assert!(!case["expected"].as_array().unwrap().is_empty());
            }
            pending += 1;
        }
    }
    assert_eq!(
        pending, 11,
        "every preserved shared-directive case is checked against its frozen expectation"
    );
}
#[test]
fn unicode_ranges_and_disable_comments_do_not_remove_dataflow_work() {
    let cases = [
        (
            "no-unassigned-vars",
            json!({"text":"export let /* 😀 */ value: string; consume(value);","expected":[{"messageId":"unassigned","column":21,"endColumn":34}]}),
        ),
        (
            "no-unassigned-vars",
            json!({"text":"let value; consume(value);\n// eslint-disable-next-line no-unassigned-vars\nvalue = 1;","expected":[]}),
        ),
        (
            "no-useless-assignment",
            json!({"text":"function f(g) {\n// eslint-disable-next-line no-useless-assignment\nlet a=0; a=1; g(a);\nlet /* 😀 */ value=0; value=1; g(value);\n}","expected":[{"messageId":"unnecessaryAssignment","line":4,"column":14,"endColumn":19}]}),
        ),
    ];
    for (rule, case) in cases {
        let file = lint(rule, &case, true).unwrap();
        let messages = file["messages"].as_array().unwrap();
        assert_eq!(
            messages.len(),
            case["expected"].as_array().unwrap().len(),
            "{case}\n{file}"
        );
        for (actual, expected) in messages.iter().zip(case["expected"].as_array().unwrap()) {
            for (key, value) in expected.as_object().unwrap() {
                assert_eq!(&actual[key], value, "{case}\n{actual}");
            }
        }
    }
}
