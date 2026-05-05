// `export =` proxying another module that uses `export default`.
import inner from './typescript-default';
export = inner;
