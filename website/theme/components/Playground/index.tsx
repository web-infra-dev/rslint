import React, { use, useEffect, useRef, useState } from 'react';
import * as Rslint from '@rslint/wasm';
import { Editor, EditorRef } from './Editor';
const wasmURL = new URL('@rslint/wasm/rslint.wasm', import.meta.url).href;
let rslintService: Rslint.RSLintService | null = null;
async function ensureService() {
  if (!rslintService) {
    rslintService = await Rslint.initialize({
      wasmURL: wasmURL,
    });
  }
  return rslintService;
}

const Playground: React.FC = () => {
  const editorRef = useRef<EditorRef | null>(null);
  async function runLint() {
    const service = await ensureService();
    const code = editorRef.current?.getValue() || 'let a:any; a.b = 10;';
    const result = await service.lint({
      fileContents: {
        '/index.ts': code,
      },
      config: 'rslint.json',
      ruleOptions: {
        '@typescript-eslint/no-unsafe-member-access': 'error',
      },
    });
    editorRef.current?.attachDiag(result.diagnostics);
  }
  useEffect(() => {
    runLint();
  }, []);
  return (
    <div className="playground-container">
      <div className="playground-main">
        <div className="editor-panel">
          <Editor ref={editorRef} onChange={() => runLint()} />
        </div>
      </div>
    </div>
  );
};

export default Playground;
