package lsp

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestParseContractRefs_Legacy(t *testing.T) {
	content := `// CONTRACT:S2-ASK-CLI.1.0
package main
`
	refs := ParseContractRefs(content, 10)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	r := refs[0]
	if r.ID != "S2-ASK-CLI" {
		t.Errorf("ID = %q, want S2-ASK-CLI", r.ID)
	}
	if r.Major != 1 || r.Minor != 0 {
		t.Errorf("version = %d.%d, want 1.0", r.Major, r.Minor)
	}
	if r.Namespace != "" {
		t.Errorf("namespace = %q, want empty", r.Namespace)
	}
}

func TestParseContractRefs_Namespaced(t *testing.T) {
	content := `// Architecture: CONTRACT:github.com/willackerly/rebar:S2-ASK-CLI.1.0
package main
`
	refs := ParseContractRefs(content, 10)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	r := refs[0]
	if r.ID != "S2-ASK-CLI" {
		t.Errorf("ID = %q, want S2-ASK-CLI", r.ID)
	}
	if r.Namespace != "github.com/willackerly/rebar" {
		t.Errorf("namespace = %q, want github.com/willackerly/rebar", r.Namespace)
	}
	if r.Major != 1 || r.Minor != 0 {
		t.Errorf("version = %d.%d, want 1.0", r.Major, r.Minor)
	}
}

func TestParseContractRefs_NoRefs(t *testing.T) {
	content := `package main

func main() {}
`
	refs := ParseContractRefs(content, 10)
	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}

func TestParseContractRefs_MaxLines(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5\n// CONTRACT:FOO.1.0\n"
	// Only scan first 3 lines — the ref on line 6 should be missed
	refs := ParseContractRefs(content, 3)
	if len(refs) != 0 {
		t.Errorf("expected 0 refs (beyond maxLines), got %d", len(refs))
	}

	// Scan all 6 lines — should find it
	refs = ParseContractRefs(content, 10)
	if len(refs) != 1 {
		t.Errorf("expected 1 ref, got %d", len(refs))
	}
}

func TestParseContractRefs_Multiple(t *testing.T) {
	content := `// CONTRACT:FOO.1.0
// CONTRACT:BAR.2.1
`
	refs := ParseContractRefs(content, 10)
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
	if refs[0].ID != "FOO" {
		t.Errorf("refs[0].ID = %q, want FOO", refs[0].ID)
	}
	if refs[1].ID != "BAR" {
		t.Errorf("refs[1].ID = %q, want BAR", refs[1].ID)
	}
}

func TestParseContractRefs_PositionAccuracy(t *testing.T) {
	content := `// Architecture: CONTRACT:github.com/foo/bar:S1-TEST.1.0`
	refs := ParseContractRefs(content, 10)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	r := refs[0]
	if r.Range.Start.Line != 0 {
		t.Errorf("start line = %d, want 0", r.Range.Start.Line)
	}
	// "CONTRACT:" starts at position 17
	if r.Range.Start.Character != 17 {
		t.Errorf("start char = %d, want 17", r.Range.Start.Character)
	}
}

func TestRefAtPosition(t *testing.T) {
	content := `// CONTRACT:FOO.1.0
// CONTRACT:BAR.2.1
`
	refs := ParseContractRefs(content, 10)

	// Position on first ref
	ref := RefAtPosition(refs, protocol.Position{Line: 0, Character: 5})
	if ref == nil {
		t.Fatal("expected ref at line 0")
	}
	if ref.ID != "FOO" {
		t.Errorf("ID = %q, want FOO", ref.ID)
	}

	// Position on second ref
	ref = RefAtPosition(refs, protocol.Position{Line: 1, Character: 5})
	if ref == nil {
		t.Fatal("expected ref at line 1")
	}
	if ref.ID != "BAR" {
		t.Errorf("ID = %q, want BAR", ref.ID)
	}

	// Position outside any ref
	ref = RefAtPosition(refs, protocol.Position{Line: 2, Character: 0})
	if ref != nil {
		t.Error("expected nil for line outside refs")
	}
}

func TestExtractBareID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"S1-STEWARD.1.0", "S1-STEWARD"},
		{"github.com/foo/bar:S1-STEWARD.1.0", "S1-STEWARD"},
		{"SIMPLE", "SIMPLE"},
		{"S2-ASK-CLI.2.3", "S2-ASK-CLI"},
	}
	for _, tt := range tests {
		got := extractBareID(tt.input)
		if got != tt.want {
			t.Errorf("extractBareID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractFilePath(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"- src/foo/bar.go", "src/foo/bar.go"},
		{"- `src/foo/bar.go`", "src/foo/bar.go"},
		{"- `bin/ask` — primary CLI entry point", "bin/ask"},
		{"- `agents/<role>/commands/*.sh` — implementations", "agents/<role>/commands/*.sh"},
		{"* internal/pkg/thing.ts", "internal/pkg/thing.ts"},
		{"- src/main.go — the entry point", "src/main.go"},
		{"not a list item", ""},
		{"- just a word", ""},
	}
	for _, tt := range tests {
		got := extractFilePath(tt.line)
		if got != tt.want {
			t.Errorf("extractFilePath(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestIdFromFilename(t *testing.T) {
	tests := []struct {
		input   string
		wantID  string
		wantVer string
	}{
		{"CONTRACT-S2-ASK-CLI.1.0.md", "S2-ASK-CLI", "S2-ASK-CLI.1.0"},
		{"CONTRACT-S1-STEWARD.1.0.md", "S1-STEWARD", "S1-STEWARD.1.0"},
		{"CONTRACT-S3-MCP-SERVER.1.0.md", "S3-MCP-SERVER", "S3-MCP-SERVER.1.0"},
	}
	for _, tt := range tests {
		id, ver := idFromFilename(tt.input)
		if id != tt.wantID || ver != tt.wantVer {
			t.Errorf("idFromFilename(%q) = (%q, %q), want (%q, %q)", tt.input, id, ver, tt.wantID, tt.wantVer)
		}
	}
}

func TestContractRef_FullID(t *testing.T) {
	r := ContractRef{Namespace: "github.com/foo/bar", ID: "S1-TEST"}
	if got := r.FullID(); got != "github.com/foo/bar:S1-TEST" {
		t.Errorf("FullID() = %q", got)
	}

	r2 := ContractRef{ID: "S1-TEST"}
	if got := r2.FullID(); got != "S1-TEST" {
		t.Errorf("FullID() = %q", got)
	}
}
