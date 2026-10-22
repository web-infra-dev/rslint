//! Default-option upstream and regression cases retain the frozen compiler options.
//! Non-default typesToIgnore/checkLiteralConstAssertions cases are outside Rsbuild's
//! enabled options, recorded explicitly in the corpus rather than silently ported.
use rslint_core::{lint_json, lint_projects_json};
use serde_json::{Value, json};

const RULE: &str = "@typescript-eslint/no-unnecessary-type-assertion";

// The project oracle consumes tsconfig spellings. The explicit program API uses
// compiler enum values and canonical library filenames. Convert the actual
// per-case values rather than silently replacing them with helper defaults.
fn program_option(key: &str, value: &Value) -> Value {
    if key == "lib" {
        return Value::Array(
            value
                .as_array()
                .expect("lib must be an array")
                .iter()
                .map(|lib| {
                    let lib = lib
                        .as_str()
                        .expect("lib entries must be strings")
                        .to_ascii_lowercase();
                    json!(if lib.starts_with("lib.") && lib.ends_with(".d.ts") {
                        lib
                    } else {
                        format!("lib.{lib}.d.ts")
                    })
                })
                .collect(),
        );
    }
    if !matches!(key, "target" | "module" | "moduleResolution" | "jsx") || value.is_number() {
        return value.clone();
    }
    let spelling = value
        .as_str()
        .expect("compiler enum must be a string or number")
        .to_ascii_lowercase();
    let number = match (key, spelling.as_str()) {
        ("target", "es5") => 1,
        ("target", "es6" | "es2015") => 2,
        ("target", "es2016") => 3,
        ("target", "es2017") => 4,
        ("target", "es2018") => 5,
        ("target", "es2019") => 6,
        ("target", "es2020") => 7,
        ("target", "es2021") => 8,
        ("target", "es2022") => 9,
        ("target", "es2023") => 10,
        ("target", "es2024") => 11,
        ("target", "es2025") => 12,
        ("target", "es2026") => 13,
        ("target", "esnext") => 99,
        ("module", "none") => 0,
        ("module", "commonjs") => 1,
        ("module", "amd") => 2,
        ("module", "umd") => 3,
        ("module", "system") => 4,
        ("module", "es6" | "es2015") => 5,
        ("module", "es2020") => 6,
        ("module", "es2022") => 7,
        ("module", "esnext") => 99,
        ("module", "node16") => 100,
        ("module", "node18") => 101,
        ("module", "node20") => 102,
        ("module", "nodenext") => 199,
        ("module", "preserve") => 200,
        ("moduleResolution", "classic") => 1,
        ("moduleResolution", "node" | "node10") => 2,
        ("moduleResolution", "node16") => 3,
        ("moduleResolution", "nodenext") => 99,
        ("moduleResolution", "bundler") => 100,
        ("jsx", "preserve") => 1,
        ("jsx", "react") => 2,
        ("jsx", "react-native") => 3,
        ("jsx", "react-jsx") => 4,
        ("jsx", "react-jsxdev") => 5,
        _ => panic!("unsupported fixture compiler option {key}: {value}"),
    };
    json!(number)
}

fn lint(entry: &Value, fix: bool) -> Value {
    let text = entry.as_str().or_else(|| entry["text"].as_str()).unwrap();
    let mut options = json!({"strict":true,"strictNullChecks":true,"esModuleInterop":true,"allowJs":false,"target":99,"module":1,"moduleResolution":2,"jsx":1,"experimentalDecorators":true,
        "lib":["lib.es2015.d.ts","lib.es2017.d.ts","lib.esnext.d.ts","lib.dom.d.ts"],"types":[]});
    if let Some(overrides) = entry.get("compilerOptions") {
        for (key, value) in overrides.as_object().unwrap() {
            options[key] = program_option(key, value);
        }
    }
    let filename = entry["fileName"]
        .as_str()
        .map(str::to_owned)
        .unwrap_or_else(|| format!("input{}", entry["extension"].as_str().unwrap_or(".ts")));
    let result = if entry.get("files").is_some() || entry.get("projectFiles").is_some() {
        // A global support file must be a program root, while only the authored
        // input is admitted for lint. Keep source overlays and root membership
        // identical to the configured JS oracle; do not append declarations to
        // the input, which changes used-before-assignment behavior and ranges.
        let roots = entry
            .get("projectFiles")
            .cloned()
            .unwrap_or_else(|| json!([filename]));
        assert!(roots.as_array().unwrap().contains(&json!(filename)));
        let mut sources = vec![json!({"path":"tsconfig.json","text":json!({
            "compilerOptions":entry["compilerOptions"],"files":roots
        }).to_string()})];
        if let Some(files) = entry["files"].as_object() {
            for (path, contents) in files {
                sources.push(json!({"path":path,"text":contents}));
            }
        }
        let request = json!({"schemaVersion":1,"cwd":"/rslint-type-assertion-tests",
            "files":[{"path":filename,"text":text,"rules":[{"name":RULE}],
                "policy":{"kind":"explicit","configFiles":["/rslint-type-assertion-tests/tsconfig.json"]}}],
            "sources":sources,"fix":fix});
        lint_projects_json(&request.to_string()).unwrap()
    } else {
        let request = json!({"schemaVersion":1,"cwd":"/rslint-type-assertion-tests",
            "files":[{"path":filename,"text":text}],"compilerOptions":options,
            "rules":[{"name":RULE}],"fix":fix});
        lint_json(&request.to_string()).unwrap()
    };
    let result: Value = serde_json::from_str(&result).unwrap();
    assert_eq!(result["files"].as_array().unwrap().len(), 1);
    assert_eq!(
        result["files"][0]["appliedRules"],
        json!([RULE]),
        "{text}\n{result}"
    );
    result["files"][0].clone()
}
#[test]
fn default_upstream_and_regression_corpus() {
    let corpus: Value = serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-type-assertion.json"
    ))
    .unwrap();
    let group = &corpus[0];
    for category in ["valid", "invalid"] {
        let cases = group[category].as_array().unwrap();
        assert!(!cases.is_empty());
        for case in cases {
            let result = lint(case, false);
            let messages = result["messages"].as_array().unwrap();
            assert_eq!(
                messages.is_empty(),
                category == "valid",
                "{}: {}\n{result}",
                case["origin"],
                case["text"]
            );
            if let Some(expected) = case.get("expectedMessageIds") {
                let mut actual: Vec<_> = messages
                    .iter()
                    .map(|message| message["messageId"].as_str().unwrap())
                    .collect();
                let mut expected: Vec<_> = expected
                    .as_array()
                    .unwrap()
                    .iter()
                    .map(|id| id.as_str().unwrap())
                    .collect();
                actual.sort_unstable();
                expected.sort_unstable();
                assert_eq!(actual, expected, "{}", case["text"]);
            }
            for message in messages {
                assert_eq!(message["ruleId"], RULE);
                assert_eq!(message["severity"], 2);
                assert!(message.get("fix").is_some());
                assert!(message.get("suggestions").is_none());
            }
            if let Some(expected) = case.get("expectedOutput") {
                let fixed = lint(case, true);
                assert_eq!(
                    &fixed["output"], expected,
                    "{}: {}",
                    case["origin"], case["text"]
                );
                assert_eq!(fixed["messages"], json!([]), "{}", case["text"]);
            }
        }
    }
}
#[test]
fn exact_messages_ranges_and_whitespace_preserving_fixes() {
    let source = "declare const value: number;\r\n/* leading */\u{00a0}value /* keep */ as /* type */ number;\r\n(<number>value);\r\nvalue!;";
    let result = lint(&json!(source), false);
    let messages = result["messages"].as_array().unwrap();
    assert_eq!(messages.len(), 3);
    for (message, (span, fix)) in messages.iter().zip([
        (
            "value /* keep */ as /* type */ number",
            " as /* type */ number",
        ),
        ("<number>value", "<number>"),
        ("value!", "!"),
    ]) {
        let begin = source.find(span).unwrap();
        let finish = begin + span.len();
        assert_eq!(
            message["range"],
            json!([
                source[..begin].encode_utf16().count(),
                source[..finish].encode_utf16().count()
            ])
        );
        assert_eq!(message["messageId"], "unnecessaryAssertion");
        assert_eq!(
            message["message"],
            "This assertion is unnecessary since it does not change the type of the expression."
        );
        let fix_start = source[begin..finish].find(fix).unwrap() + begin;
        assert_eq!(
            message["fix"]["range"],
            json!([
                source[..fix_start].encode_utf16().count(),
                source[..fix_start + fix.len()].encode_utf16().count()
            ])
        );
        assert_eq!(message["fix"]["text"], "");
    }
}
#[test]
fn non_null_assignment_and_used_before_assignment_guards() {
    let source = "declare let value: string | undefined; value! = 'ok';";
    let result = lint(&json!(source), false);
    assert_eq!(result["messages"].as_array().unwrap().len(), 1);
    assert_eq!(
        result["messages"][0]["messageId"],
        "contextuallyUnnecessary"
    );
    assert_eq!(
        result["messages"][0]["message"],
        "This assertion is unnecessary since the receiver accepts the original type of the expression."
    );
    assert_eq!(
        lint(&json!("let value: number; value!;"), false)["messages"],
        json!([])
    );
    assert_eq!(lint(&json!({"text":"let value: number; value!;","compilerOptions":{"strictNullChecks":false}}),false)["messages"].as_array().unwrap().len(),1);
}
#[test]
fn angle_assertion_fixes_keep_arrow_and_expression_grammar() {
    for (source, output) in [
        (
            "const fn = () => <{ value: number }>{ value: 1 };",
            "const fn = () => ({ value: 1 });",
        ),
        ("<{ value: number }>{ value: 1 };", "({ value: 1 });"),
    ] {
        let result = lint(&json!(source), true);
        assert_eq!(result["output"], output);
        assert_eq!(result["messages"], json!([]));
    }
}
