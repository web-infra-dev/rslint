//! Source-pinned regex RuleTester inputs, including suggestion application and
//! exact messages/locations wherever the frozen source specifies them. The
//! independent Node oracle compares every public diagnostic field as well.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};
fn lint(rule: &str, entry: &Value, fix: bool) -> Value {
    let extension = entry["extension"].as_str().unwrap_or(".ts");
    let root = "/rslint-rsbuild-regex-checks";
    let path = format!("case{extension}");
    let request = json!({"schemaVersion":1,"cwd":root,
        "sources":[{"path":"tsconfig.json","text":json!({"compilerOptions":{
            "strict":true,"allowJs":true,"target":"ES2022","module":"ESNext",
            "moduleResolution":"Bundler","jsx":"preserve"
        },"files":[path]}).to_string()}],
        "files":[{"path":path,"text":entry["text"],"rules":[{"name":rule}],
            "policy":{"kind":"explicit","configFiles":[format!("{root}/tsconfig.json")]}}],"fix":fix});
    let output: Value = serde_json::from_str(
        &lint_projects_json(&request.to_string())
            .unwrap_or_else(|e| panic!("{rule}: {e}\n{}", entry["text"])),
    )
    .unwrap();
    assert_eq!(
        output["files"].as_array().unwrap().len(),
        1,
        "silent files must be retained"
    );
    assert_eq!(
        output["files"][0]["appliedRules"],
        json!([rule]),
        "{}",
        entry["text"]
    );
    output["files"][0].clone()
}
fn apply(source: &str, fix: &Value) -> String {
    let units: Vec<_> = source.encode_utf16().collect();
    let range = fix["range"].as_array().unwrap();
    let start = range[0].as_u64().unwrap() as usize;
    let end = range[1].as_u64().unwrap() as usize;
    assert!(start <= end && end <= units.len());
    let mut result = units[..start].to_vec();
    result.extend(fix["text"].as_str().unwrap().encode_utf16());
    result.extend(&units[end..]);
    String::from_utf16(&result).unwrap()
}
fn corpus(rule: &str) {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-regex-checks.json"
    ))
    .unwrap();
    let group = groups
        .as_array()
        .unwrap()
        .iter()
        .find(|g| g["rule"] == rule)
        .unwrap();
    for category in ["valid", "invalid"] {
        assert!(!group[category].as_array().unwrap().is_empty());
        for entry in group[category].as_array().unwrap() {
            let file = lint(rule, entry, false);
            let actual = file["messages"].as_array().unwrap();
            let mut expected = entry["expected"].as_array().cloned().unwrap_or_default();
            expected.sort_by_key(|m| {
                (
                    m["line"].as_u64().unwrap_or(1),
                    m["column"].as_u64().unwrap_or(1),
                )
            });
            assert_eq!(
                actual.len(),
                expected.len(),
                "{rule}/{category}: {}\n{}\n{actual:?}",
                entry["origin"],
                entry["text"]
            );
            if expected
                .iter()
                .any(|message| message.get("line").is_none() || message.get("column").is_none())
            {
                // Some frozen RuleTester cases specify only message IDs in
                // listener traversal order. The public API instead sorts by
                // source position (a nested class can therefore come second).
                // With no expected location, substituting line/column 1 cannot
                // recover that order. Pair those partial expectations with the
                // public messages using every supplied diagnostic field, then
                // retain all ordinary field/suggestion/count assertions below.
                let mut unmatched = expected;
                expected = actual
                    .iter()
                    .map(|message| {
                        let index = unmatched
                            .iter()
                            .position(|expected| {
                                expected.as_object().unwrap().iter().all(|(key, value)| {
                                    key == "suggestions" || &message[key] == value
                                })
                            })
                            .unwrap_or_else(|| {
                                panic!(
                                    "{rule}: no frozen expectation matches {message}\n{}\n{}\n{unmatched:?}",
                                    entry["origin"], entry["text"]
                                )
                            });
                        unmatched.remove(index)
                    })
                    .collect();
                assert!(
                    unmatched.is_empty(),
                    "every frozen expectation must be used"
                );
            }
            for (actual, expected) in actual.iter().zip(expected) {
                assert_eq!(actual["ruleId"], rule);
                assert_eq!(actual["severity"], 2);
                for (key, value) in expected.as_object().unwrap() {
                    if key != "suggestions" {
                        assert_eq!(
                            &actual[key], value,
                            "{rule}: {}\n{}\n{actual}",
                            entry["origin"], entry["text"]
                        );
                    }
                }
                if let Some(suggestions) = expected.get("suggestions") {
                    let expected = suggestions.as_array().unwrap();
                    let actual = actual["suggestions"]
                        .as_array()
                        .cloned()
                        .unwrap_or_default();
                    assert_eq!(actual.len(), expected.len(), "{}", entry["text"]);
                    for (actual, expected) in actual.iter().zip(expected) {
                        assert_eq!(
                            actual["messageId"], expected["MessageId"],
                            "{}",
                            entry["text"]
                        );
                        if let Some(message) = expected.get("Message") {
                            assert_eq!(&actual["desc"], message);
                        }
                        if let Some(output) = expected.get("Output") {
                            assert_eq!(
                                apply(entry["text"].as_str().unwrap(), &actual["fix"]),
                                output.as_str().unwrap(),
                                "{}",
                                entry["text"]
                            );
                        }
                    }
                }
                if rule != "no-regex-spaces" {
                    assert!(actual.get("fix").is_none());
                }
            }
            if let Some(outputs) = entry.get("outputSteps") {
                let fixed = lint(rule, entry, true);
                if outputs.is_null() {
                    assert!(fixed.get("output").is_none());
                } else {
                    assert_eq!(
                        fixed["output"],
                        *outputs.as_array().unwrap().last().unwrap(),
                        "{}",
                        entry["text"]
                    );
                }
                let mut after = entry.clone();
                if let Some(output) = fixed.get("output") {
                    after["text"] = output.clone();
                }
                assert_eq!(
                    fixed["messages"],
                    lint(rule, &after, false)["messages"],
                    "post-fix diagnostics must be fresh"
                );
            }
        }
    }
}
#[test]
fn frozen_no_control_regex() {
    corpus("no-control-regex");
}
#[test]
fn frozen_no_invalid_regexp() {
    corpus("no-invalid-regexp");
}
#[test]
fn frozen_no_misleading_character_class() {
    corpus("no-misleading-character-class");
}
#[test]
fn frozen_no_regex_spaces() {
    corpus("no-regex-spaces");
}
#[test]
fn frozen_no_useless_backreference() {
    corpus("no-useless-backreference");
}
#[test]
fn frozen_no_useless_escape() {
    corpus("no-useless-escape");
}

#[test]
fn constructor_reference_tracking_preserves_runtime_global_scope() {
    for (source, count) in [
        (r#"type RegExp = string; RegExp("\\x1f")"#, 1),
        (r#"interface RegExp {} RegExp("\\x1f")"#, 1),
        (r#"/* global RegExp:off */ RegExp("\\x1f")"#, 0),
        (r#"const RegExp = other; RegExp("\\x1f")"#, 0),
    ] {
        assert_eq!(
            lint("no-control-regex", &json!({"text":source}), false)["messages"]
                .as_array()
                .unwrap()
                .len(),
            count,
            "{source}"
        );
    }
    for source in [
        r#"let R; R = RegExp; new R("[🇯🇵]", "u");"#,
        r#"const root = globalThis; const { ["Reg" + "Exp"]: R } = root; R("[Á]", "");"#,
        r#"let R; ({ RegExp: R } = globalThis); R("[🇯🇵]", "u");"#,
        r#"{ namespace RegExp {} RegExp("[Á]", ""); } RegExp("[Á]", "");"#,
    ] {
        assert_eq!(
            lint(
                "no-misleading-character-class",
                &json!({"text":source}),
                false
            )["messages"]
                .as_array()
                .unwrap()
                .len(),
            1,
            "{source}"
        );
    }
}
#[test]
fn block_namespaces_do_not_shadow_outer_global_references() {
    for (source, expected_columns) in [
        (r#"{ namespace RegExp {} } RegExp("[Á]", "");"#, vec![34]),
        (
            r#"{ namespace RegExp {} RegExp("[Á]", ""); } RegExp("[Á]", "");"#,
            vec![54],
        ),
        (r#"{ namespace RegExp {} RegExp("[Á]", ""); }"#, vec![]),
        (r#"namespace RegExp {} RegExp("[Á]", "");"#, vec![]),
    ] {
        let file = lint(
            "no-misleading-character-class",
            &json!({"text":source}),
            false,
        );
        let actual_columns = file["messages"]
            .as_array()
            .unwrap()
            .iter()
            .map(|message| message["column"].as_u64().unwrap())
            .collect::<Vec<_>>();
        assert_eq!(actual_columns, expected_columns, "{source}");
    }
}
#[test]
fn static_pattern_fallbacks_keep_precise_or_deduplicated_locations() {
    for source in [
        r#"new RegExp(`${"[👍🇯🇵]"}[😊]`);"#,
        r#"const RawString = String; new RegExp(RawString.raw`[👍]`)"#,
        r#"new RegExp(true ? "[👍]" : "[a]");"#,
        r#"new RegExp(String("[👍]"));"#,
    ] {
        let file = lint(
            "no-misleading-character-class",
            &json!({"text":source}),
            false,
        );
        assert_eq!(file["messages"].as_array().unwrap().len(), 1, "{source}");
        assert_eq!(
            file["messages"][0]["messageId"],
            "surrogatePairWithoutUFlag"
        );
    }
    for source in [
        r#"let pattern = "[👍]"; new RegExp(pattern);"#,
        r#"const pattern = "[👍]" as string; new RegExp(pattern);"#,
        r#"const pattern = "[👍]"; new RegExp(pattern);"#,
    ] {
        let file = lint(
            "no-misleading-character-class",
            &json!({"text":source}),
            false,
        );
        assert_eq!(file["messages"].as_array().unwrap().len(), 1, "{source}");
        assert_eq!(
            file["messages"][0]["column"],
            if source.starts_with("let ") { 17 } else { 19 },
            "{source}"
        );
    }
    let source = r#"let pattern = "[abc]"; pattern = "[👍]"; new RegExp(pattern);"#;
    assert!(
        lint(
            "no-misleading-character-class",
            &json!({"text":source}),
            false
        )["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
}
#[test]
fn regex_error_recovery_matches_frozen_character_events() {
    let source = r#"RegExp("\\x1f\\u{NOT_HEX}", "u")"#;
    assert_eq!(
        lint("no-control-regex", &json!({"text":source}), false)["messages"][0]["message"],
        "Unexpected control character(s) in regular expression: \\x1f."
    );
    let source = r#"var r = /[🇯🇵--X]/v"#;
    assert_eq!(
        lint(
            "no-misleading-character-class",
            &json!({"text":source}),
            false
        )["messages"][0]["messageId"],
        "regionalIndicatorSymbol"
    );
    let source = r#"RegExp("[^a\\p{Script=Gara}]  ", "v")"#;
    assert!(
        lint("no-regex-spaces", &json!({"text":source}), false)["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
}
#[test]
fn static_backreference_patterns_and_flow_narrowing() {
    for source in [
        r#"const pattern = '\\1(a)'; RegExp((((pattern as string) satisfies unknown)!));"#,
        r#"new RegExp(String.raw`\1${"(a)"}`);"#,
        r#"new RegExp(true ? '\\1(a)' : '(a)');"#,
        r#"new RegExp(String('\\1(a)'));"#,
        r#"const RawString = String; new RegExp(RawString.raw`\1(a)`);"#,
        "export {};\ndeclare global { var RegExp: any; }\nnew RegExp('\\\\1(a)');",
    ] {
        let file = lint("no-useless-backreference", &json!({"text":source}), false);
        assert_eq!(file["messages"].as_array().unwrap().len(), 1, "{source}");
        assert_eq!(file["messages"][0]["messageId"], "forward");
    }
    let source = r#"const R = window.RegExp; R('(\\1)')"#;
    assert!(
        lint("no-useless-backreference", &json!({"text":source}), false)["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
    let source = r#"declare const condition: boolean;
declare function isRegExpConstructor(value: RegExpConstructor | ((pattern: string) => string)): value is RegExpConstructor;
const r: RegExpConstructor | ((pattern: string) => string) = condition ? RegExp : value => value;
if (isRegExpConstructor(r)) { r('\\1(a)'); } else { r('\\1(a)'); }"#;
    let file = lint("no-useless-backreference", &json!({"text":source}), false);
    assert_eq!(file["messages"].as_array().unwrap().len(), 1);
    assert_eq!(file["messages"][0]["line"], 4);
}
