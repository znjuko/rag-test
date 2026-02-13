package rag

import (
	"fmt"
	"strings"

	milvusrepo "rag-test/internal/repository/milvus"
)

func buildChunks(hits []milvusrepo.SearchHit) []Chunk {
	chunks := make([]Chunk, 0, len(hits))
	for i, hit := range hits {
		chunks = append(chunks, Chunk{
			ID:         fmt.Sprintf("C%d", i+1),
			DataSource: hit.DataSource,
			Text:       hit.Payload,
		})
	}
	return chunks
}

func formatChunks(chunks []Chunk) string {
	if len(chunks) == 0 {
		return ""
	}

	lines := make([]string, 0, len(chunks)*4)
	for _, chunk := range chunks {
		lines = append(lines, fmt.Sprintf("[%s]", chunk.ID))
		lines = append(lines, fmt.Sprintf("data_source: %s", chunk.DataSource))
		lines = append(lines, fmt.Sprintf("text: %s", chunk.Text))
		lines = append(lines, "")
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
