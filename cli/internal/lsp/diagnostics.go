package lsp

import (
	"fmt"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const diagnosticSource = "rebar"

func (s *Server) publishDiagnostics(notify glsp.NotifyFunc, uri string, content string) {
	diags := s.computeDiagnostics(uri, content)
	notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         protocol.DocumentUri(uri),
		Diagnostics: diags,
	})
}

func (s *Server) computeDiagnostics(uri string, content string) []protocol.Diagnostic {
	diags := []protocol.Diagnostic{}

	if isContractFile(uri) {
		return diags
	}

	refs := ParseContractRefs(content, 10)
	if len(refs) == 0 {
		return diags
	}
	source := diagnosticSource

	for _, ref := range refs {
		info := s.state.ResolveRef(ref)
		if info == nil {
			severity := protocol.DiagnosticSeverityError
			diags = append(diags, protocol.Diagnostic{
				Range:    ref.Range,
				Severity: &severity,
				Source:   &source,
				Message:  fmt.Sprintf("Contract %q not found in architecture/", ref.ID),
			})
		}
	}

	return diags
}

func isContractFile(uri string) bool {
	return strings.Contains(uri, "architecture/CONTRACT-")
}

func (s *Server) textDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	content := params.TextDocument.Text
	s.state.SetDocument(uri, content)
	s.publishDiagnostics(ctx.Notify, uri, content)
	return nil
}

func (s *Server) textDocumentDidSave(ctx *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	uri := string(params.TextDocument.URI)

	if isContractFile(uri) {
		if err := s.state.Refresh(); err != nil {
			return nil
		}
		// Re-diagnose all open documents after contract index refresh
		s.rediagnoseOpenDocuments(ctx.Notify)
		return nil
	}

	content, ok := s.state.GetDocument(uri)
	if ok {
		s.publishDiagnostics(ctx.Notify, uri, content)
	}
	return nil
}

func (s *Server) textDocumentDidChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	for _, change := range params.ContentChanges {
		if c, ok := change.(protocol.TextDocumentContentChangeEventWhole); ok {
			s.state.SetDocument(uri, c.Text)
		}
	}
	return nil
}

func (s *Server) textDocumentDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	s.state.RemoveDocument(uri)
	// Clear diagnostics
	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, &protocol.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []protocol.Diagnostic{},
	})
	return nil
}

func (s *Server) rediagnoseOpenDocuments(notify glsp.NotifyFunc) {
	s.state.mu.RLock()
	docs := make(map[string]string, len(s.state.documents))
	for uri, content := range s.state.documents {
		docs[uri] = content
	}
	s.state.mu.RUnlock()

	for uri, content := range docs {
		s.publishDiagnostics(notify, uri, content)
	}
}
