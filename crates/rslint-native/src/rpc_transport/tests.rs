//! Unit tests for the bidirectional shared-memory RPC transport.
// cspell:words nonoverlapping DONTWAIT

use super::channel::RpcChannelInner;
use super::platform;
use super::protocol::{
    Layout, RpcConfiguration, GLOBAL_MAGIC_OFFSET, JS_STATUS_CANCELLED, JS_STATUS_FALLBACK,
    JS_STATUS_OK, JS_STATUS_UNAVAILABLE, RPC_MAGIC, SLOT_CANCEL_SEQ_OFF, SLOT_CLIENT_SLEEPING_OFF,
    SLOT_REQ_LEN_OFF, SLOT_REQ_SEQ_OFF, SLOT_RESP_LEN_OFF, SLOT_RESP_SEQ_OFF, SLOT_RESP_STATUS_OFF,
    SLOT_STATE_OFF, SLOT_TOTAL_RESP_LEN_OFF, STATE_ABANDONED, STATE_IDLE, STATE_IN_FLIGHT,
    STATE_REQ_READY, STATE_RESP_CHUNK, STATE_RESP_READY, WIRE_STATUS_OK,
};
use std::ptr;
use std::sync::{atomic::Ordering, Arc};
use std::thread;
use std::time::{Duration, Instant};

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
        RpcConfiguration {
            version: 2,
            ..valid
        },
        RpcConfiguration {
            slot_count: 0,
            ..valid
        },
        RpcConfiguration {
            slot_count: 65,
            ..valid
        },
        RpcConfiguration {
            slot_size: 1024,
            ..valid
        },
        RpcConfiguration {
            slot_header_stride: 32,
            ..valid
        },
        RpcConfiguration {
            slot_header_stride: 66,
            ..valid
        },
        RpcConfiguration {
            header_size: 256,
            ..valid
        },
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
        inner
            .global_word(GLOBAL_MAGIC_OFFSET)
            .load(Ordering::Acquire),
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
                    srv.slot_word(0, SLOT_RESP_SEQ_OFF)
                        .store(seq, Ordering::Relaxed);
                    srv.slot_word(0, SLOT_RESP_LEN_OFF)
                        .store(len as u32, Ordering::Relaxed);
                    srv.slot_word(0, SLOT_RESP_STATUS_OFF)
                        .store(WIRE_STATUS_OK, Ordering::Relaxed);
                    srv.slot_word(0, SLOT_STATE_OFF)
                        .store(STATE_RESP_READY, Ordering::SeqCst);
                    if srv
                        .slot_word(0, SLOT_CLIENT_SLEEPING_OFF)
                        .load(Ordering::SeqCst)
                        != 0
                    {
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
                        srv.slot_word(0, SLOT_RESP_SEQ_OFF)
                            .store(seq, Ordering::Relaxed);
                        srv.slot_word(0, SLOT_RESP_LEN_OFF)
                            .store(n as u32, Ordering::Relaxed);
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
                        if srv
                            .slot_word(0, SLOT_CLIENT_SLEEPING_OFF)
                            .load(Ordering::SeqCst)
                            != 0
                        {
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
                    while srv.slot_word(0, SLOT_STATE_OFF).load(Ordering::SeqCst) == STATE_IN_FLIGHT
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
                        srv.slot_word(0, SLOT_CANCEL_SEQ_OFF)
                            .load(Ordering::Acquire),
                        0
                    );
                    unsafe {
                        ptr::copy_nonoverlapping(req.as_ptr(), srv.slot_payload_ptr(0), len);
                    }
                    srv.slot_word(0, SLOT_RESP_SEQ_OFF)
                        .store(seq, Ordering::Relaxed);
                    srv.slot_word(0, SLOT_RESP_LEN_OFF)
                        .store(len as u32, Ordering::Relaxed);
                    srv.slot_word(0, SLOT_RESP_STATUS_OFF)
                        .store(WIRE_STATUS_OK, Ordering::Relaxed);
                    srv.slot_word(0, SLOT_STATE_OFF)
                        .store(STATE_RESP_READY, Ordering::SeqCst);
                    if srv
                        .slot_word(0, SLOT_CLIENT_SLEEPING_OFF)
                        .load(Ordering::SeqCst)
                        != 0
                    {
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
    let (st, reply) = inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
    assert_eq!(st, JS_STATUS_OK);
    assert_eq!(reply.unwrap(), msg);

    // 2. Multi-chunk 10000-byte reply
    let (st, reply) = inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
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
    let (st, reply) = inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
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
    let (st, reply) = inner.call_sync_raw(0, lease.epoch, msg.as_ptr(), msg.len(), || false);
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
        srv.slot_word(0, SLOT_RESP_SEQ_OFF)
            .store(seq, Ordering::Relaxed);
        srv.slot_word(0, SLOT_RESP_LEN_OFF)
            .store(0, Ordering::Relaxed);
        srv.slot_word(0, SLOT_TOTAL_RESP_LEN_OFF)
            .store(100, Ordering::Relaxed);
        srv.slot_word(0, SLOT_RESP_STATUS_OFF)
            .store(WIRE_STATUS_OK, Ordering::Relaxed);
        srv.slot_word(0, SLOT_STATE_OFF)
            .store(STATE_RESP_CHUNK, Ordering::SeqCst);
    });
    let (st, reply) = inner.call_sync_raw(0, lease2.epoch, msg.as_ptr(), msg.len(), || false);
    server_thread.join().unwrap();
    assert_eq!(st, JS_STATUS_FALLBACK);
    assert!(reply.is_none());

    inner.close();
}
