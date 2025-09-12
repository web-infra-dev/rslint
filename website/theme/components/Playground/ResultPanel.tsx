import React, { useEffect, useState } from 'react';
import { Button } from '@components/ui/button';
import { Share2Icon, CheckIcon } from 'lucide-react';
// Removed ToggleGroup in favor of Button to match Share style
import { Alert, AlertDescription, AlertTitle } from '@components/ui/alert';
import { AlertCircleIcon } from 'lucide-react';
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
  astTree?: ASTNode;
  initialized?: boolean;
  error?: string;
  fixedCode?: string;
  typeInfo?: string;
  loading?: boolean;
  onAstNodeSelect?: (start: number, end: number) => void;
  selectedAstNodeRange?: { start: number; end: number };
}

type TabType = 'lint' | 'fixed' | 'ast' | 'type';

interface ASTNode {
  type: string;
  start: number;
  end: number;
  name?: string;
  text?: string;
  children?: ASTNode[];
}

export const ResultPanel: React.FC<ResultPanelProps> = props => {
  const {
    diagnostics,
    ast,
    astTree,
    error,
    initialized,
    fixedCode,
    typeInfo,
    loading,
    onAstNodeSelect,
    selectedAstNodeRange,
  } = props;
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

  // AST tree view state
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set());
  const [selectedId, setSelectedId] = useState<string | null>(null);

  function nodeId(n: ASTNode) {
    return `${n.type}:${n.start}-${n.end}`;
  }

  function toggleNode(n: ASTNode) {
    const id = nodeId(n);
    setExpanded(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function clickNode(n: ASTNode) {
    setSelectedId(nodeId(n));
    onAstNodeSelect?.(n.start, n.end);
  }

  function isExpandable(n?: ASTNode) {
    return !!(n && n.children && n.children.length);
  }

  const renderNode = (n: ASTNode, depth = 0) => {
    const id = nodeId(n);
    const open = expanded.has(id);
    const hasKids = isExpandable(n);
    const preview = n.text ? n.text.replace(/\s+/g, ' ').slice(0, 40) : '';
    return (
      <div key={id} className="ast-node" style={{ paddingLeft: depth * 14 }}>
        <div
          className={`ast-node-row ${selectedId === id ? 'selected' : ''}`}
          onClick={() => clickNode(n)}
        >
          {hasKids ? (
            <button
              className={`twisty ${open ? 'open' : ''}`}
              onClick={e => {
                e.stopPropagation();
                toggleNode(n);
              }}
              aria-label={open ? 'Collapse' : 'Expand'}
            />
          ) : (
            <span className="twisty placeholder" />
          )}
          <span className="node-type">{n.type}</span>
          <span className="node-range">
            [{n.start}, {n.end}]
          </span>
          {preview && <span className="node-preview">“{preview}”</span>}
        </div>
        {open && hasKids && (
          <div className="ast-children">
            {n.children!.map(child => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  // When selection in editor changes, select smallest covering AST node and expand its ancestors
  useEffect(() => {
    if (!astTree || !selectedAstNodeRange) return;
    const { start, end } = selectedAstNodeRange;

    let best: { node: ASTNode; depth: number; path: ASTNode[] } | null = null;

    function visit(node: ASTNode, depth: number, path: ASTNode[]) {
      if (node.start <= start && node.end >= end) {
        if (!best || depth > best.depth) {
          best = { node, depth, path: [...path, node] };
        }
        if (node.children) {
          for (const c of node.children) visit(c, depth + 1, [...path, node]);
        }
      }
    }
    visit(astTree, 0, []);
    if (best) {
      const id = nodeId(best.node);
      setSelectedId(id);
      setExpanded(prev => {
        const next = new Set(prev);
        for (const p of best!.path) next.add(nodeId(p));
        return next;
      });
    }
  }, [selectedAstNodeRange, astTree]);

  // Auto-expand root when a new tree arrives
  useEffect(() => {
    if (!astTree) return;
    const id = nodeId(astTree);
    setExpanded(prev => {
      if (prev.has(id)) return prev;
      const next = new Set(prev);
      next.add(id);
      return next;
    });
  }, [astTree]);

  // Share button state and handler
  const [shareCopied, setShareCopied] = useState(false);
  async function copyShareUrl() {
    try {
      const url = window.location.href;
      await copyToClipboard(url);
      setShareCopied(true);
      window.setTimeout(() => setShareCopied(false), 1500);
    } catch (e) {
      console.warn('Share failed', e);
    }
  }

  return (
    <div className="result-panel">
      <div className="result-header">
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant={activeTab === 'lint' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setActiveTab('lint')}
            aria-pressed={activeTab === 'lint'}
          >
            Errors
          </Button>
          <Button
            type="button"
            variant={activeTab === 'ast' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setActiveTab('ast')}
            aria-pressed={activeTab === 'ast'}
          >
            AST
          </Button>
        </div>
        <div className="result-actions">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => copyShareUrl()}
            title={shareCopied ? 'Copied link' : 'Copy shareable link'}
          >
            {shareCopied ? (
              <CheckIcon className="size-4" />
            ) : (
              <Share2Icon className="size-4" />
            )}
            {shareCopied ? 'Copied' : 'Share'}
          </Button>
        </div>
      </div>

      {initialized ? (
        <div className="result-content">
          {error && (
            <Alert variant="destructive">
              <AlertCircleIcon />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {!error && activeTab === 'lint' && (
            <div className="lint-results">
              {diagnostics.length === 0 ? (
                <Alert>
                  <AlertTitle>No issues found!</AlertTitle>
                  <AlertDescription>Your code looks good.</AlertDescription>
                </Alert>
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
              {astTree ? (
                <div className="ast-tree" role="tree">
                  {renderNode(astTree)}
                </div>
              ) : ast ? (
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

function copyToClipboard(text: string) {
  if (navigator.clipboard?.writeText)
    return navigator.clipboard.writeText(text);
  return new Promise<void>((resolve, reject) => {
    try {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.setAttribute('readonly', '');
      ta.style.position = 'absolute';
      ta.style.left = '-9999px';
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
      resolve();
    } catch (e) {
      reject(e);
    }
  });
}
