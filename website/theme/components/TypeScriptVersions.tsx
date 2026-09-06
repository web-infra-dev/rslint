import React from 'react';
import { GitBranchIcon, GitCommitHorizontalIcon } from 'lucide-react';
import releases from '@/releases';
import { Badge } from '@components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@components/ui/table';

export default function TypeScriptVersions() {
  return (
    <div className="overflow-hidden rounded-lg border border-gray-200 dark:border-gray-800">
      <Table className="!my-0 !rounded-none !border-0">
        <TableHeader className="bg-gray-50 dark:bg-gray-900">
          <TableRow>
            <TableHead>Rslint</TableHead>
            <TableHead>TypeScript release</TableHead>
            <TableHead>TypeScript commit</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {[...releases].reverse().map(({ version, typescript }) =>
            typescript ? (
              <TableRow key={version}>
                <TableCell>
                  {version === 'main' ? (
                    <a href="https://github.com/web-infra-dev/rslint/tree/main">
                      <Badge variant="unreleased" className="gap-1.5 py-0.5">
                        <GitBranchIcon className="size-3.5" aria-hidden />
                        main
                      </Badge>
                    </a>
                  ) : (
                    <Badge className="overflow-hidden p-0">
                      <span className="bg-gray-700 px-2 py-0.5 text-white">
                        npm
                      </span>
                      <span className="bg-[#CB3837] px-2 py-0.5 text-white">
                        v{version}
                      </span>
                    </Badge>
                  )}
                </TableCell>
                <TableCell>
                  {typescript.releaseVersion ? (
                    <a
                      href={`https://github.com/microsoft/TypeScript/releases/tag/v${typescript.releaseVersion}`}
                    >
                      <Badge variant="version">
                        v{typescript.releaseVersion}
                      </Badge>
                    </a>
                  ) : (
                    <span className="text-gray-500 dark:text-gray-400">
                      No release tag
                    </span>
                  )}
                </TableCell>
                <TableCell>
                  <a
                    href={`https://github.com/microsoft/TypeScript/commit/${typescript.commit}`}
                    title={typescript.commit}
                    className="inline-flex items-center gap-1.5"
                  >
                    <GitCommitHorizontalIcon
                      className="size-4 shrink-0"
                      aria-hidden
                    />
                    <code>{typescript.commit.slice(0, 12)}</code>
                  </a>
                </TableCell>
              </TableRow>
            ) : null,
          )}
        </TableBody>
      </Table>
    </div>
  );
}
