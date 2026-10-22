//! Decode admitted disk sources without losing the mark needed by autofix.

use serde_json::json;
use std::collections::{BTreeMap, HashMap};
use std::sync::{Arc, RwLock};
use std::time::SystemTime;
use tsrs_api::requestfs::{RequestFileSystem, RequestParams, new_for_update};
use tsrs_vfs::{Entries, FS, FileInfo};

struct Generation {
    cwd: String,
    fs: RequestFileSystem,
    probes: Probes,
}

type Cached<T> = RwLock<HashMap<String, T>>;

/// Host sources are immutable during a compiler generation. Cache each operation
/// independently: deriving `file_exists` from `stat`, for example, would change
/// the semantics of virtual filesystems. Negative lookups matter particularly
/// for module resolution, where multiple projects probe the same missing files.
/// Text reads are the exception: `None` also represents a transient read failure,
/// which must be retried even while the source generation remains unchanged.
#[derive(Default)]
struct Probes {
    canonical: Cached<String>,
    files: Cached<bool>,
    text: Cached<Option<String>>,
    directories: Cached<bool>,
    entries: Cached<Entries>,
    metadata: Cached<Option<FileInfo>>,
    realpaths: Cached<String>,
}

/// The compiler sees one current in-memory source generation over its fixed
/// host filesystem. Queries resolve host aliases before looking in that layer,
/// including aliases that were not among the explicitly admitted lint files.
/// Root names remain unchanged; the compiler still controls preserveSymlinks.
///
/// Replace a generation only after releasing every previous compiler snapshot.
/// Each generation owns its successful and missing filesystem probes; replacing
/// it invalidates both the overlay and the host cache, including symlink targets.
/// Clear it after closing the session: the pinned compiler retains a parse-host
/// filesystem reference, which must not retain this request's source strings.
pub(crate) struct SourceFs {
    base: Arc<dyn FS>,
    generation: RwLock<Option<Generation>>,
}

impl SourceFs {
    pub(crate) fn new(base: Arc<dyn FS>) -> Self {
        Self {
            base,
            generation: RwLock::new(None),
        }
    }

    pub(crate) fn set_generation(
        &self,
        cwd: &str,
        overrides: &BTreeMap<String, String>,
    ) -> Result<(), String> {
        if !std::path::Path::new(cwd).is_absolute() {
            return Err("source filesystem current directory must be absolute".into());
        }
        // Validate the working directory even when this generation has no files.
        let cwd = canonical_path(std::path::Path::new(cwd))?;
        let mut files = BTreeMap::new();
        for (name, text) in overrides {
            if name.is_empty() {
                return Err("source overlay path must not be empty".into());
            }
            let absolute = tsrs_core::tspath::get_normalized_absolute_path(name, &cwd);
            let canonical = canonical_path(std::path::Path::new(&absolute))?;
            if self.base.directory_exists(&canonical) {
                return Err(format!("source overlay path is a directory: {name}"));
            }
            if let Some(previous) = files.insert(canonical, text)
                && previous != text
            {
                return Err(format!(
                    "conflicting source overlays for the same file: {name}"
                ));
            }
        }
        let value = tsrs_core::json::unmarshal(&json!({"kind":"layer","files":files}).to_string())?;
        let params = RequestParams::parse(&value)?;
        // Never layer over the prior generation: that would keep old source
        // strings alive and retain files absent from the new generation.
        let fs = new_for_update(
            &params,
            Arc::clone(&self.base),
            None,
            &cwd,
            &mut tsrs_project::FileChangeSummary::default(),
        )?;
        *self.generation.write().unwrap_or_else(|e| e.into_inner()) = Some(Generation {
            cwd,
            fs,
            probes: Probes::default(),
        });
        Ok(())
    }

    pub(crate) fn clear(&self) {
        *self.generation.write().unwrap_or_else(|e| e.into_inner()) = None;
    }

    fn query<T: Clone>(
        &self,
        path: &str,
        cache: impl FnOnce(&Probes) -> &Cached<T>,
        query: impl FnOnce(&dyn FS, &str) -> T,
    ) -> T {
        self.query_with_cache_policy(path, cache, query, |_| true)
    }

    fn query_with_cache_policy<T: Clone>(
        &self,
        path: &str,
        cache: impl FnOnce(&Probes) -> &Cached<T>,
        query: impl FnOnce(&dyn FS, &str) -> T,
        should_cache: impl FnOnce(&T) -> bool,
    ) -> T {
        let generation = self.generation.read().unwrap_or_else(|e| e.into_inner());
        let Some(generation) = generation.as_ref() else {
            // Outside a snapshot there is no immutability guarantee. Preserve
            // live host queries rather than retaining the last generation.
            let canonical =
                canonical_path(std::path::Path::new(path)).unwrap_or_else(|_| path.to_owned());
            return query(self.base.as_ref(), &canonical);
        };
        let absolute = tsrs_core::tspath::get_normalized_absolute_path(path, &generation.cwd);
        let canonical = cached(&generation.probes.canonical, &absolute, || {
            // Compiler probes may target inaccessible or nonexistent host paths.
            // Keep the original query on resolution failure; setting an overlay,
            // unlike such a read-only probe, rejects a resolution error above.
            canonical_path(std::path::Path::new(&absolute)).unwrap_or_else(|_| absolute.clone())
        });
        cached_if(
            cache(&generation.probes),
            &canonical,
            || query(&generation.fs, &canonical),
            should_cache,
        )
    }
}

fn cached<T: Clone>(cache: &Cached<T>, key: &str, compute: impl FnOnce() -> T) -> T {
    cached_if(cache, key, compute, |_| true)
}

fn cached_if<T: Clone>(
    cache: &Cached<T>,
    key: &str,
    compute: impl FnOnce() -> T,
    should_cache: impl FnOnce(&T) -> bool,
) -> T {
    if let Some(value) = cache.read().unwrap_or_else(|e| e.into_inner()).get(key) {
        return value.clone();
    }
    // Do not hold a cache write lock during host I/O. Parallel misses may read
    // the same immutable value; every successful lookup thereafter reuses it.
    let value = compute();
    if !should_cache(&value) {
        return value;
    }
    cache
        .write()
        .unwrap_or_else(|e| e.into_inner())
        .entry(key.to_owned())
        .or_insert(value)
        .clone()
}

impl FS for SourceFs {
    fn use_case_sensitive_file_names(&self) -> bool {
        self.base.use_case_sensitive_file_names()
    }

    fn file_exists(&self, path: &str) -> bool {
        self.query(path, |cache| &cache.files, |fs, path| fs.file_exists(path))
    }

    fn read_file(&self, path: &str) -> Option<String> {
        self.query_with_cache_policy(
            path,
            |cache| &cache.text,
            |fs, path| fs.read_file(path),
            Option::is_some,
        )
    }

    fn directory_exists(&self, path: &str) -> bool {
        self.query(
            path,
            |cache| &cache.directories,
            |fs, path| fs.directory_exists(path),
        )
    }

    fn get_accessible_entries(&self, path: &str) -> Entries {
        self.query(
            path,
            |cache| &cache.entries,
            |fs, path| fs.get_accessible_entries(path),
        )
    }

    fn stat(&self, path: &str) -> Option<FileInfo> {
        self.query(path, |cache| &cache.metadata, |fs, path| fs.stat(path))
    }

    fn realpath(&self, path: &str) -> String {
        self.query(path, |cache| &cache.realpaths, |fs, path| fs.realpath(path))
    }

    fn write_file(&self, _path: &str, _data: &str) -> Result<(), String> {
        Err("lint source filesystem is read-only".into())
    }

    fn append_file(&self, _path: &str, _data: &str) -> Result<(), String> {
        Err("lint source filesystem is read-only".into())
    }

    fn remove(&self, _path: &str) -> Result<(), String> {
        Err("lint source filesystem is read-only".into())
    }

    fn chtimes(&self, _path: &str, _a_time: SystemTime, _m_time: SystemTime) -> Result<(), String> {
        Err("lint source filesystem is read-only".into())
    }
}

fn canonical_path(path: &std::path::Path) -> Result<String, String> {
    let resolved = match std::fs::canonicalize(path) {
        Ok(resolved) => resolved,
        Err(error) if error.kind() == std::io::ErrorKind::NotFound => {
            let parent = path
                .parent()
                .ok_or_else(|| format!("cannot resolve source path: {}", path.display()))?;
            let name = path
                .file_name()
                .ok_or_else(|| format!("cannot resolve source path: {}", path.display()))?;
            std::path::PathBuf::from(canonical_path(parent)?).join(name)
        }
        Err(error) => {
            return Err(format!(
                "cannot resolve source path {}: {error}",
                path.display()
            ));
        }
    };
    let text = resolved
        .to_str()
        .ok_or("source real path is not valid Unicode")?;
    Ok(tsrs_core::tspath::normalize_path(text))
}

pub(crate) fn read(path: &str) -> Result<Option<String>, String> {
    let bytes = match std::fs::read(path) {
        Ok(bytes) => bytes,
        // Keep the compiler's existing missing-admission diagnostic.
        Err(error) if error.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(error) => return Err(format!("cannot read fix source {path}: {error}")),
    };
    decode(&bytes)
        .map(Some)
        .map_err(|error| format!("unsupported source encoding for {path}: {error}"))
}

fn decode(bytes: &[u8]) -> Result<String, String> {
    let big_endian = match bytes.get(..2) {
        Some([0xff, 0xfe]) => Some(false),
        Some([0xfe, 0xff]) => Some(true),
        _ => None,
    };
    if let Some(big_endian) = big_endian {
        let (pairs, trailing) = bytes[2..].as_chunks::<2>();
        if !trailing.is_empty() {
            return Err("incomplete UTF-16 code unit".into());
        }
        let units: Vec<_> = pairs
            .iter()
            .map(|pair| {
                if big_endian {
                    u16::from_be_bytes(*pair)
                } else {
                    u16::from_le_bytes(*pair)
                }
            })
            .collect();
        let text = String::from_utf16(&units)
            .map_err(|_| "unpaired UTF-16 surrogate is not supported yet".to_owned())?;
        return Ok(format!("\u{feff}{text}"));
    }
    String::from_utf8(bytes.to_vec())
        .map_err(|_| "source is not UTF-8 or BOM-marked UTF-16".to_owned())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;
    use std::sync::atomic::{AtomicU64, AtomicUsize, Ordering};

    struct Directory(PathBuf);

    impl Directory {
        fn new() -> Self {
            static NEXT: AtomicU64 = AtomicU64::new(0);
            let root = std::env::temp_dir().join(format!(
                "rslint-source-fs-{}-{}-{}",
                std::process::id(),
                SystemTime::now()
                    .duration_since(SystemTime::UNIX_EPOCH)
                    .unwrap()
                    .as_nanos(),
                NEXT.fetch_add(1, Ordering::Relaxed),
            ));
            std::fs::create_dir_all(&root).unwrap();
            Self(std::fs::canonicalize(root).unwrap())
        }

        fn name(&self, path: &str) -> String {
            tsrs_core::tspath::normalize_path(self.0.join(path).to_str().unwrap())
        }
    }

    impl Drop for Directory {
        fn drop(&mut self) {
            let _ = std::fs::remove_dir_all(&self.0);
        }
    }

    fn host() -> Arc<dyn FS> {
        Arc::new(tsrs_vfs::osvfs::fs())
    }

    struct TransientRead {
        path: String,
        reads: AtomicUsize,
    }

    impl FS for TransientRead {
        fn use_case_sensitive_file_names(&self) -> bool {
            tsrs_vfs::osvfs::fs().use_case_sensitive_file_names()
        }

        fn file_exists(&self, path: &str) -> bool {
            tsrs_vfs::osvfs::fs().file_exists(path)
        }

        fn read_file(&self, path: &str) -> Option<String> {
            if path == self.path && self.reads.fetch_add(1, Ordering::Relaxed) == 0 {
                return None;
            }
            tsrs_vfs::osvfs::fs().read_file(path)
        }

        fn directory_exists(&self, path: &str) -> bool {
            tsrs_vfs::osvfs::fs().directory_exists(path)
        }

        fn get_accessible_entries(&self, path: &str) -> Entries {
            tsrs_vfs::osvfs::fs().get_accessible_entries(path)
        }

        fn stat(&self, path: &str) -> Option<FileInfo> {
            tsrs_vfs::osvfs::fs().stat(path)
        }

        fn realpath(&self, path: &str) -> String {
            tsrs_vfs::osvfs::fs().realpath(path)
        }

        fn write_file(&self, _: &str, _: &str) -> Result<(), String> {
            Err("test filesystem is read-only".into())
        }

        fn append_file(&self, _: &str, _: &str) -> Result<(), String> {
            Err("test filesystem is read-only".into())
        }

        fn remove(&self, _: &str) -> Result<(), String> {
            Err("test filesystem is read-only".into())
        }

        fn chtimes(&self, _: &str, _: SystemTime, _: SystemTime) -> Result<(), String> {
            Err("test filesystem is read-only".into())
        }
    }

    #[test]
    fn preserves_utf8_and_utf16_marks_without_replacing_invalid_text() {
        let text = "\u{feff}'😀'; value!!;\r\n";
        assert_eq!(decode(text.as_bytes()).unwrap(), text);
        for big_endian in [false, true] {
            let mut bytes = vec![];
            for unit in text.encode_utf16() {
                bytes.extend(if big_endian {
                    unit.to_be_bytes()
                } else {
                    unit.to_le_bytes()
                });
            }
            assert_eq!(decode(&bytes).unwrap(), text);
        }
        assert_eq!(decode(b"").unwrap(), "");
        assert!(decode(&[0xff, 0xfe, 0x00]).is_err());
        assert!(decode(&[0xff, 0xfe, 0x00, 0xd8]).is_err());
        assert!(decode(&[0xff]).is_err());
    }

    #[test]
    fn cached_probes_preserve_positive_and_negative_values() {
        let cache = Cached::<Option<String>>::default();
        let calls = AtomicUsize::new(0);
        for _ in 0..3 {
            assert_eq!(
                cached(&cache, "present", || {
                    calls.fetch_add(1, Ordering::Relaxed);
                    Some("source".into())
                }),
                Some("source".into()),
            );
            assert_eq!(
                cached(&cache, "absent", || {
                    calls.fetch_add(1, Ordering::Relaxed);
                    None
                }),
                None,
            );
        }
        assert_eq!(calls.load(Ordering::Relaxed), 2);
    }

    #[test]
    fn generation_reuses_probes_and_replacement_invalidates_host_and_overlay() {
        let directory = Directory::new();
        std::fs::write(directory.name("entry.ts"), "disk before").unwrap();
        let fs = SourceFs::new(host());
        fs.set_generation(
            &directory.name(""),
            &BTreeMap::from([("entry.ts".into(), "overlay before".into())]),
        )
        .unwrap();

        for _ in 0..2 {
            assert_eq!(fs.read_file("entry.ts"), Some("overlay before".into()));
            assert!(fs.file_exists("entry.ts"));
            assert!(!fs.file_exists("missing.ts"));
            assert_eq!(fs.read_file("missing.ts"), None);
            assert_eq!(fs.stat("missing.ts"), None);
            assert!(fs.directory_exists("."));
            assert!(!fs.directory_exists("missing"));
            assert_eq!(fs.realpath("entry.ts"), directory.name("entry.ts"));
            assert!(!fs.get_accessible_entries(".").files.is_empty());
        }
        {
            let generation = fs.generation.read().unwrap();
            let probes = &generation.as_ref().unwrap().probes;
            assert_eq!(probes.files.read().unwrap().len(), 2);
            assert_eq!(probes.text.read().unwrap().len(), 1);
            assert_eq!(probes.metadata.read().unwrap().len(), 1);
            assert_eq!(probes.directories.read().unwrap().len(), 2);
            assert_eq!(probes.entries.read().unwrap().len(), 1);
            assert_eq!(probes.realpaths.read().unwrap().len(), 1);
        }

        std::fs::write(directory.name("entry.ts"), "disk after").unwrap();
        std::fs::write(directory.name("missing.ts"), "created after").unwrap();
        // Metadata misses remain stable during a compiler generation, while
        // failed text reads are retried because None can also mean an I/O error.
        assert!(!fs.file_exists("missing.ts"));
        assert_eq!(fs.read_file("missing.ts"), Some("created after".into()));
        fs.set_generation(&directory.name(""), &BTreeMap::new())
            .unwrap();
        assert_eq!(fs.read_file("entry.ts"), Some("disk after".into()));
        assert!(fs.file_exists("missing.ts"));
        assert_eq!(fs.read_file("missing.ts"), Some("created after".into()));
        assert!(fs.stat("missing.ts").is_some());

        fs.clear();
        assert!(fs.generation.read().unwrap().is_none());
        std::fs::write(directory.name("entry.ts"), "live host").unwrap();
        assert_eq!(
            fs.read_file(&directory.name("entry.ts")),
            Some("live host".into()),
        );
    }

    #[test]
    fn transient_read_failure_recovers_without_replacing_the_generation() {
        let directory = Directory::new();
        std::fs::write(directory.name("entry.ts"), "export const value = 1;").unwrap();
        let base = Arc::new(TransientRead {
            path: directory.name("entry.ts"),
            reads: AtomicUsize::new(0),
        });
        let fs = SourceFs::new(base.clone());
        fs.set_generation(&directory.name(""), &BTreeMap::new())
            .unwrap();
        assert!(fs.file_exists("entry.ts"));
        assert_eq!(fs.read_file("entry.ts"), None);
        assert!(
            fs.generation
                .read()
                .unwrap()
                .as_ref()
                .unwrap()
                .probes
                .text
                .read()
                .unwrap()
                .is_empty()
        );
        for _ in 0..2 {
            assert_eq!(
                fs.read_file("entry.ts"),
                Some("export const value = 1;".into()),
            );
        }
        assert_eq!(base.reads.load(Ordering::Relaxed), 2);
    }

    #[cfg(unix)]
    #[test]
    fn aliases_share_overlay_and_probe_identity_only_for_the_current_generation() {
        let directory = Directory::new();
        std::fs::write(directory.name("first.ts"), "first disk").unwrap();
        std::fs::write(directory.name("second.ts"), "second disk").unwrap();
        std::os::unix::fs::symlink(directory.name("first.ts"), directory.name("alias.ts")).unwrap();
        let fs = SourceFs::new(host());
        fs.set_generation(
            &directory.name(""),
            &BTreeMap::from([("alias.ts".into(), "first overlay".into())]),
        )
        .unwrap();
        assert_eq!(fs.read_file("alias.ts"), Some("first overlay".into()));
        assert_eq!(fs.read_file("first.ts"), Some("first overlay".into()));
        {
            let generation = fs.generation.read().unwrap();
            let generation = generation.as_ref().unwrap();
            assert_eq!(generation.probes.text.read().unwrap().len(), 1);
            assert_eq!(generation.probes.canonical.read().unwrap().len(), 2);
        }
        let conflicting = BTreeMap::from([
            ("alias.ts".into(), "one".into()),
            ("first.ts".into(), "two".into()),
        ]);
        assert!(
            fs.set_generation(&directory.name(""), &conflicting)
                .is_err()
        );
        assert_eq!(fs.read_file("alias.ts"), Some("first overlay".into()));

        std::fs::remove_file(directory.name("alias.ts")).unwrap();
        std::os::unix::fs::symlink(directory.name("second.ts"), directory.name("alias.ts"))
            .unwrap();
        fs.set_generation(
            &directory.name(""),
            &BTreeMap::from([("alias.ts".into(), "second overlay".into())]),
        )
        .unwrap();
        assert_eq!(fs.read_file("alias.ts"), Some("second overlay".into()));
        assert_eq!(fs.read_file("second.ts"), Some("second overlay".into()));
        assert_eq!(fs.read_file("first.ts"), Some("first disk".into()));
        assert_eq!(fs.realpath("alias.ts"), directory.name("second.ts"));
    }

    #[cfg(unix)]
    #[test]
    fn missing_overlay_file_resolves_through_a_symlinked_parent() {
        let directory = Directory::new();
        std::fs::create_dir(directory.name("real")).unwrap();
        std::os::unix::fs::symlink(directory.name("real"), directory.name("alias")).unwrap();
        let fs = SourceFs::new(host());
        fs.set_generation(
            &directory.name(""),
            &BTreeMap::from([("alias/new.ts".into(), "virtual source".into())]),
        )
        .unwrap();
        assert_eq!(fs.read_file("real/new.ts"), Some("virtual source".into()));
        assert_eq!(fs.read_file("alias/new.ts"), Some("virtual source".into()));
        assert!(fs.file_exists("alias/new.ts"));
        fs.set_generation(&directory.name(""), &BTreeMap::new())
            .unwrap();
        assert_eq!(fs.read_file("alias/new.ts"), None);
        assert!(!fs.file_exists("real/new.ts"));
    }
}
