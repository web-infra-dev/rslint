use std::collections::HashMap;

use serde::{Deserialize, Deserializer};
use serde_bytes::Bytes;
type TypeId = u32;
#[derive(Debug, Clone, Deserialize)]
#[non_exhaustive]
pub struct ProjectResponse<'base> {
    #[serde(borrow)]
    pub root_files: Vec<&'base str>,
    pub source_files: Vec<&'base Bytes>,
    pub module_list: Vec<&'base str>,
    pub semantic: Semantic,
    pub diagnostics: Vec<Diagnostic>,
    pub source_file_extra: Vec<SourceFileExtra>,
}
#[derive(Debug, Clone, Deserialize)]
pub struct SourceFileExtra {
    pub has_common_js_module_indicator: bool,
    pub has_external_module_indicator: bool,
}
#[derive(Debug, Clone, Deserialize)]
pub struct Location {
    pub start: u32,
    pub end: u32,
}
#[derive(Debug, Clone, Deserialize)]
pub struct Diagnostic {
    pub message: String,
    pub category: u32,
    pub file: u32,
    pub loc: Location,
}
#[derive(Debug, Clone, Deserialize)]
pub struct Semantic {
    #[serde(deserialize_with = "vecmap")]
    pub symtab: Vec<(u32, SymbolData)>,
    #[serde(deserialize_with = "vecmap")]
    pub typetab: Vec<(u32, TypeData)>,
    #[serde(deserialize_with = "vecmap")]
    pub sym2type: Vec<(u32, u32)>,
    #[serde(deserialize_with = "vecmap")]
    pub node2sym: Vec<(NodeReference, u32)>,
    #[serde(deserialize_with = "vecmap")]
    pub node2type: Vec<(NodeReference, u32)>,
    pub type_extra: TypeExtra,
    pub primtypes: PrimTypes,
    // (aliasSymbolId, targetSymbolId)
    #[serde(default, deserialize_with = "vecmap_or_empty")]
    pub alias_symbols: Vec<(u32, u32)>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct NodeReference {
    pub sourcefile_id: u32,
    pub start: u32,
    pub end: u32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct SymbolData {
    #[serde(with = "serde_bytes")]
    pub name: Vec<u8>,
    pub flags: u32,
    pub check_flags: u32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TypeData {
    pub id: u32,
    pub flags: u32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct PrimTypes {
    pub string: u32,
    pub number: u32,
    pub any: u32,
    pub error: u32,
    pub unknown: u32,
    pub never: u32,
    pub undefined: u32,
    pub null: u32,
    pub void: u32,
    pub bool: u32,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TypeExtra {
    pub name: HashMap<TypeId, serde_bytes::ByteBuf>,
    pub func: HashMap<TypeId, FunctionData>,
}
#[derive(Debug, Clone, Deserialize)]
pub struct FunctionData {
    pub signatures: Vec<Signature>,
}
#[derive(Debug, Clone, Deserialize)]
pub struct Signature {
    pub result: TypeId,
}

fn vecmap<'de, K, V, D>(deserializer: D) -> Result<Vec<(K, V)>, D::Error>
where
    D: Deserializer<'de>,
    K: Deserialize<'de>,
    V: Deserialize<'de>,
{
    use serde::de::Visitor;
    use std::marker::PhantomData;

    struct VecMap<K, V>(PhantomData<(K, V)>);

    impl<'de, K, V> Visitor<'de> for VecMap<K, V>
    where
        K: Deserialize<'de>,
        V: Deserialize<'de>,
    {
        type Value = Vec<(K, V)>;

        fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
            write!(formatter, "vec map")
        }

        fn visit_unit<E>(self) -> Result<Self::Value, E>
        where
            E: serde::de::Error,
        {
            Ok(Vec::new())
        }

        fn visit_map<A>(self, mut map: A) -> Result<Self::Value, A::Error>
        where
            A: serde::de::MapAccess<'de>,
        {
            let len = map.size_hint().unwrap_or_default();
            let len = std::cmp::min(len, 4096);
            let mut out = Vec::with_capacity(len);

            while let Some(e) = map.next_entry()? {
                out.push(e);
            }

            Ok(out)
        }
    }

    deserializer.deserialize_map(VecMap(PhantomData))
}

fn vecmap_or_empty<'de, K, V, D>(deserializer: D) -> Result<Vec<(K, V)>, D::Error>
where
    D: Deserializer<'de>,
    K: Deserialize<'de>,
    V: Deserialize<'de>,
{
    use serde::de::Visitor;
    use std::marker::PhantomData;

    struct VecMapOrEmpty<K, V>(PhantomData<(K, V)>);

    impl<'de, K, V> Visitor<'de> for VecMapOrEmpty<K, V>
    where
        K: Deserialize<'de>,
        V: Deserialize<'de>,
    {
        type Value = Vec<(K, V)>;

        fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
            write!(formatter, "vec map or nothing")
        }

        fn visit_unit<E>(self) -> Result<Self::Value, E>
        where
            E: serde::de::Error,
        {
            Ok(Vec::new())
        }

        fn visit_none<E>(self) -> Result<Self::Value, E>
        where
            E: serde::de::Error,
        {
            Ok(Vec::new())
        }

        fn visit_map<A>(self, mut map: A) -> Result<Self::Value, A::Error>
        where
            A: serde::de::MapAccess<'de>,
        {
            let len = map.size_hint().unwrap_or_default();
            let len = std::cmp::min(len, 4096);
            let mut out = Vec::with_capacity(len);

            while let Some(e) = map.next_entry()? {
                out.push(e);
            }

            Ok(out)
        }
    }

    deserializer.deserialize_any(VecMapOrEmpty(PhantomData))
}
