package lsp

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspserver "github.com/tliron/glsp/server"
)

const serverName = "rebar-lsp"

type Server struct {
	state   *WorkspaceState
	handler protocol.Handler
	version string
}

func NewServer(version string) *Server {
	s := &Server{
		state:   NewWorkspaceState(),
		version: version,
	}
	s.handler = protocol.Handler{
		Initialize:             s.initialize,
		Initialized:            s.initialized,
		Shutdown:               s.shutdown,
		SetTrace:               s.setTrace,
		TextDocumentDidOpen:    s.textDocumentDidOpen,
		TextDocumentDidChange:  s.textDocumentDidChange,
		TextDocumentDidSave:    s.textDocumentDidSave,
		TextDocumentDidClose:   s.textDocumentDidClose,
		TextDocumentCompletion: s.textDocumentCompletion,
		TextDocumentHover:      s.textDocumentHover,
		TextDocumentDefinition:     s.textDocumentDefinition,
		TextDocumentImplementation: s.textDocumentImplementation,
		TextDocumentReferences:     s.textDocumentReferences,
	}
	return s
}

func Serve(version string) error {
	commonlog.Configure(0, nil)

	s := NewServer(version)
	srv := glspserver.NewServer(&s.handler, serverName, false)
	return srv.RunStdio()
}

func (s *Server) initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := s.handler.CreateServerCapabilities()

	syncKind := protocol.TextDocumentSyncKindFull
	capabilities.TextDocumentSync = &protocol.TextDocumentSyncOptions{
		OpenClose: &protocol.True,
		Change:    &syncKind,
		Save:      &protocol.True,
	}

	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{":"},
	}

	// Initialize workspace from rootUri
	if params.RootURI != nil {
		rootPath := uriToPath(string(*params.RootURI))
		if rootPath != "" {
			if err := s.state.Init(rootPath); err != nil {
				logToStderr("init error: %v", err)
			}
		}
	} else if params.RootPath != nil {
		if err := s.state.Init(*params.RootPath); err != nil {
			logToStderr("init error: %v", err)
		}
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    serverName,
			Version: &s.version,
		},
	}, nil
}

func (s *Server) initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func (s *Server) shutdown(ctx *glsp.Context) error {
	return nil
}

func (s *Server) setTrace(ctx *glsp.Context, params *protocol.SetTraceParams) error {
	return nil
}

func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		parsed, err := url.Parse(uri)
		if err != nil {
			return strings.TrimPrefix(uri, "file://")
		}
		return parsed.Path
	}
	return uri
}

func logToStderr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[rebar-lsp] "+format+"\n", args...)
}
