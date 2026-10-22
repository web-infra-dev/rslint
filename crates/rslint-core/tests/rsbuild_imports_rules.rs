//! Frozen import-rule corpus. Only Rsbuild's actual default options/settings
//! are run; original skips and non-default configurations remain in the corpus.
//! Support files affect the selected Program but are never extra lint targets.
use rslint_core::lint_projects_json;
use serde_json::{Value, json};
use std::path::{Component, Path};

fn corpus() -> Value {
    serde_json::from_str(include_str!(
        "../../../scripts/tsrs/oracle-cases/rsbuild-imports.json"
    ))
    .expect("frozen import corpus")
}
fn lint(group: &Value, entry: &Value, fix: bool) -> Value {
    let root = "/rslint-import-rule-tests";
    let filename = entry["fileName"].as_str().unwrap();
    assert!(!Path::new(filename).is_absolute());
    assert!(
        !Path::new(filename)
            .components()
            .any(|p| matches!(p, Component::ParentDir))
    );
    let fixture = &group["fixtureSets"][entry["fixtureSet"].as_str().unwrap()];
    assert!(fixture.is_object(), "missing fixture set: {entry}");
    let mut files = fixture["files"].as_object().unwrap().clone();
    if let Some(overrides) = entry["files"].as_object() {
        files.extend(
            overrides
                .iter()
                .map(|(path, text)| (path.clone(), text.clone())),
        );
    }
    files.remove(filename); // The RuleTester overlay replaces this fixture placeholder.
    assert!(!files.contains_key("tsconfig.json"));
    let mut options = fixture["compilerOptions"].as_object().unwrap().clone();
    if let Some(overrides) = entry["compilerOptions"].as_object() {
        options.extend(
            overrides
                .iter()
                .map(|(key, value)| (key.clone(), value.clone())),
        );
    }
    let project_files = entry
        .get("projectFiles")
        .or_else(|| fixture.get("projectFiles"))
        .cloned()
        .unwrap_or_else(|| json!([filename]));
    assert!(project_files.as_array().unwrap().contains(&json!(filename)));
    let mut sources = vec![json!({"path":"tsconfig.json", "text":json!({
        "compilerOptions":options,"files":project_files
    }).to_string()})];
    for (path, text) in files {
        assert!(text.is_string());
        assert!(!Path::new(&path).is_absolute());
        assert!(
            !Path::new(&path)
                .components()
                .any(|part| matches!(part, Component::ParentDir))
        );
        sources.push(json!({"path":path,"text":text}));
    }
    let rule = group["rule"].as_str().unwrap();
    let request = json!({
        "schemaVersion":1,"cwd":root,"fix":fix,"singleThreaded":true,
        "files":[{"path":filename,"text":entry["text"],"rules":[{"name":rule}],
            "policy":{"kind":"explicit","configFiles":[format!("{root}/tsconfig.json")]}}],
        "sources":sources
    });
    let result: Value = serde_json::from_str(
        &lint_projects_json(&request.to_string())
            .unwrap_or_else(|error| panic!("{entry}\n{error}")),
    )
    .unwrap();
    assert_eq!(
        result["files"].as_array().unwrap().len(),
        1,
        "one admitted input, retaining all support files"
    );
    let file = result["files"][0].clone();
    assert!(
        file["hasSyntaxErrors"].is_null() || file["hasSyntaxErrors"] == false,
        "{entry}: {file}"
    );
    assert_eq!(file["appliedRules"], json!([rule]), "{entry}: {file}");
    file
}
fn assert_entry(group: &Value, entry: &Value, invalid: bool) {
    let file = lint(group, entry, false);
    let messages = file["messages"].as_array().unwrap();
    assert_eq!(
        !messages.is_empty(),
        invalid,
        "{}\n{}\n{file}",
        entry["origin"],
        entry["text"]
    );
    assert!(
        file.get("output").is_none(),
        "import rules do not rewrite files"
    );
    for message in messages {
        assert_eq!(message["ruleId"], group["rule"]);
        assert_eq!(message["severity"], 2);
        assert!(message.get("fix").is_none());
        assert!(message.get("suggestions").is_none());
    }
    if let Some(expected) = entry["expected"].as_array() {
        assert_eq!(messages.len(), expected.len(), "{entry}: {file}");
        let mut remaining = messages.iter().collect::<Vec<_>>();
        for expected in expected {
            let fields = expected.as_object().unwrap();
            let position = remaining
                .iter()
                .position(|actual| fields.iter().all(|(key, value)| &actual[key] == value));
            assert!(
                position.is_some(),
                "missing complete expected diagnostic {expected}\n{entry}\n{file}"
            );
            remaining.remove(position.unwrap());
        }
        assert!(remaining.is_empty());
    }
}
fn run_rule(rule: &str, count: usize) {
    let corpus = corpus();
    let group = corpus
        .as_array()
        .unwrap()
        .iter()
        .find(|group| group["rule"] == rule)
        .unwrap();
    assert_eq!(
        group["valid"].as_array().unwrap().len() + group["invalid"].as_array().unwrap().len(),
        count
    );
    for category in ["valid", "invalid"] {
        for entry in group[category].as_array().unwrap() {
            assert_entry(group, entry, category == "invalid");
        }
    }
    // No automatic edits or suggestions exist for these four rules. Exercise
    // the fix loop with both an invalid and a valid input, without losing reports.
    for category in ["valid", "invalid"] {
        let entry = &group[category][0];
        let plain = lint(group, entry, false);
        let fixed = lint(group, entry, true);
        assert_eq!(plain["messages"], fixed["messages"]);
        assert!(fixed.get("output").is_none());
    }
}
#[test]
fn default_export_graph_and_compiler_option_corpus() {
    run_rule("import/default", 520);
}
#[test]
fn named_export_and_reexport_failure_path_corpus() {
    run_rule("import/named", 195);
}
#[test]
fn namespace_scope_member_and_mutation_corpus() {
    run_rule("import/namespace", 151);
}
#[test]
fn named_default_collision_and_direct_reexport_corpus() {
    run_rule("import/no-named-as-default", 103);
}
#[test]
fn original_skips_and_non_rsbuild_settings_are_visible() {
    let corpus = corpus();
    let mut skipped = 0;
    let mut other_configuration = 0;
    for group in corpus.as_array().unwrap() {
        assert!(group["exactCompilerOptions"].as_bool().unwrap());
        for entry in group["excluded"].as_array().unwrap() {
            assert!(entry["text"].is_string());
            assert!(entry["origin"].is_string());
            assert!(entry["reason"].is_string());
            if entry["originalSkip"] == true {
                skipped += 1;
            } else {
                assert!(!entry["options"].is_null() || !entry["settings"].is_null());
                other_configuration += 1;
            }
        }
    }
    assert_eq!(skipped + other_configuration, 112);
}
