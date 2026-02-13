package main

import (
	"context"
	"log/slog"
	"os"
	"rag-test/internal/repository/embeddings"
	milvusrepo "rag-test/internal/repository/milvus"
	openairepo "rag-test/internal/repository/openai"
	"rag-test/internal/service/rag"

	docling_bridge "github.com/Dsouza10082/go-docling-bridge"
)

var (
	collectionName = "testcollection"

	needMigration = false
	token         = os.Getenv("OPENAI_TOKEN")

	embedRepo  *embeddings.Repository
	docling    = docling_bridge.NewDoclingBridge()
	vectorRepo milvusrepo.VectorRepository
	defaultDim = 384

	milvusAddres = "localhost:19530"
)

func main() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)

	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if token == "" {
		slog.Error("failed to get OPENAI_TOKEN")
		return
	}

	var (
		ctx = context.Background()
		err error
	)

	embedRepo, err = embeddings.NewRepository(token)
	if err != nil {
		slog.Error("failed to create embeddings repository", slog.String("error", err.Error()))
		return
	}

	vectorRepo, err = milvusrepo.NewMilvusRepository(ctx, milvusAddres)
	if err != nil {
		slog.Error("failed to init milvus repository", slog.String("err", err.Error()))
		return
	}

	if err := vectorRepo.EnsureCollection(ctx, collectionName, defaultDim); err != nil {
		slog.Error("failed to ensure collection", slog.String("err", err.Error()))
		return
	}

	if err := processAllFiles(ctx); err != nil {
		slog.Error("failed to process all files", slog.String("err", err.Error()))
		return
	}

	openaiRepo, err := openairepo.NewRepository(token)
	if err != nil {
		slog.Error("failed to create openaii repository", slog.String("error", err.Error()))
		return
	}

	ragSvc := rag.NewService(openaiRepo, embedRepo, vectorRepo, collectionName, 10)

	if err := runConsoleChat(ctx, ragSvc); err != nil {
		slog.Error("chat failed", slog.String("error", err.Error()))
	}
}
