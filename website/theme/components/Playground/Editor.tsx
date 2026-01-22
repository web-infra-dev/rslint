import React, { useRef, useEffect, forwardRef, Ref, createRef } from 'react';
import * as monaco from 'monaco-editor';
import { type Diagnostic } from '@rslint/wasm';
import './Editor.css';

window.MonacoEnvironment = {
  getWorker: function (_moduleId, label) {
    if (label === 'json' || label === 'jsonc') {
      return new Worker(
        new URL(
          'monaco-editor/esm/vs/language/json/json.worker',
          import.meta.url,
        ),
      );
    }
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
  revealRangeByOffset: (start: number, end: number) => void;
}

export const Editor = ({
  ref,
  onChange,
  onSelectionChange,
}: {
  ref: Ref<EditorRef>;
  onChange: (value: string) => void;
  onSelectionChange?: (start: number, end: number) => void;
}) => {
  const divEl = useRef<HTMLDivElement>(null);
  const editorRef =
    useRef<import('monaco-editor').editor.IStandaloneCodeEditor>(null);
  const urlUpdateTimer = useRef<number | null>(null);

  function decodeParam(value: string | null): string | null {
    if (!value) return null;
    try {
      return decodeURIComponent(value);
    } catch {
      return value;
    }
  }

  function getInitialCode(): string {
    if (typeof window === 'undefined') {
      return ['let a: any;', 'a.b = 10;'].join('\n');
    }
    const { search, hash } = window.location;
    const searchParams = new URLSearchParams(search);
    // URLSearchParams.get already decodes percent-encoding
    const fromSearch = searchParams.get('code');
    if (fromSearch != null) return fromSearch;
    // Also support hash like #code=...
    if (hash && hash.startsWith('#')) {
      const hashParams = new URLSearchParams(hash.slice(1));
      const fromHash = hashParams.get('code');
      if (fromHash != null) return fromHash;
    }
    return ['let a: any;', 'a.b = 10;'].join('\n');
  }

  function scheduleSerializeToUrl(value: string) {
    if (typeof window === 'undefined') return;
    if (urlUpdateTimer.current) {
      window.clearTimeout(urlUpdateTimer.current);
      urlUpdateTimer.current = null;
    }
    urlUpdateTimer.current = window.setTimeout(() => {
      try {
        const url = new URL(window.location.href);
        url.searchParams.set('code', value);
        // Remove any code=... from hash if present to avoid ambiguity
        if (url.hash && url.hash.includes('code=')) {
          url.hash = '';
        }
        window.history.replaceState(null, '', url.toString());
      } catch {
        // ignore URL update errors
      }
    }, 300);
  }
  // get value from editor using forwardRef
  const astHighlightDecorationIds = useRef<string[]>([]);

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
    revealRangeByOffset: (start: number, end: number) => {
      const editor = editorRef.current;
      const model = editor?.getModel();
      if (!editor || !model) return;
      const startPos = model.getPositionAt(Math.max(0, start));
      const endPos = model.getPositionAt(Math.max(start, end));
      const range = new monaco.Range(
        startPos.lineNumber,
        startPos.column,
        endPos.lineNumber,
        endPos.column,
      );

      editor.setSelection(range);
      editor.revealRangeInCenter(range, 0 /* Immediate */);

      // Apply a transient decoration for highlighting
      if (astHighlightDecorationIds.current.length) {
        astHighlightDecorationIds.current = editor.deltaDecorations(
          astHighlightDecorationIds.current,
          [],
        );
      }
      astHighlightDecorationIds.current = editor.deltaDecorations(
        [],
        [
          {
            range,
            options: {
              inlineClassName: 'ast-node-highlight',
              stickiness:
                monaco.editor.TrackedRangeStickiness
                  .NeverGrowsWhenTypingAtEdges,
            },
          },
        ],
      );
    },
  }));

  useEffect(() => {
    if (!divEl.current) {
      return;
    }

    const editor = monaco.editor.create(divEl.current, {
      value: getInitialCode(),
      language: 'typescript',
      automaticLayout: true,
      scrollBeyondLastLine: false,
    });
    // Ensure ref is set before first onChange so parent can read value
    editorRef.current = editor;
    // Trigger initial onChange + URL sync with initial value
    {
      const initialVal = editor.getValue() || '';
      onChange(initialVal);
      scheduleSerializeToUrl(initialVal);
    }
    editor.onDidChangeModelContent(() => {
      const val = editor.getValue() || '';
      onChange(val);
      scheduleSerializeToUrl(val);
    });
    // Selection change -> report offsets
    const selDisposable = editor.onDidChangeCursorSelection(() => {
      const model = editor.getModel();
      if (!model) return;
      const sel = editor.getSelection();
      if (!sel) return;
      const startOffset = model.getOffsetAt({
        lineNumber: sel.startLineNumber,
        column: sel.startColumn,
      });
      const endOffset = model.getOffsetAt({
        lineNumber: sel.endLineNumber,
        column: sel.endColumn,
      });
      onSelectionChange?.(
        Math.min(startOffset, endOffset),
        Math.max(startOffset, endOffset),
      );
    });

    return () => {
      selDisposable.dispose();
      editor.dispose();
    };
  }, []);
  return <div className="editor-container" ref={divEl}></div>;
};
