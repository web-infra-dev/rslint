//! Frozen typed usage rules: corpus, exact messages/ranges, cross-file symbols.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};

const BASE: &str = "@typescript-eslint/no-base-to-string";
const EVAL: &str = "@typescript-eslint/no-implied-eval";
const UNBOUND: &str = "@typescript-eslint/unbound-method";

fn compiler_options() -> Value {
    json!({"jsx":"preserve","target":"ESNext","module":"CommonJS",
        "moduleResolution":"Node10","strict":true,"strictNullChecks":true,
        "esModuleInterop":true,"lib":["es2015","es2017","esnext","DOM"],
        "experimentalDecorators":true,"types":[]})
}

fn lint(rule: &str, entry: &Value, fix: bool) -> Value {
    let root = "/rslint-rsbuild-type-usage-tests";
    let default_name = format!("input{}", entry["extension"].as_str().unwrap_or(".ts"));
    let name = entry["fileName"].as_str().unwrap_or(&default_name);
    let source = entry["text"].as_str().unwrap();
    let mut sources = Vec::new();
    if let Some(files) = entry["files"].as_object() {
        for (path, source) in files {
            sources.push(json!({"path":path,"text":source}));
        }
    }
    let mut roots = vec![name.to_string()];
    if let Some(files) = entry["projectFiles"].as_array() {
        roots.extend(files.iter().map(|path| path.as_str().unwrap().to_string()));
    }
    roots.sort();
    roots.dedup();
    sources.push(json!({"path":"tsconfig.json","text":json!({
        "compilerOptions":entry.get("compilerOptions").cloned().unwrap_or_else(compiler_options),
        "files":roots
    }).to_string()}));
    let request = json!({"schemaVersion":1,"cwd":root,"sources":sources,"fix":fix,
        "files":[{"path":name,"text":source,"rules":[{"name":rule}],
        "policy":{"kind":"explicit","configFiles":[format!("{root}/tsconfig.json")]}}]});
    let result: Value =
        serde_json::from_str(&lint_projects_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    let result = result["files"][0].clone();
    assert_eq!(result["hasTypeInformation"], true);
    assert_eq!(result["appliedRules"], json!([rule]));
    result
}

fn messages(rule: &str, source: &str) -> Vec<Value> {
    lint(rule, &json!({"text":source}), false)["messages"]
        .as_array()
        .unwrap()
        .clone()
}

#[test]
fn frozen_default_type_usage_cases_preserve_every_expected_diagnostic() {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-type-usage.json"
    ))
    .unwrap();
    let groups = groups.as_array().unwrap();
    assert_eq!(groups.len(), 3);
    for (group, counts) in
        groups
            .iter()
            .zip([[272, 81, 1, 2, 0], [48, 22, 0, 0, 0], [159, 87, 2, 4, 2]])
    {
        for (field, count) in [
            "valid",
            "invalid",
            "originalSkip",
            "excludedOptions",
            "excludedAdapters",
        ]
        .into_iter()
        .zip(counts)
        {
            assert_eq!(
                group[field].as_array().unwrap().len(),
                count,
                "{}: {field}",
                group["rule"]
            );
        }
        let rule = group["rule"].as_str().unwrap();
        for category in ["valid", "invalid"] {
            for (index, entry) in group[category].as_array().unwrap().iter().enumerate() {
                let result = lint(rule, entry, false);
                let mut actual = result["messages"].as_array().unwrap().clone();
                let source = entry["text"].as_str().unwrap();
                let expected = entry["expected"].as_array().cloned().unwrap_or_default();
                assert_eq!(
                    actual.len(),
                    expected.len(),
                    "{rule}/{category}/{index}: {source}: {actual:?}"
                );
                assert_eq!(
                    actual.is_empty(),
                    category == "valid",
                    "{rule}/{category}/{index}"
                );
                for expected in expected {
                    let found = actual.iter().position(|actual| {
                        [
                            "messageId",
                            "message",
                            "line",
                            "column",
                            "endLine",
                            "endColumn",
                        ]
                        .iter()
                        .all(|field| {
                            expected
                                .get(field)
                                .is_none_or(|value| value == &actual[field])
                        })
                    });
                    assert!(
                        found.is_some(),
                        "{rule}/{category}/{index}: {source}: expected {expected:?}; remaining {actual:?}"
                    );
                    let report = actual.remove(found.unwrap());
                    assert_eq!(report["ruleId"], rule);
                    assert_eq!(report["severity"], 2);
                    assert!(report.get("fix").is_none());
                    assert!(report["suggestions"].as_array().is_none_or(Vec::is_empty));
                }
            }
        }
    }
}

#[test]
fn base_stringification_preserves_frozen_message_trivia_and_safe_overrides() {
    let source = "const emoji = '😀'; const value = {}; '' + /* retained */ value;";
    let reports = messages(BASE, source);
    assert_eq!(reports.len(), 1);
    assert_eq!(reports[0]["messageId"], "baseToString");
    assert_eq!(
        reports[0]["message"],
        "' /* retained */ value' always use Object's default stringification format ('[object Object]') when stringified."
    );
    assert_eq!(
        reports[0]["range"],
        json!([
            source[..source.rfind("value").unwrap()]
                .encode_utf16()
                .count(),
            source[..source.rfind(';').unwrap()].encode_utf16().count()
        ])
    );
    assert!(
        messages(
            BASE,
            "const value = { toString() { return 'safe'; } }; String(value); `${value}`;"
        )
        .is_empty()
    );
    assert!(messages(BASE, "type Recursive = Recursive[]; declare const value: Recursive; String(value); value.join();").is_empty());
    let reports = messages(BASE, "const values = [{}]; values.join();");
    assert_eq!(reports[0]["messageId"], "baseArrayJoin");
    assert_eq!(
        reports[0]["message"],
        "Using `join()` for  values always use Object's default stringification format ('[object Object]') when stringified."
    );
}

#[test]
fn implied_eval_distinguishes_local_functions_and_imported_constructor_symbols() {
    let source = "const emoji = '😀'; setTimeout('alert(1)', 0); new Function('return 1');";
    let reports = messages(EVAL, source);
    assert_eq!(reports.len(), 2);
    assert_eq!(reports[0]["messageId"], "noImpliedEvalError");
    assert_eq!(
        reports[0]["message"],
        "Implied eval. Consider passing a function."
    );
    assert_eq!(reports[1]["messageId"], "noFunctionConstructor");
    assert_eq!(
        reports[1]["message"],
        "Implied eval. Do not use the Function constructor to create functions."
    );
    assert!(
        messages(
            EVAL,
            "export {}; function setTimeout(value: string) {} setTimeout('safe');"
        )
        .is_empty()
    );
    assert!(
        messages(
            EVAL,
            "declare function handler(): void; setTimeout(handler.bind(null), 0);"
        )
        .is_empty()
    );
    let entry = json!({"text":"import { Function } from './dependency'; new Function('text');",
        "files":{"dependency.ts":"export class Function { constructor(value: string) {} }"}});
    assert!(
        lint(EVAL, &entry, false)["messages"]
            .as_array()
            .unwrap()
            .is_empty()
    );
}

#[test]
fn unbound_messages_distinguish_method_annotations_fields_and_safe_uses() {
    let source = "class Service { method() {} explicit(this: Service) {} safe(this: void) {} field = function() {}; bound = () => {}; } const value = new Service(); value.method; value.explicit; value.safe; value.field; value.bound;";
    let reports = messages(UNBOUND, source);
    assert_eq!(
        reports
            .iter()
            .map(|value| value["messageId"].as_str().unwrap())
            .collect::<Vec<_>>(),
        vec!["unboundWithoutThisAnnotation", "unbound", "unbound"]
    );
    assert_eq!(
        reports[0]["message"],
        "Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead."
    );
    assert_eq!(
        reports[1]["message"],
        "Avoid referencing unbound methods which may cause unintentional scoping of `this`."
    );
    let source = "class Service { method() {} } const value = new Service(); value.method(); if (value.method) {} void value.method; value.method.bind(value);";
    assert!(messages(UNBOUND, source).is_empty());
}

#[test]
fn imported_methods_and_union_computed_keys_are_not_replaced_by_name_matching() {
    let entry = json!({"text":"import { console, Service } from './dependency'; const log = console.log; const service = new Service(); const { method } = service;",
        "files":{"dependency.ts":"export const console = { log() {} }; export class Service { method() {} }"}});
    let reports = lint(UNBOUND, &entry, false)["messages"]
        .as_array()
        .unwrap()
        .clone();
    assert_eq!(reports.len(), 2);
    assert!(
        reports
            .iter()
            .all(|report| report["messageId"] == "unboundWithoutThisAnnotation")
    );
    assert!(
        messages(
            UNBOUND,
            "const log = console.log; const floor = Math.floor; const { parseInt } = Number;"
        )
        .is_empty()
    );
    let source = "class A { method() {} } class B { method = 1; } declare const value: A | B; declare const key: 'method' | 'absent'; value[key];";
    let reports = messages(UNBOUND, source);
    assert_eq!(reports.len(), 1);
    assert_eq!(
        reports[0]["range"],
        json!([
            source.find("value[key]").unwrap(),
            source.find("value[key]").unwrap() + 10
        ])
    );
}

#[test]
fn these_rules_never_modify_sources_in_fix_mode() {
    for (rule, source) in [
        (BASE, "String({});"),
        (EVAL, "setTimeout('code', 0);"),
        (UNBOUND, "class C { method() {} } new C().method;"),
    ] {
        let entry = json!({"text":source});
        let before = lint(rule, &entry, false);
        let after = lint(rule, &entry, true);
        assert!(!before["messages"].as_array().unwrap().is_empty());
        assert_eq!(before["messages"], after["messages"]);
        assert!(after.get("output").is_none());
    }
}
