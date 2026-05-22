package lsp

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) textDocumentDefinition(ctx *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	uri := string(params.TextDocument.URI)
	content, ok := s.state.GetDocument(uri)
	if !ok {
		return nil, nil
	}

	if isContractFile(uri) {
		return s.definitionFromContract(content, params.Position)
	}

	return s.definitionFromSource(content, params.Position)
}

func (s *Server) definitionFromSource(content string, pos protocol.Position) (any, error) {
	refs := ParseContractRefs(content, 10)
	ref := RefAtPosition(refs, pos)
	if ref == nil {
		return nil, nil
	}

	info := s.state.ResolveRef(*ref)
	if info == nil {
		return nil, nil
	}

	return protocol.Location{
		URI: protocol.DocumentUri(fileURI(info.Path)),
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 0},
		},
	}, nil
}

func (s *Server) definitionFromContract(content string, pos protocol.Position) (any, error) {
	lines := strings.Split(content, "\n")
	lineIdx := int(pos.Line)
	if lineIdx >= len(lines) {
		return nil, nil
	}

	line := lines[lineIdx]
	path := extractFilePath(line)
	if path == "" {
		return nil, nil
	}

	root := s.state.RepoRoot()
	absPath := filepath.Join(root, path)

	return protocol.Location{
		URI: protocol.DocumentUri(fileURI(absPath)),
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 0},
		},
	}, nil
}

// extractFilePath tries to pull a file path from a markdown list item.
// Handles:
//   - "- src/foo/bar.go"
//   - "- `bin/ask` — description"
//   - "* path/to/file.ts"
//   - "- `agents/<role>/commands/*.sh` — note"
func extractFilePath(line string) string {
	trimmed := strings.TrimSpace(line)

	// Must be a list item
	if !strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(trimmed, "* ") {
		return ""
	}
	trimmed = strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
	trimmed = strings.TrimSpace(trimmed)

	// Extract backtick-wrapped path: `path/to/file`
	if strings.HasPrefix(trimmed, "`") {
		end := strings.Index(trimmed[1:], "`")
		if end >= 0 {
			trimmed = trimmed[1 : end+1]
		}
	} else {
		// No backticks — take everything before " —" or " -" separator
		for _, sep := range []string{" — ", " — ", " - ", "  "} {
			if idx := strings.Index(trimmed, sep); idx > 0 {
				trimmed = trimmed[:idx]
			}
		}
	}

	trimmed = strings.TrimSpace(trimmed)

	// Must contain a slash or dot to look like a path
	if trimmed == "" || (!strings.Contains(trimmed, "/") && !strings.Contains(trimmed, ".")) {
		return ""
	}

	return trimmed
}

func fileURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := &url.URL{Scheme: "file", Path: abs}
	return u.String()
}
