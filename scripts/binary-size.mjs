#!/usr/bin/env node
/**
 * Binary size monitoring for the `rslint` executable.
 *
 *   node scripts/binary-size.mjs record <binary> <out.json>
 *   node scripts/binary-size.mjs base-run <head.json>       # -> $GITHUB_OUTPUT
 *   node scripts/binary-size.mjs report <head.json> [base.json]
 *   node scripts/binary-size.mjs comment <report.md>
 *
 * The measurement itself is produced inside the existing `Test npm packages`
 * job, which has already compiled every package behind ./cmd/rslint — so the
 * release relink it measures is a link-only action off a warm Go build cache
 * rather than a second build. Each run stores its number as a workflow
 * artifact; a pull request reads back the artifact of the run that measured
 * the commit its build was merged onto. That keeps the whole system on the
 * built-in GITHUB_TOKEN: no data branch, no personal access token, no
 * repository writes.
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
const MAX_SKIPPED_BASE_COMMITS = 5;

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
 * The commit the measured binary was built on top of.
 *
 * CI compiles the merge commit GitHub generates for a pull request, so what
 * gets measured is `base branch at build time + this branch`. The first parent
 * of that merge commit is exactly that base, and subtracting it isolates the
 * pull request's own contribution.
 *
 * The event's `base.sha` cannot do that job: it is a snapshot stored on the
 * pull request and is not refreshed when the base branch moves on, while the
 * merge commit is rebuilt against the current tip for every run. Once the two
 * drift apart, a comparison against `base.sha` also counts everything that
 * landed on the base branch in between and bills it to the pull request.
 */
async function resolveBaseCommit(headPath) {
  const measuredSha = measuredCommitSha(headPath);
  if (measuredSha) {
    try {
      const commit = await api(`/repos/${repository}/commits/${measuredSha}`);
      const parents = commit.parents || [];
      if (parents.length > 1) {
        const sha = parents[0].sha;
        return { sha, subject: await commitSubject(sha) };
      }
    } catch {
      // Fall through: an unreachable commit is no reason to skip the report.
    }
  }

  // Not a merge build — the run measured a commit that stands on its own, so
  // the pull request's recorded base is the best answer available.
  const fallback = readEvent().pull_request?.base?.sha || '';
  if (!fallback) return { sha: '', subject: '' };
  return { sha: fallback, subject: await commitSubject(fallback) };
}

/** The commit this run measured, as the measurement itself recorded it. */
function measuredCommitSha(headPath) {
  if (headPath) {
    try {
      const measurement = JSON.parse(fs.readFileSync(headPath, 'utf8'));
      if (measurement.sha) return measurement.sha;
    } catch {
      // Missing or unreadable: fall back to this job's own checkout.
    }
  }
  return process.env.GITHUB_SHA || '';
}

/**
 * Locate the run that measured the base commit, walking first parents only
 * when the main CI deliberately skipped the Ubuntu measurement job.
 *
 * A run that measured no artifact, failed, or has not finished stops the
 * search. Such a missing measurement does not imply an unchanged binary.
 */
async function findBaseRun(headPath) {
  const base = await resolveBaseCommit(headPath);
  if (!base.sha) return { baseSha: '', runId: '' };

  writeBaseCommit(base);
  let sha = base.sha;
  const skippedShas = [];
  while (skippedShas.length <= MAX_SKIPPED_BASE_COMMITS) {
    const result = await mainMeasurement(sha);
    if (result.status === 'measured') {
      writeBaseCommit({
        ...base,
        measuredSha: sha,
        skippedShas: [...skippedShas].reverse(),
      });
      return { baseSha: sha, runId: result.runId };
    }
    if (
      result.status !== 'skipped' ||
      skippedShas.length === MAX_SKIPPED_BASE_COMMITS
    ) {
      break;
    }
    skippedShas.push(sha);
    const commit = await api(`/repos/${repository}/commits/${sha}`);
    sha = commit.parents?.[0]?.sha || '';
    if (!sha) break;
  }
  return { baseSha: base.sha, runId: '' };
}

/** Inspect the newest main push run for this commit, without skipping a failed measurement. */
async function mainMeasurement(sha) {
  const runs = await api(
    `/repos/${repository}/actions/runs?head_sha=${sha}&event=push&per_page=100`,
  );
  const run = (runs.workflow_runs || [])
    .filter(
      (item) =>
        item.path === '.github/workflows/ci.yml' &&
        item.head_branch === 'main' &&
        item.head_sha === sha &&
        item.event === 'push',
    )
    .sort((a, b) => b.run_number - a.run_number)[0];
  if (!run || run.status !== 'completed') return { status: 'unavailable' };

  const jobs = await api(
    `/repos/${repository}/actions/runs/${run.id}/jobs?per_page=100`,
  );
  const changed = (jobs.jobs || []).find(
    (job) => job.name === 'Detect changes',
  );
  const ubuntu = (jobs.jobs || []).find(
    (job) =>
      job.name === 'Test npm packages' ||
      job.name.startsWith('Test npm packages (rspack-ubuntu-'),
  );
  if (ubuntu?.conclusion === 'skipped' && changed?.conclusion === 'success') {
    return { status: 'skipped' };
  }
  if (ubuntu?.conclusion !== 'success') return { status: 'unavailable' };

  const artifacts = await api(
    `/repos/${repository}/actions/runs/${run.id}/artifacts?per_page=100`,
  );
  const artifact = (artifacts.artifacts || []).find(
    (item) => item.name === 'rslint-binary-size' && !item.expired,
  );
  return artifact
    ? { status: 'measured', runId: String(run.id) }
    : { status: 'unavailable' };
}

/** A commit's subject line, or an empty string when it cannot be read. */
async function commitSubject(sha) {
  try {
    const commit = await api(`/repos/${repository}/commits/${sha}`);
    return (commit.commit?.message || '').split('\n')[0];
  } catch {
    return '';
  }
}

/**
 * Record the base commit's identity next to its measurement.
 *
 * Written whether or not a measurement for it turned up, so the report can
 * always name the commit the comparison is against.
 */
function writeBaseCommit({ sha, subject, measuredSha, skippedShas }) {
  fs.writeFileSync(
    BASE_COMMIT_FILE,
    `${JSON.stringify({ sha, subject, measuredSha, skippedShas }, null, 2)}\n`,
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
  const downloadedBase =
    basePath && fs.existsSync(basePath)
      ? JSON.parse(fs.readFileSync(basePath, 'utf8'))
      : undefined;
  const baseOn = baseCommit();
  const base =
    downloadedBase?.sha === (baseOn.measuredSha || baseOn.sha)
      ? downloadedBase
      : undefined;
  const skippedShas =
    base && Array.isArray(baseOn.skippedShas) ? baseOn.skippedShas : [];
  const skippedNote = skippedShas.length
    ? `Base size was measured at main commit ${commitLink(baseOn.measuredSha)}. The Ubuntu measurement job was skipped for the following ${skippedShas.length === 1 ? 'main commit' : `${skippedShas.length} main commits`}: ${skippedShas.map(commitLink).join(', ')}.`
    : '';

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
    ...(skippedNote ? ['', skippedNote] : []),
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
      const [headPath] = args;
      const { baseSha, runId } = await findBaseRun(headPath);
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
