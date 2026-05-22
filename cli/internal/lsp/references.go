package lsp

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) textDocumentReferences(ctx *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	uri := string(params.TextDocument.URI)
	content, ok := s.state.GetDocument(uri)
	if !ok {
		return nil, nil
	}

	// In a contract file: find other contracts referencing this one
	if isContractFile(uri) {
		contractID := contractIDFromURI(uri)
		if contractID == "" {
			return nil, nil
		}
		return s.findContractReferences(contractID)
	}

	// In a source file: find contract references for the CONTRACT: ref at cursor
	refs := ParseContractRefs(content, 10)
	ref := RefAtPosition(refs, params.Position)
	if ref == nil {
		return nil, nil
	}

	return s.findContractReferences(ref.ID)
}

// findContractReferences searches contract markdown files for references to the given contract ID.
func (s *Server) findContractReferences(contractID string) ([]protocol.Location, error) {
	root := s.state.RepoRoot()
	contractDir := filepath.Join(root, "architecture")

	entries, err := os.ReadDir(contractDir)
	if err != nil {
		return nil, nil
	}

	var locations []protocol.Location
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "CONTRACT-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		if strings.Contains(name, "TEMPLATE") || strings.Contains(name, "REGISTRY") {
			continue
		}
		// Skip the contract's own file
		if strings.Contains(name, contractID) {
			continue
		}

		path := filepath.Join(contractDir, name)
		locs := findIDInFile(path, contractID)
		locations = append(locations, locs...)
	}

	return locations, nil
}

// findIDInFile scans a file for lines referencing a contract ID and returns their locations.
func findIDInFile(path, contractID string) []protocol.Location {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var locs []protocol.Location
	for lineNum, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, contractID) {
			col := strings.Index(line, contractID)
			locs = append(locs, protocol.Location{
				URI: protocol.DocumentUri(fileURI(path)),
				Range: protocol.Range{
					Start: protocol.Position{Line: protocol.UInteger(lineNum), Character: protocol.UInteger(col)},
					End:   protocol.Position{Line: protocol.UInteger(lineNum), Character: protocol.UInteger(col + len(contractID))},
				},
			})
		}
	}

	return locs
}
