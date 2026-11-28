//! Shared-memory header layout, status codes, N-API data objects, and per-slot
//! host synchronization primitives for the bidirectional RPC transport.

use napi::{Error, Result};
use napi_derive::napi;
use std::sync::{
    atomic::{AtomicBool, AtomicU32, Ordering},
    Condvar, Mutex, OnceLock,
};
use std::thread;
use std::time::Duration;

// ── Shared header layout (matches internal/ipc/rpc_channel.go) ───────────────

pub(super) const RPC_MAGIC: u32 = 0x5253_5031; // "RSP1"
pub(super) const GLOBAL_MAGIC_OFFSET: usize = 0;
pub(super) const GLOBAL_SERVER_SLEEPING_OFFSET: usize = 4;
pub(super) const GLOBAL_SERVER_CLOSED_OFFSET: usize = 8;
pub(super) const GLOBAL_CLIENT_CLOSED_OFFSET: usize = 12;
pub(super) const GLOBAL_HEADER_SIZE: usize = 128;

pub(super) const SLOT_STATE_OFF: usize = 0;
pub(super) const SLOT_REQ_SEQ_OFF: usize = 4;
pub(super) const SLOT_RESP_SEQ_OFF: usize = 8;
pub(super) const SLOT_CANCEL_SEQ_OFF: usize = 12;
pub(super) const SLOT_WORKER_EPOCH_OFF: usize = 16;
pub(super) const SLOT_REQ_LEN_OFF: usize = 20;
pub(super) const SLOT_RESP_LEN_OFF: usize = 24;
pub(super) const SLOT_RESP_STATUS_OFF: usize = 28;
pub(super) const SLOT_CLIENT_SLEEPING_OFF: usize = 32;
pub(super) const SLOT_TOTAL_RESP_LEN_OFF: usize = 36;
const SLOT_HEADER_MIN_STRIDE: usize = 64;

pub(super) const STATE_IDLE: u32 = 0;
pub(super) const STATE_REQ_READY: u32 = 1;
pub(super) const STATE_IN_FLIGHT: u32 = 2;
pub(super) const STATE_RESP_READY: u32 = 3;
pub(super) const STATE_RESP_CHUNK: u32 = 4;
pub(super) const STATE_CHUNK_ACK: u32 = 5;
pub(super) const STATE_ABANDONED: u32 = 6;

// Wire response status values written by Go in SLOT_RESP_STATUS.
pub(super) const WIRE_STATUS_OK: u32 = 0;
pub(super) const WIRE_STATUS_CLOSED: u32 = 1;
pub(super) const WIRE_STATUS_CANCELLED: u32 = 2;
pub(super) const WIRE_STATUS_ERROR: u32 = 3;

// Status codes written into JS `statusOut[0]`.
pub(super) const JS_STATUS_OK: i32 = 0;
pub(super) const JS_STATUS_CLOSED: i32 = 1;
pub(super) const JS_STATUS_CANCELLED: i32 = 2;
pub(super) const JS_STATUS_ERROR: i32 = 3;
/// Transient per-call fallback (e.g. request larger than slot, slot quarantined/busy).
pub(super) const JS_STATUS_FALLBACK: i32 = 4;
/// Permanent transport or slot-lease unavailability (closed channel, dead peer, stale epoch).
pub(super) const JS_STATUS_UNAVAILABLE: i32 = 5;

pub(super) const MAX_REPLY_BYTES: usize = 64 * 1024 * 1024;
pub(super) const CONDVAR_POLL_TIMEOUT: Duration = Duration::from_millis(10);

static SPIN_DURATION_US: OnceLock<u64> = OnceLock::new();

pub(super) fn spin_duration() -> Duration {
    let us = *SPIN_DURATION_US.get_or_init(|| {
        let cores = thread::available_parallelism()
            .map(|n| n.get())
            .unwrap_or(1);
        if cores <= 2 {
            0
        } else if cores <= 4 {
            50
        } else {
            200
        }
    });
    Duration::from_micros(us)
}

#[napi(object)]
#[derive(Clone, Copy)]
pub struct RpcConfiguration {
    pub version: u32,
    pub slot_count: u32,
    pub slot_size: u32,
    pub header_size: u32,
    pub slot_header_stride: u32,
}

#[napi(object)]
pub struct RpcMapping {
    pub version: u32,
    pub fd: Option<i32>,
    pub doorbell_fd: Option<i32>,
    pub handle: Option<String>,
    pub node_to_go_event: Option<String>,
    pub go_to_node_event: Option<String>,
    pub process_id: Option<u32>,
}

#[napi(object)]
pub struct RpcSlotLease {
    pub slot: u32,
    pub epoch: u32,
}

#[derive(Clone, Copy)]
pub(super) struct Layout {
    pub version: u32,
    pub slot_count: usize,
    pub slot_size: usize,
    pub header_size: usize,
    pub slot_header_stride: usize,
    pub capacity: usize,
}

impl Layout {
    pub(super) fn new(config: RpcConfiguration) -> Result<Self> {
        let bad = || Error::from_reason("invalid RPC shared memory configuration");
        let slot_count = usize::try_from(config.slot_count).map_err(|_| bad())?;
        let slot_size = usize::try_from(config.slot_size).map_err(|_| bad())?;
        let header_size = usize::try_from(config.header_size).map_err(|_| bad())?;
        let slot_header_stride = usize::try_from(config.slot_header_stride).map_err(|_| bad())?;
        let align = std::mem::align_of::<AtomicU32>();
        if config.version != 1
            || slot_count == 0
            || slot_count > 64
            || slot_size < 4096
            || slot_header_stride < SLOT_HEADER_MIN_STRIDE
            || slot_header_stride % align != 0
        {
            return Err(bad());
        }
        let header_needed = slot_count
            .checked_mul(slot_header_stride)
            .and_then(|slots| GLOBAL_HEADER_SIZE.checked_add(slots))
            .ok_or_else(bad)?;
        if header_size < header_needed {
            return Err(bad());
        }
        let capacity = slot_count
            .checked_mul(slot_size)
            .and_then(|payload| header_size.checked_add(payload))
            .filter(|cap| *cap <= isize::MAX as usize)
            .ok_or_else(bad)?;
        Ok(Self {
            version: config.version,
            slot_count,
            slot_size,
            header_size,
            slot_header_stride,
            capacity,
        })
    }

    #[inline]
    pub(super) fn slot_header_offset(self, slot: usize) -> usize {
        debug_assert!(slot < self.slot_count);
        GLOBAL_HEADER_SIZE + slot * self.slot_header_stride
    }

    #[inline]
    pub(super) fn slot_payload_offset(self, slot: usize) -> usize {
        debug_assert!(slot < self.slot_count);
        self.header_size + slot * self.slot_size
    }
}

pub(super) struct SlotHost {
    /// True while leased to a worker slot in WorkerPool.
    pub(super) assigned: AtomicBool,
    /// True while a host thread is actively executing `call_sync_core` on this slot.
    pub(super) in_use: AtomicBool,
    /// Active lease epoch (0 when unassigned; non-zero while assigned to one worker).
    pub(super) epoch: AtomicU32,
    /// Monotonic epoch allocator (skips 0).
    next_epoch: AtomicU32,
    /// Monotonic request sequence for this slot (skips 0; survives respawns).
    next_seq: AtomicU32,
    /// Sequence currently in flight (0 when idle), read by `rpc_abort_slot`.
    pub(super) active_seq: AtomicU32,
    /// True while a worker thread is sleeping on `wait_cv`.
    pub(super) waiting: AtomicBool,
    pub(super) wait_mu: Mutex<()>,
    pub(super) wait_cv: Condvar,
}

impl SlotHost {
    pub(super) fn new() -> Self {
        Self {
            assigned: AtomicBool::new(false),
            in_use: AtomicBool::new(false),
            epoch: AtomicU32::new(0),
            next_epoch: AtomicU32::new(1),
            next_seq: AtomicU32::new(1),
            active_seq: AtomicU32::new(0),
            waiting: AtomicBool::new(false),
            wait_mu: Mutex::new(()),
            wait_cv: Condvar::new(),
        }
    }

    pub(super) fn alloc_epoch(&self) -> u32 {
        loop {
            let cur = self.next_epoch.load(Ordering::Relaxed);
            let next = if cur == u32::MAX { 1 } else { cur + 1 };
            if self
                .next_epoch
                .compare_exchange_weak(cur, next, Ordering::Relaxed, Ordering::Relaxed)
                .is_ok()
            {
                return cur;
            }
        }
    }

    pub(super) fn next_req_seq(&self) -> u32 {
        loop {
            let cur = self.next_seq.load(Ordering::Relaxed);
            let next = if cur == u32::MAX { 1 } else { cur + 1 };
            if self
                .next_seq
                .compare_exchange_weak(cur, next, Ordering::Relaxed, Ordering::Relaxed)
                .is_ok()
            {
                return cur;
            }
        }
    }

    pub(super) fn wake(&self) {
        if self.waiting.load(Ordering::SeqCst) {
            let _guard = self.wait_mu.lock().unwrap_or_else(|e| e.into_inner());
            self.wait_cv.notify_all();
        }
    }
}

pub(super) struct BoolResetGuard<'a>(pub(super) &'a AtomicBool);
impl Drop for BoolResetGuard<'_> {
    fn drop(&mut self) {
        self.0.store(false, Ordering::Release);
    }
}

pub(super) struct ActiveSeqGuard<'a>(pub(super) &'a AtomicU32);
impl Drop for ActiveSeqGuard<'_> {
    fn drop(&mut self) {
        self.0.store(0, Ordering::Release);
    }
}
