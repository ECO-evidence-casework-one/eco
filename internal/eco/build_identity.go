package eco

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// SourceCommit is overridden by controlled builds with -ldflags -X so the
// executable can identify the exact source candidate that produced it. Local
// developer builds without an injected revision remain visibly unrecorded.
var SourceCommit = "local-unrecorded"

func DevelopmentWorkspaceRoot(base, buildID, sourceCommit string, schema int) string {
	return filepath.Join(
		base,
		"EvidenceCaseworkOne",
		"Development",
		fmt.Sprintf("schema-%d", schema),
		developmentStateComponent(buildID),
		developmentSourceComponent(sourceCommit),
	)
}

func DefaultDevelopmentWorkspaceRoot(base string) string {
	return DevelopmentWorkspaceRoot(base, BuildID, SourceCommit, Schema)
}

func developmentSourceComponent(sourceCommit string) string {
	value := strings.TrimSpace(sourceCommit)
	if value == "" {
		value = "local-unrecorded"
	}
	if value != "local-unrecorded" && len(value) > 12 {
		value = value[:12]
	}
	return developmentStateComponent(value)
}

func developmentStateComponent(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	clean := strings.Trim(b.String(), " ._")
	if clean == "" {
		clean = "unknown"
	}
	if len(clean) > 80 {
		clean = clean[:80]
	}
	return clean
}
