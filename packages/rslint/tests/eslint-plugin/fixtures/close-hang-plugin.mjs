import { writeFileSync } from 'node:fs';

export default {
  rules: {
    hang: {
      meta: { type: 'problem', schema: [] },
      create() {
        return {
          Program() {
            const marker = process.env.RSLINT_API_CLOSE_MARKER;
            if (marker) writeFileSync(marker, 'started');
            // close() must terminate this worker; the listener never yields.
            // eslint-disable-next-line no-constant-condition
            while (true) {
              // empty
            }
          },
        };
      },
    },
  },
};
