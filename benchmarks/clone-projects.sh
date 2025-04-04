#!/usr/bin/env -S bash -euxo pipefail

git clone --single-branch --depth 1 --branch 1.99.0 https://github.com/microsoft/vscode

git clone -c core.longpaths=true --single-branch --depth 1 --branch v5.8.2 https://github.com/microsoft/typescript

git clone --single-branch --depth 1 --branch 0.3.22 https://github.com/typeorm/typeorm

git clone --single-branch --depth 1 --branch v3.5.13 https://github.com/vuejs/core vuejs
