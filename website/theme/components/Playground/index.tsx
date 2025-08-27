import React, { use, useEffect, useRef, useState } from 'react';
import * as Rslint from '@rslint/wasm';
import { Editor, EditorRef } from './Editor';
import { ResultPanel, Diagnostic } from './ResultPanel';
import './index.css';

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

// Placeholder AST service - replace with actual implementation
async function getAST(code: string): Promise<string> {
  // This is a placeholder - replace with actual AST generation
  return `// Abstract Syntax Tree for the code
Program {
  body: [
    VariableDeclaration {
      kind: "let",
      declarations: [
        VariableDeclarator {
          id: Identifier { name: "a" },
          init: null
        }
      ]
    },
    ExpressionStatement {
      expression: AssignmentExpression {
        operator: "=",
        left: MemberExpression {
          object: Identifier { name: "a" },
          property: Identifier { name: "b" }
        },
        right: Literal { value: 10 }
      }
    }
  ]
}`;
}

const Playground: React.FC = () => {
  const editorRef = useRef<EditorRef | null>(null);
  const [diagnostics, setDiagnostics] = useState<Diagnostic[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [ast, setAst] = useState<string | undefined>();

  async function runLint() {
    try {
      setIsLoading(true);
      setError(undefined);

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

      // Generate AST
      try {
        const astData = await getAST(code);
        setAst(astData);
      } catch (astError) {
        console.warn('AST generation failed:', astError);
        setAst(undefined);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      setError(`Linting failed: ${errorMessage}`);
    } finally {
      setIsLoading(false);
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
        diagnostics={diagnostics}
        ast={ast}
        isLoading={isLoading}
        error={error}
      />
    </div>
  );
};

export default Playground;
