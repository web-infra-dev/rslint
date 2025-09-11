import React, { useEffect, useRef, useState } from 'react';
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
  const [astTree, setAstTree] = useState<any | undefined>();
  const [loading, setLoading] = useState(true);
  const lintTimer = useRef<number | null>(null);
  const [selectedAstRange, setSelectedAstRange] = useState<
    { start: number; end: number } | undefined
  >();

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
          const current: ASTNode = {
            type: SyntaxKind[node.kind],
            start: node.pos,
            end: node.end,
            text: (node as any).text,
          };
          const children: ASTNode[] = [];
          node.forEachChild((child: Node) => {
            children.push(RemoteNodeToEstree(child));
          });
          if (children.length) current.children = children;
          return current;
        }

        const tree = RemoteNodeToEstree(source);
        const astData = JSON.stringify(tree, null, 2);
        setAstTree(tree);
        setAst(astData);
      } catch (astError) {
        console.warn('AST generation failed:', astError);
        setAst(undefined);
        setAstTree(undefined);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      setError(`Linting failed: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  }

  // Debounce linting to reduce recomputation while typing
  function scheduleRunLint() {
    if (lintTimer.current) {
      window.clearTimeout(lintTimer.current);
      lintTimer.current = null;
    }
    lintTimer.current = window.setTimeout(() => {
      runLint();
    }, 250);
  }

  // Cleanup any pending timers on unmount
  useEffect(() => {
    return () => {
      if (lintTimer.current) {
        window.clearTimeout(lintTimer.current);
        lintTimer.current = null;
      }
    };
  }, []);
  // Initial lint is triggered by Editor's initial onChange

  return (
    <div className="playground-container">
      <div className="editor-panel">
        <Editor
          ref={editorRef}
          onChange={() => scheduleRunLint()}
          onSelectionChange={(start, end) =>
            setSelectedAstRange({ start, end })
          }
        />
      </div>
      <ResultPanel
        initialized={initialized}
        diagnostics={diagnostics}
        ast={ast}
        astTree={astTree}
        error={error}
        loading={loading}
        onAstNodeSelect={(start, end) =>
          editorRef.current?.revealRangeByOffset(start, end)
        }
        selectedAstNodeRange={selectedAstRange}
      />
    </div>
  );
};

export default Playground;
