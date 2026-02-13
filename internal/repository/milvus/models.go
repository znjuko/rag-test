package milvus

type VectorItem struct {
	ID         int64
	Embedding  []float32
	Payload    string
	DataSource string
}

type SearchHit struct {
	ID         int64
	Score      float32
	Payload    string
	DataSource string
}
