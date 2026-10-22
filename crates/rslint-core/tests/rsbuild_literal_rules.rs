//! Literal-rule API behavior. Production syntax diagnostics gate rule execution;
//! malformed legacy numbers and escapes are also covered at parsed-rule level.

use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, source: &str, extension: &str) -> Value {
    let request = json!({
        "schemaVersion":1,"cwd":"/rslint-literal-tests",
        "files":[{"path":format!("case{extension}"),"text":source}],
        "rules":[{"name":rule}]
    });
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let syntax = result["files"][0]["messages"]
        .as_array()
        .unwrap()
        .iter()
        .any(|m| {
            m["ruleId"]
                .as_str()
                .is_some_and(|id| id.starts_with("TypeScript(TS"))
        });
    assert_eq!(
        result["files"][0]["appliedRules"],
        if syntax { json!([]) } else { json!([rule]) }
    );
    result["files"][0]["messages"].clone()
}

#[test]
fn legacy_numbers_project_syntax_diagnostics_before_rule_execution() {
    let rule = "no-octal";
    let messages = lint(rule, "/* leading */\r\n  -077;", ".js");
    assert_eq!(
        messages,
        json!([{
            "ruleId":"TypeScript(TS1121)","message":"Octal literals are not allowed. Use the syntax '-0o77'.",
            "severity":2,"line":2,"column":3,"endLine":2,"endColumn":7,"range":[17,21]
        }])
    );
    for source in ["08;", "09.1;", "019e1;", "00;"] {
        let messages = lint(rule, source, ".js");
        assert_eq!(messages.as_array().unwrap().len(), 1, "{source}");
        assert!(
            messages[0]["ruleId"]
                .as_str()
                .unwrap()
                .starts_with("TypeScript(TS")
        );
        assert!(messages[0].get("messageId").is_none());
        assert!(messages[0].get("fix").is_none());
        assert!(messages[0].get("suggestions").is_none());
    }
    for source in [
        "0;", "0.1;", "0.5e1;", "0x1234;", "0o17;", "0b101;", "0n;", "1.5;",
    ] {
        assert_eq!(lint(rule, source, ".js"), json!([]), "{source}");
    }
}

#[test]
fn invalid_decimal_escapes_project_syntax_messages_with_utf16_positions() {
    let rule = "no-nonoctal-decimal-escape";
    let source = "\u{feff}'😀\\0\\8';";
    let messages = lint(rule, source, ".js");
    assert_eq!(
        messages,
        json!([{
            "ruleId":"TypeScript(TS1488)","message":"Escape sequence '\\8' is not allowed.",
            "severity":2,"line":1,"column":6,"endLine":1,"endColumn":8,"range":[5,7]
        }])
    );
    let messages = lint(rule, r"'\8\0\9'", ".js");
    assert_eq!(messages.as_array().unwrap().len(), 2);
    for message in messages.as_array().unwrap() {
        assert_eq!(message["ruleId"], "TypeScript(TS1488)");
        assert!(message.get("suggestions").is_none());
    }
    for source in [r"'8'", r"'\\8'", r"'\\9'", r"'\\\\\\8'", r"String.raw`\8`"] {
        assert_eq!(lint(rule, source, ".js"), json!([]), "{source}");
    }
}

#[test]
fn jsx_attribute_backslashes_follow_the_baseline_raw_fallback() {
    let rule = "no-nonoctal-decimal-escape";
    let source = r#"const element = <div title="\8" />;"#;
    let messages = lint(rule, source, ".tsx");
    assert_eq!(messages.as_array().unwrap().len(), 1);
    let start = source.find(r"\8").unwrap();
    assert_eq!(messages[0]["range"], json!([start, start + 2]));
    assert_eq!(messages[0]["messageId"], "decimalEscape");
    assert!(messages[0].get("fix").is_none());
    assert_eq!(messages[0]["suggestions"].as_array().unwrap().len(), 2);
    assert_eq!(
        lint(rule, r#"const element = <div title="\\8" />;"#, ".tsx"),
        json!([])
    );
    let source = "\u{feff}const element = <div title=\"😀\\0\\8\" />;";
    let messages = lint(rule, source, ".tsx");
    assert_eq!(messages.as_array().unwrap().len(), 1);
    let start = source
        .strip_prefix('\u{feff}')
        .unwrap()
        .split(r"\0")
        .next()
        .unwrap()
        .encode_utf16()
        .count();
    assert_eq!(messages[0]["range"], json!([start + 2, start + 4]));
    assert_eq!(
        messages[0]["suggestions"],
        json!([
            {"messageId":"refactor","desc":"Replace '\\0\\8' with '\\u00008'. This maintains the current functionality.","fix":{"range":[start,start+4],"text":"\\u00008"}},
            {"messageId":"refactor","desc":"Replace '\\8' with '\\u0038'. This maintains the current functionality.","fix":{"range":[start+2,start+4],"text":"\\u0038"}},
            {"messageId":"escapeBackslash","desc":"Replace '\\8' with '\\\\8' to include the actual backslash character.","fix":{"range":[start+2,start+4],"text":"\\\\8"}}
        ])
    );
}

#[test]
fn empty_classes_report_once_per_literal_and_keep_constructor_calls_silent() {
    let rule = "no-empty-character-class";
    for source in [
        "const expression = /[]/;",
        "const expression = /[][]/;",
        r"const expression = /\[[]/;",
        "const expression = /[[]]/v;",
        "const expression = /[a--[]]/v;",
        "const expression = /[[]&&b]/v;",
        "const expression = /[a[[b[]c]]d]/v;",
    ] {
        let messages = lint(rule, source, ".js");
        assert_eq!(messages.as_array().unwrap().len(), 1, "{source}");
        assert_eq!(messages[0]["messageId"], "unexpected");
        assert_eq!(messages[0]["message"], "Empty class.");
        assert_eq!(messages[0]["column"], 20);
        assert!(messages[0].get("fix").is_none());
        assert!(messages[0].get("suggestions").is_none());
    }
    for source in [
        "const expression = /[^]/;",
        "const expression = /[[]/;",
        r"const expression = /[\]]/;",
        r"const expression = /\[]/;",
        "const expression = /[[^]]/v;",
        r"const expression = /[\q{ab|cd}]/v;",
        "const expression = new RegExp('[]');",
        "const expression = RegExp('[]');",
        "const expression = '[]';",
    ] {
        assert_eq!(lint(rule, source, ".js"), json!([]), "{source}");
    }
}
