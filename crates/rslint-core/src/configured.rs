//! Execute per-file configuration over explicitly selected compiler Programs.
//! The input stays in admission order, including source-only and silent files.

use std::collections::{BTreeMap, BTreeSet};

use serde::Deserialize;
use serde_json::json;
use tsrs_core::context::{CheckerLifetime, Context, with_checker_lifetime};

use crate::{
    COMPILER_REVISION, FileObservation, InputFile, LintResult, RuleConfig, SessionOwner, autofix,
    observe_file, path, programs::Programs, projects, rules, source, validate_rules,
};

#[derive(Deserialize)]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
struct Request {
    schema_version: u32,
    cwd: String,
    // Source-only programs use the host's configuration directory, independently
    // of target admission. Configured projects retain their own cwd.
    source_only_current_directory: Option<String>,
    files: Vec<Input>,
    #[serde(default)]
    sources: Vec<InputFile>,
    #[serde(default)]
    fix: bool,
    #[serde(default)]
    single_threaded: bool,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct Input {
    path: String,
    text: Option<String>,
    rules: Vec<RuleConfig>,
    policy: projects::ProjectPolicy,
}

/// Owned per-file requests, with actual configuration and project policy.
/// Unknown rules/options fail even on source-only inputs; known type-aware
/// rules are scheduled only when the original project policy grants a checker.
pub fn lint_projects_json(request: &str) -> Result<String, String> {
    let request: Request = serde_json::from_str(request)
        .map_err(|error| format!("invalid configured lint request: {error}"))?;
    if request.schema_version != 1 || !std::path::Path::new(&request.cwd).is_absolute() {
        return Err("configured lint requires schemaVersion 1 and absolute cwd".into());
    }
    if request.source_only_current_directory.as_ref().is_some_and(|cwd| {
        !std::path::Path::new(cwd).is_absolute()
    }) {
        return Err("sourceOnlyCurrentDirectory must be absolute".into());
    }
    for input in &request.files {
        validate_rules(&input.rules)?;
    }
    std::thread::Builder::new()
        .name("rslint-rust-project-lint".into())
        .stack_size(512 << 20)
        .spawn(move || execute(request))
        .map_err(|error| format!("cannot start configured lint worker: {error}"))?
        .join()
        .map_err(|_| "Rust configured lint worker failed".to_owned())?
        .and_then(|result| serde_json::to_string(&result).map_err(|error| error.to_string()))
}

fn execute(request: Request) -> Result<LintResult, String> {
    let mut seen = BTreeSet::new();
    let mut paths = Vec::new();
    let mut overlays = BTreeMap::new();
    let mut inputs = Vec::new();
    for input in &request.files {
        let name = path(&request.cwd, &input.path);
        if !seen.insert(name.clone()) {
            return Err(format!("duplicate admitted file: {name}"));
        }
        if let Some(text) = &input.text {
            overlays.insert(name.clone(), text.clone());
        } else if request.fix
            && let Some(text) = source::read(&name)?
        {
            overlays.insert(name.clone(), text);
        }
        inputs.push(projects::InputProject {
            path: name.clone(),
            policy: input.policy.clone(),
        });
        paths.push(name);
    }
    for input in &request.sources {
        let name = path(&request.cwd, &input.path);
        if seen.contains(&name) || overlays.contains_key(&name) {
            return Err(format!("duplicate source overlay: {name}"));
        }
        let text = input.text.as_ref().ok_or("source overlay requires text")?;
        overlays.insert(name, text.clone());
    }
    let owner = SessionOwner::new(&request.cwd);
    let observe = |sources: &BTreeMap<String, String>| -> Result<Vec<FileObservation>, String> {
        owner.fs.set_generation(&request.cwd, sources)?;
        let plan = projects::select(&owner.session, &inputs)?;
        let programs = Programs::build(
            &plan,
            &paths,
            request.source_only_current_directory.as_deref().unwrap_or(&request.cwd),
            owner.fs.clone(),
            request.single_threaded,
        )?;
        let context = with_checker_lifetime(&Context::background(), CheckerLifetime::API);
        let mut observations = Vec::with_capacity(inputs.len());
        for (index, input) in request.files.iter().enumerate() {
            let lease = programs.for_file(index)?;
            let scheduled = input
                .rules
                .iter()
                .filter(|rule| {
                    lease.has_type_information || !rules::requires_type_information(&rule.name)
                })
                .cloned()
                .collect::<Vec<_>>();
            let mut observation = observe_file(
                lease.input_path(),
                lease.source_file(),
                lease.program(),
                &context,
                &scheduled,
                lease.has_type_information,
            )?;
            observation.result.has_type_information = Some(lease.has_type_information);
            observations.push(observation);
        }
        // All observations own strings/edits. The leases, checker handles and
        // every compiler Program are released before advancing the generation.
        Ok(observations)
    };
    let initial = observe(&overlays)?;
    let outcome = if request.fix {
        autofix::run(initial, |sources| {
            let mut generation = overlays.clone();
            generation.extend(sources.clone());
            observe(&generation)
        })?
    } else {
        autofix::Outcome {
            files: initial.into_iter().map(|file| file.result).collect(),
            fixed_issues: 0,
        }
    };
    Ok(LintResult {
        schema_version: 1,
        backend: "tsrs-rust",
        compiler: json!({"revision": COMPILER_REVISION}),
        files: outcome.files,
        summary: crate::LintSummary {
            fixed_issues: outcome.fixed_issues,
            threads: 1,
        },
    })
}
