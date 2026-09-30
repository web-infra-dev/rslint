// The part of @types/node the upstream fixture program relies on.
interface Console {
  log(message?: any, ...optionalParams: any[]): void;
}
declare var console: Console;
