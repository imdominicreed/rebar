const { LanguageClient, TransportKind } = require("vscode-languageclient/node");
const vscode = require("vscode");

let client;

function activate(context) {
  const config = vscode.workspace.getConfiguration("rebar");
  const configPath = config.get("serverPath");
  const command = configPath || "rebar";

  const outputChannel = vscode.window.createOutputChannel("Rebar LSP");
  outputChannel.appendLine("Starting rebar LSP: " + command + " lsp");

  const serverOptions = {
    command: command,
    args: ["lsp"],
    transport: TransportKind.stdio,
  };

  const clientOptions = {
    documentSelector: [
      { scheme: "file", language: "go" },
      { scheme: "file", language: "typescript" },
      { scheme: "file", language: "typescriptreact" },
      { scheme: "file", language: "javascript" },
      { scheme: "file", language: "python" },
      { scheme: "file", language: "rust" },
      { scheme: "file", language: "java" },
      { scheme: "file", language: "shellscript" },
      { scheme: "file", language: "markdown" },
      { scheme: "file", language: "kotlin" },
      { scheme: "file", language: "c" },
      { scheme: "file", language: "cpp" },
    ],
    outputChannel: outputChannel,
  };

  client = new LanguageClient(
    "rebar-lsp",
    "Rebar LSP",
    serverOptions,
    clientOptions
  );

  client.start();
}

function deactivate() {
  if (client) {
    return client.stop();
  }
}

module.exports = { activate, deactivate };
