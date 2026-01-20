const ANSI_RED = '\x1b[31m';
const ANSI_YELLOW = '\x1b[33m';
const ANSI_CYAN = '\x1b[36m';
const ANSI_GRAY = '\x1b[90m';
const ANSI_BRIGHT = '\x1b[97m';
const ANSI_RESET = '\x1b[0m';

function shouldColorErrors(): boolean {
  if (!process.stderr || !process.stderr.isTTY) {
    return false;
  }
  if (process.env.NO_COLOR != null) {
    return false;
  }
  return true;
}

export function colorizeError(text: string | undefined | null): string {
  if (!text) {
    return text ?? '';
  }
  if (!shouldColorErrors()) {
    return text;
  }
  return `${ANSI_RED}${text}${ANSI_RESET}`;
}

function wrap(color: string, text: string | number): string {
  return `${color}${text}${ANSI_RESET}`;
}

export function colorizeDiagnostics(text: string | undefined | null): string {
  if (!text) {
    return text ?? '';
  }
  if (!shouldColorErrors()) {
    return text;
  }
  const lines = text.split(/\r?\n/);
  const keepTrailingNewline = text.endsWith('\n');
  const colored = lines.map(colorizeDiagnosticLine);
  let output = colored.join('\n');
  if (keepTrailingNewline) {
    output += '\n';
  }
  return output;
}

function colorizeDiagnosticLine(line: string): string {
  if (!line) {
    return line;
  }
  if (line.startsWith('TSError:')) {
    return wrap(ANSI_BRIGHT, line);
  }

  const diagMatch = line.match(
    /^(.+?):(\d+):(\d+) - (\w+)(?: TS(\d+))?: (.*)$/,
  );
  if (diagMatch) {
    const [, file, lineNo, colNo, severity, code, message] = diagMatch;
    const severityLower = severity.toLowerCase();
    const severityColor =
      severityLower === 'error'
        ? ANSI_RED
        : severityLower === 'warning'
          ? ANSI_YELLOW
          : ANSI_GRAY;
    const suffix = code
      ? ` ${wrap(ANSI_GRAY, `TS${code}`)}: ${message}`
      : `: ${message}`;
    return `${wrap(ANSI_CYAN, file)}:${wrap(ANSI_YELLOW, lineNo)}:${wrap(
      ANSI_YELLOW,
      colNo,
    )} - ${wrap(severityColor, severity)}${suffix}`;
  }

  const codeLineMatch = line.match(/^(\s*)(\d+)(\s*\|\s)(.*)$/);
  if (codeLineMatch) {
    const [, indent, lineNo, separator, code] = codeLineMatch;
    return `${indent}${wrap(ANSI_GRAY, lineNo)}${wrap(ANSI_GRAY, separator)}${code}`;
  }

  const underlineMatch = line.match(/^(\s*\| )(.*)$/);
  if (underlineMatch && underlineMatch[2].includes('~')) {
    const prefix = underlineMatch[1];
    const rest = underlineMatch[2].replace(/~+/g, value =>
      wrap(ANSI_RED, value),
    );
    return `${wrap(ANSI_GRAY, prefix)}${rest}`;
  }

  return line;
}
