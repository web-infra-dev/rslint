import React, { useCallback } from 'react';
import type { NodeInfo } from './types';
import { PropRow } from './PropRow';
import { FlagsDisplay } from './FlagsDisplay';
import { KindDisplay } from './KindDisplay';
import { LazyNodeView, LazyNodeArray, LazySymbolArray } from './LazyViews';
import { useAstInfoContextOptional } from './AstInfoContext';

interface NodeInfoViewProps {
  info?: NodeInfo;
  onNodeClick?: (pos: number, end?: number) => void;
}

export const NodeInfoView: React.FC<NodeInfoViewProps> = ({
  info,
  onNodeClick,
}) => {
  const context = useAstInfoContextOptional();

  const handleMouseEnter = useCallback(() => {
    // Don't highlight if node is from external file (positions would be wrong)
    if (
      context &&
      info?.pos !== undefined &&
      info?.end !== undefined &&
      !info.fileName
    ) {
      context.highlightRange(info.pos, info.end);
    }
  }, [context, info?.pos, info?.end, info?.fileName]);

  const handleMouseLeave = useCallback(() => {
    if (context) {
      context.clearHighlight();
    }
  }, [context]);

  if (!info) {
    return (
      <div className="py-2 text-center text-xs text-gray-400">
        No node information
      </div>
    );
  }

  // Helper to render a shallow node with lazy loading
  const renderShallowNode = (label: string, node: NodeInfo | undefined) => {
    if (!node) return null;
    const kindName = node.kindName?.replace('ast.', '');
    // Don't show pos/end in preview if they are -1 (external file nodes)
    const posEndPreview =
      node.pos >= 0 ? `pos: ${node.pos}, end: ${node.end}` : '';
    const textPreview = node.text ? `text: "${node.text}"` : '';
    const preview = [posEndPreview, textPreview].filter(Boolean).join(', ');
    return (
      <LazyNodeView
        pos={node.pos}
        label={label}
        kindName={kindName}
        preview={preview || kindName || ''}
        shallow={node}
      />
    );
  };

  return (
    <div className="font-mono text-xs leading-relaxed">
      {info.id !== undefined && <PropRow label="Id" value={info.id} />}
      <PropRow
        label="Kind"
        value={
          <span
            className="underline decoration-dotted hover:decoration-solid cursor-default"
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
          >
            <KindDisplay kind={info.kind} kindName={info.kindName} />
          </span>
        }
      />
      <PropRow
        label="Pos"
        value={
          // Only allow navigation for nodes in current file (no fileName)
          onNodeClick && !info.fileName ? (
            <span
              className="cursor-pointer hover:underline text-blue-600"
              onClick={() => onNodeClick(info.pos, info.end)}
            >
              {info.pos}
            </span>
          ) : (
            info.pos
          )
        }
      />
      <PropRow label="End" value={info.end} />
      {info.fileName && <PropRow label="FileName" value={info.fileName} />}
      <PropRow
        label="Flags"
        value={<FlagsDisplay flags={info.flags} names={info.flagNames} />}
      />
      {info.text && <PropRow label="Text" value={`"${info.text}"`} />}

      {/* Node properties - lazy loaded */}
      {renderShallowNode('Parent', info.parent)}
      {renderShallowNode('Name', info.name)}
      {renderShallowNode('Expression', info.expression)}
      {renderShallowNode('Left', info.left)}
      {renderShallowNode('Right', info.right)}
      {renderShallowNode('OperatorToken', info.operatorToken)}
      {renderShallowNode('Operand', info.operand)}
      {renderShallowNode('Condition', info.condition)}
      {renderShallowNode('WhenTrue', info.whenTrue)}
      {renderShallowNode('WhenFalse', info.whenFalse)}
      {renderShallowNode('ThenStatement', info.thenStatement)}
      {renderShallowNode('ElseStatement', info.elseStatement)}
      {renderShallowNode('Body', info.body)}
      {renderShallowNode('Initializer', info.initializer)}
      {renderShallowNode('Type', info.type)}

      {/* Variable/Declaration properties */}
      {renderShallowNode('DeclarationList', info.declarationList)}

      {/* Import/Export properties */}
      {renderShallowNode('ImportClause', info.importClause)}
      {renderShallowNode('ModuleSpecifier', info.moduleSpecifier)}
      {renderShallowNode('NamedBindings', info.namedBindings)}
      {renderShallowNode('ExportClause', info.exportClause)}

      {/* Loop/Control flow properties */}
      {renderShallowNode('Incrementor', info.incrementor)}
      {renderShallowNode('Statement', info.statement)}

      {/* Switch statement properties */}
      {renderShallowNode('CaseBlock', info.caseBlock)}

      {/* Try/Catch properties */}
      {renderShallowNode('TryBlock', info.tryBlock)}
      {renderShallowNode('CatchClause', info.catchClause)}
      {renderShallowNode('FinallyBlock', info.finallyBlock)}
      {renderShallowNode('VariableDeclaration', info.variableDeclaration)}
      {renderShallowNode('Block', info.block)}

      {/* Property access properties */}
      {renderShallowNode('ArgumentExpression', info.argumentExpression)}

      {/* Shorthand property assignment properties */}
      {renderShallowNode('EqualsToken', info.equalsToken)}
      {renderShallowNode(
        'ObjectAssignmentInitializer',
        info.objectAssignmentInitializer,
      )}

      {/* Template literal properties */}
      {renderShallowNode('Head', info.head)}
      {renderShallowNode('Literal', info.literal)}
      {renderShallowNode('Tag', info.tag)}
      {renderShallowNode('Template', info.template)}

      {/* Token properties */}
      {renderShallowNode('QuestionToken', info.questionToken)}
      {renderShallowNode('DotDotDotToken', info.dotDotDotToken)}
      {renderShallowNode('ExclamationToken', info.exclamationToken)}
      {renderShallowNode('AsteriskToken', info.asteriskToken)}
      {renderShallowNode('EqualsGreaterThanToken', info.equalsGreaterThanToken)}
      {renderShallowNode('QuestionDotToken', info.questionDotToken)}

      {/* Type-related properties */}
      {renderShallowNode('Constraint', info.constraint)}
      {renderShallowNode('DefaultType', info.defaultType)}

      {/* Array properties with list metadata */}
      {info.modifiers && info.modifiers.length > 0 && (
        <LazyNodeArray
          label="Modifiers"
          items={info.modifiers}
          listMeta={info.listMetas?.Modifiers}
        />
      )}
      {info.typeParameters && info.typeParameters.length > 0 && (
        <LazyNodeArray
          label="TypeParameters"
          items={info.typeParameters}
          listMeta={info.listMetas?.TypeParameters}
        />
      )}
      {info.heritageClauses && info.heritageClauses.length > 0 && (
        <LazyNodeArray
          label="HeritageClauses"
          items={info.heritageClauses}
          listMeta={info.listMetas?.HeritageClauses}
        />
      )}
      {info.parameters && info.parameters.length > 0 && (
        <LazyNodeArray
          label="Parameters"
          items={info.parameters}
          listMeta={info.listMetas?.Parameters}
        />
      )}
      {info.members && info.members.length > 0 && (
        <LazyNodeArray
          label="Members"
          items={info.members}
          listMeta={info.listMetas?.Members}
        />
      )}
      {info.statements && info.statements.length > 0 && (
        <LazyNodeArray
          label="Statements"
          items={info.statements}
          listMeta={info.listMetas?.Statements}
        />
      )}
      {info.arguments && info.arguments.length > 0 && (
        <LazyNodeArray
          label="Arguments"
          items={info.arguments}
          listMeta={info.listMetas?.Arguments}
        />
      )}
      {info.properties && info.properties.length > 0 && (
        <LazyNodeArray
          label="Properties"
          items={info.properties}
          listMeta={info.listMetas?.Properties}
        />
      )}
      {info.elements && info.elements.length > 0 && (
        <LazyNodeArray
          label="Elements"
          items={info.elements}
          listMeta={info.listMetas?.Elements}
        />
      )}
      {info.declarations && info.declarations.length > 0 && (
        <LazyNodeArray
          label="Declarations"
          items={info.declarations}
          listMeta={info.listMetas?.Declarations}
        />
      )}
      {info.clauses && info.clauses.length > 0 && (
        <LazyNodeArray
          label="Clauses"
          items={info.clauses}
          listMeta={info.listMetas?.Clauses}
        />
      )}
      {info.templateSpans && info.templateSpans.length > 0 && (
        <LazyNodeArray
          label="TemplateSpans"
          items={info.templateSpans}
          listMeta={info.listMetas?.TemplateSpans}
        />
      )}
      {info.typeArguments && info.typeArguments.length > 0 && (
        <LazyNodeArray
          label="TypeArguments"
          items={info.typeArguments}
          listMeta={info.listMetas?.TypeArguments}
        />
      )}

      {/* Locals - symbols declared in this node's scope */}
      {info.locals && info.locals.length > 0 && (
        <LazySymbolArray label="Locals()" items={info.locals} />
      )}

      {/* SourceFile-specific properties */}
      {renderShallowNode('EndOfFileToken', info.endOfFileToken)}
      {info.imports && info.imports.length > 0 && (
        <LazyNodeArray label="Imports()" items={info.imports} />
      )}
      {info.isDeclarationFile !== undefined && (
        <PropRow
          label="IsDeclarationFile"
          value={info.isDeclarationFile ? 'true' : 'false'}
        />
      )}
      {info.scriptKind !== undefined && info.scriptKind > 0 && (
        <PropRow
          label="ScriptKind"
          value={
            [
              'Unknown',
              'JS',
              'JSX',
              'TS',
              'TSX',
              'External',
              'JSON',
              'Deferred',
            ][info.scriptKind] || info.scriptKind
          }
        />
      )}
      {info.identifierCount !== undefined && info.identifierCount > 0 && (
        <PropRow label="IdentifierCount" value={info.identifierCount} />
      )}
      {info.symbolCount !== undefined && info.symbolCount > 0 && (
        <PropRow label="SymbolCount" value={info.symbolCount} />
      )}
      {info.nodeCount !== undefined && info.nodeCount > 0 && (
        <PropRow label="NodeCount" value={info.nodeCount} />
      )}
    </div>
  );
};

// Helper to get title for top-level display
export const getNodeTitle = (info: NodeInfo) => {
  return info.kindName?.replace('ast.', '') || 'Node';
};

// Helper to get preview for top-level display
export const getNodePreview = (info: NodeInfo) => {
  const parts = [];
  if (info.pos >= 0) parts.push(`pos: ${info.pos}`);
  if (info.text) parts.push(`"${info.text}"`);
  return parts.join(', ');
};
