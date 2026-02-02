import React from 'react';
import type { SymbolInfo } from './types';
import { PropRow } from './PropRow';
import { FlagsDisplay } from './FlagsDisplay';
import { LazySymbolArray, LazyNodeView, LazyNodeArray } from './LazyViews';

interface SymbolInfoViewProps {
  info?: SymbolInfo;
}

export const SymbolInfoView: React.FC<SymbolInfoViewProps> = ({ info }) => {
  if (!info) {
    return (
      <div className="py-2 text-center text-xs text-gray-400">
        No symbol information
      </div>
    );
  }

  return (
    <div className="font-mono text-xs leading-relaxed">
      {info.id !== undefined && <PropRow label="Id" value={info.id} />}
      <PropRow label="Name" value={info.name} />
      {info.escapedName && info.escapedName !== info.name && (
        <PropRow label="EscapedName" value={info.escapedName} />
      )}
      <PropRow
        label="Flags"
        value={<FlagsDisplay flags={info.flags} names={info.flagNames} />}
      />
      <PropRow
        label="CheckFlags"
        value={
          <FlagsDisplay flags={info.checkFlags} names={info.checkFlagNames} />
        }
      />

      {info.valueDeclaration && (
        <LazyNodeView
          pos={info.valueDeclaration.pos}
          label="ValueDeclaration"
          kindName={info.valueDeclaration.kindName?.replace('ast.', '')}
          preview={[
            info.valueDeclaration.pos >= 0
              ? `pos: ${info.valueDeclaration.pos}, end: ${info.valueDeclaration.end}`
              : '',
            info.valueDeclaration.text
              ? `text: "${info.valueDeclaration.text}"`
              : '',
          ]
            .filter(Boolean)
            .join(', ')}
          shallow={info.valueDeclaration}
        />
      )}

      {info.declarations && info.declarations.length > 0 && (
        <LazyNodeArray label="Declarations" items={info.declarations} />
      )}

      {info.members && info.members.length > 0 && (
        <LazySymbolArray label="Members" items={info.members} />
      )}

      {info.exports && info.exports.length > 0 && (
        <LazySymbolArray label="Exports" items={info.exports} />
      )}
    </div>
  );
};

// Helper to get title for top-level display
export const getSymbolTitle = (info: SymbolInfo) => {
  return `Symbol(${info.name})`;
};

// Helper to get preview for top-level display
export const getSymbolPreview = (info: SymbolInfo) => {
  return (
    info.flagNames?.map(n => n.replace('ast.SymbolFlags', '')).join(' | ') || ''
  );
};
