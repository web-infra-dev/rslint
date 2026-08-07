import { useEffect, useRef } from 'react';
import './keyframes.scss';
import styles from './styles.module.scss';

/**
 * Live lint-scan demo pane — ported from the d-glitch-decode prototype's
 * `.decode-pane`. Shows a fake but faithful rslint run: a scanner band
 * sweeps a TS snippet decoding each char, a CLI-style console types the
 * `$ rslint ./demo.ts` command, a spinner+bar progress line fills, two
 * diagnostics (no-explicit-any + eqeqeq) render in the real rslint default
 * printer format, and VS Code MarkerHover tooltips appear on squiggle
 * hover. Clicking "Quick Fix..." on the eqeqeq tooltip rewrites the source
 * (`==` -> `===`) and reruns the cycle with one fewer diagnostic; the
 * "rerun" button replays the full 2-violation cycle.
 *
 * The original was vanilla DOM + rAF. We keep that imperative model: the
 * static shell is rendered as JSX, and the whole animation engine runs in
 * a single useEffect against a ref-scoped container (getElementById ->
 * root.querySelector). All rAF / setInterval / listeners are torn down on
 * unmount. No theme branching is needed here — every colour comes from CSS
 * vars that flip on rspress's `html.dark` class (see styles.module.scss).
 */
export function LintDemo() {
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const root = rootRef.current;
    if (!root) {
      return;
    }
    const $ = <T extends HTMLElement = HTMLElement>(sel: string) =>
      root.querySelector(sel) as T | null;

    // ---------- source code shown in the pane ----------
    const ORIGINAL_SOURCE = `export function hello(name: any) {
  if (name == 'Rslint') {
    return 'Welcome back!';
  }
  return \`Hello, \${name}\`;
}`;
    let SOURCE = ORIGINAL_SOURCE;

    function classify(src: string): (string | null)[] {
      const out: (string | null)[] = new Array(src.length).fill(null);
      const tag = (s: number, e: number, k: string) => {
        for (let i = s; i < e; i++) {
          out[i] = k;
        }
      };
      let re = /'[^']*'|"[^"]*"/g;
      let m: RegExpExecArray | null;
      while ((m = re.exec(src))) {
        tag(m.index, m.index + m[0].length, 's');
      }
      re = /\b\d+\b/g;
      while ((m = re.exec(src))) {
        tag(m.index, m.index + m[0].length, 'n');
      }
      re = /\/\/[^\n]*/g;
      while ((m = re.exec(src))) {
        tag(m.index, m.index + m[0].length, 'c');
      }
      re =
        /\b(import|from|export|default|const|let|function|return|if|else|new|await|async|true|false|null|undefined|while)\b/g;
      while ((m = re.exec(src))) {
        if (out[m.index] === null) {
          tag(m.index, m.index + m[0].length, 'k');
        }
      }
      re = /\b([A-Za-z_$][\w$]*)\s*\(/g;
      while ((m = re.exec(src))) {
        if (out[m.index] === null) {
          tag(m.index, m.index + m[1].length, 'p');
        }
      }
      return out;
    }
    let KINDS = classify(SOURCE);

    // ---------- build char nodes ----------
    const codeEl = $('#ld-code');
    const decodeBody = $('#ld-decodeBody');
    if (!codeEl || !decodeBody) {
      return;
    }
    let lines = SOURCE.split('\n');
    type CharNode = {
      node: HTMLSpanElement;
      char: string;
      y: number;
      kind: string | null;
      state: string;
    };
    const charNodes: CharNode[] = [];
    const LINE_H = 22;
    const TOP_PAD = 14;

    const lineCharSpans: HTMLSpanElement[][] = [];

    function buildCharNodes() {
      codeEl!.innerHTML = '';
      lines = SOURCE.split('\n');
      KINDS = classify(SOURCE);
      charNodes.length = 0;
      lineCharSpans.length = 0;
      let abs = 0;
      lines.forEach((line, ly) => {
        const lineDiv = document.createElement('div');
        lineDiv.className = 'cl';
        const lineArr: HTMLSpanElement[] = [];
        if (line.length === 0) {
          lineDiv.innerHTML = '&nbsp;';
        } else {
          for (let cx = 0; cx < line.length; cx++) {
            const s = document.createElement('span');
            s.className = 'ch enc';
            s.textContent = line[cx];
            lineDiv.appendChild(s);
            lineArr.push(s);
            charNodes.push({
              node: s,
              char: line[cx],
              y: ly * LINE_H + LINE_H / 2 + TOP_PAD,
              kind: KINDS[abs],
              state: 'enc',
            });
            abs++;
          }
        }
        lineCharSpans.push(lineArr);
        codeEl!.appendChild(lineDiv);
        abs++;
      });
    }
    buildCharNodes();

    // ---------- violations ----------
    type Violation = {
      line: number;
      col: number;
      len: number;
      severity: 'error' | 'warning';
      rule: string;
      message: string;
      quickFix: boolean;
      fix: { replace: string } | null;
      codeLines: { n: number; text: string; hl: [number, number] | null }[];
    };
    const ORIGINAL_VIOLATIONS: Violation[] = [
      {
        line: 1,
        col: 29,
        len: 3,
        severity: 'error',
        rule: '@typescript-eslint/no-explicit-any',
        message: 'Unexpected any. Specify a different type.',
        quickFix: false,
        fix: null,
        codeLines: [
          { n: 1, text: `export function hello(name: any) {`, hl: [28, 31] },
          { n: 2, text: `  if (name == 'Rslint') {`, hl: null },
        ],
      },
      {
        line: 2,
        col: 12,
        len: 2,
        severity: 'error',
        rule: 'eqeqeq',
        message: "Expected '===' and instead saw '=='.",
        quickFix: true,
        fix: { replace: '===' },
        codeLines: [
          { n: 1, text: `export function hello(name: any) {`, hl: null },
          { n: 2, text: `  if (name == 'Rslint') {`, hl: [11, 13] },
          { n: 3, text: `    return 'Welcome back!';`, hl: null },
        ],
      },
    ];
    let VIOLATIONS: Violation[] = ORIGINAL_VIOLATIONS.slice();

    // ---------- DOM refs ----------
    const scanner = $('#ld-scanner');
    const decodePane = $('#ld-decodePane');
    const consoleEl = $('#ld-console');
    const replayBtn = $('#ld-replayBtn');
    const paneFile = $('#ld-paneFile');
    if (!scanner || !decodePane || !consoleEl || !replayBtn || !paneFile) {
      return;
    }

    // ---------- console log (rslint default printer style) ----------
    function pushLog(html: string, extraClass?: string | null) {
      const div = document.createElement('div');
      div.className = 'log ' + (extraClass || 'info');
      div.innerHTML = html;
      consoleEl!.appendChild(div);
    }
    function clearLog() {
      consoleEl!.innerHTML = '';
    }
    function esc(s: string) {
      return s.replace(
        /[&<>]/g,
        (c) =>
          (
            ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }) as Record<
              string,
              string
            >
          )[c],
      );
    }

    function pushDiagnostic(
      v: Violation,
      file: string,
      firstWithRise?: boolean,
    ) {
      const sevClass = v.severity === 'error' ? 'sev-err' : 'sev-warn';
      const sevLabel = v.severity === 'error' ? 'error' : 'warning';

      pushLog(
        ` <span class="rname"> ${esc(v.rule)} </span>` +
          ` — <span class="${sevClass}">[${sevLabel}]</span> ` +
          esc(v.message),
        firstWithRise ? 'diag-in' : null,
      );

      pushLog(
        `  <span class="border">╭─┴──────────(</span> ` +
          `<span class="file">${esc(file)}:${v.line}:${v.col}</span> ` +
          `<span class="border">)─────</span>`,
      );

      const lastN = v.codeLines[v.codeLines.length - 1].n;
      const numW = String(lastN).length;
      for (const cl of v.codeLines) {
        let body: string;
        if (cl.hl) {
          const [a, b] = cl.hl;
          body =
            esc(cl.text.slice(0, a)) +
            `<span class="under">${esc(cl.text.slice(a, b))}</span>` +
            esc(cl.text.slice(b));
        } else {
          body = esc(cl.text);
        }
        const nStr = String(cl.n).padStart(numW, ' ');
        pushLog(
          `  <span class="border">│ </span><span class="gutter">${nStr}</span><span class="border"> │</span>  ${body}`,
        );
      }

      pushLog(
        `  <span class="border">╰────────────────────────────────</span>`,
      );
    }

    // ---------- VS Code MarkerHover popover ----------
    type Tooltip = {
      tip: HTMLDivElement;
      spans: HTMLSpanElement[];
      show: () => void;
      hide: () => void;
      isFixed: () => boolean;
    };
    const tooltipNodes: Tooltip[] = [];
    function spawnTooltip(v: Violation): Tooltip | null {
      const lineIdx = v.line - 1;
      const colIdx = v.col - 1;
      const spans = lineCharSpans[lineIdx] || [];
      const anchor = spans[colIdx];
      if (!anchor) {
        return null;
      }

      const tokenSpans: HTMLSpanElement[] = [];
      for (let i = colIdx; i < colIdx + v.len && i < spans.length; i++) {
        spans[i].classList.add('violation');
        tokenSpans.push(spans[i]);
      }

      const tip = document.createElement('div');
      tip.className = 'vsc-tip';

      const statusBar = v.quickFix
        ? `<div class="status-bar"><a class="action quick-fix" href="#">Quick Fix...</a></div>`
        : '';

      const lspMessage = `[${v.rule}] ${v.message}`;

      tip.innerHTML = `
        <div class="vsc-body">
          <span class="msg">${esc(lspMessage)}</span>
          <span class="details">rslint</span>
        </div>
        ${statusBar}
      `;

      decodeBody!.appendChild(tip);

      const bodyRect = decodeBody!.getBoundingClientRect();
      const anchorRect = anchor.getBoundingClientRect();

      const tipWidth = 340;
      const padding = 8;
      const anchorLeftRel = anchorRect.left - bodyRect.left;
      const anchorTopRel = anchorRect.top - bodyRect.top;

      let tipLeft = anchorLeftRel - 14;
      const maxLeft = decodeBody!.clientWidth - tipWidth - padding;
      if (tipLeft > maxLeft) {
        tipLeft = maxLeft;
      }
      if (tipLeft < padding) {
        tipLeft = padding;
      }

      const tipHeight = tip.offsetHeight;
      const tipGap = 22;
      let tipTop = anchorTopRel - tipHeight - tipGap;
      if (tipTop < padding) {
        tipTop = anchorTopRel + tipGap + 6;
      }

      tip.style.left = tipLeft + 'px';
      tip.style.top = tipTop + 'px';

      let hideTimer: ReturnType<typeof setTimeout> | null = null;
      let fixed = false;
      const show = () => {
        if (fixed) {
          return;
        }
        if (hideTimer) {
          clearTimeout(hideTimer);
          hideTimer = null;
        }
        tip.classList.add('visible');
        tokenSpans.forEach((s) => s.classList.add('hovered'));
      };
      const hide = () => {
        if (hideTimer) {
          clearTimeout(hideTimer);
        }
        hideTimer = setTimeout(() => {
          tip.classList.remove('visible');
          tokenSpans.forEach((s) => s.classList.remove('hovered'));
        }, 120);
      };
      tokenSpans.forEach((s) => {
        s.addEventListener('mouseenter', show);
        s.addEventListener('mouseleave', hide);
      });
      tip.addEventListener('mouseenter', show);
      tip.addEventListener('mouseleave', hide);

      const fixLink = tip.querySelector('.quick-fix');
      const applyFix = () => {
        if (fixed) {
          return;
        }
        if (!v.quickFix || !v.fix) {
          return;
        }
        fixed = true;
        tip.classList.remove('visible');
        triggerRelint(v);
      };
      if (fixLink) {
        fixLink.addEventListener('click', (e) => {
          e.preventDefault();
          e.stopPropagation();
          applyFix();
        });
      }

      return {
        tip,
        spans: tokenSpans,
        show,
        hide,
        isFixed() {
          return fixed;
        },
      };
    }

    function clearTooltips() {
      tooltipNodes.forEach((t) => {
        if (t.spans && t.show && t.hide) {
          t.spans.forEach((s) => {
            s.removeEventListener('mouseenter', t.show);
            s.removeEventListener('mouseleave', t.hide);
          });
        }
        if (t.tip) {
          t.tip.classList.remove('visible');
          t.tip.remove();
        }
        if (t.spans) {
          t.spans.forEach((s) =>
            s.classList.remove('violation', 'visible', 'hovered'),
          );
        }
      });
      tooltipNodes.length = 0;
    }

    // ---------- cycle ----------
    const SCAN_MS = 1100;
    const SCAN_TOP = -40;
    const SCAN_BOT = lines.length * LINE_H + TOP_PAD + 40;

    // ---------- live "rslint is running" progress line ----------
    const SPINNER_FRAMES = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
    const SPIN_INTERVAL = 80;
    const BAR_CELLS = 10;
    const BAR_FILLED = '▰';
    const BAR_EMPTY = '▱';
    let progressEl: HTMLElement | null = null;

    function renderProgress(t: number, done: boolean) {
      if (!progressEl) {
        return;
      }
      if (done) {
        const errs = VIOLATIONS.length;
        const errCls = errs === 0 ? 'ok' : 'bad';
        const errLbl = errs === 1 ? 'error' : 'errors';
        progressEl.innerHTML =
          ` <span class="ok">✓</span> ` +
          `<span class="rname">rslint completed</span>  ` +
          `<span class="meta">·</span>  ` +
          `<span class="${errCls}">${errs} ${errLbl}</span>  ` +
          `<span class="meta">in</span>  ` +
          `<span class="bold">${currentRunMs}ms</span>`;
        return;
      }
      const ratio = Math.max(0, Math.min(1, t / SCAN_MS));
      const filled = Math.round(ratio * BAR_CELLS);
      const bar =
        `<span class="rname">${BAR_FILLED.repeat(filled)}</span>` +
        `<span class="meta">${BAR_EMPTY.repeat(BAR_CELLS - filled)}</span>`;
      const pct = String(Math.round(ratio * 100)).padStart(3, ' ');
      const ms = Math.floor(ratio * currentRunMs);
      const spinIdx = Math.floor(t / SPIN_INTERVAL) % SPINNER_FRAMES.length;
      progressEl.innerHTML =
        `<span class="rname">${SPINNER_FRAMES[spinIdx]}</span>  ` +
        `${bar}  ` +
        `<span class="bold">${pct}%</span>  ` +
        `<span class="meta">·</span>  ` +
        `<span class="bold">${ms}ms</span>`;
    }

    const FILE = './demo.ts';
    const TYPE_MS = 32;
    const CMD_TEXT = `$ rslint ${FILE}`;
    // The border fill spans the WHOLE run — from the first typed char through
    // to the summary beat (typing + scan + the SCAN_MS+840 summary offset) —
    // so the ring completes exactly as the lint result lands, instead of
    // finishing early at scan-end and sitting full while diagnostics render.
    const FILL_END_MS = CMD_TEXT.length * TYPE_MS + SCAN_MS + 840;
    // Once the ring is fully drawn the shine dot keeps looping, but at this
    // calmer pace (one lap / 4s — the original idle speed) instead of the
    // faster draw speed, otherwise the post-fill loop feels too fast.
    const SHINE_CYCLE_MS = 4000;
    const randRunMs = () => 8 + Math.floor(Math.random() * 38);
    let currentRunMs = randRunMs();

    let firedFlags: Record<number, boolean> = {};
    let cycleStart = performance.now();
    // True cycle start (when the command begins typing). Drives the border
    // fill across the whole animation, independent of `cycleStart` which gets
    // offset far into the future during the typing phase.
    let animStart = performance.now();
    let stopped = false;

    // Track timers/rAF so the cleanup can cancel everything.
    let raf = 0;
    let typer: ReturnType<typeof setInterval> | null = null;

    function buildSummaryHTML() {
      const errors = VIOLATIONS.length;
      const errClass = errors === 0 ? 'ok' : 'bad';
      const errLabel = errors === 1 ? 'error' : 'errors';
      const tail =
        `<span class="meta">(linted <span class="bold">1</span> file ` +
        `with <span class="bold">${VIOLATIONS.length}</span> rules ` +
        `in <span class="bold">${currentRunMs}ms</span> ` +
        `using <span class="bold">8</span> threads)</span>`;
      return (
        `Found <span class="${errClass}">${errors} ${errLabel}</span> and ` +
        `<span class="ok">0</span> warnings ` +
        tail
      );
    }

    function armTooltip(v: Violation) {
      const t = spawnTooltip(v);
      if (!t) {
        return;
      }
      tooltipNodes.push(t);
      const beatDelay = (tooltipNodes.length - 1) * 0.7;
      t.spans.forEach((s) => {
        s.style.animationDelay = beatDelay + 's';
        s.classList.add('visible');
      });
    }

    const SCHEDULE: { at: number; fn: () => void }[] = [
      {
        at: 0,
        fn: () => {
          // Sync the border-fill clock to the moment the command line (and
          // its typing cursor) actually appears, so the ring starts drawing
          // in lockstep with the cursor instead of a frame ahead.
          animStart = performance.now();
          pushLog(
            `<span class="cmd-group">` +
              `<span class="cmd cmd-typed"></span>` +
              `<span class="cmd type-caret">▌</span>` +
              `</span>` +
              `<span class="prog-right"></span>`,
            'cmdline',
          );
          const lastLog = consoleEl!.lastElementChild;
          const cmdEl =
            lastLog &&
            (lastLog.querySelector('.cmd-typed') as HTMLElement | null);
          const caretEl =
            lastLog &&
            (lastLog.querySelector('.type-caret') as HTMLElement | null);
          const progSpan =
            lastLog &&
            (lastLog.querySelector('.prog-right') as HTMLElement | null);

          cycleStart = performance.now() + 1e7;

          const text = CMD_TEXT;
          let i = 0;
          typer = setInterval(() => {
            if (!cmdEl || !cmdEl.parentNode) {
              if (typer) {
                clearInterval(typer);
              }
              return;
            }
            if (i >= text.length) {
              if (typer) {
                clearInterval(typer);
              }
              if (caretEl && caretEl.parentNode) {
                caretEl.remove();
              }
              progressEl = progSpan;
              cycleStart = performance.now();
              renderProgress(0, false);
              return;
            }
            cmdEl.textContent = text.slice(0, ++i);
          }, TYPE_MS);
        },
      },
      {
        at: SCAN_MS,
        fn: () => {
          if (progressEl) {
            progressEl.classList.add('fading');
            progressEl = null;
          }
        },
      },
      {
        at: SCAN_MS + 100,
        fn: () => VIOLATIONS[0] && armTooltip(VIOLATIONS[0]),
      },
      {
        at: SCAN_MS + 240,
        fn: () => VIOLATIONS[1] && armTooltip(VIOLATIONS[1]),
      },
      {
        at: SCAN_MS + 200,
        fn: () => VIOLATIONS[0] && pushDiagnostic(VIOLATIONS[0], FILE, true),
      },
      {
        at: SCAN_MS + 520,
        fn: () => VIOLATIONS[1] && pushDiagnostic(VIOLATIONS[1], FILE, true),
      },
      {
        at: SCAN_MS + 840,
        fn: () => {
          pushLog(buildSummaryHTML(), 'diag-in');
          replayBtn!.classList.add('ready');
          paneFile!.textContent = `done · ${FILE}`;
          stopped = true;
        },
      },
    ];

    function resetCycle(now: number) {
      cycleStart = now;
      animStart = now;
      currentRunMs = randRunMs();
      firedFlags = {};
      stopped = false;
      progressEl = null;
      decodePane!.style.setProperty('--ld-border-fill', '0');
      replayBtn!.classList.remove('ready');
      paneFile!.textContent = `linting ${FILE}`;
      clearLog();
      clearTooltips();
      buildCharNodes();
    }

    function resetForReplay(now: number) {
      SOURCE = ORIGINAL_SOURCE;
      VIOLATIONS = ORIGINAL_VIOLATIONS.slice();
      resetCycle(now);
    }

    function triggerRelint(v: Violation) {
      if (!v || !v.fix) {
        return;
      }
      const srcLines = SOURCE.split('\n');
      const lineIdx0 = v.line - 1;
      const colIdx0 = v.col - 1;
      if (srcLines[lineIdx0]) {
        const ln = srcLines[lineIdx0];
        srcLines[lineIdx0] =
          ln.slice(0, colIdx0) + v.fix.replace + ln.slice(colIdx0 + v.len);
        SOURCE = srcLines.join('\n');
      }
      VIOLATIONS = VIOLATIONS.filter(
        (x) => !(x.rule === v.rule && x.line === v.line && x.col === v.col),
      );
      resetCycle(performance.now());
    }

    const onReplayClick = () => {
      resetForReplay(performance.now());
    };
    replayBtn.addEventListener('click', onReplayClick);

    function frame(now: number) {
      const t = now - cycleStart;

      // The border fill grows from the first frame to the summary beat, and the
      // shine dot rides exactly at its leading edge (conic starts at -90deg /
      // top, clockwise) — the dot "draws" the ring as it grows (one lap over
      // FILL_END_MS). Once the ring is full, fill clamps at 1 and the dot keeps
      // looping from the 270deg seam, but at the calmer SHINE_CYCLE_MS pace.
      const elapsed = Math.max(0, now - animStart);
      const fillRaw = elapsed / FILL_END_MS;
      decodePane!.style.setProperty(
        '--ld-border-fill',
        String(Math.min(1, fillRaw)),
      );
      // Start at -52deg (top-left corner) to match the fill conic's `from` in
      // styles.module.scss; 308 = -52 + 360 is the seam after one full lap.
      const shineDeg =
        elapsed < FILL_END_MS
          ? -52 + fillRaw * 360
          : 308 + ((elapsed - FILL_END_MS) / SHINE_CYCLE_MS) * 360;
      decodePane!.style.setProperty('--ld-shine-angle', `${shineDeg}deg`);

      let scanY: number;
      if (t < SCAN_MS) {
        scanY = SCAN_TOP + (t / SCAN_MS) * (SCAN_BOT - SCAN_TOP);
      } else {
        scanY = SCAN_BOT;
      }
      scanner!.style.transform = `translateY(${scanY - 35}px)`;

      for (const cn of charNodes) {
        let newState: string;
        if (cn.y < scanY - 32) {
          newState = 'dec';
        } else if (cn.y < scanY + 12) {
          newState = 'scr';
        } else {
          newState = 'enc';
        }

        if (newState !== cn.state) {
          cn.state = newState;
          const keep: string[] = [];
          const cls0 = cn.node.classList;
          if (cls0.contains('violation')) {
            keep.push('violation');
          }
          if (cls0.contains('visible')) {
            keep.push('visible');
          }
          if (cls0.contains('hovered')) {
            keep.push('hovered');
          }
          let cls = `ch ${newState}`;
          if (newState === 'dec' && cn.kind) {
            cls += ` t-${cn.kind}`;
          }
          if (keep.length) {
            cls += ' ' + keep.join(' ');
          }
          cn.node.className = cls;
        }
      }

      if (!stopped) {
        for (let i = 0; i < SCHEDULE.length; i++) {
          if (!firedFlags[i] && t >= SCHEDULE[i].at) {
            SCHEDULE[i].fn();
            firedFlags[i] = true;
          }
        }
      }

      if (progressEl && t < SCAN_MS) {
        renderProgress(t, false);
      }

      raf = requestAnimationFrame(frame);
    }
    raf = requestAnimationFrame(frame);

    return () => {
      cancelAnimationFrame(raf);
      if (typer) {
        clearInterval(typer);
      }
      replayBtn.removeEventListener('click', onReplayClick);
      clearTooltips();
    };
  }, []);

  return (
    <div className={styles.lintDemo} ref={rootRef}>
      <div className="stage-col">
        <div className="decode-pane" id="ld-decodePane">
          <div className="ld-frame" aria-hidden="true" />
          <div className="pane-head">
            <span className="live" />
            <span>rslint</span>
            <span className="right">
              <span id="ld-paneFile">linting ./demo.ts</span>
              <button
                className="replay"
                id="ld-replayBtn"
                type="button"
                aria-label="Rerun lint"
              >
                <span className="icon">{'↻'}</span>
                <span>rerun</span>
              </button>
            </span>
          </div>
          <div className="decode-body" id="ld-decodeBody">
            <div className="scanner" id="ld-scanner" />
            <div className="code-text" id="ld-code" />
          </div>
          <div className="console" id="ld-console" />
        </div>
      </div>
    </div>
  );
}
