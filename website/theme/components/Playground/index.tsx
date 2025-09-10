import React, { use, useEffect, useRef, useState } from 'react';
import * as Rslint from '@rslint/wasm';
import { Editor, EditorRef } from './Editor';
import { ResultPanel, Diagnostic } from './ResultPanel';
import './index.css';
import { RemoteSourceFile, type Node, SyntaxKind } from '@rslint/api';

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
  const [diagnostics, setDiagnostics] = useState<Diagnostic[]>([]);
  const [initialized, setInitialized] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [ast, setAst] = useState<string | undefined>();
  const [loading, setLoading] = useState(true);

  async function runLint() {
    try {
      setError(undefined);
      if (!initialized) setLoading(true);
      const service = await ensureService();
      const code = editorRef.current?.getValue() ?? '';

      const result = await service.lint({
        includeEncodedSourceFiles: true,
        fileContents: {
          '/index.ts': code,
        },
        config: 'rslint.json',
        ruleOptions: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
        },
      });
      setInitialized(true);

      // Convert diagnostics to the expected format
      const convertedDiagnostics: Diagnostic[] = result.diagnostics.map(
        diag => ({
          ruleName: diag.ruleName,
          message: diag.message,
          range: {
            start: {
              line: diag.range.start.line,
              column: diag.range.start.column,
            },
            end: { line: diag.range.end.line, column: diag.range.end.column },
          },
        }),
      );

      setDiagnostics(convertedDiagnostics);
      editorRef.current?.attachDiag(result.diagnostics);
      interface ASTNode {
        type: string;
        start: number;
        end: number;
        name?: string;
        children?: ASTNode[];
        text?: string;
      }

      // Generate AST
      try {
        const astBuffer = result.encodedSourceFiles!['index.ts'];
        const buffer = Uint8Array.from(atob(astBuffer), c => c.charCodeAt(0));
        const source = new RemoteSourceFile(buffer, new TextDecoder());
        // Convert a RemoteNode (from tsgo/rslint-api) to a minimal ESTree node
        
        function RemoteNodeToEstree(node: Node): ASTNode {
          return {
            type: SyntaxKind[node.kind],
            start: node.pos,
            end: node.end,
            text: node.text,
            children: node.forEachChild((child: Node) => {
              return RemoteNodeToEstree(child);
            }),
          };
        }

        const astData = JSON.stringify(RemoteNodeToEstree(source), null, 2);
        setAst(astData);
      } catch (astError) {
        console.warn('AST generation failed:', astError);
        setAst(undefined);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      setError(`Linting failed: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    runLint();
  }, []);

  return (
    <div className="playground-container">
      <div className="editor-panel">
        <Editor ref={editorRef} onChange={() => runLint()} />
      </div>
      <ResultPanel
        initialized={initialized}
        diagnostics={diagnostics}
        ast={ast}
        error={error}
        loading={loading}
      />
    </div>
  );
};

export default Playground;
