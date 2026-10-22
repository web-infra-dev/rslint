//! Rules enabled in Rsbuild: type spelling, literal assertions, and namespaces.
//! The shared corpus also runs against the independent frozen tsgo oracle.
use rslint_core::lint_json;
use serde_json::{Value, json};

const CONSTRAINT: &str = "@typescript-eslint/no-unnecessary-type-constraint";
const WRAPPER: &str = "@typescript-eslint/no-wrapper-object-types";
const AS_CONST: &str = "@typescript-eslint/prefer-as-const";
const NAMESPACE: &str = "@typescript-eslint/prefer-namespace-keyword";
const FUNCTION: &str = "@typescript-eslint/no-unsafe-function-type";

fn lint_files(rule: &str, files: &[(&str, &str)], fix: bool) -> Vec<Value> {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-ts-syntax-tests",
        "files": files.iter().map(|(path,text)|json!({"path":path,"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":rule}], "fix":fix});
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let results = result["files"].as_array().unwrap();
    assert_eq!(results.len(), files.len(), "retain silent admitted files");
    for (result, (path, _)) in results.iter().zip(files) {
        assert_eq!(
            result["filePath"],
            format!("/rslint-rsbuild-ts-syntax-tests/{path}")
        );
        assert_eq!(result["appliedRules"], json!([rule]));
    }
    results.clone()
}

fn lint(rule: &str, source: &str) -> Vec<Value> {
    lint_files(rule, &[("case.ts", source)], false)[0]["messages"]
        .as_array()
        .unwrap()
        .clone()
}

fn utf16(source: &str, offset: usize) -> usize {
    source[..offset].encode_utf16().count()
}

fn fix(source: &str, range: [usize; 2], text: &str) -> Value {
    json!({"range":[utf16(source,range[0]),utf16(source,range[1])],"text":text})
}

fn range_of(source: &str, needle: &str) -> [usize; 2] {
    let start = source.find(needle).unwrap();
    [start, start + needle.len()]
}

#[test]
fn rsbuild_ts_syntax_corpus_keeps_positive_and_negative_cases() {
    let groups: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-ts-syntax.json"
    ))
    .unwrap();
    for group in groups.as_array().unwrap() {
        let rule = group["rule"].as_str().unwrap();
        for key in ["valid", "invalid"] {
            let cases = group[key].as_array().unwrap();
            assert!(!cases.is_empty());
            // One Session per case keeps global declarations independent.
            for case in cases {
                let (source, extension) = if let Some(text) = case.as_str() {
                    (text, ".ts")
                } else {
                    (
                        case["text"].as_str().unwrap(),
                        case["extension"].as_str().unwrap(),
                    )
                };
                let results = lint_files(
                    rule,
                    &[(format!("case{extension}").as_str(), source)],
                    false,
                );
                let messages = results[0]["messages"].as_array().unwrap();
                assert_eq!(
                    messages.is_empty(),
                    key == "valid",
                    "{rule}, {key}, {source}: {messages:?}"
                );
                for message in messages {
                    assert_eq!(message["ruleId"], rule);
                    assert_eq!(message["severity"], 2);
                    assert!(message["endColumn"].as_u64().is_some());
                }
            }
        }
    }
}

#[test]
fn constraint_suggestions_preserve_defaults_and_arrow_disambiguation() {
    let source = "const f = <T extends unknown = string>() => {};";
    let diagnostics = lint_files(CONSTRAINT, &[("case.tsx", source)], false)[0]["messages"].clone();
    assert_eq!(diagnostics[0]["messageId"], "unnecessaryConstraint");
    assert_eq!(
        diagnostics[0]["message"],
        "Constraining the generic type `T` to `unknown` does nothing and is unnecessary."
    );
    assert_eq!(diagnostics[0]["column"], 12);
    assert_eq!(diagnostics[0]["endColumn"], 38);
    assert_eq!(
        diagnostics[0]["suggestions"],
        json!([{
            "messageId":"removeUnnecessaryConstraint", "desc":"Remove the unnecessary `unknown` constraint.",
            "fix":fix(source,range_of(source," extends unknown"),"")
        }])
    );
    assert!(diagnostics[0].get("fix").is_none());
    for extension in ["tsx", "mts", "cts"] {
        let source = "const f = <T extends any>() => {};";
        let diagnostics = lint_files(
            CONSTRAINT,
            &[(format!("case.{extension}").as_str(), source)],
            false,
        )[0]["messages"]
            .clone();
        assert_eq!(
            diagnostics[0]["suggestions"][0]["fix"],
            fix(source, range_of(source, " extends any"), ",")
        );
    }
    let source = "const f = <T extends any /* comma */,>() => {};";
    let diagnostics = lint_files(CONSTRAINT, &[("case.tsx", source)], false)[0]["messages"].clone();
    assert_eq!(
        diagnostics[0]["suggestions"][0]["fix"],
        fix(source, range_of(source, " extends any"), "")
    );
}

#[test]
fn wrappers_fix_only_type_references_not_heritage_targets() {
    let source = "interface I extends Number {} class C implements Number {} let x: Number;";
    let messages = lint(WRAPPER, source);
    assert_eq!(messages.len(), 3);
    for message in &messages {
        assert_eq!(message["messageId"], "bannedClassType");
        assert_eq!(
            message["message"],
            "Prefer using the primitive `number` as a type name, rather than the upper-cased `Number`."
        );
    }
    assert!(messages[0].get("fix").is_none());
    assert!(messages[1].get("fix").is_none());
    let start = source.rfind("Number").unwrap();
    assert_eq!(
        messages[2]["fix"],
        fix(source, [start, start + 6], "number")
    );
    let fixed = lint_files(WRAPPER, &[("case.ts", source)], true);
    assert_eq!(
        fixed[0]["output"],
        "interface I extends Number {} class C implements Number {} let x: number;"
    );
    assert_eq!(fixed[0]["messages"].as_array().unwrap().len(), 2);
}

#[test]
fn global_type_checks_distinguish_same_file_and_cross_file_declarations() {
    for (rule, name) in [(WRAPPER, "Number"), (FUNCTION, "Function")] {
        let augmentation =
            format!("declare global {{ interface {name} {{ custom: string }} }} export {{}};");
        let use_source = format!("let value: {name}; export {{}};");
        let files = lint_files(
            rule,
            &[("augment.ts", &augmentation), ("use.ts", &use_source)],
            false,
        );
        assert!(files[0]["messages"].as_array().unwrap().is_empty());
        assert_eq!(files[1]["messages"].as_array().unwrap().len(), 1);
        let same = format!("{augmentation}\n{use_source}");
        assert!(lint(rule, &same).is_empty());
        let catch = format!("try {{}} catch ({name}) {{ let value: {name}; }}");
        assert!(lint(rule, &catch).is_empty());
        let property = format!("const object = {{ {name}: 1 }}; let value: {name};");
        assert_eq!(lint(rule, &property).len(), 1);
    }
}

#[test]
fn prefer_as_const_preserves_comments_and_suggests_annotation_edits() {
    let source = "let 值 /* before */ : /* after: colon */ '值' = /* value */ '值';";
    let messages = lint(AS_CONST, source);
    assert_eq!(messages.len(), 1);
    let literal = range_of(source, "'值'");
    assert_eq!(
        messages[0]["range"],
        json!([utf16(source, literal[0]), utf16(source, literal[1])])
    );
    assert_eq!(messages[0]["messageId"], "variableConstAssertion");
    assert_eq!(
        messages[0]["message"],
        "Expected a `const` assertion instead of a literal type annotation."
    );
    let colon = source.find(": /*").unwrap();
    let end = source.rfind("'值'").unwrap() + "'值'".len();
    assert_eq!(
        messages[0]["suggestions"],
        json!([{
            "messageId":"variableSuggest", "desc":"You should use `as const` instead of type annotation.",
            "fix":fix(source,[colon,end]," = /* value */ '值' as const")
        }])
    );
    assert!(messages[0].get("fix").is_none());
    let suggestion_only = lint_files(AS_CONST, &[("case.ts", source)], true);
    assert!(suggestion_only[0].get("output").is_none());
    let source = "const value = /* value */ '值' as /* type */ '值';";
    let messages = lint(AS_CONST, source);
    let start = source.rfind("'值'").unwrap();
    assert_eq!(
        messages[0]["fix"],
        fix(source, [start, start + "'值'".len()], "const")
    );
    assert_eq!(
        lint_files(AS_CONST, &[("case.ts", source)], true)[0]["output"],
        "const value = /* value */ '值' as /* type */ const;"
    );
}

#[test]
fn namespace_fix_preserves_modifiers_comments_and_dotted_names() {
    let source = "/* first */ export declare /* keyword */ module /* name */ Foo.Bar {}";
    let messages = lint(NAMESPACE, source);
    assert_eq!(messages.len(), 1);
    assert_eq!(messages[0]["messageId"], "useNamespace");
    assert_eq!(
        messages[0]["message"],
        "Use 'namespace' instead of 'module' to declare custom TypeScript modules."
    );
    assert_eq!(messages[0]["column"], 13);
    assert_eq!(
        messages[0]["fix"],
        fix(source, range_of(source, "module"), "namespace")
    );
    assert_eq!(
        lint_files(NAMESPACE, &[("case.ts", source)], true)[0]["output"],
        source.replacen("module", "namespace", 1)
    );
}

#[test]
fn unsafe_function_type_has_multiline_message_and_no_edits() {
    let source = "let x: Function; interface I extends Function {}";
    let messages = lint(FUNCTION, source);
    assert_eq!(messages.len(), 2);
    for message in messages {
        assert_eq!(message["messageId"], "bannedFunctionType");
        assert_eq!(
            message["message"],
            "The `Function` type accepts any function-like value.\nPrefer explicitly defining any function parameters and return type."
        );
        assert!(message.get("fix").is_none());
        assert!(message.get("suggestions").is_none());
    }
}
