//! Frozen default binding-rule contracts; complete diagnostic/fix oracle inputs
//! live beside the other direct Rust+JS migration fixtures.
use rslint_core::lint_json;
use serde_json::{Value, json};
fn lint(rule: &str, text: &str, extension: &str, fix: bool) -> Value {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-binding-checks","files":[{"path":format!("case{extension}"),"text":text}],"rules":[{"name":rule}],"fix":fix});
    let result: Value = serde_json::from_str(
        &lint_json(&request.to_string()).unwrap_or_else(|error| panic!("{rule}: {error}\n{text}")),
    )
    .unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    assert_eq!(result["files"][0]["appliedRules"], json!([rule]));
    for message in result["files"][0]["messages"].as_array().unwrap() {
        assert_eq!(message["ruleId"], rule);
        assert_eq!(message["severity"], 2);
        assert!(message.get("suggestions").is_none());
        if !matches!(rule, "no-var" | "prefer-const") {
            assert!(message.get("fix").is_none());
        }
    }
    result["files"][0].clone()
}
fn corpus(rule: &str) {
    let cases: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-binding-checks.json"
    ))
    .unwrap();
    let group = cases
        .as_array()
        .unwrap()
        .iter()
        .find(|group| group["rule"] == rule)
        .unwrap();
    for category in ["valid", "invalid"] {
        assert!(!group[category].as_array().unwrap().is_empty());
        for entry in group[category].as_array().unwrap() {
            let text = entry["text"].as_str().unwrap();
            let extension = entry["extension"].as_str().unwrap_or(".ts");
            let file = lint(rule, text, extension, false);
            let messages = file["messages"].as_array().unwrap();
            let mut expected = entry["expected"].as_array().cloned().unwrap_or_default();
            // Public API sorts by source position; Go RuleTester records listener order.
            expected.sort_by_key(|message| {
                (
                    message["line"].as_u64().unwrap_or(1),
                    message["column"].as_u64().unwrap_or(1),
                )
            });
            assert_eq!(
                messages.len(),
                expected.len(),
                "{rule}/{category}: {}\n{text}\n{messages:?}",
                entry["origin"]
            );
            for (actual, expected) in messages.iter().zip(expected) {
                for (key, value) in expected.as_object().unwrap() {
                    assert_eq!(
                        &actual[key], value,
                        "{rule}: {}\n{text}\n{actual}",
                        entry["origin"]
                    );
                }
            }
            if let Some(output) = entry.get("outputSteps") {
                let fixed = lint(rule, text, extension, true);
                if output.is_null() {
                    assert!(fixed.get("output").is_none(), "{text}: {fixed}");
                } else {
                    assert_eq!(
                        fixed["output"],
                        *output.as_array().unwrap().last().unwrap(),
                        "{text}: {fixed}"
                    );
                }
                let after = fixed["output"].as_str().unwrap_or(text);
                assert_eq!(
                    fixed["messages"],
                    lint(rule, after, extension, false)["messages"],
                    "post-fix diagnostics must be fresh"
                );
            }
        }
    }
}
#[test]
fn frozen_no_var() {
    corpus("no-var");
}
#[test]
fn frozen_prefer_const() {
    corpus("prefer-const");
}
#[test]
fn frozen_no_global_assign() {
    corpus("no-global-assign");
}
#[test]
fn frozen_no_new_native_nonconstructor() {
    corpus("no-new-native-nonconstructor");
}

fn marked(rule: &str, source: &str, id: &str, messages: &[&str]) -> (String, Value) {
    let mut text = String::new();
    let mut begin = None;
    let mut ranges = Vec::new();
    for ch in source.chars() {
        match ch {
            '«' => assert!(begin.replace(text.len()).is_none()),
            '»' => ranges.push([begin.take().unwrap(), text.len()]),
            _ => text.push(ch),
        }
    }
    assert!(begin.is_none());
    let file = lint(rule, &text, ".ts", false);
    let actual = file["messages"].as_array().unwrap();
    assert_eq!(actual.len(), ranges.len(), "{text}: {file}");
    assert_eq!(messages.len(), ranges.len());
    for ((diagnostic, [begin, end]), message) in actual.iter().zip(ranges).zip(messages) {
        assert_eq!(diagnostic["messageId"], id);
        assert_eq!(diagnostic["message"], *message);
        assert_eq!(
            diagnostic["range"],
            json!([
                text[..begin].encode_utf16().count(),
                text[..end].encode_utf16().count()
            ])
        );
        for (offset, line, column) in [(begin, "line", "column"), (end, "endLine", "endColumn")] {
            let prefix = &text[..offset];
            assert_eq!(diagnostic[line], prefix.split('\n').count());
            assert_eq!(
                diagnostic[column],
                prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1
            );
        }
    }
    (text, file)
}
#[test]
fn var_fixes_only_when_scope_and_initialization_are_preserved() {
    let (text, file) = marked(
        "no-var",
        "/* 😀 */ «var value = 1;»",
        "unexpectedVar",
        &["Unexpected var, use let or const instead."],
    );
    assert_eq!(
        file["messages"][0]["fix"],
        json!({"range":[9,12],"text":"let"})
    );
    assert_eq!(
        lint("no-var", &text, ".ts", true)["output"],
        "/* 😀 */ let value = 1;"
    );
    for text in [
        "{ var x = 0; } consume(x);",
        "var value = value;",
        "for (;;) { var x; }",
        "export var x = 0;",
        "var x = read(); function read(){return x;}",
    ] {
        let file = lint("no-var", text, ".ts", true);
        assert!(file.get("output").is_none(), "{text}: {file}");
        assert!(!file["messages"].as_array().unwrap().is_empty());
    }
    assert!(
        lint("no-var", "declare global { var x: number; }", ".ts", false)["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
}
#[test]
fn const_preserves_shared_fixes_and_defers_to_write_binding_identity() {
    let (text, file) = marked(
        "prefer-const",
        "let «a» = 1, «b» = 2;",
        "useConst",
        &[
            "'a' is never reassigned. Use 'const' instead.",
            "'b' is never reassigned. Use 'const' instead.",
        ],
    );
    for msg in file["messages"].as_array().unwrap() {
        assert_eq!(msg["fix"], json!({"range":[0,3],"text":"const"}));
    }
    assert_eq!(
        lint("prefer-const", &text, ".ts", true)["output"],
        "const a = 1, b = 2;"
    );
    let (text, file) = marked(
        "prefer-const",
        "let a; «a» = 2;",
        "useConst",
        &["'a' is never reassigned. Use 'const' instead."],
    );
    assert!(file["messages"][0].get("fix").is_none());
    assert!(
        lint("prefer-const", &text, ".ts", true)
            .get("output")
            .is_none()
    );
    assert!(
        lint(
            "prefer-const",
            "let a = 1; (()=>{a = 2;})(); for(let i=0;i<2;i++){}",
            ".ts",
            false
        )["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
}
#[test]
fn globals_are_resolved_as_values_and_native_names_keep_source_ranges() {
    marked(
        "no-global-assign",
        "interface Object {} type Array = unknown; «Object»=1; «Array»=2;",
        "globalShouldNotBeModified",
        &[
            "Read-only global 'Object' should not be modified.",
            "Read-only global 'Array' should not be modified.",
        ],
    );
    assert!(
        lint(
            "no-global-assign",
            "function local(Object){Object=2;} {const Array=[]; Array.push(1);}",
            ".ts",
            false
        )["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
    marked(
        "no-new-native-nonconstructor",
        "new ((«Symbol» as any))(); new «BigInt»(1);",
        "noNewNonconstructor",
        &[
            "`Symbol` cannot be called as a constructor.",
            "`BigInt` cannot be called as a constructor.",
        ],
    );
    assert!(
        lint(
            "no-new-native-nonconstructor",
            "function f(Symbol:any){new Symbol();} new Array();",
            ".ts",
            false
        )["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
}

#[test]
fn frozen_no_obj_calls() {
    corpus("no-obj-calls");
}
#[test]
fn object_calls_follow_alias_cycles_with_default_global_availability() {
    marked(
        "no-obj-calls",
        "let a = JSON; let b = a; a = b; «a()»;",
        "unexpectedRefCall",
        &["'a' is reference to 'JSON', which is not a function."],
    );
    marked(
        "no-obj-calls",
        "Math = JSON; «Math()»;",
        "unexpectedRefCall",
        &["'Math' is reference to 'JSON', which is not a function."],
    );
    marked(
        "no-obj-calls",
        "const { Math: value } = globalThis; «new value()»;",
        "unexpectedRefCall",
        &["'value' is reference to 'Math', which is not a function."],
    );
    for source in [
        "Math = custom; Math();",
        "const { ...value } = globalThis; value();",
        "const key = 'Math'; globalThis[key]();",
        "globalThis[['Math']]();",
        "window.Math(); self.JSON(); global.Reflect();",
        "namespace Math {} Math();",
    ] {
        assert!(
            lint("no-obj-calls", source, ".ts", false)["messages"]
                .as_array()
                .unwrap()
                .is_empty(),
            "{source}"
        );
    }
    assert_eq!(
        lint("no-obj-calls", "global.Math()", ".cjs", false)["messages"]
            .as_array()
            .unwrap()
            .len(),
        1
    );
}
#[test]
fn object_calls_fold_global_keys_without_resolving_local_initializers() {
    for expression in [
        "String.fromCharCode(77,97,116,104)",
        "Array.of('Ma','th').join('')",
        "Symbol.keyFor(Symbol.for('Math'))",
        "({__proto__:{key:'Math'}}).key",
        "(null?.anything ?? 'Math')",
        "(unknown(),'Math')",
        "Math.max(1,2)===2 ? 'Math' : 'other'",
    ] {
        marked(
            "no-obj-calls",
            &format!("«globalThis[{expression}]()»"),
            "unexpectedRefCall",
            &["'null' is reference to 'Math', which is not a function."],
        );
    }
    assert!(
        lint(
            "no-obj-calls",
            "function f(String:any){globalThis[String.fromCharCode(77,97,116,104)]();}",
            ".ts",
            false
        )["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
}
#[test]
fn object_call_keys_keep_bigint_regex_and_numeric_boundaries() {
    for expression in [
        "9007199254740993n > 9007199254740992 ? 'Math' : 'other'",
        "9007199254740993n == '9007199254740993' ? 'Math' : 'other'",
        "(2n ** 100n) === 1267650600228229401496703205376n ? 'Math' : 'other'",
        "(-8n >> 2n) === -2n ? 'Math' : 'other'",
        "String(123n) === '123' ? 'Math' : 'other'",
        "typeof 1n === 'bigint' ? 'Math' : 'other'",
        "(/Math/).source",
        "(/x/ig).flags === 'gi' ? 'Math' : 'other'",
        "Math.f16round(1.00048828125) === 1 ? 'Math' : 'other'",
        "1 / Math.round(-0.5) === -Infinity ? 'Math' : 'other'",
        "(3 << 2) === 12 ? 'Math' : 'other'",
        "(null == undefined) ? 'Math' : 'other'",
    ] {
        marked(
            "no-obj-calls",
            &format!("«globalThis[{expression}]()»"),
            "unexpectedRefCall",
            &["'null' is reference to 'Math', which is not a function."],
        );
    }
    for source in [
        "globalThis[(1n + 1) ? 'Math' : 'Math']();",
        "globalThis[(1n >>> 1n) ? 'Math' : 'Math']();",
        "globalThis[(2n ** -1n) ? 'Math' : 'Math']();",
        "globalThis[(Symbol.for('Math') + '')]();",
    ] {
        assert!(
            lint("no-obj-calls", source, ".ts", false)["messages"]
                .as_array()
                .unwrap()
                .is_empty(),
            "{source}"
        );
    }
}
