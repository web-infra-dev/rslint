//! Behavioral cases migrated from the existing syntax-rule suites. They use
//! the public JSON boundary and actual tsrs parser, including silent files.

use rslint_core::lint_json;
use serde_json::{Value, json};

fn lint(rule: &str, sources: &[&str]) -> Vec<Value> {
    let files: Vec<Value> = sources
        .iter()
        .enumerate()
        .map(|(index, source)| json!({"path":format!("case-{index}.ts"),"text":source}))
        .collect();
    let request = json!({
        "schemaVersion":1,
        "cwd":"/rslint-syntax-tests",
        "files":files,
        "rules":[{"name":rule}]
    });
    let result: Value = serde_json::from_str(&lint_json(&request.to_string()).unwrap()).unwrap();
    assert_eq!(result["backend"], "tsrs-rust");
    let files = result["files"].as_array().unwrap();
    assert_eq!(files.len(), sources.len(), "silent files must be retained");
    files
        .iter()
        .enumerate()
        .map(|(index, file)| {
            assert_eq!(
                file["filePath"],
                format!("/rslint-syntax-tests/case-{index}.ts")
            );
            assert_eq!(file["appliedRules"], json!([rule]));
            file["messages"].clone()
        })
        .collect()
}

fn valid(rule: &str, sources: &[&str]) {
    for (source, messages) in sources.iter().zip(lint(rule, sources)) {
        assert_eq!(messages, json!([]), "{rule}: {source}");
    }
}

fn message(rule: &str, id: &str, text: &str, start: usize, end: usize) -> Value {
    json!({"ruleId":rule,"messageId":id,"message":text,"severity":2,
        "line":1,"column":start,"endLine":1,"endColumn":end,"range":[start-1,end-1]})
}

#[test]
fn debugger_statements_include_computed_assignment_keys() {
    valid(
        "no-debugger",
        &[
            "var test = { debugger: 1 }; test.debugger;",
            "const text = 'debugger'; // debugger;",
        ],
    );
    let output = lint(
        "no-debugger",
        &[
            "if (foo) debugger",
            "({ [(() => { debugger; })()]: target } = source);",
            "/* 🚀 */ debugger;",
        ],
    );
    for (messages, (start, end)) in output.into_iter().zip([(10, 18), (14, 23), (10, 19)]) {
        assert_eq!(
            messages,
            json!([message(
                "no-debugger",
                "no-debugger",
                "Unexpected 'debugger' statement.",
                start,
                end
            )])
        );
    }
}

#[test]
fn caller_checks_dot_access_and_heritage_but_preserves_ts_wrappers() {
    valid(
        "no-caller",
        &[
            "arguments.length; arguments[0]; arguments[caller];",
            "arguments['callee']; arguments['caller'];",
            "obj.callee; obj.caller; arguments.called; arguments.call;",
            "(0, arguments).callee; (true ? arguments : null).callee;",
            "arguments!.callee; (arguments as any).callee; (<any>arguments).callee;",
            "type T = arguments.callee; type U = typeof arguments.caller;",
            "interface I extends Other<arguments.callee> {}",
        ],
    );
    let sources = [
        "arguments.callee;",
        "arguments.caller;",
        "var x = ((arguments)).callee;",
        "function f(arguments) { arguments?.caller; }",
        "interface I extends arguments.callee {}",
        "interface I extends arguments.callee.Inner {}",
    ];
    let expected = [
        ("callee", 1, 17),
        ("caller", 1, 17),
        ("callee", 9, 29),
        ("caller", 25, 42),
        ("callee", 21, 37),
        ("callee", 21, 37),
    ];
    for ((source, messages), (name, start, end)) in sources
        .iter()
        .zip(lint("no-caller", &sources))
        .zip(expected)
    {
        assert_eq!(
            messages,
            json!([message(
                "no-caller",
                "unexpected",
                &format!("Avoid arguments.{name}."),
                start,
                end
            )]),
            "{source}"
        );
    }
}

#[test]
fn comparison_negative_zero_checks_all_operators_and_numeric_spellings() {
    valid(
        "no-compare-neg-zero",
        &[
            "x === 0; 0 === x; x == 0; 0 == x; x < 0; 0 <= x; x > 0; 0 >= x;",
            "x === '0'; '-0' === x; x == '-0'; x === -1; -1 === x;",
            "Object.is(x, -0); x != 0; 0 !== x;",
            "x === -0n; x === (-0 as number); x === (-0 satisfies number);",
            "x === -(-0); x === +(-0); x === (-0, y); x ** -0;",
        ],
    );
    let mut sources = Vec::new();
    let mut expected = Vec::new();
    for operator in [">", ">=", "<", "<=", "==", "===", "!=", "!=="] {
        for (left, right) in [("x", "-0"), ("-0", "x")] {
            let source = format!("{left} {operator} {right}");
            expected.push(message(
                "no-compare-neg-zero",
                "unexpected",
                &format!("Do not use the '{operator}' operator to compare against -0."),
                1,
                source.len() + 1,
            ));
            sources.push(source);
        }
    }
    for zero in [
        "(-0)",
        "-(0)",
        "(-(((0))))",
        "(/* before */ - /* after */ (0))",
        "-0.0",
        "-0e10",
        "-0x0",
        "-0b0",
        "-0o0",
    ] {
        let source = format!("x === {zero}");
        expected.push(message(
            "no-compare-neg-zero",
            "unexpected",
            "Do not use the '===' operator to compare against -0.",
            1,
            source.len() + 1,
        ));
        sources.push(source);
    }
    let source_refs: Vec<&str> = sources.iter().map(String::as_str).collect();
    for ((source, messages), expected) in sources
        .iter()
        .zip(lint("no-compare-neg-zero", &source_refs))
        .zip(expected)
    {
        assert_eq!(messages, json!([expected]), "{source}");
    }
}

#[test]
fn comparison_negative_zero_deduplicates_operands_and_reports_full_comparison() {
    let output = lint(
        "no-compare-neg-zero",
        &["-0 === -0", "x === -0 === -0", "if (x !== (-0)) {}"],
    );
    let comparison = |operator: &str, start, end| {
        message(
            "no-compare-neg-zero",
            "unexpected",
            &format!("Do not use the '{operator}' operator to compare against -0."),
            start,
            end,
        )
    };
    assert_eq!(output[0], json!([comparison("===", 1, 10)]));
    assert_eq!(
        output[1],
        json!([comparison("===", 1, 9), comparison("===", 1, 16)])
    );
    assert_eq!(output[2], json!([comparison("!==", 5, 15)]));
}

#[test]
fn template_placeholders_use_decoded_strings_and_require_nonempty_content() {
    valid(
        "no-template-curly-in-string",
        &[
            "`Hello, ${name}`;",
            "templateFunction`Hello, ${name}`;",
            "'Hello, name'; 'Hello, ' + name;",
            "'$2'; '${'; '$}'; '{foo}'; '${}';",
        ],
    );
    let sources = [
        "'Hello, ${name}'",
        "'${greeting}, ${name}'",
        "'Hello, ${index + 1}'",
        "'Hello, ${{foo: \"bar\"}.foo}'",
        "'${}${name}'",
        "'\\u0024{name}'",
        "'${\\n}'",
    ];
    for (source, messages) in sources
        .iter()
        .zip(lint("no-template-curly-in-string", &sources))
    {
        assert_eq!(
            messages,
            json!([message(
                "no-template-curly-in-string",
                "unexpectedTemplateExpression",
                "Template literal placeholder syntax in regular strings is not allowed.",
                1,
                source.len() + 1
            )]),
            "{source}"
        );
    }
}

#[test]
fn ternaries_preserve_inner_ranges_and_do_not_visit_conditional_types_as_expressions() {
    valid(
        "no-ternary",
        &[
            "\"x ? y\";",
            "type X<T> = T extends string ? 'yes' : 'no';",
            "type Y<T> = T extends string ? (T extends 'a' ? 1 : 2) : 3;",
        ],
    );
    let sources = [
        "var foo = true ? thing : stuff;",
        "var x = (a ? b : c);",
        "(a ? b : c) as any;",
        "(a ? b : c)!;",
        "a?.b ? c : d;",
    ];
    for (messages, (start, end)) in lint("no-ternary", &sources).into_iter().zip([
        (11, 31),
        (10, 19),
        (2, 11),
        (2, 11),
        (1, 13),
    ]) {
        assert_eq!(
            messages,
            json!([message(
                "no-ternary",
                "noTernaryOperator",
                "Ternary operator used.",
                start,
                end
            )])
        );
    }
    assert_eq!(
        lint("no-ternary", &["var x = a\n  ? b\n  : c;"])[0],
        json!([{
            "ruleId":"no-ternary","messageId":"noTernaryOperator",
            "message":"Ternary operator used.","severity":2,
            "line":1,"column":9,"endLine":3,"endColumn":6,"range":[8,21]
        }])
    );
}

#[test]
fn nested_ternary_checks_only_direct_branches_after_parentheses() {
    valid(
        "no-nested-ternary",
        &[
            "foo ? doBar() : doBaz();",
            "a ? b : c; d ? e : f;",
            "var x = (a ? b : c) ? d : e;",
            "var x = a ? () => b : () => (c ? d : e);",
            "var x = a ? { k: b ? c : d } : e;",
            "var x = a ? [b ? c : d] : e;",
            "var x = a ? foo(b ? c : d) : e;",
            "var x = a ? (b ? c : d) as any : e;",
            "var x = a ? (b ? c : d)! : e;",
            "var x = a ? (b ? c : d) satisfies unknown : e;",
            "var x = a ? <any>(b ? c : d) : e;",
            "type T = A extends B ? C : D extends E ? F : G;",
        ],
    );
    let sources = [
        "a ? b : (c ? d : e)",
        "a ? ((b ? c : d)) : e",
        "a ? b ? c : d : e ? f : g",
    ];
    for (source, messages) in sources.iter().zip(lint("no-nested-ternary", &sources)) {
        assert_eq!(
            messages,
            json!([message(
                "no-nested-ternary",
                "noNestedTernary",
                "Do not nest ternary expressions.",
                1,
                source.len() + 1
            )])
        );
    }
    let result = lint("no-nested-ternary", &["var a = a ? b : c ? d : e ? f : g;"]);
    assert_eq!(
        result[0],
        json!([
            message(
                "no-nested-ternary",
                "noNestedTernary",
                "Do not nest ternary expressions.",
                9,
                34
            ),
            message(
                "no-nested-ternary",
                "noNestedTernary",
                "Do not nest ternary expressions.",
                17,
                34
            )
        ])
    );
}

#[test]
fn with_statements_report_the_complete_statement_and_nested_statements() {
    valid(
        "no-with",
        &[
            "foo.bar(); obj.with(1);",
            "var obj = { with: function() {} }; obj.with();",
            "var s = 'with';",
        ],
    );
    let output = lint(
        "no-with",
        &["with(foo) { bar() }", "with(a) { with(b) { c() } }"],
    );
    let expected = |start, end| {
        message(
            "no-with",
            "unexpectedWith",
            "Unexpected use of 'with' statement.",
            start,
            end,
        )
    };
    assert_eq!(output[0], json!([expected(1, 20)]));
    assert_eq!(output[1], json!([expected(1, 28), expected(11, 26)]));
}

#[test]
fn optionless_syntax_rules_reject_options_instead_of_ignoring_them() {
    let request = json!({
        "schemaVersion":1,"cwd":"/rslint-syntax-tests",
        "files":[{"path":"case.ts","text":"debugger;"}],
        "rules":[{"name":"no-debugger","options":[{}]}]
    });
    assert!(
        lint_json(&request.to_string())
            .unwrap_err()
            .contains("no-debugger does not accept options")
    );
}
