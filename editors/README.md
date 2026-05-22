# Rebar LSP — Editor Setup

The rebar LSP server provides editor integration for `CONTRACT:` references in source file headers (first 10 lines).

## Features

### Diagnostics

Red squiggly underlines appear on any `CONTRACT:` reference that doesn't match an actual contract file in `architecture/`. The error reads: *Contract "XYZ" not found in architecture/*. Diagnostics update on open and save. Saving a contract file refreshes the index and re-checks all open files, so adding a new contract clears errors immediately.

### Go to Definition

From a **source file**: jump from a `CONTRACT:` reference to the contract's markdown file in `architecture/`.

From a **contract file**: jump from a file path in a list item (e.g. `- src/foo/bar.go`) to that source file. Supports backtick-wrapped paths and dash separators.

### Go to Implementation

Shows all source files that implement a given contract. Works from both contract files and source files — it searches for `CONTRACT:` headers across the repo using `git grep`. Select a result to jump to the exact header line in that file.

### Find References

Shows which **other contracts** reference a given contract ID. This reveals the contract dependency graph — which contracts depend on or mention a given contract. Only searches `architecture/CONTRACT-*.md` files (use Go to Implementation to find source files).

### Completions

After typing `CONTRACT:`, a dropdown lists all known contracts in the workspace. Each item shows the fully qualified reference (e.g. `github.com/willackerly/rebar:S2-ASK-CLI.1.0`) and the contract's human-readable name.

### Hover

Hovering over a `CONTRACT:` reference shows a tooltip with the contract ID, version, name, and a short excerpt from the "Why this exists" section (up to 300 characters).

## Install

Download the latest binary and VS Code extension from the [GitHub release](https://github.com/imdominicreed/rebar/releases/tag/v0.1.0-lsp).

```bash
# Download binary (macOS arm64)
curl -L -o rebar https://github.com/imdominicreed/rebar/releases/download/v0.1.0-lsp/rebar
chmod +x rebar && sudo mv rebar /usr/local/bin/

# Download and install VS Code extension
curl -L -o rebar-lsp.vsix https://github.com/imdominicreed/rebar/releases/download/v0.1.0-lsp/rebar-lsp-0.1.0.vsix
code --install-extension rebar-lsp.vsix && rm rebar-lsp.vsix

# Verify
rebar lsp --help
```

If you're building from source:

```bash
cd cli
go build -o /usr/local/bin/rebar ./main.go
```

## Neovim (0.11+)

Add to your LSP config (e.g. `lua/plugins/lsp/lspconfig.lua`):

```lua
vim.lsp.config["rebar"] = {
  cmd = { "rebar", "lsp" },
  filetypes = { "go", "typescript", "typescriptreact", "javascript",
                "python", "rust", "java", "sh", "markdown" },
  root_markers = { ".rebarrc", ".rebar" },
}

vim.lsp.enable("rebar")
```

Restart Neovim. The LSP attaches automatically in any repo with a `.rebarrc` or `.rebar` directory.

## VS Code

If the `code` command isn't available in your terminal, open VS Code and run `Cmd+Shift+P` → **Shell Command: Install 'code' command in PATH**.

Download and install the extension:

```bash
curl -L -o rebar-lsp.vsix https://github.com/imdominicreed/rebar/releases/download/v0.1.0-lsp/rebar-lsp-0.1.0.vsix
code --install-extension rebar-lsp.vsix && rm rebar-lsp.vsix
```

Or in VS Code: `Cmd+Shift+P` → **Extensions: Install from VSIX...** and select the downloaded file.

If `rebar` isn't on your PATH, set the binary location in VS Code settings:

```json
{
  "rebar.serverPath": "/path/to/rebar"
}
```

Reload the window (`Cmd+Shift+P` → **Developer: Reload Window**).

## JetBrains (IntelliJ, GoLand, WebStorm)

1. Install the [LSP4IJ](https://plugins.jetbrains.com/plugin/23257-lsp4ij) plugin
2. Go to **Settings → Languages & Frameworks → Language Servers**
3. Click **+** to add a new server:
   - **Name:** Rebar
   - **Command:** `rebar lsp`
   - **File patterns:** `*.go`, `*.ts`, `*.tsx`, `*.js`, `*.py`, `*.rs`, `*.java`, `*.sh`, `*.md`

## Keybindings

| Action | Neovim | VS Code | What it does |
|--------|--------|---------|--------------|
| Go to definition | `gd` | `F12` / `Cmd+Click` | Jump to contract file from `CONTRACT:` ref |
| Go to implementation | `gi` | `Cmd+F12` | Find source files implementing a contract |
| References | `gr` | `Shift+F12` | Find contracts that reference this contract |
| Hover | `K` | Mouse hover | Show contract summary |
| Completions | Type `CONTRACT:` | Type `CONTRACT:` | Auto-suggest contract IDs |

## Troubleshooting

**LSP not attaching:** Make sure the project has a `.rebarrc` or `.rebar` directory at the root.

**"Contract not found" errors on valid refs:** Run `rebar verify` to check that your contracts exist in `architecture/`. The LSP discovers contracts from `architecture/CONTRACT-*.md` filenames.

**VS Code "server crashed":** Check the Output panel → **Rebar LSP** for error details. Most likely `rebar` on PATH points to the bash framework, not the Go CLI. Set `rebar.serverPath` to the full path of the Go binary.
