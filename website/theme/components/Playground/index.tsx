import React, { useEffect, useRef, useState } from 'react';
import * as Rslint from '@rslint/wasm';
import type { RemoteTypeChecker } from '@rslint/wasm';
import { parse as parseJsonc } from 'jsonc-parser';
import { EditorTabs, EditorTabsRef } from './EditorTabs';
import { ResultPanel, Diagnostic, NodeDetailInfo } from './ResultPanel';
import {
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle,
} from '@components/ui/resizable';
import './index.css';
import { RemoteSourceFile, type Node, SyntaxKind, RemoteNode } from '@rslint/api';

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

const DEFAULT_RSLINT_CONFIG = JSON.stringify(
  [
    {
      languageOptions: {
        parserOptions: {
          project: ['./tsconfig.json'],
        },
      },
      rules: {
        '@typescript-eslint/no-unsafe-member-access': 'error',
      },
      plugins: ['@typescript-eslint'],
    },
  ],
  null,
  2,
);

const DEFAULT_TSCONFIG = JSON.stringify({
  compilerOptions: {},
}, null, 2);

const Playground: React.FC = () => {
  const editorTabsRef = useRef<EditorTabsRef | null>(null);

  const [code, setCode] = useState('');
  const [rslintConfig, setRslintConfig] = useState(DEFAULT_RSLINT_CONFIG);
  const [tsconfigContent, setTsconfigContent] = useState(DEFAULT_TSCONFIG);

  const [diagnostics, setDiagnostics] = useState<Diagnostic[]>([]);
  const [initialized, setInitialized] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [ast, setAst] = useState<string | undefined>();
  const [astTree, setAstTree] = useState<any | undefined>();
  const [tsAstTree, setTsAstTree] = useState<any | undefined>();
  const tsModuleRef = useRef<null | typeof import('typescript')>(null);
  const tsAstActiveRef = useRef<boolean>(false);
  const lastSourceTextRef = useRef<string>('');
  const [loading, setLoading] = useState(true);
  const lintTimer = useRef<number | null>(null);
  const [selectedAstRange, setSelectedAstRange] = useState<
    { start: number; end: number } | undefined
  >();
  const [selectedNodeDetail, setSelectedNodeDetail] = useState<NodeDetailInfo | undefined>();
  // Store the RemoteSourceFile for node lookup
  const remoteSourceFileRef = useRef<RemoteSourceFile | null>(null);
  // Store the RemoteTypeChecker for type queries
  const typeCheckerRef = useRef<RemoteTypeChecker | null>(null);

  async function runLint() {
    try {
      setError(undefined);
      if (!initialized) setLoading(true);
      const service = await ensureService();
      // Parse JSONC to extract rules and convert to standard JSON
      const rslintConfigParsed = parseJsonc(rslintConfig);
      const rslintConfigJson = JSON.stringify(rslintConfigParsed);
      const tsconfigParsed = parseJsonc(tsconfigContent);
      const tsconfigJson = JSON.stringify(tsconfigParsed);

      // Extract rules from rslint config (support both array and object format)
      let ruleOptions: Record<string, unknown> = {};
      if (Array.isArray(rslintConfigParsed)) {
        // Flat config format: array of config objects
        for (const config of rslintConfigParsed) {
          if (config?.rules) {
            ruleOptions = { ...ruleOptions, ...config.rules };
          }
        }
      } else if (rslintConfigParsed?.rules) {
        // Legacy config format: single object with rules
        ruleOptions = rslintConfigParsed.rules;
      }

      const result = await service.lint({
        includeEncodedSourceFiles: true,
        includeTypeChecker: true, // Request a type checker session
        fileContents: {
          '/index.ts': code,
          '/tsconfig.json': tsconfigJson,
          '/rslint.json': rslintConfigJson,
        },
        config: 'rslint.json',
        ruleOptions: ruleOptions as Record<string, string>,
      });
      setInitialized(true);

      // Store the type checker for later use
      if (result.typeChecker) {
        typeCheckerRef.current = result.typeChecker;
      }

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
      editorTabsRef.current?.getEditorRef()?.attachDiag(result.diagnostics);
      interface ASTNode {
        type: string;
        start: number;
        end: number;
        name?: string;
        children?: ASTNode[];
        text?: string;
      }

      // Generate AST (tsgo)
      let sourceTextForTs: string | undefined;
      try {
        const astBuffer = result.encodedSourceFiles!['index.ts'];
        const buffer = Uint8Array.from(atob(astBuffer), c => c.charCodeAt(0));
        const source = new RemoteSourceFile(buffer, new TextDecoder());
        // Store for later node lookup
        remoteSourceFileRef.current = source;
        // capture the exact source text from encoded source file
        try {
          sourceTextForTs = (source as any).text as string | undefined;
        } catch {}
        // Convert a RemoteNode (from tsgo/rslint-api) to a minimal ESTree node

        function RemoteNodeToEstree(node: RemoteNode): ASTNode {
          const current: ASTNode = {
            type: SyntaxKind[node.kind],
            start: node.pos,
            end: node.end,
            text: (node as any).text,
          };
          // console.log('node: ', current.type, node.pos, node.end, node.kind, node.flags);
          const children: ASTNode[] = [];
          node.forEachChild((child: RemoteNode) => {
            children.push(RemoteNodeToEstree(child));
          });
          if (children.length) current.children = children;
          return current;
        }

        const tree = RemoteNodeToEstree(source);
        setAstTree(tree);
        setAst(undefined);
        // persist source text for TS parsing alignment
        lastSourceTextRef.current = sourceTextForTs ?? code;
      } catch (astError) {
        console.warn('AST generation failed:', astError);
        setAst(undefined);
        setAstTree(undefined);
        remoteSourceFileRef.current = null;
        lastSourceTextRef.current = code;
      }
      // If TS AST tab has been opened and TS module loaded, refresh TS AST too
      if (tsModuleRef.current && tsAstActiveRef.current) {
        await buildTypeScriptAst(lastSourceTextRef.current);
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

  // Find RemoteNode by position and return both the node and its details
  function findNodeAtPosition(start: number, end: number): { node: RemoteNode; detail: NodeDetailInfo } | undefined {
    const source = remoteSourceFileRef.current;
    if (!source) return undefined;

    let bestNode: RemoteNode | null = null;
    let bestDepth = -1;

    function visit(node: RemoteNode, depth: number) {
      if (node.pos <= start && node.end >= end) {
        if (depth > bestDepth) {
          bestNode = node;
          bestDepth = depth;
        }
        node.forEachChild((child: RemoteNode) => visit(child, depth + 1));
      }
    }

    visit(source, 0);

    if (!bestNode) return undefined;

    const node = bestNode as RemoteNode;
    const kindName = SyntaxKind[node.kind] || `Unknown(${node.kind})`;

    // Extract flags (basic parsing)
    const flags = node.flags;
    const flagNames: string[] = [];
    // Common NodeFlags values (from TypeScript)
    if (flags & 1) flagNames.push('Let');
    if (flags & 2) flagNames.push('Const');
    if (flags & 4) flagNames.push('NestedNamespace');
    if (flags & 8) flagNames.push('Synthesized');
    if (flags & 16) flagNames.push('Namespace');
    if (flags & 32) flagNames.push('ExportContext');
    if (flags & 64) flagNames.push('ContainsThis');
    if (flags & 128) flagNames.push('HasImplicitReturn');
    if (flags & 256) flagNames.push('HasExplicitReturn');
    if (flags & 512) flagNames.push('GlobalAugmentation');
    if (flags & 1024) flagNames.push('HasAsyncFunctions');
    if (flags & 2048) flagNames.push('DisallowInContext');
    if (flags & 4096) flagNames.push('YieldContext');
    if (flags & 8192) flagNames.push('DecoratorContext');
    if (flags & 16384) flagNames.push('AwaitContext');
    if (flags & 32768) flagNames.push('ThisNodeHasError');
    if (flags & 65536) flagNames.push('JavaScriptFile');
    if (flags & 131072) flagNames.push('ThisNodeOrAnySubNodesHasError');
    if (flags & 262144) flagNames.push('HasAggregatedChildData');

    const detail: NodeDetailInfo = {
      kind: node.kind,
      kindName,
      flags,
      flagNames: flagNames.length > 0 ? flagNames : undefined,
      pos: node.pos,
      end: node.end,
      text: (node as any).text,
      // Type/Symbol info will be fetched asynchronously
      type: undefined,
      contextualType: undefined,
      symbol: undefined,
      signature: undefined,
      flowNode: undefined,
    };

    return { node, detail };
  }

  // The file path used for type checker queries (must match the key in encodedSourceFiles/sourceFiles)
  // Note: Go side uses relative paths, so this should be 'index.ts' not '/index.ts'
  const MAIN_FILE_PATH = 'index.ts';

  // Update node detail when selection changes
  async function handleAstNodeSelect(start: number, end: number) {
    setSelectedAstRange({ start, end });
    const result = findNodeAtPosition(start, end);

    if (!result) {
      setSelectedNodeDetail(undefined);
      return;
    }

    const { node, detail } = result;

    // Use the lazy-loading APIs for type info
    if (typeCheckerRef.current) {
      try {
        const nodeLocation = {
          filePath: MAIN_FILE_PATH,
          pos: node.pos,
          kind: node.kind,
        };

        // Fetch type, symbol, signature, and flow node info in parallel
        const [typeResp, symbolResp, signatureResp, flowNodeResp] = await Promise.all([
          typeCheckerRef.current.getNodeType(nodeLocation),
          typeCheckerRef.current.getNodeSymbol(nodeLocation),
          typeCheckerRef.current.getNodeSignature(nodeLocation),
          typeCheckerRef.current.getNodeFlowNode(nodeLocation),
        ]);

        // Merge all related types and symbols from different responses
        const mergedRelatedTypes: Record<number, any> = {
          ...typeResp?.RelatedTypes,
          ...symbolResp?.RelatedTypes,
          ...signatureResp?.RelatedTypes,
        };
        const mergedRelatedSymbols: Record<number, any> = {
          ...typeResp?.RelatedSymbols,
          ...symbolResp?.RelatedSymbols,
          ...signatureResp?.RelatedSymbols,
        };

        // Update detail with full type/symbol/signature/flowNode info
        const updatedDetail: NodeDetailInfo = {
          ...detail,
          type: typeResp?.Type,
          contextualType: typeResp?.ContextualType,
          symbol: symbolResp?.Symbol,
          signature: signatureResp?.Signature,
          flowNode: flowNodeResp?.FlowNode,
          relatedTypes: Object.keys(mergedRelatedTypes).length > 0 ? mergedRelatedTypes : undefined,
          relatedSymbols: Object.keys(mergedRelatedSymbols).length > 0 ? mergedRelatedSymbols : undefined,
        };
        setSelectedNodeDetail(updatedDetail);
        return;
      } catch (e) {
        console.warn('Failed to get node type info:', e);
      }
    }

    setSelectedNodeDetail(detail);
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

  // Re-run lint when code or config changes
  useEffect(() => {
    if (code) {
      scheduleRunLint();
    }
  }, [code, rslintConfig, tsconfigContent]);

  async function buildTypeScriptAst(text: string) {
    const ts = tsModuleRef.current!;
    try {
      const sf = ts.createSourceFile(
        'index.ts',
        text,
        ts.ScriptTarget.Latest,
        /*setParentNodes*/ true,
        ts.ScriptKind.TS,
      );

      interface TSAstNode {
        type: string;
        start: number;
        end: number;
        text?: string;
        children?: TSAstNode[];
      }

      function tsNodeToTree(node: any): TSAstNode {
        const current: TSAstNode = {
          type: (ts as any).Debug.formatSyntaxKind(node.kind),
          start: node.pos,
          end: node.end,
          text: node.text,
        };
        const children: TSAstNode[] = [];
        ts.forEachChild(node, (child: any) => {
          children.push(tsNodeToTree(child));
        });
        if (children.length) current.children = children;
        return current;
      }

      setTsAstTree(tsNodeToTree(sf));
    } catch (e) {
      console.warn('TypeScript AST generation failed:', e);
      setTsAstTree(undefined);
    }
  }

  // Get saved layout from localStorage
  const getDefaultLayout = () => {
    if (typeof window === 'undefined') return undefined;
    try {
      const saved = localStorage.getItem('playground-layout');
      if (saved) {
        return JSON.parse(saved);
      }
    } catch {
      // ignore
    }
    return undefined;
  };

  const handleLayoutChanged = (layout: Record<string, number>) => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('playground-layout', JSON.stringify(layout));
    }
  };

  const savedLayout = getDefaultLayout();

  return (
    <div className="playground-container">
      <ResizablePanelGroup
        orientation="horizontal"
        id="playground-layout"
        defaultLayout={savedLayout}
        onLayoutChanged={handleLayoutChanged}
      >
        <ResizablePanel id="editor" defaultSize={savedLayout?.editor ?? 60} minSize={30}>
          <div className="editor-panel h-full">
            <EditorTabs
              ref={editorTabsRef}
              defaultRslintConfig={DEFAULT_RSLINT_CONFIG}
              defaultTsconfig={DEFAULT_TSCONFIG}
              onCodeChange={setCode}
              onSelectionChange={(start, end) => handleAstNodeSelect(start, end)}
              onRslintConfigChange={setRslintConfig}
              onTsconfigChange={setTsconfigContent}
            />
          </div>
        </ResizablePanel>
        <ResizableHandle withHandle />
        <ResizablePanel id="result" defaultSize={savedLayout?.result ?? 40} minSize={20}>
          <ResultPanel
            initialized={initialized}
            diagnostics={diagnostics}
            ast={ast}
            astTree={astTree}
            tsAstTree={tsAstTree}
            error={error}
            loading={loading}
            onAstNodeSelect={(start, end) => {
              editorTabsRef.current?.getEditorRef()?.revealRangeByOffset(start, end);
              handleAstNodeSelect(start, end);
            }}
            selectedAstNodeRange={selectedAstRange}
            selectedNodeDetail={selectedNodeDetail}
            onRequestTsAst={async () => {
              tsAstActiveRef.current = true;
              if (!tsModuleRef.current) {
                try {
                  const mod = await import('typescript');
                  tsModuleRef.current = mod as any;
                } catch (e) {
                  console.warn('Failed to load TypeScript module:', e);
                  return;
                }
              }
              await buildTypeScriptAst(
                lastSourceTextRef.current || editorTabsRef.current?.getCodeValue() || '',
              );
            }}
          />
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
};

export default Playground;
