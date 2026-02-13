package openai

const defaultModel = "gpt-5.2"

type Role string

const (
	RoleSystem    Role = "system"
	RoleDeveloper Role = "developer"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleFunction  Role = "function"
)

type Message struct {
	Role       Role
	Content    string
	Name       string
	ToolCallID string
}

type ResponseFormatType string

const (
	ResponseFormatTypeText       ResponseFormatType = "text"
	ResponseFormatTypeJSONObject ResponseFormatType = "json_object"
	ResponseFormatTypeJSONSchema ResponseFormatType = "json_schema"
)

type ResponseFormat struct {
	Type ResponseFormatType
}

type ChatCompletionRequest struct {
	Messages       []Message
	Temperature    float32
	MaxTokens      int
	TopP           float32
	ResponseFormat *ResponseFormat
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type ChatCompletionResponse struct {
	Content      string
	Role         Role
	FinishReason string
	Usage        Usage
}
