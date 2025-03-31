#!/usr/bin/env -S bash -euxo pipefail

git clone --single-branch --depth 1 https://github.com/microsoft/vscode

git clone --single-branch --depth 1 https://github.com/microsoft/typescript

git clone --single-branch --depth 1 https://github.com/typeorm/typeorm
