//! OS handles are local implementation details. The wire carries a descriptor,
//! never a process address or a source-file path. All supported npm targets use
//! the same fixed-slot protocol above this module.
// cspell:words munmap syscall memfd CLOEXEC CREAT RDWR fcntl SETFD ftruncate READWRITE

use super::SourceMapping;
use std::io;

unsafe fn publication(data: *mut u8, slot: usize) -> u32 {
    use std::sync::atomic::{fence, AtomicU32, Ordering};
    // A small relaxed load is guaranteed to work on read-only mappings on our
    // x64/arm64 targets. Acquire loads do not have that guarantee; pair the
    // relaxed load with an acquire fence instead.
    // https://doc.rust-lang.org/std/sync/atomic/#atomic-accesses-to-read-only-memory
    let generation = AtomicU32::from_ptr(data.add(slot * 4).cast()).load(Ordering::Relaxed);
    fence(Ordering::Acquire);
    generation
}

#[cfg(unix)]
mod platform {
    use super::*;
    use std::os::fd::{AsRawFd, FromRawFd, OwnedFd};
    use std::ptr;

    pub struct Mapping {
        data: *mut u8,
        length: usize,
        fd: OwnedFd,
    }

    impl Mapping {
        pub fn published(&self, slot: usize) -> u32 {
            // The caller bounds slot to SLOT_COUNT. This control word is never
            // included in an immutable source slice.
            unsafe { super::publication(self.data, slot) }
        }

        #[cfg(test)]
        pub fn publish_for_test(&self, slot: usize, generation: u32) {
            unsafe {
                let view = libc::mmap(
                    ptr::null_mut(),
                    self.length,
                    libc::PROT_READ | libc::PROT_WRITE,
                    libc::MAP_SHARED,
                    self.fd.as_raw_fd(),
                    0,
                );
                assert_ne!(view, libc::MAP_FAILED);
                let word = (view as *mut u32).add(slot);
                std::sync::atomic::AtomicU32::from_ptr(word)
                    .store(generation, std::sync::atomic::Ordering::Release);
                libc::munmap(view, self.length);
            }
        }

        pub fn new(length: usize) -> io::Result<Self> {
            #[cfg(target_os = "linux")]
            let raw = unsafe {
                libc::syscall(
                    libc::SYS_memfd_create,
                    c"rslint-source".as_ptr(),
                    libc::MFD_CLOEXEC | libc::MFD_ALLOW_SEALING,
                ) as i32
            };
            #[cfg(not(target_os = "linux"))]
            let raw = {
                use std::ffi::CString;
                use std::sync::atomic::{AtomicU32, Ordering};
                static NEXT: AtomicU32 = AtomicU32::new(0);
                let name = CString::new(format!(
                    "/rslint-{}-{}",
                    std::process::id(),
                    NEXT.fetch_add(1, Ordering::Relaxed)
                ))
                .unwrap();
                let fd = unsafe {
                    libc::shm_open(
                        name.as_ptr(),
                        libc::O_CREAT | libc::O_EXCL | libc::O_RDWR,
                        0o600,
                    )
                };
                if fd >= 0 && unsafe { libc::shm_unlink(name.as_ptr()) } != 0 {
                    let error = io::Error::last_os_error();
                    unsafe {
                        libc::close(fd);
                    }
                    return Err(error);
                }
                fd
            };
            if raw < 0 {
                return Err(io::Error::last_os_error());
            }
            let fd = unsafe { OwnedFd::from_raw_fd(raw) };
            if unsafe { libc::fcntl(raw, libc::F_SETFD, libc::FD_CLOEXEC) } < 0
                || unsafe { libc::ftruncate(raw, length as libc::off_t) } != 0
            {
                return Err(io::Error::last_os_error());
            }
            #[cfg(target_os = "linux")]
            if unsafe {
                libc::fcntl(
                    raw,
                    libc::F_ADD_SEALS,
                    libc::F_SEAL_GROW | libc::F_SEAL_SHRINK | libc::F_SEAL_SEAL,
                )
            } < 0
            {
                return Err(io::Error::last_os_error());
            }
            let data = unsafe {
                libc::mmap(
                    ptr::null_mut(),
                    length,
                    libc::PROT_READ,
                    libc::MAP_SHARED,
                    raw,
                    0,
                )
            };
            if data == libc::MAP_FAILED {
                return Err(io::Error::last_os_error());
            }
            Ok(Self {
                data: data.cast(),
                length,
                fd,
            })
        }

        pub fn descriptor(&self) -> SourceMapping {
            SourceMapping {
                version: 1,
                fd: Some(self.fd.as_raw_fd()),
                handle: None,
                process_id: None,
            }
        }

        pub unsafe fn bytes(&self, offset: usize, length: usize) -> &[u8] {
            std::slice::from_raw_parts(self.data.add(offset), length)
        }
    }

    impl Drop for Mapping {
        fn drop(&mut self) {
            unsafe {
                libc::munmap(self.data.cast(), self.length);
            }
        }
    }
}

#[cfg(windows)]
mod platform {
    use super::*;
    use std::ptr;
    use windows_sys::Win32::{
        Foundation::{CloseHandle, HANDLE, INVALID_HANDLE_VALUE},
        System::Memory::{
            CreateFileMappingW, MapViewOfFile, UnmapViewOfFile, FILE_MAP_READ,
            MEMORY_MAPPED_VIEW_ADDRESS, PAGE_READWRITE,
        },
    };

    pub struct Mapping {
        data: *mut u8,
        handle: HANDLE,
    }

    impl Mapping {
        pub fn published(&self, slot: usize) -> u32 {
            unsafe { super::publication(self.data, slot) }
        }

        #[cfg(test)]
        pub fn publish_for_test(&self, slot: usize, generation: u32) {
            unsafe {
                let view = MapViewOfFile(
                    self.handle,
                    windows_sys::Win32::System::Memory::FILE_MAP_WRITE,
                    0,
                    0,
                    super::super::CAPACITY,
                );
                assert!(!view.Value.is_null());
                let word = (view.Value as *mut u32).add(slot);
                std::sync::atomic::AtomicU32::from_ptr(word)
                    .store(generation, std::sync::atomic::Ordering::Release);
                UnmapViewOfFile(view);
            }
        }

        pub fn new(length: usize) -> io::Result<Self> {
            let handle = unsafe {
                CreateFileMappingW(
                    INVALID_HANDLE_VALUE,
                    ptr::null(),
                    PAGE_READWRITE,
                    0,
                    length as u32,
                    ptr::null(),
                )
            };
            if handle.is_null() {
                return Err(io::Error::last_os_error());
            }
            let view = unsafe { MapViewOfFile(handle, FILE_MAP_READ, 0, 0, length) };
            if view.Value.is_null() {
                let error = io::Error::last_os_error();
                unsafe {
                    CloseHandle(handle);
                }
                return Err(error);
            }
            Ok(Self {
                data: view.Value.cast(),
                handle,
            })
        }

        pub fn descriptor(&self) -> SourceMapping {
            SourceMapping {
                version: 1,
                fd: None,
                handle: Some((self.handle as usize).to_string()),
                process_id: Some(std::process::id()),
            }
        }

        pub unsafe fn bytes(&self, offset: usize, length: usize) -> &[u8] {
            std::slice::from_raw_parts(self.data.add(offset), length)
        }
    }

    impl Drop for Mapping {
        fn drop(&mut self) {
            unsafe {
                UnmapViewOfFile(MEMORY_MAPPED_VIEW_ADDRESS {
                    Value: self.data.cast(),
                });
                CloseHandle(self.handle);
            }
        }
    }
}

pub(super) use platform::Mapping;

// SAFETY: only Lease::bytes exposes a slice. Its lifetime pins the mapping and
// prevents the Go producer from receiving permission to reuse that slot.
unsafe impl Send for Mapping {}
unsafe impl Sync for Mapping {}
