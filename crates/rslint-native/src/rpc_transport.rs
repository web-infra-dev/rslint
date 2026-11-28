//! Generic bidirectional shared-memory RPC transport between Node plugin worker
//! threads and the Go IPC server.
//!
//! Design invariants:
//! - Main thread owns `RpcChannel` (and the underlying SHM + OS doorbell).
//! - Worker threads hold ONLY plain integer handles (`channelId`, `slot`, `epoch`)
//!   and invoke stateless `rpc_call_sync(...)` — zero worker-isolate finalizers.
//! - Host-side `SlotHost::in_use` + `SlotHost::epoch` enforce exclusive ownership
//!   before any shared-memory payload read/write, preventing UB even under
//!   concurrent or stale `rpc_call_sync` invocations.
//! - All shared control words are accessed strictly through `AtomicU32::from_ptr`
//!   at explicit offsets (128-byte cache-line isolated). Never construct whole-
//!   mapping `&[u8]` or `&mut [u8]` references.
//! - Every Dekker sleep/wake pair uses `Ordering::SeqCst` on both sides so ARM64
//!   (Apple Silicon `+rcpc`) never reorders StoreLoad across processes.
//! - Replies larger than one slot stream over the same slot via `RESP_CHUNK` ↔
//!   `CHUNK_ACK` without re-executing the server handler.
//! - Cancel/timeout/close abandons in-flight slots via CAS (`IN_FLIGHT -> ABANDONED`)
//!   and quarantines the slot until Go resets it to `IDLE`.
// cspell:words finalizers rcpc nonoverlapping syscalls typedarray arraybuffer DONTWAIT

mod platform;

use napi::{Env, Error, JsValue, Result, Unknown};
use napi_derive::napi;
use std::collections::HashMap;
use std::hint;
use std::ptr;
use std::sync::{
    atomic::{AtomicBool, AtomicI32, AtomicU32, Ordering},
    Arc, Condvar, Mutex, OnceLock,
};
use std::thread;
use std::time::{Duration, Instant};

use crate::NativeResult;
use platform::DoorbellEvent;

// ── Shared header layout (matches internal/ipc/rpc_channel.go) ───────────────

pub(super) const RPC_MAGIC: u32 = 0x5253_5031; // "RSP1"
pub(super) const GLOBAL_MAGIC_OFFSET: usize = 0;
const GLOBAL_SERVER_SLEEPING_OFFSET: usize = 4;
const GLOBAL_SERVER_CLOSED_OFFSET: usize = 8;
const GLOBAL_CLIENT_CLOSED_OFFSET: usize = 12;
const GLOBAL_HEADER_SIZE: usize = 128;

const SLOT_STATE_OFF: usize = 0;
const SLOT_REQ_SEQ_OFF: usize = 4;
const SLOT_RESP_SEQ_OFF: usize = 8;
const SLOT_CANCEL_SEQ_OFF: usize = 12;
const SLOT_WORKER_EPOCH_OFF: usize = 16;
const SLOT_REQ_LEN_OFF: usize = 20;
const SLOT_RESP_LEN_OFF: usize = 24;
const SLOT_RESP_STATUS_OFF: usize = 28;
const SLOT_CLIENT_SLEEPING_OFF: usize = 32;
const SLOT_TOTAL_RESP_LEN_OFF: usize = 36;
const SLOT_HEADER_MIN_STRIDE: usize = 64;

const STATE_IDLE: u32 = 0;
const STATE_REQ_READY: u32 = 1;
const STATE_IN_FLIGHT: u32 = 2;
const STATE_RESP_READY: u32 = 3;
const STATE_RESP_CHUNK: u32 = 4;
const STATE_CHUNK_ACK: u32 = 5;
const STATE_ABANDONED: u32 = 6;

// Wire response status values written by Go in SLOT_RESP_STATUS.
const WIRE_STATUS_OK: u32 = 0;
const WIRE_STATUS_CLOSED: u32 = 1;
const WIRE_STATUS_CANCELLED: u32 = 2;
const WIRE_STATUS_ERROR: u32 = 3;

// Status codes written into JS `statusOut[0]`.
const JS_STATUS_OK: i32 = 0;
const JS_STATUS_CLOSED: i32 = 1;
const JS_STATUS_CANCELLED: i32 = 2;
const JS_STATUS_ERROR: i32 = 3;
/// Transient per-call fallback (e.g. request larger than slot, slot quarantined/busy).
const JS_STATUS_FALLBACK: i32 = 4;
/// Permanent transport or slot-lease unavailability (closed channel, dead peer, stale epoch).
const JS_STATUS_UNAVAILABLE: i32 = 5;

const MAX_REPLY_BYTES: usize = 64 * 1024 * 1024;
const CONDVAR_POLL_TIMEOUT: Duration = Duration::from_millis(10);

static CHANNELS: OnceLock<Mutex<HashMap<u32, Arc<RpcChannelInner>>>> = OnceLock::new();
static NEXT_CHANNEL_ID: AtomicU32 = AtomicU32::new(1);
static SPIN_DURATION_US: OnceLock<u64> = OnceLock::new();

fn channels() -> &'static Mutex<HashMap<u32, Arc<RpcChannelInner>>> {
    CHANNELS.get_or_init(|| Mutex::new(HashMap::new()))
}

fn spin_duration() -> Duration {
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

fn invalid() -> Error {
    Error::from_reason("invalid or closed RPC channel")
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
    fn new(config: RpcConfiguration) -> Result<Self> {
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
    fn slot_header_offset(self, slot: usize) -> usize {
        debug_assert!(slot < self.slot_count);
        GLOBAL_HEADER_SIZE + slot * self.slot_header_stride
    }

    #[inline]
    fn slot_payload_offset(self, slot: usize) -> usize {
        debug_assert!(slot < self.slot_count);
        self.header_size + slot * self.slot_size
    }
}

struct SlotHost {
    /// True while leased to a worker slot in WorkerPool.
    assigned: AtomicBool,
    /// True while a host thread is actively executing `call_sync_core` on this slot.
    in_use: AtomicBool,
    /// Active lease epoch (0 when unassigned; non-zero while assigned to one worker).
    epoch: AtomicU32,
    /// Monotonic epoch allocator (skips 0).
    next_epoch: AtomicU32,
    /// Monotonic request sequence for this slot (skips 0; survives respawns).
    next_seq: AtomicU32,
    /// Sequence currently in flight (0 when idle), read by `rpc_abort_slot`.
    active_seq: AtomicU32,
    /// True while a worker thread is sleeping on `wait_cv`.
    waiting: AtomicBool,
    wait_mu: Mutex<()>,
    wait_cv: Condvar,
}

impl SlotHost {
    fn new() -> Self {
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

    fn alloc_epoch(&self) -> u32 {
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

    fn next_req_seq(&self) -> u32 {
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

    fn wake(&self) {
        if self.waiting.load(Ordering::SeqCst) {
            let _guard = self.wait_mu.lock().unwrap_or_else(|e| e.into_inner());
            self.wait_cv.notify_all();
        }
    }
}

struct BoolResetGuard<'a>(&'a AtomicBool);
impl Drop for BoolResetGuard<'_> {
    fn drop(&mut self) {
        self.0.store(false, Ordering::Release);
    }
}

struct ActiveSeqGuard<'a>(&'a AtomicU32);
impl Drop for ActiveSeqGuard<'_> {
    fn drop(&mut self) {
        self.0.store(0, Ordering::Release);
    }
}

struct RpcChannelInner {
    mapping: platform::Mapping,
    slots: Vec<SlotHost>,
    closed: AtomicBool,
    peer_closed: AtomicBool,
    reader_started: AtomicBool,
    reader_start_mu: Mutex<()>,
}

impl RpcChannelInner {
    fn new(mapping: platform::Mapping) -> Self {
        let count = mapping.layout().slot_count;
        let mut slots = Vec::with_capacity(count);
        for _ in 0..count {
            slots.push(SlotHost::new());
        }
        Self {
            mapping,
            slots,
            closed: AtomicBool::new(false),
            peer_closed: AtomicBool::new(false),
            reader_started: AtomicBool::new(false),
            reader_start_mu: Mutex::new(()),
        }
    }

    #[inline]
    fn global_word(&self, offset: usize) -> &AtomicU32 {
        debug_assert!(offset + 4 <= GLOBAL_HEADER_SIZE);
        // SAFETY: mapping is valid for the lifetime of Self, offset is 4-byte aligned in header.
        unsafe { AtomicU32::from_ptr(self.mapping.base().add(offset).cast()) }
    }

    #[inline]
    fn slot_word(&self, slot: usize, field_offset: usize) -> &AtomicU32 {
        let off = self.mapping.layout().slot_header_offset(slot) + field_offset;
        debug_assert!(off + 4 <= self.mapping.layout().header_size);
        // SAFETY: mapping is valid for the lifetime of Self, offset is 4-byte aligned in slot header.
        unsafe { AtomicU32::from_ptr(self.mapping.base().add(off).cast()) }
    }

    #[inline]
    fn slot_payload_ptr(&self, slot: usize) -> *mut u8 {
        let off = self.mapping.layout().slot_payload_offset(slot);
        // SAFETY: offset is within mapped capacity.
        unsafe { self.mapping.base().add(off) }
    }

    fn is_unavailable(&self) -> bool {
        self.closed.load(Ordering::SeqCst)
            || self.peer_closed.load(Ordering::SeqCst)
            || self
                .global_word(GLOBAL_SERVER_CLOSED_OFFSET)
                .load(Ordering::SeqCst)
                != 0
    }

    fn wake_all_slots(&self) {
        for slot in &self.slots {
            let _guard = slot.wait_mu.lock().unwrap_or_else(|e| e.into_inner());
            slot.wait_cv.notify_all();
        }
    }

    fn ensure_reader_started(self: &Arc<Self>) {
        if self.reader_started.load(Ordering::Acquire) {
            return;
        }
        let _guard = self.reader_start_mu.lock().unwrap_or_else(|e| e.into_inner());
        if self.reader_started.load(Ordering::Relaxed) || self.closed.load(Ordering::SeqCst) {
            return;
        }
        let cloned = Arc::clone(self);
        let spawn = thread::Builder::new()
            .name("rslint-rpc-wake".into())
            .spawn(move || {
                loop {
                    if cloned.closed.load(Ordering::SeqCst) {
                        break;
                    }
                    match cloned.mapping.wait_go_to_node() {
                        DoorbellEvent::Wake => {
                            for (idx, slot) in cloned.slots.iter().enumerate() {
                                if slot.waiting.load(Ordering::SeqCst) {
                                    let st = cloned
                                        .slot_word(idx, SLOT_STATE_OFF)
                                        .load(Ordering::SeqCst);
                                    if st == STATE_RESP_READY
                                        || st == STATE_RESP_CHUNK
                                        || st == STATE_IDLE
                                        || st == STATE_ABANDONED
                                        || cloned.is_unavailable()
                                    {
                                        slot.wake();
                                    }
                                }
                            }
                        }
                        DoorbellEvent::Closed => {
                            cloned.peer_closed.store(true, Ordering::SeqCst);
                            cloned.wake_all_slots();
                            break;
                        }
                    }
                }
            });
        if spawn.is_ok() {
            self.reader_started.store(true, Ordering::Release);
        } else {
            self.peer_closed.store(true, Ordering::SeqCst);
            self.wake_all_slots();
        }
    }

    fn close(&self) {
        if self.closed.swap(true, Ordering::SeqCst) {
            return;
        }
        self.global_word(GLOBAL_CLIENT_CLOSED_OFFSET)
            .store(1, Ordering::SeqCst);
        self.mapping.ring_node_to_go();
        self.mapping.signal_close();
        self.wake_all_slots();
    }

    fn acquire_slot(&self) -> Option<RpcSlotLease> {
        if self.is_unavailable() {
            return None;
        }
        for (idx, slot) in self.slots.iter().enumerate() {
            if slot.in_use.load(Ordering::Acquire)
                || self.slot_word(idx, SLOT_STATE_OFF).load(Ordering::Acquire) != STATE_IDLE
            {
                continue;
            }
            if slot
                .assigned
                .compare_exchange(false, true, Ordering::AcqRel, Ordering::Relaxed)
                .is_ok()
            {
                // Re-check after winning exclusive assignment so a racing call or
                // quarantined slot can never be leased out. Note: `epoch` stays 0
                // until both checks pass, so no stale worker can match it.
                if slot.in_use.load(Ordering::Acquire)
                    || self.slot_word(idx, SLOT_STATE_OFF).load(Ordering::Acquire) != STATE_IDLE
                {
                    slot.assigned.store(false, Ordering::Release);
                    continue;
                }
                let epoch = slot.alloc_epoch();
                slot.epoch.store(epoch, Ordering::SeqCst);
                return Some(RpcSlotLease {
                    slot: idx as u32,
                    epoch,
                });
            }
        }
        None
    }

    fn release_slot(&self, slot_idx: usize) {
        if let Some(slot) = self.slots.get(slot_idx) {
            // Invalidate epoch BEFORE aborting and clearing `assigned` so no stale
            // worker can pass the epoch check while the slot is being recycled.
            slot.epoch.store(0, Ordering::SeqCst);
            self.abort_slot(slot_idx);
            slot.assigned.store(false, Ordering::Release);
        }
    }

    fn abort_slot(&self, slot_idx: usize) {
        let Some(slot) = self.slots.get(slot_idx) else {
            return;
        };
        let seq = slot.active_seq.load(Ordering::SeqCst);
        if seq != 0 {
            self.slot_word(slot_idx, SLOT_CANCEL_SEQ_OFF)
                .store(seq, Ordering::SeqCst);
        }
        let _guard = slot.wait_mu.lock().unwrap_or_else(|e| e.into_inner());
        slot.wait_cv.notify_all();
    }

    /// Cleanly abandon an in-flight request without leaving the slot stuck in
    /// `RESP_READY` or `RESP_CHUNK`, and without aliasing payload memory while
    /// Go is still executing.
    fn abandon_slot(&self, slot_idx: usize, req_seq: u32) {
        self.slot_word(slot_idx, SLOT_CANCEL_SEQ_OFF)
            .store(req_seq, Ordering::SeqCst);
        let state_word = self.slot_word(slot_idx, SLOT_STATE_OFF);
        loop {
            let cur = state_word.load(Ordering::SeqCst);
            match cur {
                STATE_IDLE | STATE_ABANDONED => break,
                STATE_REQ_READY | STATE_RESP_READY => {
                    if state_word
                        .compare_exchange(cur, STATE_IDLE, Ordering::SeqCst, Ordering::Acquire)
                        .is_ok()
                    {
                        break;
                    }
                }
                STATE_IN_FLIGHT => {
                    if state_word
                        .compare_exchange(
                            STATE_IN_FLIGHT,
                            STATE_ABANDONED,
                            Ordering::SeqCst,
                            Ordering::Acquire,
                        )
                        .is_ok()
                    {
                        break;
                    }
                }
                STATE_RESP_CHUNK | STATE_CHUNK_ACK => {
                    if state_word
                        .compare_exchange(
                            cur,
                            STATE_ABANDONED,
                            Ordering::SeqCst,
                            Ordering::Acquire,
                        )
                        .is_ok()
                    {
                        // Wake Go's chunk writer so it observes ABANDONED and resets to IDLE.
                        self.mapping.ring_node_to_go();
                        break;
                    }
                }
                _ => break,
            }
        }
    }

    /// Unified synchronous RPC state machine shared by both N-API production
    /// calls (`call_sync_napi`) and Rust unit tests (`call_sync_raw`).
    ///
    /// `emit_payload(prefix_chunks, tail_ptr, tail_len)` constructs the caller's
    /// owned reply directly from SHM (single-copy for replies <= slot_size).
    fn call_sync_core<R, F, Emit>(
        self: &Arc<Self>,
        slot_idx: usize,
        worker_epoch: u32,
        req_ptr: *const u8,
        req_len: usize,
        is_cancelled: F,
        emit_payload: Emit,
    ) -> Result<(i32, Option<R>)>
    where
        F: Fn() -> bool,
        Emit: FnOnce(Option<&[u8]>, *const u8, usize) -> Result<R>,
    {
        if is_cancelled() {
            return Ok((JS_STATUS_CANCELLED, None));
        }
        if self.is_unavailable() || worker_epoch == 0 {
            return Ok((JS_STATUS_UNAVAILABLE, None));
        }
        let layout = self.mapping.layout();
        if slot_idx >= layout.slot_count {
            return Ok((JS_STATUS_UNAVAILABLE, None));
        }
        if req_len > layout.slot_size {
            return Ok((JS_STATUS_FALLBACK, None));
        }

        let slot = &self.slots[slot_idx];
        // Enforce host-side exclusive ownership BEFORE touching shared memory so
        // concurrent or stale callers can never race on slot payload/headers.
        if slot
            .in_use
            .compare_exchange(false, true, Ordering::AcqRel, Ordering::Acquire)
            .is_err()
        {
            return Ok((JS_STATUS_FALLBACK, None));
        }
        let _in_use_guard = BoolResetGuard(&slot.in_use);

        if slot.epoch.load(Ordering::Acquire) != worker_epoch {
            return Ok((JS_STATUS_UNAVAILABLE, None));
        }

        // If this slot is still quarantined from a previously abandoned call,
        // fall back cleanly for this call so we never alias payload memory.
        let state_word = self.slot_word(slot_idx, SLOT_STATE_OFF);
        if state_word.load(Ordering::Acquire) != STATE_IDLE {
            return Ok((JS_STATUS_FALLBACK, None));
        }

        self.ensure_reader_started();
        if self.is_unavailable() {
            return Ok((JS_STATUS_UNAVAILABLE, None));
        }

        let req_seq = slot.next_req_seq();

        // Copy request payload into this exclusively owned slot.
        if req_len > 0 {
            // SAFETY: slot is exclusively owned (`in_use == true`, valid `epoch`, `STATE_IDLE`),
            // `req_len <= layout.slot_size`, and `req_ptr` is backed by a non-shared buffer.
            unsafe {
                ptr::copy_nonoverlapping(req_ptr, self.slot_payload_ptr(slot_idx), req_len);
            }
        }

        // Clear per-request header words (including SLOT_CANCEL_SEQ / SLOT_RESP_SEQ to
        // prevent wraparound false positives) BEFORE publishing `active_seq`.
        self.slot_word(slot_idx, SLOT_CANCEL_SEQ_OFF)
            .store(0, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_RESP_SEQ_OFF)
            .store(0, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_REQ_SEQ_OFF)
            .store(req_seq, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_WORKER_EPOCH_OFF)
            .store(worker_epoch, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_REQ_LEN_OFF)
            .store(req_len as u32, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_RESP_LEN_OFF)
            .store(0, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_RESP_STATUS_OFF)
            .store(0, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_TOTAL_RESP_LEN_OFF)
            .store(0, Ordering::Relaxed);
        self.slot_word(slot_idx, SLOT_CLIENT_SLEEPING_OFF)
            .store(0, Ordering::Relaxed);

        // Publish active_seq with SeqCst and re-check cancel/abort/epoch BEFORE
        // making the request visible to Go, closing the abort_slot submission race.
        slot.active_seq.store(req_seq, Ordering::SeqCst);
        let _active_guard = ActiveSeqGuard(&slot.active_seq);
        if is_cancelled()
            || self
                .slot_word(slot_idx, SLOT_CANCEL_SEQ_OFF)
                .load(Ordering::SeqCst)
                == req_seq
        {
            return Ok((JS_STATUS_CANCELLED, None));
        }
        if self.is_unavailable() || slot.epoch.load(Ordering::SeqCst) != worker_epoch {
            return Ok((JS_STATUS_UNAVAILABLE, None));
        }

        // Dekker SeqCst pair: publish REQ_READY, then read server_sleeping.
        state_word.store(STATE_REQ_READY, Ordering::SeqCst);
        if self
            .global_word(GLOBAL_SERVER_SLEEPING_OFFSET)
            .load(Ordering::SeqCst)
            != 0
        {
            self.mapping.ring_node_to_go();
        }

        let mut chunks: Option<Vec<u8>> = None;
        let mut expected_total: usize = 0;

        loop {
            match self.wait_for_response_state(slot_idx, req_seq, &is_cancelled) {
                WaitOutcome::Cancelled => {
                    self.abandon_slot(slot_idx, req_seq);
                    return Ok((JS_STATUS_CANCELLED, None));
                }
                WaitOutcome::Unavailable => {
                    let cancelled = is_cancelled()
                        || self
                            .slot_word(slot_idx, SLOT_CANCEL_SEQ_OFF)
                            .load(Ordering::Acquire)
                            == req_seq;
                    self.abandon_slot(slot_idx, req_seq);
                    return Ok((
                        if cancelled {
                            JS_STATUS_CANCELLED
                        } else if self.is_unavailable() {
                            JS_STATUS_UNAVAILABLE
                        } else {
                            JS_STATUS_FALLBACK
                        },
                        None,
                    ));
                }
                WaitOutcome::Chunk => {
                    let resp_seq = self
                        .slot_word(slot_idx, SLOT_RESP_SEQ_OFF)
                        .load(Ordering::Acquire);
                    let resp_len = self
                        .slot_word(slot_idx, SLOT_RESP_LEN_OFF)
                        .load(Ordering::Acquire) as usize;
                    let total_len = self
                        .slot_word(slot_idx, SLOT_TOTAL_RESP_LEN_OFF)
                        .load(Ordering::Acquire) as usize;
                    if resp_seq != req_seq
                        || resp_len == 0
                        || resp_len > layout.slot_size
                        || total_len == 0
                        || total_len > MAX_REPLY_BYTES
                    {
                        self.abandon_slot(slot_idx, req_seq);
                        return Ok((JS_STATUS_FALLBACK, None));
                    }
                    if chunks.is_none() {
                        let mut v = Vec::new();
                        if v.try_reserve_exact(total_len).is_err() {
                            self.abandon_slot(slot_idx, req_seq);
                            return Ok((JS_STATUS_FALLBACK, None));
                        }
                        expected_total = total_len;
                        chunks = Some(v);
                    }
                    let buf = chunks.as_mut().unwrap();
                    if total_len != expected_total
                        || buf.len().saturating_add(resp_len) > expected_total
                    {
                        self.abandon_slot(slot_idx, req_seq);
                        return Ok((JS_STATUS_FALLBACK, None));
                    }
                    let start = buf.len();
                    buf.resize(start + resp_len, 0);
                    // SAFETY: state == RESP_CHUNK grants exclusive read of payload[0..resp_len].
                    unsafe {
                        ptr::copy_nonoverlapping(
                            self.slot_payload_ptr(slot_idx),
                            buf.as_mut_ptr().add(start),
                            resp_len,
                        );
                    }
                    state_word.store(STATE_CHUNK_ACK, Ordering::SeqCst);
                    // Always ring doorbell on chunk ACK (rare >64 KiB path) to wake chunk writer immediately.
                    self.mapping.ring_node_to_go();
                }
                WaitOutcome::Ready => {
                    let resp_seq = self
                        .slot_word(slot_idx, SLOT_RESP_SEQ_OFF)
                        .load(Ordering::Acquire);
                    let resp_len = self
                        .slot_word(slot_idx, SLOT_RESP_LEN_OFF)
                        .load(Ordering::Acquire) as usize;
                    let wire_status = self
                        .slot_word(slot_idx, SLOT_RESP_STATUS_OFF)
                        .load(Ordering::Acquire);
                    if resp_seq != req_seq || resp_len > layout.slot_size {
                        state_word.store(STATE_IDLE, Ordering::Release);
                        return Ok((JS_STATUS_FALLBACK, None));
                    }

                    let js_status = match wire_status {
                        WIRE_STATUS_OK => JS_STATUS_OK,
                        WIRE_STATUS_CLOSED => JS_STATUS_CLOSED,
                        WIRE_STATUS_CANCELLED => JS_STATUS_CANCELLED,
                        WIRE_STATUS_ERROR => JS_STATUS_ERROR,
                        _ => JS_STATUS_FALLBACK,
                    };

                    if js_status != JS_STATUS_OK && js_status != JS_STATUS_ERROR {
                        state_word.store(STATE_IDLE, Ordering::Release);
                        return Ok((js_status, None));
                    }

                    if let Some(ref buf) = chunks {
                        let Some(total) = buf.len().checked_add(resp_len) else {
                            state_word.store(STATE_IDLE, Ordering::Release);
                            return Ok((JS_STATUS_FALLBACK, None));
                        };
                        if total != expected_total || total > MAX_REPLY_BYTES {
                            state_word.store(STATE_IDLE, Ordering::Release);
                            return Ok((JS_STATUS_FALLBACK, None));
                        }
                    }

                    let emitted = emit_payload(
                        chunks.as_deref(),
                        self.slot_payload_ptr(slot_idx),
                        resp_len,
                    );
                    state_word.store(STATE_IDLE, Ordering::Release);
                    return Ok((js_status, Some(emitted?)));
                }
            }
        }
    }

    #[cfg(test)]
    fn call_sync_raw<F>(
        self: &Arc<Self>,
        slot_idx: usize,
        worker_epoch: u32,
        req_ptr: *const u8,
        req_len: usize,
        is_cancelled: F,
    ) -> (i32, Option<Vec<u8>>)
    where
        F: Fn() -> bool,
    {
        self.call_sync_core(
            slot_idx,
            worker_epoch,
            req_ptr,
            req_len,
            is_cancelled,
            |prefix, tail_ptr, tail_len| {
                let prefix_len = prefix.map_or(0, <[u8]>::len);
                let mut out = vec![0u8; prefix_len + tail_len];
                if let Some(p) = prefix {
                    out[..prefix_len].copy_from_slice(p);
                }
                if tail_len > 0 {
                    // SAFETY: state == RESP_READY grants exclusive read of tail_ptr[0..tail_len].
                    unsafe {
                        ptr::copy_nonoverlapping(
                            tail_ptr,
                            out.as_mut_ptr().add(prefix_len),
                            tail_len,
                        );
                    }
                }
                Ok(out)
            },
        )
        .expect("in-memory Vec reply allocation should not fail in tests")
    }

    /// Single-copy fast path for N-API: when the reply fits in a single slot,
    /// copies directly from SHM into the Node-owned `napi_create_buffer` without
    /// an intermediate `Vec<u8>`.
    fn call_sync_napi<F>(
        self: &Arc<Self>,
        env: &Env,
        slot_idx: usize,
        worker_epoch: u32,
        req_ptr: *const u8,
        req_len: usize,
        is_cancelled: F,
    ) -> Result<(i32, Option<napi::sys::napi_value>)>
    where
        F: Fn() -> bool,
    {
        self.call_sync_core(
            slot_idx,
            worker_epoch,
            req_ptr,
            req_len,
            is_cancelled,
            |prefix, tail_ptr, tail_len| {
                let prefix_len = prefix.map_or(0, <[u8]>::len);
                let total = prefix_len + tail_len;
                let mut value = ptr::null_mut();
                let mut dest = ptr::null_mut();
                napi::check_status!(
                    unsafe {
                        napi::sys::napi_create_buffer(env.raw(), total, &mut dest, &mut value)
                    },
                    "failed to allocate RPC response buffer"
                )?;
                if total > 0 {
                    if dest.is_null() {
                        return Err(Error::from_reason("missing allocated RPC buffer"));
                    }
                    let dst = dest.cast::<u8>();
                    if let Some(p) = prefix {
                        if !p.is_empty() {
                            unsafe {
                                ptr::copy_nonoverlapping(p.as_ptr(), dst, p.len());
                            }
                        }
                    }
                    if tail_len > 0 {
                        // SAFETY: state == RESP_READY grants exclusive read of tail_ptr[0..tail_len].
                        unsafe {
                            ptr::copy_nonoverlapping(tail_ptr, dst.add(prefix_len), tail_len);
                        }
                    }
                }
                Ok(value)
            },
        )
    }

    fn wait_for_response_state<F>(
        &self,
        slot_idx: usize,
        req_seq: u32,
        is_cancelled: &F,
    ) -> WaitOutcome
    where
        F: Fn() -> bool,
    {
        let state_word = self.slot_word(slot_idx, SLOT_STATE_OFF);
        let cancel_seq_word = self.slot_word(slot_idx, SLOT_CANCEL_SEQ_OFF);
        let sleeping_word = self.slot_word(slot_idx, SLOT_CLIENT_SLEEPING_OFF);
        let slot = &self.slots[slot_idx];

        // Tier 1: short adaptive spin on >2-core hosts (zero syscalls for fast local
        // Go checker replies; disabled on <=2-core hosts to avoid starving Go).
        let spin_dur = spin_duration();
        if !spin_dur.is_zero() {
            let start = Instant::now();
            let mut spins: u32 = 0;
            loop {
                let st = state_word.load(Ordering::Acquire);
                if st == STATE_RESP_READY {
                    return WaitOutcome::Ready;
                }
                if st == STATE_RESP_CHUNK {
                    return WaitOutcome::Chunk;
                }
                if st == STATE_IDLE || st == STATE_ABANDONED {
                    return WaitOutcome::Unavailable;
                }
                spins = spins.wrapping_add(1);
                if spins & 255 == 0 {
                    if is_cancelled() || cancel_seq_word.load(Ordering::Acquire) == req_seq {
                        return WaitOutcome::Cancelled;
                    }
                    if self.is_unavailable() {
                        return WaitOutcome::Unavailable;
                    }
                    if start.elapsed() >= spin_dur {
                        break;
                    }
                }
                hint::spin_loop();
            }
        } else {
            let st = state_word.load(Ordering::Acquire);
            if st == STATE_RESP_READY {
                return WaitOutcome::Ready;
            }
            if st == STATE_RESP_CHUNK {
                return WaitOutcome::Chunk;
            }
            if st == STATE_IDLE || st == STATE_ABANDONED {
                return WaitOutcome::Unavailable;
            }
        }

        // Tier 2: Dekker SeqCst sleep on in-process Condvar (woken by doorbell reader thread
        // or main-thread `rpc_abort_slot`, with 10ms safety poll for raw SAB cancelFlag updates).
        let mut guard = slot.wait_mu.lock().unwrap_or_else(|e| e.into_inner());
        loop {
            slot.waiting.store(true, Ordering::SeqCst);
            sleeping_word.store(1, Ordering::SeqCst);

            let st = state_word.load(Ordering::SeqCst);
            if st == STATE_RESP_READY {
                sleeping_word.store(0, Ordering::Release);
                slot.waiting.store(false, Ordering::Release);
                return WaitOutcome::Ready;
            }
            if st == STATE_RESP_CHUNK {
                sleeping_word.store(0, Ordering::Release);
                slot.waiting.store(false, Ordering::Release);
                return WaitOutcome::Chunk;
            }
            if st == STATE_IDLE || st == STATE_ABANDONED {
                sleeping_word.store(0, Ordering::Release);
                slot.waiting.store(false, Ordering::Release);
                return WaitOutcome::Unavailable;
            }
            if is_cancelled() || cancel_seq_word.load(Ordering::SeqCst) == req_seq {
                sleeping_word.store(0, Ordering::Release);
                slot.waiting.store(false, Ordering::Release);
                return WaitOutcome::Cancelled;
            }
            if self.is_unavailable() {
                sleeping_word.store(0, Ordering::Release);
                slot.waiting.store(false, Ordering::Release);
                return WaitOutcome::Unavailable;
            }

            let (next_guard, _) = slot
                .wait_cv
                .wait_timeout(guard, CONDVAR_POLL_TIMEOUT)
                .unwrap_or_else(|e| e.into_inner());
            guard = next_guard;
            sleeping_word.store(0, Ordering::Release);
            slot.waiting.store(false, Ordering::Release);
        }
    }
}

enum WaitOutcome {
    Ready,
    Chunk,
    Cancelled,
    Unavailable,
}

/// Owned by one Node IPC session on the main thread.
#[napi]
pub struct RpcChannel {
    backing: Option<platform::Backing>,
    inner: Option<Arc<RpcChannelInner>>,
    channel_id: Option<u32>,
}

#[napi]
impl RpcChannel {
    #[napi(constructor)]
    pub fn new() -> Result<Self> {
        let backing =
            platform::Backing::new().map_err(|error| Error::from_reason(error.to_string()))?;
        Ok(Self {
            backing: Some(backing),
            inner: None,
            channel_id: None,
        })
    }

    #[napi]
    pub fn fd(&self) -> Option<i32> {
        self.backing.as_ref().and_then(platform::Backing::fd)
    }

    #[napi]
    pub fn doorbell_fd(&self) -> Option<i32> {
        self.backing
            .as_ref()
            .and_then(platform::Backing::doorbell_fd)
    }

    /// Close the parent's copy of the child's doorbell fd right after `spawn`.
    #[napi]
    pub fn after_spawn(&mut self) {
        if let Some(backing) = self.backing.as_mut() {
            backing.after_spawn();
        }
    }

    /// Lazily size, seal, map, and register the RPC shared memory when Go
    /// advertises RPC capability and Node constructs its first plugin host.
    #[napi]
    pub fn configure(
        &mut self,
        config: RpcConfiguration,
        inherited_fd: Option<i32>,
        inherited_doorbell_fd: Option<i32>,
        peer_pid: Option<u32>,
    ) -> Result<RpcMapping> {
        let backing = self.backing.take().ok_or_else(invalid)?;
        let layout = Layout::new(config)?;
        let mapping = backing
            .configure(layout, peer_pid)
            .map_err(|error| Error::from_reason(error.to_string()))?;
        let descriptor = mapping.descriptor(inherited_fd, inherited_doorbell_fd);
        let inner = Arc::new(RpcChannelInner::new(mapping));

        let mut reg = channels().lock().map_err(|_| invalid())?;
        reg.try_reserve(1).map_err(|_| invalid())?;
        let id = NEXT_CHANNEL_ID
            .fetch_update(Ordering::Relaxed, Ordering::Relaxed, |id| id.checked_add(1))
            .map_err(|_| invalid())?;
        reg.insert(id, Arc::clone(&inner));

        self.inner = Some(inner);
        self.channel_id = Some(id);
        Ok(descriptor)
    }

    #[napi]
    pub fn channel_id(&self) -> Option<u32> {
        self.channel_id
    }

    #[napi]
    pub fn close(&mut self) {
        if let Some(id) = self.channel_id.take() {
            let mut reg = channels().lock().unwrap_or_else(|e| e.into_inner());
            reg.remove(&id);
        }
        if let Some(inner) = self.inner.take() {
            inner.close();
        }
        self.backing = None;
    }
}

impl Drop for RpcChannel {
    fn drop(&mut self) {
        self.close();
    }
}

fn lookup_channel(channel_id: u32) -> Option<Arc<RpcChannelInner>> {
    channels()
        .lock()
        .ok()?
        .get(&channel_id)
        .cloned()
}

#[napi]
pub fn rpc_acquire_slot(channel_id: u32) -> Option<RpcSlotLease> {
    let channel = lookup_channel(channel_id)?;
    channel.acquire_slot()
}

#[napi]
pub fn rpc_release_slot(channel_id: u32, slot: u32) {
    if let Some(channel) = lookup_channel(channel_id) {
        channel.release_slot(slot as usize);
    }
}

#[napi]
pub fn rpc_abort_slot(channel_id: u32, slot: u32) {
    if let Some(channel) = lookup_channel(channel_id) {
        channel.abort_slot(slot as usize);
    }
}

/// Extract raw pointer + element count from a TypedArray without constructing
/// Rust slices over potentially shared (`SharedArrayBuffer`) backing memory.
unsafe fn typed_array_raw(
    env: napi::sys::napi_env,
    value: napi::sys::napi_value,
    expected_type: napi::sys::napi_typedarray_type,
    require_array_buffer: bool,
) -> Result<(*mut u8, usize)> {
    let mut ty = expected_type;
    let mut len: usize = 0;
    let mut data = ptr::null_mut();
    let mut ab = ptr::null_mut();
    let mut byte_offset: usize = 0;
    napi::check_status!(
        napi::sys::napi_get_typedarray_info(
            env,
            value,
            &mut ty,
            &mut len,
            &mut data,
            &mut ab,
            &mut byte_offset,
        ),
        "expected TypedArray argument"
    )?;
    if ty != expected_type {
        return Err(Error::from_reason("unexpected TypedArray element type"));
    }
    if require_array_buffer {
        let mut is_ab = false;
        napi::check_status!(
            napi::sys::napi_is_arraybuffer(env, ab, &mut is_ab),
            "failed to inspect TypedArray backing buffer"
        )?;
        if !is_ab {
            return Err(Error::from_reason(
                "RPC request buffer must be backed by an ArrayBuffer",
            ));
        }
    }
    Ok((data.cast::<u8>(), len))
}

/// Synchronous RPC call invoked directly from worker threads.
/// Writes one of `JS_STATUS_*` into `status_out[0]` and returns `Buffer` for
/// `OK`/`ERROR` payloads or `null` otherwise.
#[allow(clippy::too_many_arguments)]
#[napi(catch_unwind, ts_return_type = "Buffer | null")]
pub fn rpc_call_sync<'a>(
    env: &'a Env,
    channel_id: u32,
    slot: u32,
    worker_epoch: u32,
    req_buf: Unknown<'a>,
    req_len: u32,
    cancel_flag: Unknown<'a>,
    status_out: Unknown<'a>,
) -> NativeResult<Unknown<'a>> {
    NativeResult((|| {
        let (req_ptr, req_cap) = unsafe {
            typed_array_raw(
                env.raw(),
                req_buf.raw(),
                napi::sys::TypedarrayType::uint8_array,
                true,
            )?
        };
        let (cancel_ptr, cancel_len) = unsafe {
            typed_array_raw(
                env.raw(),
                cancel_flag.raw(),
                napi::sys::TypedarrayType::int32_array,
                false,
            )?
        };
        let (status_ptr, status_len) = unsafe {
            typed_array_raw(
                env.raw(),
                status_out.raw(),
                napi::sys::TypedarrayType::int32_array,
                false,
            )?
        };

        let req_len = req_len as usize;
        let align = std::mem::align_of::<AtomicI32>();
        if req_len > req_cap
            || (req_len > 0 && req_ptr.is_null())
            || cancel_len < 1
            || cancel_ptr.is_null()
            || !(cancel_ptr as usize).is_multiple_of(align)
            || status_len < 1
            || status_ptr.is_null()
            || !(status_ptr as usize).is_multiple_of(align)
        {
            return Err(Error::from_reason("invalid RPC TypedArray arguments"));
        }

        // SAFETY: V8 roots all TypedArray arguments for the duration of this
        // synchronous N-API call; access shared memory strictly via AtomicI32.
        let cancel_atomic = unsafe { AtomicI32::from_ptr(cancel_ptr.cast()) };
        let status_atomic = unsafe { AtomicI32::from_ptr(status_ptr.cast()) };

        let is_cancelled = || -> bool { cancel_atomic.load(Ordering::SeqCst) != 0 };

        let Some(channel) = lookup_channel(channel_id) else {
            let status = if is_cancelled() {
                JS_STATUS_CANCELLED
            } else {
                JS_STATUS_UNAVAILABLE
            };
            status_atomic.store(status, Ordering::Release);
            let mut null_val = ptr::null_mut();
            napi::check_status!(
                unsafe { napi::sys::napi_get_null(env.raw(), &mut null_val) },
                "failed to create null"
            )?;
            return Ok(unsafe { Unknown::from_raw_unchecked(env.raw(), null_val) });
        };

        let (status, maybe_buf) = channel.call_sync_napi(
            env,
            slot as usize,
            worker_epoch,
            req_ptr,
            req_len,
            is_cancelled,
        )?;
        status_atomic.store(status, Ordering::Release);

        let raw_val = match maybe_buf {
            Some(v) => v,
            None => {
                let mut null_val = ptr::null_mut();
                napi::check_status!(
                    unsafe { napi::sys::napi_get_null(env.raw(), &mut null_val) },
                    "failed to create null"
                )?;
                null_val
            }
        };
        Ok(unsafe { Unknown::from_raw_unchecked(env.raw(), raw_val) })
    })())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_config() -> RpcConfiguration {
        RpcConfiguration {
            version: 1,
            slot_count: 4,
            slot_size: 4096,
            header_size: 4096,
            slot_header_stride: 128,
        }
    }

    #[test]
    fn layout_validation_rejects_bad_configs() {
        let valid = test_config();
        assert!(Layout::new(valid).is_ok());
        for bad in [
            RpcConfiguration { version: 2, ..valid },
            RpcConfiguration { slot_count: 0, ..valid },
            RpcConfiguration { slot_count: 65, ..valid },
            RpcConfiguration { slot_size: 1024, ..valid },
            RpcConfiguration { slot_header_stride: 32, ..valid },
            RpcConfiguration { slot_header_stride: 66, ..valid },
            RpcConfiguration { header_size: 256, ..valid },
        ] {
            assert!(Layout::new(bad).is_err());
        }
    }

    #[cfg(unix)]
    #[test]
    fn roundtrip_single_and_chunked_and_abandon_and_eof() {
        use std::os::fd::AsRawFd;

        let (backing, child_sock) = platform::Mapping::test_take_child_doorbell_backing().unwrap();
        let layout = Layout::new(test_config()).unwrap();
        let mapping = backing.configure(layout, None).unwrap();
        let inner = Arc::new(RpcChannelInner::new(mapping));

        // Verify magic stamped before peer maps.
        assert_eq!(
            inner.global_word(GLOBAL_MAGIC_OFFSET).load(Ordering::Acquire),
            RPC_MAGIC
        );

        let lease = inner.acquire_slot().unwrap();
        assert_eq!(lease.slot, 0);
        assert_eq!(lease.epoch, 1);

        // Stale / wrong epoch is rejected immediately as UNAVAILABLE without touching SHM.
        let msg = b"hello shm rpc";
        let (st, reply) = inner.call_sync_raw(0, 999, msg.as_ptr(), msg.len(), || false);
        assert_eq!(st, JS_STATUS_UNAVAILABLE);
        assert!(reply.is_none());

        // Spawn simulated Go server sharing the same Arc<RpcChannelInner>.
        let srv = Arc::clone(&inner);
        let child_fd = child_sock.as_raw_fd();
        let server_thread = thread::spawn(move || {
            let mut served = 0u32;
            while served < 4 {
                let st = srv.slot_word(0, SLOT_STATE_OFF).load(Ordering::SeqCst);
                if st == STATE_REQ_READY
                    && srv
                        .slot_word(0, SLOT_STATE_OFF)
                        .compare_exchange(
                            STATE_REQ_READY,
                            STATE_IN_FLIGHT,
                            Ordering::SeqCst,
                            Ordering::Acquire,
                        )
                        .is_ok()
                {
                    let seq = srv.slot_word(0, SLOT_REQ_SEQ_OFF).load(Ordering::Acquire);
                    let len = srv.slot_word(0, SLOT_REQ_LEN_OFF).load(Ordering::Acquire) as usize;
                    let mut req = vec![0u8; len];
                    unsafe {
                        ptr::copy_nonoverlapping(srv.slot_payload_ptr(0), req.as_mut_ptr(), len);
                    }
                    served += 1;
                    if served == 1 {
                        // Case 1: Single-slot echo.
                        unsafe {
                            ptr::copy_nonoverlapping(req.as_ptr(), srv.slot_payload_ptr(0), len);
                        }
                        srv.slot_word(0, SLOT_RESP_SEQ_OFF).store(seq, Ordering::Relaxed);
                        srv.slot_word(0, SLOT_RESP_LEN_OFF).store(len as u32, Ordering::Relaxed);
                        srv.slot_word(0, SLOT_RESP_STATUS_OFF)
                            .store(WIRE_STATUS_OK, Ordering::Relaxed);
                        srv.slot_word(0, SLOT_STATE_OFF)
                            .store(STATE_RESP_READY, Ordering::SeqCst);
                        if srv.slot_word(0, SLOT_CLIENT_SLEEPING_OFF).load(Ordering::SeqCst) != 0 {
                            let b: u8 = 1;
                            unsafe {
                                libc::send(child_fd, (&raw const b).cast(), 1, libc::MSG_DONTWAIT);
                            }
                        }
                    } else if served == 2 {
                        // Case 2: Multi-chunk reply (10000 bytes across 3 chunks of <=4096).
                        let big: Vec<u8> = (0..10000).map(|i| (i % 251) as u8).collect();
                        let mut off = 0usize;
                        while off < big.len() {
                            let n = (big.len() - off).min(layout.slot_size);
                            unsafe {
                                ptr::copy_nonoverlapping(
                                    big.as_ptr().add(off),
                                    srv.slot_payload_ptr(0),
                                    n,
                                );
                            }
                            off += n;
                            srv.slot_word(0, SLOT_RESP_SEQ_OFF).store(seq, Ordering::Relaxed);
                            srv.slot_word(0, SLOT_RESP_LEN_OFF).store(n as u32, Ordering::Relaxed);
                            srv.slot_word(0, SLOT_TOTAL_RESP_LEN_OFF)
                                .store(big.len() as u32, Ordering::Relaxed);
                            srv.slot_word(0, SLOT_RESP_STATUS_OFF)
                                .store(WIRE_STATUS_OK, Ordering::Relaxed);
                            let next_state = if off < big.len() {
                                STATE_RESP_CHUNK
                            } else {
                                STATE_RESP_READY
                            };
                            srv.slot_word(0, SLOT_STATE_OFF)
                                .store(next_state, Ordering::SeqCst);
                            if srv.slot_word(0, SLOT_CLIENT_SLEEPING_OFF).load(Ordering::SeqCst) != 0 {
                                let b: u8 = 1;
                                unsafe {
                                    libc::send(child_fd, (&raw const b).cast(), 1, libc::MSG_DONTWAIT);
                                }
                            }
                            if next_state == STATE_RESP_CHUNK {
                                while srv.slot_word(0, SLOT_STATE_OFF).load(Ordering::SeqCst)
                                    == STATE_RESP_CHUNK
                                {
                                    thread::yield_now();
                                }
                            }
                        }
                    } else if served == 3 {
                        // Case 3: Wait for client to cancel via abort_slot ALONE (without is_cancelled).
                        while srv.slot_word(0, SLOT_STATE_OFF).load(Ordering::SeqCst)
                            == STATE_IN_FLIGHT
                        {
                            thread::sleep(Duration::from_millis(1));
                        }
                        assert_eq!(
                            srv.slot_word(0, SLOT_STATE_OFF).load(Ordering::SeqCst),
                            STATE_ABANDONED
                        );
                        srv.slot_word(0, SLOT_STATE_OFF)
                            .store(STATE_IDLE, Ordering::Release);
                    } else {
                        // Case 4: Verify SLOT_CANCEL_SEQ was cleared on subsequent request and echo.
                        assert_eq!(
                            srv.slot_word(0, SLOT_CANCEL_SEQ_OFF).load(Ordering::Acquire),
                            0
                        );
                        unsafe {
                            ptr::copy_nonoverlapping(req.as_ptr(), srv.slot_payload_ptr(0), len);
                        }
                        srv.slot_word(0, SLOT_RESP_SEQ_OFF).store(seq, Ordering::Relaxed);
                        srv.slot_word(0, SLOT_RESP_LEN_OFF).store(len as u32, Ordering::Relaxed);
                        srv.slot_word(0, SLOT_RESP_STATUS_OFF)
                            .store(WIRE_STATUS_OK, Ordering::Relaxed);
                        srv.slot_word(0, SLOT_STATE_OFF)
                            .store(STATE_RESP_READY, Ordering::SeqCst);
                        if srv.slot_word(0, SLOT_CLIENT_SLEEPING_OFF).load(Ordering::SeqCst) != 0 {
                            let b: u8 = 1;
                            unsafe {
                                libc::send(child_fd, (&raw const b).cast(), 1, libc::MSG_DONTWAIT);
                            }
                        }
                    }
                } else {
                    thread::yield_now();
                }
            }
            // Drop child_sock to trigger EOF on peer death!
            drop(child_sock);
        });

        // 1. Single-slot echo
        let (st, reply) =
            inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
        assert_eq!(st, JS_STATUS_OK);
        assert_eq!(reply.unwrap(), msg);

        // 2. Multi-chunk 10000-byte reply
        let (st, reply) =
            inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
        assert_eq!(st, JS_STATUS_OK);
        let expected: Vec<u8> = (0..10000).map(|i| (i % 251) as u8).collect();
        assert_eq!(reply.unwrap(), expected);

        // 3. Mid-flight abort_slot ALONE (with is_cancelled == false!) must cancel & wake immediately.
        // Also verify concurrent call on same slot returns JS_STATUS_FALLBACK via in_use guard.
        let inner_abort = Arc::clone(&inner);
        let lease_epoch = lease.epoch;
        let aborter = thread::spawn(move || {
            thread::sleep(Duration::from_millis(5));
            let probe = b"x";
            let (dup_st, dup_reply) =
                inner_abort.call_sync_raw(0, lease_epoch, probe.as_ptr(), probe.len(), || false);
            assert_eq!(dup_st, JS_STATUS_FALLBACK);
            assert!(dup_reply.is_none());
            inner_abort.abort_slot(0);
        });
        let (st, reply) = inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
        aborter.join().unwrap();
        assert_eq!(st, JS_STATUS_CANCELLED);
        assert!(reply.is_none());

        // Wait until simulated server resets slot 0 from ABANDONED -> IDLE.
        let reset_deadline = Instant::now() + Duration::from_secs(2);
        while inner.slot_word(0, SLOT_STATE_OFF).load(Ordering::Acquire) != STATE_IDLE
            && Instant::now() < reset_deadline
        {
            thread::sleep(Duration::from_millis(1));
        }

        // 4. Next call after cancel clears SLOT_CANCEL_SEQ and succeeds.
        let (st, reply) =
            inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
        assert_eq!(st, JS_STATUS_OK);
        assert_eq!(reply.unwrap(), msg);

        // Wait for server to exit and close child_sock -> triggers EOF -> peer_closed!
        server_thread.join().unwrap();
        let deadline = Instant::now() + Duration::from_secs(2);
        while !inner.is_unavailable() && Instant::now() < deadline {
            thread::sleep(Duration::from_millis(2));
        }
        assert!(inner.is_unavailable());

        // 5. Subsequent call reports UNAVAILABLE without blocking.
        let (st, reply) =
            inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
        assert_eq!(st, JS_STATUS_UNAVAILABLE);
        assert!(reply.is_none());

        inner.close();
    }

    #[cfg(unix)]
    #[test]
    fn release_invalidates_epoch_and_rejects_truncated_or_empty_chunks() {
        let (backing, _child_sock) = platform::Mapping::test_take_child_doorbell_backing().unwrap();
        let layout = Layout::new(test_config()).unwrap();
        let mapping = backing.configure(layout, None).unwrap();
        let inner = Arc::new(RpcChannelInner::new(mapping));

        let lease1 = inner.acquire_slot().unwrap();
        assert_eq!(lease1.slot, 0);
        assert_eq!(lease1.epoch, 1);
        inner.release_slot(0);
        // Released slot rejects old epoch.
        let msg = b"hi";
        let (st, _) = inner.call_sync_raw(0, lease1.epoch, msg.as_ptr(), msg.len(), || false);
        assert_eq!(st, JS_STATUS_UNAVAILABLE);

        // Re-acquiring yields a strictly new epoch; old epoch remains rejected.
        let lease2 = inner.acquire_slot().unwrap();
        assert_eq!(lease2.slot, 0);
        assert_eq!(lease2.epoch, 2);
        let (st, _) = inner.call_sync_raw(0, lease1.epoch, msg.as_ptr(), msg.len(), || false);
        assert_eq!(st, JS_STATUS_UNAVAILABLE);

        // Simulated buggy peer sending a 0-byte RESP_CHUNK must fall back instead of looping forever.
        let srv = Arc::clone(&inner);
        let server_thread = thread::spawn(move || {
            while srv.slot_word(0, SLOT_STATE_OFF).load(Ordering::SeqCst) != STATE_REQ_READY {
                thread::yield_now();
            }
            let seq = srv.slot_word(0, SLOT_REQ_SEQ_OFF).load(Ordering::Acquire);
            srv.slot_word(0, SLOT_RESP_SEQ_OFF).store(seq, Ordering::Relaxed);
            srv.slot_word(0, SLOT_RESP_LEN_OFF).store(0, Ordering::Relaxed);
            srv.slot_word(0, SLOT_TOTAL_RESP_LEN_OFF).store(100, Ordering::Relaxed);
            srv.slot_word(0, SLOT_RESP_STATUS_OFF).store(WIRE_STATUS_OK, Ordering::Relaxed);
            srv.slot_word(0, SLOT_STATE_OFF).store(STATE_RESP_CHUNK, Ordering::SeqCst);
        });
        let (st, reply) =
            inner.call_sync_raw(0, lease2.epoch, msg.as_ptr(), msg.len(), || false);
        server_thread.join().unwrap();
        assert_eq!(st, JS_STATUS_FALLBACK);
        assert!(reply.is_none());

        inner.close();
    }
}
