package rag

import (
	"fmt"
	"strings"
)

func formatDialogContext(history []DialogMessage) string {
	if len(history) == 0 {
		return ""
	}

	lines := make([]string, 0, len(history))
	for _, msg := range history {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		role := strings.TrimSpace(string(msg.Role))
		if role == "" {
			role = "user"
		}
		lines = append(lines, fmt.Sprintf("%s: %s", role, content))
	}

	return strings.Join(lines, "\n")
}

func resolveDialogContext(req Request) string {
	context := strings.TrimSpace(req.DialogContext)
	if context != "" {
		return context
	}

	return formatDialogContext(req.History)
}

func copyStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	out := make([]string, len(values))
	copy(out, values)
	return out
}

func copyChunks(values []Chunk) []Chunk {
	if len(values) == 0 {
		return nil
	}

	out := make([]Chunk, len(values))
	copy(out, values)
	return out
}

func copyCitations(values []Citation) []Citation {
	if len(values) == 0 {
		return nil
	}

	out := make([]Citation, len(values))
	copy(out, values)
	return out
}

func trimOptional(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}
