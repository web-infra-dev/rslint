#!/usr/bin/env node
const path = require('node:path');
const binPath = path.resolve(__dirname,'./rslint');
require("child_process").execFileSync(binPath, process.argv.slice(2), { stdio: "inherit" });
