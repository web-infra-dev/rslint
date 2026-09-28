//! Task-independent shared byte storage. The Go peer is the sole writer and
//! publishes immutable batches before IPC registers a reader capability.
//! One lease pins every batch of a request. Reuse is acknowledged only after
//! revocation proves all native readers have returned.

mod mapping;

use napi::{Env, Error, Result, Unknown};
use napi_derive::napi;
use std::borrow::Cow;
use std::collections::{HashMap, HashSet};
use std::ops::Range;
use std::sync::{
    atomic::{AtomicU32, Ordering},
    Arc, Mutex, OnceLock,
};

use crate::NativeResult;

// Native addon statics are shared by Node's worker isolates. An Arc pins both
// the immutable batch contents and their mapping across revocation and shutdown.
static READERS: OnceLock<Mutex<HashMap<u32, Arc<Lease>>>> = OnceLock::new();
static NEXT_LEASE: AtomicU32 = AtomicU32::new(1);

fn readers() -> &'static Mutex<HashMap<u32, Arc<Lease>>> {
    READERS.get_or_init(|| Mutex::new(HashMap::new()))
}

fn invalid() -> Error {
    Error::from_reason("invalid or expired shared bytes")
}

/// Selected by the Go peer for this session, before either side maps bytes.
#[napi(object)]
#[derive(Clone, Copy)]
pub struct MemoryConfiguration {
    pub version: u32,
    pub slot_count: u32,
    pub slot_size: u32,
    pub header_size: u32,
    pub publication_stride: u32,
}

#[napi(object)]
pub struct MemoryMapping {
    pub version: u32,
    pub fd: Option<i32>,
    pub handle: Option<String>,
    pub process_id: Option<u32>,
}

#[napi(object)]
#[derive(Clone, Copy)]
pub struct MemoryBatch {
    pub slot: u32,
    pub generation: u32,
    pub length: u32,
}

// SharedBytes uses u32 offsets/lengths even when the host can map more bytes.
fn logical_length(batches: &[MemoryBatch]) -> Result<u32> {
    batches.iter().try_fold(0u32, |length, batch| {
        length.checked_add(batch.length).ok_or_else(invalid)
    })
}

/// Checked once before reserving views. Only the v1 atomic word representation
/// is fixed by the reader algorithm; every layout dimension comes from Go.
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
    fn new(config: MemoryConfiguration) -> Result<Self> {
        let invalid = || Error::from_reason("invalid shared memory configuration");
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

struct Region {
    start: usize,
    offset: usize,
    length: usize,
}

struct Lease {
    mapping: Arc<mapping::Mapping>,
    regions: Vec<Region>,
    length: usize,
}

impl Lease {
    fn range(&self, offset: u32, length: u32) -> Result<Range<usize>> {
        let offset = offset as usize;
        let length = length as usize;
        if offset > self.length || length > self.length - offset {
            return Err(invalid());
        }
        Ok(offset..offset + length)
    }

    fn chunks(&self, range: Range<usize>) -> impl Iterator<Item = &[u8]> {
        self.regions.iter().filter_map(move |region| {
            let start = range.start.max(region.offset);
            let end = range.end.min(region.offset + region.length);
            if start >= end {
                return None;
            }
            // SAFETY: registration validates every published region. The lease
            // prevents successful reuse acknowledgement while any of its bytes
            // are borrowed. Never borrow the whole arena: Go may write other slots.
            Some(unsafe {
                self.mapping
                    .bytes(region.start + start - region.offset, end - start)
            })
        })
    }

    fn bytes(&self, offset: u32, length: u32) -> Result<Cow<'_, [u8]>> {
        let range = self.range(offset, length)?;
        let mut chunks = self.chunks(range);
        let Some(first) = chunks.next() else {
            return Ok(Cow::Borrowed(&[]));
        };
        if first.len() == length as usize {
            return Ok(Cow::Borrowed(first));
        }
        // A consumer sees one logical byte range, even across batch boundaries.
        // In particular, UTF-8 consumers must decode only after this assembly.
        let mut bytes = Vec::new();
        bytes
            .try_reserve_exact(length as usize)
            .map_err(|_| invalid())?;
        bytes.extend_from_slice(first);
        for chunk in chunks {
            bytes.extend_from_slice(chunk);
        }
        debug_assert_eq!(bytes.len(), length as usize);
        Ok(Cow::Owned(bytes))
    }

    /// The range is validated before the destination is allocated. Each byte is
    /// copied once, including cross-batch ranges; no temporary assembly is needed.
    fn copy_into(&self, range: Range<usize>, mut destination: &mut [u8]) {
        debug_assert_eq!(range.len(), destination.len());
        for chunk in self.chunks(range) {
            let (head, tail) = destination.split_at_mut(chunk.len());
            head.copy_from_slice(chunk);
            destination = tail;
        }
        debug_assert!(destination.is_empty());
    }
}

#[derive(Default)]
struct Slot {
    generation: u32,
    lease: u32,
    retired: bool,
}

/// Owned by one IPC session. Failure to create or configure this optional store
/// leaves transport policy free to carry the same complete bytes inline.
#[napi]
pub struct MemoryArena {
    backing: Option<mapping::Backing>,
    mapping: Option<Arc<mapping::Mapping>>,
    // Advertised slot count alone must not cause proportional heap allocation.
    slots: HashMap<u32, Slot>,
}

#[napi]
impl MemoryArena {
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

    /// Unix inherits this empty backing when Go starts; Windows needs no handle
    /// until the peer supplies the runtime dimensions.
    #[napi]
    pub fn fd(&self) -> Option<i32> {
        self.backing
            .as_ref()
            .and_then(mapping::Backing::fd)
            .or_else(|| self.mapping.as_ref().and_then(|mapping| mapping.fd()))
    }

    /// One attempt per arena, including failures: configuration can resize or
    /// seal the backing even when a later step fails.
    #[napi]
    pub fn configure(&mut self, config: MemoryConfiguration) -> Result<()> {
        let backing = self.backing.take().ok_or_else(invalid)?;
        let layout = Layout::new(config)?;
        let mapping = backing
            .configure(layout)
            .map_err(|error| Error::from_reason(error.to_string()))?;
        self.mapping = Some(Arc::new(mapping));
        Ok(())
    }

    #[napi]
    pub fn descriptor(&self) -> Result<MemoryMapping> {
        Ok(self.mapping.as_ref().ok_or_else(invalid)?.descriptor())
    }

    /// Register all batches atomically as one ordered logical byte space. No
    /// slot state or capability is published until every batch has been checked.
    #[napi]
    pub fn register(&mut self, batches: Vec<MemoryBatch>) -> Result<u32> {
        let mapping = self.mapping.as_ref().ok_or_else(invalid)?;
        let layout = mapping.layout();
        if batches.is_empty() || batches.len() > layout.slot_count {
            return Err(invalid());
        }
        let length = logical_length(&batches)?;
        let mut seen = HashSet::new();
        seen.try_reserve(batches.len()).map_err(|_| invalid())?;
        let mut regions = Vec::new();
        regions
            .try_reserve_exact(batches.len())
            .map_err(|_| invalid())?;
        let mut offset = 0usize;
        for batch in &batches {
            if batch.slot as usize >= layout.slot_count
                || batch.length as usize > layout.slot_size
                || batch.generation == 0
                || !seen.insert(batch.slot)
                || self.slots.get(&batch.slot).is_some_and(|entry| {
                    entry.retired || entry.lease != 0 || batch.generation <= entry.generation
                })
            {
                return Err(invalid());
            }
            regions.push(Region {
                start: layout.slot_start(batch.slot as usize),
                offset,
                length: batch.length as usize,
            });
            offset += batch.length as usize;
        }
        debug_assert_eq!(offset, length as usize);
        for batch in &batches {
            // Pair with Go's release publication. Empty batches carry no bytes
            // whose visibility needs to be acquired.
            if batch.length != 0 && mapping.published(batch.slot as usize) != batch.generation {
                return Err(invalid());
            }
        }
        self.slots
            .try_reserve(batches.len())
            .map_err(|_| invalid())?;
        let mut registry = readers().lock().map_err(|_| invalid())?;
        registry.try_reserve(1).map_err(|_| invalid())?;
        // Never wrap and accidentally validate a capability from an old session.
        let id = NEXT_LEASE
            .fetch_update(Ordering::Relaxed, Ordering::Relaxed, |id| id.checked_add(1))
            .map_err(|_| invalid())?;
        registry.insert(
            id,
            Arc::new(Lease {
                mapping: Arc::clone(mapping),
                regions,
                length: length as usize,
            }),
        );
        for batch in batches {
            let entry = self.slots.entry(batch.slot).or_default();
            entry.generation = batch.generation;
            entry.lease = id;
        }
        Ok(id)
    }

    /// Revoke new reads first, then decide reuse for the whole lease. Any active
    /// native reader permanently retires every batch, even if it reads one range.
    #[napi]
    pub fn release(&mut self, lease: u32) -> bool {
        if lease == 0 || !self.slots.values().any(|slot| slot.lease == lease) {
            return false;
        }
        let mut registry = readers().lock().unwrap_or_else(|error| error.into_inner());
        let reader = registry.remove(&lease);
        let released = reader.is_some_and(|reader| Arc::try_unwrap(reader).is_ok());
        for slot in self.slots.values_mut().filter(|slot| slot.lease == lease) {
            slot.lease = 0;
            slot.retired = !released;
        }
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

impl Drop for MemoryArena {
    fn drop(&mut self) {
        self.close();
    }
}

#[cfg(feature = "test-worker-termination")]
impl MemoryArena {
    /// Test-only writer used by parser lifecycle fixtures; not a native export.
    pub(crate) fn publish_for_test(
        &self,
        slot: usize,
        generation: u32,
        bytes: &[u8],
    ) -> Result<()> {
        let mapping = self.mapping.as_ref().ok_or_else(invalid)?;
        if slot >= mapping.layout().slot_count || bytes.len() > mapping.layout().slot_size {
            return Err(invalid());
        }
        mapping.write_for_test(slot, generation, bytes);
        Ok(())
    }
}

#[napi(object)]
pub struct SharedBytes {
    pub lease: u32,
    pub offset: u32,
    pub length: u32,
}

fn acquire(bytes: &SharedBytes) -> Result<Arc<Lease>> {
    readers()
        .lock()
        .map_err(|_| invalid())?
        .get(&bytes.lease)
        .cloned()
        .ok_or_else(invalid)
}

/// The lease remains pinned for the entire callback, including a cross-batch
/// assembly. Borrowed slices cannot escape this scope or outlive revocation.
pub(crate) fn with_bytes<T>(
    bytes: SharedBytes,
    read: impl FnOnce(&[u8]) -> Result<T>,
) -> Result<T> {
    let lease = acquire(&bytes)?;
    let data = lease.bytes(bytes.offset, bytes.length)?;
    read(data.as_ref())
}

/// Return independent, writable Node memory. Never expose read-only mapped
/// pages as a Buffer or rely on JavaScript to honor Rust's immutable borrows.
#[napi(catch_unwind, ts_return_type = "Buffer")]
pub fn read_bytes(env: &Env, bytes: SharedBytes) -> NativeResult<Unknown<'_>> {
    NativeResult((|| {
        let lease = acquire(&bytes)?;
        let range = lease.range(bytes.offset, bytes.length)?;
        let mut value = std::ptr::null_mut();
        let mut destination = std::ptr::null_mut();
        // Node owns the allocation from this point. Unlike external Vec-backed
        // buffers, failure/worker termination cannot strand a Rust finalizer.
        napi::check_status!(
            unsafe {
                napi::sys::napi_create_buffer(env.raw(), range.len(), &mut destination, &mut value)
            },
            "failed to allocate byte buffer"
        )?;
        if !range.is_empty() {
            if destination.is_null() {
                return Err(Error::from_reason("missing allocated byte buffer"));
            }
            // SAFETY: successful napi_create_buffer supplies range.len() writable
            // bytes owned by this env. The value has not been exposed to JS, so
            // no user code can observe a partially copied buffer or mutate it.
            let output =
                unsafe { std::slice::from_raw_parts_mut(destination.cast::<u8>(), range.len()) };
            lease.copy_into(range, output);
        }
        // SAFETY: creation succeeded in this env. Returning the existing value
        // requires no further N-API calls; NativeResult preserves pending errors.
        Ok(unsafe { Unknown::from_raw_unchecked(env.raw(), value) })
    })())
}

#[cfg(test)]
fn test_configuration() -> MemoryConfiguration {
    MemoryConfiguration {
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

    fn batch(slot: u32, generation: u32, length: u32) -> MemoryBatch {
        MemoryBatch {
            slot,
            generation,
            length,
        }
    }

    fn configured_arena() -> MemoryArena {
        let mut arena = MemoryArena::new().unwrap();
        arena.configure(test_configuration()).unwrap();
        arena
    }

    #[test]
    fn requires_one_configuration_before_reading() {
        let mut arena = MemoryArena::new().unwrap();
        let fd = arena.fd();
        #[cfg(unix)]
        assert!(fd.is_some());
        #[cfg(windows)]
        assert!(fd.is_none());
        assert!(arena.descriptor().is_err());
        assert!(arena.register(vec![batch(0, 1, 0)]).is_err());

        let config = test_configuration();
        arena.configure(config).unwrap();
        assert_eq!(arena.fd(), fd);
        assert_eq!(arena.descriptor().unwrap().version, config.version);
        assert!(arena.slots.is_empty());
        let lease = arena.register(vec![batch(0, 1, 0)]).unwrap();
        assert!(arena.configure(config).is_err());
        assert!(arena.release(lease));
    }

    #[test]
    fn close_before_configuration_disposes_backing() {
        let mut arena = MemoryArena::new().unwrap();
        arena.close();
        arena.close();
        assert!(arena.fd().is_none());
        assert!(arena.descriptor().is_err());
        assert!(arena.configure(test_configuration()).is_err());
        assert!(arena.register(vec![batch(0, 1, 0)]).is_err());
    }

    #[test]
    fn rejects_invalid_layouts_without_allowing_reconfiguration() {
        let valid = test_configuration();
        for config in [
            MemoryConfiguration {
                version: 0,
                ..valid
            },
            MemoryConfiguration {
                version: 2,
                ..valid
            },
            MemoryConfiguration {
                slot_count: 0,
                ..valid
            },
            MemoryConfiguration {
                slot_size: 0,
                ..valid
            },
            MemoryConfiguration {
                header_size: 0,
                ..valid
            },
            MemoryConfiguration {
                header_size: 4,
                ..valid
            },
            MemoryConfiguration {
                publication_stride: 0,
                ..valid
            },
            MemoryConfiguration {
                publication_stride: 2,
                ..valid
            },
            MemoryConfiguration {
                publication_stride: 6,
                ..valid
            },
            MemoryConfiguration {
                slot_count: u32::MAX,
                slot_size: u32::MAX - 3,
                header_size: u32::MAX - 3,
                publication_stride: 4,
                ..valid
            },
        ] {
            let mut arena = MemoryArena::new().unwrap();
            assert!(arena.configure(config).is_err());
            assert!(arena.fd().is_none());
            assert!(arena.descriptor().is_err());
            assert!(arena.configure(valid).is_err());
            assert!(arena.slots.is_empty());
        }
    }

    #[test]
    fn slot_state_is_allocated_only_for_registered_slots() {
        let config = MemoryConfiguration {
            slot_count: 1_000_000,
            slot_size: 4,
            header_size: 4_000_000,
            publication_stride: 4,
            ..test_configuration()
        };
        let mut arena = MemoryArena::new().unwrap();
        arena.configure(config).unwrap();
        assert!(arena.slots.is_empty());
        let lease = arena
            .register(vec![batch(config.slot_count - 1, 1, 0)])
            .unwrap();
        assert_eq!(arena.slots.len(), 1);
        assert!(arena.release(lease));
    }

    #[test]
    fn rejects_invalid_and_replayed_leases() {
        let mut arena = configured_arena();
        let config = test_configuration();
        assert!(arena
            .register(vec![batch(config.slot_count, 1, 1)])
            .is_err());
        assert!(arena.register(vec![batch(0, 0, 1)]).is_err());
        assert!(arena
            .register(vec![batch(0, 1, config.slot_size + 1)])
            .is_err());
        assert!(arena.register(vec![batch(0, 1, 8)]).is_err()); // no producer publication
        arena.mapping.as_ref().unwrap().publish_for_test(0, 1);
        let id = arena.register(vec![batch(0, 1, 8)]).unwrap();
        assert!(arena.register(vec![batch(0, 2, 8)]).is_err());
        {
            let registry = readers().lock().unwrap();
            let lease = registry.get(&id).unwrap();
            assert_eq!(lease.bytes(8, 0).unwrap().as_ref(), b"");
            assert!(lease.bytes(8, 1).is_err());
            assert!(lease.bytes(u32::MAX, u32::MAX).is_err());
        }
        assert!(arena.release(id));
        assert!(!arena.release(id));
        assert!(!readers().lock().unwrap().contains_key(&id));
        assert!(arena.register(vec![batch(0, 1, 8)]).is_err());
        arena.mapping.as_ref().unwrap().publish_for_test(0, 2);
        let next = arena.register(vec![batch(0, 2, 8)]).unwrap();
        assert_ne!(id, next);
        assert!(!arena.release(id));
        assert!(arena.release(next));
    }

    #[test]
    fn revocation_during_read_retires_slot_and_pins_mapping() {
        let mut arena = configured_arena();
        arena.mapping.as_ref().unwrap().publish_for_test(0, 1);
        let id = arena.register(vec![batch(0, 1, 8)]).unwrap();
        with_bytes(
            SharedBytes {
                lease: id,
                offset: 0,
                length: 8,
            },
            |bytes| {
                assert!(!arena.release(id));
                assert!(!readers().lock().unwrap().contains_key(&id));
                arena.mapping.as_ref().unwrap().publish_for_test(0, 2);
                assert!(arena.register(vec![batch(0, 2, 8)]).is_err());
                arena.close();
                assert!(arena.descriptor().is_err());
                assert!(arena.configure(test_configuration()).is_err());
                assert!(arena.register(vec![batch(1, 1, 8)]).is_err());
                assert_eq!(bytes, &[0; 8]);
                Ok(())
            },
        )
        .unwrap();
        arena.close();
    }

    #[test]
    fn arenas_cannot_revoke_each_others_bytes() {
        let mut first = configured_arena();
        let mut second = configured_arena();
        let one = first.register(vec![batch(0, 1, 0)]).unwrap();
        let two = second.register(vec![batch(0, 1, 0)]).unwrap();
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
        let lease = arena
            .register(vec![batch(slot, 1, config.slot_size)])
            .unwrap();
        with_bytes(
            SharedBytes {
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
            SharedBytes {
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
            SharedBytes {
                lease,
                offset: config.slot_size,
                length: 1
            },
            |_| Ok(()),
        )
        .is_err());
        assert!(arena.release(lease));
    }

    #[test]
    fn registration_checks_all_batches_before_publishing_any_lease() {
        let mut arena = configured_arena();
        let mapping = arena.mapping.as_ref().unwrap();
        mapping.write_for_test(0, 1, b"first");
        mapping.write_for_test(1, 1, b"second");
        assert!(arena.register(vec![]).is_err());
        for batches in [
            vec![batch(0, 1, 5), batch(0, 1, 5)],
            vec![batch(0, 1, 5), batch(3, 1, 6)],
            vec![batch(0, 1, 5), batch(1, 2, 6)],
            vec![batch(0, 1, 5), batch(1, 0, 6)],
            vec![
                batch(0, 1, 5),
                batch(1, 1, test_configuration().slot_size + 1),
            ],
        ] {
            assert!(arena.register(batches).is_err());
            assert!(arena.slots.is_empty());
        }
        let occupied = arena.register(vec![batch(2, 1, 0)]).unwrap();
        assert!(arena
            .register(vec![batch(0, 1, 5), batch(2, 2, 0)])
            .is_err());
        assert!(!arena.slots.contains_key(&0));
        assert_eq!(arena.slots[&2].lease, occupied);
        let lease = arena
            .register(vec![batch(0, 1, 5), batch(1, 1, 6)])
            .unwrap();
        assert_eq!(arena.slots[&0].lease, lease);
        assert_eq!(arena.slots[&1].lease, lease);
        assert!(arena.release(lease));
        assert!(arena.release(occupied));
    }

    #[test]
    fn logical_space_must_fit_capability_offsets() {
        assert_eq!(
            logical_length(&[batch(0, 1, u32::MAX - 3), batch(1, 1, 3)]).unwrap(),
            u32::MAX
        );
        assert!(logical_length(&[batch(0, 1, u32::MAX - 3), batch(1, 1, 4)]).is_err());
    }

    #[test]
    fn ordered_batches_preserve_binary_bytes_and_cross_batch_ranges() {
        let mut arena = configured_arena();
        let mapping = arena.mapping.as_ref().unwrap();
        // Deliberately reverse physical slot order and split a UTF-8 scalar.
        mapping.write_for_test(2, 1, &[0, 0xff, 0xf0]);
        mapping.write_for_test(0, 1, &[0x9f, 0x92, 0xa9, 42, 0]);
        let id = arena
            .register(vec![batch(2, 1, 3), batch(1, 1, 0), batch(0, 1, 5)])
            .unwrap();
        {
            let registry = readers().lock().unwrap();
            let lease = registry.get(&id).unwrap();
            let first = lease.bytes(0, 2).unwrap();
            assert!(matches!(first, Cow::Borrowed(_)));
            assert_eq!(first.as_ref(), &[0, 0xff]);
            let scalar = lease.bytes(2, 4).unwrap();
            assert!(matches!(scalar, Cow::Owned(_)));
            assert_eq!(std::str::from_utf8(scalar.as_ref()).unwrap(), "\u{1f4a9}");
            let expected = [0, 0xff, 0xf0, 0x9f, 0x92, 0xa9, 42, 0];
            let all = lease.bytes(0, 8).unwrap();
            assert_eq!(all.as_ref(), expected);
            let mut copy = [0; 8];
            lease.copy_into(lease.range(0, 8).unwrap(), &mut copy);
            assert_eq!(copy, expected);
            copy.fill(9);
            assert_eq!(lease.bytes(0, 8).unwrap().as_ref(), expected);
            assert_eq!(lease.bytes(8, 0).unwrap().as_ref(), b"");
            assert!(lease.bytes(9, 0).is_err());
            assert!(lease.bytes(8, 1).is_err());
            assert!(lease.bytes(u32::MAX, 1).is_err());
        }
        assert!(arena.release(id));
        assert!(with_bytes(
            SharedBytes {
                lease: id,
                offset: 8,
                length: 0
            },
            |_| Ok(())
        )
        .is_err());
    }

    #[test]
    fn a_reader_retires_every_batch_even_when_borrowing_one_range() {
        let mut arena = configured_arena();
        let mapping = arena.mapping.as_ref().unwrap();
        mapping.write_for_test(2, 1, b"abc");
        mapping.write_for_test(0, 1, b"de");
        let id = arena
            .register(vec![batch(2, 1, 3), batch(0, 1, 2)])
            .unwrap();
        with_bytes(
            SharedBytes {
                lease: id,
                offset: 1,
                length: 1,
            },
            |bytes| {
                assert!(!arena.release(id));
                assert!(arena.slots[&0].retired && arena.slots[&2].retired);
                assert_eq!(arena.slots[&0].lease, 0);
                assert_eq!(arena.slots[&2].lease, 0);
                assert!(with_bytes(
                    SharedBytes {
                        lease: id,
                        offset: 0,
                        length: 0
                    },
                    |_| Ok(())
                )
                .is_err());
                let mapping = arena.mapping.as_ref().unwrap();
                mapping.publish_for_test(0, 2);
                mapping.publish_for_test(2, 2);
                assert!(arena.register(vec![batch(0, 2, 2)]).is_err());
                assert!(arena.register(vec![batch(2, 2, 3)]).is_err());
                let unrelated = arena.register(vec![batch(1, 1, 0)]).unwrap();
                assert!(arena.release(unrelated));
                arena.close();
                assert_eq!(bytes, b"b");
                Ok(())
            },
        )
        .unwrap();
    }

    #[test]
    fn cross_batch_reader_pins_whole_lease_through_close() {
        let mut arena = configured_arena();
        let mapping = arena.mapping.as_ref().unwrap();
        mapping.write_for_test(0, 1, b"ab");
        mapping.write_for_test(1, 1, b"cd");
        let id = arena
            .register(vec![batch(0, 1, 2), batch(1, 1, 2)])
            .unwrap();
        with_bytes(
            SharedBytes {
                lease: id,
                offset: 1,
                length: 2,
            },
            |bytes| {
                assert!(!arena.release(id));
                assert!(arena.slots[&0].retired && arena.slots[&1].retired);
                arena.close();
                assert_eq!(bytes, b"bc");
                Ok(())
            },
        )
        .unwrap();
    }

    #[test]
    fn successful_release_reuses_all_batches_without_reviving_old_capabilities() {
        let mut arena = configured_arena();
        let mapping = arena.mapping.as_ref().unwrap();
        mapping.write_for_test(0, 1, b"ab");
        mapping.write_for_test(1, 1, b"cd");
        let old = arena
            .register(vec![batch(0, 1, 2), batch(1, 1, 2)])
            .unwrap();
        assert!(arena.release(old));
        assert!(!arena.slots[&0].retired && !arena.slots[&1].retired);
        let mapping = arena.mapping.as_ref().unwrap();
        mapping.write_for_test(0, 2, b"new");
        mapping.write_for_test(1, 2, b"data");
        let current = arena
            .register(vec![batch(1, 2, 4), batch(0, 2, 3)])
            .unwrap();
        assert_ne!(old, current);
        assert!(!arena.release(old));
        assert!(with_bytes(
            SharedBytes {
                lease: old,
                offset: 0,
                length: 1
            },
            |_| Ok(())
        )
        .is_err());
        with_bytes(
            SharedBytes {
                lease: current,
                offset: 0,
                length: 7,
            },
            |bytes| {
                assert_eq!(bytes, b"datanew"); // cspell:ignore datanew
                Ok(())
            },
        )
        .unwrap();
        assert!(arena.release(current));
    }
}
