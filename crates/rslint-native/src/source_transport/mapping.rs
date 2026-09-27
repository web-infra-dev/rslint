//! OS handles are local implementation details. The wire carries a descriptor,
//! never a process address or a source-file path. All supported npm targets use
//! the same fixed-slot protocol above this module.
// cspell:words munmap syscall memfd CLOEXEC CREAT RDWR fcntl SETFD ftruncate READWRITE EFAULT nonoverlapping fstat

use super::{Layout, SourceMapping};
use std::io;

unsafe fn publication(control: *mut u8, offset: usize) -> u32 {
    use std::sync::atomic::{AtomicU32, Ordering};
    // AtomicU32::from_ptr requires readable and writable memory, even for loads.
    // The separate control view satisfies that contract; source slices always
    // use the read-only data view.
    AtomicU32::from_ptr(control.add(offset).cast()).load(Ordering::Acquire)
}

#[cfg(unix)]
mod platform {
    use super::*;
    use std::os::fd::{AsRawFd, FromRawFd, OwnedFd};
    use std::ptr;

    /// Created before spawn so the child inherits the same anonymous object.
    /// Its dimensions and memory views do not exist until Go supplies a layout.
    pub struct Backing {
        fd: OwnedFd,
    }

    impl Backing {
        pub fn new() -> io::Result<Self> {
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
            if unsafe { libc::fcntl(raw, libc::F_SETFD, libc::FD_CLOEXEC) } < 0 {
                return Err(io::Error::last_os_error());
            }
            Ok(Self { fd })
        }

        pub fn fd(&self) -> Option<i32> {
            Some(self.fd.as_raw_fd())
        }

        pub fn configure(self, layout: Layout) -> io::Result<Mapping> {
            let raw = self.fd.as_raw_fd();
            let length = libc::off_t::try_from(layout.capacity)
                .map_err(|error| io::Error::new(io::ErrorKind::InvalidInput, error))?;
            if unsafe { libc::ftruncate(raw, length) } != 0 {
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
                    layout.capacity,
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
                    layout.header_size,
                    libc::PROT_READ | libc::PROT_WRITE,
                    libc::MAP_SHARED,
                    raw,
                    0,
                )
            };
            if control == libc::MAP_FAILED {
                let error = io::Error::last_os_error();
                unsafe { libc::munmap(data, layout.capacity) };
                return Err(error);
            }
            Ok(Mapping {
                data: data.cast(),
                control: control.cast(),
                layout,
                fd: self.fd,
            })
        }
    }

    pub struct Mapping {
        data: *mut u8,
        control: *mut u8,
        layout: Layout,
        fd: OwnedFd,
    }

    impl Mapping {
        pub fn layout(&self) -> Layout {
            self.layout
        }

        pub fn fd(&self) -> Option<i32> {
            Some(self.fd.as_raw_fd())
        }

        pub fn published(&self, slot: usize) -> u32 {
            // The caller bounds slot against the configured count. This control
            // word is never included in an immutable source slice.
            unsafe { super::publication(self.control, self.layout.publication_offset(slot)) }
        }

        #[cfg(test)]
        pub fn publish_for_test(&self, slot: usize, generation: u32) {
            self.write_for_test(slot, generation, &[]);
        }

        #[cfg(test)]
        pub fn write_for_test(&self, slot: usize, generation: u32, source: &[u8]) {
            assert!(source.len() <= self.layout.slot_size);
            unsafe {
                let view = libc::mmap(
                    ptr::null_mut(),
                    self.layout.capacity,
                    libc::PROT_READ | libc::PROT_WRITE,
                    libc::MAP_SHARED,
                    self.fd.as_raw_fd(),
                    0,
                );
                assert_ne!(view, libc::MAP_FAILED);
                ptr::copy_nonoverlapping(
                    source.as_ptr(),
                    (view as *mut u8).add(self.layout.header_size + slot * self.layout.slot_size),
                    source.len(),
                );
                let word = (view as *mut u8)
                    .add(slot * self.layout.publication_stride)
                    .cast();
                std::sync::atomic::AtomicU32::from_ptr(word)
                    .store(generation, std::sync::atomic::Ordering::Release);
                libc::munmap(view, self.layout.capacity);
            }
        }

        pub fn descriptor(&self) -> SourceMapping {
            SourceMapping {
                version: self.layout.version,
                fd: self.fd(),
                handle: None,
                process_id: None,
            }
        }

        pub unsafe fn bytes(&self, offset: usize, length: usize) -> &[u8] {
            debug_assert!(
                offset <= self.layout.capacity && length <= self.layout.capacity - offset
            );
            std::slice::from_raw_parts(self.data.add(offset), length)
        }
    }

    impl Drop for Mapping {
        fn drop(&mut self) {
            unsafe {
                libc::munmap(self.control.cast(), self.layout.header_size);
                libc::munmap(self.data.cast(), self.layout.capacity);
            }
        }
    }

    #[cfg(test)]
    #[test]
    fn control_is_writable_and_data_is_read_only() {
        let layout = Layout::new(super::super::test_configuration()).unwrap();
        let backing = Backing::new().unwrap();
        let mut stat: libc::stat = unsafe { std::mem::zeroed() };
        assert_eq!(unsafe { libc::fstat(backing.fd().unwrap(), &mut stat) }, 0);
        assert_eq!(stat.st_size, 0);
        let mapping = backing.configure(layout).unwrap();
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
                mapping.data.add(layout.header_size).cast(),
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

    /// Windows transfers a mapping handle after configuration, so there is no
    /// OS resource to allocate before its size has been selected by Go.
    pub struct Backing;

    impl Backing {
        pub fn new() -> io::Result<Self> {
            Ok(Self)
        }

        pub fn fd(&self) -> Option<i32> {
            None
        }

        pub fn configure(self, layout: Layout) -> io::Result<Mapping> {
            Mapping::new(layout)
        }
    }

    pub struct Mapping {
        data: *mut u8,
        control: *mut u8,
        handle: HANDLE,
        layout: Layout,
    }

    impl Mapping {
        pub fn layout(&self) -> Layout {
            self.layout
        }

        pub fn fd(&self) -> Option<i32> {
            None
        }

        pub fn published(&self, slot: usize) -> u32 {
            unsafe { super::publication(self.control, self.layout.publication_offset(slot)) }
        }

        #[cfg(test)]
        pub fn publish_for_test(&self, slot: usize, generation: u32) {
            self.write_for_test(slot, generation, &[]);
        }

        #[cfg(test)]
        pub fn write_for_test(&self, slot: usize, generation: u32, source: &[u8]) {
            assert!(source.len() <= self.layout.slot_size);
            unsafe {
                let view = MapViewOfFile(self.handle, FILE_MAP_WRITE, 0, 0, self.layout.capacity);
                assert!(!view.Value.is_null());
                let start = self.layout.header_size + slot * self.layout.slot_size;
                assert!(!VirtualAlloc(
                    view.Value.cast::<u8>().add(start).cast(),
                    self.layout.slot_size,
                    MEM_COMMIT,
                    PAGE_READWRITE,
                )
                .is_null());
                ptr::copy_nonoverlapping(
                    source.as_ptr(),
                    view.Value.cast::<u8>().add(start),
                    source.len(),
                );
                let word = (view.Value as *mut u8)
                    .add(slot * self.layout.publication_stride)
                    .cast();
                std::sync::atomic::AtomicU32::from_ptr(word)
                    .store(generation, std::sync::atomic::Ordering::Release);
                UnmapViewOfFile(view);
            }
        }

        fn new(layout: Layout) -> io::Result<Self> {
            let capacity = layout.capacity as u64;
            let handle = unsafe {
                CreateFileMappingW(
                    INVALID_HANDLE_VALUE,
                    ptr::null(),
                    PAGE_READWRITE | SEC_RESERVE,
                    (capacity >> 32) as u32,
                    capacity as u32,
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
                unsafe { MapViewOfFile(handle, FILE_MAP_WRITE, 0, 0, layout.header_size) };
            if control.Value.is_null() {
                let error = io::Error::last_os_error();
                unsafe { CloseHandle(handle) };
                return Err(error);
            }
            let committed = unsafe {
                VirtualAlloc(
                    control.Value,
                    layout.header_size,
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
            let view = unsafe { MapViewOfFile(handle, FILE_MAP_READ, 0, 0, layout.capacity) };
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
                layout,
            })
        }

        pub fn descriptor(&self) -> SourceMapping {
            SourceMapping {
                version: self.layout.version,
                fd: self.fd(),
                handle: Some((self.handle as usize).to_string()),
                process_id: Some(std::process::id()),
            }
        }

        pub unsafe fn bytes(&self, offset: usize, length: usize) -> &[u8] {
            debug_assert!(
                offset <= self.layout.capacity && length <= self.layout.capacity - offset
            );
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

        let layout = Layout::new(super::super::test_configuration()).unwrap();
        let mapping = Backing::new().unwrap().configure(layout).unwrap();
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
        let first = layout.header_size + layout.slot_size / 2;
        let second = first + layout.slot_size;
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

pub(super) use platform::{Backing, Mapping};

// SAFETY: only Lease::bytes exposes a slice. Its lifetime pins the mapping and
// prevents the Go producer from receiving permission to reuse that slot.
unsafe impl Send for Mapping {}
unsafe impl Sync for Mapping {}
