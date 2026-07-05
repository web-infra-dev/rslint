export {};

export interface ReactTestRendererJSON {
  type: string;
  props: { [propName: string]: any };
  children: null | ReactTestRendererNode[];
}

export type ReactTestRendererNode = ReactTestRendererJSON | string;

export function create(nextElement: any, options?: any): any;
