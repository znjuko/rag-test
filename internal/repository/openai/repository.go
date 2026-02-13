package openai

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	goopenai "github.com/sashabaranov/go-openai"
)

type Repository struct {
	cli   *goopenai.Client
	model string
}

func NewRepository(token string) (*Repository, error) {
	if token == "" {
		return nil, errors.New("openai token is empty")
	}

	return &Repository{
		cli:   goopenai.NewClient(token),
		model: defaultModel,
	}, nil
}

func (r *Repository) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("openai: messages is empty")
	}

	openaiReq := goopenai.ChatCompletionRequest{
		Model:                           r.model,
		Messages:                        toOpenAIMessages(req.Messages),
		Temperature:                     req.Temperature,
		TopP:                            req.TopP,
		ChatCompletionRequestExtensions: goopenai.ChatCompletionRequestExtensions{},
	}
	if req.ResponseFormat != nil {
		responseFormat, err := toOpenAIResponseFormat(req.ResponseFormat)
		if err != nil {
			return nil, err
		}
		openaiReq.ResponseFormat = responseFormat
	}

	resp, err := r.cli.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		slog.Error("failed to create chat completion", slog.String("error", err.Error()))
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("openai: empty response")
	}

	choice := resp.Choices[0]
	if choice.Message.Content == "" {
		var choises []string
		for _, ch := range resp.Choices {
			choises = append(choises, ch.Message.Content)
		}

		ch := strings.Join(choises, "\n")
		slog.Error("openai: choice is empty", slog.String("choice", ch))
	}

	return &ChatCompletionResponse{
		Content:      choice.Message.Content,
		Role:         Role(choice.Message.Role),
		FinishReason: string(choice.FinishReason),
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func toOpenAIMessages(messages []Message) []goopenai.ChatCompletionMessage {
	result := make([]goopenai.ChatCompletionMessage, 0, len(messages))
	for _, message := range messages {
		result = append(result, goopenai.ChatCompletionMessage{
			Role:       string(message.Role),
			Content:    message.Content,
			Name:       message.Name,
			ToolCallID: message.ToolCallID,
		})
	}
	return result
}

func toOpenAIResponseFormat(format *ResponseFormat) (*goopenai.ChatCompletionResponseFormat, error) {
	if format == nil {
		return nil, nil
	}

	switch format.Type {
	case ResponseFormatTypeJSONObject:
		return &goopenai.ChatCompletionResponseFormat{
			Type: goopenai.ChatCompletionResponseFormatTypeJSONObject,
		}, nil
	case ResponseFormatTypeText:
		return &goopenai.ChatCompletionResponseFormat{
			Type: goopenai.ChatCompletionResponseFormatTypeText,
		}, nil
	case ResponseFormatTypeJSONSchema:
		return &goopenai.ChatCompletionResponseFormat{
			Type: goopenai.ChatCompletionResponseFormatTypeJSONSchema,
		}, nil
	default:
		return nil, errors.New("openai: unsupported response format")
	}
}
