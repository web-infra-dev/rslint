//! Frozen argument/return suites, with tsconfig semantics and original assertions.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

const ARGUMENT: &str = "@typescript-eslint/no-unsafe-argument";
const RETURN: &str = "@typescript-eslint/no-unsafe-return";
fn fixture_options() -> Value {
    json!({"jsx":"preserve","target":"esnext","module":"commonjs","strict":true,
        "strictNullChecks":true,"esModuleInterop":true,"lib":["es2015","es2017","esnext","DOM"],
        "experimentalDecorators":true,"types":[]})
}
fn lint(rule: &str, source: &str, support: Value, options: Value, fix: bool) -> Value {
    let cwd = "/rslint-rsbuild-unsafe-values-tests";
    let config = format!("{cwd}/tsconfig.json");
    let mut sources = Vec::new();
    let mut roots = vec!["input.ts".to_owned()];
    for (path, text) in support.as_object().into_iter().flatten() {
        sources.push(json!({"path":path,"text":text}));
        roots.push(path.clone());
    }
    sources.push(json!({"path":"tsconfig.json","text":json!({"compilerOptions":options,"files":roots}).to_string()}));
    let request = json!({"schemaVersion":1,"cwd":cwd,"files":[{"path":"input.ts","text":source,
        "rules":[{"name":rule}],"policy":{"kind":"explicit","configFiles":[config]}}],
        "sources":sources,"fix":fix});
    let output: Value =
        serde_json::from_str(&lint_projects_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(output["backend"], "tsrs-rust");
    assert_eq!(output["files"].as_array().unwrap().len(), 1);
    let result = output["files"][0].clone();
    assert_eq!(result["filePath"], format!("{cwd}/input.ts"));
    assert_eq!(result["appliedRules"], json!([rule]));
    assert_eq!(result["hasTypeInformation"], true);
    result
}
fn run(rule: &str, text: &str) -> Value {
    lint(rule, text, json!({}), fixture_options(), false)
}
#[test]
fn frozen_unsafe_values_corpus_keeps_every_default_case_and_original_assertion() {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-unsafe-values.json"
    ))
    .unwrap();
    let groups = groups.as_array().unwrap();
    assert_eq!(groups.len(), 2);
    assert_eq!(groups[0]["valid"].as_array().unwrap().len(), 46);
    assert_eq!(groups[0]["invalid"].as_array().unwrap().len(), 44);
    assert_eq!(groups[1]["valid"].as_array().unwrap().len(), 37);
    assert_eq!(groups[1]["invalid"].as_array().unwrap().len(), 53);
    for group in groups {
        assert_eq!(group["originalSkip"], json!([]));
        assert_eq!(group["excludedNonDefaultCases"], json!([]));
        let rule = group["rule"].as_str().unwrap();
        for category in ["valid", "invalid"] {
            for (index, case) in group[category].as_array().unwrap().iter().enumerate() {
                let source = case["text"].as_str().unwrap();
                let result = lint(
                    rule,
                    source,
                    case.get("files").cloned().unwrap_or(json!({})),
                    case["compilerOptions"].clone(),
                    false,
                );
                let messages = result["messages"].as_array().unwrap();
                let label = format!("{rule}/{category}/{index}, {}: {source}", case["origin"]);
                assert_eq!(
                    messages.is_empty(),
                    category == "valid",
                    "{label}: {messages:?}"
                );
                if category == "invalid" {
                    let mut expected = case["expected"].as_array().unwrap().clone();
                    expected.sort_by_key(|message| {
                        (
                            message["line"].as_u64().unwrap_or(0),
                            message["column"].as_u64().unwrap_or(0),
                        )
                    });
                    assert_eq!(messages.len(), expected.len(), "{label}: {messages:?}");
                    for (actual, expected) in messages.iter().zip(expected) {
                        for (key, value) in expected.as_object().unwrap() {
                            assert_eq!(&actual[key], value, "{label}: {actual:?}");
                        }
                    }
                }
                for message in messages {
                    assert_eq!(message["ruleId"], rule, "{label}");
                    assert_eq!(message["severity"], 2, "{label}");
                    assert!(message.get("fix").is_none(), "{label}");
                    assert!(message.get("suggestions").is_none(), "{label}");
                    assert!(message["endLine"].as_u64().is_some(), "{label}");
                    assert!(message["endColumn"].as_u64().is_some(), "{label}");
                }
                assert!(result.get("output").is_none(), "{label}");
            }
        }
    }
}
#[test]
fn unsafe_argument_reports_exact_utf16_range_without_autofix() {
    let text = "const emoji = '😀'; declare function take(value: number): void; take((1 as any));";
    let result = run(ARGUMENT, text);
    let messages = result["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 1);
    let message = &messages[0];
    assert_eq!(message["messageId"], "unsafeArgument");
    assert_eq!(
        message["message"],
        "Unsafe argument of type `any` assigned to a parameter of type `number`."
    );
    let start = text.find("1 as any").unwrap();
    assert_eq!(message["column"], text[..start].encode_utf16().count() + 1);
    assert_eq!(
        message["endColumn"],
        text[..start + "1 as any".len()].encode_utf16().count() + 1
    );
    let fixed = lint(ARGUMENT, text, json!({}), fixture_options(), true);
    assert_eq!(fixed["messages"], result["messages"]);
    assert!(fixed.get("output").is_none());
}
#[test]
fn unsafe_values_resolve_shared_generic_targets_and_nested_arguments_across_files() {
    let support = json!({"types.ts":"export interface Box<T> { value: T } export declare const unsafe: Box<Box<any>>; export declare const safe: Box<Box<string>>; export declare function take(value: Box<Box<string>>): void;"});
    let source = "import { unsafe, safe, take } from './types'; take(unsafe); take(safe);";
    let result = lint(ARGUMENT, source, support.clone(), fixture_options(), false);
    assert_eq!(result["messages"].as_array().unwrap().len(), 1);
    assert_eq!(
        result["messages"][0]["message"],
        "Unsafe argument of type `Box<Box<any>>` assigned to a parameter of type `Box<Box<string>>`."
    );
    let source = "import { unsafe, safe, type Box } from './types'; export const bad: () => Box<Box<string>> = () => unsafe; export const good: () => Box<Box<string>> = () => safe;";
    let result = lint(RETURN, source, support, fixture_options(), false);
    assert_eq!(result["messages"].as_array().unwrap().len(), 1);
    assert_eq!(result["messages"][0]["messageId"], "unsafeReturnAssignment");
    assert_eq!(
        result["messages"][0]["message"],
        "Unsafe return of type `Box<Box<any>>` from function with return type `Box<Box<string>>`."
    );
}
#[test]
fn tuple_rest_and_tagged_templates_consume_the_frozen_signature_cursor() {
    let text = "declare function take(a: number, ...rest: string[]): void; declare const tuple: readonly [any, ...string[]]; take(...tuple, 1 as any);";
    let result = run(ARGUMENT, text);
    let messages = result["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 2);
    assert_eq!(messages[0]["messageId"], "unsafeTupleSpread");
    assert_eq!(messages[1]["messageId"], "unsafeArgument");
    let result = run(
        ARGUMENT,
        "declare function tag(strings: TemplateStringsArray, n: number): void; tag`${1 as any}`;",
    );
    assert_eq!(result["messages"].as_array().unwrap().len(), 1);
    assert_eq!(
        result["messages"][0]["message"],
        "Unsafe argument of type `any` assigned to a parameter of type `number`."
    );
}
#[test]
fn return_preserves_implicit_this_message_and_explicit_intent_exemptions() {
    let source = "function value() { return this; }";
    let mut options = fixture_options();
    options["noImplicitThis"] = json!(false);
    let result = lint(RETURN, source, json!({}), options, false);
    assert_eq!(result["messages"].as_array().unwrap().len(), 1);
    assert_eq!(result["messages"][0]["messageId"], "unsafeReturnThis");
    assert_eq!(
        result["messages"][0]["message"],
        "Unsafe return of a value of type ``any``. `this` is typed as `any`.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function."
    );
    for source in [
        "function value(): any { return 1 as any; }",
        "function value(): unknown { return 1 as any; }",
        "function value(): unknown[] { return [] as any[]; }",
        "async function value(): Promise<unknown> { return Promise.resolve(1 as any); }",
        "function value() { return Promise.resolve(1 as any); }",
    ] {
        assert_eq!(run(RETURN, source)["messages"], json!([]), "{source}");
    }
}
#[test]
fn async_thenables_and_contextual_returns_preserve_type_classification() {
    let result = run(
        RETURN,
        "async function value() { return Promise.resolve(1 as any); }",
    );
    assert_eq!(result["messages"].as_array().unwrap().len(), 1);
    assert_eq!(
        result["messages"][0]["message"],
        "Unsafe return of a value of type `Promise<any>`."
    );
    let result = run(
        RETURN,
        "const value: () => Set<string> = () => new Set<any>();",
    );
    assert_eq!(result["messages"].as_array().unwrap().len(), 1);
    assert_eq!(result["messages"][0]["messageId"], "unsafeReturnAssignment");
    assert_eq!(
        run(
            RETURN,
            "const value: () => Map<string, string> = () => new Map();"
        )["messages"],
        json!([])
    );
    let result = run(RETURN, "function value() { return unknownName(); }");
    assert_eq!(
        result["messages"][0]["message"],
        "Unsafe return of a value of type error."
    );
}
