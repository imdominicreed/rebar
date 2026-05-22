package lsp

import (
	"fmt"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) textDocumentHover(ctx *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := string(params.TextDocument.URI)
	content, ok := s.state.GetDocument(uri)
	if !ok {
		return nil, nil
	}

	refs := ParseContractRefs(content, 10)
	ref := RefAtPosition(refs, params.Position)
	if ref == nil {
		return nil, nil
	}

	info := s.state.ResolveRef(*ref)
	if info == nil {
		return nil, nil
	}

	var md strings.Builder
	fmt.Fprintf(&md, "**%s** (v%s)\n\n", info.ID, versionFromFullID(info.Version))

	if info.Name != "" {
		fmt.Fprintf(&md, "%s\n\n", info.Name)
	}

	if summary, ok := info.Sections["Why this exists"]; ok && summary != "" {
		text := truncate(summary, 300)
		fmt.Fprintf(&md, "---\n\n%s\n", text)
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: md.String(),
		},
		Range: &ref.Range,
	}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
