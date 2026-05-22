package lsp

import (
	"regexp"
	"strconv"
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

var (
	legacyRefRE     = regexp.MustCompile(`CONTRACT:([A-Z][A-Za-z0-9_-]*)\.(\d+)\.(\d+)\b`)
	namespacedRefRE = regexp.MustCompile(`CONTRACT:([a-zA-Z0-9][a-zA-Z0-9_./-]+):([A-Z][A-Za-z0-9_-]*)\.(\d+)\.(\d+)\b`)
)

type ContractRef struct {
	Namespace string
	ID        string
	Major     int
	Minor     int
	Range     protocol.Range
	RawText   string
}

func (r ContractRef) FullID() string {
	if r.Namespace != "" {
		return r.Namespace + ":" + r.ID
	}
	return r.ID
}

func (r ContractRef) VersionedID() string {
	return r.ID + "." + strconv.Itoa(r.Major) + "." + strconv.Itoa(r.Minor)
}

// ParseContractRefs scans the first maxLines lines for CONTRACT: references.
func ParseContractRefs(content string, maxLines int) []ContractRef {
	lines := strings.SplitN(content, "\n", maxLines+1)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var refs []ContractRef

	for lineNum, line := range lines {
		// Try namespaced first (more specific)
		for _, match := range namespacedRefRE.FindAllStringSubmatchIndex(line, -1) {
			ns := line[match[2]:match[3]]
			id := line[match[4]:match[5]]
			major, _ := strconv.Atoi(line[match[6]:match[7]])
			minor, _ := strconv.Atoi(line[match[8]:match[9]])
			refs = append(refs, ContractRef{
				Namespace: ns,
				ID:        id,
				Major:     major,
				Minor:     minor,
				Range: protocol.Range{
					Start: protocol.Position{Line: protocol.UInteger(lineNum), Character: protocol.UInteger(match[0])},
					End:   protocol.Position{Line: protocol.UInteger(lineNum), Character: protocol.UInteger(match[1])},
				},
				RawText: line[match[0]:match[1]],
			})
		}

		// Legacy refs (only if not already captured as namespaced)
		for _, match := range legacyRefRE.FindAllStringSubmatchIndex(line, -1) {
			raw := line[match[0]:match[1]]
			if isAlreadyCaptured(refs, lineNum, match[0]) {
				continue
			}
			id := line[match[2]:match[3]]
			major, _ := strconv.Atoi(line[match[4]:match[5]])
			minor, _ := strconv.Atoi(line[match[6]:match[7]])
			refs = append(refs, ContractRef{
				ID:    id,
				Major: major,
				Minor: minor,
				Range: protocol.Range{
					Start: protocol.Position{Line: protocol.UInteger(lineNum), Character: protocol.UInteger(match[0])},
					End:   protocol.Position{Line: protocol.UInteger(lineNum), Character: protocol.UInteger(match[1])},
				},
				RawText: raw,
			})
		}
	}

	return refs
}

func isAlreadyCaptured(refs []ContractRef, line, col int) bool {
	for _, r := range refs {
		if int(r.Range.Start.Line) == line &&
			int(r.Range.Start.Character) <= col &&
			int(r.Range.End.Character) > col {
			return true
		}
	}
	return false
}

// RefAtPosition returns the ContractRef at the given position, or nil.
func RefAtPosition(refs []ContractRef, pos protocol.Position) *ContractRef {
	for i := range refs {
		r := &refs[i]
		if pos.Line == r.Range.Start.Line &&
			pos.Character >= r.Range.Start.Character &&
			pos.Character <= r.Range.End.Character {
			return r
		}
	}
	return nil
}
