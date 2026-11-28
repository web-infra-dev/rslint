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
// cspell:words finalizers rcpc

mod channel;
mod napi_api;
mod platform;
mod protocol;
#[cfg(test)]
mod tests;

pub use napi_api::{rpc_abort_slot, rpc_acquire_slot, rpc_call_sync, rpc_release_slot, RpcChannel};
