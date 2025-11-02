declare module '*.mdx' {
  let MDXComponent: () => JSX.Element;
  export default MDXComponent;
}

declare module '*.module.scss' {
  const classes: Readonly<Record<string, string>>;
  export default classes;
}
