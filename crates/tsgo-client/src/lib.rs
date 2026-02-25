pub mod client;
pub mod proto;
pub mod symbolflags;

pub struct Api {
    client: client::Client,
}
impl Api {
    pub fn with_uninitialized_client(
        client: client::UninitializedClient,
    ) -> Result<Api, client::ProtocolError> {
        let client = client.init()?;
        Ok(Api { client })
    }

    pub fn load_project<'buf>(
        mut self,
        buf: &'buf mut Vec<u8>,
    ) -> Result<proto::ProjectResponse<'buf>, client::ProtocolError> {
        self.client.load_project(buf)
    }
}
