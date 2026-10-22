//! Regression cases for enum literals, declaration syntax, and binder merges.
//! No diagnostic, admitted file, fix, or suggestion is discarded by the checks.

use rslint_core::lint_json;
use serde_json::{Value, json};

const DUPLICATE: &str = "@typescript-eslint/no-duplicate-enum-values";
const NEW: &str = "@typescript-eslint/no-misused-new";
const MERGING: &str = "@typescript-eslint/no-unsafe-declaration-merging";

fn lint_files(rule: &str, sources: &[(&str, &str)]) -> Vec<Value> {
    let input = json!({
        "schemaVersion": 1,
        "cwd": "/rslint-ts-structural-tests",
        "files": sources.iter().map(|(path,text)| json!({"path":path,"text":text})).collect::<Vec<_>>(),
        "rules": [{"name":rule}]
    });
    let result: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let files = result["files"].as_array().unwrap();
    assert_eq!(files.len(), sources.len(), "retain silent admitted files");
    files
        .iter()
        .zip(sources)
        .map(|(file, (path, _))| {
            assert_eq!(
                file["filePath"],
                format!("/rslint-ts-structural-tests/{path}")
            );
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

fn diagnostic(
    rule: &str,
    source: &str,
    text: &str,
    occurrence: usize,
    id: &str,
    message: &str,
) -> Value {
    let start = source.match_indices(text).nth(occurrence).unwrap().0;
    let end = start + text.len();
    fn position(source: &str, offset: usize) -> (usize, usize, usize) {
        let prefix = &source[..offset];
        (
            prefix.encode_utf16().count(),
            prefix.bytes().filter(|b| *b == b'\n').count() + 1,
            prefix.rsplit('\n').next().unwrap().encode_utf16().count() + 1,
        )
    }
    let (start, line, column) = position(source, start);
    let (end, end_line, end_column) = position(source, end);
    json!({"ruleId":rule,"messageId":id,"message":message,"severity":2,
        "line":line,"column":column,"endLine":end_line,"endColumn":end_column,"range":[start,end]})
}

fn duplicate(source: &str, text: &str, value: &str) -> Value {
    diagnostic(
        DUPLICATE,
        source,
        text,
        0,
        "duplicateValue",
        &format!("Duplicate enum member value {value}."),
    )
}

fn merging(source: &str, occurrence: usize) -> Value {
    diagnostic(
        MERGING,
        source,
        "Foo",
        occurrence,
        "unsafeMerging",
        "Unsafe declaration merging between classes and interfaces.",
    )
}

#[test]
fn duplicate_enum_values_preserve_literal_kinds_and_unary_js_conversions() {
    let cases = [
        ("enum E { A = ((1)), B = 1 }", "B = 1", "1"),
        ("enum E { A = +(('1')), B = 1 }", "B = 1", "1"),
        ("enum E { A = +'0x10', B = 16 }", "B = 16", "16"),
        ("enum E { A = +'0b10', B = 2 }", "B = 2", "2"),
        ("enum E { A = +'0o10', B = 8 }", "B = 8", "8"),
        ("enum E { A = +'  ', B = 0 }", "B = 0", "0"),
        (
            "enum E { A = +'\\uFEFF\\u20001\\u2028', B = 1 }",
            "B = 1",
            "1",
        ),
        (
            "enum E { A = +'Infinity', B = 1e9999 }",
            "B = 1e9999",
            "Infinity",
        ),
        (
            "enum E { A = -'Infinity', B = -1e9999 }",
            "B = -1e9999",
            "-Infinity",
        ),
        ("enum E { A = +'.0000001', B = 1e-7 }", "B = 1e-7", "1e-7"),
        (
            "enum E { A = +'.000001', B = 1e-6 }",
            "B = 1e-6",
            "0.000001",
        ),
        ("enum E { A = +'1e21', B = 1e21 }", "B = 1e21", "1e+21"),
        (
            "enum E { A = +'0xffffffffffffffff', B = 18446744073709552000 }",
            "B = 18446744073709552000",
            "18446744073709552000",
        ),
        ("enum E { A = 1_000, B = 1000 }", "B = 1000", "1000"),
        ("enum E { A = '\\x41', B = 'A' }", "B = 'A'", "A"),
        ("enum E { A = `\\u0042`, B = 'B' }", "B = 'B'", "B"),
        ("enum E { A = -(-2), B = +2 }", "B = +2", "2"),
        ("enum E { A = -(+2), B = -2 }", "B = -2", "-2"),
    ];
    let sources: Vec<_> = cases.iter().map(|case| case.0).collect();
    for ((source, text, value), messages) in cases.iter().zip(lint_cases(DUPLICATE, &sources)) {
        assert_eq!(
            messages,
            json!([duplicate(source, text, value)]),
            "{source}"
        );
    }
}

#[test]
fn duplicate_enum_values_keep_nonliteral_expressions_and_nan_silent() {
    let sources = [
        "enum E { A, B, C = 1 }",
        "enum E { A = +'not a number', B = +'not a number' }",
        "enum E { A = +'-0x1', B = -1 }",
        "enum E { A = +(1 + 1), B = 2 }",
        "enum E { A = +'-0', B = 0 }",
        "enum E { A = '0', B = 0 }",
        "enum E { A = 'Infinity', B = 1e9999 }",
        "enum E { A = 1n, B = 1n }",
        "enum E { A = ~1, B = ~1 }",
        "enum E { A = 1 as number, B = 1 }",
        "enum E { A = +('1_000'), B = 1000 }",
        "enum E { A = +'\\u00851', B = 1 }",
        "enum E { A = +'Inf', B = 1e9999 }",
        "enum E { A = `a${1}`, B = `a${1}` }",
    ];
    for (source, messages) in sources.iter().zip(lint_cases(DUPLICATE, &sources)) {
        assert_eq!(messages, json!([]), "{source}");
    }
}

#[test]
fn public_entry_reports_invalid_template_syntax_without_running_rules() {
    // Isolated rule semantics remain separately tested with the actual parser.
    let input = json!({"schemaVersion":1,"cwd":"/rslint-ts-structural-tests",
        "files":[{"path":"invalid-template.ts","text":r"enum E { A = `\unicode`, B = `\unicode` }"}],
        "rules":[{"name":DUPLICATE}]});
    let result: Value = serde_json::from_str(&lint_json(&input.to_string()).unwrap()).unwrap();
    assert_eq!(result["files"][0]["hasSyntaxErrors"], true);
    assert_eq!(result["files"][0]["appliedRules"], json!([]));
    let messages = result["files"][0]["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 2);
    for message in messages {
        assert_eq!(message["ruleId"], "TypeScript(TS1125)");
        assert_eq!(message["message"], "Hexadecimal digit expected.");
        assert_eq!(message["severity"], 2);
        assert!(message.get("messageId").is_none());
        assert!(message.get("fix").is_none());
    }
}

#[test]
fn duplicate_enum_values_distinguish_signed_zero_through_overflow_storage() {
    let source = "enum E { S = '0', P = 0, N = -0, A = 1, B = 2, C = 3, D = 4, E = 5, F = 6, G = 7, H = 8, DS = '0', DP = +0, DN = -0, MN = -1, DMN = -1, NN = -(-0), PN = +-0 }";
    assert_eq!(
        lint_cases(DUPLICATE, &[source])[0],
        json!([
            duplicate(source, "DS = '0'", "0"),
            duplicate(source, "DP = +0", "0"),
            duplicate(source, "DN = -0", "0"),
            duplicate(source, "DMN = -1", "-1"),
            duplicate(source, "NN = -(-0)", "0"),
            duplicate(source, "PN = +-0", "0"),
        ])
    );
}

#[test]
fn enum_diagnostics_keep_utf16_positions_and_declaration_local_state() {
    let source = "/* 😀 */ enum E { A = 'same', B = 'same' }\nenum E { C = 'same' }";
    let result = lint_files(
        DUPLICATE,
        &[
            ("a.ts", source),
            ("b.ts", "enum E { D = 'same' }"),
            ("silent.ts", "const untouched = 1;"),
        ],
    );
    assert_eq!(result[0], json!([duplicate(source, "B = 'same'", "same")]));
    assert_eq!(result[1], json!([]));
    assert_eq!(result[2], json!([]));
}

#[test]
fn misused_new_preserves_parent_type_and_static_name_checks() {
    let cases = [
        (
            "class C { new(): C; }",
            "new(): C;",
            "errorMessageClass",
            "Class cannot have method named `new`.",
        ),
        (
            "declare abstract class C { ['new'](): C; }",
            "['new'](): C;",
            "errorMessageClass",
            "Class cannot have method named `new`.",
        ),
        (
            "const c = class C { [`new`](): C; };",
            "[`new`](): C;",
            "errorMessageClass",
            "Class cannot have method named `new`.",
        ),
        (
            "interface I { new (): I; }",
            "new (): I;",
            "errorMessageInterface",
            "interfaces cannot be constructed, only classes.",
        ),
        (
            "interface G { new <T>(): G<T>; }",
            "new <T>(): G<T>;",
            "errorMessageInterface",
            "interfaces cannot be constructed, only classes.",
        ),
        (
            "type T = { constructor(): void; };",
            "constructor(): void;",
            "errorMessageInterface",
            "interfaces cannot be constructed, only classes.",
        ),
        (
            "interface I { ['constructor'](): ''; }",
            "['constructor'](): '';",
            "errorMessageInterface",
            "interfaces cannot be constructed, only classes.",
        ),
        (
            "interface I { [`constructor`](): void; }",
            "[`constructor`](): void;",
            "errorMessageInterface",
            "interfaces cannot be constructed, only classes.",
        ),
    ];
    let sources: Vec<_> = cases.iter().map(|case| case.0).collect();
    for ((source, text, id, message), messages) in cases.iter().zip(lint_cases(NEW, &sources)) {
        assert_eq!(
            messages,
            json!([diagnostic(NEW, source, text, 0, id, message)]),
            "{source}"
        );
    }
}

#[test]
fn misused_new_does_not_expand_to_implementations_qualified_or_wrapped_types() {
    let sources = [
        "declare abstract class C { foo() {} get new(); bar(); }",
        "class C { constructor(); }",
        "const c = class { new(): X; };",
        "class C { new(): C { return this; } }",
        "class C { [`new`](): C {} }",
        "interface I { new (): {}; }",
        "type T = { new (): T };",
        "export default class { new(): X; }",
        "interface I { new <T>(): Other<T>; }",
        "interface I { new <T>(): 'x'; }",
        "declare class C { [methodName](): C; new(): Namespace.C; new(): (C); }",
        "declare class C { ['new' as string](): C; ['n' + 'ew'](): C; 'new'(): C; }",
        "interface I { ['constructor' as string](): void; 'constructor'(): void; }",
    ];
    for (source, messages) in sources.iter().zip(lint_cases(NEW, &sources)) {
        assert_eq!(messages, json!([]), "{source}");
    }
}

#[test]
fn declaration_merging_checks_same_scope_and_default_export_local_symbols() {
    let sources = [
        "interface Foo {} class Foo {}",
        "{ interface Foo {} class Foo {} }",
        "namespace N { interface Foo {} class Foo {} }",
        "function outer() { interface Foo {} class Foo {} }",
        "const outer = () => { interface Foo {} class Foo {} };",
        "class Holder { static { interface Foo {} class Foo {} } }",
        "declare module 'example' { interface Foo {} class Foo {} }",
        "export interface Foo {} export class Foo {}",
        "interface Foo {} export default class Foo {}",
        "export default class Foo {} interface Foo {}",
        "/* 😀 */ interface Foo {}\nabstract class Foo {}",
    ];
    for (source, messages) in sources.iter().zip(lint_cases(MERGING, &sources)) {
        assert_eq!(
            messages,
            json!([merging(source, 0), merging(source, 1)]),
            "{source}"
        );
    }
}

#[test]
fn declaration_merging_preserves_generic_interface_listener_asymmetry() {
    let cases = [
        ("interface Foo<T> {} class Foo {}", 1),
        ("class Foo<T> {} interface Foo<T> {}", 0),
        (
            "interface Foo<T extends string> {} class Foo<T extends string> {}",
            1,
        ),
        ("interface Foo<T = number> {} class Foo {}", 1),
    ];
    let sources: Vec<_> = cases.iter().map(|case| case.0).collect();
    for ((source, occurrence), messages) in cases.iter().zip(lint_cases(MERGING, &sources)) {
        assert_eq!(messages, json!([merging(source, *occurrence)]), "{source}");
    }
}

#[test]
fn declaration_merging_keeps_distinct_bindings_and_reopened_scopes_silent() {
    let sources = [
        "interface Foo {} const Other = class Foo {};",
        "interface Foo {} export default class {}",
        "interface Foo {} function f() { class Foo {} }",
        "class Foo {} const f = () => { interface Foo {} };",
        "class Holder { static { interface Foo {} } } class Foo {}",
        "namespace A { class Foo {} } namespace B { interface Foo {} }",
        "namespace N { export interface Foo {} } namespace N { export class Foo {} }",
        "{ interface Foo {} } { class Foo {} }",
        "class Foo {} namespace Foo { export const value = 1; }",
        "interface Foo {} namespace Foo { export const value = 1; }",
        "interface Foo {} interface Foo { value: number }",
        "declare module 'lib' { interface Foo {} } declare module 'lib' { class Foo {} }",
        "declare global { interface Foo {} } class Foo {}",
        "declare global { class Foo {} } namespace nested { interface Foo {} }",
    ];
    for (source, messages) in sources.iter().zip(lint_cases(MERGING, &sources)) {
        assert_eq!(messages, json!([]), "{source}");
    }
}

#[test]
fn declaration_merging_caches_large_symbols_without_losing_scopes_or_reports() {
    let mut source = "class Foo {}\n".to_owned();
    source.push_str(&"interface Foo {}\n".repeat(12));
    let messages = lint_cases(MERGING, &[&source]);
    let expected: Vec<_> = (0..13).map(|index| merging(&source, index)).collect();
    assert_eq!(messages[0], json!(expected));

    let mut reopened = "namespace N { export class Foo {} }\nnamespace N {\n".to_owned();
    reopened.push_str(&"export interface Foo {}\n".repeat(12));
    reopened.push('}');
    assert_eq!(lint_cases(MERGING, &[&reopened])[0], json!([]));
}

#[test]
fn declaration_merging_distinguishes_cross_file_symbols_and_module_augmentation() {
    let result = lint_files(
        MERGING,
        &[
            ("global-class.ts", "class Foo {}"),
            ("global-interface.ts", "interface Foo {}"),
            ("library.ts", "export class Service {}"),
            (
                "augment.ts",
                "import { Service } from './library'; declare module './library' { interface Service { additional: string } }",
            ),
            ("silent.ts", "export const untouched = 1;"),
        ],
    );
    for messages in result {
        assert_eq!(messages, json!([]));
    }
}
