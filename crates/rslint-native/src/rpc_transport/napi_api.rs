//! N-API exports and channel registry for the bidirectional shared-memory RPC
//! transport (`RpcChannel`, slot lease functions, and stateless `rpc_call_sync`).
// cspell:words typedarray arraybuffer

use napi::{Env, Error, JsValue, Result, Unknown};
use napi_derive::napi;
use std::collections::HashMap;
use std::ptr;
use std::sync::{
    atomic::{AtomicI32, AtomicU32, Ordering},
    Arc, Mutex, OnceLock,
};

use super::channel::RpcChannelInner;
use super::platform;
use super::protocol::{
    Layout, RpcConfiguration, RpcMapping, RpcSlotLease, JS_STATUS_CANCELLED, JS_STATUS_UNAVAILABLE,
};
use crate::NativeResult;

static CHANNELS: OnceLock<Mutex<HashMap<u32, Arc<RpcChannelInner>>>> = OnceLock::new();
static NEXT_CHANNEL_ID: AtomicU32 = AtomicU32::new(1);

fn channels() -> &'static Mutex<HashMap<u32, Arc<RpcChannelInner>>> {
    CHANNELS.get_or_init(|| Mutex::new(HashMap::new()))
}

fn invalid() -> Error {
    Error::from_reason("invalid or closed RPC channel")
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
    channels().lock().ok()?.get(&channel_id).cloned()
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
