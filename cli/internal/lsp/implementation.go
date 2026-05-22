package lsp

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/willackerly/rebar/cli/internal/repo"
)

func (s *Server) textDocumentImplementation(ctx *glsp.Context, params *protocol.ImplementationParams) (any, error) {
	uri := string(params.TextDocument.URI)
	content, ok := s.state.GetDocument(uri)
	if !ok {
		return nil, nil
	}

	root := s.state.RepoRoot()
	if root == "" {
		return nil, nil
	}

	// In a contract file: find all source files implementing this contract
	if isContractFile(uri) {
		contractID := contractIDFromURI(uri)
		if contractID == "" {
			return nil, nil
		}
		return findImplementingFiles(root, contractID)
	}

	// In a source file: find implementations for the CONTRACT: ref at cursor
	refs := ParseContractRefs(content, 10)
	ref := RefAtPosition(refs, params.Position)
	if ref == nil {
		return nil, nil
	}

	return findImplementingFiles(root, ref.ID)
}

// contractIDFromURI extracts the bare contract ID from a contract file URI.
// "file:///repo/architecture/CONTRACT-S2-ASK-CLI.1.0.md" → "S2-ASK-CLI"
func contractIDFromURI(uri string) string {
	path := uriToPath(uri)
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "CONTRACT-") {
		return ""
	}
	id, _ := idFromFilename(base)
	return id
}

// findImplementingFiles uses git grep to find all source files with a CONTRACT: header matching the given ID.
func findImplementingFiles(repoRoot, contractID string) ([]protocol.Location, error) {
	pattern := "CONTRACT:.*" + contractID + "\\."
	files, err := repo.TrackedFilesGrep(repoRoot, pattern,
		"*.go", "*.ts", "*.tsx", "*.js", "*.jsx", "*.py", "*.rs",
		"*.java", "*.rb", "*.sh", "*.kt", "*.c", "*.cpp")
	if err != nil {
		return nil, nil
	}

	var locations []protocol.Location
	for _, relPath := range files {
		absPath := filepath.Join(repoRoot, relPath)
		line, col := findContractLine(absPath, contractID)
		locations = append(locations, protocol.Location{
			URI: protocol.DocumentUri(fileURI(absPath)),
			Range: protocol.Range{
				Start: protocol.Position{Line: protocol.UInteger(line), Character: protocol.UInteger(col)},
				End:   protocol.Position{Line: protocol.UInteger(line), Character: protocol.UInteger(col)},
			},
		})
	}

	return locations, nil
}

// findContractLine scans the first 10 lines for the CONTRACT: header and returns its position.
func findContractLine(path, contractID string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for line := 0; scanner.Scan() && line < 10; line++ {
		text := scanner.Text()
		idx := strings.Index(text, "CONTRACT:")
		if idx >= 0 && strings.Contains(text, contractID) {
			return line, idx
		}
	}
	return 0, 0
}
