//! Windows shared-memory section and event-handle doorbell backend.
// cspell:words READWRITE wakeups socketpair

use super::DoorbellEvent;
use crate::rpc_transport::protocol::{Layout, RpcMapping, GLOBAL_MAGIC_OFFSET, RPC_MAGIC};
use std::io;
use std::ptr;
use std::sync::atomic::{AtomicU32, Ordering};
use windows_sys::Win32::{
    Foundation::{CloseHandle, HANDLE, INVALID_HANDLE_VALUE, WAIT_OBJECT_0},
    System::{
        Memory::{
            CreateFileMappingW, MapViewOfFile, UnmapViewOfFile, FILE_MAP_WRITE,
            MEMORY_MAPPED_VIEW_ADDRESS, PAGE_READWRITE,
        },
        Threading::{
            CreateEventW, OpenProcess, SetEvent, WaitForMultipleObjects, INFINITE,
            PROCESS_SYNCHRONIZE,
        },
    },
};

/// Windows transfers duplicated handles dynamically during `transportRpcOpen`,
/// so no OS resources are allocated before `configure()`.
pub struct Backing;

impl Backing {
    pub fn new() -> io::Result<Self> {
        Ok(Self)
    }

    pub fn fd(&self) -> Option<i32> {
        None
    }

    pub fn doorbell_fd(&self) -> Option<i32> {
        None
    }

    pub fn after_spawn(&mut self) {}

    pub fn configure(self, layout: Layout, peer_pid: Option<u32>) -> io::Result<Mapping> {
        let capacity = layout.capacity as u64;
        // Plain PAGE_READWRITE (implicit SEC_COMMIT) ensures both Node and Go
        // can read/write every slot without VirtualAlloc races.
        let handle = unsafe {
            CreateFileMappingW(
                INVALID_HANDLE_VALUE,
                ptr::null(),
                PAGE_READWRITE,
                (capacity >> 32) as u32,
                capacity as u32,
                ptr::null(),
            )
        };
        if handle.is_null() {
            return Err(io::Error::last_os_error());
        }
        let view = unsafe { MapViewOfFile(handle, FILE_MAP_WRITE, 0, 0, layout.capacity) };
        if view.Value.is_null() {
            let error = io::Error::last_os_error();
            unsafe { CloseHandle(handle) };
            return Err(error);
        }
        let node_to_go = unsafe { CreateEventW(ptr::null(), 0, 0, ptr::null()) };
        if node_to_go.is_null() {
            let error = io::Error::last_os_error();
            unsafe {
                UnmapViewOfFile(view);
                CloseHandle(handle);
            }
            return Err(error);
        }
        let go_to_node = unsafe { CreateEventW(ptr::null(), 0, 0, ptr::null()) };
        if go_to_node.is_null() {
            let error = io::Error::last_os_error();
            unsafe {
                CloseHandle(node_to_go);
                UnmapViewOfFile(view);
                CloseHandle(handle);
            }
            return Err(error);
        }
        let close_event = unsafe { CreateEventW(ptr::null(), 1, 0, ptr::null()) };
        if close_event.is_null() {
            let error = io::Error::last_os_error();
            unsafe {
                CloseHandle(go_to_node);
                CloseHandle(node_to_go);
                UnmapViewOfFile(view);
                CloseHandle(handle);
            }
            return Err(error);
        }
        // Best-effort peer process handle for immediate crash wakeups on Windows
        // (mirrors Unix socketpair EOF). If restricted, close_event still wakes on IpcClient close.
        let peer_process = peer_pid
            .filter(|&pid| pid != 0)
            .map(|pid| unsafe { OpenProcess(PROCESS_SYNCHRONIZE, 0, pid) })
            .filter(|h| !h.is_null());
        let base: *mut u8 = view.Value.cast();
        unsafe {
            AtomicU32::from_ptr(base.add(GLOBAL_MAGIC_OFFSET).cast())
                .store(RPC_MAGIC, Ordering::Release);
        }
        Ok(Mapping {
            base,
            layout,
            handle,
            node_to_go,
            go_to_node,
            close_event,
            peer_process,
        })
    }
}

pub struct Mapping {
    base: *mut u8,
    layout: Layout,
    handle: HANDLE,
    node_to_go: HANDLE,
    go_to_node: HANDLE,
    close_event: HANDLE,
    peer_process: Option<HANDLE>,
}

impl Mapping {
    pub fn base(&self) -> *mut u8 {
        self.base
    }

    pub fn layout(&self) -> Layout {
        self.layout
    }

    pub fn descriptor(
        &self,
        _inherited_fd: Option<i32>,
        _inherited_doorbell_fd: Option<i32>,
    ) -> RpcMapping {
        RpcMapping {
            version: self.layout.version,
            fd: None,
            doorbell_fd: None,
            handle: Some((self.handle as usize).to_string()),
            node_to_go_event: Some((self.node_to_go as usize).to_string()),
            go_to_node_event: Some((self.go_to_node as usize).to_string()),
            process_id: Some(std::process::id()),
        }
    }

    pub fn ring_node_to_go(&self) {
        unsafe {
            let _ = SetEvent(self.node_to_go);
        }
    }

    pub fn signal_close(&self) {
        unsafe {
            let _ = SetEvent(self.close_event);
        }
    }

    pub fn wait_go_to_node(&self) -> DoorbellEvent {
        let (handles, count) = match self.peer_process {
            Some(peer) => ([self.go_to_node, self.close_event, peer], 3u32),
            None => ([self.go_to_node, self.close_event, ptr::null_mut()], 2u32),
        };
        let rc = unsafe { WaitForMultipleObjects(count, handles.as_ptr(), 0, INFINITE) };
        if rc == WAIT_OBJECT_0 {
            DoorbellEvent::Wake
        } else {
            DoorbellEvent::Closed
        }
    }
}

impl Drop for Mapping {
    fn drop(&mut self) {
        unsafe {
            UnmapViewOfFile(MEMORY_MAPPED_VIEW_ADDRESS {
                Value: self.base.cast(),
            });
            if let Some(peer) = self.peer_process {
                CloseHandle(peer);
            }
            CloseHandle(self.close_event);
            CloseHandle(self.go_to_node);
            CloseHandle(self.node_to_go);
            CloseHandle(self.handle);
        }
    }
}
