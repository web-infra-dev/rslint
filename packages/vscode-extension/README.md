# Rslint VS Code Extension

Language support for Rslint in Visual Studio Code.

## Features

- **Real-time linting**: Get instant feedback on TypeScript and JavaScript code
- **TypeScript-powered analysis**: Leverages full TypeScript semantics for accurate linting
- **Language Server Protocol**: High-performance linting with LSP integration

## Commands

### Restart Server

Use the **"rslint: Restart Server"** command to restart the Rslint language server:

1. Open the Command Palette (`Ctrl/Cmd + Shift + P`)
2. Type "rslint: Restart Server"
3. Press Enter

This is useful when:

- The language server becomes unresponsive
- You've updated your Rslint configuration
- You want to reload the language server with fresh settings

## Configuration

You can configure the extension through VS Code settings:

- `rslint.enable`: Enable/disable rslint (default: `true`)
- `rslint.binPath`: Path to rslint executable (optional)
- `rslint.trace.server`: Traces the communication between VS Code and the language server

## Supported Languages

- TypeScript (`.ts`)
- TypeScript React (`.tsx`)
- JavaScript (`.js`)
- JavaScript React (`.jsx`)

## Requirements

- Visual Studio Code 1.74.0 or higher
- Rslint binary (automatically included with the extension)
