//! Select a configured project by its parsed roots, before loading Programs.
//!
//! An imported source is not a lint root. Selection is independent of the
//! compiler's warm/default-project state and never borrows a checker from a
//! different target's Program. Config parsing uses the same request filesystem
//! as source loading, including in-memory config and source overlays.

use std::collections::{BTreeMap, BTreeSet};
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use serde_json::{Value, json};
use tsrs_api::{Handler as _, Response, Session};
use tsrs_core::{JsxEmit, ModuleKind, ScriptTarget, tspath};

/// Project options after flat-config merging. Explicit patterns must already
/// have been expanded, in declaration order, by the admission layer. No
/// implicit default project is introduced for a source-only target.
#[derive(Clone, Debug, Deserialize, Serialize)]
#[serde(tag = "kind", rename_all = "camelCase", deny_unknown_fields)]
pub(crate) enum ProjectPolicy {
    SourceOnly,
    Service {
        #[serde(rename = "rootDirectory")]
        root_directory: String,
    },
    Explicit {
        #[serde(rename = "configFiles")]
        config_files: Vec<String>,
    },
}

#[derive(Clone, Debug)]
pub(crate) struct InputProject {
    pub path: String,
    pub policy: ProjectPolicy,
}

/// Owned config metadata; no compiler pointers survive selection. The parent
/// executor must bind each target to exactly this selection, and confirm that
/// the compiler admitted the selected root before acquiring a checker.
#[derive(Clone, Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct SelectedProject {
    pub config_file_name: String,
    pub current_directory: String,
    pub root_files: Vec<String>,
    pub compiler_options: Value,
    pub project_references: Vec<Value>,
    pub config_file_parsing_diagnostics: Vec<Value>,
    /// Baseline projectService construction redirects to referenced sources;
    /// an explicit project declaration keeps ordinary project construction.
    pub source_references: bool,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub(crate) struct ProjectPlan {
    pub projects: Vec<SelectedProject>,
    /// Aligned with the original input order. None is a source-only gap and
    /// grants no type-information capability, even if another project imports
    /// that source. It does not mean that the file can be dropped from linting.
    pub file_projects: Vec<Option<usize>>,
}

/// Syntax/binding options shared by all source-only targets. These do not
/// grant a checker: the separate file_projects capability remains decisive.
pub(crate) fn source_only_compiler_options() -> Value {
    json!({
        "target": ScriptTarget::ESNext.value(),
        "module": ModuleKind::ESNext.value(),
        "jsx": JsxEmit::Preserve.value(),
        "allowJs": true,
        "noLib": true,
        "noResolve": true,
    })
}

struct ParsedProject {
    metadata: SelectedProject,
    roots: BTreeSet<String>,
    references: Vec<String>,
}

struct Discovery<'a> {
    session: &'a Session,
    configs: BTreeMap<String, Option<Arc<ParsedProject>>>,
}

/// Resolve all targets in one source generation. Recreate this plan after a
/// generation change; caching config roots across autofix generations would
/// conceal edits to configs or source membership.
pub(crate) fn select(session: &Session, inputs: &[InputProject]) -> Result<ProjectPlan, String> {
    let mut discovery = Discovery {
        session,
        configs: BTreeMap::new(),
    };
    let mut projects = Vec::new();
    let mut project_indexes = BTreeMap::new();
    let mut file_projects = Vec::with_capacity(inputs.len());
    for input in inputs {
        let file = absolute_literal(&input.path, "admitted project input")?;
        let (selected, source_references) = match &input.policy {
            ProjectPolicy::SourceOnly => (None, false),
            ProjectPolicy::Service { root_directory } => {
                let root = absolute_literal(root_directory, "project service root")?;
                (discovery.find(&file, &root)?, true)
            }
            ProjectPolicy::Explicit { config_files } => {
                // Expand all declarations before membership selection. An
                // unmatched pattern is represented by no paths, not a fallback
                // to the nearest tsconfig or some other file's project.
                let names = config_files
                    .iter()
                    .map(|name| absolute_literal(name, "explicit project"))
                    .collect::<Result<Vec<_>, _>>()?;
                (discovery.explicit(&file, &names)?, false)
            }
        };
        let index = if let Some(selected) = selected {
            let key = (
                discovery.path_key(&selected.metadata.config_file_name),
                source_references,
            );
            Some(*project_indexes.entry(key).or_insert_with(|| {
                let index = projects.len();
                let mut metadata = selected.metadata.clone();
                metadata.source_references = source_references;
                if source_references {
                    // Apply only after parsing roots, as the baseline service
                    // loader does. This admits explicitly listed JavaScript
                    // without changing allowJs glob expansion or poisoning an
                    // explicit declaration that shares the parsed-config cache.
                    metadata.compiler_options["allowNonTsExtensions"] = Value::Bool(true);
                }
                projects.push(metadata);
                index
            }))
        } else {
            None
        };
        file_projects.push(index);
    }
    Ok(ProjectPlan {
        projects,
        file_projects,
    })
}

fn absolute_literal(name: &str, label: &str) -> Result<String, String> {
    if name.is_empty() || !tspath::is_rooted_disk_path(name) {
        return Err(format!(
            "{label} must be an absolute filesystem path: {name:?}"
        ));
    }
    // Project selectors are configuration paths, not glob expressions. Source
    // filenames themselves may legitimately contain glob punctuation.
    if label == "explicit project" && name.contains(['*', '?', '[', ']', '{', '}']) {
        return Err(format!(
            "explicit project patterns must be expanded before project selection: {name:?}"
        ));
    }
    Ok(tspath::normalize_path(name))
}

impl Discovery<'_> {
    fn path_key(&self, name: &str) -> String {
        tspath::to_path(
            name,
            "",
            self.session.base_fs().use_case_sensitive_file_names(),
        )
        .to_string()
    }

    fn config(&mut self, name: &str) -> Result<Option<Arc<ParsedProject>>, String> {
        let key = self.path_key(name);
        if let Some(cached) = self.configs.get(&key) {
            return Ok(cached.clone());
        }
        if !self.session.base_fs().file_exists(name) {
            self.configs.insert(key, None);
            return Ok(None);
        }
        let response = self
            .session
            .handle_request(
                "parseConfigFile",
                &serde_json::to_vec(&json!({"file": name})).map_err(|e| e.to_string())?,
            )
            .map_err(|error| format!("parse TypeScript project {name:?}: {error}"))?;
        let parsed: Value = match response {
            Response::Json(text) => serde_json::from_str(&text).map_err(|e| e.to_string())?,
            Response::Binary(_) => return Err("unexpected binary parsed project metadata".into()),
        };
        let root_files = parsed["fileNames"]
            .as_array()
            .ok_or_else(|| format!("parsed project {name:?} has no root file list"))?
            .iter()
            .map(|value| {
                value
                    .as_str()
                    .ok_or_else(|| format!("parsed project {name:?} has a non-string root"))
                    .and_then(|value| absolute_literal(value, "parsed project root"))
            })
            .collect::<Result<Vec<_>, _>>()?;
        let compiler_options = parsed["options"].clone();
        if !compiler_options.is_object() {
            return Err(format!("parsed project {name:?} has no compiler options"));
        }
        let project_references = optional_array(&parsed, "projectReferences", name)?;
        let references = project_references
            .iter()
            .map(|reference| {
                let path = reference["path"]
                    .as_str()
                    .ok_or_else(|| format!("parsed project {name:?} has an invalid reference"))?;
                let path = absolute_literal(path, "parsed project reference")?;
                Ok(tsrs_core::resolve_config_file_name_of_project_reference(
                    &path,
                ))
            })
            .collect::<Result<Vec<_>, String>>()?;
        let roots = root_files.iter().map(|name| self.path_key(name)).collect();
        let config = Arc::new(ParsedProject {
            metadata: SelectedProject {
                config_file_name: name.to_owned(),
                current_directory: tspath::get_directory_path(name),
                root_files,
                compiler_options,
                project_references,
                // The original lint loader is lenient about config parsing
                // diagnostics. Preserve them for Program construction rather
                // than making a metadata-only query fail earlier than lint.
                config_file_parsing_diagnostics: optional_array(&parsed, "errors", name)?,
                source_references: false,
            },
            roots,
            references,
        });
        self.configs.insert(key, Some(Arc::clone(&config)));
        Ok(Some(config))
    }

    fn explicit(
        &mut self,
        file: &str,
        names: &[String],
    ) -> Result<Option<Arc<ParsedProject>>, String> {
        let exact = self.path_key(file);
        let canonical = self.canonical_key(file);
        for name in names {
            let Some(parsed) = self.config(name)? else {
                return Err(format!("couldn't read tsconfig at {name}"));
            };
            if parsed.roots.contains(&exact)
                || parsed
                    .metadata
                    .root_files
                    .iter()
                    .any(|root| self.canonical_key(root) == canonical)
            {
                return Ok(Some(parsed));
            }
        }
        Ok(None)
    }

    fn canonical_key(&self, file: &str) -> String {
        let canonical = self.session.base_fs().realpath(file);
        self.path_key(if canonical.is_empty() {
            file
        } else {
            &canonical
        })
    }

    fn find(&mut self, file: &str, root: &str) -> Result<Option<Arc<ParsedProject>>, String> {
        let mut search = Search {
            discovery: self,
            file: String::new(),
            fallback: None,
            seen: BTreeSet::new(),
        };
        search.file = search.discovery.path_key(file);
        let root = search.discovery.path_key(root);
        let mut directory = tspath::get_directory_path(file);
        loop {
            for name in ["tsconfig.json", "jsconfig.json"] {
                let name = tspath::resolve_path(&directory, &[name]);
                let Some(parsed) = search.discovery.config(&name)? else {
                    continue;
                };
                search.seen.insert(search.discovery.path_key(&name));
                if let Some(selected) = search.direct(&parsed)? {
                    return Ok(Some(selected));
                }
                if let Some(selected) = search.references(&parsed)? {
                    return Ok(Some(selected));
                }
                if enabled(&parsed, "disableSolutionSearching") {
                    return Ok(search.fallback);
                }
            }
            let parent = tspath::get_directory_path(&directory);
            let directory_key = search.discovery.path_key(&directory);
            // A configured root need not be an ancestor (an explicit config
            // module can live in node_modules). Match the original boundary:
            // equality, node_modules, or filesystem root; do not reject that
            // valid invocation or impose a new containment test.
            if directory_key == root
                || tspath::get_base_file_name(&directory_key) == "node_modules"
                || parent == directory
            {
                return Ok(search.fallback);
            }
            directory = parent;
        }
    }
}

fn optional_array(parsed: &Value, field: &str, name: &str) -> Result<Vec<Value>, String> {
    match parsed.get(field) {
        None | Some(Value::Null) => Ok(Vec::new()),
        Some(Value::Array(values)) => Ok(values.clone()),
        Some(_) => Err(format!("parsed project {name:?} has invalid {field}")),
    }
}

fn enabled(parsed: &ParsedProject, option: &str) -> bool {
    parsed.metadata.compiler_options[option] == Value::Bool(true)
}

struct Search<'a, 'session> {
    discovery: &'a mut Discovery<'session>,
    file: String,
    fallback: Option<Arc<ParsedProject>>,
    seen: BTreeSet<String>,
}

impl Search<'_, '_> {
    fn referenced_root(
        &mut self,
        parsed: &ParsedProject,
        seen: &mut BTreeSet<String>,
    ) -> Result<bool, String> {
        for name in &parsed.references {
            if !seen.insert(self.discovery.path_key(name)) {
                continue;
            }
            let Some(child) = self.discovery.config(name)? else {
                continue;
            };
            if child.roots.contains(&self.file) || self.referenced_root(&child, seen)? {
                return Ok(true);
            }
        }
        Ok(false)
    }

    fn direct(
        &mut self,
        parsed: &Arc<ParsedProject>,
    ) -> Result<Option<Arc<ParsedProject>>, String> {
        if !parsed.roots.contains(&self.file) {
            return Ok(None);
        }
        if enabled(parsed, "disableSourceOfProjectReferenceRedirect") {
            return Ok(Some(Arc::clone(parsed)));
        }
        let mut seen = BTreeSet::from([self.discovery.path_key(&parsed.metadata.config_file_name)]);
        if !self.referenced_root(parsed, &mut seen)? {
            return Ok(Some(Arc::clone(parsed)));
        }
        if self.fallback.is_none() {
            self.fallback = Some(Arc::clone(parsed));
        }
        Ok(None)
    }

    fn references(&mut self, parsed: &ParsedProject) -> Result<Option<Arc<ParsedProject>>, String> {
        if enabled(parsed, "disableReferencedProjectLoad") {
            return Ok(None);
        }
        let mut children = Vec::new();
        for name in &parsed.references {
            if !self.seen.insert(self.discovery.path_key(name)) {
                continue;
            }
            let Some(child) = self.discovery.config(name)? else {
                continue;
            };
            if let Some(selected) = self.direct(&child)? {
                return Ok(Some(selected));
            }
            children.push(child);
        }
        // Immediate references take precedence over any earlier sibling's
        // subtree. A depth-first walk produces a different selected checker.
        for child in children {
            if let Some(selected) = self.references(&child)? {
                return Ok(Some(selected));
            }
        }
        Ok(None)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;
    use std::sync::atomic::{AtomicU64, Ordering};
    use tsrs_api::SessionOptions;
    use tsrs_vfs::{bundled, osvfs};

    static NEXT: AtomicU64 = AtomicU64::new(0);

    struct Fixture {
        root: PathBuf,
        session: Arc<Session>,
    }

    impl Fixture {
        fn new(files: &[(&str, &str)]) -> Self {
            let root = std::env::temp_dir().join(format!(
                "rslint-project-selection-{}-{}",
                std::process::id(),
                NEXT.fetch_add(1, Ordering::Relaxed)
            ));
            std::fs::create_dir_all(&root).unwrap();
            let root = root.canonicalize().unwrap();
            for (name, text) in files {
                let file = root.join(name);
                std::fs::create_dir_all(file.parent().unwrap()).unwrap();
                std::fs::write(file, text).unwrap();
            }
            let session = Session::new(SessionOptions::new(
                root.to_string_lossy().into_owned(),
                bundled::default_lib_path(),
                Arc::new(bundled::wrap_fs(osvfs::fs())),
                false,
            ));
            Self { root, session }
        }

        fn path(&self, name: &str) -> String {
            tspath::normalize_path(&self.root.join(name).to_string_lossy())
        }

        fn service(&self, names: &[&str]) -> ProjectPlan {
            let inputs = names
                .iter()
                .map(|name| InputProject {
                    path: self.path(name),
                    policy: ProjectPolicy::Service {
                        root_directory: self.path(""),
                    },
                })
                .collect::<Vec<_>>();
            select(&self.session, &inputs).unwrap()
        }

        fn selected<'a>(&self, plan: &'a ProjectPlan, file: usize) -> Option<&'a SelectedProject> {
            plan.file_projects[file].map(|index| &plan.projects[index])
        }
    }

    impl Drop for Fixture {
        fn drop(&mut self) {
            self.session.close();
            std::fs::remove_dir_all(&self.root).unwrap();
        }
    }

    #[test]
    fn imports_do_not_make_targets_configured_roots() {
        let fixture = Fixture::new(&[
            ("tsconfig.json", r#"{"files":["entry.ts"]}"#),
            ("entry.ts", "import { value } from './dependency'; value;"),
            ("dependency.ts", "export const value = 1;"),
        ]);
        let plan = fixture.service(&["entry.ts", "dependency.ts"]);
        assert_eq!(plan.projects.len(), 1);
        assert_eq!(plan.file_projects, vec![Some(0), None]);
        assert!(plan.projects[0].source_references);
        assert_eq!(
            plan.projects[0].compiler_options["allowNonTsExtensions"],
            true
        );
        assert_eq!(plan.projects[0].root_files, vec![fixture.path("entry.ts")]);
    }

    #[test]
    fn search_continues_to_parent_when_nearer_config_excludes_file() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"include":["child/*.ts"],"compilerOptions":{"strict":true}}"#,
            ),
            ("child/tsconfig.json", r#"{"files":["other.ts"]}"#),
            ("child/target.ts", "export {};"),
            ("child/other.ts", "export {};"),
        ]);
        let plan = fixture.service(&["child/target.ts", "child/other.ts"]);
        assert_eq!(
            fixture.selected(&plan, 0).unwrap().config_file_name,
            fixture.path("tsconfig.json")
        );
        assert_eq!(
            fixture.selected(&plan, 1).unwrap().config_file_name,
            fixture.path("child/tsconfig.json")
        );
        assert_eq!(
            fixture.selected(&plan, 0).unwrap().compiler_options["strict"],
            true
        );
    }

    #[test]
    fn referenced_source_uses_child_options_and_cycles_terminate() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["target.ts"],"references":[{"path":"./child"}],"compilerOptions":{"strict":false}}"#,
            ),
            (
                "child/tsconfig.json",
                r#"{"files":["../target.ts"],"references":[{"path":".."}],"compilerOptions":{"strict":true}}"#,
            ),
            ("target.ts", "export {};"),
        ]);
        let plan = fixture.service(&["target.ts"]);
        // Both configs redirect to each other; neither is a terminal direct
        // owner, so the first containing config remains the fallback.
        assert_eq!(
            plan.projects[0].config_file_name,
            fixture.path("tsconfig.json")
        );

        std::fs::write(
            fixture.root.join("child/tsconfig.json"),
            r#"{"files":["../target.ts"],"compilerOptions":{"strict":true}}"#,
        )
        .unwrap();
        let plan = fixture.service(&["target.ts"]);
        assert_eq!(
            plan.projects[0].config_file_name,
            fixture.path("child/tsconfig.json")
        );
        assert_eq!(plan.projects[0].compiler_options["strict"], true);
    }

    #[test]
    fn immediate_references_precede_earlier_reference_subtrees() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":[],"references":[{"path":"./a"},{"path":"./b"}]}"#,
            ),
            (
                "a/tsconfig.json",
                r#"{"files":[],"references":[{"path":"../c"}]}"#,
            ),
            ("b/tsconfig.json", r#"{"files":["../target.ts"]}"#),
            ("c/tsconfig.json", r#"{"files":["../target.ts"]}"#),
            ("target.ts", "export {};"),
        ]);
        let plan = fixture.service(&["target.ts"]);
        assert_eq!(
            plan.projects[0].config_file_name,
            fixture.path("b/tsconfig.json")
        );
    }

    #[test]
    fn redirect_and_reference_loading_flags_preserve_parent_fallback() {
        for flag in [
            "disableSourceOfProjectReferenceRedirect",
            "disableReferencedProjectLoad",
        ] {
            let root = json!({"files":["target.ts"],"references":[{"path":"./child"}],"compilerOptions":{flag:true}}).to_string();
            let fixture = Fixture::new(&[
                ("tsconfig.json", &root),
                ("child/tsconfig.json", r#"{"files":["../target.ts"]}"#),
                ("target.ts", "export {};"),
            ]);
            let plan = fixture.service(&["target.ts"]);
            assert_eq!(
                plan.projects[0].config_file_name,
                fixture.path("tsconfig.json"),
                "{flag}"
            );
        }
    }

    #[test]
    fn solution_search_flag_and_service_root_stop_ancestor_search() {
        let fixture = Fixture::new(&[
            ("tsconfig.json", r#"{"files":["child/target.ts"]}"#),
            (
                "child/tsconfig.json",
                r#"{"files":["other.ts"],"compilerOptions":{"disableSolutionSearching":true}}"#,
            ),
            ("child/target.ts", "export {};"),
            ("child/other.ts", "export {};"),
        ]);
        assert_eq!(
            fixture.service(&["child/target.ts"]).file_projects,
            vec![None]
        );
        std::fs::remove_file(fixture.root.join("child/tsconfig.json")).unwrap();
        let plan = select(
            &fixture.session,
            &[InputProject {
                path: fixture.path("child/target.ts"),
                policy: ProjectPolicy::Service {
                    root_directory: fixture.path("child"),
                },
            }],
        )
        .unwrap();
        assert_eq!(plan.file_projects, vec![None]);
    }

    #[test]
    fn jsconfig_is_checked_after_unmatched_tsconfig() {
        let fixture = Fixture::new(&[
            ("tsconfig.json", r#"{"files":["other.ts"]}"#),
            ("jsconfig.json", r#"{"files":["target.js"]}"#),
            ("target.js", "export {};"),
            ("other.ts", "export {};"),
        ]);
        let plan = fixture.service(&["target.js"]);
        assert_eq!(
            plan.projects[0].config_file_name,
            fixture.path("jsconfig.json")
        );
    }

    #[test]
    fn nonancestor_module_root_does_not_hide_repository_projects() {
        let fixture = Fixture::new(&[
            ("tsconfig.json", r#"{"files":["target.ts"]}"#),
            ("target.ts", "export {};"),
        ]);
        let plan = select(
            &fixture.session,
            &[InputProject {
                path: fixture.path("target.ts"),
                policy: ProjectPolicy::Service {
                    root_directory: fixture.path("node_modules/rstack/dist"),
                },
            }],
        )
        .unwrap();
        assert_eq!(
            plan.projects[0].config_file_name,
            fixture.path("tsconfig.json")
        );
    }

    #[test]
    fn source_only_targets_do_not_parse_nearby_invalid_configs() {
        let fixture = Fixture::new(&[
            ("tsconfig.json", "invalid json"),
            ("target.ts", "export {};"),
        ]);
        let plan = select(
            &fixture.session,
            &[InputProject {
                path: fixture.path("target.ts"),
                policy: ProjectPolicy::SourceOnly,
            }],
        )
        .unwrap();
        assert!(plan.projects.is_empty());
        assert_eq!(plan.file_projects, vec![None]);
        assert_eq!(
            source_only_compiler_options(),
            json!({
                "target":99,"module":99,"jsx":1,"allowJs":true,"noLib":true,"noResolve":true,
            })
        );
    }

    #[test]
    fn explicit_projects_keep_order_and_do_not_gain_reference_discovery() {
        let fixture = Fixture::new(&[
            ("a.json", r#"{"files":["target.ts"]}"#),
            ("b.json", r#"{"files":["target.ts","second.ts"]}"#),
            ("target.ts", "import './imported';"),
            ("second.ts", "export {};"),
            ("imported.ts", "export {};"),
        ]);
        let inputs = ["target.ts", "second.ts", "imported.ts"].map(|name| InputProject {
            path: fixture.path(name),
            policy: ProjectPolicy::Explicit {
                config_files: vec![fixture.path("a.json"), fixture.path("b.json")],
            },
        });
        let plan = select(&fixture.session, &inputs).unwrap();
        assert_eq!(plan.file_projects, vec![Some(0), Some(1), None]);
        assert_eq!(plan.projects[0].config_file_name, fixture.path("a.json"));
        assert!(!plan.projects[0].source_references);
    }

    #[test]
    fn inherited_config_roots_are_resolved_from_their_authored_directory() {
        let fixture = Fixture::new(&[
            ("tsconfig.json", r#"{"extends":"./config/base.json"}"#),
            (
                "config/base.json",
                r#"{"include":["../src/*.ts"],"compilerOptions":{"strict":true}}"#,
            ),
            ("src/target.ts", "export {};"),
        ]);
        let plan = fixture.service(&["src/target.ts"]);
        assert_eq!(
            plan.projects[0].root_files,
            vec![fixture.path("src/target.ts")]
        );
        assert_eq!(plan.projects[0].compiler_options["strict"], true);
    }

    #[test]
    fn invalid_or_unexpanded_project_policies_fail_explicitly() {
        let fixture = Fixture::new(&[("target.ts", "export {};")]);
        for policy in [
            ProjectPolicy::Service {
                root_directory: "relative".into(),
            },
            ProjectPolicy::Explicit {
                config_files: vec![fixture.path("**/tsconfig.json")],
            },
            ProjectPolicy::Explicit {
                config_files: vec![fixture.path("missing.json")],
            },
        ] {
            assert!(
                select(
                    &fixture.session,
                    &[InputProject {
                        path: fixture.path("target.ts"),
                        policy
                    }]
                )
                .is_err()
            );
        }
    }
}
