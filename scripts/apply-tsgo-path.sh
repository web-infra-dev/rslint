#! /bin/bash

# apply tsgo path
cd typescript-go
if git apply --check ../__patches__/typescript-go.patch 2>/dev/null; then
    echo "Patch can be applied, applying..."
    git apply ../__patches__/typescript-go.patch
else
    echo "Patch already applied or cannot be applied, skipping..."
fi