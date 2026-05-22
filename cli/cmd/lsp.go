package cmd

import (
	"github.com/spf13/cobra"
	"github.com/willackerly/rebar/cli/internal/lsp"
)

var lspStdio bool

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the Language Server Protocol server",
	Long: `Starts a JSON-RPC 2.0 LSP server over stdio for editor integration.

Provides diagnostics, go-to-definition, completions, and hover
for CONTRACT: references in source files.

Configure your editor to run "rebar lsp" as a language server:

  Neovim:
    vim.lsp.start({
      name = "rebar",
      cmd = { "rebar", "lsp" },
      root_dir = vim.fs.dirname(
        vim.fs.find({ ".rebarrc", ".rebar" }, { upward = true })[1]
      ),
    })

  VS Code (settings.json):
    Add rebar to your LSP client extension configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return lsp.Serve(Version)
	},
}

func init() {
	lspCmd.Flags().BoolVar(&lspStdio, "stdio", false, "use stdio transport (default, accepted for editor compatibility)")
	lspCmd.Flags().String("clientProcessId", "", "editor client process ID (accepted for editor compatibility)")
}
