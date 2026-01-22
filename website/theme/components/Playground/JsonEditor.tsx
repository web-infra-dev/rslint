import React, { useRef, useEffect, Ref } from 'react';
import * as monaco from 'monaco-editor';

// Configure JSON defaults to allow comments and trailing commas
monaco.json.jsonDefaults.setDiagnosticsOptions({
  validate: true,
  allowComments: true,
  trailingCommas: 'ignore',
});

export interface JsonEditorRef {
  getValue: () => string | undefined;
}

export const JsonEditor = ({
  ref,
  defaultValue,
  onChange,
}: {
  ref: Ref<JsonEditorRef>;
  defaultValue: string;
  onChange?: (value: string) => void;
}) => {
  const divEl = useRef<HTMLDivElement>(null);
  const editorRef =
    useRef<import('monaco-editor').editor.IStandaloneCodeEditor>(null);

  React.useImperativeHandle(ref, () => ({
    getValue: () => editorRef.current?.getValue(),
  }));

  useEffect(() => {
    if (!divEl.current) {
      return;
    }

    const editor = monaco.editor.create(divEl.current, {
      value: defaultValue,
      language: 'json',
      automaticLayout: true,
      scrollBeyondLastLine: false,
      minimap: { enabled: false },
    });

    editorRef.current = editor;

    editor.onDidChangeModelContent(() => {
      const val = editor.getValue() || '';
      onChange?.(val);
    });

    // Trigger initial onChange
    onChange?.(editor.getValue() || '');

    return () => {
      editor.dispose();
    };
  }, []);

  return <div className="h-full w-full" ref={divEl}></div>;
};
