import {
  createConnection,
  TextDocuments,
  ProposedFeatures,
  InitializeParams,
  DidChangeConfigurationNotification,
  TextDocumentSyncKind,
  InitializeResult,
  DocumentDiagnosticParams,
  DocumentDiagnosticReport,
  DocumentDiagnosticReportKind,
  Diagnostic,
  DiagnosticSeverity,
  Position,
  Range
} from 'vscode-languageserver/node';
import { TextDocument } from 'vscode-languageserver-textdocument';
import { spawn } from 'child_process';
import * as path from 'path';

const connection = createConnection(ProposedFeatures.all);
const documents: TextDocuments<TextDocument> = new TextDocuments(TextDocument);

let hasConfigurationCapability = false;
let hasWorkspaceFolderCapability = false;
let hasDiagnosticRelatedInformationCapability = false;

interface TSGoLintSettings {
  enable: boolean;
  executablePath: string;
}

const defaultSettings: TSGoLintSettings = {
  enable: true,
  executablePath: 'tsgolint'
};

let globalSettings: TSGoLintSettings = defaultSettings;
const documentSettings: Map<string, Thenable<TSGoLintSettings>> = new Map();

connection.onInitialize((params: InitializeParams) => {
  const capabilities = params.capabilities;

  hasConfigurationCapability = !!(
    capabilities.workspace && !!capabilities.workspace.configuration
  );
  hasWorkspaceFolderCapability = !!(
    capabilities.workspace && !!capabilities.workspace.workspaceFolders
  );
  hasDiagnosticRelatedInformationCapability = !!(
    capabilities.textDocument &&
    capabilities.textDocument.publishDiagnostics &&
    capabilities.textDocument.publishDiagnostics.relatedInformation
  );

  const result: InitializeResult = {
    capabilities: {
      textDocumentSync: TextDocumentSyncKind.Incremental,
      diagnosticProvider: {
        interFileDependencies: false,
        workspaceDiagnostics: false
      }
    }
  };

  if (hasWorkspaceFolderCapability) {
    result.capabilities.workspace = {
      workspaceFolders: {
        supported: true
      }
    };
  }

  return result;
});

connection.onInitialized(() => {
  if (hasConfigurationCapability) {
    connection.client.register(DidChangeConfigurationNotification.type, undefined);
  }
  if (hasWorkspaceFolderCapability) {
    connection.workspace.onDidChangeWorkspaceFolders(_event => {
      connection.console.log('Workspace folder change event received.');
    });
  }
});

connection.onDidChangeConfiguration(change => {
  if (hasConfigurationCapability) {
    documentSettings.clear();
  } else {
    globalSettings = <TSGoLintSettings>(
      (change.settings.tsgolint || defaultSettings)
    );
  }

  documents.all().forEach(validateTextDocument);
});

function getDocumentSettings(resource: string): Thenable<TSGoLintSettings> {
  if (!hasConfigurationCapability) {
    return Promise.resolve(globalSettings);
  }
  let result = documentSettings.get(resource);
  if (!result) {
    result = connection.workspace.getConfiguration({
      scopeUri: resource,
      section: 'tsgolint'
    });
    documentSettings.set(resource, result);
  }
  return result;
}

documents.onDidClose(e => {
  documentSettings.delete(e.document.uri);
});

documents.onDidChangeContent(change => {
  validateTextDocument(change.document);
});

async function validateTextDocument(textDocument: TextDocument): Promise<void> {
  const settings = await getDocumentSettings(textDocument.uri);
  
  if (!settings.enable) {
    connection.sendDiagnostics({ uri: textDocument.uri, diagnostics: [] });
    return;
  }

  const diagnostics = await runTSGoLint(textDocument, settings);
  connection.sendDiagnostics({ uri: textDocument.uri, diagnostics });
}

interface TSGoLintOutput {
  file: string;
  line: number;
  column: number;
  severity: 'error' | 'warning' | 'info';
  message: string;
  rule?: string;
}

async function runTSGoLint(textDocument: TextDocument, settings: TSGoLintSettings): Promise<Diagnostic[]> {
  return new Promise((resolve) => {
    const diagnostics: Diagnostic[] = [];
    const filePath = textDocument.uri.replace('file://', '');
    
    const tsgolint = spawn(settings.executablePath, [filePath, '--format', 'json'], {
      cwd: path.dirname(filePath)
    });

    let output = '';
    let errorOutput = '';

    tsgolint.stdout.on('data', (data) => {
      output += data.toString();
    });

    tsgolint.stderr.on('data', (data) => {
      errorOutput += data.toString();
    });

    tsgolint.on('close', (code) => {
      if (errorOutput) {
        connection.console.error(`TSGoLint error: ${errorOutput}`);
      }

      try {
        const results: TSGoLintOutput[] = JSON.parse(output);
        
        for (const result of results) {
          const diagnostic: Diagnostic = {
            severity: result.severity === 'error' ? DiagnosticSeverity.Error :
                     result.severity === 'warning' ? DiagnosticSeverity.Warning :
                     DiagnosticSeverity.Information,
            range: {
              start: Position.create(result.line - 1, result.column - 1),
              end: Position.create(result.line - 1, result.column)
            },
            message: result.message,
            source: 'tsgolint'
          };

          if (result.rule) {
            diagnostic.code = result.rule;
          }

          diagnostics.push(diagnostic);
        }
      } catch (e) {
        connection.console.error(`Failed to parse TSGoLint output: ${e}`);
      }

      resolve(diagnostics);
    });

    tsgolint.on('error', (err) => {
      connection.console.error(`Failed to run TSGoLint: ${err.message}`);
      resolve([]);
    });
  });
}

connection.languages.diagnostics.on(async (params: DocumentDiagnosticParams) => {
  const document = documents.get(params.textDocument.uri);
  if (document !== undefined) {
    return {
      kind: DocumentDiagnosticReportKind.Full,
      items: await validateTextDocument(document).then(() => [])
    } as DocumentDiagnosticReport;
  } else {
    return {
      kind: DocumentDiagnosticReportKind.Full,
      items: []
    } as DocumentDiagnosticReport;
  }
});

documents.listen(connection);
connection.listen();