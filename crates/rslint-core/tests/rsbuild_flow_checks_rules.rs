//! Frozen default-option tests plus precise control-flow diagnostic contracts.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

fn lint(rule: &str, text: &str, options: Value, fix: bool) -> Value {
    let request = json!({
        "schemaVersion":1,"cwd":"/rslint-flow-checks","fix":fix,
        "files":[{"path":"input.ts","text":text,"rules":[{"name":rule}],"policy":{"kind":"explicit","configFiles":["/rslint-flow-checks/tsconfig.json"]}}],
        "sources":[{"path":"tsconfig.json","text":json!({"compilerOptions":options,"files":["input.ts"]}).to_string()}]
    });
    let result: Value = serde_json::from_str(
        &lint_projects_json(&request.to_string())
            .unwrap_or_else(|error| panic!("{rule}: {error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let file = &result["files"][0];
    assert_eq!(file["appliedRules"], json!([rule]), "{text}: {result}");
    assert!(
        file.get("output").is_none(),
        "flow rules must not modify source: {result}"
    );
    file["messages"].clone()
}
fn corpus(rule: &str) {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-flow-checks.json"
    ))
    .unwrap();
    let group = groups
        .as_array()
        .unwrap()
        .iter()
        .find(|group| group["rule"] == rule)
        .unwrap();
    for (name, valid) in [("valid", true), ("invalid", false)] {
        let cases = group[name].as_array().unwrap();
        assert!(!cases.is_empty());
        for (index, case) in cases.iter().enumerate() {
            let text = case["text"].as_str().unwrap();
            let messages = lint(rule, text, group["compilerOptions"].clone(), false);
            let messages = messages.as_array().unwrap();
            assert_eq!(
                messages.is_empty(),
                valid,
                "{rule}/{name}/{index}: {text}\n{messages:?}"
            );
            if let Some(expected) = case.get("expected").and_then(Value::as_array) {
                assert_eq!(
                    messages.len(),
                    expected.len(),
                    "{rule}/{name}/{index}: {text}\n{messages:?}"
                );
                // Public diagnostics are sorted by range, unlike a rule tester's
                // callback-order assertions (getter bare returns precede heads).
                let mut expected = expected.clone();
                if expected
                    .iter()
                    .all(|value| value.get("line").is_some())
                {
                    expected
                        .sort_by_key(|value| (value["line"].as_u64(), value["column"].as_u64()));
                }
                for (actual, expected) in messages.iter().zip(expected) {
                    for key in [
                        "messageId",
                        "message",
                        "line",
                        "column",
                        "endLine",
                        "endColumn",
                    ] {
                        if let Some(value) = expected.get(key) {
                            assert_eq!(
                                &actual[key], value,
                                "{rule}/{name}/{index}/{key}: {text}\n{messages:?}"
                            );
                        }
                    }
                }
            }
            for message in messages {
                assert_eq!(message["ruleId"], rule);
                assert_eq!(message["severity"], 2);
                assert!(message.get("fix").is_none());
                assert!(message.get("suggestions").is_none());
            }
        }
    }
}
#[test]
fn constructor_super_frozen_cases() {
    corpus("constructor-super");
}
#[test]
fn no_this_before_super_frozen_cases() {
    corpus("no-this-before-super");
}
#[test]
fn getter_return_frozen_cases() {
    corpus("getter-return");
}
#[test]
fn no_unreachable_frozen_cases() {
    corpus("no-unreachable");
}
fn marked(rule: &str, source: &str, id: &str, description: &str) {
    let mut text = String::new();
    let mut start = None;
    let mut ranges = Vec::new();
    for c in source.chars() {
        match c {
            '«' => assert!(start.replace(text.len()).is_none()),
            '»' => ranges.push([start.take().unwrap(), text.len()]),
            _ => text.push(c),
        }
    }
    assert!(start.is_none());
    let messages = lint(
        rule,
        &text,
        json!({"strict":true,"target":"esnext","module":"commonjs","types":[]}),
        true,
    );
    let messages = messages.as_array().unwrap();
    assert_eq!(messages.len(), ranges.len(), "{text}: {messages:?}");
    for (message, [start, end]) in messages.iter().zip(ranges) {
        assert_eq!(message["messageId"], id);
        assert_eq!(message["message"], description);
        assert_eq!(
            message["range"],
            json!([
                text[..start].encode_utf16().count(),
                text[..end].encode_utf16().count()
            ])
        );
        for (offset, line, column) in [(start, "line", "column"), (end, "endLine", "endColumn")] {
            let prefix = &text[..offset];
            assert_eq!(message[line], prefix.split('\n').count());
            assert_eq!(
                message[column],
                prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1
            );
        }
    }
}
#[test]
fn constructor_duplicate_and_missing_ranges() {
    marked(
        "constructor-super",
        "/* 😀 */ class C extends B { constructor() { super(); «super»(); } }",
        "duplicate",
        "Unexpected duplicate 'super()' call.",
    );
    marked(
        "constructor-super",
        "class C extends B { «constructor() { if (ok) super(); }» }",
        "missingSome",
        "Expected to call 'super()' in some paths.",
    );
    marked(
        "constructor-super",
        "class C extends null { constructor() { «super»(); } }",
        "badSuper",
        "Unexpected 'super()' call because this class does not extend a valid constructor.",
    );
}
#[test]
fn this_default_arguments_and_finally_ranges() {
    marked(
        "no-this-before-super",
        "class C extends B { constructor() { switch (super()) {} this.value = 1; } }",
        "noBeforeSuper",
        "'this' is not allowed before 'super()'.",
    );
    marked(
        "no-this-before-super",
        "/* 😀 */ class C extends B { constructor(x = «this».x) { super(); } }",
        "noBeforeSuper",
        "'this' is not allowed before 'super()'.",
    );
    marked(
        "no-this-before-super",
        "class C extends B { constructor() { try { super(); } finally { «this».value = 1; } } }",
        "noBeforeSuper",
        "'this' is not allowed before 'super()'.",
    );
}
#[test]
fn getter_function_heads_and_bare_returns() {
    for source in [
        "let Object; Object.defineProperty(target, 'value', { get() {} });",
        "function f() { Object.defineProperty(target, 'value', { get() {} }); var Object; }",
        "/* globals Object:off */ Object.defineProperty(target, 'value', { get() {} });",
        "/* globals Object:readonly */ function f(Object) { Object.defineProperty(target, 'value', { get() {} }); }",
    ] {
        marked(
            "getter-return",
            source,
            "expected",
            "Expected to return a value in method 'get'.",
        );
    }
    marked(
        "getter-return",
        "interface Object {} Object.defineProperty(target, 'value', { «get»() {} });",
        "expected",
        "Expected to return a value in method 'get'.",
    );
    marked(
        "getter-return",
        "/* 😀 */ class C { «static get value»() {} }",
        "expected",
        "Expected to return a value in static getter 'value'.",
    );
    marked(
        "getter-return",
        "Object.defineProperty(target, 'value', { «get: function »() {} });",
        "expected",
        "Expected to return a value in method 'get'.",
    );
    marked(
        "getter-return",
        "class C { get value() { «return;» } }",
        "expected",
        "Expected to return a value in getter 'value'.",
    );
}
#[test]
fn unreachable_groups_skip_hoisted_statements() {
    marked(
        "no-unreachable",
        "class C { static { if (true) throw new Error(); after(); } }",
        "unreachableCode",
        "Unreachable code.",
    );
    marked(
        "no-unreachable",
        "/* 😀 */ function f() { return; «a(); b();» var hoisted; «c();» }",
        "unreachableCode",
        "Unreachable code.",
    );
    marked(
        "no-unreachable",
        "function f() { if (true) return; live(); return; «dead();» }",
        "unreachableCode",
        "Unreachable code.",
    );
}
