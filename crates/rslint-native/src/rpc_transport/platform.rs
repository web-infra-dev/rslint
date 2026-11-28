//! Platform-specific bidirectional shared memory and anonymous OS doorbell
//! backing for the synchronous RPC transport.

#[cfg(unix)]
#[path = "platform_unix.rs"]
mod imp;

#[cfg(windows)]
#[path = "platform_windows.rs"]
mod imp;

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum DoorbellEvent {
    Wake,
    Closed,
}

pub(super) use imp::{Backing, Mapping};

// SAFETY: Mapping owns the OS mapping and doorbell handles. Control words are
// accessed strictly via AtomicU32::from_ptr at explicit offsets, and each slot's
// payload region is exclusively owned according to the slot state machine.
unsafe impl Send for Mapping {}
unsafe impl Sync for Mapping {}
