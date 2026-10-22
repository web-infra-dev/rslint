//! Frozen rsbuild's unsafe access rules: actual checker types, exact messages,
//! chain suppression, library identity, and strictness-dependent `this` advice.

use rslint_core::lint_json;
use serde_json::{Value, json};

const CALL: &str = "@typescript-eslint/no-unsafe-call";
const MEMBER: &str = "@typescript-eslint/no-unsafe-member-access";

fn lint(rule: &str, sources: &[(String, String)], options: Option<Value>) -> Vec<Value> {
    // Preserve each oracle input's script/module status and global scope. Each
    // file receives a separate program; unrelated fixtures must not merge globals.
    sources
        .iter()
        .enumerate()
        .map(|(index, (text, extension))| {
            let mut request = json!({
                "schemaVersion": 1, "cwd": "/rslint-rsbuild-unsafe-access-tests",
                "files": [{"path": format!("case-{index}{extension}"), "text": text}],
                "rules": [{"name": rule}],
            });
            if let Some(options) = &options {
                request["compilerOptions"] = options.clone();
            }
            let result: Value =
                serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
            let files = result["files"].as_array().unwrap();
            assert_eq!(files.len(), 1, "silent files remain admitted");
            assert_eq!(
                files[0]["filePath"],
                format!("/rslint-rsbuild-unsafe-access-tests/case-{index}{extension}")
            );
            assert_eq!(files[0]["appliedRules"], json!([rule]));
            files[0]["messages"].clone()
        })
        .collect()
}

fn one(rule: &str, source: &str, options: Option<Value>) -> Vec<Value> {
    lint(rule, &[(source.to_string(), ".ts".to_string())], options)[0]
        .as_array()
        .unwrap()
        .clone()
}

fn verify_cases(rule: &str) {
    let cases: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-unsafe-access.json"
    ))
    .unwrap();
    let group = cases
        .as_array()
        .unwrap()
        .iter()
        .find(|group| group["rule"] == rule)
        .unwrap();
    for category in ["valid", "invalid"] {
        let entries = group[category].as_array().unwrap();
        assert!(!entries.is_empty());
        for entry in entries {
            let (source, extension) = if let Some(text) = entry.as_str() {
                (text.to_string(), ".ts".to_string())
            } else {
                (
                    entry["text"].as_str().unwrap().to_string(),
                    entry["extension"].as_str().unwrap_or(".ts").to_string(),
                )
            };
            let options = entry.get("compilerOptions").map(|overrides| {
                let mut options = json!({"strict":true, "allowJs":true, "target":9, "module":99, "moduleResolution":100});
                options.as_object_mut().unwrap().extend(overrides.as_object().unwrap().clone());
                options
            });
            let messages = lint(rule, &[(source.clone(), extension)], options).remove(0);
            let messages = messages.as_array().unwrap();
            if category == "valid" {
                assert!(messages.is_empty(), "{rule}: {source}: {messages:?}");
            } else {
                assert!(!messages.is_empty(), "{rule}: {source}");
            }
            for diagnostic in messages {
                assert_eq!(diagnostic["ruleId"], rule);
                assert_eq!(diagnostic["severity"], 2);
                assert!(diagnostic["fix"].is_null());
                assert!(
                    diagnostic["suggestions"].is_null(),
                    "unsafe accesses have no automatic safe replacement"
                );
            }
        }
    }
}

#[test]
fn unsafe_call_handles_calls_construction_tags_function_subtypes_and_this() {
    verify_cases(CALL);
}

#[test]
fn unsafe_members_handle_error_types_computed_keys_chains_heritage_and_jsx() {
    verify_cases(MEMBER);
}

#[test]
fn call_message_variants_and_reporting_ranges_match_the_frozen_baseline() {
    let source = "declare const value: NotKnown; value(); new value(); value`text`;";
    let messages = one(CALL, source, None);
    assert_eq!(messages.len(), 3);
    for (message, id, text, range) in [
        (
            &messages[0],
            "unsafeCall",
            "Unsafe call of a(n) `error` type typed value.",
            [31, 36],
        ),
        (
            &messages[1],
            "unsafeNew",
            "Unsafe construction of a(n) `error` type typed value.",
            [40, 51],
        ),
        (
            &messages[2],
            "unsafeTemplateTag",
            "Unsafe use of a(n) `error` type typed template tag.",
            [53, 58],
        ),
    ] {
        assert_eq!(message["messageId"], id);
        assert_eq!(message["message"], text);
        assert_eq!(message["range"], json!(range));
        assert_eq!(message["line"], 1);
        assert_eq!(message["column"], range[0] + 1);
        assert_eq!(message["endLine"], 1);
        assert_eq!(message["endColumn"], range[1] + 1);
    }
}

#[test]
fn computed_receiver_and_key_report_independently_at_the_unwrapped_utf16_range() {
    let source = "declare const value: any; declare const key: any; value[(( /* 😀 */ key ))];";
    let messages = one(MEMBER, source, None);
    assert_eq!(messages.len(), 2);
    let start = source[..source.rfind("key").unwrap()]
        .encode_utf16()
        .count();
    assert_eq!(messages[0]["messageId"], "unsafeMemberExpression");
    assert_eq!(
        messages[0]["message"],
        "Unsafe member access [key] on an `any` value."
    );
    assert_eq!(messages[1]["messageId"], "unsafeComputedMemberAccess");
    assert_eq!(
        messages[1]["message"],
        "Computed name [key] resolves to an `any` value."
    );
    for message in messages {
        assert_eq!(message["range"], json!([start, start + 3]));
        assert_eq!(message["column"], start + 1);
        assert_eq!(message["endColumn"], start + 4);
    }
}

#[test]
fn call_boundaries_end_chain_suppression_and_preserve_source_order() {
    let source = "declare const value: any; value.method().result; (value.a).b; (value?.c).d;";
    let messages = one(MEMBER, source, None);
    assert_eq!(
        messages
            .iter()
            .map(|message| message["message"].as_str().unwrap())
            .collect::<Vec<_>>(),
        [
            "Unsafe member access .method on an `any` value.",
            "Unsafe member access .result on an `any` value.",
            "Unsafe member access .a on an `any` value.",
            "Unsafe member access .c on an `any` value.",
            "Unsafe member access .d on an `any` value.",
        ]
    );
}

#[test]
fn this_messages_respect_each_rules_distinct_strictness_behavior() {
    let advice = "\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.";
    for (annotation, member_id, member_text, call_type) in [
        (
            "any",
            "unsafeThisMemberExpression",
            "Unsafe member access .property on an `any` value. `this` is typed as `any`.",
            "`any`",
        ),
        (
            "NotKnown",
            "errorThisMemberExpression",
            "Unsafe member access .property. The type of `this` cannot be resolved.",
            "`error` type",
        ),
    ] {
        let source = format!("function read(this: {annotation}) {{ this.property(); }}");
        let loose = Some(
            json!({"strict":true, "noImplicitThis":false, "target":9, "module":99, "moduleResolution":100}),
        );
        let member = one(MEMBER, &source, loose.clone());
        assert_eq!(member.len(), 1);
        assert_eq!(member[0]["messageId"], member_id);
        assert_eq!(member[0]["message"], format!("{member_text}{advice}"));
        let strict_member = one(MEMBER, &source, None);
        assert_eq!(
            strict_member[0]["messageId"],
            if annotation == "any" {
                "unsafeMemberExpression"
            } else {
                "errorMemberExpression"
            }
        );
        for options in [None, loose] {
            let call = one(CALL, &source, options);
            assert_eq!(call.len(), 1);
            assert_eq!(call[0]["messageId"], "unsafeCallThis");
            assert_eq!(
                call[0]["message"],
                format!(
                    "Unsafe call of a(n) {call_type} typed value. `this` is typed as {call_type}.{advice}"
                )
            );
        }
    }
}

#[test]
fn intrinsic_error_receiver_and_key_keep_the_baseline_error_message_ids() {
    let source = "declare const value: NotKnown; declare const key: Missing; value[key];";
    let messages = one(MEMBER, source, None);
    assert_eq!(messages.len(), 2);
    assert_eq!(messages[0]["messageId"], "errorMemberExpression");
    assert_eq!(
        messages[0]["message"],
        "Unsafe member access [key] on a type that cannot be resolved."
    );
    assert_eq!(messages[1]["messageId"], "errorComputedMemberAccess");
    assert_eq!(
        messages[1]["message"],
        "The type of computed name [key] cannot be resolved."
    );
}
