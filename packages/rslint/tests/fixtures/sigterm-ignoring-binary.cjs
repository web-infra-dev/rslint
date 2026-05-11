#!/usr/bin/env node
'use strict';

/**
 * Test fixture: deliberately ignores SIGTERM so we can verify the
 * SIGKILL fallback in engine.ts's safeKillGo actually escalates. Stays
 * alive on stdin closed too, until the kernel delivers SIGKILL.
 */

process.on('SIGTERM', () => {
  // No-op. The whole point of this fixture is to demonstrate that a
  // misbehaving binary that ignores SIGTERM is still terminated.
  process.stderr.write('[fixture] received SIGTERM, ignoring\n');
});

process.on('SIGINT', () => {
  process.stderr.write('[fixture] received SIGINT, ignoring\n');
});

// Mark "alive" so the test can confirm we actually got past startup.
process.stderr.write('__FIXTURE_READY__\n');

// Park forever. setInterval keeps the event loop alive.
setInterval(() => {
  // tick — no-op
}, 60_000);
