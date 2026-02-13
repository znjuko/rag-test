package rag

import (
	"context"
	"log/slog"

	openairepo "rag-test/internal/repository/openai"
)

func (s *Service) chat(ctx context.Context, systemPrompt, userPrompt string, maxTokens int, stage string, responseFormat *openairepo.ResponseFormat) (string, error) {
	resp, err := s.openaiRepo.CreateChatCompletion(ctx, openairepo.ChatCompletionRequest{
		Messages: []openairepo.Message{
			{Role: openairepo.RoleSystem, Content: systemPrompt},
			{Role: openairepo.RoleUser, Content: userPrompt},
		},
		Temperature:    0,
		MaxTokens:      maxTokens,
		ResponseFormat: responseFormat,
	})
	if err != nil {
		slog.Error("openai request failed", slog.String("stage", stage), slog.String("error", err.Error()))
		return "", err
	}

	return resp.Content, nil
}
