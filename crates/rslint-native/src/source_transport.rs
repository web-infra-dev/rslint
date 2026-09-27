//! Private CLI source transport. The Go adapter is the sole writer. It publishes
//! a slot through IPC only after copying the complete immutable source snapshots.
//! The host registers that slot before dispatch and revokes it before replying.
//! A slot is reusable only if revocation proves every native reader has returned.
//! No mapped memory or pointer is exposed to JavaScript; AST JSON is unchanged.

mod mapping;
mod protocol_generated;

use protocol_generated::{
    SourceMapping, CAPACITY, HEADER_SIZE, PROTOCOL_VERSION, PUBLICATION_STRIDE, SLOT_COUNT,
    SLOT_SIZE,
};

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
    mapping: Option<Arc<mapping::Mapping>>,
    slots: [Slot; SLOT_COUNT],
}

#[napi]
impl SourceArena {
    #[napi(constructor)]
    pub fn new() -> Result<Self> {
        let mapping = mapping::Mapping::new(CAPACITY)
            .map_err(|error| Error::from_reason(error.to_string()))?;
        Ok(Self {
            mapping: Some(Arc::new(mapping)),
            slots: std::array::from_fn(|_| Slot::default()),
        })
    }

    #[napi]
    pub fn descriptor(&self) -> Result<SourceMapping> {
        let mapping = self.mapping.as_ref().ok_or_else(invalid)?;
        Ok(mapping.descriptor())
    }

    #[napi]
    pub fn register(&mut self, slot: u32, generation: u32, length: u32) -> Result<u32> {
        let mapping = self.mapping.as_ref().ok_or_else(invalid)?;
        let entry = self.slots.get_mut(slot as usize).ok_or_else(invalid)?;
        if entry.retired
            || entry.lease != 0
            || generation <= entry.generation
            || length as usize > SLOT_SIZE
        {
            return Err(invalid());
        }
        // All shipped targets have lock-free 32-bit atomics with four-byte
        // alignment. Pair with Go's atomic publication before borrowing data.
        // An empty slot has no source bytes requiring publication.
        if length != 0 && mapping.published(slot as usize) != generation {
            return Err(invalid());
        }
        // Never wrap and accidentally make an old descriptor valid again.
        let id = NEXT_LEASE
            .fetch_update(Ordering::Relaxed, Ordering::Relaxed, |id| id.checked_add(1))
            .map_err(|_| invalid())?;
        let lease = Arc::new(Lease {
            mapping: Arc::clone(mapping),
            start: HEADER_SIZE + slot as usize * SLOT_SIZE,
            length: length as usize,
        });
        readers().lock().map_err(|_| invalid())?.insert(id, lease);
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
            .iter_mut()
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
        for slot in &mut self.slots {
            registry.remove(&slot.lease);
            slot.lease = 0;
            slot.retired = true;
        }
        self.mapping = None;
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
mod tests {
    use super::*;

    #[test]
    fn rejects_invalid_and_replayed_leases() {
        let mut arena = SourceArena::new().unwrap();
        assert!(arena.register(SLOT_COUNT as u32, 1, 1).is_err());
        assert!(arena.register(0, 0, 1).is_err());
        assert!(arena.register(0, 1, SLOT_SIZE as u32 + 1).is_err());
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
        let mut arena = SourceArena::new().unwrap();
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
                assert!(arena.register(0, 2, 8).is_err());
                arena.close();
                assert!(arena.descriptor().is_err());
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
        let mut first = SourceArena::new().unwrap();
        let mut second = SourceArena::new().unwrap();
        let one = first.register(0, 1, 0).unwrap();
        let two = second.register(0, 1, 0).unwrap();
        assert_ne!(one, two);
        assert!(!first.release(two));
        first.close();
        assert!(!readers().lock().unwrap().contains_key(&one));
        assert!(second.release(two));
    }
}
