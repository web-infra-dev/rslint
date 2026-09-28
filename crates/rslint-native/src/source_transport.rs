//! Private CLI source transport. The Go adapter is the sole writer. It publishes
//! a slot through IPC only after copying the complete immutable source snapshots.
//! The host registers that slot before dispatch and revokes it before replying.
//! A slot is reusable only if revocation proves every native reader has returned.
//! No mapped memory or pointer is exposed to JavaScript; AST JSON is unchanged.

mod mapping;
#[cfg(feature = "test-worker-termination")]
pub(crate) mod worker_test;

use napi::{Error, Result};
use napi_derive::napi;
use std::collections::HashMap;
use std::sync::{
    atomic::{AtomicU32, Ordering},
    Arc, Mutex, OnceLock,
};

// Native addon statics are shared by Node's worker isolates. An Arc pins BOTH
// the slot contents and their mapping across revocation, shutdown and GC.
static READERS: OnceLock<Mutex<HashMap<u32, Arc<Lease>>>> = OnceLock::new();
static NEXT_LEASE: AtomicU32 = AtomicU32::new(1);

fn readers() -> &'static Mutex<HashMap<u32, Arc<Lease>>> {
    READERS.get_or_init(|| Mutex::new(HashMap::new()))
}

fn invalid() -> Error {
    Error::from_reason("invalid or expired shared source")
}

/// Selected by the Go peer for this session, before either side maps sources.
#[napi(object)]
#[derive(Clone, Copy)]
pub struct SourceConfiguration {
    pub version: u32,
    pub slot_count: u32,
    pub slot_size: u32,
    pub header_size: u32,
    pub publication_stride: u32,
}

#[napi(object)]
pub struct SourceMapping {
    pub version: u32,
    pub fd: Option<i32>,
    pub handle: Option<String>,
    pub process_id: Option<u32>,
}

/// Checked once before reserving any views. These are runtime dimensions;
/// only the v1 atomic word representation is fixed by the reader algorithm.
#[derive(Clone, Copy)]
struct Layout {
    version: u32,
    slot_count: usize,
    slot_size: usize,
    header_size: usize,
    publication_stride: usize,
    capacity: usize,
}

impl Layout {
    fn new(config: SourceConfiguration) -> Result<Self> {
        let invalid = || Error::from_reason("invalid shared source configuration");
        let slot_count = usize::try_from(config.slot_count).map_err(|_| invalid())?;
        let slot_size = usize::try_from(config.slot_size).map_err(|_| invalid())?;
        let header_size = usize::try_from(config.header_size).map_err(|_| invalid())?;
        let publication_stride =
            usize::try_from(config.publication_stride).map_err(|_| invalid())?;
        let alignment = std::mem::align_of::<AtomicU32>();
        if config.version != 1
            || slot_count == 0
            || slot_size == 0
            || publication_stride < std::mem::size_of::<AtomicU32>()
            || publication_stride % alignment != 0
        {
            return Err(invalid());
        }
        let controls = slot_count
            .checked_mul(publication_stride)
            .ok_or_else(invalid)?;
        let capacity = slot_count
            .checked_mul(slot_size)
            .and_then(|payload| header_size.checked_add(payload))
            .filter(|capacity| *capacity <= isize::MAX as usize)
            .ok_or_else(invalid)?;
        if header_size < controls {
            return Err(invalid());
        }
        Ok(Self {
            version: config.version,
            slot_count,
            slot_size,
            header_size,
            publication_stride,
            capacity,
        })
    }

    fn slot_start(self, slot: usize) -> usize {
        debug_assert!(slot < self.slot_count);
        self.header_size + slot * self.slot_size
    }

    fn publication_offset(self, slot: usize) -> usize {
        debug_assert!(slot < self.slot_count);
        slot * self.publication_stride
    }
}

struct Lease {
    mapping: Arc<mapping::Mapping>,
    start: usize,
    length: usize,
}

impl Lease {
    fn bytes(&self, offset: u32, length: u32) -> Result<&[u8]> {
        let offset = offset as usize;
        let length = length as usize;
        if offset > self.length || length > self.length - offset {
            return Err(invalid());
        }
        // SAFETY: register() checks the whole slot's bounds. Holding this Lease
        // prevents a successful release acknowledgement, so the producer cannot
        // overwrite it. Only this published range is borrowed, never the arena
        // (Go may be writing another slot concurrently).
        Ok(unsafe { self.mapping.bytes(self.start + offset, length) })
    }
}

#[derive(Default)]
struct Slot {
    generation: u32,
    lease: u32,
    retired: bool,
}

/// Owned by one CLI engine, never by a plugin worker. Creation may fail; the
/// engine then keeps the existing complete-source JSON transport.
#[napi]
pub struct SourceArena {
    backing: Option<mapping::Backing>,
    mapping: Option<Arc<mapping::Mapping>>,
    // Track only slots actually used. An advertised count must not cause a
    // proportional heap allocation before any source has been published.
    slots: HashMap<u32, Slot>,
}

#[napi]
impl SourceArena {
    #[napi(constructor)]
    pub fn new() -> Result<Self> {
        let backing =
            mapping::Backing::new().map_err(|error| Error::from_reason(error.to_string()))?;
        Ok(Self {
            backing: Some(backing),
            mapping: None,
            slots: HashMap::new(),
        })
    }

    /// Unix must inherit this empty backing object when the Go peer is spawned.
    /// Windows creates its mapping only after receiving the runtime dimensions.
    #[napi]
    pub fn fd(&self) -> Option<i32> {
        self.backing
            .as_ref()
            .and_then(mapping::Backing::fd)
            .or_else(|| self.mapping.as_ref().and_then(|mapping| mapping.fd()))
    }

    /// One attempt per arena, including failures. A failed mapping may already
    /// have resized or sealed its backing, so it cannot safely be configured again.
    #[napi]
    pub fn configure(&mut self, config: SourceConfiguration) -> Result<()> {
        let backing = self.backing.take().ok_or_else(invalid)?;
        let layout = Layout::new(config)?;
        let mapping = backing
            .configure(layout)
            .map_err(|error| Error::from_reason(error.to_string()))?;
        self.mapping = Some(Arc::new(mapping));
        Ok(())
    }

    #[napi]
    pub fn descriptor(&self) -> Result<SourceMapping> {
        let mapping = self.mapping.as_ref().ok_or_else(invalid)?;
        Ok(mapping.descriptor())
    }

    #[napi]
    pub fn register(&mut self, slot: u32, generation: u32, length: u32) -> Result<u32> {
        let mapping = self.mapping.as_ref().ok_or_else(invalid)?;
        let layout = mapping.layout();
        if slot as usize >= layout.slot_count
            || length as usize > layout.slot_size
            || generation == 0
            || self.slots.get(&slot).is_some_and(|entry| {
                entry.retired || entry.lease != 0 || generation <= entry.generation
            })
        {
            return Err(invalid());
        }
        // All shipped targets have lock-free 32-bit atomics with four-byte
        // alignment. Pair with Go's atomic publication before borrowing data.
        // An empty slot has no source bytes requiring publication.
        if length != 0 && mapping.published(slot as usize) != generation {
            return Err(invalid());
        }
        self.slots.try_reserve(1).map_err(|_| invalid())?;
        let mut registry = readers().lock().map_err(|_| invalid())?;
        registry.try_reserve(1).map_err(|_| invalid())?;
        // Never wrap and accidentally make an old descriptor valid again.
        let id = NEXT_LEASE
            .fetch_update(Ordering::Relaxed, Ordering::Relaxed, |id| id.checked_add(1))
            .map_err(|_| invalid())?;
        let lease = Arc::new(Lease {
            mapping: Arc::clone(mapping),
            start: layout.slot_start(slot as usize),
            length: length as usize,
        });
        registry.insert(id, lease);
        let entry = self.slots.entry(slot).or_default();
        entry.generation = generation;
        entry.lease = id;
        Ok(id)
    }

    /// Revoke future reads first, then prove that no earlier read is in flight.
    /// A false result permanently retires the slot: a timed-out worker may still
    /// be inside the parser. No timers or diagnostic heuristics are involved.
    #[napi]
    pub fn release(&mut self, lease: u32) -> bool {
        let Some(slot) = self
            .slots
            .values_mut()
            .find(|slot| slot.lease == lease && lease != 0)
        else {
            return false;
        };
        let reader = readers()
            .lock()
            .unwrap_or_else(|error| error.into_inner())
            .remove(&lease);
        slot.lease = 0;
        let released = reader.is_some_and(|reader| Arc::try_unwrap(reader).is_ok());
        slot.retired = !released;
        released
    }

    #[napi]
    pub fn close(&mut self) {
        let mut registry = readers().lock().unwrap_or_else(|error| error.into_inner());
        for slot in self.slots.values() {
            registry.remove(&slot.lease);
        }
        self.slots.clear();
        self.mapping = None;
        self.backing = None;
    }
}

impl Drop for SourceArena {
    fn drop(&mut self) {
        self.close();
    }
}

#[napi(object)]
pub struct SharedSource {
    pub lease: u32,
    pub offset: u32,
    pub length: u32,
}

/// Pin the registered lease for the whole callback. Its borrowed slice cannot
/// escape the callback, so revocation cannot unmap or reuse bytes while read.
pub(crate) fn with_bytes<T>(
    source: SharedSource,
    read: impl FnOnce(&[u8]) -> Result<T>,
) -> Result<T> {
    let lease = readers()
        .lock()
        .map_err(|_| invalid())?
        .get(&source.lease)
        .cloned()
        .ok_or_else(invalid)?;
    read(lease.bytes(source.offset, source.length)?)
}

#[cfg(test)]
fn test_configuration() -> SourceConfiguration {
    SourceConfiguration {
        version: 1,
        slot_count: 3,
        slot_size: 32 * 1024 + 3,
        header_size: 8 * 1024 + 1,
        publication_stride: 16,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn configured_arena() -> SourceArena {
        let mut arena = SourceArena::new().unwrap();
        arena.configure(test_configuration()).unwrap();
        arena
    }

    #[test]
    fn requires_one_configuration_before_reading() {
        let mut arena = SourceArena::new().unwrap();
        let fd = arena.fd();
        #[cfg(unix)]
        assert!(fd.is_some());
        #[cfg(windows)]
        assert!(fd.is_none());
        assert!(arena.descriptor().is_err());
        assert!(arena.register(0, 1, 0).is_err());

        let config = test_configuration();
        arena.configure(config).unwrap();
        assert_eq!(arena.fd(), fd);
        assert_eq!(arena.descriptor().unwrap().version, config.version);
        assert!(arena.slots.is_empty());
        let lease = arena.register(0, 1, 0).unwrap();
        assert!(arena.configure(config).is_err());
        assert!(arena.release(lease));
    }

    #[test]
    fn close_before_configuration_disposes_backing() {
        let mut arena = SourceArena::new().unwrap();
        arena.close();
        arena.close();
        assert!(arena.fd().is_none());
        assert!(arena.descriptor().is_err());
        assert!(arena.configure(test_configuration()).is_err());
        assert!(arena.register(0, 1, 0).is_err());
    }

    #[test]
    fn rejects_invalid_layouts_without_allowing_reconfiguration() {
        let valid = test_configuration();
        for config in [
            SourceConfiguration {
                version: 0,
                ..valid
            },
            SourceConfiguration {
                version: 2,
                ..valid
            },
            SourceConfiguration {
                slot_count: 0,
                ..valid
            },
            SourceConfiguration {
                slot_size: 0,
                ..valid
            },
            SourceConfiguration {
                header_size: 0,
                ..valid
            },
            SourceConfiguration {
                header_size: 4,
                ..valid
            },
            SourceConfiguration {
                publication_stride: 0,
                ..valid
            },
            SourceConfiguration {
                publication_stride: 2,
                ..valid
            },
            SourceConfiguration {
                publication_stride: 6,
                ..valid
            },
            SourceConfiguration {
                slot_count: u32::MAX,
                slot_size: u32::MAX - 3,
                header_size: u32::MAX - 3,
                publication_stride: 4,
                ..valid
            },
        ] {
            let mut arena = SourceArena::new().unwrap();
            assert!(arena.configure(config).is_err());
            assert!(arena.fd().is_none());
            assert!(arena.descriptor().is_err());
            assert!(arena.configure(valid).is_err());
            assert!(arena.slots.is_empty());
        }
    }

    #[test]
    fn slot_state_is_allocated_only_for_registered_slots() {
        let config = SourceConfiguration {
            slot_count: 1_000_000,
            slot_size: 4,
            header_size: 4_000_000,
            publication_stride: 4,
            ..test_configuration()
        };
        let mut arena = SourceArena::new().unwrap();
        arena.configure(config).unwrap();
        assert!(arena.slots.is_empty());
        let lease = arena.register(config.slot_count - 1, 1, 0).unwrap();
        assert_eq!(arena.slots.len(), 1);
        assert!(arena.release(lease));
    }

    #[test]
    fn rejects_invalid_and_replayed_leases() {
        let mut arena = configured_arena();
        let config = test_configuration();
        assert!(arena.register(config.slot_count, 1, 1).is_err());
        assert!(arena.register(0, 0, 1).is_err());
        assert!(arena.register(0, 1, config.slot_size + 1).is_err());
        assert!(arena.register(0, 1, 8).is_err()); // no producer publication
        arena.mapping.as_ref().unwrap().publish_for_test(0, 1);
        let id = arena.register(0, 1, 8).unwrap();
        assert!(arena.register(0, 2, 8).is_err());
        {
            let registry = readers().lock().unwrap();
            let lease = registry.get(&id).unwrap();
            assert_eq!(lease.bytes(8, 0).unwrap(), b"");
            assert!(lease.bytes(8, 1).is_err());
            assert!(lease.bytes(u32::MAX, u32::MAX).is_err());
        }
        assert!(arena.release(id));
        assert!(!arena.release(id));
        assert!(!readers().lock().unwrap().contains_key(&id));
        assert!(arena.register(0, 1, 8).is_err());
        arena.mapping.as_ref().unwrap().publish_for_test(0, 2);
        let next = arena.register(0, 2, 8).unwrap();
        assert_ne!(id, next);
        assert!(!arena.release(id));
        assert!(arena.release(next));
    }

    #[test]
    fn revocation_during_read_retires_slot_and_pins_mapping() {
        let mut arena = configured_arena();
        arena.mapping.as_ref().unwrap().publish_for_test(0, 1);
        let id = arena.register(0, 1, 8).unwrap();
        with_bytes(
            SharedSource {
                lease: id,
                offset: 0,
                length: 8,
            },
            |bytes| {
                assert!(!arena.release(id));
                assert!(!readers().lock().unwrap().contains_key(&id));
                arena.mapping.as_ref().unwrap().publish_for_test(0, 2);
                assert!(arena.register(0, 2, 8).is_err());
                arena.close();
                assert!(arena.descriptor().is_err());
                assert!(arena.configure(test_configuration()).is_err());
                assert!(arena.register(1, 1, 8).is_err());
                assert_eq!(bytes, &[0; 8]);
                Ok(())
            },
        )
        .unwrap();
        arena.close();
    }

    #[test]
    fn engines_cannot_revoke_each_others_sources() {
        let mut first = configured_arena();
        let mut second = configured_arena();
        let one = first.register(0, 1, 0).unwrap();
        let two = second.register(0, 1, 0).unwrap();
        assert_ne!(one, two);
        assert!(!first.release(two));
        first.close();
        assert!(!readers().lock().unwrap().contains_key(&one));
        assert!(second.release(two));
    }

    #[test]
    fn non_default_layout_bounds_the_last_slot() {
        let mut arena = configured_arena();
        let config = test_configuration();
        let slot = config.slot_count - 1;
        let mapping = arena.mapping.as_ref().unwrap();
        let mut source = vec![0; config.slot_size as usize];
        source[0] = 17;
        source[config.slot_size as usize - 1] = 29;
        mapping.write_for_test(slot as usize, 1, &source);
        assert_eq!(mapping.published(slot as usize), 1);
        assert_eq!(mapping.published(0), 0);
        let lease = arena.register(slot, 1, config.slot_size).unwrap();
        with_bytes(
            SharedSource {
                lease,
                offset: 0,
                length: config.slot_size,
            },
            |bytes| {
                assert_eq!(bytes, source);
                Ok(())
            },
        )
        .unwrap();
        with_bytes(
            SharedSource {
                lease,
                offset: config.slot_size - 1,
                length: 1,
            },
            |bytes| {
                assert_eq!(bytes, &[29]);
                Ok(())
            },
        )
        .unwrap();
        assert!(with_bytes(
            SharedSource {
                lease,
                offset: config.slot_size,
                length: 1
            },
            |_| Ok(()),
        )
        .is_err());
        assert!(arena.release(lease));
    }
}
