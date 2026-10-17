//! Shared byte transport for rslint's IPC host and plugin workers.
mod memory_transport;
#[cfg(feature = "test-worker-termination")]
mod reader_worker_test;
pub use memory_transport::{read_bytes, read_bytes_into};
#[cfg(feature = "test-worker-termination")]
pub use reader_worker_test::{
    create_worker_termination_fixture, republish_worker_fixture, resume_worker_read,
    wait_for_worker_read, worker_read_bytes,
};

/// A terminated worker may reject N-API conversion with PendingException even
/// though no ordinary JavaScript exception exists. Return directly to Node in
/// that case: trying to manufacture and throw another error can abort a debug
/// addon. Application callers remain responsible for ignoring cancelled results.
pub struct NativeResult<T>(napi::Result<T>);

impl<T: napi::bindgen_prelude::ToNapiValue> napi::bindgen_prelude::ToNapiValue for NativeResult<T> {
    unsafe fn to_napi_value(
        env: napi::sys::napi_env,
        value: Self,
    ) -> napi::Result<napi::sys::napi_value> {
        match value.0.and_then(|value| T::to_napi_value(env, value)) {
            // Node-API callbacks may return NULL with an exception pending.
            // Preserve the original exception/termination; make no N-API calls.
            Err(error) if error.status == napi::Status::PendingException => {
                Ok(std::ptr::null_mut())
            }
            result => result,
        }
    }
}
