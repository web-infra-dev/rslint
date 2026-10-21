//! Control-flow cases through the public Rust boundary. Angle markers identify
//! the complete diagnostic spans; markers are removed before parsing.

use rslint_core::lint_json;
use serde_json::{Value, json};

fn fixture(rule: &str, marked: &str, id: &str, text: &str) -> (String, Vec<Value>) {
    let mut source = String::new();
    let mut start = None;
    let mut spans = Vec::new();
    for character in marked.chars() {
        match character {
            '«' => {
                assert!(start.replace(source.len()).is_none());
            }
            '»' => spans.push((
                start.take().expect("unmatched closing marker"),
                source.len(),
            )),
            _ => source.push(character),
        }
    }
    assert!(start.is_none(), "unclosed marker: {marked}");
    let messages = spans
        .into_iter()
        .map(|(start, end)| {
            let before = &source[..start];
            let through = &source[..end];
            json!({
                "ruleId":rule,"messageId":id,"message":text,"severity":2,
                "line":before.split('\n').count(),
                "column":before.rsplit('\n').next().unwrap().encode_utf16().count()+1,
                "endLine":through.split('\n').count(),
                "endColumn":through.rsplit('\n').next().unwrap().encode_utf16().count()+1,
                "range":[before.encode_utf16().count(),through.encode_utf16().count()]
            })
        })
        .collect();
    (source, messages)
}

fn run_rule(rule: &str, valid: &[&str], invalid: &[(&str, &str, &str)]) {
    let cases: Vec<_> = valid
        .iter()
        .map(|source| fixture(rule, source, "", ""))
        .chain(
            invalid
                .iter()
                .map(|(source, id, text)| fixture(rule, source, id, text)),
        )
        .collect();
    assert!(!cases.is_empty());
    let input_files: Vec<_> = cases
        .iter()
        .enumerate()
        .map(|(index, (source, _))| json!({"path":format!("case-{index}.ts"),"text":source}))
        .collect();
    let request = json!({
        "schemaVersion":1,"cwd":"/rslint-control-flow-tests",
        "files":input_files,"rules":[{"name":rule}]
    });
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let files = result["files"].as_array().unwrap();
    assert_eq!(
        files.len(),
        cases.len(),
        "silent files must remain admitted"
    );
    for (index, ((source, messages), file)) in cases.iter().zip(files).enumerate() {
        assert_eq!(
            file,
            &json!({
                "filePath":format!("/rslint-control-flow-tests/case-{index}.ts"),
                "messages":messages,"appliedRules":[rule]
            }),
            "{rule}: {source}"
        );
    }
}

#[test]
fn continue_statements_keep_labels_and_complete_ranges() {
    let invalid = [
        "var sum = 0, i; for(i = 0; i < 10; i++){ if(i <= 5) { «continue;» } sum += i; }",
        "var sum = 0, i; myLabel: for(i = 0; i < 10; i++){ if(i <= 5) { «continue myLabel;» } sum += i; }",
        "var sum = 0, i = 0; while(i < 10) { if(i <= 5) { i++; «continue;» } sum += i; i++; }",
        "var sum = 0, i = 0; myLabel: while(i < 10) { if(i <= 5) { i++; «continue myLabel;» } sum += i; i++; }",
        "for (;;) {\n  /* 🚀 */ «continue;»\n}",
        "label: while (true) { if (x) { «continue label;» } «continue;» }",
        "for (const x of items) { «continue;» }",
        "for (const x in items) { «continue;» }",
        "do { «continue;» } while (true);",
        "while (true) { «continue»\n}",
    ];
    run_rule(
        "no-continue",
        &[
            "var sum = 0, i; for(i = 0; i < 10; i++){ if(i > 5) { sum += i; } }",
            "var sum = 0, i = 0; while(i < 10) { if(i > 5) { sum += i; } i++; }",
            "const object = { continue: true }; object.continue; const value = 'continue';",
        ],
        &invalid.map(|source| {
            (
                source,
                "unexpected",
                "Unexpected use of continue statement.",
            )
        }),
    );
}

#[test]
fn awaits_respect_each_loop_position_and_function_boundary() {
    let invalid = [
        "async function foo() { while (baz) { «await bar»; } }",
        "async function foo() { while («await foo()») { } }",
        "async function foo() { while (baz) { for «await» (x of xs); } }",
        "async function foo() { do { «await bar»; } while (baz); }",
        "async function foo() { do { } while («await bar»); }",
        "async function foo() { for (var bar of baz) { «await bar»; } }",
        "async function foo() { for (var bar in baz) { «await bar»; } }",
        "async function foo() { for (var i = 0; «await foo(i)»; i++) { } }",
        "async function foo() { for (var i = 0; i < n; i = «await bar») { } }",
        "async function foo() { for (var i = 0; i < n; i++) { «await bar»; } }",
        "async function foo() { while (true) { if (bar) { «await baz»; } } }",
        "async function foo() { while (true) { while (true) { «await bar»; } } }",
        "async function foo() { for await (const x of xs) { while (true) { «await x»; } } }",
        "async function foo() { while (true) { for «await» (const x of xs) { await x; } } }",
    ];
    run_rule(
        "no-await-in-loop",
        &[
            "async function foo() { await bar; }",
            "async function foo() { for (var bar in await baz) { } }",
            "async function foo() { for (var bar of await baz) { } }",
            "async function foo() { for await (var bar of await baz) { } }",
            "async function foo() { for await (var x of xs) { await f(x) } }",
            "async function foo() { for (var i = await bar; i < n; i++) { } }",
            "async function foo() { while (true) { async function bar() { await baz; } } }",
            "async function foo() { while (true) { var bar = async () => { await baz; } } }",
            "async function foo() { while (true) { var bar = async function() { await baz; } } }",
            "async function foo() { while (true) { class bar { async baz() { await qux; } } } }",
            "async function foo() { for (var i = 0; i < 10; i++) { } }",
            "async function foo() { for (var i = 0; i < 10; i++) { bar(); } }",
            // Preserve the baseline's nearest-loop initializer behavior.
            "async function foo() { while (true) { for (let x = await bar; x; x--) {} } }",
        ],
        &invalid.map(|source| {
            (
                source,
                "unexpectedAwait",
                "Unexpected `await` inside a loop.",
            )
        }),
    );
}

#[test]
fn constructor_returns_preserve_the_baseline_boundary_set() {
    let invalid = [
        "class C { constructor() { «return ''» } }",
        "class C { constructor(a) { if (!a) { «return ''» } else { a() } } }",
        "class C { constructor() { «return 1» } }",
        "class C { constructor() { «return {}» } }",
        "class C { constructor() { «return null» } }",
        "class C { constructor() { «return undefined» } }",
        "class C { constructor() { «return true» } }",
        "class C { constructor(x) { if (x) «return 'yes';» «return 'no';» } }",
        "class C { constructor() { «return []» } }",
        "class C { constructor() { «return foo()» } }",
        "class C { constructor() { «return this» } }",
        "class C { constructor(private x: number) { «return 1» } }",
        // Get/set accessors are absent from the baseline's explicit sentinel set.
        "class C { constructor() { const o = { get x() { «return 1;» } }; } }",
    ];
    run_rule(
        "no-constructor-return",
        &[
            "function fn() { return }",
            "function fn(value) { if (value) { return value } }",
            "const fn = function () { return }",
            "const fn = function () { if (value) { return value } }",
            "const fn = () => { return }",
            "const fn = () => { if (value) { return value } }",
            "class C {}",
            "class C { constructor() {} }",
            "class C { constructor() { let v } }",
            "class C { method() { return '' } }",
            "class C { get value() { return '' } }",
            "class C { constructor(a) { if (!a) { return } else { a() } } }",
            "class C { constructor() { return } }",
            "class C { constructor() { { return } } }",
            "class C { constructor() { function fn() { return true } } }",
            "class C { constructor() { this.fn = function () { return true } } }",
            "class C { constructor() { this.fn = () => { return true } } }",
            "class C { constructor() { const o = { method() { return true; } }; } }",
            "class C { constructor(); constructor(a?: string) {} }",
            "class C { constructor(a: string); constructor(a?: string) { if (!a) { return } } }",
            "class C { constructor(private x: number) {} }",
            "class C { constructor(public readonly y: string) { return } }",
        ],
        &invalid.map(|source| {
            (
                source,
                "unexpected",
                "Unexpected return statement in constructor.",
            )
        }),
    );
}

#[test]
fn negated_conditions_check_plain_else_and_transparent_parentheses() {
    let invalid = [
        "«if (!a) {;} else {;}»",
        "«if (a != b) {;} else {;}»",
        "«if (a !== b) {;} else {;}»",
        "«!a ? b : c»",
        "«a != b ? c : d»",
        "«a !== b ? c : d»",
        "«if ((!a)) {} else {}»",
        "«if (((!a))) {} else {}»",
        "«if ((a != b)) {} else {}»",
        "«if (!a!) {} else {}»",
        "«if (!a?.b) {} else {}»",
        "if (a) {} else «if (!b) {} else {}»",
        "if (a) {} else if (b) {} else «if (!c) {} else {}»",
        "«if (!a) { if (b) {} else {} } else {}»",
        "a ? b : «!c ? d : e»",
        "«if (!a); else;»",
        "«if (!user?.isActive) { markInactive(); } else { markActive(); }»",
        "«if (!(value as MyType).isValid) { reject(); } else { accept(); }»",
    ];
    run_rule(
        "no-negated-condition",
        &[
            "if (a) {}",
            "if (a) {} else {}",
            "if (!a) {}",
            "if (!a) {} else if (b) {}",
            "if (!a) {} else if (b) {} else {}",
            "if (a == b) {}",
            "if (a == b) {} else {}",
            "if (a != b) {}",
            "if (a != b) {} else if (b) {}",
            "if (a != b) {} else if (b) {} else {}",
            "if (a !== b) {}",
            "if (a === b) {} else {}",
            "a ? b : c",
            "if ((a != b) as any) {} else {}",
            "if ((a != b) satisfies boolean) {} else {}",
            "if (-a) {} else {}",
            "if (typeof a) {} else {}",
            "if (a && b) {} else {}",
            "if (a || b) {} else {}",
        ],
        &invalid.map(|source| (source, "unexpectedNegated", "Unexpected negated condition.")),
    );
}

#[test]
fn finally_control_flow_uses_function_class_loop_switch_and_label_sentinels() {
    let valid = [
        "function f() { try { return 1; } catch (e) { return 2; } finally { work(); } }",
        "function f() { try {} finally { function g() { return 1; } } }",
        "function f() { try {} finally { const g = function() { return 1; }; } }",
        "function f() { try {} finally { const g = () => { return 1; }; } }",
        "function f() { try {} finally { const o = { method() { return 1; } }; } }",
        "function f() { try {} finally { const o = { get x() { return 1; } }; } }",
        "function f() { try {} finally { const o = { set x(v) { return; } }; } }",
        "function f() { try {} finally { class C { constructor() { return; } } } }",
        "function f() { try {} finally { class C { static { throw new Error(); } } } }",
        "function f() { try {} finally { const C = class { static { throw new Error(); } }; } }",
        "function f() { try {} finally { function g() { throw new Error(); } } }",
        "function f() { try {} finally { while (true) break; } }",
        "function f() { try {} finally { for (;;) break; } }",
        "function f() { try {} finally { for (const x in obj) break; } }",
        "function f() { try {} finally { for (const x of obj) break; } }",
        "function f() { try {} finally { do { break; } while (true); } }",
        "function f() { try {} finally { while (true) continue; } }",
        "function f() { try {} finally { for (;;) continue; } }",
        "function f() { try {} finally { for (const x in obj) continue; } }",
        "function f() { try {} finally { for (const x of obj) continue; } }",
        "function f() { try {} finally { do { continue; } while (true); } }",
        "function f() { try {} finally { switch (x) { case 1: break; } } }",
        "function f() { try {} finally { label: while (true) { break label; } } }",
        "function f() { try {} finally { label: while (true) { continue label; } } }",
        "function f() { try {} finally { label: { break label; } } }",
        // A loop inside finally stops labeled continue, but not labeled break.
        "label: while (true) { try {} finally { while (true) { continue label; } } }",
        "function f() { try { return 1; } catch(e) { return 2; } }",
        "function f() { try {} finally { try { while (true) { break; } } catch(e) {} } }",
    ];
    let mut invalid = Vec::new();
    for source in [
        "function f() { try { return 1; } finally { «return 3;» } }",
        "function f() { try {} finally { if (x) { «return;» } } }",
        "function f() { try {} finally { while (true) { «return 1;» } } }",
        "function f() { try {} finally { try { «return 1;» } finally { log(); } } }",
    ] {
        invalid.push((source, "unsafeUsage", "Unsafe usage of ReturnStatement."));
    }
    for source in [
        "function f() { try {} finally { «throw new Error();» } }",
        "function f() { try {} finally { switch (x) { case 1: «throw x;» } } }",
        "function f() { try {} finally { try {} catch(e) { «throw e;» } } }",
    ] {
        invalid.push((source, "unsafeUsage", "Unsafe usage of ThrowStatement."));
    }
    for source in [
        "while (true) try {} finally { «break;» }",
        "switch (x) { case 1: try {} finally { «break;» } }",
        "label: while (true) { try {} finally { «break label;» } }",
        "label: while (true) { try {} finally { while (true) { «break label;» } } }",
        "outer: { try {} finally { inner: { «break outer;» } } }",
    ] {
        invalid.push((source, "unsafeUsage", "Unsafe usage of BreakStatement."));
    }
    for source in [
        "while (true) try {} finally { «continue;» }",
        "label: while (true) { try {} finally { «continue label;» } }",
        "while (true) { try {} finally { switch (x) { case 1: «continue;» } } }",
    ] {
        invalid.push((source, "unsafeUsage", "Unsafe usage of ContinueStatement."));
    }
    run_rule("no-unsafe-finally", &valid, &invalid);
}

#[test]
fn useless_catch_preserves_first_statement_and_finally_report_targets() {
    let without_finally = [
        "«try { foo(); } catch (err) { throw err; }»",
        "«try { foo(); } catch (err) { /* comment */ throw err; }»",
        "«try { foo(); } catch (err: unknown) { throw err; }»",
        "«try { foo(); } catch (err) { throw (err); }»",
        "«try { foo(); } catch (err) { throw ((err)); }»",
        "«try { foo(); } catch (err) { throw err; unreachable(); }»",
    ];
    let mut invalid: Vec<_> = without_finally
        .map(|source| (source, "unnecessaryCatch", "Unnecessary try/catch wrapper."))
        .into_iter()
        .collect();
    invalid.push((
        "try { foo(); } «catch (err) { throw err; }» finally { foo(); }",
        "unnecessaryCatchClause",
        "Unnecessary catch clause.",
    ));
    invalid.push((
        "try { foo(); }\r\n«catch (err) {\r\n  throw (err);\r\n}» finally {}",
        "unnecessaryCatchClause",
        "Unnecessary catch clause.",
    ));
    run_rule(
        "no-useless-catch",
        &[
            "try { foo(); } catch (err) { console.error(err); }",
            "try { foo(); } catch (err) { console.error(err); } finally { bar(); }",
            "try { foo(); } catch (err) { doSomethingBeforeRethrow(); throw err; }",
            "try { foo(); } catch (err) { throw err.msg; }",
            "try { foo(); } catch (err) { throw new Error('whoops!'); }",
            "try { foo(); } catch (err) { throw bar; }",
            "try { foo(); } catch (err) {}",
            "try { foo(); } catch ({ err }) { throw err; }",
            "try { foo(); } catch ([ err ]) { throw err; }",
            "try { throw new Error('foo'); } catch { throw new Error('foo'); }",
            "try { foo(); } catch (err) { throw err!; }",
            "try { foo(); } catch (err) { throw err as Error; }",
            "try { foo(); } catch (err) { ; throw err; }",
        ],
        &invalid,
    );
}
