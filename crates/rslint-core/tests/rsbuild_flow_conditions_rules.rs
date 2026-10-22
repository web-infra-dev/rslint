//! Frozen default-option tests plus precise control-flow diagnostic contracts.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

fn lint(rule: &str, text: &str, options: Value, fix: bool) -> Value {
    let request = json!({
        "schemaVersion":1,"cwd":"/rslint-flow-conditions","fix":fix,
        "files":[{"path":"input.ts","text":text,"rules":[{"name":rule}],"policy":{"kind":"explicit","configFiles":["/rslint-flow-conditions/tsconfig.json"]}}],
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
        "../../../scripts/tsrs/oracle-cases/rsbuild-flow-conditions.json"
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
                    .all(|value| value.get("line").is_some() && value.get("column").is_some())
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
fn no_constant_binary_expression_frozen_cases() {
    corpus("no-constant-binary-expression");
}
#[test]
fn no_constant_condition_frozen_cases() {
    corpus("no-constant-condition");
}
#[test]
fn no_dupe_else_if_frozen_cases() {
    corpus("no-dupe-else-if");
}
#[test]
fn no_fallthrough_frozen_cases() {
    corpus("no-fallthrough");
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
fn constant_comparison_ranges_preserve_unicode() {
    marked(
        "no-constant-binary-expression",
        "/* 😀 */ const result = «[] === value»;",
        "alwaysNew",
        "Unexpected comparison to newly constructed object. These two values can never be equal.",
    );
    marked(
        "no-constant-binary-expression",
        "const result = «(value + 1) ?? fallback»;",
        "constantShortCircuit",
        "Unexpected constant ?? on the left-hand side of a `??` expression.",
    );
    marked(
        "no-constant-binary-expression",
        "if («undefined === 42») {}",
        "constantBinaryOperand",
        "Unexpected constant binary expression. Comparisons will always evaluate the same.",
    );
}
#[test]
fn constant_conditions_preserve_default_loop_and_yield_policy() {
    marked(
        "no-constant-condition",
        "/* 😀 */ if («(true)») {} while (true) {}",
        "unexpected",
        "Unexpected constant condition.",
    );
    marked(
        "no-constant-condition",
        "for (; «true»;) { break; }",
        "unexpected",
        "Unexpected constant condition.",
    );
    let messages = lint(
        "no-constant-condition",
        "function* f() { while (1) { yield 1; } }",
        json!({"target":"esnext","types":[]}),
        false,
    );
    assert_eq!(messages, json!([]));
}
#[test]
fn else_if_coverage_is_cumulative_and_tokens_ignore_comments() {
    marked(
        "no-dupe-else-if",
        "if (a) {} else if (b) {} else if («b || a») {}",
        "unexpected",
        "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
    );
    marked(
        "no-dupe-else-if",
        "/* 😀 */ if (f(a)) {} else if («f(/* comment */ a)») {}",
        "unexpected",
        "This branch can never execute. Its condition is a duplicate or covered by previous conditions in the if-else-if chain.",
    );
}
#[test]
fn switch_case_ranges_and_unicode_fallthrough_comments() {
    marked(
        "no-fallthrough",
        "/* 😀 */ switch (x) { case 1: first(); «case 2: second(); break;» }",
        "case",
        "Expected a 'break' statement before 'case'.",
    );
    marked(
        "no-fallthrough",
        "switch (x) { case 1: first(); «default: second();» }",
        "default",
        "Expected a 'break' statement before 'default'.",
    );
    for comment in [
        "fallthrough",
        "FALLS THROUGH",
        "fallſ\u{a0}through",
        "falls\u{feff}through",
    ] {
        let text = format!("switch (x) {{ case 1: first(); /* {comment} */ case 2: break; }}");
        assert_eq!(
            lint("no-fallthrough", &text, json!({"types":[]}), true),
            json!([]),
            "{comment}"
        );
    }
}
