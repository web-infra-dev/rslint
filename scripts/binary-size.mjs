#!/usr/bin/env node
/**
 * Binary size monitoring for the `rslint` executable.
 *
 *   node scripts/binary-size.mjs record <binary> <out.json>
 *   node scripts/binary-size.mjs base-run                    # -> $GITHUB_OUTPUT
 *   node scripts/binary-size.mjs report <head.json> [base.json]
 *   node scripts/binary-size.mjs comment <report.md>
 *
 * The measurement itself is produced inside the existing `Test npm packages`
 * job, which has already compiled every package behind ./cmd/rslint — so the
 * release relink it measures is a link-only action off a warm Go build cache
 * rather than a second build. Each run stores its number as a workflow
 * artifact; a pull request reads back the artifact of the main run for its
 * base commit. That keeps the whole system on the built-in GITHUB_TOKEN: no
 * data branch, no personal access token, no repository writes.
 */
import fs from 'node:fs';
import path from 'node:path';

const SCHEMA_VERSION = 1;

// Ties an update to the comment this workflow posted last time, instead of
// opening a new thread on every push.
const COMMENT_MARKER = '<!-- rslint-binary-size -->';

// Written by `base-run`, read by `report`, so the base commit can be named
// whether or not a measurement for it turned up.
const BASE_COMMIT_FILE = 'base-commit.json';

const apiBase = process.env.GITHUB_API_URL || 'https://api.github.com';
const repository = process.env.GITHUB_REPOSITORY || '';

async function api(endpoint) {
  const response = await fetch(`${apiBase}${endpoint}`, {
    headers: githubHeaders(),
  });
  if (!response.ok) {
    throw new Error(
      `GET ${endpoint} failed: ${response.status} ${response.statusText}`,
    );
  }
  return response.json();
}

/** The `rel="next"` target of an API `Link` header, if there is another page. */
function nextPageUrl(link) {
  if (!link) return '';
  for (const part of link.split(',')) {
    const match = /<([^>]+)>;\s*rel="next"/.exec(part);
    if (match) return match[1];
  }
  return '';
}

/**
 * Walk a list endpoint page by page.
 *
 * Reading only the first page would miss the marker comment once a pull
 * request has more than a page of comments, and every later push would then
 * post a fresh report instead of editing the existing one.
 */
async function* apiPages(endpoint) {
  let url = `${apiBase}${endpoint}`;
  while (url) {
    const response = await fetch(url, { headers: githubHeaders() });
    if (!response.ok) {
      throw new Error(
        `GET ${url} failed: ${response.status} ${response.statusText}`,
      );
    }
    yield await response.json();
    url = nextPageUrl(response.headers.get('link'));
  }
}

function githubHeaders() {
  const token = process.env.GITHUB_TOKEN;
  if (!token) throw new Error('GITHUB_TOKEN is not set');
  return {
    accept: 'application/vnd.github+json',
    authorization: `Bearer ${token}`,
    'content-type': 'application/json',
    'x-github-api-version': '2022-11-28',
  };
}

function readEvent() {
  const eventPath = process.env.GITHUB_EVENT_PATH;
  if (!eventPath) throw new Error('GITHUB_EVENT_PATH is not set');
  return JSON.parse(fs.readFileSync(eventPath, 'utf8'));
}

/**
 * The commit a reader can actually find in the pull request.
 *
 * On a `pull_request` event GITHUB_SHA is the ephemeral merge commit GitHub
 * built for CI; it appears nowhere in the pull request's commit list. Since
 * the report is edited in place across pushes, and a build takes minutes, the
 * comment has to name the head commit it measured so a stale report is
 * obvious.
 */
function headCommitSha() {
  try {
    return readEvent().pull_request?.head?.sha || process.env.GITHUB_SHA || '';
  } catch {
    return process.env.GITHUB_SHA || '';
  }
}

/** Measure `binary` and write the record this run will publish as an artifact. */
function record(binary, out) {
  const size = fs.statSync(binary).size;
  const measurement = {
    schemaVersion: SCHEMA_VERSION,
    sha: process.env.GITHUB_SHA || '',
    headSha: headCommitSha(),
    ref: process.env.GITHUB_REF_NAME || '',
    event: process.env.GITHUB_EVENT_NAME || '',
    runId: process.env.GITHUB_RUN_ID || '',
    target: 'linux-x64-gnu',
    goVersion: process.env.BINARY_SIZE_GO_VERSION || '',
    buildFlags: '-ldflags=-s -w',
    size,
    measuredAt: new Date().toISOString(),
  };
  fs.mkdirSync(path.dirname(path.resolve(out)), { recursive: true });
  fs.writeFileSync(out, `${JSON.stringify(measurement, null, 2)}\n`);
  return measurement;
}

/**
 * Locate the workflow run that measured the pull request's base commit.
 *
 * The base commit is on main, so its own CI run holds its measurement.
 * A miss is not an error: the base may predate this workflow, its run may
 * still be going, or its artifact may have expired. The report then just
 * states the current size.
 */
async function findBaseRun() {
  const baseSha = readEvent().pull_request?.base?.sha;
  if (!baseSha) return { baseSha: '', runId: '' };

  await writeBaseCommit(baseSha);

  const runs = await api(
    `/repos/${repository}/actions/runs?head_sha=${baseSha}&per_page=100`,
  );
  const candidates = (runs.workflow_runs || [])
    .filter((run) => run.path === '.github/workflows/ci.yml')
    .sort((a, b) => b.run_number - a.run_number);

  for (const run of candidates) {
    const artifacts = await api(
      `/repos/${repository}/actions/runs/${run.id}/artifacts?per_page=100`,
    );
    const artifact = (artifacts.artifacts || []).find(
      (item) => item.name === 'rslint-binary-size' && !item.expired,
    );
    if (artifact) return { baseSha, runId: String(run.id) };
  }

  return { baseSha, runId: '' };
}

/**
 * Record the base commit's identity next to its measurement.
 *
 * Looked up rather than read off the base measurement so the report can name
 * the base commit even when no measurement for it exists.
 */
async function writeBaseCommit(sha) {
  let subject;
  try {
    const commit = await api(`/repos/${repository}/commits/${sha}`);
    subject = (commit.commit?.message || '').split('\n')[0];
  } catch {
    subject = '';
  }
  fs.writeFileSync(
    BASE_COMMIT_FILE,
    `${JSON.stringify({ sha, subject }, null, 2)}\n`,
  );
}

function formatBytes(bytes) {
  const mib = bytes / 1024 / 1024;
  if (Math.abs(mib) >= 1) return `${mib.toFixed(2)} MiB`;
  return `${(bytes / 1024).toFixed(2)} KiB`;
}

function formatDelta(delta) {
  const sign = delta > 0 ? '+' : delta < 0 ? '-' : '';
  return `${sign}${formatBytes(Math.abs(delta))}`;
}

function formatPercent(delta, base) {
  if (base === 0) return 'n/a';
  const percent = (delta / base) * 100;
  const sign = percent > 0 ? '+' : percent < 0 ? '-' : '';
  return `${sign}${Math.abs(percent).toFixed(2)}%`;
}

/** Link a commit by its short SHA, so a reader can open exactly what was measured. */
function commitLink(sha) {
  if (!sha) return '`unknown`';
  const short = `\`${sha.slice(0, 7)}\``;
  const server = process.env.GITHUB_SERVER_URL;
  if (!server || !repository) return short;
  return `[${short}](${server}/${repository}/commit/${sha})`;
}

/**
 * A commit subject as an inline code span.
 *
 * A code span rather than plain text on purpose: it keeps a `#1234` in the
 * subject from auto-linking, which would post a cross-reference onto whatever
 * unrelated pull request that number belongs to every time this comment is
 * edited. Backticks are dropped so the span cannot be broken out of.
 */
function formatSubject(subject) {
  if (!subject) return '';
  const clean = subject.replaceAll('`', '');
  const trimmed = clean.length > 72 ? `${clean.slice(0, 71)}…` : clean;
  return trimmed ? ` — \`${trimmed}\`` : '';
}

/** "2026-09-07 03:22 UTC" — minute precision is enough to spot a stale report. */
function formatTimestamp(iso) {
  if (!iso) return '';
  return `${iso.slice(0, 10)} ${iso.slice(11, 16)} UTC`;
}

/** The base commit this pull request is measured against. */
function baseCommit() {
  try {
    return JSON.parse(fs.readFileSync(BASE_COMMIT_FILE, 'utf8'));
  } catch {
    return { sha: readEvent().pull_request?.base?.sha || '', subject: '' };
  }
}

/** Render the Markdown shared by the job summary and the pull request comment. */
function report(headPath, basePath) {
  const head = JSON.parse(fs.readFileSync(headPath, 'utf8'));
  const base =
    basePath && fs.existsSync(basePath)
      ? JSON.parse(fs.readFileSync(basePath, 'utf8'))
      : undefined;
  const baseOn = baseCommit();

  // One table either way: an em dash where a missing measurement would go says
  // everything a sentence about it would.
  const delta = base ? head.size - base.size : undefined;
  const row = [
    `\`rslint\` (${head.target})`,
    base ? formatBytes(base.size) : '—',
    formatBytes(head.size),
    delta === undefined
      ? '—'
      : `${formatDelta(delta)} (${formatPercent(delta, base.size)})`,
  ];

  const goVersion = head.goVersion ? `Go ${head.goVersion}, ` : '';
  const server = process.env.GITHUB_SERVER_URL;
  const run =
    head.runId && server && repository
      ? `[run](${server}/${repository}/actions/runs/${head.runId})`
      : 'run';

  return [
    COMMENT_MARKER,
    '',
    '## 🦀📦 Binary size',
    '',
    `Commit ${commitLink(head.headSha)} merged into base ${commitLink(baseOn.sha)}${formatSubject(baseOn.subject)}.`,
    '',
    '| Binary | Base | This PR | Change |',
    '| --- | ---: | ---: | ---: |',
    `| ${row.join(' | ')} |`,
    '',
    `<sub>Stripped \`go build -ldflags="-s -w" ./cmd/rslint\`, ${goVersion}linux/amd64 · ${run} · ${formatTimestamp(head.measuredAt)}</sub>`,
    '',
  ].join('\n');
}

/**
 * Post or update the report on the pull request.
 *
 * Best-effort by design: a pull request from a fork gets a read-only
 * GITHUB_TOKEN, so this call fails there and the job summary stays the only
 * channel. That is why the caller runs this step with `continue-on-error`.
 */
async function comment(reportPath) {
  const body = fs.readFileSync(reportPath, 'utf8');
  const issueNumber = readEvent().pull_request?.number;
  if (!issueNumber) throw new Error('no pull request number in the event');

  let previous;
  for await (const page of apiPages(
    `/repos/${repository}/issues/${issueNumber}/comments?per_page=100`,
  )) {
    previous = page.find((item) => item.body?.includes(COMMENT_MARKER));
    if (previous) break;
  }

  const endpoint = previous
    ? `/repos/${repository}/issues/comments/${previous.id}`
    : `/repos/${repository}/issues/${issueNumber}/comments`;
  const response = await fetch(`${apiBase}${endpoint}`, {
    method: previous ? 'PATCH' : 'POST',
    headers: githubHeaders(),
    body: JSON.stringify({ body }),
  });
  if (!response.ok) {
    throw new Error(
      `${previous ? 'PATCH' : 'POST'} ${endpoint} failed: ${response.status} ${response.statusText}`,
    );
  }
}

function appendOutput(key, value) {
  const outputPath = process.env.GITHUB_OUTPUT;
  if (!outputPath) return;
  fs.appendFileSync(outputPath, `${key}=${value}\n`);
}

async function main() {
  const [command, ...args] = process.argv.slice(2);

  switch (command) {
    case 'record': {
      const [binary, out] = args;
      if (!binary || !out) throw new Error('usage: record <binary> <out.json>');
      const measurement = record(binary, out);
      process.stdout.write(
        `${measurement.target} rslint: ${measurement.size} bytes (${formatBytes(measurement.size)})\n`,
      );
      break;
    }
    case 'base-run': {
      const { baseSha, runId } = await findBaseRun();
      appendOutput('run-id', runId);
      process.stdout.write(
        runId
          ? `run measuring ${baseSha}: ${runId}\n`
          : `no measurement found for ${baseSha || 'the base commit'}\n`,
      );
      break;
    }
    case 'report': {
      const [headPath, basePath] = args;
      if (!headPath) throw new Error('usage: report <head.json> [base.json]');
      process.stdout.write(report(headPath, basePath));
      break;
    }
    case 'comment': {
      const [reportPath] = args;
      if (!reportPath) throw new Error('usage: comment <report.md>');
      await comment(reportPath);
      break;
    }
    default:
      throw new Error(`unknown command: ${command ?? '(none)'}`);
  }
}

await main();
