import assert from 'node:assert/strict';
import fs from 'node:fs';
import { test } from 'node:test';

test('the built release page renders the shared history as a table', () => {
  const html = fs.readFileSync(
    new URL('../doc_build/guide/releases.html', import.meta.url),
    'utf8',
  );
  const releases = JSON.parse(
    fs.readFileSync(new URL('../../releases.json', import.meta.url), 'utf8'),
  );
  const table = html.match(/<table\b[^>]*>([\s\S]*?)<\/table>/)?.[1];
  assert.ok(
    table,
    'The release component must render, not appear as literal Markdown',
  );
  const rows = [...table.matchAll(/<tr\b[^>]*>([\s\S]*?)<\/tr>/g)].map(
    (match) => match[1],
  );
  assert.equal(
    rows.length,
    releases.length + 1,
    'Every release needs a visible row',
  );
  for (const [index, entry] of [...releases].reverse().entries()) {
    const row = rows[index + 1];
    if (entry.typescript) {
      assert.ok(
        row.includes(
          `https://github.com/microsoft/TypeScript/commit/${entry.typescript.commit}`,
        ),
      );
      if (entry.typescript.releaseVersion) {
        assert.ok(
          row.includes(`/releases/tag/v${entry.typescript.releaseVersion}`),
        );
      } else {
        assert.ok(row.includes('No matching release'));
      }
    } else {
      assert.ok(
        row.includes('Not recorded'),
        'Historical releases must not invent a compiler binding',
      );
    }
  }
});
