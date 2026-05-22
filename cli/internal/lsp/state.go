package lsp

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/willackerly/rebar/cli/internal/config"
	"github.com/willackerly/rebar/cli/internal/repo"
	"github.com/willackerly/rebar/cli/internal/spec"
)

type ContractInfo struct {
	ID       string
	Version  string
	Name     string
	Path     string
	Sections map[string]string
}

type WorkspaceState struct {
	mu          sync.RWMutex
	repoRoot    string
	cfg         *config.Config
	namespace   string
	contracts   map[string]*ContractInfo // bare ID → info
	contractDir string
	documents   map[string]string // URI → content for open documents
}

func NewWorkspaceState() *WorkspaceState {
	return &WorkspaceState{
		contracts: make(map[string]*ContractInfo),
		documents: make(map[string]string),
	}
}

func (ws *WorkspaceState) Init(rootPath string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	root, err := config.FindRepoRoot(rootPath)
	if err != nil {
		return fmt.Errorf("not a rebar repo: %w", err)
	}
	ws.repoRoot = root
	ws.contractDir = filepath.Join(root, "architecture")

	ws.cfg, err = config.Load(root)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ws.namespace = ws.cfg.ContractNamespace
	if ws.namespace == "" {
		ws.namespace, _ = repo.InferNamespace(root)
	}

	return ws.refreshLocked()
}

func (ws *WorkspaceState) Refresh() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.refreshLocked()
}

func (ws *WorkspaceState) refreshLocked() error {
	ws.contracts = make(map[string]*ContractInfo)

	paths, err := spec.FindContracts(ws.contractDir, nil)
	if err != nil {
		return err
	}

	for _, path := range paths {
		base := filepath.Base(path)
		if strings.Contains(base, "TEMPLATE") || strings.HasSuffix(base, ".impl.md") || base == "CONTRACT-REGISTRY.md" {
			continue
		}

		c, err := spec.ParseContract(path)
		if err != nil {
			continue
		}

		id := extractBareID(c.ID)
		version := c.ID

		// Fallback: spec.ParseContract's regex doesn't match namespaced titles,
		// so extract ID from filename: CONTRACT-S2-ASK-CLI.1.0.md → S2-ASK-CLI.1.0
		if id == "" {
			id, version = idFromFilename(base)
		}
		if id == "" {
			continue
		}

		ws.contracts[id] = &ContractInfo{
			ID:       id,
			Version:  version,
			Name:     c.Name,
			Path:     path,
			Sections: c.Sections,
		}
	}

	return nil
}

// idFromFilename extracts bare ID and full versioned ID from a contract filename.
// "CONTRACT-S2-ASK-CLI.1.0.md" → ("S2-ASK-CLI", "S2-ASK-CLI.1.0")
func idFromFilename(base string) (string, string) {
	name := strings.TrimPrefix(base, "CONTRACT-")
	name = strings.TrimSuffix(name, ".md")
	// name is now e.g. "S2-ASK-CLI.1.0"
	id := extractBareID(name)
	return id, name
}

// extractBareID strips any namespace prefix and version suffix from a contract ID.
// "github.com/foo/bar:S1-STEWARD.1.0" → "S1-STEWARD"
// "S1-STEWARD.1.0" → "S1-STEWARD"
func extractBareID(raw string) string {
	// Strip namespace prefix
	if idx := strings.LastIndex(raw, ":"); idx >= 0 {
		raw = raw[idx+1:]
	}
	// Strip version suffix (.MAJOR.MINOR)
	parts := strings.Split(raw, ".")
	if len(parts) >= 3 {
		return strings.Join(parts[:len(parts)-2], ".")
	}
	return raw
}

func (ws *WorkspaceState) ResolveRef(ref ContractRef) *ContractInfo {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.contracts[ref.ID]
}

func (ws *WorkspaceState) AllContracts() []*ContractInfo {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	result := make([]*ContractInfo, 0, len(ws.contracts))
	for _, c := range ws.contracts {
		result = append(result, c)
	}
	return result
}

func (ws *WorkspaceState) Namespace() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.namespace
}

func (ws *WorkspaceState) RepoRoot() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.repoRoot
}

func (ws *WorkspaceState) SetDocument(uri, content string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.documents[uri] = content
}

func (ws *WorkspaceState) GetDocument(uri string) (string, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	content, ok := ws.documents[uri]
	return content, ok
}

func (ws *WorkspaceState) RemoveDocument(uri string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	delete(ws.documents, uri)
}
