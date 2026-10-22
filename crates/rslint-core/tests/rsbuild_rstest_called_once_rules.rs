//! The frozen rule's full test inputs, source-order boundaries and exact edits.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

const RULE: &str = "rstest/prefer-called-exactly-once-with";

fn lint(text: &str, fix: bool, typed: bool) -> Value {
    let root = "/rslint-rstest-called-once";
    let request = json!({
        "schemaVersion":1,"cwd":root,"fix":fix,
        "files":[{"path":"input.ts","text":text,"rules":[{"name":RULE}],
            "policy":if typed {json!({"kind":"explicit","configFiles":[format!("{root}/tsconfig.json")]})}
                else {json!({"kind":"sourceOnly"})}}],
        "sources":[{"path":"tsconfig.json","text":json!({"compilerOptions":{
            "jsx":"preserve","target":"ESNext","module":"CommonJS","strict":true,
            "esModuleInterop":true,"lib":["es2015","es2017","esnext","DOM"],"types":[]
        },"files":["input.ts"]}).to_string()}]
    });
    let result: Value = serde_json::from_str(
        &lint_projects_json(&request.to_string()).unwrap_or_else(|error| panic!("{error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = result["files"][0].clone();
    assert_eq!(file["appliedRules"], json!([RULE]));
    assert_eq!(file["hasTypeInformation"], typed);
    assert!(file["hasSyntaxErrors"].is_null(), "{text}: {file}");
    for message in file["messages"].as_array().unwrap() {
        assert_eq!(message["ruleId"], RULE);
        assert_eq!(message["messageId"], "preferCalledExactlyOnceWith");
        assert_eq!(message["severity"], 2);
        assert!(message["suggestions"].is_null());
    }
    file
}

#[test]
fn all_frozen_inputs_preserve_reports_and_fix_outputs() {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-rstest-called-once.json"
    ))
    .unwrap();
    let group = &groups[0];
    assert_eq!(group["originalSkip"], json!([]));
    assert_eq!(group["nonDefaultOptions"], json!([]));
    assert_eq!(group["valid"].as_array().unwrap().len(), 92);
    assert_eq!(group["invalid"].as_array().unwrap().len(), 54);
    for category in ["valid", "invalid"] {
        for entry in group[category].as_array().unwrap() {
            let text = entry["text"].as_str().unwrap();
            let file = lint(text, false, true);
            let messages = file["messages"].as_array().unwrap();
            if category == "valid" {
                assert!(messages.is_empty(), "{}\n{text}\n{file}", entry["origin"]);
            } else {
                assert!(!messages.is_empty(), "{}\n{text}\n{file}", entry["origin"]);
            }
            if let Some(expected) = entry["expected"].as_array() {
                assert_eq!(
                    messages.len(),
                    expected.len(),
                    "{}\n{text}\n{file}",
                    entry["origin"]
                );
                for (actual, expected) in messages.iter().zip(expected) {
                    for (key, value) in expected.as_object().unwrap() {
                        assert_eq!(&actual[key], value, "{}\n{text}\n{file}", entry["origin"]);
                    }
                }
            }
            if let Some(output_steps) = entry.get("outputSteps") {
                let fixed = lint(text, true, true);
                if output_steps.is_null() {
                    assert!(fixed["output"].is_null(), "{text}: {fixed}");
                } else {
                    assert_eq!(
                        &fixed["output"],
                        output_steps.as_array().unwrap().last().unwrap(),
                        "{}\n{text}\n{fixed}",
                        entry["origin"]
                    );
                }
                let after = fixed["output"].as_str().unwrap_or(text);
                assert_eq!(
                    fixed["messages"],
                    lint(after, false, true)["messages"],
                    "Post-fix diagnostics must be a fresh observation: {text}"
                );
            }
        }
    }
}

#[test]
fn exact_diagnostic_and_atomic_fix_keep_intervening_source() {
    let text = "expect(x).toHaveBeenCalledOnce();\nexpect(x).toHaveBeenCalledWith('a');";
    let file = lint(text, false, true);
    assert_eq!(
        file["messages"],
        json!([{
            "ruleId":RULE,"messageId":"preferCalledExactlyOnceWith","severity":2,
            "message":"Using `toHaveBeenCalledOnce` and `toHaveBeenCalledWith` on the same target; prefer `toHaveBeenCalledExactlyOnceWith` instead.",
        "line":2,"column":11,"endLine":2,"endColumn":31,"range":[44,64],
        "fix":{"range":[0,64],"text":"expect(x).toHaveBeenCalledExactlyOnceWith"}
        }])
    );
    let fixed = lint(text, true, true);
    assert_eq!(
        fixed["output"],
        "expect(x).toHaveBeenCalledExactlyOnceWith('a');"
    );
    assert_eq!(fixed["messages"], json!([]));
}

#[test]
fn source_only_intervals_do_not_infer_library_call_safety() {
    let text = "expect(x).toHaveBeenCalledOnce(); console.log('checkpoint'); expect(x).toHaveBeenCalledWith('a');";
    assert_eq!(lint(text, false, false)["messages"], json!([]));
    assert_eq!(
        lint(text, false, true)["messages"]
            .as_array()
            .unwrap()
            .len(),
        1
    );
    for middle in [
        "",
        "const unrelated = 1; let value; var other = 2;",
        "expect(other).toBe(1);",
    ] {
        let text = format!(
            "expect(x).toHaveBeenCalledOnce(); {middle} expect(x).toHaveBeenCalledWith('a');"
        );
        assert_eq!(
            lint(&text, false, false)["messages"]
                .as_array()
                .unwrap()
                .len(),
            1
        );
    }
}

#[test]
fn quote_delimiters_trivia_and_unicode_fix_ranges_are_preserved() {
    for accessor in [
        "\"calledWith\"",
        "'calledWith'",
        "`calledWith`",
        "(\"calledWith\")",
    ] {
        let text = format!("/* 😀 */ expect(x).calledOnce; expect(x)[{accessor}]('a');");
        let fixed = lint(&text, true, true);
        assert_eq!(
            fixed["output"],
            format!(
                "/* 😀 */  expect(x)[{}]('a');",
                accessor.replace("calledWith", "calledOnceWith")
            )
        );
        assert_eq!(fixed["messages"], json!([]));
    }
    let text = "expect(x).calledOnce; // preserve\r\nexpect(x).calledWith('a');";
    assert_eq!(
        lint(text, true, true)["output"],
        " // preserve\r\nexpect(x).calledOnceWith('a');"
    );
}

#[test]
fn self_contained_and_unstable_assertions_are_report_only() {
    for text in [
        "expect(x).calledOnce.and.calledWith('a');",
        "expect(getMock()).calledOnce; expect(getMock()).calledWith('a');",
        "expect(x, message()).calledOnce; expect(x, message()).calledWith('a');",
        "expect(x).calledOnce.and.ok; expect(x).calledWith('a');",
    ] {
        let file = lint(text, true, true);
        assert_eq!(
            file["messages"].as_array().unwrap().len(),
            1,
            "{text}: {file}"
        );
        assert!(file["messages"][0]["fix"].is_null(), "{text}: {file}");
        assert!(file["output"].is_null());
    }
}
