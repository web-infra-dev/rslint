import React, { useEffect, useState } from 'react';
import './ResultPanel.css';

export interface Diagnostic {
  ruleName: string;
  message: string;
  range: {
    start: { line: number; column: number };
    end: { line: number; column: number };
  };
}

interface ResultPanelProps {
  diagnostics: Diagnostic[];
  ast?: string;
  initialized?: boolean;
  error?: string;
  fixedCode?: string;
  typeInfo?: string;
  loading?: boolean;
}

type TabType = 'lint' | 'fixed' | 'ast' | 'type';

export const ResultPanel: React.FC<ResultPanelProps> = props => {
  const { diagnostics, ast, error, initialized, fixedCode, typeInfo, loading } =
    props;
  const [activeTab, setActiveTab] = useState<TabType>(() => {
    if (typeof window === 'undefined') return 'lint';
    const params = new URLSearchParams(window.location.search);
    let tab = params.get('tab');
    if (!tab && window.location.hash) {
      const hashParams = new URLSearchParams(window.location.hash.slice(1));
      tab = hashParams.get('tab');
    }
    if (tab === 'lint' || tab === 'ast' || tab === 'fixed' || tab === 'type') {
      return tab as TabType;
    }
    return 'lint';
  });

  // Keep the URL in sync with the selected tab
  useEffect(() => {
    if (typeof window === 'undefined') return;
    try {
      const url = new URL(window.location.href);
      url.searchParams.set('tab', activeTab);
      if (url.hash && new URLSearchParams(url.hash.slice(1)).has('tab')) {
        url.hash = '';
      }
      window.history.replaceState(null, '', url.toString());
    } catch {
      // ignore URL update errors
    }
  }, [activeTab]);

  // Respond to browser navigation updating the tab
  useEffect(() => {
    if (typeof window === 'undefined') return;
    const handler = () => {
      try {
        const params = new URLSearchParams(window.location.search);
        let tab = params.get('tab');
        if (!tab && window.location.hash) {
          const hashParams = new URLSearchParams(window.location.hash.slice(1));
          tab = hashParams.get('tab');
        }
        if (
          tab === 'lint' ||
          tab === 'ast' ||
          tab === 'fixed' ||
          tab === 'type'
        ) {
          setActiveTab(tab as TabType);
        }
      } catch {
        // ignore
      }
    };
    window.addEventListener('popstate', handler);
    return () => window.removeEventListener('popstate', handler);
  }, []);

  return (
    <div className="result-panel">
      <div className="result-header">
        <div className="result-tabs">
          <div
            className={`result-tab ${activeTab === 'lint' ? 'active' : ''}`}
            onClick={() => setActiveTab('lint')}
            title="Errors"
          >
            Errors
          </div>
          <div
            className={`result-tab ${activeTab === 'ast' ? 'active' : ''}`}
            onClick={() => setActiveTab('ast')}
            title="AST"
          >
            AST
          </div>
        </div>
      </div>

      {initialized ? (
        <div className="result-content">
          {error && (
            <div className="error-message">
              <div className="error-icon">⚠️</div>
              <div className="error-text">
                <strong>Error:</strong> {error}
              </div>
            </div>
          )}

          {!error && activeTab === 'lint' && (
            <div className="lint-results">
              {diagnostics.length === 0 ? (
                <div className="success-message">
                  <div className="success-icon">✅</div>
                  <div className="success-text">
                    <strong>No issues found!</strong> Your code looks good.
                  </div>
                </div>
              ) : (
                <div className="diagnostics-list">
                  {diagnostics.map((diagnostic, index) => (
                    <div key={index} className="diagnostic-item">
                      <div className="diagnostic-header">
                        <h4>{diagnostic.ruleName}</h4>
                      </div>
                      <div className="diagnostic-message">
                        {diagnostic.message} {diagnostic.range.start.line}:
                        {diagnostic.range.start.column} -{' '}
                        {diagnostic.range.end.line}:
                        {diagnostic.range.end.column}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {!error && activeTab === 'ast' && (
            <div className="ast-view">
              {ast ? (
                <div className="code-block-wrapper">
                  <pre className="ast-content">{ast}</pre>
                </div>
              ) : (
                <div className="empty-state">
                  <div className="empty-text">AST will be displayed here</div>
                </div>
              )}
            </div>
          )}
        </div>
      ) : (
        <div className="result-content">
          {loading ? (
            <div className="loading-state">
              <div className="spinner"></div>
              <div>Loading WASM...</div>
            </div>
          ) : error ? (
            <div className="error-message">
              <div className="error-icon">⚠️</div>
              <div className="error-text">
                <strong>Error:</strong> {error}
              </div>
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
};
