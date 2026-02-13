package rag

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
)

type clarificationResult struct {
	NeedClarification  bool     `json:"need_clarification"`
	ClarifyingQuestion *string  `json:"clarifying_question"`
	MissingSlots       []string `json:"missing_slots"`
	Assumptions        []string `json:"assumptions"`
}

type rewriteResult struct {
	Queries []string `json:"queries"`
}

type answerResult struct {
	Text          string     `json:"text"`
	CitationsUsed []string   `json:"citations_used"`
	Citations     []Citation `json:"citations"`
}

type validationPayload struct {
	OK                bool     `json:"ok"`
	UnsupportedClaims []string `json:"unsupported_claims"`
	Notes             string   `json:"notes"`
}

func normalizeAnswer(parsed answerResult) answerResult {
	parsed.Text = strings.TrimSpace(parsed.Text)
	if parsed.Text == "" {
		return answerResult{Text: unknownAnswer}
	}
	if parsed.Text == unknownAnswer {
		parsed.CitationsUsed = nil
		parsed.Citations = nil
		return parsed
	}
	if len(parsed.CitationsUsed) == 0 || len(parsed.Citations) == 0 {
		return answerResult{Text: unknownAnswer}
	}
	return parsed
}

func parseAnswer(content, userPrompt, chunks, question, stage string) (answerResult, error) {
	var parsed answerResult
	if err := decodeJSON(content, &parsed); err != nil {
		slog.Error(
			"failed to parse answer response",
			slog.String("error", err.Error()),
			slog.String("stage", stage),
			slog.String("content", content),
			slog.String("user_prompt", userPrompt),
			slog.String("chunks", chunks),
			slog.String("question", question),
		)
		return answerResult{}, err
	}

	return normalizeAnswer(parsed), nil
}

func buildValidationFeedback(validation ValidationResult) string {
	payload := struct {
		Notes             string   `json:"notes"`
		UnsupportedClaims []string `json:"unsupported_claims"`
	}{
		Notes:             strings.TrimSpace(validation.Notes),
		UnsupportedClaims: validation.UnsupportedClaims,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		if payload.Notes != "" {
			return payload.Notes
		}
		return "no feedback"
	}

	return string(data)
}

func (s *Service) checkClarification(ctx context.Context, question, dialogContext string) (clarificationResult, error) {
	userPrompt := buildClarificationUserPrompt(question, dialogContext)
	content, err := s.chat(ctx, analysisSystemPrompt, userPrompt, analysisMaxTokens, "clarification", jsonObjectResponseFormat)
	if err != nil {
		return clarificationResult{}, err
	}

	var parsed clarificationResult
	if err := decodeJSON(content, &parsed); err != nil {
		slog.Error("failed to parse clarification response", slog.String("error", err.Error()))
		return clarificationResult{}, err
	}

	return parsed, nil
}

func (s *Service) rewriteQueries(ctx context.Context, question, dialogContext string, analysis clarificationResult) (rewriteResult, error) {
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return rewriteResult{}, err
	}

	userPrompt := buildRewriteUserPrompt(question, dialogContext, string(analysisJSON))
	content, err := s.chat(ctx, rewriteSystemPrompt, userPrompt, rewriteMaxTokens, "rewrite", jsonObjectResponseFormat)
	if err != nil {
		return rewriteResult{}, err
	}

	var parsed rewriteResult
	if err := decodeJSON(content, &parsed); err != nil {
		slog.Error("failed to parse rewrite response", slog.String("error", err.Error()))
		return rewriteResult{}, err
	}

	return parsed, nil
}

func (s *Service) fetchChunks(ctx context.Context, question string, topK int) ([]Chunk, error) {
	if topK <= 0 {
		topK = s.defaultTopK
	}

	vectors, err := s.embeddingsRepo.CreateEmbeddings(ctx, question)
	if err != nil {
		slog.Error("failed to create embeddings", slog.String("error", err.Error()))
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, errors.New("empty embeddings")
	}

	hits, err := s.vectorRepo.Search(ctx, s.collection, vectors[0], topK)
	if err != nil {
		slog.Error("failed to search chunks", slog.String("error", err.Error()))
		return nil, err
	}

	return buildChunks(hits), nil
}

func (s *Service) generateAnswer(ctx context.Context, question, dialogContext, chunks string) (answerResult, error) {
	userPrompt := buildAnswerUserPrompt(question, dialogContext, chunks)
	content, err := s.chat(ctx, answerSystemPrompt, userPrompt, answerMaxTokens, "answer", jsonObjectResponseFormat)
	if err != nil {
		return answerResult{}, err
	}

	return parseAnswer(content, userPrompt, chunks, question, "answer")
}

func (s *Service) rewriteAnswer(ctx context.Context, question, dialogContext, chunks, answerText string, validation ValidationResult) (answerResult, error) {
	feedback := buildValidationFeedback(validation)
	userPrompt := buildAnswerRewriteUserPrompt(question, dialogContext, chunks, answerText, feedback)
	content, err := s.chat(ctx, answerRewriteSystemPrompt, userPrompt, answerMaxTokens, "answer_rewrite", jsonObjectResponseFormat)
	if err != nil {
		return answerResult{}, err
	}

	return parseAnswer(content, userPrompt, chunks, question, "answer_rewrite")
}

func (s *Service) validateAnswer(ctx context.Context, question, answerText, chunks string) (ValidationResult, error) {
	trimmedAnswer := strings.TrimSpace(answerText)
	if trimmedAnswer == "" {
		slog.Warn("skip validation: empty answer text")
		return ValidationResult{
			OK:    false,
			Notes: "empty answer text",
		}, nil
	}
	if trimmedAnswer == unknownAnswer {
		return ValidationResult{
			OK:    true,
			Notes: "validation skipped for unknown answer",
		}, nil
	}

	userPrompt := buildValidationUserPrompt(question, answerText, chunks)
	content, err := s.chat(ctx, validationSystemPrompt, userPrompt, validationMaxTokens, "validation", jsonObjectResponseFormat)
	if err != nil {
		return ValidationResult{}, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		slog.Warn("validation returned empty response")
		return ValidationResult{
			OK:    false,
			Notes: "empty validator response",
		}, nil
	}

	var parsed validationPayload
	if err := decodeJSON(content, &parsed); err != nil {
		slog.Warn(
			"failed to parse validation response",
			slog.String("error", err.Error()),
			slog.String("content", content),
			slog.String("user_prompt", userPrompt),
			slog.String("chunks", chunks),
			slog.String("question", question),
			slog.String("answer_text", answerText),
		)
		return ValidationResult{
			OK:    false,
			Notes: "invalid validator response",
		}, nil
	}

	return ValidationResult{
		OK:                parsed.OK,
		UnsupportedClaims: copyStrings(parsed.UnsupportedClaims),
		Notes:             parsed.Notes,
	}, nil
}
