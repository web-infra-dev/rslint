use cbor4ii::serde::DecodeError;

use crate::proto::{self, Diagnostic};
use std::convert::Infallible;
use std::ffi::OsStr;
use std::io::Read;
use std::path::PathBuf;
use std::{io, process};

pub struct Client {
    stdout: process::ChildStdout,
    child: process::Child,
}

#[derive(Default)]
pub struct Options {
    pub cwd: Option<PathBuf>,
    pub log_file: Option<String>,
    pub config_file: String,
}

pub struct Builder {
    cmd: process::Command,
    options: Options,
}

pub struct UninitializedClient {
    client: Client,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct MessageType(u8);

#[derive(thiserror::Error, Debug)]
#[error("process error: {0}")]
pub struct ProcessError(#[from] io::Error);

#[derive(thiserror::Error, Debug)]
#[non_exhaustive]
pub enum ProtocolError {
    #[error("decode error: {0}")]
    DecodeError(DecodeError<Infallible>),
    #[error("invalid TypeScript project with diagnostics: {0:?}")]
    InvalidTsProject(Vec<Diagnostic>),
    #[error("remote error message: {0}")]
    Error(String),
}

impl Client {
    pub fn builder(exe: &OsStr, options: Options) -> Builder {
        let mut cmd = process::Command::new(exe);
        cmd.arg("--api");
        Client::with_command(cmd, options)
    }

    pub fn with_command(cmd: process::Command, options: Options) -> Builder {
        Builder { cmd, options }
    }

    pub fn load_project<'buf>(
        &mut self,
        resp: &'buf mut Vec<u8>,
    ) -> Result<proto::ProjectResponse<'buf>, ProtocolError> {
        resp.clear();
        self.stdout
            .read_to_end(resp)
            .map_err(|err| ProtocolError::Error(err.to_string()))?;
        cbor4ii::serde::from_slice(resp).map_err(ProtocolError::DecodeError)
    }
}

impl Builder {
    pub fn cwd(mut self, path: PathBuf) -> Builder {
        self.options.cwd = Some(path);
        self
    }

    pub fn log_file(mut self, filename: String) -> Builder {
        self.options.log_file = Some(filename);
        self
    }

    pub fn build(mut self) -> Result<UninitializedClient, ProcessError> {
        // Add config file argument if provided
        if !self.options.config_file.is_empty() {
            self.cmd.arg("-config").arg(&self.options.config_file);
        }

        // Set the working directory if provided
        if let Some(cwd) = &self.options.cwd {
            self.cmd.current_dir(cwd);
        }

        let mut child = self
            .cmd
            .stdin(process::Stdio::piped())
            .stdout(process::Stdio::piped())
            .stderr(process::Stdio::inherit())
            .spawn()?;
        let stdout = child.stdout.take().unwrap();

        let client = Client { stdout, child };
        Ok(UninitializedClient { client })
    }
}

impl UninitializedClient {
    pub fn init(self) -> Result<Client, ProtocolError> {
        Ok(self.client)
    }
}

impl Client {
    pub fn close(mut self) -> io::Result<()> {
        self.child.kill()?;
        self.child.wait()?;
        Ok(())
    }
}

impl Drop for Client {
    fn drop(&mut self) {
        let _ = self.child.kill();
    }
}
