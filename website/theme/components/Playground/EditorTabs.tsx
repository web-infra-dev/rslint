import { useState, useRef, useEffect, useImperativeHandle, Ref } from 'react';
import * as monaco from 'monaco-editor';
import { type Diagnostic } from '@rslint/wasm';
import { Button } from '@components/ui/button';
// Monaco-specific styles only (ast-node-highlight)
import './EditorTabs.css';

window.MonacoEnvironment = {
  getWorker: function (_moduleId, label) {
    if (label === 'typescript' || label === 'javascript') {
      return new Worker(
        new URL(
          'monaco-editor/esm/vs/language/typescript/ts.worker',
          import.meta.url,
        ),
      );
    }
    if (label === 'json') {
      return new Worker(
        new URL(
          'monaco-editor/esm/vs/language/json/json.worker',
          import.meta.url,
        ),
      );
    }
    return new Worker(
      new URL('monaco-editor/esm/vs/editor/editor.worker', import.meta.url),
    );
  },
};

export type EditorTabType = 'code' | 'rslint' | 'tsconfig';

export interface EditorTabsRef {
  getValue: () => string | undefined;
  getCodeValue: () => string | undefined;
  getRslintConfig: () => any | null;
  getTsConfig: () => any | null;
  attachDiag: (diags: Diagnostic[]) => void;
  revealRangeByOffset: (start: number, end: number) => void;
  /** Add a temporary highlight decoration (for hover) without changing selection */
  highlightRange: (start: number, end: number) => void;
  /** Clear the temporary highlight decoration */
  clearHighlight: () => void;
}

interface EditorTabsProps {
  ref: Ref<EditorTabsRef>;
  onChange: (value: string) => void;
  onSelectionChange?: (start: number, end: number) => void;
  onConfigChange?: () => void;
}

const DEFAULT_RSLINT_CONFIG = `[
  {
    "languageOptions": {
      "parserOptions": {
        "project": ["./tsconfig.json"]
      }
    },
    "rules": {
      "@typescript-eslint/no-unsafe-member-access": "error"
    },
    "plugins": ["@typescript-eslint"]
  }
]`;

const DEFAULT_TSCONFIG = `{
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "strict": true,
    "strictNullChecks": true
  }
}`;

function parseJsonc(content: string): any | null {
  try {
    // Remove single-line comments
    let cleaned = content.replace(/\/\/.*$/gm, '');
    // Remove multi-line comments
    cleaned = cleaned.replace(/\/\*[\s\S]*?\*\//g, '');
    // Remove trailing commas before } or ]
    cleaned = cleaned.replace(/,(\s*[}\]])/g, '$1');
    return JSON.parse(cleaned);
  } catch {
    return null;
  }
}

export const EditorTabs = ({
  ref,
  onChange,
  onSelectionChange,
  onConfigChange,
}: EditorTabsProps) => {
  const [activeTab, setActiveTab] = useState<EditorTabType>('code');

  const codeEditorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(
    null,
  );
  const rslintEditorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(
    null,
  );
  const tsconfigEditorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(
    null,
  );

  const codeContainerRef = useRef<HTMLDivElement>(null);
  const rslintContainerRef = useRef<HTMLDivElement>(null);
  const tsconfigContainerRef = useRef<HTMLDivElement>(null);

  const urlUpdateTimer = useRef<number | null>(null);
  const astHighlightDecorationIds = useRef<string[]>([]);
  const hoverHighlightDecorationIds = useRef<string[]>([]);
  const isEditingRef = useRef<boolean>(false);
  const editingTimer = useRef<number | null>(null);

  // Store last valid parsed configs
  const lastValidRslintConfig = useRef<any>(null);
  const lastValidTsConfig = useRef<any>(null);

  function getInitialCode(): string {
    if (typeof window === 'undefined') {
      return ['let a: any;', 'a.b = 10;'].join('\n');
    }
    const { search, hash } = window.location;
    const searchParams = new URLSearchParams(search);
    const fromSearch = searchParams.get('code');
    if (fromSearch != null) return fromSearch;
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
        if (url.hash && url.hash.includes('code=')) {
          url.hash = '';
        }
        window.history.replaceState(null, '', url.toString());
      } catch {
        // ignore URL update errors
      }
    }, 300);
  }

  // Initialize last valid configs
  useEffect(() => {
    lastValidRslintConfig.current = parseJsonc(DEFAULT_RSLINT_CONFIG);
    lastValidTsConfig.current = parseJsonc(DEFAULT_TSCONFIG);
  }, []);

  useImperativeHandle(ref, () => ({
    getValue: () => codeEditorRef.current?.getValue(),
    getCodeValue: () => codeEditorRef.current?.getValue(),
    getRslintConfig: () => {
      const content = rslintEditorRef.current?.getValue() || '';
      const parsed = parseJsonc(content);
      if (parsed !== null) {
        lastValidRslintConfig.current = parsed;
      }
      return lastValidRslintConfig.current;
    },
    getTsConfig: () => {
      const content = tsconfigEditorRef.current?.getValue() || '';
      const parsed = parseJsonc(content);
      if (parsed !== null) {
        lastValidTsConfig.current = parsed;
      }
      return lastValidTsConfig.current;
    },
    attachDiag: (diags: Diagnostic[]) => {
      const model = codeEditorRef.current?.getModel();
      if (model) {
        const markers = diags.map(diag => {
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
        monaco.editor.setModelMarkers(model, 'rslint', markers);
      }
    },
    revealRangeByOffset: (start: number, end: number) => {
      const editor = codeEditorRef.current;
      const model = editor?.getModel();
      if (!editor || !model) return;

      // Switch to code tab if not active
      if (activeTab !== 'code') {
        setActiveTab('code');
      }

      const startPos = model.getPositionAt(Math.max(0, start));
      const endPos = model.getPositionAt(Math.max(start, end));
      const range = new monaco.Range(
        startPos.lineNumber,
        startPos.column,
        endPos.lineNumber,
        endPos.column,
      );

      editor.setSelection(range);
      editor.revealRangeInCenter(range, 0);

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
    highlightRange: (start: number, end: number) => {
      const editor = codeEditorRef.current;
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

      // Clear previous hover highlight
      if (hoverHighlightDecorationIds.current.length) {
        hoverHighlightDecorationIds.current = editor.deltaDecorations(
          hoverHighlightDecorationIds.current,
          [],
        );
      }

      // Add new hover highlight decoration (different style from selection highlight)
      hoverHighlightDecorationIds.current = editor.deltaDecorations(
        [],
        [
          {
            range,
            options: {
              className: 'ast-node-hover-highlight',
              stickiness:
                monaco.editor.TrackedRangeStickiness
                  .NeverGrowsWhenTypingAtEdges,
            },
          },
        ],
      );
    },
    clearHighlight: () => {
      const editor = codeEditorRef.current;
      if (!editor) return;

      if (hoverHighlightDecorationIds.current.length) {
        hoverHighlightDecorationIds.current = editor.deltaDecorations(
          hoverHighlightDecorationIds.current,
          [],
        );
      }
    },
  }));

  // Create code editor
  useEffect(() => {
    if (!codeContainerRef.current) return;

    const editor = monaco.editor.create(codeContainerRef.current, {
      value: getInitialCode(),
      language: 'typescript',
      automaticLayout: true,
      scrollBeyondLastLine: false,
    });
    codeEditorRef.current = editor;

    // Trigger initial onChange
    const initialVal = editor.getValue() || '';
    onChange(initialVal);
    scheduleSerializeToUrl(initialVal);

    editor.onDidChangeModelContent(() => {
      const val = editor.getValue() || '';
      onChange(val);
      scheduleSerializeToUrl(val);

      // Mark as editing to prevent AST tree selection during typing
      isEditingRef.current = true;
      if (editingTimer.current) {
        window.clearTimeout(editingTimer.current);
      }
      // Clear editing flag after 1.5 seconds of no typing
      editingTimer.current = window.setTimeout(() => {
        isEditingRef.current = false;
      }, 1500);
    });

    const selDisposable = editor.onDidChangeCursorSelection(() => {
      // Skip selection change during editing to prevent AST tree flickering
      if (isEditingRef.current) return;

      const model = editor.getModel();
      if (!model) return;
      const sel = editor.getSelection();
      if (!sel) return;

      // Clear AST node highlight when user changes selection in editor
      if (astHighlightDecorationIds.current.length) {
        astHighlightDecorationIds.current = editor.deltaDecorations(
          astHighlightDecorationIds.current,
          [],
        );
      }

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
      if (editingTimer.current) {
        window.clearTimeout(editingTimer.current);
      }
    };
  }, []);

  // Configure Monaco JSON to allow comments and trailing commas (JSONC)
  useEffect(() => {
    // Use type assertion to access jsonDefaults which may be typed as deprecated
    const jsonLang = monaco.languages.json as any;
    if (jsonLang.jsonDefaults?.setDiagnosticsOptions) {
      jsonLang.jsonDefaults.setDiagnosticsOptions({
        validate: true,
        allowComments: true,
        trailingCommas: 'ignore',
        schemaValidation: 'ignore',
      });
    }
  }, []);

  // Create rslint.json editor
  useEffect(() => {
    if (!rslintContainerRef.current) return;

    const editor = monaco.editor.create(rslintContainerRef.current, {
      value: DEFAULT_RSLINT_CONFIG,
      language: 'json',
      automaticLayout: true,
      scrollBeyondLastLine: false,
    });
    rslintEditorRef.current = editor;

    editor.onDidChangeModelContent(() => {
      onConfigChange?.();
    });

    return () => {
      editor.dispose();
    };
  }, []);

  // Create tsconfig.json editor
  useEffect(() => {
    if (!tsconfigContainerRef.current) return;

    const editor = monaco.editor.create(tsconfigContainerRef.current, {
      value: DEFAULT_TSCONFIG,
      language: 'json',
      automaticLayout: true,
      scrollBeyondLastLine: false,
    });
    tsconfigEditorRef.current = editor;

    editor.onDidChangeModelContent(() => {
      onConfigChange?.();
    });

    return () => {
      editor.dispose();
    };
  }, []);

  const tabs: { key: EditorTabType; label: string }[] = [
    { key: 'code', label: 'Code' },
    { key: 'rslint', label: 'rslint.json' },
    { key: 'tsconfig', label: 'tsconfig' },
  ];

  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex items-center gap-2 bg-gray-50 p-2 flex-shrink-0">
        {tabs.map(tab => (
          <Button
            key={tab.key}
            type="button"
            variant={activeTab === tab.key ? 'default' : 'outline'}
            size="sm"
            onClick={() => setActiveTab(tab.key)}
            aria-pressed={activeTab === tab.key}
            className="dark:text-accent dark:border-muted/20"
          >
            {tab.label}
          </Button>
        ))}
      </div>
      <div className="flex-1 relative overflow-visible">
        <div
          ref={codeContainerRef}
          className={`absolute inset-0 ${activeTab === 'code' ? 'block' : 'hidden'}`}
        />
        <div
          ref={rslintContainerRef}
          className={`absolute inset-0 ${activeTab === 'rslint' ? 'block' : 'hidden'}`}
        />
        <div
          ref={tsconfigContainerRef}
          className={`absolute inset-0 ${activeTab === 'tsconfig' ? 'block' : 'hidden'}`}
        />
      </div>
    </div>
  );
};
