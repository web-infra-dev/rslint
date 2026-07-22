import { useEffect, useRef } from 'react';
import styles from './styles.module.scss';

/**
 * Parallax star field — ambient full-page background canvas.
 *
 * Ported near-verbatim from the d-glitch-decode prototype's `starfield`
 * IIFE. Three depth layers drift leftward at their own speed; per-star
 * twinkle modulates alpha + luminance so the field gently breathes, and a
 * very low-amplitude mouse force field leaves a soft wake.
 *
 * Theme detection is inverted vs the prototype: it used
 * `classList.contains('theme-light')`, but rspress marks DARK mode with
 * `.dark` on <html> (theme/index.scss:11). So light == NO `.dark` class.
 * A MutationObserver on <html>'s class reflects rspress theme switches in
 * real time (the prototype polled / toggled a class directly).
 */
export function Starfield() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !canvas.getContext) {
      return;
    }
    const ctx = canvas.getContext('2d', { alpha: true });
    if (!ctx) {
      return;
    }

    const MAX_DPR = 1.75;

    type LayerSpec = {
      count: number;
      speed: number;
      sizeMin: number;
      sizeMax: number;
      alphaMin: number;
      alphaMax: number;
      twinkle: number;
    };

    const LAYERS_DARK: LayerSpec[] = [
      {
        count: 90,
        speed: 18,
        sizeMin: 1.1,
        sizeMax: 2.4,
        alphaMin: 0.55,
        alphaMax: 1.0,
        twinkle: 0.4,
      },
      {
        count: 160,
        speed: 10,
        sizeMin: 0.7,
        sizeMax: 1.5,
        alphaMin: 0.35,
        alphaMax: 0.75,
        twinkle: 0.3,
      },
      {
        count: 240,
        speed: 4,
        sizeMin: 0.4,
        sizeMax: 0.9,
        alphaMin: 0.18,
        alphaMax: 0.45,
        twinkle: 0.18,
      },
    ];
    const LAYERS_LIGHT: LayerSpec[] = [
      {
        count: 70,
        speed: 16,
        sizeMin: 0.9,
        sizeMax: 2.0,
        alphaMin: 0.35,
        alphaMax: 0.7,
        twinkle: 0.28,
      },
      {
        count: 130,
        speed: 9,
        sizeMin: 0.6,
        sizeMax: 1.3,
        alphaMin: 0.22,
        alphaMax: 0.5,
        twinkle: 0.2,
      },
      {
        count: 200,
        speed: 3,
        sizeMin: 0.3,
        sizeMax: 0.8,
        alphaMin: 0.12,
        alphaMax: 0.32,
        twinkle: 0.12,
      },
    ];

    // [r, g, b, pick-weight]
    const TINTS_DARK: number[][] = [
      [255, 245, 232, 0.45],
      [192, 150, 255, 0.25],
      [255, 158, 84, 0.18],
      [142, 200, 255, 0.12],
    ];
    const TINTS_LIGHT: number[][] = [
      [144, 144, 152, 0.5],
      [196, 168, 120, 0.28],
      [168, 144, 196, 0.14],
      [196, 86, 32, 0.08],
    ];

    let W = 0;
    let H = 0;
    let DPR = 1;

    function resize() {
      DPR = Math.min(MAX_DPR, window.devicePixelRatio || 1);
      W = window.innerWidth;
      H = window.innerHeight;
      canvas!.style.width = W + 'px';
      canvas!.style.height = H + 'px';
      canvas!.width = Math.floor(W * DPR);
      canvas!.height = Math.floor(H * DPR);
      ctx!.setTransform(DPR, 0, 0, DPR, 0, 0);
    }

    type Layer = {
      specDark: LayerSpec;
      specLight: LayerSpec;
      capacity: number;
      x: Float32Array;
      y: Float32Array;
      size: Float32Array;
      baseAlpha: Float32Array;
      twinklePhase: Float32Array;
      tintIdx: Uint8Array;
    };

    let layers: Layer[] = [];

    function pickWeightedTint(list: number[][]) {
      const r = Math.random();
      let acc = 0;
      for (let i = 0; i < list.length; i++) {
        acc += list[i][3];
        if (r < acc) {
          return i;
        }
      }
      return list.length - 1;
    }

    function seed() {
      layers = [];
      for (let li = 0; li < LAYERS_DARK.length; li++) {
        const d = LAYERS_DARK[li];
        const l = LAYERS_LIGHT[li];
        const N = Math.max(d.count, l.count);
        const x = new Float32Array(N);
        const y = new Float32Array(N);
        const sz = new Float32Array(N);
        const a0 = new Float32Array(N);
        const tp = new Float32Array(N);
        const ti = new Uint8Array(N);
        for (let i = 0; i < N; i++) {
          x[i] = Math.random() * W;
          y[i] = Math.random() * H;
          sz[i] = d.sizeMin + Math.random() * (d.sizeMax - d.sizeMin);
          a0[i] = d.alphaMin + Math.random() * (d.alphaMax - d.alphaMin);
          tp[i] = Math.random() * Math.PI * 2;
          ti[i] = pickWeightedTint(TINTS_DARK);
        }
        layers.push({
          specDark: d,
          specLight: l,
          capacity: N,
          x,
          y,
          size: sz,
          baseAlpha: a0,
          twinklePhase: tp,
          tintIdx: ti,
        });
      }
    }

    resize();
    seed();

    const onResize = () => {
      resize();
      seed();
    };
    window.addEventListener('resize', onResize);

    // Mouse — very subtle wake.
    const MOUSE_FORCE = 220;
    let mx = -9999;
    let my = -9999;
    let hasMouse = false;
    const onMouseMove = (e: MouseEvent) => {
      mx = e.clientX;
      my = e.clientY;
      hasMouse = true;
    };
    const onMouseLeave = () => {
      hasMouse = false;
    };
    window.addEventListener('mousemove', onMouseMove, { passive: true });
    window.addEventListener('mouseleave', onMouseLeave);

    // rspress: dark mode == html.dark; light == no .dark class.
    const isLight = () => !document.documentElement.classList.contains('dark');

    let raf = 0;
    let prev = performance.now();
    function render(now: number) {
      const dt = Math.min(0.08, (now - prev) / 1000);
      prev = now;

      const light = isLight();
      const layerSpecs = light ? LAYERS_LIGHT : LAYERS_DARK;
      const tintList = light ? TINTS_LIGHT : TINTS_DARK;

      ctx!.clearRect(0, 0, W, H);
      ctx!.globalCompositeOperation = light ? 'source-over' : 'lighter';

      for (let li = 0; li < layers.length; li++) {
        const L = layers[li];
        const spec = layerSpecs[li];
        const N = Math.min(L.capacity, spec.count);
        const speed = spec.speed;
        const twinkleAmp = spec.twinkle;

        for (let i = 0; i < N; i++) {
          let x = L.x[i] - speed * dt;
          let y = L.y[i] + speed * dt * 0.08;

          if (x < -4) {
            x = W + 4;
          }
          if (y > H + 4) {
            y = -4;
          }

          if (hasMouse) {
            const dx = x - mx;
            const dy = y - my;
            const d2 = dx * dx + dy * dy + 60;
            if (d2 < 30000) {
              const f = MOUSE_FORCE / d2;
              x += dx * f * dt;
              y += dy * f * dt;
            }
          }

          L.x[i] = x;
          L.y[i] = y;

          const tw = Math.sin(now * 0.0014 + L.twinklePhase[i]);
          const a = Math.max(0.04, L.baseAlpha[i] + tw * twinkleAmp * 0.5);
          const lumMult = 1 + tw * 0.15;
          const tint = tintList[L.tintIdx[i]];
          const r = Math.min(255, Math.round(tint[0] * lumMult));
          const g = Math.min(255, Math.round(tint[1] * lumMult));
          const b = Math.min(255, Math.round(tint[2] * lumMult));

          ctx!.fillStyle =
            'rgba(' + r + ',' + g + ',' + b + ',' + a.toFixed(3) + ')';
          ctx!.beginPath();
          ctx!.arc(x, y, L.size[i], 0, Math.PI * 2);
          ctx!.fill();
        }
      }

      raf = requestAnimationFrame(render);
    }
    raf = requestAnimationFrame(render);

    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener('resize', onResize);
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseleave', onMouseLeave);
    };
  }, []);

  return (
    <canvas ref={canvasRef} className={styles.starfield} aria-hidden="true" />
  );
}
