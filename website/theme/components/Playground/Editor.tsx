import React, { useRef, useEffect, forwardRef, Ref, createRef } from 'react';
import * as monaco from 'monaco-editor';
import { type Diagnostic } from '@rslint/wasm';
import './Editor.css';

window.MonacoEnvironment = {
  getWorker: function (moduleId, label) {
    if (label === 'typescript' || label === 'javascript') {
      return new Worker(
        new URL(
          'monaco-editor/esm/vs/language/typescript/ts.worker',
          import.meta.url,
        ),
      );
    }
    return new Worker(
      new URL('monaco-editor/esm/vs/editor/editor.worker', import.meta.url),
    );
  },
};

export interface EditorRef {
  getValue: () => string | undefined;
  attachDiag: (diags: Diagnostic[]) => void;
}

export const Editor = ({
  ref,
  onChange,
}: {
  ref: Ref<{ getValue: () => string | undefined }>;
  onChange: (value: string) => void;
}) => {
  const divEl = useRef<HTMLDivElement>(null);
  const editorRef =
    useRef<import('monaco-editor').editor.IStandaloneCodeEditor>(null);
  // get value from editor using forwardRef
  React.useImperativeHandle(ref, () => ({
    getValue: () => editorRef.current?.getValue(),
    attachDiag: (diags: Diagnostic[]) => {
      const model = editorRef.current?.getModel();

      if (model) {
        const markers = diags.map(diag => {
          // Convert character offset to line/column position
          const start = diag.range.start;
          const end = diag.range.end;

          return {
            startLineNumber: start.line,
            startColumn: start.column,
            endLineNumber: end.line,
            endColumn: end.column,
            source: 'rslint',
            severity: monaco.MarkerSeverity.Error,
            message: diag.message,
          } as monaco.editor.IMarkerData;
        });
        // attach warning to text
        monaco.editor.setModelMarkers(model, 'rslint', markers);
      }
    },
  }));

  useEffect(() => {
    if (!divEl.current) {
      return;
    }

    const editor = monaco.editor.create(divEl.current, {
      value: ['let a: any;', 'a.b = 10;'].join('\n'),
      language: 'typescript',
      automaticLayout: true,
      scrollBeyondLastLine: false,
    });
    editor.onDidChangeModelContent(() => {
      onChange(editorRef.current?.getValue() || '');
    });
    editorRef.current = editor;

    return () => {
      editor.dispose();
    };
  }, []);
  return <div className="editor-container" ref={divEl}></div>;
};
