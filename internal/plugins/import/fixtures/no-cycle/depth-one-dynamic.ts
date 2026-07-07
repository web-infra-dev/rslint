export const dynamicDepthOne = () =>
  import('../file').then(({ rootValue }) => rootValue);
