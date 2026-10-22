//! One immutable lint generation built with the public tsrs compiler APIs.
//!
//! The project API always enables reference-source redirects. Direct compiler
//! construction keeps that switch explicit, preserves parsed ConfigFile
//! identity through reference cycles, and gives every selected project its
//! own working directory. No project is selected by import membership.

use std::collections::{BTreeMap, HashMap};
use std::fmt::Display;
use std::hash::{DefaultHasher, Hash, Hasher};
use std::sync::{Arc, Mutex, OnceLock};

use serde_json::Value;
use tsrs_ast::{SourceFile, SourceFileParseOptions};
use tsrs_compiler::{CompilerHost, Program, ProgramOptions, ProgramRegions};
use tsrs_core::{
    CompilerOptions, JsxEmit, ModuleKind, P, ScriptKind, ScriptTarget, Tristate,
    arena::{Region, RegionScope},
    ensure_script_kind_from_file_name,
    tspath::{self, ComparePathsOptions, Path},
};
use tsrs_diagnostics::Message;
use tsrs_tsoptions::{ParseConfigHost, ParsedCommandLine};
use tsrs_vfs::{FS, bundled};

use crate::projects::{ProjectPlan, SelectedProject};

/// Owns all compiler state for one source generation. Rebuild rather than
/// updating these Programs after an edit: free_unshared_program requires that
/// no Program was used as the source or result of update_program.
pub(crate) struct Programs {
    configured: Vec<ProgramOwner>,
    source_only: Option<ProgramOwner>,
    file_projects: Vec<Option<usize>>,
    input_bindings: Vec<InputBinding>,
    // Last field: every Program and its checker pool must be freed before the
    // source regions shared by those Programs. Each host also keeps a clone so
    // construction/unwind cannot leave a pointer without a live region owner.
    _parse_cache: Arc<ParseCache>,
}

/// One immutable source generation, not a process-wide compiler cache. The
/// first successful read of each exact filename is authoritative until the
/// generation is dropped (including aliases: never merge by realpath).
///
/// Each AST has its own registered region, following tsrs_project's parse cache:
/// parse, bind and later owner-routed lazy allocations share its lifetime. The
/// cache, rather than whichever Program happened to parse first, owns it. No
/// per-Program config arena or checker arena owns shared binder/AST allocations.
#[derive(Default)]
struct ParseCache {
    snapshots: Mutex<HashMap<String, SourceSlot>>,
    files: Mutex<HashMap<ParseKey, Arc<OnceLock<CachedSource>>>>,
}

type SourceSlot = Arc<Mutex<Option<Arc<SourceSnapshot>>>>;

struct SourceSnapshot {
    text: Arc<str>,
    hash: u64,
}

impl SourceSnapshot {
    fn new(text: String) -> Self {
        let text: Arc<str> = text.into();
        let mut hasher = DefaultHasher::new();
        text.hash(&mut hasher);
        Self {
            text,
            hash: hasher.finish(),
        }
    }
}

#[derive(Clone)]
struct ParseKey {
    options: SourceFileParseOptions,
    kind: ScriptKind,
    snapshot: Arc<SourceSnapshot>,
}

impl Hash for ParseKey {
    fn hash<H: Hasher>(&self, state: &mut H) {
        self.options.hash(state);
        self.kind.hash(state);
        self.snapshot.hash.hash(state);
    }
}

impl PartialEq for ParseKey {
    fn eq(&self, other: &Self) -> bool {
        self.options == other.options
            && self.kind == other.kind
            && self.snapshot.hash == other.snapshot.hash
            // A hash collision must never reuse a different source text.
            && (Arc::ptr_eq(&self.snapshot, &other.snapshot)
                || self.snapshot.text == other.snapshot.text)
    }
}

impl Eq for ParseKey {}

struct CachedSource {
    file: P<SourceFile>,
    _region: Region,
}

impl ParseCache {
    fn get(&self, fs: &dyn FS, options: SourceFileParseOptions) -> Option<P<SourceFile>> {
        let snapshot = {
            let mut snapshots = self
                .snapshots
                .lock()
                .unwrap_or_else(|error| error.into_inner());
            Arc::clone(snapshots.entry(options.file_name.clone()).or_default())
        };
        // Do not memoize failures: a missing input is retried by subsequent
        // callers, matching the original host and avoiding stale negative IO.
        let snapshot = {
            let mut value = snapshot.lock().unwrap_or_else(|error| error.into_inner());
            if value.is_none() {
                *value = Some(Arc::new(SourceSnapshot::new(
                    fs.read_file(&options.file_name)?,
                )));
            }
            Arc::clone(value.as_ref().unwrap())
        };
        let kind = ensure_script_kind_from_file_name(&options.file_name);
        let key = ParseKey {
            options,
            kind,
            snapshot,
        };
        let entry = {
            let mut files = self.files.lock().unwrap_or_else(|error| error.into_inner());
            Arc::clone(files.entry(key.clone()).or_default())
        };
        // OnceLock serializes only callers of the same parse key. Neither map
        // lock is held while reading, parsing or binding; unrelated files can
        // parse in parallel. Source reads are likewise single-flight per exact
        // filename, and a panicking initializer leaves no blocked flight.
        let source = entry.get_or_init(|| {
            let region = Region::new(
                key.snapshot
                    .text
                    .len()
                    .saturating_mul(10)
                    .saturating_add(4 << 10)
                    .min(64 << 20),
            );
            let file = {
                let _scope = region.enter();
                let file = tsrs_parser::parse_source_file(
                    key.options.clone(),
                    &key.snapshot.text,
                    key.kind,
                );
                let text_index = file.text_index.get();
                region.on_free(Box::new(move || {
                    tsrs_ast::unregister_source_text(text_index)
                }));
                // Keep SourceFile.hash at its original zero value: it is
                // public encoded AST data, not this cache's implementation key.
                tsrs_binder::bind_source_file(file);
                file
            };
            region.trim();
            CachedSource {
                file,
                _region: region,
            }
        });
        Some(source.file)
    }
}

struct InputBinding {
    // Diagnostics and edits retain the identity admitted by the caller, even
    // when an explicit project lists another spelling of the same real file.
    input_path: String,
    root_path: String,
}

/// The compiler returns arena pointers without Rust lifetimes. Keep this lease
/// alive while reading its AST/checker, convert diagnostics to owned values,
/// and drop every checker handle before this lease. Compiler pointers must not
/// be retained in the parent request, result, or a later source generation.
pub(crate) struct ProgramLease<'a> {
    program: &'a Program,
    source_file: P<SourceFile>,
    input_path: &'a str,
    pub has_type_information: bool,
    _scope: RegionScope,
}

impl ProgramLease<'_> {
    pub(crate) fn program(&self) -> &Program {
        self.program
    }

    /// The compiler file of this input's selected root. Do not look up the
    /// admitted path again: explicit projects may bind a canonical alias.
    pub(crate) fn source_file(&self) -> P<SourceFile> {
        self.source_file
    }

    /// The caller's lexical path, preserved for diagnostics and edit routing.
    pub(crate) fn input_path(&self) -> &str {
        self.input_path
    }
}

impl Programs {
    pub(crate) fn build(
        plan: &ProjectPlan,
        inputs: &[String],
        cwd: &str,
        fs: Arc<dyn FS>,
        single_threaded: bool,
    ) -> Result<Self, String> {
        if inputs.len() != plan.file_projects.len() {
            return Err("project plan must map every admitted input exactly once".into());
        }
        if !tspath::is_rooted_disk_path(cwd) {
            return Err("program current directory must be absolute".into());
        }
        if plan
            .file_projects
            .iter()
            .flatten()
            .any(|index| *index >= plan.projects.len())
        {
            return Err("project plan contains an unknown selected project".into());
        }
        let input_bindings = inputs
            .iter()
            .zip(&plan.file_projects)
            .map(|(input, project)| {
                if !tspath::is_rooted_disk_path(input) {
                    return Err(format!("admitted lint input must be absolute: {input:?}"));
                }
                let root_path = match project {
                    Some(index) => selected_root(&plan.projects[*index], input, fs.as_ref())?,
                    None => tspath::normalize_path(input),
                };
                Ok(InputBinding {
                    input_path: input.clone(),
                    root_path,
                })
            })
            .collect::<Result<Vec<_>, String>>()?;
        let parse_cache = Arc::new(ParseCache::default());
        let mut configured = Vec::with_capacity(plan.projects.len());
        for selected in &plan.projects {
            configured.push(ProgramOwner::configured(
                selected,
                Arc::clone(&fs),
                single_threaded,
                Arc::clone(&parse_cache),
            )?);
        }
        let roots = inputs
            .iter()
            .zip(&plan.file_projects)
            .filter(|(_, project)| project.is_none())
            .map(|(name, _)| name.clone())
            .collect::<Vec<_>>();
        let source_only = if roots.is_empty() {
            None
        } else {
            Some(ProgramOwner::source_only(
                roots,
                cwd,
                fs,
                single_threaded,
                Arc::clone(&parse_cache),
            )?)
        };
        let programs = Self {
            configured,
            source_only,
            file_projects: plan.file_projects.clone(),
            input_bindings,
            _parse_cache: parse_cache,
        };
        // Detect a listed root that the compiler did not admit. Do not silently
        // select a different Program which happens to import the same path.
        for index in 0..inputs.len() {
            programs.for_file(index)?;
        }
        Ok(programs)
    }

    pub(crate) fn for_file(&self, file_index: usize) -> Result<ProgramLease<'_>, String> {
        let project = self
            .file_projects
            .get(file_index)
            .ok_or("lint input index is outside the project plan")?;
        let owner = match project {
            Some(index) => self
                .configured
                .get(*index)
                .ok_or("selected compiler project is missing")?,
            None => self
                .source_only
                .as_ref()
                .ok_or("source-only compiler program is missing")?,
        };
        owner.lease(&self.input_bindings[file_index], project.is_some())
    }
}

fn selected_root(selected: &SelectedProject, input: &str, fs: &dyn FS) -> Result<String, String> {
    let key = |path: &str| {
        tspath::to_path(
            path,
            &selected.current_directory,
            fs.use_case_sensitive_file_names(),
        )
    };
    let input_key = key(input);
    if let Some(root) = selected
        .root_files
        .iter()
        .find(|root| key(root) == input_key)
    {
        return Ok(root.clone());
    }
    if !selected.source_references {
        // Explicit project discovery admits canonical aliases. Restrict that
        // same comparison to the selected config's roots, in parsed order;
        // an imported file with a matching realpath grants no root membership.
        let canonical_key = |path: &str| {
            let real = fs.realpath(path);
            key(if real.is_empty() { path } else { &real })
        };
        let canonical_input = canonical_key(input);
        if let Some(root) = selected
            .root_files
            .iter()
            .find(|root| canonical_key(root) == canonical_input)
        {
            return Ok(root.clone());
        }
    }
    Err(format!(
        "selected project {:?} does not list admitted input {input:?} as a root",
        selected.config_file_name
    ))
}

struct OwnedParseHost {
    fs: Arc<dyn FS>,
    cwd: String,
}

impl ParseConfigHost for OwnedParseHost {
    fn fs(&self) -> &dyn FS {
        self.fs.as_ref()
    }

    fn get_current_directory(&self) -> &str {
        &self.cwd
    }
}

struct ProgramOwner {
    program: Option<&'static Program>,
    file_regions: Option<Arc<ProgramRegions>>,
    config_region: Option<Region>,
    // ProgramRegions deliberately uses unregistered scratch regions. Lazy
    // SourceFile/ParsedCommandLine caches must instead enter this registered
    // generation-owned region, rather than the permanent thread arena. Never
    // enter it across new_program or a checker call: those can wait on workers
    // which need the same region. Region's shared owner lock is reentrant on
    // one thread and its public handle is Send + Sync.
    lazy_region: Option<Region>,
    // tsrs ParsedCommandLine stores a static ParseConfigHost reference. This
    // box has a stable address and is freed only after all Program/config/file
    // allocations using that reference have been freed. Unlike Box::leak in
    // the compiler's generic reference host, no request filesystem survives.
    parse_host: Option<Box<OwnedParseHost>>,
    parse_cache: Arc<ParseCache>,
}

impl ProgramOwner {
    fn empty(cwd: &str, fs: Arc<dyn FS>, parse_cache: Arc<ParseCache>) -> Self {
        Self {
            program: None,
            file_regions: Some(Arc::new(ProgramRegions::default())),
            config_region: Some(Region::new(1 << 20)),
            lazy_region: Some(Region::new(1 << 16)),
            parse_host: Some(Box::new(OwnedParseHost {
                fs,
                cwd: cwd.to_owned(),
            })),
            parse_cache,
        }
    }

    fn parse_host(&self) -> &'static OwnedParseHost {
        let pointer = std::ptr::from_ref(self.parse_host.as_deref().unwrap());
        // SAFETY: the pointee is in a Box, not in this movable owner. The box
        // remains live until Drop has freed every Program and arena that may
        // reference it. Only the private compiler host and parsed configs get
        // this widened reference; neither can escape this ProgramOwner.
        unsafe { &*pointer }
    }

    fn configured(
        selected: &SelectedProject,
        fs: Arc<dyn FS>,
        single_threaded: bool,
        parse_cache: Arc<ParseCache>,
    ) -> Result<Self, String> {
        let mut owner = Self::empty(&selected.current_directory, fs, parse_cache);
        let _scope = owner.config_region.as_ref().unwrap().enter();
        let (parsed, errors) = tsrs_tsoptions::get_parsed_command_line_of_config_file(
            &selected.config_file_name,
            None,
            None,
            owner.parse_host(),
            None,
        );
        let mut parsed = parsed.ok_or_else(|| {
            format!(
                "could not parse selected project {:?}: {}",
                selected.config_file_name,
                errors
                    .iter()
                    .map(|error| error.localize())
                    .collect::<Vec<_>>()
                    .join("; ")
            )
        })?;
        if selected.source_references {
            let mut options = parsed
                .compiler_options()
                .ok_or("selected project has no compiler options")?
                .get()
                .clone();
            options.allow_non_ts_extensions = Tristate::True;
            parsed.set_compiler_options(P::new(options));
        }
        verify_selection(selected, &parsed)?;
        let parsed = P::new(parsed);
        let host = owner.host(Some(parsed), &selected.config_file_name);
        owner.create(parsed, host, selected.source_references, single_threaded)?;
        Ok(owner)
    }

    fn source_only(
        roots: Vec<String>,
        cwd: &str,
        fs: Arc<dyn FS>,
        single_threaded: bool,
        parse_cache: Arc<ParseCache>,
    ) -> Result<Self, String> {
        let mut owner = Self::empty(cwd, fs, parse_cache);
        let _scope = owner.config_region.as_ref().unwrap().enter();
        let options = P::new(CompilerOptions {
            target: ScriptTarget::ESNext,
            module: ModuleKind::ESNext,
            jsx: JsxEmit::Preserve,
            allow_js: Tristate::True,
            no_lib: Tristate::True,
            no_resolve: Tristate::True,
            ..Default::default()
        });
        let parsed = P::new(tsrs_tsoptions::new_parsed_command_line(
            options,
            roots,
            Vec::new(),
            ComparePathsOptions {
                current_directory: cwd.to_owned(),
                use_case_sensitive_file_names: owner
                    .parse_host()
                    .fs
                    .use_case_sensitive_file_names(),
            },
        ));
        let host = owner.host(None, "");
        owner.create(parsed, host, false, single_threaded)?;
        Ok(owner)
    }

    fn host(&self, root: Option<P<ParsedCommandLine>>, root_name: &str) -> Arc<dyn CompilerHost> {
        let parse_host = self.parse_host();
        let base = tsrs_compiler::new_compiler_host(
            &parse_host.cwd,
            Arc::clone(&parse_host.fs),
            &bundled::default_lib_path(),
            None,
            None,
            None,
        );
        Arc::new(RootHost {
            base,
            parse_host,
            lazy_region: self.lazy_region.as_ref().unwrap().clone(),
            root,
            root_path: tspath::to_path(
                root_name,
                &parse_host.cwd,
                parse_host.fs.use_case_sensitive_file_names(),
            ),
            references: Mutex::new(BTreeMap::new()),
            parse_cache: Arc::clone(&self.parse_cache),
        })
    }

    fn create(
        &mut self,
        config: P<ParsedCommandLine>,
        host: Arc<dyn CompilerHost>,
        source_references: bool,
        single_threaded: bool,
    ) -> Result<(), String> {
        let mut options = ProgramOptions::new(config, host);
        options.use_source_of_project_reference = source_references;
        options.single_threaded = if single_threaded {
            Tristate::True
        } else {
            Tristate::False
        };
        // Keep all ASTs and checker state valid until lint and fix collection
        // finish. One-shot CLI leaf retirement/recycling is unsuitable here.
        options.canonical_diagnostics = false;
        options.checker_regions = true;
        options.checker_diagnostic_regions = true;
        options.regions = self.file_regions.clone();
        let program = tsrs_compiler::new_program(options);
        self.program = Some(program);
        if let Some(error) = program.module_resolution_error() {
            return Err(format!(
                "selected project module resolution failed: {error}"
            ));
        }
        // Source-only syntax rules still need binder-owned symbols. Bind while
        // this owner's arena is entered; none of these symbols may be allocated
        // in a temporary checker's region that can retire independently.
        // Program::bind_source_files uses Rayon when parsing is parallel, but
        // those binder jobs do not enter ProgramRegions. Bind on this thread
        // so symbols remain in the owner arena rather than worker leak arenas.
        // Reverse order matches the compiler's single-threaded binder queue.
        for file in program.source_files().iter().rev() {
            tsrs_binder::bind_source_file(*file);
        }
        Ok(())
    }

    fn lease<'a>(
        &'a self,
        input: &'a InputBinding,
        has_type_information: bool,
    ) -> Result<ProgramLease<'a>, String> {
        let scope = self.config_region.as_ref().unwrap().enter();
        let program = self.program.unwrap();
        let source_file = program.get_source_file(&input.root_path).ok_or_else(|| {
            format!(
                "selected compiler project did not admit lint root {:?} for input {:?}",
                input.root_path, input.input_path
            )
        })?;
        Ok(ProgramLease {
            program,
            source_file,
            input_path: &input.input_path,
            has_type_information,
            _scope: scope,
        })
    }
}

impl Drop for ProgramOwner {
    fn drop(&mut self) {
        if let Some(program) = self.program.take() {
            // SAFETY: only fresh new_program is used, never update/reuse. The
            // owner is inaccessible while a ProgramLease borrows it. All
            // checker handles and AST-backed values must end with that lease;
            // the request exposes only owned JSON. This releases the Program,
            // its checkers, resolver hosts and unshared processed-file/mapper
            // data before their file/config arenas or parse host are freed.
            unsafe { tsrs_compiler::free_unshared_program(program) };
        }
        // Revoke adopted address ownership before the AST/config storage can
        // be reused by another request. These arenas only hold non-owning P
        // caches; their destructors do not dereference lazy allocations.
        drop(self.lazy_region.take());
        drop(self.file_regions.take());
        drop(self.config_region.take());
        drop(self.parse_host.take());
    }
}

fn as_json(value: &tsrs_core::json::Value) -> Result<Value, String> {
    serde_json::from_str(&tsrs_core::json::marshal(value)?).map_err(|error| error.to_string())
}

fn verify_selection(selected: &SelectedProject, parsed: &ParsedCommandLine) -> Result<(), String> {
    let options = parsed
        .compiler_options()
        .ok_or("selected project has no compiler options")?;
    let options = as_json(&tsrs_tsoptions::gojson::compiler_options_to_go_json(
        &options,
    ))?;
    let references = parsed
        .project_references()
        .iter()
        .map(|reference| {
            as_json(&tsrs_tsoptions::gojson::project_reference_to_go_json(
                reference,
            ))
        })
        .collect::<Result<Vec<_>, _>>()?;
    if parsed.file_names() != selected.root_files
        || options != selected.compiler_options
        || references != selected.project_references
    {
        return Err(format!(
            "selected project changed between admission and construction: {:?}",
            selected.config_file_name
        ));
    }
    Ok(())
}

struct RootHost {
    base: Arc<dyn CompilerHost>,
    parse_host: &'static OwnedParseHost,
    lazy_region: Region,
    root: Option<P<ParsedCommandLine>>,
    root_path: Path,
    references: Mutex<BTreeMap<Path, Option<P<ParsedCommandLine>>>>,
    parse_cache: Arc<ParseCache>,
}

impl RootHost {
    fn adopt_unregistered_owner(&self, address: usize) {
        // Do not insert a point registration inside an already registered
        // chunk: that would mask the chunk's following addresses in tsrs's
        // ordered registry. In particular, the original root config already
        // belongs to config_region. Newly parsed scratch objects have no owner.
        if Region::containing(address).is_none() {
            self.lazy_region.adopt_owner(address);
        }
    }
}

impl CompilerHost for RootHost {
    fn fs(&self) -> &dyn FS {
        self.base.fs()
    }

    fn default_library_path(&self) -> &str {
        self.base.default_library_path()
    }

    fn get_current_directory(&self) -> &str {
        self.base.get_current_directory()
    }

    fn trace(&self, message: &'static Message, args: &[&dyn Display]) {
        self.base.trace(message, args);
    }

    fn get_source_file(&self, options: SourceFileParseOptions) -> Option<P<SourceFile>> {
        self.parse_cache.get(self.fs(), options)
    }

    fn get_content_mapped_source_files(
        &self,
        options: SourceFileParseOptions,
        mapper: &tsrs_contentmapper::Mapper,
    ) -> Result<tsrs_contentmapper::SourceFiles, tsrs_contentmapper::Error> {
        let files = self.base.get_content_mapped_source_files(options, mapper)?;
        for source in files.canonical.iter().chain(&files.supplemental) {
            self.adopt_unregistered_owner(source.addr());
        }
        Ok(files)
    }

    fn content_mapper_project(&self) -> Option<Arc<dyn tsrs_contentmapper::Project>> {
        self.base.content_mapper_project()
    }

    fn get_resolved_project_reference(
        &self,
        file_name: &str,
        path: Path,
    ) -> Option<P<ParsedCommandLine>> {
        if path == self.root_path && self.root.is_some() {
            return self.root;
        }
        let mut references = self
            .references
            .lock()
            .unwrap_or_else(|error| error.into_inner());
        if let Some(parsed) = references.get(&path) {
            return *parsed;
        }
        // The file loader parses references inside an unregistered
        // ProgramRegions scope. Adopt each result before returning: the
        // reference parser immediately asks parse_input_output_names to fill
        // its lazy maps through enter_owner, including on parallel builds.
        let (parsed, _) = tsrs_tsoptions::get_parsed_command_line_of_config_file_path(
            file_name,
            path.clone(),
            None,
            None,
            self.parse_host,
            None,
        );
        let parsed = parsed.map(|parsed| {
            let parsed = P::new(parsed);
            self.adopt_unregistered_owner(parsed.addr());
            if let Some(config_file) = parsed.config_file {
                self.adopt_unregistered_owner(config_file.source_file.addr());
            }
            parsed
        });
        references.insert(path, parsed);
        parsed
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::projects::{InputProject, ProjectPolicy, select};
    use std::path::PathBuf;
    use std::sync::atomic::{AtomicU64, Ordering};
    use tsrs_api::{Session, SessionOptions};
    use tsrs_vfs::osvfs;

    static NEXT: AtomicU64 = AtomicU64::new(0);

    struct Fixture {
        root: PathBuf,
        fs: Arc<dyn FS>,
        session: Arc<Session>,
    }

    impl Fixture {
        fn new(files: &[(&str, &str)]) -> Self {
            let root = std::env::temp_dir().join(format!(
                "rslint-program-generation-{}-{}",
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
            let fs: Arc<dyn FS> = Arc::new(bundled::wrap_fs(osvfs::fs()));
            let session = Session::new(SessionOptions::new(
                root.to_string_lossy().into_owned(),
                bundled::default_lib_path(),
                Arc::clone(&fs),
                false,
            ));
            Self { root, fs, session }
        }

        fn path(&self, name: &str) -> String {
            tspath::normalize_path(&self.root.join(name).to_string_lossy())
        }

        fn build(&self, inputs: &[InputProject]) -> Result<Programs, String> {
            let plan = select(&self.session, inputs)?;
            Programs::build(
                &plan,
                &inputs
                    .iter()
                    .map(|input| input.path.clone())
                    .collect::<Vec<_>>(),
                &self.path(""),
                Arc::clone(&self.fs),
                true,
            )
        }
    }

    impl Drop for Fixture {
        fn drop(&mut self) {
            self.session.close();
            std::fs::remove_dir_all(&self.root).unwrap();
        }
    }

    #[test]
    fn cache_shares_full_parse_keys_but_preserves_module_detection_and_exact_paths() {
        let fixture = Fixture::new(&[("entry.ts", "const value = 1;")]);
        let cache = Arc::new(ParseCache::default());
        let options = SourceFileParseOptions {
            file_name: fixture.path("entry.ts"),
            path: tspath::to_path(&fixture.path("entry.ts"), "", true),
            ..Default::default()
        };
        let first = cache.get(fixture.fs.as_ref(), options.clone()).unwrap();
        assert!(first.external_module_indicator().is_none());
        assert_eq!(first.hash.get(), 0);
        std::thread::scope(|scope| {
            let handles = (0..8)
                .map(|_| scope.spawn(|| cache.get(fixture.fs.as_ref(), options.clone()).unwrap()))
                .collect::<Vec<_>>();
            for handle in handles {
                assert_eq!(handle.join().unwrap(), first);
            }
        });
        assert_eq!(cache.files.lock().unwrap().len(), 1);
        assert_eq!(cache.snapshots.lock().unwrap().len(), 1);
        let forced = cache
            .get(
                fixture.fs.as_ref(),
                SourceFileParseOptions {
                    external_module_indicator_options: tsrs_ast::ExternalModuleIndicatorOptions {
                        force: true,
                        jsx: false,
                    },
                    ..options.clone()
                },
            )
            .unwrap();
        assert_ne!(forced, first);
        assert!(forced.external_module_indicator().is_some());
        let distinct_path = cache
            .get(
                fixture.fs.as_ref(),
                SourceFileParseOptions {
                    path: tspath::to_path(&fixture.path("other.ts"), "", true),
                    ..options.clone()
                },
            )
            .unwrap();
        assert_ne!(distinct_path, first);
        assert_eq!(cache.files.lock().unwrap().len(), 3);
        assert_eq!(cache.snapshots.lock().unwrap().len(), 1);

        // Failed reads are not a permanent negative snapshot.
        let missing = SourceFileParseOptions {
            file_name: fixture.path("late.ts"),
            path: tspath::to_path(&fixture.path("late.ts"), "", true),
            ..Default::default()
        };
        assert!(cache.get(fixture.fs.as_ref(), missing.clone()).is_none());
        std::fs::write(fixture.path("late.ts"), "export const late = 1;").unwrap();
        assert!(cache.get(fixture.fs.as_ref(), missing).is_some());
        let addresses = [first.addr(), forced.addr(), distinct_path.addr()];
        assert!(
            addresses
                .into_iter()
                .all(|address| Region::containing(address).is_some())
        );
        drop(cache);
        assert!(
            addresses
                .into_iter()
                .all(|address| Region::containing(address).is_none())
        );
    }

    #[test]
    fn projects_share_source_regions_but_keep_independent_checker_options() {
        let fixture = Fixture::new(&[
            (
                "strict/tsconfig.json",
                r#"{"files":["entry.ts"],"compilerOptions":{"strictNullChecks":true,"noLib":true,"module":"esnext"}}"#,
            ),
            (
                "loose/tsconfig.json",
                r#"{"files":["entry.ts"],"compilerOptions":{"strictNullChecks":false,"noLib":true,"module":"esnext"}}"#,
            ),
            ("strict/entry.ts", "export { identity } from '../shared';"),
            ("loose/entry.ts", "export { identity } from '../shared';"),
            (
                "shared.ts",
                "export function identity(value: number | undefined) { return value; }",
            ),
        ]);
        let inputs = ["strict", "loose"].map(|name| InputProject {
            path: fixture.path(&format!("{name}/entry.ts")),
            policy: ProjectPolicy::Explicit {
                config_files: vec![fixture.path(&format!("{name}/tsconfig.json"))],
            },
        });
        let mut programs = fixture.build(&inputs).unwrap();
        assert_eq!(programs.configured.len(), 2);
        let context = tsrs_core::context::with_checker_lifetime(
            &tsrs_core::context::Context::background(),
            tsrs_core::context::CheckerLifetime::API,
        );
        let mut shared = None;
        for (index, expected) in ["number | undefined", "number"].into_iter().enumerate() {
            let lease = programs.for_file(index).unwrap();
            let file = lease
                .program()
                .get_source_file(&fixture.path("shared.ts"))
                .unwrap();
            if let Some(first) = shared {
                assert_eq!(
                    first, file,
                    "equal parse inputs must share their AST and binder state"
                );
            }
            shared = Some(file);
            let node = file.statements.nodes()[0].parameters()[0].name().unwrap();
            let mut checker = lease
                .program()
                .get_type_checker_for_file_exclusive(&context, file);
            let ty = checker.get_type_at_location(node);
            assert_eq!(checker.type_to_string_exported(ty), expected);
        }
        assert_eq!(programs._parse_cache.files.lock().unwrap().len(), 3);
        let shared = shared.unwrap();
        let address = shared.addr();
        // The first Program happened to populate this entry; freeing that
        // Program must not invalidate the other Program's AST/binder storage.
        drop(programs.configured.remove(0));
        assert_eq!(
            shared.text(),
            "export function identity(value: number | undefined) { return value; }"
        );
        let remaining = programs.configured[0].program.unwrap();
        let mut checker = remaining.get_type_checker_for_file_exclusive(&context, shared);
        let ty = checker
            .get_type_at_location(shared.statements.nodes()[0].parameters()[0].name().unwrap());
        assert_eq!(checker.type_to_string_exported(ty), "number");
        drop(checker);
        drop(programs);
        assert!(Region::containing(address).is_none());
    }

    #[test]
    fn new_program_generation_observes_current_overlay_without_reusing_stale_ast() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["entry.ts"],"compilerOptions":{"noLib":true}}"#,
            ),
            ("entry.ts", "export const value = 'disk';"),
        ]);
        let input = InputProject {
            path: fixture.path("entry.ts"),
            policy: ProjectPolicy::Explicit {
                config_files: vec![fixture.path("tsconfig.json")],
            },
        };
        let plan = select(&fixture.session, &[input]).unwrap();
        let fs = Arc::new(crate::source::SourceFs::new(Arc::clone(&fixture.fs)));
        for text in [
            "export const value = 'before';",
            "export const value = 'after';",
        ] {
            fs.set_generation(
                &fixture.path(""),
                &BTreeMap::from([(fixture.path("entry.ts"), text.to_owned())]),
            )
            .unwrap();
            let programs = Programs::build(
                &plan,
                &[fixture.path("entry.ts")],
                &fixture.path(""),
                fs.clone(),
                true,
            )
            .unwrap();
            let lease = programs.for_file(0).unwrap();
            assert_eq!(lease.source_file().text(), text);
            assert_eq!(programs._parse_cache.files.lock().unwrap().len(), 1);
        }
    }

    #[test]
    fn configured_and_gap_files_keep_distinct_programs_and_type_capability() {
        let fixture = Fixture::new(&[
            (
                "project/tsconfig.json",
                r#"{"files":["entry.ts"],"compilerOptions":{"strict":true,"noLib":true}}"#,
            ),
            ("project/entry.ts", "import './dependency'; export {};"),
            ("project/dependency.ts", "export const value = 1;"),
        ]);
        let inputs = ["entry.ts", "dependency.ts"].map(|name| InputProject {
            path: fixture.path(&format!("project/{name}")),
            policy: ProjectPolicy::Service {
                root_directory: fixture.path(""),
            },
        });
        let plan = select(&fixture.session, &inputs).unwrap();
        let paths = inputs
            .iter()
            .map(|input| input.path.clone())
            .collect::<Vec<_>>();
        let api_config_directory = fixture.path("config-origin");
        let programs = Programs::build(
            &plan,
            &paths,
            &api_config_directory,
            fixture.fs.clone(),
            true,
        )
        .unwrap();
        let typed = programs.for_file(0).unwrap();
        assert!(typed.has_type_information);
        assert_eq!(typed.input_path(), inputs[0].path);
        assert_eq!(typed.source_file().file_name(), inputs[0].path);
        assert_eq!(
            typed.program().get_current_directory(),
            fixture.path("project")
        );
        assert!(typed.program().get_source_file(&inputs[1].path).is_some());
        assert!(typed.program().options().allow_non_ts_extensions.is_true());
        drop(typed);
        let gap = programs.for_file(1).unwrap();
        assert_eq!(gap.program().get_current_directory(), api_config_directory);
        assert!(!gap.has_type_information);
        assert_eq!(gap.input_path(), inputs[1].path);
        assert_eq!(gap.source_file().file_name(), inputs[1].path);
        assert!(gap.program().options().no_lib.is_true());
        assert!(gap.program().options().no_resolve.is_true());
        assert_eq!(gap.program().source_files().len(), 1);
        assert!(
            gap.program()
                .get_source_file(&inputs[1].path)
                .unwrap()
                .as_node()
                .locals()
                .is_some()
        );
    }

    #[cfg(unix)]
    #[test]
    fn explicit_project_aliases_keep_lexical_identity_and_read_the_current_overlay() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["entry.ts"],"compilerOptions":{"noLib":true}}"#,
            ),
            ("entry.ts", "export const value = 'disk';"),
        ]);
        for alias in ["first.ts", "second.ts"] {
            std::os::unix::fs::symlink(fixture.path("entry.ts"), fixture.path(alias)).unwrap();
        }
        let inputs = ["first.ts", "second.ts"].map(|name| InputProject {
            path: fixture.path(name),
            policy: ProjectPolicy::Explicit {
                config_files: vec![fixture.path("tsconfig.json")],
            },
        });
        let plan = select(&fixture.session, &inputs).unwrap();
        assert_eq!(plan.file_projects, vec![Some(0), Some(0)]);
        let source_fs = Arc::new(crate::source::SourceFs::new(Arc::clone(&fixture.fs)));
        source_fs
            .set_generation(
                &fixture.path(""),
                &BTreeMap::from([(
                    inputs[0].path.clone(),
                    "export const value = 'overlay';".to_owned(),
                )]),
            )
            .unwrap();
        let programs = Programs::build(
            &plan,
            &inputs
                .iter()
                .map(|input| input.path.clone())
                .collect::<Vec<_>>(),
            &fixture.path(""),
            source_fs,
            true,
        )
        .unwrap();
        let first = programs.for_file(0).unwrap();
        let second = programs.for_file(1).unwrap();
        assert!(first.has_type_information && second.has_type_information);
        assert_eq!(first.input_path(), fixture.path("first.ts"));
        assert_eq!(second.input_path(), fixture.path("second.ts"));
        assert_eq!(first.source_file(), second.source_file());
        assert_eq!(first.source_file().file_name(), fixture.path("entry.ts"));
        assert_eq!(
            first.source_file().text(),
            "export const value = 'overlay';"
        );
        assert!(programs.for_file(2).is_err());
    }

    #[cfg(unix)]
    #[test]
    fn explicit_project_can_list_the_alias_and_admit_the_real_file() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["alias.ts"],"compilerOptions":{"noLib":true}}"#,
            ),
            ("entry.ts", "export const value = 1;"),
        ]);
        std::os::unix::fs::symlink(fixture.path("entry.ts"), fixture.path("alias.ts")).unwrap();
        let input = InputProject {
            path: fixture.path("entry.ts"),
            policy: ProjectPolicy::Explicit {
                config_files: vec![fixture.path("tsconfig.json")],
            },
        };
        let programs = fixture.build(&[input]).unwrap();
        let lease = programs.for_file(0).unwrap();
        assert!(lease.has_type_information);
        assert_eq!(lease.input_path(), fixture.path("entry.ts"));
        assert_eq!(lease.source_file().file_name(), fixture.path("alias.ts"));
    }

    #[cfg(unix)]
    #[test]
    fn canonical_import_membership_cannot_replace_selected_root_membership() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["entry.ts"],"compilerOptions":{"noLib":true}}"#,
            ),
            ("entry.ts", "import './imported';"),
            ("imported.ts", "export const value = 1;"),
        ]);
        std::os::unix::fs::symlink(fixture.path("imported.ts"), fixture.path("alias.ts")).unwrap();
        let inputs = ["entry.ts", "alias.ts"].map(|name| InputProject {
            path: fixture.path(name),
            policy: ProjectPolicy::Explicit {
                config_files: vec![fixture.path("tsconfig.json")],
            },
        });
        let mut plan = select(&fixture.session, &inputs).unwrap();
        assert_eq!(plan.file_projects, vec![Some(0), None]);
        // A corrupt/stale caller map must not borrow the typed Program merely
        // because it imports the alias target. Reject before constructing it.
        plan.file_projects[1] = Some(0);
        let error = Programs::build(
            &plan,
            &inputs
                .iter()
                .map(|input| input.path.clone())
                .collect::<Vec<_>>(),
            &fixture.path(""),
            Arc::clone(&fixture.fs),
            true,
        )
        .err()
        .unwrap();
        assert!(error.contains("does not list admitted input"));
    }

    #[test]
    fn explicit_and_service_modes_resolve_referenced_sources_differently() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["entry.ts"],"references":[{"path":"./child"}],"compilerOptions":{"noLib":true,"module":"esnext","moduleResolution":"bundler"}}"#,
            ),
            (
                "entry.ts",
                "import { value } from './child/src/value'; export { value };",
            ),
            (
                "child/tsconfig.json",
                r#"{"files":["src/value.ts"],"compilerOptions":{"composite":true,"rootDir":"src","outDir":"dist","noLib":true}}"#,
            ),
            ("child/src/value.ts", "export const value = 1;"),
            (
                "child/dist/value.d.ts",
                "export declare const value: 'declaration';",
            ),
        ]);
        let explicit = InputProject {
            path: fixture.path("entry.ts"),
            policy: ProjectPolicy::Explicit {
                config_files: vec![fixture.path("tsconfig.json")],
            },
        };
        let service = InputProject {
            path: fixture.path("entry.ts"),
            policy: ProjectPolicy::Service {
                root_directory: fixture.path(""),
            },
        };
        for (input, expected, absent) in [
            (explicit, "child/dist/value.d.ts", "child/src/value.ts"),
            (service, "child/src/value.ts", "child/dist/value.d.ts"),
        ] {
            let programs = fixture.build(&[input]).unwrap();
            let lease = programs.for_file(0).unwrap();
            assert!(
                lease
                    .program()
                    .get_source_file(&fixture.path(expected))
                    .is_some(),
                "{expected}"
            );
            assert!(
                lease
                    .program()
                    .get_source_file(&fixture.path(absent))
                    .is_none(),
                "{absent}"
            );
        }
    }

    #[test]
    fn reference_cycle_preserves_the_original_root_config_identity() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["entry.ts"],"references":[{"path":"./child"}],"compilerOptions":{"composite":true,"noLib":true}}"#,
            ),
            ("entry.ts", "export const root = 1;"),
            (
                "child/tsconfig.json",
                r#"{"files":["child.ts"],"references":[{"path":".."}],"compilerOptions":{"composite":true,"noLib":true}}"#,
            ),
            ("child/child.ts", "export {};"),
        ]);
        let input = InputProject {
            path: fixture.path("entry.ts"),
            policy: ProjectPolicy::Service {
                root_directory: fixture.path(""),
            },
        };
        let programs = fixture.build(&[input]).unwrap();
        let owner = &programs.configured[0];
        let lease = programs.for_file(0).unwrap();
        let config = lease.program().command_line();
        let host = owner.host(Some(config), &fixture.path("tsconfig.json"));
        let referenced = host
            .get_resolved_project_reference(
                &fixture.path("tsconfig.json"),
                tspath::to_path(
                    &fixture.path("tsconfig.json"),
                    "",
                    fixture.fs.use_case_sensitive_file_names(),
                ),
            )
            .unwrap();
        assert_eq!(referenced, config);
        assert_eq!(referenced.config_file, config.config_file);
        assert!(
            lease
                .program()
                .get_source_file(&fixture.path("entry.ts"))
                .is_some()
        );
    }

    #[test]
    fn stale_selection_is_an_error_not_a_different_project() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["entry.ts"],"compilerOptions":{"noLib":true}}"#,
            ),
            ("entry.ts", "export {};"),
            ("other.ts", "export {};"),
        ]);
        let input = InputProject {
            path: fixture.path("entry.ts"),
            policy: ProjectPolicy::Service {
                root_directory: fixture.path(""),
            },
        };
        let plan = select(&fixture.session, &[input]).unwrap();
        std::fs::write(
            fixture.root.join("tsconfig.json"),
            r#"{"files":["other.ts"],"compilerOptions":{"noLib":true}}"#,
        )
        .unwrap();
        let result = Programs::build(
            &plan,
            &[fixture.path("entry.ts")],
            &fixture.path(""),
            Arc::clone(&fixture.fs),
            true,
        );
        assert!(
            result
                .err()
                .unwrap()
                .contains("changed between admission and construction")
        );
    }

    #[test]
    fn repeated_generations_release_request_filesystems_and_lazy_regions() {
        let fixture = Fixture::new(&[
            (
                "tsconfig.json",
                r#"{"files":["entry.ts"],"references":[{"path":"./child"}],"compilerOptions":{"noLib":true}}"#,
            ),
            (
                "entry.ts",
                "/** Entry documentation. */\nimport './child/value';\nexport {};",
            ),
            (
                "child/tsconfig.json",
                r#"{"files":["value.ts"],"compilerOptions":{"composite":true,"noLib":true}}"#,
            ),
            ("child/value.ts", "export const value = 1;"),
        ]);
        let input = InputProject {
            path: fixture.path("entry.ts"),
            policy: ProjectPolicy::Service {
                root_directory: fixture.path(""),
            },
        };
        let plan = select(&fixture.session, &[input]).unwrap();
        let freed = Arc::new(AtomicU64::new(0));
        for (generation, single_threaded) in [true, false, true, false].into_iter().enumerate() {
            // Use an independent wrapper: the metadata Session's stock parse
            // host has its own lifetime and must not make this check vacuous.
            let fs: Arc<dyn FS> = Arc::new(bundled::wrap_fs(osvfs::fs()));
            let weak = Arc::downgrade(&fs);
            let programs = Programs::build(
                &plan,
                &[fixture.path("entry.ts")],
                &fixture.path(""),
                fs,
                single_threaded,
            )
            .unwrap();
            {
                let count = Arc::clone(&freed);
                programs.configured[0].config_region.as_ref().unwrap().on_free(Box::new(move || {
                    assert_eq!(count.load(Ordering::SeqCst), generation as u64 + 1,
                        "lazy address registrations must be revoked before config storage is freed");
                }));
                let lazy_region = programs.configured[0].lazy_region.as_ref().unwrap();
                let freed = Arc::clone(&freed);
                lazy_region.on_free(Box::new(move || {
                    freed.fetch_add(1, Ordering::SeqCst);
                }));
                let lease = programs.for_file(0).unwrap();
                for file in lease.program().source_files() {
                    let source_region = Region::containing(file.addr())
                        .expect("shared parsed files must own registered source regions");
                    assert!(
                        !source_region.ptr_eq(lazy_region),
                        "shared AST storage must not belong to one Program's lazy region"
                    );
                    let map = file.get_position_map();
                    assert!(
                        Region::containing(map.addr())
                            .is_some_and(|region| region.ptr_eq(&source_region)),
                        "position map must remain owned by the shared source region"
                    );
                }
                let references = lease.program().get_resolved_project_references();
                let child = references
                    .iter()
                    .flatten()
                    .find(|config| config.config_name() == fixture.path("child/tsconfig.json"))
                    .expect("the child reference must be loaded");
                assert!(
                    Region::containing(child.addr())
                        .is_some_and(|region| region.ptr_eq(lazy_region))
                );
                let mappings = child.source_to_project_reference().unwrap();
                assert!(!mappings.is_empty());
                for mapping in mappings.values() {
                    assert!(
                        Region::containing(mapping.addr())
                            .is_some_and(|region| region.ptr_eq(lazy_region)),
                        "reference output map escaped into a permanent thread arena"
                    );
                }
                let context = tsrs_core::context::with_checker_lifetime(
                    &tsrs_core::context::Context::background(),
                    tsrs_core::context::CheckerLifetime::API,
                );
                let file = lease
                    .program()
                    .get_source_file(&fixture.path("entry.ts"))
                    .unwrap();
                let _checker = lease
                    .program()
                    .get_type_checker_for_file_exclusive(&context, file);
                assert!(weak.upgrade().is_some());
            }
            drop(programs);
            assert!(
                weak.upgrade().is_none(),
                "request filesystem escaped the Program generation"
            );
            assert_eq!(
                freed.load(Ordering::SeqCst),
                generation as u64 + 1,
                "lazy allocation region survived its generation"
            );
        }
    }
}
