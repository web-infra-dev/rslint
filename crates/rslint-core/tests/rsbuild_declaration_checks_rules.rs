//! Default-configuration contracts and frozen-oracle inputs for declaration checks.
use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, text: &str, fix: bool) -> Value {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-declaration-checks","files":[{"path":"case.ts","text":text}],"rules":[{"name":rule}],"fix":fix});
    let response: Value = serde_json::from_str(
        &lint_json(&request.to_string()).unwrap_or_else(|error| panic!("{rule}: {error}\n{text}")),
    )
    .unwrap();
    assert_eq!(response["files"].as_array().unwrap().len(), 1);
    assert_eq!(
        response["files"][0]["appliedRules"],
        json!([rule]),
        "{rule}: {text}\n{response}"
    );
    assert!(
        response["files"][0].get("output").is_none(),
        "none of these rules automatically fixes source"
    );
    response["files"][0]["messages"].clone()
}
#[test]
fn frozen_default_positive_and_negative_cases() {
    let suites: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-declaration-checks.json"
    ))
    .unwrap();
    for suite in suites.as_array().unwrap() {
        let rule = suite["rule"].as_str().unwrap();
        for (group, empty) in [("valid", true), ("invalid", false)] {
            let cases = suite[group].as_array().unwrap();
            assert!(!cases.is_empty());
            for case in cases {
                let text = case.as_str().or_else(|| case["text"].as_str()).unwrap();
                let messages = lint(rule, text, false);
                assert_eq!(
                    messages.as_array().unwrap().is_empty(),
                    empty,
                    "{rule}/{group}: {text}\n{messages}"
                );
                for message in messages.as_array().unwrap() {
                    assert_eq!(message["ruleId"], rule);
                    assert_eq!(message["severity"], 2);
                    assert!(message.get("fix").is_none());
                }
            }
        }
    }
}
fn unmark(marked: &str) -> (String, Vec<[usize; 2]>) {
    let mut text = String::new();
    let mut begin = None;
    let mut ranges = Vec::new();
    for character in marked.chars() {
        match character {
            '«' => assert!(begin.replace(text.len()).is_none()),
            '»' => ranges.push([begin.take().unwrap(), text.len()]),
            _ => text.push(character),
        }
    }
    assert!(begin.is_none());
    (text, ranges)
}
fn utf16(text: &str, byte: usize) -> usize {
    text[..byte].encode_utf16().count()
}
fn marked(rule: &str, source: &str, id: &str, description: &str) -> (String, Vec<Value>) {
    let (text, ranges) = unmark(source);
    let messages = lint(rule, &text, true);
    let messages = messages.as_array().unwrap();
    assert_eq!(messages.len(), ranges.len(), "{text}: {messages:?}");
    for (message, [begin, end]) in messages.iter().zip(ranges) {
        assert_eq!(message["messageId"], id);
        assert_eq!(message["message"], description);
        assert_eq!(
            message["range"],
            json!([utf16(&text, begin), utf16(&text, end)])
        );
        for (offset, line, column) in [(begin, "line", "column"), (end, "endLine", "endColumn")] {
            let prefix = &text[..offset];
            assert_eq!(message[line], prefix.split('\n').count());
            assert_eq!(
                message[column],
                prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1
            );
        }
    }
    (text, messages.to_vec())
}
#[test]
fn setter_returns_own_function_bodies_and_expression_arrows() {
    marked(
        "no-setter-return",
        "/* 😀 */ class C { set value(v) { function inner() { return 0; } «return v;» } }",
        "returnsValue",
        "Setter cannot return a value.",
    );
    marked(
        "no-setter-return",
        "Object.defineProperty(target, 'x', {set: value => «(value + 1)»});",
        "returnsValue",
        "Setter cannot return a value.",
    );
    marked(
        "no-setter-return",
        "Object.defineProperties(target, {x: { ['set'](value) { «return value;» } }});",
        "returnsValue",
        "Setter cannot return a value.",
    );
}
#[test]
fn prototype_suggestion_preserves_call_parentheses_and_sequence_receiver() {
    let (text, messages) = marked(
        "no-prototype-builtins",
        "/* 😀 */ ((first, second).«hasOwnProperty») /* call */ ('x');",
        "prototypeBuildIn",
        "Do not access Object.prototype method 'hasOwnProperty' from target object.",
    );
    let begin = text.find("(first, second)").unwrap();
    let open = text.find("('x')").unwrap();
    let between = &text[text.find("hasOwnProperty").unwrap() + "hasOwnProperty".len()..open + 1];
    assert_eq!(
        messages[0]["suggestions"],
        json!([{"messageId":"callObjectPrototype","desc":"Call Object.prototype.hasOwnProperty explicitly.","fix":{"range":[utf16(&text,begin),utf16(&text,open+1)],"text":format!("Object.prototype.hasOwnProperty.call{between}(first, second), ")}}])
    );
    for text in [
        "foo?.hasOwnProperty('x')",
        "foo.hasOwnProperty?.('x')",
        "(foo?.hasOwnProperty)('x')",
        "function f(Object) { foo.hasOwnProperty('x') }",
    ] {
        let messages = lint("no-prototype-builtins", text, false);
        assert_eq!(messages.as_array().unwrap().len(), 1);
        assert!(messages[0].get("suggestions").is_none());
    }
}
#[test]
fn conditional_assignments_preserve_default_parenthesis_depth() {
    marked(
        "no-cond-assign",
        "/* 😀 */ if («value = other») {} while ((value = other)) {}",
        "missing",
        "Expected a conditional expression and instead saw an assignment.",
    );
    marked(
        "no-cond-assign",
        "const result = («value = other») ? yes : no;",
        "missing",
        "Expected a conditional expression and instead saw an assignment.",
    );
    assert!(
        lint(
            "no-cond-assign",
            "const result = ((value = other)) ? yes : no;",
            false
        )
        .as_array()
        .unwrap()
        .is_empty()
    );
}
#[test]
fn duplicate_property_names_canonicalize_js_numbers_bigints_and_regex_flags() {
    for (rule, source, name) in [
        ("no-dupe-keys", "({1e21: 0, «'1e+21'»: 1})", "1e+21"),
        (
            "no-dupe-keys",
            "({[0x10000000000000000n]: 0, «'18446744073709551616'»: 1})",
            "18446744073709551616",
        ),
        (
            "no-dupe-keys",
            "({[/foo/mi]: 0, [«/foo/im»]: 1})",
            "/foo/im",
        ),
        (
            "no-dupe-class-members",
            "class C { [1e-7]() {} ['0.0000001']() {} [«'1e-7'»]() {} }",
            "1e-7",
        ),
    ] {
        let description = if rule == "no-dupe-keys" {
            format!("Duplicate key '{name}'.")
        } else {
            format!("Duplicate name '{name}'.")
        };
        marked(rule, source, "unexpected", &description);
    }
}
#[test]
fn computed_key_report_parenthesis_boundary_matches_each_frozen_rule() {
    marked(
        "no-dupe-keys",
        "/* 😀 */ ({ x: 1, [(«'x'»)]: 2 })",
        "unexpected",
        "Duplicate key 'x'.",
    );
    marked(
        "no-dupe-class-members",
        "/* 😀 */ class C { x() {} [«('x')»]() {} }",
        "unexpected",
        "Duplicate name 'x'.",
    );
    marked(
        "no-dupe-class-members",
        "class C { get x() {} set x(v) {} «x»=1; static x=1; }",
        "unexpected",
        "Duplicate name 'x'.",
    );
}
#[test]
fn restricted_binding_reports_are_deduplicated_and_keep_runtime_scope() {
    marked(
        "no-shadow-restricted-names",
        "/* 😀 */ function «NaN»(«NaN») { try {} catch («NaN») {} }",
        "shadowingRestrictedName",
        "Shadowing of global property 'NaN'.",
    );
    marked(
        "no-shadow-restricted-names",
        "var «undefined»; ({undefined}=value);",
        "shadowingRestrictedName",
        "Shadowing of global property 'undefined'.",
    );
    assert!(
        lint(
            "no-shadow-restricted-names",
            "var undefined; function f() { let undefined; } type F = (undefined: string) => void;",
            false
        )
        .as_array()
        .unwrap()
        .is_empty()
    );
}
#[test]
fn nondefault_modes_fail_explicitly() {
    for (rule, options) in [
        ("no-cond-assign", json!(["always"])),
        (
            "no-shadow-restricted-names",
            json!([{"reportGlobalThis":false}]),
        ),
    ] {
        let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-declaration-checks","files":[{"path":"case.ts","text":""}],"rules":[{"name":rule,"options":options}]});
        assert!(
            lint_json(&request.to_string())
                .unwrap_err()
                .contains("does not accept options")
        );
    }
}

#[test]
fn directives_use_the_public_rule_name_after_rule_tester_alias_conversion() {
    assert!(
        lint(
            "no-cond-assign",
            "// eslint-disable-next-line no-cond-assign\nif (x = y) {}",
            false
        )
        .as_array()
        .unwrap()
        .is_empty()
    );
    assert_eq!(
        lint(
            "no-cond-assign",
            "// eslint-disable-next-line test\nif (x = y) {}",
            false
        )
        .as_array()
        .unwrap()
        .len(),
        1
    );
}
