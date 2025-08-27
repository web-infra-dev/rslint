import React, { use, useEffect, useRef, useState } from 'react';
import * as Rslint from '@rslint/wasm';

import './Playground.css';
const wasmURL = new URL('@rslint/wasm/rslint.wasm', import.meta.url).href;
console.log(wasmURL);

interface Diagnostic {
  ruleName: string;
  message: string;
  range: {
    start: { line: number; column: number };
    end: { line: number; column: number };
  };
}

interface LintResult {
  diagnostics: Diagnostic[];
}

const Playground: React.FC = () => {
  const [service, setService] = useState<any>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [results, setResults] = useState<LintResult | null>(null);
  const [status, setStatus] = useState<string>('Ready to lint your code');
  const [error, setError] = useState<string | null>(null);

  const editorRef = useRef<HTMLTextAreaElement>(null);
  const lineNumbersRef = useRef<HTMLDivElement>(null);
  const resultsCountRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    runLint();
  }, []);

  // Initialize line numbers and error highlighting
  const initializeCodeEditor = () => {
    const editor = editorRef.current;
    const lineNumbers = lineNumbersRef.current;

    if (!editor || !lineNumbers) return;

    const updateLineNumbers = () => {
      const lines = editor.value.split('\n');
      lineNumbers.innerHTML = lines.map((_, index) => index + 1).join('\n');
    };

    const updateLineNumbersHeight = () => {
      lineNumbers.style.height = editor.scrollHeight + 'px';
    };

    const handleInput = () => {
      updateLineNumbers();
      updateLineNumbersHeight();
    };

    const handleScroll = () => {
      lineNumbers.scrollTop = editor.scrollTop;
    };

    editor.addEventListener('input', handleInput);
    editor.addEventListener('scroll', handleScroll);

    // Initial setup
    updateLineNumbers();
    updateLineNumbersHeight();

    // Cleanup function
    return () => {
      editor.removeEventListener('input', handleInput);
      editor.removeEventListener('scroll', handleScroll);
    };
  };

  // Initialize RSLint service
  const initializeService = async () => {
    if (isInitialized) return service;

    try {
      setIsLoading(true);
      setStatus('Initializing RSLint...');
      const rslintService = await Rslint.initialize({
        wasmURL: wasmURL,
      });

      setService(rslintService);
      setIsInitialized(true);
      setStatus('RSLint initialized successfully!');

      setTimeout(() => {
        setStatus('Ready to lint your code');
        setIsLoading(false);
      }, 1000);
      return rslintService;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      setError(`Failed to initialize RSLint: ${errorMessage}`);
      setIsLoading(false);
    }
  };

  // Run linting with RSLint
  const runLint = async () => {
    const service = await initializeService();

    try {
      setIsLoading(true);
      setStatus('Running linter...');
      setError(null);

      const code = editorRef.current?.value || '';

      // Build rule options based on checkboxes
      const ruleOptions: Record<string, string> = {};
      const noUnsafeMemberAccess = document.getElementById(
        'noUnsafeMemberAccess',
      ) as HTMLInputElement;
      const noUnsafeAssignment = document.getElementById(
        'noUnsafeAssignment',
      ) as HTMLInputElement;
      const noUnsafeCall = document.getElementById(
        'noUnsafeCall',
      ) as HTMLInputElement;

      if (noUnsafeMemberAccess?.checked) {
        ruleOptions['@typescript-eslint/no-unsafe-member-access'] = 'error';
      }
      if (noUnsafeAssignment?.checked) {
        ruleOptions['@typescript-eslint/no-unsafe-assignment'] = 'error';
      }
      if (noUnsafeCall?.checked) {
        ruleOptions['@typescript-eslint/no-unsafe-call'] = 'error';
      }

      const result = await service.lint({
        config: 'rslint.json',
        ruleOptions,
        fileContents: {
          '/index.ts': code,
        },
      });

      setResults(result);
      setStatus('Linting completed');
      setIsLoading(false);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      setError(`Linting failed: ${errorMessage}`);
      setIsLoading(false);
    }
  };

  // Reset code to default
  const resetCode = () => {
    if (editorRef.current) {
      editorRef.current.value = `let a: any;
a.b = 10;

function example() {
  const obj: any = { prop: 'value' };
  return obj.unknownProperty;
}

const result = example();
console.log(result);`;

      // Trigger input event to update line numbers
      const event = new Event('input');
      editorRef.current.dispatchEvent(event);
    }
  };

  // Format code (basic implementation)
  const formatCode = () => {
    const editor = editorRef.current;
    if (!editor) return;

    const code = editor.value;

    // Basic indentation fix
    const lines = code.split('\n');
    const formatted = lines.map(line => line.trim()).join('\n');

    editor.value = formatted;

    // Trigger input event to update line numbers
    const event = new Event('input');
    editor.dispatchEvent(event);
  };

  // Update results count
  useEffect(() => {
    if (resultsCountRef.current) {
      if (!results || results.diagnostics.length === 0) {
        resultsCountRef.current.textContent = '0 issues';
      } else {
        resultsCountRef.current.textContent = `${results.diagnostics.length} issue${results.diagnostics.length !== 1 ? 's' : ''}`;
      }
    }
  }, [results]);

  // Initialize on component mount
  useEffect(() => {
    const cleanup = initializeCodeEditor();
    initializeService();

    return cleanup;
  }, []);

  return (
    <div className="playground-container">
      <div className="playground-header">
        <div className="playground-title">Code Editor</div>
        <div className="playground-actions">
          <button className="btn btn-secondary" onClick={resetCode}>
            Reset
          </button>
          <button className="btn btn-secondary" onClick={formatCode}>
            Format
          </button>
        </div>
      </div>

      <div className="playground-main">
        <div className="editor-panel">
          <div className="code-editor-container">
            <div className="line-numbers" ref={lineNumbersRef}></div>
            <textarea
              ref={editorRef}
              className="code-editor"
              placeholder="Enter your TypeScript/JavaScript code here..."
              spellCheck={false}
              defaultValue={`let a: any;
a.b = 10;

function example() {
  const obj: any = { prop: 'value' };
  return obj.unknownProperty;
}

const result = example();
console.log(result);`}
            />
          </div>
        </div>

        <div className="config-panel">
          <div className="config-header">
            <div className="config-title">Configuration</div>
            <div className="config-actions">
              <button
                className="btn btn-primary"
                onClick={runLint}
                disabled={isLoading}
              >
                {isLoading ? 'Running...' : 'Run Lint'}
              </button>
            </div>
          </div>

          <div className="config-content">
            <div className="config-group">
              <label className="config-label">Language</label>
              <select className="config-input" defaultValue="typescript">
                <option value="typescript">TypeScript</option>
                <option value="javascript">JavaScript</option>
              </select>
            </div>

            <div className="config-group">
              <label className="config-label">Rules</label>
              <div className="rule-checkbox">
                <label>
                  <input
                    type="checkbox"
                    id="noUnsafeMemberAccess"
                    defaultChecked
                  />
                  @typescript-eslint/no-unsafe-member-access
                </label>
              </div>
              <div className="rule-checkbox">
                <label>
                  <input type="checkbox" id="noUnsafeAssignment" />
                  @typescript-eslint/no-unsafe-assignment
                </label>
              </div>
              <div className="rule-checkbox">
                <label>
                  <input type="checkbox" id="noUnsafeCall" />
                  @typescript-eslint/no-unsafe-call
                </label>
              </div>
            </div>

            <div className="config-group">
              <label className="config-label">Strict Mode</label>
              <label className="rule-checkbox">
                <input type="checkbox" id="strictMode" defaultChecked />
                Enable strict type checking
              </label>
            </div>
          </div>
        </div>
      </div>

      <div className="results-panel">
        <div className="results-header">
          <div className="results-title">Linting Results</div>
          <div className="results-count" ref={resultsCountRef}>
            No results yet
          </div>
        </div>
        <div className="results-content">
          {error && (
            <div className="error">
              <strong>Error:</strong> {error}
            </div>
          )}

          {isLoading && (
            <div className="loading">
              <div className="spinner"></div>
              {status}
            </div>
          )}

          {!isLoading &&
            !error &&
            results &&
            results.diagnostics.length === 0 && (
              <div className="success">
                <strong>No issues found!</strong> Your code looks good.
              </div>
            )}

          {!isLoading &&
            !error &&
            results &&
            results.diagnostics.length > 0 && (
              <>
                {results.diagnostics.map((diagnostic, index) => (
                  <div key={index} className="result-item">
                    <div className="result-header">
                      <span className="result-rule">{diagnostic.ruleName}</span>
                      <span className="result-severity">Error</span>
                    </div>
                    <div className="result-message">{diagnostic.message}</div>
                    <div className="result-location">
                      Line {diagnostic.range.start.line}, Column{' '}
                      {diagnostic.range.start.column}
                    </div>
                  </div>
                ))}
              </>
            )}

          {!isLoading && !error && !results && (
            <div className="loading">Ready to lint your code</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Playground;
