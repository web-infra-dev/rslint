//! Declaration rules used by the frozen Rsbuild config. The same corpus is
//! checked against the independent original engine by rust-rule-oracle.mjs.
use rslint_core::lint_json;
use serde_json::{Value, json};

const EMPTY: &str = "@typescript-eslint/no-empty-object-type";
const NAMESPACE: &str = "@typescript-eslint/no-namespace";
const THIS: &str = "@typescript-eslint/no-this-alias";
const ARRAY: &str = "@typescript-eslint/no-array-constructor";

fn lint_files(rule: &str, files: &[(&str, &str)], fix: bool) -> Vec<Value> {
    let request = json!({"schemaVersion":1,"cwd":"/rslint-rsbuild-declaration-tests",
        "files":files.iter().map(|(path,text)|json!({"path":path,"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":rule}],"fix":fix});
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let results = result["files"].as_array().unwrap();
    assert_eq!(results.len(), files.len(), "retain silent admitted files");
    for (result, (path, _)) in results.iter().zip(files) {
        assert_eq!(
            result["filePath"],
            format!("/rslint-rsbuild-declaration-tests/{path}")
        );
        assert_eq!(result["appliedRules"], json!([rule]));
    }
    results.clone()
}

fn messages(rule: &str, source: &str) -> Vec<Value> {
    lint_files(rule, &[("input.ts", source)], false)[0]["messages"]
        .as_array()
        .unwrap()
        .clone()
}

fn utf16(text: &str, index: usize) -> usize {
    text[..index].encode_utf16().count()
}

fn fix(source: &str, range: [usize; 2], replacement: &str) -> Value {
    json!({"range":[utf16(source,range[0]),utf16(source,range[1])],"text":replacement})
}

#[test]
fn declaration_corpus_has_positive_and_negative_coverage() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-ts-declarations.json"
    ))
    .unwrap();
    for group in corpus.as_array().unwrap() {
        let rule = group["rule"].as_str().unwrap();
        for key in ["valid", "invalid"] {
            let cases = group[key].as_array().unwrap();
            assert!(!cases.is_empty());
            for case in cases {
                let (source, extension) = case.as_str().map_or_else(
                    || {
                        (
                            case["text"].as_str().unwrap(),
                            case["extension"].as_str().unwrap_or(".ts"),
                        )
                    },
                    |source| (source, ".ts"),
                );
                let results = lint_files(
                    rule,
                    &[(format!("case{extension}").as_str(), source)],
                    false,
                );
                let diagnostics = results[0]["messages"].as_array().unwrap();
                assert_eq!(
                    diagnostics.is_empty(),
                    key == "valid",
                    "{rule}: {source}: {diagnostics:?}"
                );
                for diagnostic in diagnostics {
                    assert_eq!(diagnostic["ruleId"], rule);
                    assert_eq!(diagnostic["severity"], 2);
                    assert!(diagnostic["endColumn"].as_u64().is_some());
                }
            }
        }
    }
}

#[test]
fn empty_interface_suggestions_keep_modifiers_trivia_and_generic_clause() {
    let source = "export declare /* keep */ interface Box< /* first */ T, /* last */ > {}";
    let report = &messages(EMPTY, source)[0];
    assert_eq!(report["messageId"], "noEmptyInterface");
    assert_eq!(
        report["message"],
        "An empty interface declaration allows any non-nullish value, including literals like `0` and `\"\"`.\n- If that's what you want, disable this lint rule with an inline comment or configure the 'allowInterfaces' rule option.\n- If you want a type meaning \"any object\", you probably want `object` instead.\n- If you want a type meaning \"any value\", you probably want `unknown` instead."
    );
    let start = source.find("Box").unwrap();
    assert_eq!(report["line"], 1);
    assert_eq!(report["column"], start + 1);
    assert_eq!(report["endColumn"], start + 4);
    assert!(report.get("fix").is_none());
    let suggestions = report["suggestions"].as_array().unwrap();
    assert_eq!(suggestions.len(), 2);
    for (suggestion, replacement) in suggestions.iter().zip(["object", "unknown"]) {
        assert_eq!(suggestion["messageId"], "replaceEmptyInterface");
        assert_eq!(suggestion["data"], json!({"replacement": replacement}));
        assert_eq!(
            suggestion["desc"],
            format!("Replace empty interface with `{replacement}`.")
        );
        assert_eq!(
            suggestion["fix"],
            fix(
                source,
                [source.find("interface").unwrap(), source.len()],
                &format!("type Box< /* first */ T, /* last */ > = {replacement}")
            )
        );
    }
    let fixed = lint_files(EMPTY, &[("input.ts", source)], true);
    assert!(
        fixed[0].get("output").is_none(),
        "suggestions are not automatic fixes"
    );
    assert_eq!(fixed[0]["messages"][0], *report);
}

#[test]
fn empty_interface_merge_suppresses_unsafe_replacement_suggestions() {
    for source in [
        "interface Shared {} interface Shared { value: string }",
        "interface Shared {} class Shared { value = 'value'; }",
        "class Shared { value = 'value'; } interface Shared {}",
        "export default interface Shared {}",
        "export default interface Shared extends Base {}",
    ] {
        let diagnostics = messages(EMPTY, source);
        assert_eq!(diagnostics.len(), 1, "{source}");
        assert!(diagnostics[0].get("suggestions").is_none(), "{source}");
    }
    for other in [
        "interface Shared { value: string }",
        "class Shared { value = 'value'; }",
    ] {
        let results = lint_files(
            EMPTY,
            &[("empty.ts", "interface Shared {}"), ("other.ts", other)],
            false,
        );
        assert_eq!(results[0]["messages"].as_array().unwrap().len(), 1);
        assert!(results[0]["messages"][0].get("suggestions").is_none());
        assert_eq!(results[1]["messages"], json!([]));
    }
    let results = lint_files(
        EMPTY,
        &[
            ("empty.ts", "export interface Shared {}"),
            ("other.ts", "export interface Shared { value: string }"),
        ],
        false,
    );
    assert_eq!(
        results[0]["messages"][0]["suggestions"]
            .as_array()
            .unwrap()
            .len(),
        2,
        "unrelated module bindings must not suppress suggestions"
    );
}

#[test]
fn empty_object_and_single_base_have_exact_distinct_suggestions() {
    let source = "interface Box<T> extends Parent<T> {}";
    let report = &messages(EMPTY, source)[0];
    assert_eq!(report["messageId"], "noEmptyInterfaceWithSuper");
    assert_eq!(
        report["message"],
        "An interface declaring no members is equivalent to its supertype."
    );
    assert_eq!(
        report["suggestions"],
        json!([{
            "messageId":"replaceEmptyInterfaceWithSuper",
            "desc":"Replace empty interface with a type alias.",
            "fix":fix(source,[0,source.len()],"type Box<T> = Parent<T>")
        }])
    );
    let source = "type Empty = {};";
    let report = &messages(EMPTY, source)[0];
    assert_eq!(report["messageId"], "noEmptyObject");
    assert_eq!(report["column"], 14);
    assert_eq!(report["endColumn"], 16);
    assert_eq!(
        report["suggestions"],
        json!([
            {"messageId":"replaceEmptyObjectType","desc":"Replace `{}` with `object`.","data":{"replacement":"object"},"fix":fix(source,[13,15],"object")},
            {"messageId":"replaceEmptyObjectType","desc":"Replace `{}` with `unknown`.","data":{"replacement":"unknown"},"fix":fix(source,[13,15],"unknown")}
        ])
    );
}

#[test]
fn namespace_chain_reports_outer_declaration_but_nested_body_reports_both() {
    let source = "/* lead */ declare namespace A.B.C {}";
    let diagnostics = messages(NAMESPACE, source);
    assert_eq!(diagnostics.len(), 1);
    assert_eq!(diagnostics[0]["messageId"], "moduleSyntaxIsPreferred");
    assert_eq!(
        diagnostics[0]["message"],
        "ES2015 module syntax is preferred over namespaces."
    );
    assert_eq!(
        diagnostics[0]["column"],
        source.find("declare").unwrap() + 1
    );
    assert_eq!(diagnostics[0]["endColumn"], source.len() + 1);
    assert!(diagnostics[0].get("fix").is_none());
    assert!(diagnostics[0].get("suggestions").is_none());
    let source = "namespace Outer { namespace Inner {} }";
    let diagnostics = messages(NAMESPACE, source);
    assert_eq!(diagnostics.len(), 2);
    assert_eq!(
        diagnostics[1]["column"],
        source.find("namespace Inner").unwrap() + 1
    );
    for extension in [".d.ts", ".d.mts", ".d.cts"] {
        let results = lint_files(
            NAMESPACE,
            &[(format!("case{extension}").as_str(), source)],
            false,
        );
        assert_eq!(results[0]["messages"], json!([]));
    }
}

#[test]
fn this_alias_uses_binding_annotation_and_utf16_target_positions() {
    let source = "const emoji = '😀'; const self: unknown = (((this)));";
    let report = &messages(THIS, source)[0];
    assert_eq!(report["messageId"], "thisAssignment");
    assert_eq!(
        report["message"],
        "Unexpected aliasing of 'this' to local variable."
    );
    assert_eq!(
        report["column"],
        utf16(source, source.find("self").unwrap()) + 1
    );
    assert_eq!(
        report["endColumn"],
        utf16(source, source.find("unknown").unwrap() + "unknown".len()) + 1
    );
    assert!(report.get("fix").is_none());
    assert!(report.get("suggestions").is_none());
    let source = "let self; ((self)) ||= (this);";
    let report = &messages(THIS, source)[0];
    assert_eq!(report["column"], source.rfind("self").unwrap() + 1);
    assert_eq!(report["endColumn"], source.rfind("self").unwrap() + 5);
}

#[test]
fn array_rewrite_preserves_argument_comments_and_applies_all_passes() {
    let source = "const emoji = '😀'; new Array( /*a*/ first, /*b*/ second, /*c*/ );";
    let start = source.find("new Array").unwrap();
    let end = source.rfind(')').unwrap() + 1;
    let diagnostics = messages(ARRAY, source);
    assert_eq!(diagnostics.len(), 1);
    let report = &diagnostics[0];
    assert_eq!(report["messageId"], "useLiteral");
    assert_eq!(
        report["message"],
        "The array literal notation [] is preferable."
    );
    assert_eq!(report["column"], utf16(source, start) + 1);
    assert_eq!(report["endColumn"], utf16(source, end) + 1);
    assert_eq!(
        report["fix"],
        fix(source, [start, end], "[ /*a*/ first, /*b*/ second, /*c*/ ]")
    );
    assert!(report.get("suggestions").is_none());
    let fixed = lint_files(ARRAY, &[("input.ts", source)], true);
    assert_eq!(fixed[0]["messages"], json!([]));
    assert_eq!(
        fixed[0]["output"],
        "const emoji = '😀'; [ /*a*/ first, /*b*/ second, /*c*/ ];"
    );
    let fixed = lint_files(
        ARRAY,
        &[("input.ts", "Array(Array(), Array()); new Array;")],
        true,
    );
    assert_eq!(fixed[0]["messages"], json!([]));
    assert_eq!(fixed[0]["output"], "[[], []]; [];");
}
