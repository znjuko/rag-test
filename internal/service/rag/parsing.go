package rag

import (
	"encoding/json"
	"strings"
)

func decodeJSON(raw string, target any) error {
	trimmed := strings.TrimSpace(raw)
	if err := json.Unmarshal([]byte(trimmed), target); err == nil {
		return nil
	} else {
		start := strings.Index(trimmed, "{")
		end := strings.LastIndex(trimmed, "}")
		if start == -1 || end == -1 || end <= start {
			return err
		}
		return json.Unmarshal([]byte(trimmed[start:end+1]), target)
	}
}
