package rag

import openairepo "rag-test/internal/repository/openai"

type DialogMessage struct {
	Role    openairepo.Role
	Content string
}

type Request struct {
	Question      string
	History       []DialogMessage
	DialogContext string
	TopK          int
}

type Response struct {
	NeedClarification  bool
	ClarifyingQuestion string
	MissingSlots       []string
	Assumptions        []string
	SuggestedQueries   []string

	Answer        string
	CitationsUsed []string
	Citations     []Citation
	Chunks        []Chunk
	Validation    ValidationResult
}

type Chunk struct {
	ID         string
	DataSource string
	Text       string
}

type Citation struct {
	ID         string
	Quote      string
	DataSource string
}

type ValidationResult struct {
	OK                bool
	UnsupportedClaims []string
	Notes             string
}
