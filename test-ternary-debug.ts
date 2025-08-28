// This should be reported with ignoreTernaryTests: false
declare let x: string | null;
const result = x ? x : 'default';
