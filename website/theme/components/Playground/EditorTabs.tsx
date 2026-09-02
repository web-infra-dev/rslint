import {
  useState,
  useRef,
  useEffect,
  useLayoutEffect,
  useImperativeHandle,
  Ref,
  ReactNode,
} from 'react';
import * as monaco from 'monaco-editor';
import { jsonDefaults } from 'monaco-editor/languages/features/json/register';
import {
  javascriptDefaults,
  ModuleKind,
  ModuleResolutionKind,
  ScriptTarget,
} from 'monaco-editor/languages/features/typescript/register';
import { type Diagnostic } from '@rslint/core/service';
import { useDark } from '@rspress/core/runtime';
import { Button } from '@components/ui/button';
import { evaluateConfig, type PlaygroundConfig } from './config';
import {
  DEFAULT_RSLINT_CONFIG,
  readShareState,
  writeShareState,
} from './share-url';
import { installRslintCoreTypes } from './config-types';
// Monaco-specific styles only (ast-node-highlight)
import './EditorTabs.css';

window.MonacoEnvironment = {
  getWorker: function (_moduleId, label) {
    if (label === 'typescript' || label === 'javascript') {
      return new Worker(
        new URL('monaco-editor/language/typescript/ts.worker', import.meta.url),
      );
    }
    if (label === 'json') {
      return new Worker(
        new URL('monaco-editor/language/json/json.worker', import.meta.url),
      );
    }
    return new Worker(
      new URL('monaco-editor/editor/editor.worker', import.meta.url),
    );
  },
};

export type EditorTabType = 'code' | 'rslint' | 'tsconfig';

export interface EditorTabsRef {
  getValue: () => string | undefined;
  getCodeValue: () => string | undefined;
  getRslintConfig: (wasmVersion: string) => Promise<PlaygroundConfig>;
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
  toolbarEnd?: ReactNode;
}

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
  toolbarEnd,
}: EditorTabsProps) => {
  const [activeTab, setActiveTab] = useState<EditorTabType>('code');
  const [initialState] = useState(readShareState);
  const isDark = useDark();
  const editorTheme = isDark ? 'vs-dark' : 'vs';

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
  const onChangeRef = useRef(onChange);
  const onSelectionChangeRef = useRef(onSelectionChange);
  const onConfigChangeRef = useRef(onConfigChange);

  const lastValidTsConfig = useRef<any>(null);

  // Monaco keeps these subscriptions for the editor lifetime, so forward them
  // through refs instead of retaining callbacks from the initial render.
  useLayoutEffect(() => {
    onChangeRef.current = onChange;
    onSelectionChangeRef.current = onSelectionChange;
    onConfigChangeRef.current = onConfigChange;
  }, [onChange, onSelectionChange, onConfigChange]);

  function scheduleSerializeToUrl() {
    if (typeof window === 'undefined') return;
    if (urlUpdateTimer.current) {
      window.clearTimeout(urlUpdateTimer.current);
      urlUpdateTimer.current = null;
    }
    urlUpdateTimer.current = window.setTimeout(() => {
      writeShareState({
        code: codeEditorRef.current?.getValue() ?? initialState.code,
        rslintConfig:
          rslintEditorRef.current?.getValue() ?? initialState.rslintConfig,
        tsconfig:
          tsconfigEditorRef.current?.getValue() ?? initialState.tsconfig,
      });
    }, 300);
  }

  // Initialize the last valid tsconfig.
  useEffect(() => {
    lastValidTsConfig.current = parseJsonc(initialState.tsconfig);
  }, []);

  useImperativeHandle(ref, () => ({
    getValue: () => codeEditorRef.current?.getValue(),
    getCodeValue: () => codeEditorRef.current?.getValue(),
    getRslintConfig: async (wasmVersion) => {
      await installRslintCoreTypes(wasmVersion);
      return evaluateConfig(
        rslintEditorRef.current?.getValue() ?? DEFAULT_RSLINT_CONFIG,
        wasmVersion,
      );
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
        const markers = diags.map((diag) => {
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
      value: initialState.code,
      language: 'typescript',
      theme: editorTheme,
      automaticLayout: true,
      scrollBeyondLastLine: false,
      fixedOverflowWidgets: true,
    });
    codeEditorRef.current = editor;

    // Trigger initial onChange
    const initialVal = editor.getValue() || '';
    onChangeRef.current(initialVal);
    scheduleSerializeToUrl();

    editor.onDidChangeModelContent(() => {
      const val = editor.getValue() || '';
      onChangeRef.current(val);
      scheduleSerializeToUrl();

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
      onSelectionChangeRef.current?.(
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
    jsonDefaults.setDiagnosticsOptions({
      validate: true,
      allowComments: true,
      trailingCommas: 'ignore',
      schemaValidation: 'ignore',
    });
  }, []);

  // Create rslint.config.js editor.
  useEffect(() => {
    if (!rslintContainerRef.current) return;

    javascriptDefaults.setCompilerOptions({
      allowJs: true,
      checkJs: true,
      module: ModuleKind.ESNext,
      moduleResolution: ModuleResolutionKind.NodeJs,
      target: ScriptTarget.ESNext,
    });
    javascriptDefaults.setDiagnosticsOptions({
      noSemanticValidation: false,
      noSyntaxValidation: false,
    });
    const model = monaco.editor.createModel(
      initialState.rslintConfig,
      'javascript',
      monaco.Uri.parse('file:///rslint.config.js'),
    );
    const editor = monaco.editor.create(rslintContainerRef.current, {
      model,
      theme: editorTheme,
      automaticLayout: true,
      scrollBeyondLastLine: false,
      fixedOverflowWidgets: true,
    });
    rslintEditorRef.current = editor;

    editor.onDidChangeModelContent(() => {
      onConfigChangeRef.current?.();
      scheduleSerializeToUrl();
    });

    return () => {
      editor.dispose();
      model.dispose();
    };
  }, []);

  // Create tsconfig.json editor
  useEffect(() => {
    if (!tsconfigContainerRef.current) return;

    const editor = monaco.editor.create(tsconfigContainerRef.current, {
      value: initialState.tsconfig,
      language: 'json',
      theme: editorTheme,
      automaticLayout: true,
      scrollBeyondLastLine: false,
      fixedOverflowWidgets: true,
    });
    tsconfigEditorRef.current = editor;

    editor.onDidChangeModelContent(() => {
      onConfigChangeRef.current?.();
      scheduleSerializeToUrl();
    });

    return () => {
      editor.dispose();
    };
  }, []);

  useEffect(() => {
    monaco.editor.setTheme(editorTheme);
  }, [editorTheme]);

  const tabs: { key: EditorTabType; label: string }[] = [
    { key: 'code', label: 'Code' },
    { key: 'rslint', label: 'rslint.config.js' },
    { key: 'tsconfig', label: 'tsconfig.json' },
  ];

  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex items-center justify-between gap-2 bg-[var(--rp-c-bg-soft)] p-2 flex-shrink-0">
        <div className="flex items-center gap-2">
          {tabs.map((tab) => (
            <Button
              key={tab.key}
              type="button"
              variant={activeTab === tab.key ? 'default' : 'outline'}
              size="sm"
              onClick={() => setActiveTab(tab.key)}
              aria-pressed={activeTab === tab.key}
            >
              {tab.label}
            </Button>
          ))}
        </div>
        {toolbarEnd}
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
