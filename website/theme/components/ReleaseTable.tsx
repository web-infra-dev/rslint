import React from 'react';
import releases from '@/releases';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@components/ui/table';

export default function ReleaseTable() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Rslint</TableHead>
          <TableHead>TS release</TableHead>
          <TableHead>TS commit</TableHead>
          <TableHead>New rules</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {[...releases].reverse().map(({ version, rules, typescript }) => (
          <TableRow key={version}>
            <TableCell>
              {version === 'unreleased' ? (
                'Unreleased'
              ) : (
                <a
                  href={`https://github.com/web-infra-dev/rslint/releases/tag/v${version}`}
                >
                  {version}
                </a>
              )}
            </TableCell>
            <TableCell>
              {typescript?.releaseVersion ? (
                <a
                  href={`https://github.com/microsoft/TypeScript/releases/tag/v${typescript.releaseVersion}`}
                >
                  {typescript.releaseVersion}
                </a>
              ) : typescript ? (
                'No matching release'
              ) : (
                'Not recorded'
              )}
            </TableCell>
            <TableCell>
              {typescript ? (
                <a
                  href={`https://github.com/microsoft/TypeScript/commit/${typescript.commit}`}
                  title={typescript.commit}
                >
                  <code>{typescript.commit.slice(0, 12)}</code>
                </a>
              ) : (
                'Not recorded'
              )}
            </TableCell>
            <TableCell>
              {/^\d+\.\d+\.\d+$/.test(version) ? rules.length : '—'}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
