/// <reference types="@rsbuild/core/types" />

interface ImportMetaEnv {
  readonly SSG_MD: boolean;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
