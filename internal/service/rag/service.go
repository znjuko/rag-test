package rag

import (
	"context"
	"errors"
	"strings"

	"rag-test/internal/repository/embeddings"
	milvusrepo "rag-test/internal/repository/milvus"
	openairepo "rag-test/internal/repository/openai"
)

type Service struct {
	openaiRepo     *openairepo.Repository
	embeddingsRepo *embeddings.Repository
	vectorRepo     milvusrepo.VectorRepository
	collection     string
	defaultTopK    int
}

func NewService(
	openaiRepo *openairepo.Repository,
	embeddingsRepo *embeddings.Repository,
	vectorRepo milvusrepo.VectorRepository,
	collection string,
	topK int,
) *Service {
	if topK <= 0 {
		topK = defaultTopK
	}

	return &Service{
		openaiRepo:     openaiRepo,
		embeddingsRepo: embeddingsRepo,
		vectorRepo:     vectorRepo,
		collection:     collection,
		defaultTopK:    topK,
	}
}

func (s *Service) Answer(ctx context.Context, req Request) (*Response, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, errors.New("question is empty")
	}

	dialogContext := resolveDialogContext(req)

	chunks, err := s.fetchChunks(ctx, question, req.TopK)
	if err != nil {
		return nil, err
	}

	//clarification, err := s.checkClarification(ctx, question, dialogContext)
	//if err != nil {
	//	return nil, err
	//}

	response := &Response{
		//NeedClarification:  clarification.NeedClarification,
		//ClarifyingQuestion: trimOptional(clarification.ClarifyingQuestion),
		//MissingSlots:       copyStrings(clarification.MissingSlots),
		//Assumptions:        copyStrings(clarification.Assumptions),
	}

	//if clarification.NeedClarification {
	//	rewrite, err := s.rewriteQueries(ctx, question, dialogContext, clarification)
	//	if err != nil {
	//		return nil, err
	//	}
	//	response.SuggestedQueries = copyStrings(rewrite.Queries)
	//	return response, nil
	//}

	response.Chunks = copyChunks(chunks)

	chunksText := formatChunks(chunks)
	answer, err := s.generateAnswer(ctx, question, dialogContext, chunksText)
	if err != nil {
		return nil, err
	}

	response.Answer = strings.TrimSpace(answer.Text)
	response.CitationsUsed = copyStrings(answer.CitationsUsed)
	response.Citations = copyCitations(answer.Citations)

	validation, err := s.validateAnswer(ctx, question, response.Answer, chunksText)
	if err != nil {
		return nil, err
	}
	response.Validation = validation

	if !validation.OK {
		rewritten, err := s.rewriteAnswer(ctx, question, dialogContext, chunksText, response.Answer, validation)
		if err != nil {
			return nil, err
		}

		response.Answer = strings.TrimSpace(rewritten.Text)
		response.CitationsUsed = copyStrings(rewritten.CitationsUsed)
		response.Citations = copyCitations(rewritten.Citations)
	}

	return response, nil
}
