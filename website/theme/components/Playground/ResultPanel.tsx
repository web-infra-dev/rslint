import React, { useState } from 'react';
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
}

type TabType = 'lint' | 'fixed' | 'ast' | 'type';

export const ResultPanel: React.FC<ResultPanelProps> = props => {
  const { diagnostics, ast, error, initialized, fixedCode, typeInfo } = props;
  const [activeTab, setActiveTab] = useState<TabType>('lint');

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
      ) : null}
    </div>
  );
};
