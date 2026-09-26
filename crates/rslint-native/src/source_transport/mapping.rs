//! OS handles are local implementation details. The wire carries a descriptor,
//! never a process address or a source-file path. All supported npm targets use
//! the same fixed-slot protocol above this module.
// cspell:words munmap syscall memfd CLOEXEC CREAT RDWR fcntl SETFD ftruncate READWRITE EFAULT

use super::SourceMapping;
use std::io;

unsafe fn publication(control: *mut u8, slot: usize) -> u32 {
    use std::sync::atomic::{AtomicU32, Ordering};
    // AtomicU32::from_ptr requires readable and writable memory, even for loads.
    // The separate control view satisfies that contract; source slices always
    // use the read-only data view.
    AtomicU32::from_ptr(control.add(slot * 4).cast()).load(Ordering::Acquire)
}

#[cfg(unix)]
mod platform {
    use super::*;
    use std::os::fd::{AsRawFd, FromRawFd, OwnedFd};
    use std::ptr;

    pub struct Mapping {
        data: *mut u8,
        control: *mut u8,
        length: usize,
        fd: OwnedFd,
    }

    impl Mapping {
        pub fn published(&self, slot: usize) -> u32 {
            // The caller bounds slot to SLOT_COUNT. This control word is never
            // included in an immutable source slice.
            unsafe { super::publication(self.control, slot) }
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
            // Request only the header. The OS may round this view up to a host
            // page (for example, 16 KiB on macOS), so never borrow source bytes
            // through it.
            let control = unsafe {
                libc::mmap(
                    ptr::null_mut(),
                    super::super::HEADER_SIZE,
                    libc::PROT_READ | libc::PROT_WRITE,
                    libc::MAP_SHARED,
                    raw,
                    0,
                )
            };
            if control == libc::MAP_FAILED {
                let error = io::Error::last_os_error();
                unsafe { libc::munmap(data, length) };
                return Err(error);
            }
            Ok(Self {
                data: data.cast(),
                control: control.cast(),
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
                libc::munmap(self.control.cast(), super::super::HEADER_SIZE);
                libc::munmap(self.data.cast(), self.length);
            }
        }
    }

    #[cfg(test)]
    #[test]
    fn control_is_writable_and_data_is_read_only() {
        let mapping = Mapping::new(super::super::CAPACITY).unwrap();
        let zero = std::fs::File::open("/dev/zero").unwrap();
        mapping.publish_for_test(0, 1);
        assert_eq!(mapping.published(0), 1);
        // Kernel writes report inaccessible destinations as EFAULT instead of
        // deliberately crashing the test process with a direct write.
        assert_eq!(
            unsafe { libc::read(zero.as_raw_fd(), mapping.control.cast(), 4) },
            4,
        );
        assert_eq!(mapping.published(0), 0);
        let result = unsafe {
            libc::read(
                zero.as_raw_fd(),
                mapping.data.add(super::super::HEADER_SIZE).cast(),
                1,
            )
        };
        let error = io::Error::last_os_error();
        assert_eq!(result, -1);
        assert_eq!(error.raw_os_error(), Some(libc::EFAULT));
    }
}

#[cfg(windows)]
mod platform {
    use super::*;
    use std::ptr;
    use windows_sys::Win32::{
        Foundation::{CloseHandle, HANDLE, INVALID_HANDLE_VALUE},
        System::Memory::{
            CreateFileMappingW, MapViewOfFile, UnmapViewOfFile, VirtualAlloc, FILE_MAP_READ,
            FILE_MAP_WRITE, MEMORY_MAPPED_VIEW_ADDRESS, MEM_COMMIT, PAGE_READWRITE, SEC_RESERVE,
        },
    };

    pub struct Mapping {
        data: *mut u8,
        control: *mut u8,
        handle: HANDLE,
    }

    impl Mapping {
        pub fn published(&self, slot: usize) -> u32 {
            unsafe { super::publication(self.control, slot) }
        }

        #[cfg(test)]
        pub fn publish_for_test(&self, slot: usize, generation: u32) {
            unsafe {
                let view = MapViewOfFile(self.handle, FILE_MAP_WRITE, 0, 0, super::super::CAPACITY);
                assert!(!view.Value.is_null());
                let start = super::super::HEADER_SIZE + slot * super::super::SLOT_SIZE;
                assert!(!VirtualAlloc(
                    view.Value.cast::<u8>().add(start).cast(),
                    super::super::SLOT_SIZE,
                    MEM_COMMIT,
                    PAGE_READWRITE,
                )
                .is_null());
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
                    PAGE_READWRITE | SEC_RESERVE,
                    0,
                    length as u32,
                    ptr::null(),
                )
            };
            if handle.is_null() {
                return Err(io::Error::last_os_error());
            }
            // Reserve payload pages until the Go writer needs a slot. Plain
            // PAGE_READWRITE defaults to SEC_COMMIT and would charge the full
            // arena even for native-only CLI runs. Only the control page must
            // be readable before the first producer publication.
            let control =
                unsafe { MapViewOfFile(handle, FILE_MAP_WRITE, 0, 0, super::super::HEADER_SIZE) };
            if control.Value.is_null() {
                let error = io::Error::last_os_error();
                unsafe { CloseHandle(handle) };
                return Err(error);
            }
            let committed = unsafe {
                VirtualAlloc(
                    control.Value,
                    super::super::HEADER_SIZE,
                    MEM_COMMIT,
                    PAGE_READWRITE,
                )
            };
            if committed.is_null() {
                let error = io::Error::last_os_error();
                unsafe {
                    UnmapViewOfFile(control);
                    CloseHandle(handle);
                }
                return Err(error);
            }
            let view = unsafe { MapViewOfFile(handle, FILE_MAP_READ, 0, 0, length) };
            if view.Value.is_null() {
                let error = io::Error::last_os_error();
                unsafe {
                    UnmapViewOfFile(control);
                    CloseHandle(handle);
                }
                return Err(error);
            }
            Ok(Self {
                data: view.Value.cast(),
                control: control.Value.cast(),
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
                UnmapViewOfFile(MEMORY_MAPPED_VIEW_ADDRESS {
                    Value: self.control.cast(),
                });
                CloseHandle(self.handle);
            }
        }
    }

    #[cfg(test)]
    #[test]
    fn payload_is_committed_only_when_published() {
        use windows_sys::Win32::System::Memory::{
            VirtualQuery, MEMORY_BASIC_INFORMATION, MEM_RESERVE, PAGE_READONLY,
        };

        let mapping = Mapping::new(super::super::CAPACITY).unwrap();
        let page = |base: *mut u8, offset: usize| {
            let mut info = MEMORY_BASIC_INFORMATION::default();
            assert_ne!(
                unsafe {
                    VirtualQuery(
                        base.add(offset).cast(),
                        &mut info,
                        std::mem::size_of::<MEMORY_BASIC_INFORMATION>(),
                    )
                },
                0,
            );
            info
        };
        assert_eq!(page(mapping.control, 0).State, MEM_COMMIT);
        assert_eq!(page(mapping.control, 0).Protect, PAGE_READWRITE);
        assert_eq!(page(mapping.data, 0).State, MEM_COMMIT);
        assert_eq!(page(mapping.data, 0).Protect, PAGE_READONLY);
        // Query inside the payload, away from any rounding of the control page.
        let first = super::super::HEADER_SIZE + super::super::SLOT_SIZE / 2;
        let second = first + super::super::SLOT_SIZE;
        assert_eq!(page(mapping.data, first).State, MEM_RESERVE);
        assert_eq!(page(mapping.data, second).State, MEM_RESERVE);

        mapping.publish_for_test(0, 1);

        assert_eq!(page(mapping.control, 0).Protect, PAGE_READWRITE);
        assert_eq!(page(mapping.data, first).State, MEM_COMMIT);
        assert_eq!(page(mapping.data, first).Protect, PAGE_READONLY);
        assert_eq!(page(mapping.data, second).State, MEM_RESERVE);
        assert_eq!(mapping.published(0), 1);
    }
}

pub(super) use platform::Mapping;

// SAFETY: only Lease::bytes exposes a slice. Its lifetime pins the mapping and
// prevents the Go producer from receiving permission to reuse that slot.
unsafe impl Send for Mapping {}
unsafe impl Sync for Mapping {}
