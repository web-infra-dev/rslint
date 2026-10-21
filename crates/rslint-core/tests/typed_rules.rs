//! Native checker regression cases for the optionless type-aware rule ports.
//! Expected diagnostics include exact messages, full ranges, and silent files.

use rslint_core::lint_json;
use serde_json::{Value, json};

const MIXED: &str = "@typescript-eslint/no-mixed-enums";
const COMPARISON: &str = "@typescript-eslint/no-unsafe-enum-comparison";
const ACCESSORS: &str = "@typescript-eslint/related-getter-setter-pairs";

fn lint_files(rule: &str, sources: &[(&str, &str)]) -> Vec<Value> {
    let input = json!({
        "schemaVersion":1,
        "cwd":"/rslint-typed-tests",
        "files":sources.iter().map(|(path,text)|json!({"path":path,"text":text})).collect::<Vec<_>>(),
        "rules":[{"name":rule}]
    });
    let output: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(output["backend"], "tsrs-rust");
    let files = output["files"].as_array().unwrap();
    assert_eq!(
        files.len(),
        sources.len(),
        "silent files must stay admitted"
    );
    files
        .iter()
        .zip(sources)
        .map(|(file, (path, _))| {
            assert_eq!(file["filePath"], format!("/rslint-typed-tests/{path}"));
            assert_eq!(file["appliedRules"], json!([rule]));
            file["messages"].clone()
        })
        .collect()
}

fn lint_cases(rule: &str, sources: &[&str]) -> Vec<Value> {
    let files: Vec<_> = sources
        .iter()
        .enumerate()
        .map(|(index, source)| {
            (
                format!("case-{index}.ts"),
                format!("{source}\nexport {{}};"),
            )
        })
        .collect();
    lint_files(
        rule,
        &files
            .iter()
            .map(|(path, text)| (path.as_str(), text.as_str()))
            .collect::<Vec<_>>(),
    )
}

fn expected(rule: &str, source: &str, text: &str, id: &str, message: &str) -> Value {
    let start = source.find(text).unwrap();
    let end = start + text.len();
    expected_at(rule, source, start, end, id, message)
}

fn expected_at(
    rule: &str,
    source: &str,
    start: usize,
    end: usize,
    id: &str,
    message: &str,
) -> Value {
    fn location(source: &str, offset: usize) -> (usize, usize, usize) {
        let prefix = &source[..offset];
        let range = prefix.encode_utf16().count();
        let line = prefix.bytes().filter(|b| *b == b'\n').count() + 1;
        let column = prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1;
        (range, line, column)
    }
    let (start, line, column) = location(source, start);
    let (end, end_line, end_column) = location(source, end);
    json!({"ruleId":rule,"messageId":id,"message":message,"severity":2,
        "line":line,"column":column,"endLine":end_line,"endColumn":end_column,"range":[start,end]})
}

fn mixed(source: &str, text: &str) -> Value {
    expected(
        MIXED,
        source,
        text,
        "mixed",
        "Mixing number and string enums can be confusing.",
    )
}

fn comparison(source: &str, text: &str) -> Value {
    expected(
        COMPARISON,
        source,
        text,
        "mismatchedCondition",
        "The two values in this comparison do not have a shared enum type.",
    )
}

fn accessor(source: &str, text: &str) -> Value {
    expected(
        ACCESSORS,
        source,
        text,
        "mismatch",
        "`get()` type should be assignable to its equivalent `set()` type.",
    )
}

#[test]
fn mixed_enums_use_computed_initializer_types_and_merged_declarations() {
    let sources = [
        "enum E { A, B = 'b', C = 'c' }",
        "const getValue = () => 1; enum E { A = getValue(), B = 'b' }",
        "enum E { A = 'a' } enum E { B = 2 }",
        "enum E { A = `a`, B }",
        "/* 😀 */ enum E { A = 0, B = 'b' }",
    ];
    let outputs = lint_cases(MIXED, &sources);
    for ((source, output), text) in sources
        .iter()
        .zip(outputs)
        .zip(["'b'", "'b'", "2", "B", "'b'"])
    {
        assert_eq!(output, json!([mixed(source, text)]), "{source}");
    }
}

#[test]
fn mixed_enums_stop_at_unknown_members_and_keep_consistent_enums_silent() {
    let sources = [
        "enum E {}",
        "enum E { A, B, C = 2 }",
        "enum E { A = 'a', B = `b` }",
        "const getValue = () => ''; enum E { A = getValue(), B = 'b' }",
        "enum E { A = false, B = 'b' }",
        "enum E { A, B = false, C = 'c' }",
        "enum E { A = 0 } enum E { B = 1 }",
        "namespace Outer { namespace Inner { export enum E { A = 1 } } } namespace Outer { namespace Inner { export enum E { B = 'b' } } }",
    ];
    for (source, output) in sources.iter().zip(lint_cases(MIXED, &sources)) {
        assert_eq!(output, json!([]), "{source}");
    }
}

#[test]
fn mixed_enums_query_cross_file_types_and_global_declaration_merging() {
    let imported = "import { value } from './values'; enum E { A = 0, B = value }";
    let merged = "enum Shared { B = 'b' }";
    let sources = [
        ("values.ts", "export const value = 'text';"),
        ("main.ts", imported),
        ("first.ts", "enum Shared { A = 1 }"),
        ("second.ts", merged),
        ("silent.ts", "export const silent = true;"),
    ];
    let results = lint_files(MIXED, &sources);
    assert_eq!(results[0], json!([]));
    let start = imported.rfind("value").unwrap();
    assert_eq!(
        results[1],
        json!([expected_at(
            MIXED,
            imported,
            start,
            start + 5,
            "mixed",
            "Mixing number and string enums can be confusing."
        )])
    );
    assert_eq!(results[2], json!([]));
    assert_eq!(results[3], json!([mixed(merged, "'b'")]));
    assert_eq!(results[4], json!([]));
}

#[test]
fn enum_comparisons_cover_every_comparison_operator_and_both_directions() {
    let sources: Vec<_> = ["<", "<=", ">", ">=", "==", "===", "!=", "!=="]
        .iter()
        .map(|operator| format!("enum E {{ A = 1 }} E.A {operator} 1;"))
        .collect();
    let refs: Vec<_> = sources.iter().map(String::as_str).collect();
    for (source, output) in sources.iter().zip(lint_cases(COMPARISON, &refs)) {
        let expression = source.rsplit("} ").next().unwrap().trim_end_matches(';');
        assert_eq!(output, json!([comparison(source, expression)]), "{source}");
    }
    let source = "enum E { A = 'a' } 'a' === E.A;";
    assert_eq!(
        lint_cases(COMPARISON, &[source])[0],
        json!([comparison(source, "'a' === E.A")])
    );
}

#[test]
fn enum_comparisons_preserve_union_intersection_and_shared_type_semantics() {
    let valid = [
        "'a' === 'b'; 1 === 2;",
        "enum E { A = 0, B = 1 } E.A === E.B;",
        "enum E { A = 'a' } declare const value: E | 'a'; value === 'a';",
        "enum E { A = 1 } declare const value: E | 0; value === 0;",
        "enum E { A = 1 } declare const value: E | number; value === 0;",
        "enum E { A = 1 } E.A === ({} as any); E.A === undefined; E.A === null;",
        "enum E { A = 1 } enum F { A = 2 } declare const value: E | F; value === E.A;",
        "enum E { A = 1 } function f<T extends E>(value: T) { return value === E.A; }",
        "enum E { A = 1 } E.A + 1; E.A & 1;",
    ];
    for (source, output) in valid.iter().zip(lint_cases(COMPARISON, &valid)) {
        assert_eq!(output, json!([]), "{source}");
    }
    let invalid = [
        "enum E { A = 1 } enum F { A = 2 } E.A === F.A;",
        "enum E { A = 1 } declare const value: number & { readonly brand: unique symbol }; E.A === value;",
        "enum E { A = 'a' } declare const value: 'a' | 'b'; E.A === value;",
    ];
    for ((source, output), expression) in invalid
        .iter()
        .zip(lint_cases(COMPARISON, &invalid))
        .zip(["E.A === F.A", "E.A === value", "E.A === value"])
    {
        assert_eq!(output, json!([comparison(source, expression)]), "{source}");
    }
}

#[test]
fn enum_switch_cases_and_imported_enums_have_exact_diagnostics_without_fixes() {
    let source = "import { E } from './types';\nconst x: E = E.A;\nswitch (x) { case 1: break; case E.A: break; }\nE.A === 1;";
    let results = lint_files(
        COMPARISON,
        &[
            ("types.ts", "export enum E { A = 1 }"),
            ("main.ts", source),
            ("clean.ts", "export const clean = 1;"),
        ],
    );
    assert_eq!(results[0], json!([]));
    assert_eq!(results[2], json!([]));
    assert_eq!(
        results[1],
        json!([
            expected(
                COMPARISON,
                source,
                "case 1: break;",
                "mismatchedCase",
                "The case statement does not have a shared enum type with the switch predicate."
            ),
            comparison(source, "E.A === 1")
        ])
    );
}

#[test]
fn accessors_use_full_assignability_and_both_declaration_orders() {
    let sources = [
        "interface Example { get value(): string | undefined; set value(value: string); }",
        "type Example = { set value(value: string); get value(): number; };",
        "class Example { get value(): boolean { return true; } set value(value: string) {} }",
        "const Example = class { get value(): number { return 1; } set value(value: string) {} };",
        "interface Example<T extends string> { get value(): string; set value(value: T); }",
        "interface Example { get value(): { a: string; b: string }; set value(value: { c: string }); }",
    ];
    for ((source, output), text) in sources.iter().zip(lint_cases(ACCESSORS, &sources)).zip([
        "string | undefined",
        "number",
        "boolean",
        "number",
        "string",
        "{ a: string; b: string }",
    ]) {
        let start = source.find("get value(): ").unwrap() + "get value(): ".len();
        assert_eq!(&source[start..start + text.len()], text);
        assert_eq!(
            output,
            json!([expected_at(
                ACCESSORS,
                source,
                start,
                start + text.len(),
                "mismatch",
                "`get()` type should be assignable to its equivalent `set()` type."
            )]),
            "{source}"
        );
    }
}

#[test]
fn accessors_keep_untyped_and_missing_or_invalid_pairs_silent() {
    let sources = [
        "interface E { get value(): string; set value(value: string | undefined); }",
        "interface E<T extends string> { get value(): T; set value(value: string); }",
        "interface E { get value(): number; }",
        "interface E { set value(value: string); }",
        "interface E { get value(): number; set value(); }",
        "interface E { get value(): number; set value(a: string, b: string); }",
        "class E { get value() { return 1; } set value(value: string) {} }",
        "type E = { get a(): number; set b(value: string); };",
        "const value = { get a(): number { return 1; }, set a(value: string) {} };",
    ];
    for (source, output) in sources.iter().zip(lint_cases(ACCESSORS, &sources)) {
        assert_eq!(output, json!([]), "{source}");
    }
}

#[test]
fn accessor_names_and_imported_types_preserve_existing_matching() {
    let source = "import { Read, Write } from './types';\ninterface Example { get value(): Read; set value(value: Write); }";
    let result = lint_files(
        ACCESSORS,
        &[
            (
                "types.ts",
                "export type Read = { a: string }; export type Write = { b: string };",
            ),
            ("main.ts", source),
        ],
    );
    assert_eq!(result[0], json!([]));
    // The first Read is in the import; the reported annotation is the second.
    let start = source.find("Read; set").unwrap();
    let report = expected_at(
        ACCESSORS,
        source,
        start,
        start + 4,
        "mismatch",
        "`get()` type should be assignable to its equivalent `set()` type.",
    );
    assert_eq!(result[1], json!([report]));

    let sources = [
        "interface E { get ['value'](): number; set value(value: string); }",
        "class E { get #value(): number { return 1; } set #value(value: string) {} }",
        "interface E { get ['not-valid'](): number; set \"not-valid\"(value: string); }",
    ];
    for (source, output) in sources.iter().zip(lint_cases(ACCESSORS, &sources)) {
        assert_eq!(output, json!([accessor(source, "number")]));
    }
}
