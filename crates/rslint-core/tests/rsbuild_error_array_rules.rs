//! Frozen default-option caught-error and array predicate contracts.
//! Support files retain the original Program without adding lint targets.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

fn corpus() -> Value {
    serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-error-array.json"
    ))
    .unwrap()
}

fn lint(group: &Value, entry: &Value, source: &str, fix: bool) -> Value {
    let root = "/rslint-error-array-tests";
    let fixture = &group["fixtureSets"][entry["fixtureSet"].as_str().unwrap()];
    let name = entry["fileName"].as_str().unwrap();
    let mut sources = vec![json!({"path":"tsconfig.json","text":json!({
        "compilerOptions":fixture["compilerOptions"],"files":entry["projectFiles"]
    }).to_string()})];
    for (path, text) in fixture["files"].as_object().unwrap() {
        if path != name {
            sources.push(json!({"path":path,"text":text}));
        }
    }
    let request = json!({"schemaVersion":1,"cwd":root,"singleThreaded":true,
    "sources":sources,"fix":fix,"files":[{
        "path":name,"text":source,"rules":[{"name":group["rule"]}],
        "policy":{"kind":"explicit","configFiles":[format!("{root}/tsconfig.json")]}
    }]});
    let result: Value = serde_json::from_str(
        &lint_projects_json(&request.to_string())
            .unwrap_or_else(|error| panic!("{entry}\n{error}")),
    )
    .unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = result["files"][0].clone();
    assert_eq!(file["appliedRules"], json!([group["rule"]]));
    assert_eq!(file["hasTypeInformation"], true);
    assert!(
        file["hasSyntaxErrors"].is_null() || file["hasSyntaxErrors"] == false,
        "{entry}: {file}"
    );
    for report in file["messages"].as_array().unwrap() {
        assert_eq!(report["ruleId"], group["rule"]);
        assert_eq!(report["severity"], 2);
    }
    file
}

fn suggestion_output(source: &str, suggestion: &Value) -> String {
    let units = source.encode_utf16().collect::<Vec<_>>();
    let start = suggestion["fix"]["range"][0].as_u64().unwrap() as usize;
    let end = suggestion["fix"]["range"][1].as_u64().unwrap() as usize;
    assert!(start <= end && end <= units.len());
    String::from_utf16(&units[..start]).unwrap()
        + suggestion["fix"]["text"].as_str().unwrap()
        + &String::from_utf16(&units[end..]).unwrap()
}

fn expected_matches(source: &str, expected: &Value, actual: &Value) -> bool {
    for field in [
        "messageId",
        "message",
        "line",
        "column",
        "endLine",
        "endColumn",
    ] {
        if let Some(value) = expected.get(field)
            && value != &actual[field]
        {
            return false;
        }
    }
    let expected_suggestions = expected["suggestions"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    let actual_suggestions = actual["suggestions"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    expected_suggestions.len() == actual_suggestions.len()
        && expected_suggestions
            .iter()
            .zip(actual_suggestions.iter())
            .all(|(expected, actual)| {
                expected["messageId"] == actual["messageId"]
                    && expected["desc"] == actual["desc"]
                    && expected["output"].as_str()
                        == Some(suggestion_output(source, actual).as_str())
            })
}

fn run_rule(rule: &str, valid: usize, invalid: usize, excluded: usize) {
    let corpus = corpus();
    let group = corpus
        .as_array()
        .unwrap()
        .iter()
        .find(|group| group["rule"] == rule)
        .unwrap();
    assert_eq!(group["valid"].as_array().unwrap().len(), valid);
    assert_eq!(group["invalid"].as_array().unwrap().len(), invalid);
    assert_eq!(group["excluded"].as_array().unwrap().len(), excluded);
    for category in ["valid", "invalid"] {
        for entry in group[category].as_array().unwrap() {
            let source = entry["text"].as_str().unwrap();
            let plain = lint(group, entry, source, false);
            let mut actual = plain["messages"].as_array().unwrap().clone();
            let expected = entry["expected"].as_array().cloned().unwrap_or_default();
            assert_eq!(actual.len(), expected.len(), "{entry}: {plain}");
            assert_eq!(actual.is_empty(), category == "valid", "{entry}: {plain}");
            for expected in expected {
                let index = actual
                    .iter()
                    .position(|actual| expected_matches(source, &expected, actual));
                assert!(
                    index.is_some(),
                    "missing expected diagnostic {expected}\n{entry}\n{plain}"
                );
                let report = actual.remove(index.unwrap());
                assert_eq!(
                    report.get("fix").is_some(),
                    entry["outputSteps"].is_array(),
                    "{entry}: {report}"
                );
            }
            assert!(plain.get("output").is_none());
            let fixed = lint(group, entry, source, true);
            if let Some(steps) = entry["outputSteps"].as_array() {
                assert!(!steps.is_empty());
                assert_eq!(fixed["output"], *steps.last().unwrap(), "{entry}: {fixed}");
                let after = fixed["output"].as_str().unwrap();
                assert_eq!(
                    fixed["messages"],
                    lint(group, entry, after, false)["messages"],
                    "final diagnostics are computed after autofix"
                );
            } else {
                assert!(
                    fixed.get("output").is_none(),
                    "suggestions are not autofixes: {entry}: {fixed}"
                );
                assert_eq!(plain["messages"], fixed["messages"], "{entry}");
            }
        }
    }
}

#[test]
fn caught_error_default_cases_preserve_messages_ranges_and_suggestions() {
    run_rule("preserve-caught-error", 66, 82, 53);
}

#[test]
fn array_some_cases_preserve_types_messages_autofixes_and_suggestions() {
    run_rule("unicorn/prefer-array-some", 156, 105, 0);
}

#[test]
fn nondefault_configuration_exclusions_and_explicit_defaults_are_visible() {
    let corpus = corpus();
    let group = &corpus[0];
    let mut global_overrides = 0;
    for entry in group["excluded"].as_array().unwrap() {
        assert!(entry["origin"].is_string() && entry["reason"].is_string());
        assert_ne!(entry["originalSkip"], true);
        if !entry["globals"].is_null() {
            assert_eq!(entry["globals"], json!({"Error":"off"}));
            global_overrides += 1;
        } else {
            assert!(!entry["options"].is_null());
        }
    }
    assert_eq!(global_overrides, 1);
    let defaults = group["valid"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(|entry| entry.get("originalOptions"))
        .collect::<Vec<_>>();
    assert_eq!(defaults.len(), 3);
    assert!(defaults.contains(&&json!({})));
    assert!(defaults.contains(&&json!({"requireCatchParameter":false})));
    assert!(defaults.contains(&&json!({"errorClassNames":[]})));
}
