//! Suppression and explicit unsupported-directive handling use real comments,
//! never strings, regular expressions, template text, or JSX text.

use rslint_core::lint_json;
use serde_json::{Value, json};

fn request(files: &[(&str, &str)]) -> String {
    json!({
        "schemaVersion":1,
        "cwd":"/rslint-comment-tests",
        "files":files.iter().map(|(path,text)| json!({"path":path,"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":"no-debugger"}],
        "compilerOptions":{"strict":true,"target":9,"module":99,"moduleResolution":100,"jsx":1,"allowJs":true}
    }).to_string()
}

fn supported(files: &[(&str, &str)]) {
    let output = lint_json(&request(files)).unwrap();
    let output: Value = serde_json::from_str(&output).unwrap();
    let actual = output["files"].as_array().unwrap();
    assert_eq!(actual.len(), files.len());
    for (actual, (path, _)) in actual.iter().zip(files) {
        assert_eq!(actual["filePath"], format!("/rslint-comment-tests/{path}"));
        assert_eq!(actual["appliedRules"], json!(["no-debugger"]));
        assert_eq!(actual["messages"], json!([]));
    }
}

fn rejected(path: &str, source: &str, marker: &str, kind: &str, content: &str) {
    let error = lint_json(&request(&[(path, source)])).unwrap_err();
    let offset = source.find(marker).unwrap();
    assert_eq!(
        error,
        format!(
            "unsupported {kind} lint directive at byte {offset}: {content} in /rslint-comment-tests/{path}"
        ),
        "{source:?}"
    );
}

#[test]
fn ordinary_comments_and_line_only_nondirective_prefixes_remain_supported() {
    supported(&[
        ("empty.ts", ""),
        ("comment-only.ts", "/* first */\r\n// second"),
        ("shebang.ts", "#!/usr/bin/env node\r\n/* ordinary */\r\n"),
        (
            "tokens.ts",
            "/* first */ const /* name */ value /* equals */ = /* number */ 1; // tail\n",
        ),
        (
            "line-prefixes.ts",
            "// global custom\n// globals custom\n// exported custom\n// eslint no-debugger: off\nconst custom=1;",
        ),
        (
            "ordinary-prefixes.ts",
            "/* discusses eslint-disable */ /* globalThis */ /* exportedValue */ /* rslint setting */",
        ),
        (
            "jsdoc.js",
            "/** @param {number[]} values */\nfunction use(values) { return values; } // normal\n",
        ),
    ]);
}

#[test]
fn directive_like_string_regex_template_and_jsx_contents_are_not_comments() {
    supported(&[
        (
            "strings.ts",
            r#"const a = "/* eslint-disable */"; const b = '// rslint-disable';"#,
        ),
        (
            "regexp.ts",
            r#"const a = /[/*] eslint-disable [*/]/; const b = /https?:\/\/eslint-disable/;"#,
        ),
        (
            "templates.ts",
            "const a = `/* eslint-disable */ ${1 /* ordinary */ + 2} // rslint-disable`;",
        ),
        (
            "nested-templates.ts",
            "const a = `/* eslint-disable */ ${`// rslint-disable ${1}`} /* global x */`;",
        ),
        (
            "jsx.tsx",
            "const value = <><A />/* eslint-disable */{1}/* global x */<A></A>/* rslint-disable */<></>/* exported x */</>;",
        ),
        (
            "jsx-attribute.tsx",
            "const value = <A text='/* eslint-disable */' other={'// rslint-disable'}>/* global x */</A>;",
        ),
        (
            "js-whitespace.ts",
            "/* \u{0085}eslint-disable */ /* \u{200b}rslint-disable */",
        ),
    ]);
}

#[test]
fn disable_and_enable_comments_are_accepted_without_skipping_file_admission() {
    supported(&[
        (
            "next-line.ts",
            "// eslint-disable-next-line no-debugger\ndebugger;",
        ),
        (
            "same-line.ts",
            "debugger; // rslint-disable-line no-debugger",
        ),
        ("all.ts", "/* eslint-disable */ debugger;"),
        ("enable.ts", "/* rslint-enable */ const value = 1;"),
        (
            "shebang.ts",
            "#!/usr/bin/env node\n// rslint-disable\ndebugger;",
        ),
        (
            "ordinary.ts",
            "/* ordinary */\n// eslint-disable\ndebugger;",
        ),
    ]);
}

#[test]
fn auxiliary_directive_forms_remain_explicitly_rejected_in_real_comments() {
    for (source, marker, kind, content) in [
        (
            "/*eslint no-debugger: off*/ debugger;",
            "/*",
            "InlineConfig",
            "eslint no-debugger: off",
        ),
        (
            "/*global external:readonly*/ external;",
            "/*",
            "Globals",
            "global external:readonly",
        ),
        (
            "/*globals external:false*/ external;",
            "/*",
            "Globals",
            "globals external:false",
        ),
        ("// eslint-unknown", "//", "Unknown", "eslint-unknown"),
        (
            "/*\u{feff}\u{00a0}eslint-disable */",
            "/*",
            "Unknown",
            "eslint-disable",
        ),
    ] {
        rejected("directive.ts", source, marker, kind, content);
    }
}

#[test]
fn only_matching_report_ranges_are_suppressed_and_enables_restore_reports() {
    let source = "debugger;\n/* eslint-disable no-debugger */\ndebugger;\n/* rslint-enable no-debugger */\ndebugger;\n// eslint-disable-next-line no-console\ndebugger;";
    let output: Value =
        serde_json::from_str(&lint_json(&request(&[("transitions.ts", source)])).unwrap()).unwrap();
    let files = output["files"].as_array().unwrap();
    assert_eq!(files.len(), 1);
    assert_eq!(files[0]["appliedRules"], json!(["no-debugger"]));
    let messages = files[0]["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 3);
    for (message, line) in messages.iter().zip([1, 5, 7]) {
        let start = source
            .split_inclusive('\n')
            .take(line - 1)
            .map(str::len)
            .sum::<usize>();
        assert_eq!(
            message,
            &json!({
                "ruleId":"no-debugger", "messageId":"no-debugger",
                "message":"Unexpected 'debugger' statement.", "severity":2,
                "line":line, "column":1, "endLine":line, "endColumn":10,
                "range":[start,start+9]
            })
        );
    }
}

#[test]
fn automatic_fix_preserves_disabled_occurrences_and_fixes_enabled_occurrences() {
    let source = "let value: string | null = null;\n// rslint-disable-next-line @typescript-eslint/no-extra-non-null-assertion\nvalue!!;\nvalue!!;";
    let mut input: Value = serde_json::from_str(&request(&[("fix.ts", source)])).unwrap();
    input["rules"] = json!([{"name":"@typescript-eslint/no-extra-non-null-assertion"}]);
    let observed: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    let messages = observed["files"][0]["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 1);
    assert_eq!(messages[0]["line"], 4);
    assert_eq!(
        messages[0]["ruleId"],
        "@typescript-eslint/no-extra-non-null-assertion"
    );
    assert!(messages[0]["fix"].is_object());
    input["fix"] = json!(true);
    let fixed: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(fixed["files"][0]["messages"], json!([]));
    assert_eq!(
        fixed["files"][0]["output"],
        source.replacen("value!!;\nvalue!!;", "value!!;\nvalue!;", 1)
    );
    assert_eq!(
        fixed["files"][0]["appliedRules"],
        json!(["@typescript-eslint/no-extra-non-null-assertion"])
    );
    input["files"][0]["text"] = json!("/* eslint-disable */ value!!;");
    let suppressed: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(suppressed["files"][0]["messages"], json!([]));
    assert!(suppressed["files"][0].get("output").is_none());
}

#[test]
fn comments_between_tokens_and_inside_template_expressions_are_seen() {
    for source in [
        "const /* eslint no-debugger: off */ value = 1;",
        "const value = /* eslint no-debugger: off */ 1;",
        "const value = (1 /* eslint no-debugger: off */ + 2);",
        "const value = `raw ${1 /* eslint no-debugger: off */ + 2} tail`;",
        "const value = /[/*]/; /* eslint no-debugger: off */ value.test('x');",
        "const value = 1;\u{2028}/* eslint no-debugger: off */\u{2029}",
        "const value = 1;          /* eslint no-debugger: off */",
    ] {
        rejected(
            "tokens.ts",
            source,
            "/* eslint no-debugger: off",
            "InlineConfig",
            "eslint no-debugger: off",
        );
    }
}

#[test]
fn jsx_expression_attribute_type_argument_and_outer_trailing_comments_are_seen() {
    for source in [
        "const value = <>{/* eslint no-debugger: off */ 1}</>;",
        "const value = <A prop={1 /* eslint no-debugger: off */}>text</A>;",
        "const value = <A<T /* eslint no-debugger: off */> prop={1}>text</A>;",
        "const value = <><A /></>; /* eslint no-debugger: off */",
        "const value = <A></A>; /* eslint no-debugger: off */",
    ] {
        rejected(
            "jsx.tsx",
            source,
            "/* eslint no-debugger: off",
            "InlineConfig",
            "eslint no-debugger: off",
        );
    }
}

#[test]
fn exported_globals_do_not_change_default_module_diagnostics() {
    let result: Value = serde_json::from_str(
        &lint_json(&request(&[(
            "module.ts",
            "/*exported value*/ var value=1; debugger;",
        )]))
        .unwrap(),
    )
    .unwrap();
    assert_eq!(result["files"][0]["messages"].as_array().unwrap().len(), 1);
    assert_eq!(result["files"][0]["messages"][0]["ruleId"], "no-debugger");
}
