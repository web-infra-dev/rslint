export function parseEnvBool(
  value: string | undefined | null,
  fallback: boolean,
): boolean {
  if (value == null) {
    return fallback;
  }
  switch (String(value).toLowerCase()) {
    case '1':
    case 'true':
    case 'yes':
    case 'on':
      return true;
    case '0':
    case 'false':
    case 'no':
    case 'off':
      return false;
    default:
      return fallback;
  }
}
