import child_process from "node:child_process";
import path from "node:path";
export async function lint(tsconfig: string) {
  let binPath = path.resolve(import.meta.dirname, "../bin/rslint");
  let cmd = `${binPath}`;
  let args = ["--format=jsonline", `--tsconfig=${tsconfig}`];
  let defaultCwd = process.cwd();
  let child = child_process.spawn(cmd, args, {
    stdio: ["pipe", "pipe", "inherit"],
    cwd: defaultCwd,
  });
  return new Promise((resolve, reject) => {
    let chunks: Buffer[] = [];
    child.stdout.on("data", (chunk: Buffer) => {
       chunks.push(chunk);
    });
    child.stdout.on('end', () => {
      let message = Buffer.concat(chunks).toString();
      let diags = message
        .split("\n")
        .filter((x) => {
          // FIXME: we should not generate empty line when generate diags
          return x.trim().length !== 0;
        })
        .map((x) => {
          return JSON.parse(x);
        });
      resolve(diags);
    })
    child.on("error", (err) => {
      reject(err);
    });
    child.on("exit", (code) => {
      if (code !== 0) {
        reject(new Error(`Linting process exited with code ${code}`));
      }
    });
  });
}
