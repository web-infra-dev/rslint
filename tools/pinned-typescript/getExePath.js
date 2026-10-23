// Keep executable resolution with the pinned upstream checkout on every host.
module.exports = async function getExePath() {
  const { default: resolve } =
    await import('../../typescript-go/packages/typescript/lib/getExePath.js');
  return resolve();
};
