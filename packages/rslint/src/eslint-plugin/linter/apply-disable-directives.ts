import { ConfigCommentParser } from '@eslint/plugin-kit';
import type { Comment } from '../../languages/js/token-builder.js';
import type { Diagnostic } from './diagnostic-builder.js';

/** Original comment location, not a next-line directive's target location. */
export interface DisableDirectiveUsage {
  /** Zero-based ECMAScript line and UTF-16 column, against BOM-stripped text. */
  line: number;
  column: number;
  /** The declared rule entry; null identifies the wildcard entry. */
  ruleId: string | null;
}

interface Directive {
  type: 'disable' | 'enable';
  pos: number;
  ruleId: string | null;
  usage?: DisableDirectiveUsage;
}

interface Options {
  comments: readonly Comment[];
  diagnostics: readonly Diagnostic[];
  lineStartOffsets: readonly number[];
  textLength: number;
  collectUsage: boolean;
}

const parser = new ConfigCommentParser();

/** Filter before byte conversion and serialization; return each used entry once. */
export function applyDisableDirectives(options: Options): {
  diagnostics: readonly Diagnostic[];
  usedDisableDirectives?: DisableDirectiveUsage[];
} {
  const used = options.collectUsage
    ? new Set<DisableDirectiveUsage>()
    : undefined;
  if (options.diagnostics.length === 0) {
    return {
      diagnostics: options.diagnostics,
      ...(used ? { usedDisableDirectives: [] } : {}),
    };
  }
  const blocks: Directive[] = [];
  const lines: Directive[] = [];
  for (const comment of options.comments) {
    if (comment.type !== 'Line' && comment.type !== 'Block') continue;
    const parsed = parser.parseDirective(comment.value);
    if (!parsed) continue;
    const label = parsed.label.replace(/^rslint-/u, 'eslint-');
    if (
      label !== 'eslint-disable' &&
      label !== 'eslint-enable' &&
      label !== 'eslint-disable-line' &&
      label !== 'eslint-disable-next-line'
    ) {
      continue;
    }
    const block = label === 'eslint-disable' || label === 'eslint-enable';
    if (block && comment.type !== 'Block') continue;
    if (
      label === 'eslint-disable-line' &&
      comment.loc.start.line !== comment.loc.end.line
    ) {
      continue;
    }
    const names = parseRuleNames(parsed.value);
    for (const ruleId of names.length ? names : [null]) {
      const type = label === 'eslint-enable' ? 'enable' : 'disable';
      const usage =
        used && type === 'disable'
          ? {
              line: comment.loc.start.line - 1,
              column: comment.loc.start.column,
              ruleId,
            }
          : undefined;
      if (block) {
        blocks.push({ type, pos: comment.range[0], ruleId, usage });
      } else {
        const target =
          label === 'eslint-disable-line'
            ? comment.loc.start.line
            : comment.loc.end.line + 1;
        // A nonexistent following line must not clear a disable at EOF.
        lines.push({
          type: 'disable',
          pos: options.lineStartOffsets[target - 1] ?? Infinity,
          ruleId,
          usage,
        });
        lines.push({
          type: 'enable',
          pos: options.lineStartOffsets[target] ?? Infinity,
          ruleId,
        });
      }
    }
  }
  blocks.sort((a, b) => a.pos - b.pos);
  lines.sort((a, b) => a.pos - b.pos);
  // Block suppression gets usage attribution before line suppression. Stable
  // sorting preserves source ordering when line directives target the same line.
  const afterBlocks = filter(
    options.diagnostics,
    blocks,
    options.textLength,
    used,
  );
  const diagnostics = filter(afterBlocks, lines, options.textLength, used);
  return {
    diagnostics,
    ...(used ? { usedDisableDirectives: [...used] } : {}),
  };
}

function parseRuleNames(value: string): string[] {
  const names = new Set<string>();
  for (const part of value.split(',')) {
    let name = part.trim();
    if (
      name.length >= 2 &&
      (name[0] === '"' || name[0] === "'") &&
      name[name.length - 1] === name[0]
    ) {
      name = name.slice(1, -1);
    }
    if (name) names.add(name);
  }
  return [...names];
}

function filter(
  diagnostics: readonly Diagnostic[],
  directives: readonly Directive[],
  textLength: number,
  used: Set<DisableDirectiveUsage> | undefined,
): readonly Diagnostic[] {
  if (directives.length === 0) return diagnostics;
  const visible: Diagnostic[] = [];
  for (const diagnostic of diagnostics) {
    // Match the byte-boundary normalization applied to visible diagnostics.
    const pos = Number.isFinite(diagnostic.startPos)
      ? Math.min(textLength, Math.max(0, Math.trunc(diagnostic.startPos)))
      : 0;
    let active: Directive | undefined;
    for (const directive of directives) {
      if (directive.pos > pos) break;
      if (
        directive.ruleId === null ||
        directive.ruleId === diagnostic.ruleName
      ) {
        active = directive.type === 'disable' ? directive : undefined;
      }
    }
    if (active) {
      if (used && active.usage) used.add(active.usage);
    } else {
      visible.push(diagnostic);
    }
  }
  return visible;
}
