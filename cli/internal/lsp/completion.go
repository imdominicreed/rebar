package lsp

import (
	"fmt"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) textDocumentCompletion(ctx *glsp.Context, params *protocol.CompletionParams) (any, error) {
	uri := string(params.TextDocument.URI)
	content, ok := s.state.GetDocument(uri)
	if !ok {
		return nil, nil
	}

	lines := strings.Split(content, "\n")
	lineIdx := int(params.Position.Line)
	if lineIdx >= len(lines) {
		return nil, nil
	}

	line := lines[lineIdx]
	col := int(params.Position.Character)
	if col > len(line) {
		col = len(line)
	}
	prefix := line[:col]

	if !strings.Contains(prefix, "CONTRACT:") {
		return nil, nil
	}

	ns := s.state.Namespace()
	contracts := s.state.AllContracts()
	kind := protocol.CompletionItemKindReference

	var items []protocol.CompletionItem
	for _, c := range contracts {
		label := fmt.Sprintf("%s.%s", c.ID, versionFromFullID(c.Version))
		if ns != "" {
			label = ns + ":" + label
		}
		detail := c.Name
		items = append(items, protocol.CompletionItem{
			Label:  label,
			Kind:   &kind,
			Detail: &detail,
		})
	}

	return items, nil
}

// versionFromFullID extracts "1.0" from "S1-STEWARD.1.0"
func versionFromFullID(fullID string) string {
	parts := strings.Split(fullID, ".")
	if len(parts) >= 3 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return fullID
}
