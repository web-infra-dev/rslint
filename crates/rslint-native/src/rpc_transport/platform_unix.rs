//! Unix (Linux/macOS) shared-memory and anonymous socketpair doorbell backend.
// cspell:words munmap syscall memfd CLOEXEC CREAT RDWR fcntl SETFD ftruncate NOSIGPIPE DONTWAIT NOSIGNAL POLLIN POLLHUP POLLERR POLLNVAL PSHMNAMLEN nonblock GETFL SETFL NONBLOCK setsockopt socklen rsrpc socketpair pfds pollfd revents wakeup

use super::DoorbellEvent;
use crate::rpc_transport::protocol::{Layout, RpcMapping, GLOBAL_MAGIC_OFFSET, RPC_MAGIC};
use std::io;
use std::os::fd::{AsRawFd, FromRawFd, OwnedFd};
use std::ptr;
use std::sync::atomic::{AtomicU32, Ordering};

fn set_cloexec(fd: libc::c_int) -> io::Result<()> {
    if unsafe { libc::fcntl(fd, libc::F_SETFD, libc::FD_CLOEXEC) } < 0 {
        return Err(io::Error::last_os_error());
    }
    Ok(())
}

fn set_nonblock(fd: libc::c_int) -> io::Result<()> {
    let flags = unsafe { libc::fcntl(fd, libc::F_GETFL) };
    if flags < 0 {
        return Err(io::Error::last_os_error());
    }
    if unsafe { libc::fcntl(fd, libc::F_SETFL, flags | libc::O_NONBLOCK) } < 0 {
        return Err(io::Error::last_os_error());
    }
    Ok(())
}

#[cfg(target_os = "macos")]
fn set_nosigpipe(fd: libc::c_int) -> io::Result<()> {
    let one: libc::c_int = 1;
    if unsafe {
        libc::setsockopt(
            fd,
            libc::SOL_SOCKET,
            libc::SO_NOSIGPIPE,
            (&raw const one).cast(),
            std::mem::size_of::<libc::c_int>() as libc::socklen_t,
        )
    } < 0
    {
        return Err(io::Error::last_os_error());
    }
    Ok(())
}

#[cfg(not(target_os = "macos"))]
fn set_nosigpipe(_fd: libc::c_int) -> io::Result<()> {
    Ok(())
}

fn create_shm_fd() -> io::Result<OwnedFd> {
    #[cfg(target_os = "linux")]
    let raw = unsafe {
        libc::syscall(
            libc::SYS_memfd_create,
            c"rslint-rpc".as_ptr(),
            libc::MFD_CLOEXEC | libc::MFD_ALLOW_SEALING,
        ) as i32
    };
    #[cfg(not(target_os = "linux"))]
    let raw = {
        use std::ffi::CString;
        static NEXT: AtomicU32 = AtomicU32::new(0);
        // Keep well below macOS PSHMNAMLEN (31 bytes) and distinct from MemoryArena.
        let name = CString::new(format!(
            "/rsrpc-{}-{}",
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
    set_cloexec(fd.as_raw_fd())?;
    Ok(fd)
}

fn create_socketpair() -> io::Result<(OwnedFd, OwnedFd)> {
    let mut fds = [-1 as libc::c_int; 2];
    if unsafe { libc::socketpair(libc::AF_UNIX, libc::SOCK_STREAM, 0, fds.as_mut_ptr()) } != 0 {
        return Err(io::Error::last_os_error());
    }
    let parent = unsafe { OwnedFd::from_raw_fd(fds[0]) };
    let child = unsafe { OwnedFd::from_raw_fd(fds[1]) };
    set_cloexec(parent.as_raw_fd())?;
    set_cloexec(child.as_raw_fd())?;
    set_nosigpipe(parent.as_raw_fd())?;
    set_nosigpipe(child.as_raw_fd())?;
    Ok((parent, child))
}

fn create_wake_pipe() -> io::Result<(OwnedFd, OwnedFd)> {
    let mut fds = [-1 as libc::c_int; 2];
    if unsafe { libc::pipe(fds.as_mut_ptr()) } != 0 {
        return Err(io::Error::last_os_error());
    }
    let read_end = unsafe { OwnedFd::from_raw_fd(fds[0]) };
    let write_end = unsafe { OwnedFd::from_raw_fd(fds[1]) };
    set_cloexec(read_end.as_raw_fd())?;
    set_cloexec(write_end.as_raw_fd())?;
    set_nonblock(read_end.as_raw_fd())?;
    set_nonblock(write_end.as_raw_fd())?;
    Ok((read_end, write_end))
}

/// Created before spawn so the child inherits the anonymous SHM fd and one
/// end of the anonymous Unix socketpair. Zero bytes are allocated until
/// `configure()` runs lazily when a plugin host is actually created.
pub struct Backing {
    shm_fd: OwnedFd,
    parent_sock: OwnedFd,
    child_sock: Option<OwnedFd>,
}

impl Backing {
    pub fn new() -> io::Result<Self> {
        let shm_fd = create_shm_fd()?;
        let (parent_sock, child_sock) = create_socketpair()?;
        Ok(Self {
            shm_fd,
            parent_sock,
            child_sock: Some(child_sock),
        })
    }

    pub fn fd(&self) -> Option<i32> {
        Some(self.shm_fd.as_raw_fd())
    }

    pub fn doorbell_fd(&self) -> Option<i32> {
        self.child_sock.as_ref().map(AsRawFd::as_raw_fd)
    }

    /// Drop the parent's copy of the child's socket end immediately after
    /// spawning so peer death produces EOF on `parent_sock`.
    pub fn after_spawn(&mut self) {
        self.child_sock = None;
    }

    pub fn configure(self, layout: Layout, _peer_pid: Option<u32>) -> io::Result<Mapping> {
        let raw = self.shm_fd.as_raw_fd();
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
        let (wake_read, wake_write) = create_wake_pipe()?;
        let base = unsafe {
            libc::mmap(
                ptr::null_mut(),
                layout.capacity,
                libc::PROT_READ | libc::PROT_WRITE,
                libc::MAP_SHARED,
                raw,
                0,
            )
        };
        if base == libc::MAP_FAILED {
            return Err(io::Error::last_os_error());
        }
        let base: *mut u8 = base.cast();
        // ftruncate from 0 zero-initializes all pages; stamp magic before Go maps.
        unsafe {
            AtomicU32::from_ptr(base.add(GLOBAL_MAGIC_OFFSET).cast())
                .store(RPC_MAGIC, Ordering::Release);
        }
        Ok(Mapping {
            base,
            layout,
            shm_fd: self.shm_fd,
            parent_sock: self.parent_sock,
            wake_read,
            wake_write,
        })
    }
}

pub struct Mapping {
    base: *mut u8,
    layout: Layout,
    shm_fd: OwnedFd,
    parent_sock: OwnedFd,
    wake_read: OwnedFd,
    wake_write: OwnedFd,
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
        inherited_fd: Option<i32>,
        inherited_doorbell_fd: Option<i32>,
    ) -> RpcMapping {
        RpcMapping {
            version: self.layout.version,
            fd: inherited_fd.or(Some(self.shm_fd.as_raw_fd())),
            doorbell_fd: inherited_doorbell_fd,
            handle: None,
            node_to_go_event: None,
            go_to_node_event: None,
            process_id: None,
        }
    }

    pub fn ring_node_to_go(&self) {
        let byte: u8 = 1;
        #[cfg(target_os = "linux")]
        let flags = libc::MSG_DONTWAIT | libc::MSG_NOSIGNAL;
        #[cfg(not(target_os = "linux"))]
        let flags = libc::MSG_DONTWAIT;
        unsafe {
            let _ = libc::send(
                self.parent_sock.as_raw_fd(),
                (&raw const byte).cast(),
                1,
                flags,
            );
        }
    }

    pub fn signal_close(&self) {
        let byte: u8 = 1;
        unsafe {
            let _ = libc::write(self.wake_write.as_raw_fd(), (&raw const byte).cast(), 1);
        }
    }

    pub fn wait_go_to_node(&self) -> DoorbellEvent {
        let mut pfds = [
            libc::pollfd {
                fd: self.parent_sock.as_raw_fd(),
                events: libc::POLLIN,
                revents: 0,
            },
            libc::pollfd {
                fd: self.wake_read.as_raw_fd(),
                events: libc::POLLIN,
                revents: 0,
            },
        ];
        loop {
            let rc = unsafe { libc::poll(pfds.as_mut_ptr(), 2, -1) };
            if rc < 0 {
                let err = io::Error::last_os_error();
                if err.kind() == io::ErrorKind::Interrupted {
                    continue;
                }
                return DoorbellEvent::Closed;
            }
            if (pfds[1].revents & (libc::POLLIN | libc::POLLHUP | libc::POLLERR | libc::POLLNVAL))
                != 0
            {
                return DoorbellEvent::Closed;
            }
            if (pfds[0].revents & (libc::POLLIN | libc::POLLHUP | libc::POLLERR | libc::POLLNVAL))
                != 0
            {
                let mut buf = [0u8; 64];
                let n = unsafe {
                    libc::recv(
                        self.parent_sock.as_raw_fd(),
                        buf.as_mut_ptr().cast(),
                        buf.len(),
                        libc::MSG_DONTWAIT,
                    )
                };
                if n > 0 {
                    // Drain any coalesced wakeup bytes without blocking.
                    loop {
                        let more = unsafe {
                            libc::recv(
                                self.parent_sock.as_raw_fd(),
                                buf.as_mut_ptr().cast(),
                                buf.len(),
                                libc::MSG_DONTWAIT,
                            )
                        };
                        if more <= 0 {
                            break;
                        }
                    }
                    return DoorbellEvent::Wake;
                }
                if n == 0 {
                    return DoorbellEvent::Closed;
                }
                let err = io::Error::last_os_error();
                if err.kind() == io::ErrorKind::WouldBlock
                    || err.kind() == io::ErrorKind::Interrupted
                {
                    if (pfds[0].revents & (libc::POLLHUP | libc::POLLERR | libc::POLLNVAL)) != 0 {
                        return DoorbellEvent::Closed;
                    }
                    continue;
                }
                return DoorbellEvent::Closed;
            }
        }
    }

    #[cfg(test)]
    pub fn test_take_child_doorbell_backing() -> io::Result<(Backing, OwnedFd)> {
        let mut backing = Backing::new()?;
        let child = backing.child_sock.take().unwrap();
        Ok((backing, child))
    }
}

impl Drop for Mapping {
    fn drop(&mut self) {
        unsafe {
            libc::munmap(self.base.cast(), self.layout.capacity);
        }
    }
}
