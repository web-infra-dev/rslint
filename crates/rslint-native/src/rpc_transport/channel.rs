//! Core shared-memory RPC channel state machine (`RpcChannelInner`): slot
//! leasing, Dekker sleep/wake handshakes, chunk streaming, and abandonment.
// cspell:words nonoverlapping syscalls

use napi::{Env, Error, Result};
use std::hint;
use std::ptr;
use std::sync::{
    atomic::{AtomicBool, AtomicU32, Ordering},
    Arc, Mutex,
};
use std::thread;
use std::time::Instant;

use super::platform::{self, DoorbellEvent};
use super::protocol::{
    spin_duration, ActiveSeqGuard, BoolResetGuard, RpcSlotLease, SlotHost, CONDVAR_POLL_TIMEOUT,
    GLOBAL_CLIENT_CLOSED_OFFSET, GLOBAL_HEADER_SIZE, GLOBAL_SERVER_CLOSED_OFFSET,
    GLOBAL_SERVER_SLEEPING_OFFSET, JS_STATUS_CANCELLED, JS_STATUS_CLOSED, JS_STATUS_ERROR,
    JS_STATUS_FALLBACK, JS_STATUS_OK, JS_STATUS_UNAVAILABLE, MAX_REPLY_BYTES, SLOT_CANCEL_SEQ_OFF,
    SLOT_CLIENT_SLEEPING_OFF, SLOT_REQ_LEN_OFF, SLOT_REQ_SEQ_OFF, SLOT_RESP_LEN_OFF,
    SLOT_RESP_SEQ_OFF, SLOT_RESP_STATUS_OFF, SLOT_STATE_OFF, SLOT_TOTAL_RESP_LEN_OFF,
    SLOT_WORKER_EPOCH_OFF, STATE_ABANDONED, STATE_CHUNK_ACK, STATE_IDLE, STATE_IN_FLIGHT,
    STATE_REQ_READY, STATE_RESP_CHUNK, STATE_RESP_READY, WIRE_STATUS_CANCELLED, WIRE_STATUS_CLOSED,
    WIRE_STATUS_ERROR, WIRE_STATUS_OK,
};

pub(super) struct RpcChannelInner {
    mapping: platform::Mapping,
    slots: Vec<SlotHost>,
    closed: AtomicBool,
    peer_closed: AtomicBool,
    reader_started: AtomicBool,
    reader_start_mu: Mutex<()>,
}

enum WaitOutcome {
    Ready,
    Chunk,
    Cancelled,
    Unavailable,
}

impl RpcChannelInner {
    pub(super) fn new(mapping: platform::Mapping) -> Self {
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
    pub(super) fn global_word(&self, offset: usize) -> &AtomicU32 {
        debug_assert!(offset + 4 <= GLOBAL_HEADER_SIZE);
        // SAFETY: mapping is valid for the lifetime of Self, offset is 4-byte aligned in header.
        unsafe { AtomicU32::from_ptr(self.mapping.base().add(offset).cast()) }
    }

    #[inline]
    pub(super) fn slot_word(&self, slot: usize, field_offset: usize) -> &AtomicU32 {
        let off = self.mapping.layout().slot_header_offset(slot) + field_offset;
        debug_assert!(off + 4 <= self.mapping.layout().header_size);
        // SAFETY: mapping is valid for the lifetime of Self, offset is 4-byte aligned in slot header.
        unsafe { AtomicU32::from_ptr(self.mapping.base().add(off).cast()) }
    }

    #[inline]
    pub(super) fn slot_payload_ptr(&self, slot: usize) -> *mut u8 {
        let off = self.mapping.layout().slot_payload_offset(slot);
        // SAFETY: offset is within mapped capacity.
        unsafe { self.mapping.base().add(off) }
    }

    pub(super) fn is_unavailable(&self) -> bool {
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

    pub(super) fn close(&self) {
        if self.closed.swap(true, Ordering::SeqCst) {
            return;
        }
        self.global_word(GLOBAL_CLIENT_CLOSED_OFFSET)
            .store(1, Ordering::SeqCst);
        self.mapping.ring_node_to_go();
        self.mapping.signal_close();
        self.wake_all_slots();
    }

    pub(super) fn acquire_slot(&self) -> Option<RpcSlotLease> {
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

    pub(super) fn release_slot(&self, slot_idx: usize) {
        if let Some(slot) = self.slots.get(slot_idx) {
            // Invalidate epoch BEFORE aborting and clearing `assigned` so no stale
            // worker can pass the epoch check while the slot is being recycled.
            slot.epoch.store(0, Ordering::SeqCst);
            self.abort_slot(slot_idx);
            slot.assigned.store(false, Ordering::Release);
        }
    }

    pub(super) fn abort_slot(&self, slot_idx: usize) {
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
    pub(super) fn call_sync_raw<F>(
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
    pub(super) fn call_sync_napi<F>(
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
