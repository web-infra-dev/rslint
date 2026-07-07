export const dynamicDepthOne = () =>
  import('./consumer').then(({ rootValue }) => rootValue);
