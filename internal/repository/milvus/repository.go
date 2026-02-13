package milvus

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type VectorRepository interface {
	EnsureCollection(ctx context.Context, name string, dim int) error
	Upsert(ctx context.Context, collection string, items []VectorItem) error
	Search(ctx context.Context, collection string, vector []float32, topK int) ([]SearchHit, error)
	SearchByDataSource(ctx context.Context, collection string, vector []float32, topK int, dataSource string) ([]SearchHit, error)
	Close() error
}

type MilvusRepository struct {
	client client.Client
}

func NewMilvusRepository(ctx context.Context, addr string) (*MilvusRepository, error) {
	c, err := client.NewGrpcClient(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &MilvusRepository{client: c}, nil
}

func (r *MilvusRepository) Close() error {
	return r.client.Close()
}

func (r *MilvusRepository) EnsureCollection(ctx context.Context, name string, dim int) error {
	exists, err := r.client.HasCollection(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		schema := &entity.Schema{
			CollectionName: name,
			Description:    "Documents with embeddings",
			AutoID:         false,
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     false,
				},
				{
					Name:       "embedding",
					DataType:   entity.FieldTypeFloatVector,
					TypeParams: map[string]string{"dim": strconv.Itoa(dim)},
				},
				{
					Name:       "payload",
					DataType:   entity.FieldTypeVarChar,
					TypeParams: map[string]string{"max_length": "4096"},
				},
				{
					Name:       "data_source",
					DataType:   entity.FieldTypeVarChar,
					TypeParams: map[string]string{"max_length": "1024"},
				},
			},
		}

		if err := r.client.CreateCollection(ctx, schema, 2); err != nil {
			return err
		}

		index, err := entity.NewIndexIvfFlat(entity.L2, 128)
		if err != nil {
			return err
		}
		if err := r.client.CreateIndex(ctx, name, "embedding", index, false); err != nil {
			return err
		}
	}

	return r.client.LoadCollection(ctx, name, false)
}

func (r *MilvusRepository) Upsert(ctx context.Context, collection string, items []VectorItem) error {
	if len(items) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(items))
	vectors := make([][]float32, 0, len(items))
	payloads := make([]string, 0, len(items))
	dataSources := make([]string, 0, len(items))

	var dim int
	for i, item := range items {
		if len(item.Embedding) == 0 {
			return fmt.Errorf("empty embedding for item id %d", item.ID)
		}
		if i == 0 {
			dim = len(item.Embedding)
		} else if len(item.Embedding) != dim {
			return fmt.Errorf("embedding dimension mismatch for item id %d", item.ID)
		}
		if strings.TrimSpace(item.DataSource) == "" {
			return fmt.Errorf("data_source is required for item id %d", item.ID)
		}

		ids = append(ids, item.ID)
		vectors = append(vectors, item.Embedding)
		payloads = append(payloads, item.Payload)
		dataSources = append(dataSources, item.DataSource)
	}

	columns := []entity.Column{
		entity.NewColumnInt64("id", ids),
		entity.NewColumnFloatVector("embedding", dim, vectors),
		entity.NewColumnVarChar("payload", payloads),
		entity.NewColumnVarChar("data_source", dataSources),
	}

	_, err := r.client.Insert(ctx, collection, "", columns...)
	if err != nil {
		return err
	}

	return r.client.Flush(ctx, collection, false)
}

func (r *MilvusRepository) Search(ctx context.Context, collection string, vector []float32, topK int) ([]SearchHit, error) {
	if len(vector) == 0 {
		return nil, fmt.Errorf("empty query vector")
	}

	query := []entity.Vector{entity.FloatVector(vector)}
	searchParams, err := entity.NewIndexIvfFlatSearchParam(64)
	if err != nil {
		return nil, err
	}

	results, err := r.client.Search(
		ctx,
		collection,
		[]string{},
		"",
		[]string{"id", "payload", "data_source"},
		query,
		"embedding",
		entity.L2,
		topK,
		searchParams,
	)
	if err != nil {
		return nil, err
	}

	hits := make([]SearchHit, 0, topK)
	for _, result := range results {
		idColumn := result.Fields.GetColumn("id")
		if idColumn == nil {
			idColumn = result.IDs
		}
		payloadColumn := result.Fields.GetColumn("payload")
		sourceColumn := result.Fields.GetColumn("data_source")
		if idColumn == nil {
			return nil, fmt.Errorf("missing id column in search result")
		}
		if payloadColumn == nil {
			return nil, fmt.Errorf("missing payload column in search result")
		}
		if sourceColumn == nil {
			return nil, fmt.Errorf("missing data_source column in search result")
		}

		for i := 0; i < result.ResultCount; i++ {
			id, err := idColumn.GetAsInt64(i)
			if err != nil {
				return nil, err
			}
			payload, err := payloadColumn.GetAsString(i)
			if err != nil {
				return nil, err
			}
			dataSource, err := sourceColumn.GetAsString(i)
			if err != nil {
				return nil, err
			}

			hits = append(hits, SearchHit{
				ID:         id,
				Score:      result.Scores[i],
				Payload:    payload,
				DataSource: dataSource,
			})
		}
	}

	return hits, nil
}

func (r *MilvusRepository) SearchByDataSource(ctx context.Context, collection string, vector []float32, topK int, dataSource string) ([]SearchHit, error) {
	if len(vector) == 0 {
		return nil, fmt.Errorf("empty query vector")
	}
	if strings.TrimSpace(dataSource) == "" {
		return nil, fmt.Errorf("data_source is required")
	}

	query := []entity.Vector{entity.FloatVector(vector)}
	searchParams, err := entity.NewIndexIvfFlatSearchParam(64)
	if err != nil {
		return nil, err
	}

	results, err := r.client.Search(
		ctx,
		collection,
		[]string{},
		buildSingleDataSourceExpr(dataSource),
		[]string{"id", "payload", "data_source"},
		query,
		"embedding",
		entity.L2,
		topK,
		searchParams,
	)
	if err != nil {
		return nil, err
	}

	hits := make([]SearchHit, 0, topK)
	for _, result := range results {
		idColumn := result.Fields.GetColumn("id")
		if idColumn == nil {
			idColumn = result.IDs
		}
		payloadColumn := result.Fields.GetColumn("payload")
		sourceColumn := result.Fields.GetColumn("data_source")
		if idColumn == nil {
			return nil, fmt.Errorf("missing id column in search result")
		}
		if payloadColumn == nil {
			return nil, fmt.Errorf("missing payload column in search result")
		}
		if sourceColumn == nil {
			return nil, fmt.Errorf("missing data_source column in search result")
		}

		for i := 0; i < result.ResultCount; i++ {
			id, err := idColumn.GetAsInt64(i)
			if err != nil {
				return nil, err
			}
			payload, err := payloadColumn.GetAsString(i)
			if err != nil {
				return nil, err
			}
			source, err := sourceColumn.GetAsString(i)
			if err != nil {
				return nil, err
			}

			hits = append(hits, SearchHit{
				ID:         id,
				Score:      result.Scores[i],
				Payload:    payload,
				DataSource: source,
			})
		}
	}

	return hits, nil
}

func buildSingleDataSourceExpr(source string) string {
	trimmed := strings.TrimSpace(source)
	if trimmed == "" {
		return ""
	}
	return "data_source == \"" + escapeMilvusString(trimmed) + "\""
}

func escapeMilvusString(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
}
