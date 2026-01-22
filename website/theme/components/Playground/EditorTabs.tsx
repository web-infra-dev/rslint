import React, { useRef, useState, Ref } from 'react';
import { parse, ParseError } from 'jsonc-parser';
import { Editor, EditorRef } from './Editor';
import { JsonEditor, JsonEditorRef } from './JsonEditor';
import { Button } from '@components/ui/button';

function isValidJsonc(text: string): boolean {
  const errors: ParseError[] = [];
  parse(text, errors);
  return errors.length === 0;
}

export type EditorTabType = 'code' | 'rslint.json' | 'tsconfig';

export interface EditorTabsRef {
  getCodeValue: () => string | undefined;
  getRslintConfig: () => string;
  getTsconfigContent: () => string;
  getEditorRef: () => EditorRef | null;
}

interface EditorTabsProps {
  defaultRslintConfig: string;
  defaultTsconfig: string;
  onCodeChange?: (value: string) => void;
  onSelectionChange?: (start: number, end: number) => void;
  onRslintConfigChange?: (value: string) => void;
  onTsconfigChange?: (value: string) => void;
}

export const EditorTabs = ({
  ref,
  defaultRslintConfig,
  defaultTsconfig,
  onCodeChange,
  onSelectionChange,
  onRslintConfigChange,
  onTsconfigChange,
}: EditorTabsProps & { ref: Ref<EditorTabsRef> }) => {
  const editorRef = useRef<EditorRef | null>(null);
  const rslintEditorRef = useRef<JsonEditorRef | null>(null);
  const tsconfigEditorRef = useRef<JsonEditorRef | null>(null);

  const [activeTab, setActiveTab] = useState<EditorTabType>('code');
  const [rslintConfig, setRslintConfig] = useState(defaultRslintConfig);
  const [tsconfigContent, setTsconfigContent] = useState(defaultTsconfig);

  React.useImperativeHandle(ref, () => ({
    getCodeValue: () => editorRef.current?.getValue(),
    getRslintConfig: () => rslintConfig,
    getTsconfigContent: () => tsconfigContent,
    getEditorRef: () => editorRef.current,
  }));

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 p-2">
        <Button
          type="button"
          variant={activeTab === 'code' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setActiveTab('code')}
          aria-pressed={activeTab === 'code'}
        >
          code
        </Button>
        <Button
          type="button"
          variant={activeTab === 'rslint.json' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setActiveTab('rslint.json')}
          aria-pressed={activeTab === 'rslint.json'}
        >
          rslint.json
        </Button>
        <Button
          type="button"
          variant={activeTab === 'tsconfig' ? 'default' : 'outline'}
          size="sm"
          onClick={() => setActiveTab('tsconfig')}
          aria-pressed={activeTab === 'tsconfig'}
        >
          tsconfig
        </Button>
      </div>
      <div className={`flex-1 ${activeTab !== 'code' ? 'hidden' : ''}`}>
        <Editor
          ref={editorRef}
          onChange={v => onCodeChange?.(v)}
          onSelectionChange={onSelectionChange}
        />
      </div>
      <div className={`flex-1 ${activeTab !== 'rslint.json' ? 'hidden' : ''}`}>
        <JsonEditor
          ref={rslintEditorRef}
          defaultValue={defaultRslintConfig}
          onChange={v => {
            if (isValidJsonc(v)) {
              setRslintConfig(v);
              onRslintConfigChange?.(v);
            }
          }}
        />
      </div>
      <div className={`flex-1 ${activeTab !== 'tsconfig' ? 'hidden' : ''}`}>
        <JsonEditor
          ref={tsconfigEditorRef}
          defaultValue={defaultTsconfig}
          onChange={v => {
            if (isValidJsonc(v)) {
              setTsconfigContent(v);
              onTsconfigChange?.(v);
            }
          }}
        />
      </div>
    </div>
  );
};
