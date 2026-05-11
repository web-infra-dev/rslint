// Gap file — NOT in tsconfig `include` (which is src/ only). The plugin
// rule `fx/no-forbidden` must still fire here (plugin dispatch via the
// worker does not require the file to be in the TS program), alongside
// the native syntax rule `no-debugger`.
debugger;
const forbidden = 2;
export { forbidden };
