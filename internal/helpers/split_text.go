package helpers

import (
	"log/slog"
	"strings"

	"github.com/tmc/langchaingo/textsplitter"
)

func SplitTextByChunks(text string) ([]string, error) {
	split, err := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(500),
		textsplitter.WithChunkOverlap(80),
		textsplitter.WithHeadingHierarchy(true),
		textsplitter.WithModelName("gpt-5.1"),
	).SplitText(text)
	if err != nil {
		slog.Error("failed to split text", slog.String("err", err.Error()))
		return nil, err
	}

	split = removeUnusedChunks(split)

	return split, nil
}

func removeUnusedChunks(split []string) []string {
	if len(split) == 0 || len(split) == 1 {
		return split
	}

	if strings.Contains(split[len(split)-2], split[len(split)-1]) {
		return removeUnusedChunks(split[:len(split)-1])
	}

	return split
}
