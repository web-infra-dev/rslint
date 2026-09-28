//! Deterministic native-read barrier for the isolated worker integration test.
//! This entire module is absent from ordinary builds, including its N-API exports.

use crate::memory_transport::{MemoryArena, MemoryBatch, MemoryConfiguration, SharedBytes};
use napi::{Error, Result};
use napi_derive::napi;
use std::sync::{Condvar, Mutex, OnceLock};
use std::time::Duration;

#[derive(Default)]
struct State {
    lease: u32,
    entered: bool,
    resume: bool,
    program: Option<String>,
}

static BARRIER: OnceLock<(Mutex<State>, Condvar)> = OnceLock::new();

fn barrier() -> &'static (Mutex<State>, Condvar) {
    BARRIER.get_or_init(|| (Mutex::new(State::default()), Condvar::new()))
}

#[napi(object, object_from_js = false)]
pub struct WorkerTerminationFixture {
    pub arena: MemoryArena,
    pub source: SharedBytes,
}

/// A single fixture per isolated process. The writer uses the native platform
/// mapping implementation, so the JavaScript test never handles an fd or handle.
#[napi]
pub fn create_worker_termination_fixture(text: String) -> Result<WorkerTerminationFixture> {
    let config = MemoryConfiguration {
        version: 1,
        slot_count: 1,
        slot_size: 64 * 1024,
        header_size: 64 * 1024,
        publication_stride: 4,
    };
    if text.is_empty() || text.len() > config.slot_size as usize {
        return Err(Error::from_reason(
            "invalid worker termination fixture source",
        ));
    }
    let mut state = barrier().0.lock().unwrap();
    if state.lease != 0 {
        return Err(Error::from_reason(
            "worker termination fixture already exists",
        ));
    }
    let mut arena = MemoryArena::new()?;
    arena.configure(config)?;
    arena.publish_for_test(0, 1, text.as_bytes())?;
    let length = text.len() as u32;
    let lease = arena.register(vec![MemoryBatch {
        slot: 0,
        generation: 1,
        length,
    }])?;
    state.lease = lease;
    Ok(WorkerTerminationFixture {
        arena,
        source: SharedBytes {
            lease,
            offset: 0,
            length,
        },
    })
}

/// Change only the publication word, leaving the pinned payload untouched. A
/// rejected second registration must prove retirement, not a stale generation.
#[napi]
pub fn republish_worker_fixture(arena: &MemoryArena) -> Result<()> {
    arena.publish_for_test(0, 2, &[])
}

pub(crate) fn before_parse(lease: u32) {
    let (mutex, changed) = barrier();
    let mut state = mutex.lock().unwrap();
    if state.lease != lease {
        return;
    }
    state.entered = true;
    changed.notify_all();
    while !state.resume {
        state = changed.wait(state).unwrap();
    }
}

pub(crate) fn after_parse(lease: u32, program: &str) {
    let mut state = barrier().0.lock().unwrap();
    if state.lease == lease {
        state.program = Some(program.to_owned());
    }
}

/// Wait for an observed event, never an assumed parser duration. The timeout
/// only fails a worker that never reached the barrier (for example load errors).
#[napi]
pub fn wait_for_worker_parse() -> Result<()> {
    let (mutex, changed) = barrier();
    let state = mutex.lock().unwrap();
    let (state, _) = changed
        .wait_timeout_while(state, Duration::from_secs(60), |state| !state.entered)
        .unwrap();
    if !state.entered {
        return Err(Error::from_reason(
            "worker did not enter the native parse barrier",
        ));
    }
    Ok(())
}

#[napi]
pub fn resume_worker_parse() {
    let (mutex, changed) = barrier();
    mutex.lock().unwrap().resume = true;
    changed.notify_all();
}

#[napi]
pub fn worker_parsed_program() -> Option<String> {
    barrier().0.lock().unwrap().program.clone()
}
