//! The direct engine must reject unsupported directives without mistaking
//! strings, regular expressions, template text, or JSX text for comments.

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

fn rejected(path: &str, source: &str, marker: &str) {
    let error = lint_json(&request(&[(path, source)])).unwrap_err();
    assert!(
        error.starts_with("unsupported lint directive in "),
        "{source:?}: {error}"
    );
    let offset = source.find(marker).unwrap();
    assert!(
        error.contains(&format!("/rslint-comment-tests/{path} at byte {offset}:")),
        "{source:?}: {error}"
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
fn all_directive_forms_are_explicitly_rejected_in_real_comments() {
    for (source, marker) in [
        ("// eslint-disable-next-line no-debugger\ndebugger;", "//"),
        ("debugger; // rslint-disable-line no-debugger", "//"),
        ("/* eslint-disable */ debugger;", "/*"),
        ("/* rslint-enable */ const value = 1;", "/*"),
        ("/*eslint no-debugger: off*/ debugger;", "/*"),
        ("/*global external:readonly*/ external;", "/*"),
        ("/*globals external:false*/ external;", "/*"),
        ("/*exported value*/ var value=1;", "/*"),
        ("#!/usr/bin/env node\n// rslint-disable", "//"),
        ("/* ordinary */\n// eslint-disable", "//"),
        ("/*\u{feff}\u{00a0}eslint-disable */", "/*"),
    ] {
        rejected("directive.ts", source, marker);
    }
}

#[test]
fn comments_between_tokens_and_inside_template_expressions_are_seen() {
    for source in [
        "const /* eslint-disable */ value = 1;",
        "const value = /* rslint-disable */ 1;",
        "const value = (1 /* eslint-disable */ + 2);",
        "const value = `raw ${1 /* eslint-disable */ + 2} tail`;",
        "const value = /[/*]/; /* eslint-disable */ value.test('x');",
        "const value = 1;\u{2028}/* eslint-disable */\u{2029}",
        "const value = 1;          /* eslint-disable */",
    ] {
        let marker = if source.contains("/* rslint-disable") {
            "/* rslint-disable"
        } else {
            "/* eslint-disable"
        };
        rejected("tokens.ts", source, marker);
    }
}

#[test]
fn jsx_expression_attribute_type_argument_and_outer_trailing_comments_are_seen() {
    for source in [
        "const value = <>{/* eslint-disable */ 1}</>;",
        "const value = <A prop={1 /* eslint-disable */}>text</A>;",
        "const value = <A<T /* eslint-disable */> prop={1}>text</A>;",
        "const value = <><A /></>; /* eslint-disable */",
        "const value = <A></A>; /* eslint-disable */",
    ] {
        rejected("jsx.tsx", source, "/* eslint-disable");
    }
}
