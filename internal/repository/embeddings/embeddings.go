package embeddings

import (
	"context"
	"log/slog"

	"github.com/tmc/langchaingo/llms/openai"
)

type Repository struct {
	cli *openai.LLM
}

func NewRepository(token string) (*Repository, error) {
	opts := []openai.Option{
		openai.WithToken(token),
		openai.WithModel(modelName),
		openai.WithEmbeddingModel(embeddingModelName),
		openai.WithEmbeddingDimensions(defaultDim),
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	return &Repository{cli: llm}, nil
}

func (r *Repository) CreateEmbeddings(ctx context.Context, data ...string) ([][]float32, error) {
	embeddings, err := r.cli.CreateEmbedding(ctx, data)
	if err != nil {
		slog.Error("failed to create embeddings", slog.String("error", err.Error()))
		return nil, err
	}

	return embeddings, nil
}
