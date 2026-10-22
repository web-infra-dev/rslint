//! Typed await rules: frozen source cases plus diagnostic/suggestion contracts.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

const AWAIT: &str = "@typescript-eslint/await-thenable";
const REQUIRE: &str = "@typescript-eslint/require-await";

fn fixture_options() -> Value {
    json!({"jsx":"preserve","target":"ESNext","module":"CommonJS","moduleResolution":"Node10",
        "strict":true,"strictNullChecks":true,"esModuleInterop":true,
        "lib":["es2015","es2017","esnext","DOM"],"experimentalDecorators":true,"types":[]})
}
fn lint(rule: &str, files: &[(&str, &str)], options: Value, fix: bool) -> Vec<Value> {
    // The project API accepts tsconfig spellings and gives every overlay its
    // original program-root identity. Do not concatenate support declarations:
    // that changes module resolution, source ranges, and declaration identity.
    let root = "/rslint-rsbuild-await-tests";
    let config = format!("{root}/tsconfig.json");
    let sources = vec![json!({"path":"tsconfig.json","text":json!({
        "compilerOptions":options,
        "files":files.iter().map(|(path,_)|*path).collect::<Vec<_>>()
    }).to_string()})];
    let request = json!({"schemaVersion":1,"cwd":root,
        "files":files.iter().map(|(path,text)|json!({"path":path,"text":text,
            "rules":[{"name":rule}],"policy":{"kind":"explicit","configFiles":[config]}})).collect::<Vec<_>>(),
        "sources":sources,"fix":fix});
    let output: Value =
        serde_json::from_str(&lint_projects_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(output["backend"], "tsrs-rust");
    let results = output["files"].as_array().unwrap();
    assert_eq!(results.len(), files.len(), "retain silent admitted files");
    for result in results {
        assert_eq!(result["appliedRules"], json!([rule]));
    }
    results.clone()
}
fn messages(rule: &str, source: &str) -> Vec<Value> {
    lint(rule, &[("input.ts", source)], fixture_options(), false)[0]["messages"]
        .as_array()
        .unwrap()
        .clone()
}
fn apply_suggestion(source: &str, suggestion: &Value) -> String {
    let units = source.encode_utf16().collect::<Vec<_>>();
    let fix = &suggestion["fix"];
    let start = fix["range"][0].as_u64().unwrap() as usize;
    let end = fix["range"][1].as_u64().unwrap() as usize;
    String::from_utf16(&units[..start]).unwrap()
        + fix["text"].as_str().unwrap()
        + &String::from_utf16(&units[end..]).unwrap()
}

#[test]
fn frozen_await_corpus_keeps_all_default_rule_inputs() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-await-checks.json"
    ))
    .unwrap();
    let groups = corpus.as_array().unwrap();
    assert_eq!(groups.len(), 2);
    for group in groups {
        let rule = group["rule"].as_str().unwrap();
        assert_eq!(group["originalSkip"], json!([]));
        for category in ["valid", "invalid"] {
            let cases = group[category].as_array().unwrap();
            assert!(!cases.is_empty());
            for (index, case) in cases.iter().enumerate() {
                let source = case["text"].as_str().or_else(|| case.as_str()).unwrap();
                let extension = case["extension"].as_str().unwrap_or(".ts");
                let name = format!("input{extension}");
                let mut files = vec![(name.as_str(), source)];
                if let Some(support) = case["files"].as_object() {
                    files.extend(
                        support
                            .iter()
                            .map(|(path, text)| (path.as_str(), text.as_str().unwrap())),
                    );
                }
                let options = case
                    .get("compilerOptions")
                    .cloned()
                    .unwrap_or_else(fixture_options);
                let results = lint(rule, &files, options, false);
                let diagnostics = results[0]["messages"].as_array().unwrap();
                assert_eq!(
                    diagnostics.is_empty(),
                    category == "valid",
                    "{rule}/{category}/{index}: {source}: {diagnostics:?}"
                );
                for diagnostic in diagnostics {
                    assert_eq!(diagnostic["ruleId"], rule);
                    assert_eq!(diagnostic["severity"], 2);
                    assert!(
                        diagnostic.get("fix").is_none(),
                        "these rules only suggest edits"
                    );
                    for suggestion in diagnostic["suggestions"].as_array().into_iter().flatten() {
                        assert!(
                            suggestion.get("data").is_none(),
                            "frozen suggestions have no template data"
                        );
                    }
                }
            }
        }
    }
}

#[test]
fn unnecessary_await_suggestion_preserves_comments_and_utf16() {
    // Frozen public API treats parenthesized top-level `await` in a script as a
    // call expression. Keep that exact input and its module variant separate.
    let script = "const emoji = '😀'; await /* retained */ (1);";
    assert!(messages(AWAIT, script).is_empty());
    let source = "const emoji = '😀'; await /* retained */ (1);\nexport {};";
    let reports = messages(AWAIT, source);
    assert_eq!(reports.len(), 1);
    let report = &reports[0];
    assert_eq!(report["messageId"], "await");
    assert_eq!(
        report["message"],
        "Unexpected `await` of a non-Promise (non-\"Thenable\") value."
    );
    let begin = source.find("await").unwrap();
    assert_eq!(report["column"], source[..begin].encode_utf16().count() + 1);
    let first_line = source.lines().next().unwrap();
    assert_eq!(report["line"], 1);
    assert_eq!(report["endLine"], 1);
    assert_eq!(
        report["endColumn"],
        first_line[..first_line.len() - 1].encode_utf16().count() + 1
    );
    let suggestion = &report["suggestions"][0];
    assert_eq!(suggestion["messageId"], "removeAwait");
    assert_eq!(suggestion["desc"], "Remove unnecessary `await`.");
    assert_eq!(
        apply_suggestion(source, suggestion),
        "const emoji = '😀';  /* retained */ (1);\nexport {};"
    );
    let fixed = lint(AWAIT, &[("input.ts", source)], fixture_options(), true);
    assert!(
        fixed[0].get("output").is_none(),
        "a suggestion must not auto-apply"
    );
    assert_eq!(fixed[0]["messages"], json!(reports));
}

#[test]
fn require_await_suggestions_unwrap_only_native_return_type_names() {
    for (source, expected) in [
        (
            "async function value(): Promise<number> { return 1; }",
            "function value(): number { return 1; }",
        ),
        (
            "async function value(): ((Promise /* keep */ <number>)) { return 1; }",
            "function value(): ((number)) { return 1; }",
        ),
        (
            "async function* value(): AsyncGenerator<number> { yield 1; }",
            "function* value(): Generator<number> { yield 1; }",
        ),
        (
            "async function value(): globalThis.Promise<number> { return 1; }",
            "function value(): globalThis.Promise<number> { return 1; }",
        ),
        (
            "async /* keep */ function value() { return 1; }",
            "/* keep */ function value() { return 1; }",
        ),
    ] {
        let reports = messages(REQUIRE, source);
        assert_eq!(reports.len(), 1, "{source}");
        assert_eq!(reports[0]["messageId"], "missingAwait");
        let suggestion = &reports[0]["suggestions"][0];
        assert_eq!(suggestion["messageId"], "removeAsync");
        assert_eq!(suggestion["desc"], "Remove 'async'.");
        assert_eq!(apply_suggestion(source, suggestion), expected, "{source}");
    }
    let source = "async function value(): Promise<number> { return 1; }";
    let report = &messages(REQUIRE, source)[0];
    assert_eq!(
        report["message"],
        "Async function 'value' has no 'await' expression."
    );
    assert_eq!(report["column"], 1);
    assert_eq!(report["endColumn"], 21);
}

#[test]
fn require_await_preserves_class_asi_and_jsdoc_attachment() {
    for (source, expected) in [
        (
            "class C { field = value\n async [key]() { return 1; } }",
            "class C { field = value\n ;[key]() { return 1; } }",
        ),
        (
            "class C { field = value\n /** @deprecated */\n async [key]() { return 1; } }",
            "class C { field = value;\n /** @deprecated */\n [key]() { return 1; } }",
        ),
        (
            "class C { field = value++\n async [key]() { return 1; } }",
            "class C { field = value++\n [key]() { return 1; } }",
        ),
        (
            "class C { field = value++\n async *method() { yield 1; } }",
            "class C { field = value++\n ;*method() { yield 1; } }",
        ),
    ] {
        let reports = messages(REQUIRE, source);
        assert_eq!(reports.len(), 1);
        assert_eq!(
            apply_suggestion(source, &reports[0]["suggestions"][0]),
            expected
        );
    }
    let reports = messages(REQUIRE, "const value = async /** @deprecated */ () => 1;");
    assert_eq!(reports.len(), 1);
    assert!(
        reports[0].get("suggestions").is_none(),
        "do not turn an internal JSDoc into declaration metadata"
    );
}

#[test]
fn cross_file_thenables_use_callback_types_and_well_known_symbols() {
    let support = "export interface Good { then(...callbacks: ((n: number) => unknown)[]): unknown; }\nexport interface Bad { then(callback: number): unknown; }\nexport declare const good: Good; export declare const bad: Bad;";
    for (rule, source, expected) in [
        (
            AWAIT,
            "import {good, bad} from './types'; async function run() { await good; await bad; }",
            vec!["await"],
        ),
        (
            REQUIRE,
            "import {good, bad} from './types'; async function goodReturn() { return good; } async function badReturn() { return bad; }",
            vec!["missingAwait"],
        ),
    ] {
        let result = lint(
            rule,
            &[("input.ts", source), ("types.ts", support)],
            fixture_options(),
            false,
        );
        let reports = result[0]["messages"].as_array().unwrap();
        assert_eq!(
            reports
                .iter()
                .map(|m| m["messageId"].as_str().unwrap())
                .collect::<Vec<_>>(),
            expected
        );
        assert_eq!(result[1]["messages"], json!([]));
    }
    assert!(
        messages(
            REQUIRE,
            "async function* f(value: AsyncIterable<number>) { yield* value; }"
        )
        .is_empty()
    );
    assert!(
        messages(
            AWAIT,
            "async function f(value: AsyncIterable<number>) { for await (const item of value) {} }"
        )
        .is_empty()
    );
    let reports = messages(
        AWAIT,
        "async function f(value: Iterable<Promise<number>>) { for await (const item of value) {} }",
    );
    assert_eq!(reports.len(), 1);
    assert_eq!(reports[0]["messageId"], "forAwaitOfNonAsyncIterable");
}

#[test]
fn aggregator_identity_and_static_keys_are_checked_semantically() {
    for source in [
        "const method = 'all'; Promise[method]([1]);",
        "const names = Object.freeze({key: 'all'}); Promise[names.key]([1]);",
        "const part = 'll'; Promise[`a${part}`]([1]);",
        "Promise[String('all')]([1]);",
        "Promise[['a', 'll'].join('')]([1]);",
        "const encode = String; Promise[encode('all')]([1]);",
        "const max = Math.max; Promise[max(1,2)===2 ? 'all' : 'other']([1]);",
        "Promise[Number.parseInt('0010',2)===2 ? 'all' : 'other']([1]);",
        "Promise[String.raw`a${'ll'}`]([1]);",
    ] {
        let reports = messages(AWAIT, source);
        assert_eq!(reports.len(), 1, "{source}: {reports:?}");
        assert_eq!(reports[0]["messageId"], "invalidPromiseAggregatorInput");
        assert!(reports[0].get("suggestions").is_none());
    }
    for source in [
        "const Promise = {all(values: unknown[]) { return values; }}; Promise.all([1]); export {};",
        "declare const key: 'all'; Promise[key]([1]);",
        "let key: 'all' | 'race' = 'all'; key = 'race'; Promise[key]([1]);",
        "class P extends Promise<unknown> {} P.all([1]);",
    ] {
        assert!(messages(AWAIT, source).is_empty(), "{source}");
    }
}

#[test]
fn static_key_mutation_tracking_matches_frozen_public_diagnostics() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-await-checks.json"
    ))
    .unwrap();
    let cases = corpus
        .as_array()
        .unwrap()
        .iter()
        .find(|entry| entry["rule"] == AWAIT)
        .unwrap();
    let mut count = 0;
    for category in ["valid", "invalid"] {
        for case in cases[category].as_array().unwrap() {
            if !case["regression"]
                .as_str()
                .is_some_and(|tag| tag.starts_with("staticMutation:"))
            {
                continue;
            }
            count += 1;
            let source = case["text"].as_str().unwrap();
            let mut actual = messages(AWAIT, source);
            for message in &mut actual {
                // The public API omits the internal range. Verify its UTF-16
                // projection separately before comparing all public fields.
                let offset = |line: u64, column: u64| -> usize {
                    source
                        .split_inclusive('\n')
                        .take(line as usize - 1)
                        .map(|text| text.encode_utf16().count())
                        .sum::<usize>()
                        + column as usize
                        - 1
                };
                assert_eq!(
                    message["range"],
                    json!([
                        offset(
                            message["line"].as_u64().unwrap(),
                            message["column"].as_u64().unwrap()
                        ),
                        offset(
                            message["endLine"].as_u64().unwrap(),
                            message["endColumn"].as_u64().unwrap()
                        )
                    ])
                );
                message.as_object_mut().unwrap().remove("range");
            }
            assert_eq!(
                json!(actual),
                case["expectedMessages"],
                "{}: {source}",
                case["regression"]
            );
        }
    }
    assert_eq!(
        count, 21,
        "the frozen mutation comparison must not silently select no cases"
    );
}
